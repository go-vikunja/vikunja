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
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth/oauth2server"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// insertOAuthCode stores a code like Authorize would, for states Authorize never creates.
func insertOAuthCode(t *testing.T, c models.OAuthCode, expiresAt time.Time) string {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	if c.UserID == 0 {
		c.UserID = testuser1.ID
	}
	if c.ClientID == "" {
		c.ClientID = scopedClientID
	}
	c.RedirectURI = scopedRedirectURI
	c.CodeChallenge = pkceChallenge(scopedVerifier)
	c.CodeChallengeMethod = "S256"
	code, err := models.CreateOAuthCode(s, &c)
	require.NoError(t, err)
	if !expiresAt.IsZero() {
		_, err = s.Where("code = ?", models.HashSessionToken(code)).Cols("expires_at").Update(&models.OAuthCode{ExpiresAt: expiresAt})
		require.NoError(t, err)
	}
	require.NoError(t, s.Commit())
	return code
}

func TestOAuthTokenErrors(t *testing.T) {
	e, metadataClientID, host := metadataClientEnv(t)
	const mcpScope = "mcp:read"

	usedCode := insertOAuthCode(t, models.OAuthCode{}, time.Time{})
	require.Equal(t, http.StatusOK, scopedExchange(t, e, "v2", usedCode, nil).code)

	for _, tc := range []struct {
		name string
		// form replaces the defaults of scopedExchange; code is the code to exchange.
		code      func() string
		form      url.Values
		error     oauth2server.TokenErrorCode
		vikunja   int
		v1Status  int
		v1Vikunja int
	}{
		{
			name:    "unknown code",
			code:    func() string { return "not-a-code" },
			error:   oauth2server.TokenErrorInvalidGrant,
			vikunja: models.ErrCodeOAuthCodeInvalid,
		},
		{
			name:    "used code",
			code:    func() string { return usedCode },
			error:   oauth2server.TokenErrorInvalidGrant,
			vikunja: models.ErrCodeOAuthCodeInvalid,
		},
		{
			name: "expired code",
			code: func() string {
				return insertOAuthCode(t, models.OAuthCode{}, time.Now().Add(-time.Minute))
			},
			error:   oauth2server.TokenErrorInvalidGrant,
			vikunja: models.ErrCodeOAuthCodeExpired,
		},
		{
			name:    "PKCE mismatch",
			form:    url.Values{"code_verifier": {"another-verifier-0123456789abcdef0123456789"}},
			error:   oauth2server.TokenErrorInvalidGrant,
			vikunja: models.ErrCodeOAuthPKCEVerifyFailed,
		},
		{
			name:    "redirect mismatch",
			form:    url.Values{"redirect_uri": {"http://127.0.0.1:1/other"}},
			error:   oauth2server.TokenErrorInvalidGrant,
			vikunja: models.ErrCodeOAuthInvalidRedirectURI,
		},
		{
			name:    "client mismatch",
			form:    url.Values{"client_id": {"another-client"}},
			error:   oauth2server.TokenErrorInvalidGrant,
			vikunja: models.ErrCodeOAuthClientNotFound,
		},
		{
			name:    "unsupported grant type",
			form:    url.Values{"grant_type": {"password"}},
			error:   oauth2server.TokenErrorUnsupportedGrantType,
			vikunja: models.ErrCodeOAuthInvalidGrantType,
		},
		{
			name:    "missing grant type",
			form:    url.Values{"grant_type": {""}},
			error:   oauth2server.TokenErrorInvalidRequest,
			vikunja: models.ErrCodeOAuthInvalidGrantType,
		},
		{
			name: "foreign resource",
			code: func() string {
				return insertOAuthCode(t, models.OAuthCode{Scope: mcpScope, Resource: mcpResourceURL}, time.Time{})
			},
			form:    url.Values{"resource": {"https://evil.example/api/v2/mcp"}},
			error:   oauth2server.TokenErrorInvalidTarget,
			vikunja: models.ErrCodeOAuthInvalidTarget,
		},
		{
			name: "unscoped code of a metadata document client",
			code: func() string {
				return insertOAuthCode(t, models.OAuthCode{ClientID: metadataClientID}, time.Time{})
			},
			form:    url.Values{"client_id": {metadataClientID}},
			error:   oauth2server.TokenErrorInvalidScope,
			vikunja: models.ErrCodeOAuthInvalidScope,
		},
		{
			name: "unreachable metadata document",
			code: func() string {
				return insertOAuthCode(t, models.OAuthCode{ClientID: host.URL + "/missing.json", Scope: mcpScope, Resource: mcpResourceURL}, time.Time{})
			},
			form:      url.Values{"client_id": {host.URL + "/missing.json"}},
			error:     oauth2server.TokenErrorInvalidClient,
			vikunja:   models.ErrCodeOAuthClientMetadataUnavailable,
			v1Status:  http.StatusBadGateway,
			v1Vikunja: models.ErrCodeOAuthClientMetadataUnavailable,
		},
		{
			name: "invalid metadata document",
			code: func() string {
				return insertOAuthCode(t, models.OAuthCode{ClientID: host.URL + "/broken.json", Scope: mcpScope, Resource: mcpResourceURL}, time.Time{})
			},
			form:    url.Values{"client_id": {host.URL + "/broken.json"}},
			error:   oauth2server.TokenErrorInvalidClient,
			vikunja: models.ErrCodeOAuthInvalidClientMetadata,
		},
		{
			name: "disabled user",
			code: func() string {
				return insertOAuthCode(t, models.OAuthCode{UserID: 17}, time.Time{})
			},
			error: oauth2server.TokenErrorInvalidGrant,
		},
		{
			name: "invalid refresh token",
			form: url.Values{
				"grant_type":    {"refresh_token"},
				"refresh_token": {"not-a-refresh-token"},
			},
			error:     oauth2server.TokenErrorInvalidGrant,
			vikunja:   models.ErrCodeInvalidRefreshToken,
			v1Status:  http.StatusUnauthorized,
			v1Vikunja: models.ErrCodeInvalidRefreshToken,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			code := func() string {
				if tc.code != nil {
					return tc.code()
				}
				return insertOAuthCode(t, models.OAuthCode{}, time.Time{})
			}

			res := scopedExchange(t, e, "v2", code(), tc.form)
			require.Equal(t, http.StatusBadRequest, res.code, string(res.body))
			var body struct {
				Error            oauth2server.TokenErrorCode `json:"error"`
				ErrorDescription string                      `json:"error_description"`
				Detail           string                      `json:"detail"`
				Code             int                         `json:"code"`
			}
			require.NoError(t, json.Unmarshal(res.body, &body), string(res.body))
			assert.Equal(t, tc.error, body.Error)
			assert.NotEmpty(t, body.ErrorDescription)
			assert.Equal(t, body.Detail, body.ErrorDescription)
			if tc.vikunja != 0 {
				assert.Equal(t, tc.vikunja, body.Code)
			}

			if tc.v1Status != 0 {
				v1 := scopedExchange(t, e, "v1", code(), tc.form)
				assert.Equal(t, tc.v1Status, v1.code, string(v1.body))
				assert.Equal(t, tc.v1Vikunja, v1.errorCode(t))
				assert.NotContains(t, string(v1.body), `"error"`, "v1 is frozen")
			}
		})
	}
}

