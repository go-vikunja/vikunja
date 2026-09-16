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
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/routes"

	"github.com/labstack/echo/v5"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const unauthRateLimitForTest = 4

// /api/v2 has no unauthenticated route group - without an explicit per-path
// limiter its credential endpoints would be unthrottled on a default install,
// where the global rate limiter is off.
func TestV2UnauthRateLimit(t *testing.T) {
	_, err := setupTestEnv()
	require.NoError(t, err)
	require.False(t, config.RateLimitEnabled.GetBool(), "the global rate limiter must default to off for this test to be meaningful")

	previous := config.RateLimitNoAuthRoutesLimit.GetInt64()
	config.RateLimitNoAuthRoutesLimit.Set(unauthRateLimitForTest)
	defer config.RateLimitNoAuthRoutesLimit.Set(previous)

	// A fresh instance per subtest: RegisterRoutes builds the limiter, so this
	// hands every subtest an untouched budget.
	newRoutes := func() *echo.Echo {
		e := routes.NewEcho()
		routes.RegisterRoutes(e)
		return e
	}

	for _, path := range []string{
		"/api/v2/register",
		"/api/v2/user/password/token",
		"/api/v2/user/password/reset",
		"/api/v2/user/confirm",
		"/api/v2/login",
		"/api/v2/shares/1/auth",
	} {
		t.Run(path, func(t *testing.T) {
			e := newRoutes()

			for i := 0; i < unauthRateLimitForTest; i++ {
				rec := humaRequest(t, e, http.MethodPost, path, "{}", "", "")
				require.NotEqual(t, http.StatusTooManyRequests, rec.Code, "request %d, body: %s", i, rec.Body.String())
				assert.Equal(t, strconv.Itoa(unauthRateLimitForTest), rec.Header().Get("X-RateLimit-Limit"))
				assert.Equal(t, strconv.Itoa(unauthRateLimitForTest-i-1), rec.Header().Get("X-RateLimit-Remaining"), "request %d", i)
			}

			rec := humaRequest(t, e, http.MethodPost, path, "{}", "", "")
			assert.Equal(t, http.StatusTooManyRequests, rec.Code, "body: %s", rec.Body.String())
			assert.Equal(t, "0", rec.Header().Get("X-RateLimit-Remaining"))
		})
	}

	// Per-version limiters would hand an attacker twice the guesses.
	t.Run("v1 and v2 share the budget", func(t *testing.T) {
		e := newRoutes()

		paths := []string{
			"/api/v2/login",
			"/api/v1/login",
			"/api/v2/register",
			"/api/v1/register",
		}
		for i, path := range paths {
			rec := humaRequest(t, e, http.MethodPost, path, "{}", "", "")
			require.NotEqual(t, http.StatusTooManyRequests, rec.Code, "request %d (%s), body: %s", i, path, rec.Body.String())
			assert.Equal(t, strconv.Itoa(len(paths)-i-1), rec.Header().Get("X-RateLimit-Remaining"), "request %d (%s)", i, path)
		}

		rec := humaRequest(t, e, http.MethodPost, "/api/v1/login", "{}", "", "")
		assert.Equal(t, http.StatusTooManyRequests, rec.Code, "body: %s", rec.Body.String())
	})

	t.Run("oauth token endpoint uses the renewal budget", func(t *testing.T) {
		e := newRoutes()

		rec := humaRequest(t, e, http.MethodPost, "/api/v2/oauth/token", `{"grant_type":"refresh_token","refresh_token":"garbage"}`, "", "")
		require.NotEqual(t, http.StatusTooManyRequests, rec.Code, "body: %s", rec.Body.String())
		assert.Equal(t, strconv.FormatInt(config.RateLimitTokenRefreshLimit.GetInt64(), 10), rec.Header().Get("X-RateLimit-Limit"))

		rec = humaRequest(t, e, http.MethodPost, "/api/v2/login", "{}", "", "")
		assert.Equal(t, strconv.Itoa(unauthRateLimitForTest-1), rec.Header().Get("X-RateLimit-Remaining"), "the renewal above must not spend a login guess")
	})

	t.Run("unrelated paths are not charged", func(t *testing.T) {
		e := newRoutes()

		for i := 0; i < unauthRateLimitForTest+2; i++ {
			rec := humaRequest(t, e, http.MethodGet, "/api/v2/info", "", "", "")
			require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
			assert.Empty(t, rec.Header().Get("X-RateLimit-Limit"))
		}

		rec := humaRequest(t, e, http.MethodPost, "/api/v2/login", "{}", "", "")
		assert.Equal(t, strconv.Itoa(unauthRateLimitForTest-1), rec.Header().Get("X-RateLimit-Remaining"))
	})
}

