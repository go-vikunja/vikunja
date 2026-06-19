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
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/modules/auth/oauth2server"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// authorizeRequestBody builds a JSON body for the authorize endpoint.
func authorizeRequestBody(responseType, clientID, redirectURI, codeChallenge, codeChallengeMethod, state string) []byte {
	body, _ := json.Marshal(map[string]string{ //nolint:errchkjson
		"response_type":         responseType,
		"client_id":             clientID,
		"redirect_uri":          redirectURI,
		"code_challenge":        codeChallenge,
		"code_challenge_method": codeChallengeMethod,
		"state":                 state,
	})
	return body
}

// doAuthorize performs a POST to /api/v1/oauth/authorize with the given JWT token and returns the recorder.
func doAuthorize(e http.Handler, token string, body []byte) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/oauth/authorize", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	e.ServeHTTP(rec, req)
	return rec
}

// doTokenRequest performs a JSON POST to /api/v1/oauth/token and returns the recorder.
func doTokenRequest(e http.Handler, params map[string]string) *httptest.ResponseRecorder {
	body, _ := json.Marshal(params) //nolint:errchkjson
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/oauth/token", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	e.ServeHTTP(rec, req)
	return rec
}

func TestOAuth2AuthorizeEndpoint(t *testing.T) {
	t.Run("rejects unauthenticated request", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		body := authorizeRequestBody("code", "vikunja", "vikunja-flutter://callback", "abc123", "S256", "teststate")
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/oauth/authorize", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("issues code for authenticated user", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		token, err := auth.NewUserJWTAuthtoken(&testuser1, "test-session-id")
		require.NoError(t, err)

		body := authorizeRequestBody("code", "vikunja", "vikunja-flutter://callback", "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM", "S256", "teststate")
		rec := doAuthorize(e, token, body)

		require.Equal(t, http.StatusOK, rec.Code)

		var resp oauth2server.AuthorizeResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.NotEmpty(t, resp.Code)
		assert.Equal(t, "vikunja-flutter://callback", resp.RedirectURI)
		assert.Equal(t, "teststate", resp.State)
	})

	t.Run("rejects invalid redirect_uri", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		token, err := auth.NewUserJWTAuthtoken(&testuser1, "test-session-id")
		require.NoError(t, err)

		body := authorizeRequestBody("code", "vikunja", "https://evil.com/callback", "test", "S256", "")
		rec := doAuthorize(e, token, body)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("rejects missing PKCE", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		token, err := auth.NewUserJWTAuthtoken(&testuser1, "test-session-id")
		require.NoError(t, err)

		body := authorizeRequestBody("code", "vikunja", "vikunja-flutter://callback", "", "", "")
		rec := doAuthorize(e, token, body)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

// getAuthorizationCode performs the authorize step and returns the code from the JSON response.
func getAuthorizationCode(t *testing.T, e http.Handler, codeChallenge, state string) string {
	t.Helper()

	token, err := auth.NewUserJWTAuthtoken(&testuser1, "test-session-id")
	require.NoError(t, err)

	body := authorizeRequestBody("code", "vikunja", "vikunja-flutter://callback", codeChallenge, "S256", state)
	rec := doAuthorize(e, token, body)
	require.Equal(t, http.StatusOK, rec.Code)

	var resp oauth2server.AuthorizeResponse
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	require.NotEmpty(t, resp.Code)

	return resp.Code
}

func TestOAuth2TokenEndpoint(t *testing.T) {
	t.Run("full authorization code flow with PKCE", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		codeVerifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
		h := sha256.Sum256([]byte(codeVerifier))
		codeChallenge := base64.RawURLEncoding.EncodeToString(h[:])

		code := getAuthorizationCode(t, e, codeChallenge, "xyz")

		rec := doTokenRequest(e, map[string]string{
			"grant_type":    "authorization_code",
			"code":          code,
			"client_id":     "vikunja",
			"redirect_uri":  "vikunja-flutter://callback",
			"code_verifier": codeVerifier,
		})

		require.Equal(t, http.StatusOK, rec.Code)

		var tokenResp oauth2server.TokenResponse
		err = json.Unmarshal(rec.Body.Bytes(), &tokenResp)
		require.NoError(t, err)
		assert.NotEmpty(t, tokenResp.AccessToken)
		assert.Equal(t, "bearer", tokenResp.TokenType)
		assert.NotEmpty(t, tokenResp.RefreshToken)
		assert.Positive(t, tokenResp.ExpiresIn)
	})

	t.Run("code is single-use", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		codeVerifier := "test-verifier-for-single-use-check"
		h := sha256.Sum256([]byte(codeVerifier))
		codeChallenge := base64.RawURLEncoding.EncodeToString(h[:])

		code := getAuthorizationCode(t, e, codeChallenge, "")

		tokenParams := map[string]string{
			"grant_type":    "authorization_code",
			"code":          code,
			"client_id":     "vikunja",
			"redirect_uri":  "vikunja-flutter://callback",
			"code_verifier": codeVerifier,
		}

		// First exchange succeeds
		rec := doTokenRequest(e, tokenParams)
		require.Equal(t, http.StatusOK, rec.Code)

		// Second exchange fails
		rec2 := doTokenRequest(e, tokenParams)
		assert.Equal(t, http.StatusBadRequest, rec2.Code)
	})

	t.Run("rejects wrong PKCE verifier", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		codeVerifier := "correct-verifier"
		h := sha256.Sum256([]byte(codeVerifier))
		codeChallenge := base64.RawURLEncoding.EncodeToString(h[:])

		code := getAuthorizationCode(t, e, codeChallenge, "")

		rec := doTokenRequest(e, map[string]string{
			"grant_type":    "authorization_code",
			"code":          code,
			"client_id":     "vikunja",
			"redirect_uri":  "vikunja-flutter://callback",
			"code_verifier": "wrong-verifier",
		})
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("rejects invalid grant_type", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		rec := doTokenRequest(e, map[string]string{
			"grant_type": "password",
			"client_id":  "vikunja",
		})
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("refresh token flow", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		codeVerifier := "refresh-flow-test-verifier"
		h := sha256.Sum256([]byte(codeVerifier))
		codeChallenge := base64.RawURLEncoding.EncodeToString(h[:])

		code := getAuthorizationCode(t, e, codeChallenge, "")

		rec := doTokenRequest(e, map[string]string{
			"grant_type":    "authorization_code",
			"code":          code,
			"client_id":     "vikunja",
			"redirect_uri":  "vikunja-flutter://callback",
			"code_verifier": codeVerifier,
		})
		require.Equal(t, http.StatusOK, rec.Code)

		var tokenResp oauth2server.TokenResponse
		_ = json.Unmarshal(rec.Body.Bytes(), &tokenResp)

		// Use the refresh token to get new tokens
		rec2 := doTokenRequest(e, map[string]string{
			"grant_type":    "refresh_token",
			"refresh_token": tokenResp.RefreshToken,
			"client_id":     "vikunja",
		})

		require.Equal(t, http.StatusOK, rec2.Code)

		var refreshResp oauth2server.TokenResponse
		err = json.Unmarshal(rec2.Body.Bytes(), &refreshResp)
		require.NoError(t, err)
		assert.NotEmpty(t, refreshResp.AccessToken)
		assert.NotEmpty(t, refreshResp.RefreshToken)
		assert.NotEqual(t, tokenResp.RefreshToken, refreshResp.RefreshToken)
	})

	t.Run("refresh token rotation prevents replay", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		codeVerifier := "replay-test-verifier"
		h := sha256.Sum256([]byte(codeVerifier))
		codeChallenge := base64.RawURLEncoding.EncodeToString(h[:])

		code := getAuthorizationCode(t, e, codeChallenge, "")

		rec := doTokenRequest(e, map[string]string{
			"grant_type":    "authorization_code",
			"code":          code,
			"client_id":     "vikunja",
			"redirect_uri":  "vikunja-flutter://callback",
			"code_verifier": codeVerifier,
		})

		var tokenResp oauth2server.TokenResponse
		_ = json.Unmarshal(rec.Body.Bytes(), &tokenResp)
		oldRefreshToken := tokenResp.RefreshToken

		refreshParams := map[string]string{
			"grant_type":    "refresh_token",
			"refresh_token": oldRefreshToken,
			"client_id":     "vikunja",
		}

		// First refresh succeeds
		rec2 := doTokenRequest(e, refreshParams)
		require.Equal(t, http.StatusOK, rec2.Code)

		// Replay the same old refresh token — should fail
		rec3 := doTokenRequest(e, refreshParams)
		assert.Equal(t, http.StatusUnauthorized, rec3.Code)
	})
}