// RFC 6749 §3.2: the token endpoint MUST ignore unrecognized request parameters.
func TestOAuthTokenIgnoresUnknownParameters(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)

	code := insertOAuthCode(t, models.OAuthCode{Scope: "mcp:read", Resource: mcpResourceURL}, time.Time{})
	res := scopedExchange(t, e, "v2", code, url.Values{"scope": {"mcp:read"}, "state": {"xyz"}})
	assert.Equal(t, http.StatusOK, res.code, string(res.body))
}

func TestOAuthTokenCORS(t *testing.T) {
	e := discoveryEnv(t, map[config.Key]any{
		config.CorsEnable:  true,
		config.CorsOrigins: []string{"https://allowed.example"},
	})
	const origin = "https://mcp-client.example"

	req := httptest.NewRequest(http.MethodOptions, "/api/v2/oauth/token", nil)
	req.Header.Set("Origin", origin)
	req.Header.Set(echo.HeaderAccessControlRequestMethod, http.MethodPost)
	req.Header.Set(echo.HeaderAccessControlRequestHeaders, "content-type")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, "*", rec.Header().Get(echo.HeaderAccessControlAllowOrigin))
	assert.Contains(t, rec.Header().Get(echo.HeaderAccessControlAllowMethods), http.MethodPost)
	assert.Equal(t, "content-type", rec.Header().Get(echo.HeaderAccessControlAllowHeaders))
	assert.Empty(t, rec.Header().Get(echo.HeaderAccessControlAllowCredentials))

	for name, code := range map[string]string{
		"success":  insertOAuthCode(t, models.OAuthCode{Scope: "mcp:read", Resource: mcpResourceURL}, time.Time{}),
		"rejected": "not-a-code",
	} {
		form := url.Values{
			"grant_type":    {"authorization_code"},
			"code":          {code},
			"client_id":     {scopedClientID},
			"redirect_uri":  {scopedRedirectURI},
			"code_verifier": {scopedVerifier},
		}
		for _, o := range []string{origin, "https://allowed.example"} {
			req := httptest.NewRequest(http.MethodPost, "/api/v2/oauth/token", strings.NewReader(form.Encode()))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
			req.Header.Set("Origin", o)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			assert.Equal(t, "*", rec.Header().Get(echo.HeaderAccessControlAllowOrigin), "%s from %s: %d %s", name, o, rec.Code, rec.Body.String())
			assert.Empty(t, rec.Header().Get(echo.HeaderAccessControlAllowCredentials), "%s from %s", name, o)
			form.Set("code", "not-a-code")
		}
	}

	// Only the token endpoint: the rest of the API keeps cors.origins.
	req = httptest.NewRequest(http.MethodGet, "/api/v2/info", nil)
	req.Header.Set("Origin", origin)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Empty(t, rec.Header().Get(echo.HeaderAccessControlAllowOrigin))
}

