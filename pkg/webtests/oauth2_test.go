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
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/modules/auth/oauth2server"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOAuth2AuthorizeEndpoint(t *testing.T) {
	t.Run("redirects unauthenticated user to login", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/oauth/authorize?response_type=code&client_id=vikunja-flutter&redirect_uri=vikunja://callback&code_challenge=abc123&code_challenge_method=S256&state=teststate", nil)
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusFound, rec.Code)
		location := rec.Header().Get("Location")
		assert.Contains(t, location, "/login")
		assert.Contains(t, location, "redirect=")
	})

	t.Run("issues code for authenticated user", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		token, err := auth.NewUserJWTAuthtoken(&testuser1, "test-session-id")
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/oauth/authorize?response_type=code&client_id=vikunja-flutter&redirect_uri=vikunja://callback&code_challenge=E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM&code_challenge_method=S256&state=teststate", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusFound, rec.Code)
		location := rec.Header().Get("Location")
		assert.True(t, strings.HasPrefix(location, "vikunja://callback"))
		locationURL, err := url.Parse(location)
		require.NoError(t, err)
		assert.NotEmpty(t, locationURL.Query().Get("code"))
		assert.Equal(t, "teststate", locationURL.Query().Get("state"))
	})

	t.Run("rejects invalid client_id", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		token, err := auth.NewUserJWTAuthtoken(&testuser1, "test-session-id")
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/oauth/authorize?response_type=code&client_id=unknown&redirect_uri=vikunja://callback&code_challenge=test&code_challenge_method=S256", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("rejects invalid redirect_uri", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		token, err := auth.NewUserJWTAuthtoken(&testuser1, "test-session-id")
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/oauth/authorize?response_type=code&client_id=vikunja-flutter&redirect_uri=https://evil.com/callback&code_challenge=test&code_challenge_method=S256", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("rejects missing PKCE", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		token, err := auth.NewUserJWTAuthtoken(&testuser1, "test-session-id")
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/oauth/authorize?response_type=code&client_id=vikunja-flutter&redirect_uri=vikunja://callback", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestOAuth2TokenEndpoint(t *testing.T) {
	t.Run("full authorization code flow with PKCE", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		// Step 1: Get an authorization code
		codeVerifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
		h := sha256.Sum256([]byte(codeVerifier))
		codeChallenge := base64.RawURLEncoding.EncodeToString(h[:])

		token, err := auth.NewUserJWTAuthtoken(&testuser1, "test-session-id")
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/oauth/authorize?response_type=code&client_id=vikunja-flutter&redirect_uri=vikunja://callback&code_challenge="+codeChallenge+"&code_challenge_method=S256&state=xyz", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		e.ServeHTTP(rec, req)

		require.Equal(t, http.StatusFound, rec.Code)
		location, err := url.Parse(rec.Header().Get("Location"))
		require.NoError(t, err)
		code := location.Query().Get("code")
		require.NotEmpty(t, code)

		// Step 2: Exchange the code for tokens
		form := url.Values{}
		form.Set("grant_type", "authorization_code")
		form.Set("code", code)
		form.Set("client_id", "vikunja-flutter")
		form.Set("redirect_uri", "vikunja://callback")
		form.Set("code_verifier", codeVerifier)

		rec2 := httptest.NewRecorder()
		req2 := httptest.NewRequest(http.MethodPost, "/api/v1/oauth/token", strings.NewReader(form.Encode()))
		req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		e.ServeHTTP(rec2, req2)

		require.Equal(t, http.StatusOK, rec2.Code)

		var tokenResp oauth2server.TokenResponse
		err = json.Unmarshal(rec2.Body.Bytes(), &tokenResp)
		require.NoError(t, err)
		assert.NotEmpty(t, tokenResp.AccessToken)
		assert.Equal(t, "bearer", tokenResp.TokenType)
		assert.NotEmpty(t, tokenResp.RefreshToken)
		assert.Greater(t, tokenResp.ExpiresIn, int64(0))
	})

	t.Run("code is single-use", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		// Get a code
		codeVerifier := "test-verifier-for-single-use-check"
		h := sha256.Sum256([]byte(codeVerifier))
		codeChallenge := base64.RawURLEncoding.EncodeToString(h[:])

		token, err := auth.NewUserJWTAuthtoken(&testuser1, "test-session-id")
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/oauth/authorize?response_type=code&client_id=vikunja-flutter&redirect_uri=vikunja://callback&code_challenge="+codeChallenge+"&code_challenge_method=S256", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		e.ServeHTTP(rec, req)

		location, _ := url.Parse(rec.Header().Get("Location"))
		code := location.Query().Get("code")

		// First exchange succeeds
		form := url.Values{}
		form.Set("grant_type", "authorization_code")
		form.Set("code", code)
		form.Set("client_id", "vikunja-flutter")
		form.Set("redirect_uri", "vikunja://callback")
		form.Set("code_verifier", codeVerifier)

		rec2 := httptest.NewRecorder()
		req2 := httptest.NewRequest(http.MethodPost, "/api/v1/oauth/token", strings.NewReader(form.Encode()))
		req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		e.ServeHTTP(rec2, req2)
		require.Equal(t, http.StatusOK, rec2.Code)

		// Second exchange fails
		rec3 := httptest.NewRecorder()
		req3 := httptest.NewRequest(http.MethodPost, "/api/v1/oauth/token", strings.NewReader(form.Encode()))
		req3.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		e.ServeHTTP(rec3, req3)
		assert.Equal(t, http.StatusBadRequest, rec3.Code)
	})

	t.Run("rejects wrong PKCE verifier", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		codeVerifier := "correct-verifier"
		h := sha256.Sum256([]byte(codeVerifier))
		codeChallenge := base64.RawURLEncoding.EncodeToString(h[:])

		token, err := auth.NewUserJWTAuthtoken(&testuser1, "test-session-id")
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/oauth/authorize?response_type=code&client_id=vikunja-flutter&redirect_uri=vikunja://callback&code_challenge="+codeChallenge+"&code_challenge_method=S256", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		e.ServeHTTP(rec, req)

		location, _ := url.Parse(rec.Header().Get("Location"))
		code := location.Query().Get("code")

		form := url.Values{}
		form.Set("grant_type", "authorization_code")
		form.Set("code", code)
		form.Set("client_id", "vikunja-flutter")
		form.Set("redirect_uri", "vikunja://callback")
		form.Set("code_verifier", "wrong-verifier")

		rec2 := httptest.NewRecorder()
		req2 := httptest.NewRequest(http.MethodPost, "/api/v1/oauth/token", strings.NewReader(form.Encode()))
		req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		e.ServeHTTP(rec2, req2)
		assert.Equal(t, http.StatusBadRequest, rec2.Code)
	})

	t.Run("rejects invalid grant_type", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		form := url.Values{}
		form.Set("grant_type", "password")
		form.Set("client_id", "vikunja-flutter")

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/oauth/token", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("refresh token flow", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		// Get initial tokens via authorization code flow
		codeVerifier := "refresh-flow-test-verifier"
		h := sha256.Sum256([]byte(codeVerifier))
		codeChallenge := base64.RawURLEncoding.EncodeToString(h[:])

		token, err := auth.NewUserJWTAuthtoken(&testuser1, "test-session-id")
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/oauth/authorize?response_type=code&client_id=vikunja-flutter&redirect_uri=vikunja://callback&code_challenge="+codeChallenge+"&code_challenge_method=S256", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		e.ServeHTTP(rec, req)

		location, _ := url.Parse(rec.Header().Get("Location"))
		code := location.Query().Get("code")

		form := url.Values{}
		form.Set("grant_type", "authorization_code")
		form.Set("code", code)
		form.Set("client_id", "vikunja-flutter")
		form.Set("redirect_uri", "vikunja://callback")
		form.Set("code_verifier", codeVerifier)

		rec2 := httptest.NewRecorder()
		req2 := httptest.NewRequest(http.MethodPost, "/api/v1/oauth/token", strings.NewReader(form.Encode()))
		req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		e.ServeHTTP(rec2, req2)
		require.Equal(t, http.StatusOK, rec2.Code)

		var tokenResp oauth2server.TokenResponse
		_ = json.Unmarshal(rec2.Body.Bytes(), &tokenResp)

		// Use the refresh token to get new tokens
		refreshForm := url.Values{}
		refreshForm.Set("grant_type", "refresh_token")
		refreshForm.Set("refresh_token", tokenResp.RefreshToken)
		refreshForm.Set("client_id", "vikunja-flutter")

		rec3 := httptest.NewRecorder()
		req3 := httptest.NewRequest(http.MethodPost, "/api/v1/oauth/token", strings.NewReader(refreshForm.Encode()))
		req3.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		e.ServeHTTP(rec3, req3)

		require.Equal(t, http.StatusOK, rec3.Code)

		var refreshResp oauth2server.TokenResponse
		err = json.Unmarshal(rec3.Body.Bytes(), &refreshResp)
		require.NoError(t, err)
		assert.NotEmpty(t, refreshResp.AccessToken)
		assert.NotEmpty(t, refreshResp.RefreshToken)
		// The refresh token should have been rotated
		assert.NotEqual(t, tokenResp.RefreshToken, refreshResp.RefreshToken)
	})

	t.Run("refresh token rotation prevents replay", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		// Get initial tokens
		codeVerifier := "replay-test-verifier"
		h := sha256.Sum256([]byte(codeVerifier))
		codeChallenge := base64.RawURLEncoding.EncodeToString(h[:])

		token, err := auth.NewUserJWTAuthtoken(&testuser1, "test-session-id")
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/oauth/authorize?response_type=code&client_id=vikunja-flutter&redirect_uri=vikunja://callback&code_challenge="+codeChallenge+"&code_challenge_method=S256", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		e.ServeHTTP(rec, req)

		location, _ := url.Parse(rec.Header().Get("Location"))
		code := location.Query().Get("code")

		form := url.Values{}
		form.Set("grant_type", "authorization_code")
		form.Set("code", code)
		form.Set("client_id", "vikunja-flutter")
		form.Set("redirect_uri", "vikunja://callback")
		form.Set("code_verifier", codeVerifier)

		rec2 := httptest.NewRecorder()
		req2 := httptest.NewRequest(http.MethodPost, "/api/v1/oauth/token", strings.NewReader(form.Encode()))
		req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		e.ServeHTTP(rec2, req2)

		var tokenResp oauth2server.TokenResponse
		_ = json.Unmarshal(rec2.Body.Bytes(), &tokenResp)
		oldRefreshToken := tokenResp.RefreshToken

		// First refresh succeeds
		refreshForm := url.Values{}
		refreshForm.Set("grant_type", "refresh_token")
		refreshForm.Set("refresh_token", oldRefreshToken)
		refreshForm.Set("client_id", "vikunja-flutter")

		rec3 := httptest.NewRecorder()
		req3 := httptest.NewRequest(http.MethodPost, "/api/v1/oauth/token", strings.NewReader(refreshForm.Encode()))
		req3.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		e.ServeHTTP(rec3, req3)
		require.Equal(t, http.StatusOK, rec3.Code)

		// Replay the same old refresh token — should fail
		rec4 := httptest.NewRecorder()
		req4 := httptest.NewRequest(http.MethodPost, "/api/v1/oauth/token", strings.NewReader(refreshForm.Encode()))
		req4.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		e.ServeHTTP(rec4, req4)
		assert.Equal(t, http.StatusUnauthorized, rec4.Code)
	})
}
