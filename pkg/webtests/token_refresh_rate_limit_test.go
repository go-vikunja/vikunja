// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package webtests

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/routes"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Regression test for #3474: routine renewals must not exhaust the /login floor.
func TestTokenRefreshRateLimit(t *testing.T) {
	_, err := setupTestEnv()
	require.NoError(t, err)
	require.False(t, config.RateLimitEnabled.GetBool(), "the global rate limiter must default to off for this test to be meaningful")

	previousNoAuth := config.RateLimitNoAuthRoutesLimit.GetInt64()
	previousRefresh := config.RateLimitTokenRefreshLimit.GetInt64()
	config.RateLimitNoAuthRoutesLimit.Set(2)
	config.RateLimitTokenRefreshLimit.Set(4)
	defer func() {
		config.RateLimitNoAuthRoutesLimit.Set(previousNoAuth)
		config.RateLimitTokenRefreshLimit.Set(previousRefresh)
	}()

	e := routes.NewEcho()
	routes.RegisterRoutes(e)

	// Burn more session renewals than the login floor allows. The requests
	// themselves fail - there is no valid refresh token - but they must be
	// accounted against the renewal budget only.
	for _, path := range []string{
		"/api/v1/user/token/refresh",
		"/api/v1/user/token/refresh",
		"/api/v1/oauth/token",
	} {
		rec := humaRequest(t, e, http.MethodPost, path, `{"grant_type":"refresh_token","refresh_token":"garbage"}`, "", "")
		require.NotEqual(t, http.StatusTooManyRequests, rec.Code, "path: %s, body: %s", path, rec.Body.String())
		assert.Equal(t, "4", rec.Header().Get("X-RateLimit-Limit"))
	}

	t.Run("login still has its full budget", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodPost, "/api/v1/login", `{"username":"user1","password":"12345678"}`, "", "")
		assert.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Equal(t, "2", rec.Header().Get("X-RateLimit-Limit"))
		assert.Equal(t, "1", rec.Header().Get("X-RateLimit-Remaining"))
	})

	t.Run("renewals stay bounded", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodPost, "/api/v1/user/token/refresh", "", "", "")
		require.NotEqual(t, http.StatusTooManyRequests, rec.Code, "body: %s", rec.Body.String())

		rec = humaRequest(t, e, http.MethodPost, "/api/v1/user/token/refresh", "", "", "")
		assert.Equal(t, http.StatusTooManyRequests, rec.Code, "body: %s", rec.Body.String())
		assert.Equal(t, "0", rec.Header().Get("X-RateLimit-Remaining"))
	})

	// Fresh instance: v1 and v2 count against the same per-IP budget, so reusing
	// the outer e would carry over budget already spent above.
	t.Run("v1 and v2 share the renewal budget", func(t *testing.T) {
		// CORS is read at registration, so it must be set before building e.
		prevEnable, prevOrigins := config.CorsEnable.GetBool(), config.CorsOrigins.GetStringSlice()
		config.CorsEnable.Set(true)
		config.CorsOrigins.Set([]string{"http://localhost:*"})
		t.Cleanup(func() {
			config.CorsEnable.Set(prevEnable)
			config.CorsOrigins.Set(prevOrigins)
		})

		e := routes.NewEcho()
		routes.RegisterRoutes(e)

		paths := []string{
			"/api/v2/user/token/refresh",
			"/api/v1/user/token/refresh",
			"/api/v2/user/token/refresh",
			"/api/v1/user/token/refresh",
		}
		for i, path := range paths {
			rec := humaRequest(t, e, http.MethodPost, path, "", "", "")
			require.NotEqual(t, http.StatusTooManyRequests, rec.Code, "request %d (%s), body: %s", i, path, rec.Body.String())
			assert.Equal(t, strconv.Itoa(len(paths)-i-1), rec.Header().Get("X-RateLimit-Remaining"), "request %d (%s)", i, path)
		}

		for _, path := range []string{
			"/api/v1/user/token/refresh",
			"/api/v2/user/token/refresh",
		} {
			req := httptest.NewRequest(http.MethodPost, path, nil)
			req.Header.Set(echo.HeaderOrigin, "http://localhost:5173")
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			assert.Equal(t, http.StatusTooManyRequests, rec.Code, "path: %s, body: %s", path, rec.Body.String())
			assert.Equal(t, "0", rec.Header().Get("X-RateLimit-Remaining"), "path: %s", path)
			assertRetryAfter(t, rec)
			// A cross-origin frontend can only read Retry-After if CORS exposes it.
			assert.Equal(t, "http://localhost:5173", rec.Header().Get(echo.HeaderAccessControlAllowOrigin), "path: %s", path)
			assert.Contains(t, rec.Header().Values(echo.HeaderAccessControlExposeHeaders), echo.HeaderRetryAfter, "path: %s", path)
		}
	})
}

func assertRetryAfter(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	retryAfter, err := strconv.Atoi(rec.Header().Get("Retry-After"))
	require.NoError(t, err, "Retry-After: %q", rec.Header().Get("Retry-After"))
	assert.GreaterOrEqual(t, retryAfter, 1)
	assert.LessOrEqual(t, retryAfter, 61)
}