// RFC 6749 §2.3.1: the client_id may come as the HTTP Basic username, form-urlencoded.
func TestOAuthTokenBasicClientAuth(t *testing.T) {
	e, metadataClientID, host := metadataClientEnv(t)
	require.NotEqual(t, metadataClientID, url.QueryEscape(metadataClientID), "the client id must need form-encoding")

	basic := func(clientID, secret string) string {
		return "Basic " + base64.StdEncoding.EncodeToString([]byte(url.QueryEscape(clientID)+":"+url.QueryEscape(secret)))
	}
	exchange := func(code, authorization string, extra url.Values) *httptest.ResponseRecorder {
		form := url.Values{
			"grant_type":    {"authorization_code"},
			"code":          {code},
			"redirect_uri":  {scopedRedirectURI},
			"code_verifier": {scopedVerifier},
		}
		for k, v := range extra {
			form[k] = v
		}
		req := httptest.NewRequest(http.MethodPost, "/api/v2/oauth/token", strings.NewReader(form.Encode()))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
		if authorization != "" {
			req.Header.Set(echo.HeaderAuthorization, authorization)
		}
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		return rec
	}
	scopedCode := func(clientID string) string {
		return insertOAuthCode(t, models.OAuthCode{ClientID: clientID, Scope: "mcp:read", Resource: mcpResourceURL}, time.Time{})
	}

	t.Run("accepted", func(t *testing.T) {
		for _, tc := range []struct {
			name          string
			code          string
			authorization string
			form          url.Values
		}{
			{"metadata document client in Basic only", scopedCode(metadataClientID), basic(metadataClientID, ""), nil},
			{"legacy client in Basic only", insertOAuthCode(t, models.OAuthCode{}, time.Time{}), basic(scopedClientID, ""), nil},
			{"Basic and body agree", scopedCode(metadataClientID), basic(metadataClientID, ""), url.Values{"client_id": {metadataClientID}}},
			{"no Authorization header", scopedCode(scopedClientID), "", url.Values{"client_id": {scopedClientID}}},
		} {
			t.Run(tc.name, func(t *testing.T) {
				rec := exchange(tc.code, tc.authorization, tc.form)
				require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
				var token oauth2server.TokenResponse
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &token))
				assert.NotEmpty(t, token.AccessToken)
			})
		}
	})

	t.Run("rejected without consuming the code", func(t *testing.T) {
		for _, tc := range []struct {
			name          string
			authorization string
			form          url.Values
			status        int
			error         oauth2server.TokenErrorCode
			vikunja       int
		}{
			{
				name:          "Basic and body disagree",
				authorization: basic(scopedClientID, ""),
				form:          url.Values{"client_id": {"another-client"}},
				status:        http.StatusBadRequest,
				error:         oauth2server.TokenErrorInvalidRequest,
				vikunja:       models.ErrCodeOAuthMalformedClientCredentials,
			},
			{
				name:          "malformed Basic credentials",
				authorization: "Basic not-base64!",
				status:        http.StatusBadRequest,
				error:         oauth2server.TokenErrorInvalidRequest,
				vikunja:       models.ErrCodeOAuthMalformedClientCredentials,
			},
			{
				name:          "client secret",
				authorization: basic(scopedClientID, "s3cret"),
				status:        http.StatusUnauthorized,
				error:         oauth2server.TokenErrorInvalidClient,
				vikunja:       models.ErrCodeOAuthClientSecretUnsupported,
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				code := scopedCode(scopedClientID)
				rec := exchange(code, tc.authorization, tc.form)
				require.Equal(t, tc.status, rec.Code, rec.Body.String())
				var body struct {
					Error oauth2server.TokenErrorCode `json:"error"`
					Code  int                         `json:"code"`
				}
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body), rec.Body.String())
				assert.Equal(t, tc.error, body.Error)
				assert.Equal(t, tc.vikunja, body.Code)
				if tc.status == http.StatusUnauthorized {
					assert.True(t, strings.HasPrefix(rec.Header().Get(echo.HeaderWWWAuthenticate), "Basic "), rec.Header().Get(echo.HeaderWWWAuthenticate))
				} else {
					assert.Empty(t, rec.Header().Get(echo.HeaderWWWAuthenticate))
				}

				retry := exchange(code, basic(scopedClientID, ""), nil)
				assert.Equal(t, http.StatusOK, retry.Code, retry.Body.String())
			})
		}
	})

	// RFC 6749 §5.2: invalid_client after Basic authentication is a 401 with a Basic challenge.
	t.Run("invalid client after Basic", func(t *testing.T) {
		clientID := host.URL + "/missing.json"
		rec := exchange(scopedCode(clientID), basic(clientID, ""), nil)
		require.Equal(t, http.StatusUnauthorized, rec.Code, rec.Body.String())
		assert.True(t, strings.HasPrefix(rec.Header().Get(echo.HeaderWWWAuthenticate), "Basic "))
		assert.Contains(t, rec.Body.String(), `"error":"invalid_client"`)
	})
}
