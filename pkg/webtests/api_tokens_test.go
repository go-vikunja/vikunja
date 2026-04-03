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
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/modules/auth/oauth2server"
	"code.vikunja.io/api/pkg/routes"
	"code.vikunja.io/api/pkg/user"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPITokenRoutesIncludesCaldav(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)

	s := db.NewSession()
	defer s.Close()
	u, err := user.GetUserByID(s, 1)
	require.NoError(t, err)
	jwt, err := auth.NewUserJWTAuthtoken(u, "test-session-id")
	require.NoError(t, err)

	ctx := context.Background()
	req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/api/v1/routes", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+jwt)
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Contains(t, res.Body.String(), `"caldav"`)
	assert.Contains(t, res.Body.String(), `"access"`)
}

func TestAPIToken(t *testing.T) {
	t.Run("valid token", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/tasks", nil)
		res := httptest.NewRecorder()
		c := e.NewContext(req, res)
		h := routes.SetupTokenMiddleware()(func(c *echo.Context) error {
			u, err := auth.GetAuthFromClaims(c)
			if err != nil {
				return c.String(http.StatusInternalServerError, err.Error())
			}

			return c.JSON(http.StatusOK, u)
		})

		req.Header.Set(echo.HeaderAuthorization, "Bearer tk_2eef46f40ebab3304919ab2e7e39993f75f29d2e") // Token 1
		require.NoError(t, h(c))
		// check if the request handlers "see" the request as if it came directly from that user
		assert.Contains(t, res.Body.String(), `"username":"user1"`)
	})
	t.Run("invalid token", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/tasks", nil)
		res := httptest.NewRecorder()
		c := e.NewContext(req, res)
		h := routes.SetupTokenMiddleware()(func(c *echo.Context) error {
			return c.String(http.StatusOK, "test")
		})

		req.Header.Set(echo.HeaderAuthorization, "Bearer tk_loremipsumdolorsitamet")
		require.NoError(t, h(c))
		assert.Equal(t, http.StatusUnauthorized, res.Code)
		assert.Contains(t, res.Body.String(), `"code":11`)
	})
	t.Run("expired token", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/tasks", nil)
		res := httptest.NewRecorder()
		c := e.NewContext(req, res)
		h := routes.SetupTokenMiddleware()(func(c *echo.Context) error {
			return c.String(http.StatusOK, "test")
		})

		req.Header.Set(echo.HeaderAuthorization, "Bearer tk_a5e6f92ddbad68f49ee2c63e52174db0235008c8") // Token 2
		require.NoError(t, h(c))
		assert.Equal(t, http.StatusUnauthorized, res.Code)
		assert.Contains(t, res.Body.String(), `"code":11`)
	})
	t.Run("valid token, invalid scope", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/projects", nil)
		res := httptest.NewRecorder()
		c := e.NewContext(req, res)
		h := routes.SetupTokenMiddleware()(func(c *echo.Context) error {
			return c.String(http.StatusOK, "test")
		})

		req.Header.Set(echo.HeaderAuthorization, "Bearer tk_2eef46f40ebab3304919ab2e7e39993f75f29d2e")
		require.NoError(t, h(c))
		assert.Equal(t, http.StatusUnauthorized, res.Code)
		assert.Contains(t, res.Body.String(), `"code":11`)
	})
	t.Run("disabled user token rejected", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/tasks", nil)
		res := httptest.NewRecorder()
		c := e.NewContext(req, res)
		h := routes.SetupTokenMiddleware()(func(c *echo.Context) error {
			return c.String(http.StatusOK, "test")
		})

		req.Header.Set(echo.HeaderAuthorization, "Bearer tk_disabled_user_test_token_000000001234abcd") // Token 4 (disabled user 17)
		require.NoError(t, h(c))
		assert.Equal(t, http.StatusUnauthorized, res.Code)
		assert.Contains(t, res.Body.String(), `"code":11`)
	})
	t.Run("locked user token rejected", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/tasks", nil)
		res := httptest.NewRecorder()
		c := e.NewContext(req, res)
		h := routes.SetupTokenMiddleware()(func(c *echo.Context) error {
			return c.String(http.StatusOK, "test")
		})

		req.Header.Set(echo.HeaderAuthorization, "Bearer tk_locked_user_test_token_0000000012345678") // Token 5 (locked user 18)
		require.NoError(t, h(c))
		assert.Equal(t, http.StatusUnauthorized, res.Code)
		assert.Contains(t, res.Body.String(), `"code":11`)
	})
	t.Run("jwt", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/tasks", nil)
		res := httptest.NewRecorder()
		c := e.NewContext(req, res)
		h := routes.SetupTokenMiddleware()(func(c *echo.Context) error {
			return c.String(http.StatusOK, "test")
		})

		s := db.NewSession()
		defer s.Close()
		u, err := user.GetUserByID(s, 1)
		require.NoError(t, err)
		jwt, err := auth.NewUserJWTAuthtoken(u, "test-session-id")
		require.NoError(t, err)

		req.Header.Set(echo.HeaderAuthorization, "Bearer "+jwt)
		require.NoError(t, h(c))
	})
}

