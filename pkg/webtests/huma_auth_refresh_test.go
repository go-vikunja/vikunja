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
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/modules/auth"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// refreshRequest posts to the v2 refresh endpoint with the given refresh-token
// cookie value (empty value omits the cookie entirely), driving the full
// echo+Huma stack so cookie reading and Set-Cookie writing are exercised.
func refreshRequest(e *echo.Echo, refreshToken string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/v2/user/token/refresh", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/json")
	if refreshToken != "" {
		req.AddCookie(&http.Cookie{Name: auth.RefreshTokenCookieName, Value: refreshToken})
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

// TestHumaRefreshToken ports the v1 refresh-token coverage to /api/v2: a valid
// cookie yields a new JWT and a rotated HttpOnly cookie, the old token then stops
// working, and missing/invalid cookies map to the same 401 v1 returns.
func TestHumaRefreshToken(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)

	t.Run("valid refresh token", func(t *testing.T) {
		rec := refreshRequest(e, "testtoken_session1")
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"token":"`)
		assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"))

		cookie := refreshCookie(rec)
		require.NotNil(t, cookie, "refresh must set a new refresh-token cookie")
		assert.NotEmpty(t, cookie.Value)
		assert.NotEqual(t, "testtoken_session1", cookie.Value, "refresh token must be rotated")
		assert.True(t, cookie.HttpOnly, "refresh cookie must be HttpOnly")
	})

	t.Run("rotation invalidates the old token", func(t *testing.T) {
		// session2 is a separate session so this case does not depend on the
		// one above. The first refresh succeeds and rotates the token.
		first := refreshRequest(e, "testtoken_session2")
		require.Equal(t, http.StatusOK, first.Code, first.Body.String())
		newCookie := refreshCookie(first)
		require.NotNil(t, newCookie)

		// Replaying the now-rotated token must fail.
		replay := refreshRequest(e, "testtoken_session2")
		assert.Equal(t, http.StatusUnauthorized, replay.Code)

		// The freshly rotated token still works.
		next := refreshRequest(e, newCookie.Value)
		assert.Equal(t, http.StatusOK, next.Code, next.Body.String())
	})

	t.Run("missing cookie", func(t *testing.T) {
		rec := refreshRequest(e, "")
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("invalid cookie", func(t *testing.T) {
		rec := refreshRequest(e, "garbage")
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}