// Guards the per-IP bcrypt budget (GHSA-m469-88xx-8rx2).
func TestBasicAuthRateLimit(t *testing.T) {
	_, err := setupTestEnv()
	require.NoError(t, err)

	const basicAuthLimitForTest = 3
	previous := config.RateLimitBasicAuthLimit.GetString()
	config.RateLimitBasicAuthLimit.Set(basicAuthLimitForTest)
	defer config.RateLimitBasicAuthLimit.Set(previous)

	previousNoAuth := config.RateLimitNoAuthRoutesLimit.GetString()
	config.RateLimitNoAuthRoutesLimit.Set(unauthRateLimitForTest)
	defer config.RateLimitNoAuthRoutesLimit.Set(previousNoAuth)

	// RegisterRoutes constructs the limiter, giving each subtest a fresh budget.
	newRoutes := func() *echo.Echo {
		e := routes.NewEcho()
		routes.RegisterRoutes(e)
		return e
	}

	caldavRequest := func(t *testing.T, e *echo.Echo, username, password string) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest("PROPFIND", "/dav/projects", strings.NewReader(""))
		req.Header.Set(echo.HeaderContentType, echo.MIMETextXML)
		req.Header.Set("Depth", "0")
		if username != "" {
			req.SetBasicAuth(username, password)
		}
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		return rec
	}

	t.Run("wrong dav passwords eventually reach 429", func(t *testing.T) {
		e := newRoutes()

		for i := 0; i < basicAuthLimitForTest; i++ {
			rec := caldavRequest(t, e, testuser15.Username, "wrong-password")
			require.Equal(t, http.StatusUnauthorized, rec.Code, "failed attempt %d, body: %s", i, rec.Body.String())
		}

		rec := caldavRequest(t, e, testuser15.Username, "wrong-password")
		assert.Equal(t, http.StatusTooManyRequests, rec.Code, "body: %s", rec.Body.String())
	})

	t.Run("successful propfind bursts are never limited", func(t *testing.T) {
		e := newRoutes()

		for i := 0; i < 20; i++ {
			rec := caldavRequest(t, e, testuser15.Username, "12345678")
			require.NotEqual(t, http.StatusTooManyRequests, rec.Code, "request %d", i)
			require.NotEqual(t, http.StatusUnauthorized, rec.Code, "the valid credentials must authenticate, request %d", i)
		}
	})

	t.Run("rejection happens before the password check", func(t *testing.T) {
		e := newRoutes()

		for i := 0; i < basicAuthLimitForTest; i++ {
			rec := caldavRequest(t, e, testuser15.Username, "wrong-password")
			require.Equal(t, http.StatusUnauthorized, rec.Code)
		}

		rec := caldavRequest(t, e, testuser15.Username, "12345678")
		assert.Equal(t, http.StatusTooManyRequests, rec.Code, "body: %s", rec.Body.String())
	})

	t.Run("requests without credentials are only challenged", func(t *testing.T) {
		e := newRoutes()

		for i := 0; i < 20; i++ {
			rec := caldavRequest(t, e, "", "")
			require.Equal(t, http.StatusUnauthorized, rec.Code, "the 401 challenge is not a guess, request %d", i)
		}
	})

	t.Run("the budget is isolated from the login budget", func(t *testing.T) {
		e := newRoutes()

		for i := 0; i < basicAuthLimitForTest; i++ {
			rec := caldavRequest(t, e, testuser15.Username, "wrong-password")
			require.Equal(t, http.StatusUnauthorized, rec.Code)
		}
		rec := caldavRequest(t, e, testuser15.Username, "wrong-password")
		require.Equal(t, http.StatusTooManyRequests, rec.Code)

		rec = humaRequest(t, e, http.MethodPost, "/api/v1/login", "{}", "", "")
		assert.NotEqual(t, http.StatusTooManyRequests, rec.Code, "body: %s", rec.Body.String())
	})

	t.Run("the feed route shares the basic auth budget", func(t *testing.T) {
		e := newRoutes()

		req := httptest.NewRequest(http.MethodGet, "/feeds/notifications.atom", nil)
		req.SetBasicAuth(testuser15.Username, "wrong-password")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		require.Equal(t, http.StatusUnauthorized, rec.Code)

		rec = caldavRequest(t, e, testuser15.Username, "wrong-password")
		require.Equal(t, http.StatusUnauthorized, rec.Code)
		rec = caldavRequest(t, e, testuser15.Username, "wrong-password")
		require.Equal(t, http.StatusUnauthorized, rec.Code)

		rec = caldavRequest(t, e, testuser15.Username, "wrong-password")
		assert.Equal(t, http.StatusTooManyRequests, rec.Code,
			"the failed feed auth must have spent the shared basic auth budget")
	})
}