func TestOAuth2ClientRegistration(t *testing.T) {
	t.Run("successful registration and token flow", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		body, _ := json.Marshal(map[string]interface{}{
			"client_name":   "Test App",
			"redirect_uris": []string{"https://example.com/callback"},
		})

		req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/auth/openid/register", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		res := httptest.NewRecorder()
		e.ServeHTTP(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		var resp oauth2server.DynamicClientRegistrationResponse
		err = json.Unmarshal(res.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.NotEmpty(t, resp.ClientID)
		assert.Equal(t, "Test App", resp.ClientName)
		assert.Contains(t, resp.GrantTypes, "authorization_code")
		assert.Contains(t, resp.GrantTypes, "refresh_token")

		token, err := auth.NewUserJWTAuthtoken(&testuser1, "test-session-id")
		require.NoError(t, err)

		codeVerifier := "test-verifier-12345"
		h := sha256.Sum256([]byte(codeVerifier))
		codeChallenge := base64.RawURLEncoding.EncodeToString(h[:])

		authBody, _ := json.Marshal(map[string]string{
			"response_type":         "code",
			"client_id":             resp.ClientID,
			"redirect_uri":          "https://example.com/callback",
			"code_challenge":        codeChallenge,
			"code_challenge_method": "S256",
			"state":                 "teststate",
		})

		authReq := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/oauth/authorize", bytes.NewReader(authBody))
		authReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		authReq.Header.Set("Authorization", "Bearer "+token)
		authRes := httptest.NewRecorder()
		e.ServeHTTP(authRes, authReq)

		require.Equal(t, http.StatusOK, authRes.Code)

		var authResp oauth2server.AuthorizeResponse
		err = json.Unmarshal(authRes.Body.Bytes(), &authResp)
		require.NoError(t, err)
		assert.NotEmpty(t, authResp.Code)

		tokenBody, _ := json.Marshal(map[string]string{
			"grant_type":    "authorization_code",
			"code":          authResp.Code,
			"client_id":     resp.ClientID,
			"redirect_uri":  "https://example.com/callback",
			"code_verifier": codeVerifier,
		})

		tokenReq := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/oauth/token", bytes.NewReader(tokenBody))
		tokenReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		tokenRes := httptest.NewRecorder()
		e.ServeHTTP(tokenRes, tokenReq)

		require.Equal(t, http.StatusOK, tokenRes.Code)

		var tokenResp oauth2server.TokenResponse
		err = json.Unmarshal(tokenRes.Body.Bytes(), &tokenResp)
		require.NoError(t, err)
		assert.NotEmpty(t, tokenResp.AccessToken)
		assert.NotEmpty(t, tokenResp.RefreshToken)
		assert.Equal(t, "bearer", tokenResp.TokenType)
	})

	t.Run("registration with multiple redirect uris", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		body, _ := json.Marshal(map[string]interface{}{
			"client_name":   "Multi Redirect App",
			"redirect_uris": []string{"https://example.com/callback", "http://localhost:8080/callback"},
		})

		req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/auth/openid/register", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		res := httptest.NewRecorder()
		e.ServeHTTP(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		var resp oauth2server.DynamicClientRegistrationResponse
		err = json.Unmarshal(res.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.NotEmpty(t, resp.ClientID)
	})

	t.Run("registration without client name fails", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		body, _ := json.Marshal(map[string]interface{}{
			"redirect_uris": []string{"https://example.com/callback"},
		})

		req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/auth/openid/register", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		res := httptest.NewRecorder()
		e.ServeHTTP(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)
		assert.Contains(t, res.Body.String(), "client_name")
	})

	t.Run("registration without redirect uris fails", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		body, _ := json.Marshal(map[string]interface{}{
			"client_name": "Test App No Redirect",
		})

		req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/auth/openid/register", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		res := httptest.NewRecorder()
		e.ServeHTTP(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)
		assert.Contains(t, res.Body.String(), "redirect_uris")
	})
}
