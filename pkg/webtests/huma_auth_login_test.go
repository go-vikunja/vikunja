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
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/user"

	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// refreshCookie returns the Set-Cookie value for the refresh-token cookie, or ""
// if the response set no such cookie.
func refreshCookie(rec *httptest.ResponseRecorder) *http.Cookie {
	for _, c := range rec.Result().Cookies() {
		if c.Name == auth.RefreshTokenCookieName {
			return c
		}
	}
	return nil
}

// TestHumaLogin ports the v1 login coverage to /api/v2: it asserts the token
// response, the HttpOnly refresh cookie, the no-store header, and the credential
// and TOTP gates.
func TestHumaLogin(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)

	login := func(body string) *httptest.ResponseRecorder {
		return humaRequest(t, e, http.MethodPost, "/api/v2/login", body, "", "")
	}

	t.Run("normal login", func(t *testing.T) {
		rec := login(`{"username":"user1","password":"12345678"}`)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"token":"`)

		assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"))

		cookie := refreshCookie(rec)
		require.NotNil(t, cookie, "login must set the refresh-token cookie")
		assert.NotEmpty(t, cookie.Value)
		assert.True(t, cookie.HttpOnly, "refresh cookie must be HttpOnly")
	})

	t.Run("wrong password", func(t *testing.T) {
		rec := login(`{"username":"user1","password":"wrong"}`)
		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.Equal(t, user.ErrCodeWrongUsernameOrPassword, problemCode(t, rec))
		assert.Nil(t, refreshCookie(rec), "a failed login must not set a refresh cookie")
	})

	t.Run("nonexistent user", func(t *testing.T) {
		rec := login(`{"username":"userWhichDoesNotExist","password":"12345678"}`)
		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.Equal(t, user.ErrCodeWrongUsernameOrPassword, problemCode(t, rec))
	})

	t.Run("unconfirmed email", func(t *testing.T) {
		rec := login(`{"username":"user5","password":"12345678"}`)
		assert.Equal(t, http.StatusPreconditionFailed, rec.Code)
		assert.Equal(t, user.ErrCodeEmailNotConfirmed, problemCode(t, rec))
	})

	t.Run("disabled account", func(t *testing.T) {
		rec := login(`{"username":"user17","password":"12345678"}`)
		assert.Equal(t, http.StatusPreconditionFailed, rec.Code)
		assert.Equal(t, user.ErrCodeAccountDisabled, problemCode(t, rec))
	})

	t.Run("locked account", func(t *testing.T) {
		rec := login(`{"username":"user18","password":"12345678"}`)
		assert.Equal(t, http.StatusPreconditionFailed, rec.Code)
		assert.Equal(t, user.ErrCodeAccountLocked, problemCode(t, rec))
	})

	t.Run("TOTP required but missing", func(t *testing.T) {
		rec := login(`{"username":"user10","password":"12345678"}`)
		assert.Equal(t, http.StatusPreconditionFailed, rec.Code)
		assert.Equal(t, user.ErrCodeInvalidTOTPPasscode, problemCode(t, rec))
	})

	t.Run("TOTP wrong", func(t *testing.T) {
		rec := login(`{"username":"user10","password":"12345678","totp_passcode":"000000"}`)
		assert.Equal(t, http.StatusPreconditionFailed, rec.Code)
		assert.Equal(t, user.ErrCodeInvalidTOTPPasscode, problemCode(t, rec))
	})

	t.Run("TOTP correct", func(t *testing.T) {
		code, err := totp.GenerateCode("JBSWY3DPEHPK3PXP", time.Now())
		require.NoError(t, err)
		rec := login(`{"username":"user10","password":"12345678","totp_passcode":"` + code + `"}`)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"token":"`)
		assert.NotNil(t, refreshCookie(rec))
	})
}

// TestHumaLogout proves the v2 logout deletes the session server-side and clears
// the refresh-token cookie.
func TestHumaLogout(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)

	// Create a session so logout has something to delete, then mint a JWT whose
	// sid claim points at it.
	s := db.NewSession()
	session, err := models.CreateSession(s, testuser1.ID, "test", "127.0.0.1", false)
	require.NoError(t, err)
	require.NoError(t, s.Commit())
	require.NoError(t, s.Close())

	token, err := auth.NewUserJWTAuthtoken(&testuser1, session.ID)
	require.NoError(t, err)

	rec := humaRequest(t, e, http.MethodPost, "/api/v2/logout", "", token, "")
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.Contains(t, rec.Body.String(), "Successfully logged out.")

	cookie := refreshCookie(rec)
	require.NotNil(t, cookie, "logout must clear the refresh cookie")
	assert.Empty(t, cookie.Value, "cleared cookie has no value")
	assert.Negative(t, cookie.MaxAge, "cleared cookie is expired")

	// The session must be gone.
	check := db.NewSession()
	defer check.Close()
	exists, err := check.Where("id = ?", session.ID).Exist(&models.Session{})
	require.NoError(t, err)
	assert.False(t, exists, "logout must delete the session")
}

// TestHumaLoginUnauthenticated proves login needs no token (it is a public op).
func TestHumaLoginUnauthenticated(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)

	rec := humaRequest(t, e, http.MethodPost, "/api/v2/login", `{"username":"user1","password":"12345678"}`, "", "")
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
}

// TestHumaOpenIDGating proves the OIDC callback route only exists when OpenID is
// enabled, mirroring the registrar gate.
func TestHumaOpenIDGating(t *testing.T) {
	body := `{"code":"abc","redirect_url":"https://example.com"}`

	t.Run("disabled returns 404", func(t *testing.T) {
		config.AuthOpenIDEnabled.Set(false)

		e, err := setupTestEnv()
		require.NoError(t, err)

		rec := humaRequest(t, e, http.MethodPost, "/api/v2/auth/openid/test/callback", body, "", "")
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("enabled does not require auth", func(t *testing.T) {
		config.AuthOpenIDEnabled.Set(true)
		defer config.AuthOpenIDEnabled.Set(false)

		e, err := setupTestEnv()
		require.NoError(t, err)

		// No provider is configured, so the call fails downstream — but it must
		// not 404 as an unknown route nor 401 for missing auth, which proves the
		// public route is registered.
		rec := humaRequest(t, e, http.MethodPost, "/api/v2/auth/openid/doesnotexist/callback", body, "", "")
		assert.NotEqual(t, http.StatusNotFound, rec.Code)
		assert.NotEqual(t, http.StatusUnauthorized, rec.Code)
	})
}
