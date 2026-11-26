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
	"encoding/json"
	"net/http"
	"net/url"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth/oauth2server"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	scopedClientID    = "test-mcp-client"
	scopedRedirectURI = "http://127.0.0.1:33418/callback"
	scopedVerifier    = "scoped-grant-test-verifier-0123456789abcdef"
)

func scopedAuthorize(t *testing.T, e *echo.Echo, version string, params map[string]string) *httpResult {
	t.Helper()
	body := map[string]string{
		"response_type":         "code",
		"client_id":             scopedClientID,
		"redirect_uri":          scopedRedirectURI,
		"code_challenge":        pkceChallenge(scopedVerifier),
		"code_challenge_method": "S256",
	}
	for k, v := range params {
		body[k] = v
	}
	raw, err := json.Marshal(body)
	require.NoError(t, err)
	rec := humaRequest(t, e, http.MethodPost, "/api/"+version+"/oauth/authorize", string(raw), humaTokenFor(t, &testuser1), "")
	return &httpResult{
		code: rec.Code,
		body: rec.Body.Bytes(),
	}
}

func scopedExchange(t *testing.T, e *echo.Echo, version, code string, extra url.Values) *httpResult {
	t.Helper()
	form := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"client_id":     {scopedClientID},
		"redirect_uri":  {scopedRedirectURI},
		"code_verifier": {scopedVerifier},
	}
	for k, v := range extra {
		form[k] = v
	}
	rec := humaRequest(t, e, http.MethodPost, "/api/"+version+"/oauth/token", form.Encode(), "", "application/x-www-form-urlencoded")
	return &httpResult{
		code: rec.Code,
		body: rec.Body.Bytes(),
	}
}

type httpResult struct {
	code int
	body []byte
}

func (r *httpResult) errorCode(t *testing.T) int {
	t.Helper()
	var body struct {
		Code int `json:"code"`
	}
	require.NoError(t, json.Unmarshal(r.body, &body), string(r.body))
	return body.Code
}

func scopedGrant(t *testing.T, e *echo.Echo, params map[string]string) (oauth2server.AuthorizeResponse, oauth2server.TokenResponse, map[string]any) {
	t.Helper()
	authz := scopedAuthorize(t, e, "v2", params)
	require.Equal(t, http.StatusOK, authz.code, string(authz.body))
	var authorized oauth2server.AuthorizeResponse
	require.NoError(t, json.Unmarshal(authz.body, &authorized))

	tok := scopedExchange(t, e, "v2", authorized.Code, nil)
	require.Equal(t, http.StatusOK, tok.code, string(tok.body))
	var token oauth2server.TokenResponse
	require.NoError(t, json.Unmarshal(tok.body, &token))
	var raw map[string]any
	require.NoError(t, json.Unmarshal(tok.body, &raw))
	return authorized, token, raw
}

func storedToken(t *testing.T, cleartext string) *models.APIToken {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	token, err := models.GetTokenFromTokenString(s, cleartext)
	require.NoError(t, err)
	return token
}

func hasNonReadPermission(permissions models.APIPermissions) bool {
	for group, granted := range permissions {
		for _, p := range granted {
			if group != "mcp" && p != "read_one" && p != "read_all" {
				return true
			}
		}
	}
	return false
}

func TestOAuthScopedGrant(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)

	t.Run("mcp:read issues a read-only MCP token without refresh token", func(t *testing.T) {
		authorized, token, raw := scopedGrant(t, e, map[string]string{
			"scope": "mcp:read offline_access",
		})
		assert.Equal(t, "mcp:read", authorized.Scope)

		assert.Regexp(t, `^tk_`, token.AccessToken)
		assert.Equal(t, "bearer", token.TokenType)
		assert.Equal(t, int64(90*24*60*60), token.ExpiresIn)
		assert.Equal(t, "mcp:read", token.Scope)
		assert.NotContains(t, raw, "refresh_token")

		stored := storedToken(t, token.AccessToken)
		assert.Equal(t, scopedClientID, stored.OAuthClientID)
		assert.Equal(t, mcpResourceURL, stored.OAuthResource)
		assert.Equal(t, int64(1), stored.OwnerID)
		assert.True(t, stored.HasMCPAccess())
		assert.True(t, stored.HasPermission("projects", "read_all"))
		assert.True(t, stored.HasPermission("tasks", "read_one"))
		assert.False(t, hasNonReadPermission(stored.APIPermissions), "%v", stored.APIPermissions)
	})

	t.Run("mcp:write implies mcp:read and grants write permissions", func(t *testing.T) {
		authorized, token, _ := scopedGrant(t, e, map[string]string{
			"scope": "mcp:write",
		})
		assert.Equal(t, "mcp:read mcp:write", authorized.Scope)
		assert.Equal(t, "mcp:read mcp:write", token.Scope)

		stored := storedToken(t, token.AccessToken)
		assert.True(t, stored.HasMCPAccess())
		assert.True(t, stored.HasPermission("projects", "read_all"))
		assert.True(t, stored.HasPermission("tasks", "create"))
		assert.True(t, stored.HasPermission("tasks", "delete"))
	})

	t.Run("resource without scope grants read and write", func(t *testing.T) {
		authorized, token, _ := scopedGrant(t, e, map[string]string{
			"resource": "HTTPS://LOCALHOST/api/v2/mcp/",
		})
		assert.Equal(t, "mcp:read mcp:write", authorized.Scope)
		assert.Equal(t, mcpResourceURL, storedToken(t, token.AccessToken).OAuthResource)
	})

	t.Run("token request may repeat the resource", func(t *testing.T) {
		authz := scopedAuthorize(t, e, "v2", map[string]string{
			"scope":    "mcp:read",
			"resource": mcpResourceURL,
		})
		require.Equal(t, http.StatusOK, authz.code, string(authz.body))
		var authorized oauth2server.AuthorizeResponse
		require.NoError(t, json.Unmarshal(authz.body, &authorized))

		tok := scopedExchange(t, e, "v2", authorized.Code, url.Values{"resource": {mcpResourceURL + "/"}})
		assert.Equal(t, http.StatusOK, tok.code, string(tok.body))
	})

	t.Run("rejects a foreign resource", func(t *testing.T) {
		authz := scopedAuthorize(t, e, "v2", map[string]string{
			"scope":    "mcp:read",
			"resource": "https://evil.example/api/v2/mcp",
		})
		require.Equal(t, http.StatusBadRequest, authz.code, string(authz.body))
		assert.Equal(t, models.ErrCodeOAuthInvalidTarget, authz.errorCode(t))
	})

	t.Run("a scope without mcp values and no resource returns a session", func(t *testing.T) {
		authz := scopedAuthorize(t, e, "v2", map[string]string{
			"scope": "offline_access openid",
		})
		require.Equal(t, http.StatusOK, authz.code, string(authz.body))
		var authorized oauth2server.AuthorizeResponse
		require.NoError(t, json.Unmarshal(authz.body, &authorized))
		assert.Empty(t, authorized.Scope)

		tok := scopedExchange(t, e, "v2", authorized.Code, nil)
		require.Equal(t, http.StatusOK, tok.code, string(tok.body))
		var raw map[string]any
		require.NoError(t, json.Unmarshal(tok.body, &raw))
		assert.NotEmpty(t, raw["refresh_token"])
		assert.NotRegexp(t, `^tk_`, raw["access_token"])
	})

	t.Run("rejects a resource with a scope without mcp values", func(t *testing.T) {
		authz := scopedAuthorize(t, e, "v2", map[string]string{
			"scope":    "offline_access openid",
			"resource": mcpResourceURL,
		})
		require.Equal(t, http.StatusBadRequest, authz.code, string(authz.body))
		assert.Equal(t, models.ErrCodeOAuthInvalidScope, authz.errorCode(t))
	})

	t.Run("rejects a scoped grant without client id", func(t *testing.T) {
		authz := scopedAuthorize(t, e, "v2", map[string]string{
			"scope":     "mcp:read",
			"client_id": "",
		})
		require.Equal(t, http.StatusBadRequest, authz.code, string(authz.body))
		assert.Equal(t, models.ErrCodeOAuthClientNotFound, authz.errorCode(t))
	})

	t.Run("rejects a mismatched resource at the token endpoint and burns the code", func(t *testing.T) {
		authz := scopedAuthorize(t, e, "v2", map[string]string{
			"scope": "mcp:read",
		})
		require.Equal(t, http.StatusOK, authz.code, string(authz.body))
		var authorized oauth2server.AuthorizeResponse
		require.NoError(t, json.Unmarshal(authz.body, &authorized))

		tok := scopedExchange(t, e, "v2", authorized.Code, url.Values{"resource": {"https://evil.example/api/v2/mcp"}})
		require.Equal(t, http.StatusBadRequest, tok.code, string(tok.body))
		assert.Equal(t, models.ErrCodeOAuthInvalidTarget, tok.errorCode(t))

		retry := scopedExchange(t, e, "v2", authorized.Code, nil)
		require.Equal(t, http.StatusBadRequest, retry.code, string(retry.body))
		assert.Equal(t, models.ErrCodeOAuthCodeInvalid, retry.errorCode(t))
	})

	t.Run("an unscoped code never yields a session when a resource is sent", func(t *testing.T) {
		authz := scopedAuthorize(t, e, "v2", nil)
		require.Equal(t, http.StatusOK, authz.code, string(authz.body))
		var authorized oauth2server.AuthorizeResponse
		require.NoError(t, json.Unmarshal(authz.body, &authorized))

		tok := scopedExchange(t, e, "v2", authorized.Code, url.Values{"resource": {mcpResourceURL}})
		require.Equal(t, http.StatusBadRequest, tok.code, string(tok.body))
		assert.Equal(t, models.ErrCodeOAuthInvalidTarget, tok.errorCode(t))
	})

	t.Run("unscoped flow still returns a session with refresh token", func(t *testing.T) {
		authz := scopedAuthorize(t, e, "v2", nil)
		require.Equal(t, http.StatusOK, authz.code, string(authz.body))
		var authorized oauth2server.AuthorizeResponse
		require.NoError(t, json.Unmarshal(authz.body, &authorized))
		assert.Empty(t, authorized.Scope)
		assert.NotContains(t, string(authz.body), `"scope"`)

		tok := scopedExchange(t, e, "v2", authorized.Code, nil)
		require.Equal(t, http.StatusOK, tok.code, string(tok.body))
		var raw map[string]any
		require.NoError(t, json.Unmarshal(tok.body, &raw))
		assert.NotEmpty(t, raw["refresh_token"])
		assert.NotContains(t, raw, "scope")
		assert.NotRegexp(t, `^tk_`, raw["access_token"])
	})

	t.Run("v1 endpoints share the scoped flow", func(t *testing.T) {
		authz := scopedAuthorize(t, e, "v1", map[string]string{
			"scope": "mcp:read",
		})
		require.Equal(t, http.StatusOK, authz.code, string(authz.body))
		var authorized oauth2server.AuthorizeResponse
		require.NoError(t, json.Unmarshal(authz.body, &authorized))
		assert.Equal(t, "mcp:read", authorized.Scope)

		tok := scopedExchange(t, e, "v1", authorized.Code, nil)
		require.Equal(t, http.StatusOK, tok.code, string(tok.body))
		var token oauth2server.TokenResponse
		require.NoError(t, json.Unmarshal(tok.body, &token))
		assert.Regexp(t, `^tk_`, token.AccessToken)
		assert.Empty(t, token.RefreshToken)
	})
}

func TestOAuthScopedGrantOverMCP(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)

	_, readToken, _ := scopedGrant(t, e, map[string]string{"scope": "mcp:read"})
	_, writeToken, _ := scopedGrant(t, e, map[string]string{"scope": "mcp:write"})

	t.Run("mcp:read sees and calls only read tools", func(t *testing.T) {
		c := newMCPClientOn(t, e, readToken.AccessToken)
		names := c.toolNames()
		assert.True(t, names["projects_list"])
		assert.True(t, names["tasks_read"])
		assert.False(t, names["projects_create"])
		assert.False(t, names["tasks_create"])
		assert.False(t, names["tasks_delete"])

		var projects []map[string]any
		readAllItems(t, c.callTool("projects_list", map[string]any{}), &projects)
		assert.NotEmpty(t, projects)

		denied := c.rpc("tools/call", map[string]any{
			"name": "tasks_create",
			"arguments": map[string]any{
				"project": 1,
				"title":   "nope",
			},
		})
		assert.NotContains(t, denied, "result")
		assert.Equal(t, map[string]any{
			"code":    float64(-32602),
			"message": `unknown tool "tasks_create"`,
		}, denied["error"])
		db.AssertMissing(t, "tasks", map[string]interface{}{
			"title": "nope",
		})
	})

	t.Run("mcp:write sees and calls write tools", func(t *testing.T) {
		c := newMCPClientOn(t, e, writeToken.AccessToken)
		names := c.toolNames()
		assert.True(t, names["projects_list"])
		assert.True(t, names["tasks_create"])
		assert.True(t, names["tasks_update"])

		var task map[string]any
		toolResultJSON(t, c.callTool("tasks_create", map[string]any{
			"project": 1,
			"title":   "created over oauth",
		}), &task)
		assert.Equal(t, "created over oauth", task["title"])
		require.NotContains(t, c.callTool("tasks_update", map[string]any{
			"task": task["id"],
			"done": true,
		}), "isError")
		db.AssertExists(t, "tasks", map[string]interface{}{
			"id":   task["id"],
			"done": true,
		}, false)
	})
}

func TestAPITokenCreateCannotSetOAuthFields(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	const fields = `"title":"forged","permissions":{"projects":["read_all"]},"expires_at":"2099-01-01T00:00:00Z","oauth_client_id":"forged-client"`

	for _, tc := range []struct {
		method string
		path   string
		body   string
	}{
		{
			method: http.MethodPost,
			path:   "/api/v2/tokens",
			body:   "{" + fields + "}",
		},
		{
			method: http.MethodPut,
			path:   "/api/v1/tokens",
			body:   "{" + fields + `,"oauth_resource":"https://localhost/api/v2/mcp"}`,
		},
	} {
		t.Run(tc.path, func(t *testing.T) {
			rec := humaRequest(t, e, tc.method, tc.path, tc.body, humaTokenFor(t, &testuser1), "")
			require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
			var created models.APIToken
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created))
			assert.NotContains(t, rec.Body.String(), "oauth_client_id")

			stored := storedToken(t, created.Token)
			assert.False(t, stored.IsOAuthGrant())
			assert.Empty(t, stored.OAuthResource)

			list := humaRequest(t, e, http.MethodGet, "/api/v2/projects", "", created.Token, "")
			assert.Equal(t, http.StatusOK, list.Code, list.Body.String())
		})
	}

	t.Run("v2 rejects oauth_resource as unknown", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodPost, "/api/v2/tokens", "{"+fields+`,"oauth_resource":"https://localhost/api/v2/mcp"}`, humaTokenFor(t, &testuser1), "")
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code, rec.Body.String())
	})
}

func TestAPITokenListExposesOAuthClientID(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	insertOAuthToken(t, mcpResourceURL, projectsReadPermissions())

	rec := humaRequest(t, e, http.MethodGet, "/api/v2/tokens?per_page=100", "", humaTokenFor(t, &testuser1), "")
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.Contains(t, rec.Body.String(), `"oauth_client_id":"test-client"`)
	assert.NotContains(t, rec.Body.String(), "oauth_resource")
}

func authorizePreview(t *testing.T, e *echo.Echo, token string, params map[string]string) *httpResult {
	t.Helper()
	body := map[string]string{
		"response_type":         "code",
		"client_id":             scopedClientID,
		"redirect_uri":          scopedRedirectURI,
		"state":                 "preview-state",
		"code_challenge":        pkceChallenge(scopedVerifier),
		"code_challenge_method": "S256",
	}
	for k, v := range params {
		body[k] = v
	}
	raw, err := json.Marshal(body)
	require.NoError(t, err)
	rec := humaRequest(t, e, http.MethodPost, "/api/v2/oauth/authorize/preview", string(raw), token, "")
	return &httpResult{code: rec.Code, body: rec.Body.Bytes()}
}

func countOAuthCodes(t *testing.T) int64 {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	n, err := s.Count(&models.OAuthCode{})
	require.NoError(t, err)
	return n
}

func TestOAuthAuthorizePreview(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	userToken := humaTokenFor(t, &testuser1)

	t.Run("scoped request", func(t *testing.T) {
		before := countOAuthCodes(t)
		res := authorizePreview(t, e, userToken, map[string]string{
			"scope":    "mcp:write offline_access",
			"resource": mcpResourceURL,
		})
		require.Equal(t, http.StatusOK, res.code, string(res.body))
		assert.JSONEq(t, `{
			"client": {"id": "test-mcp-client"},
			"redirect_uri": "http://127.0.0.1:33418/callback",
			"scopes": ["mcp:read", "mcp:write"],
			"scoped": true
		}`, stripSchema(t, res.body))
		assert.Equal(t, before, countOAuthCodes(t), "preview must not create a code")
	})

	t.Run("read-only scope", func(t *testing.T) {
		res := authorizePreview(t, e, userToken, map[string]string{"scope": "mcp:read"})
		require.Equal(t, http.StatusOK, res.code, string(res.body))
		var preview oauth2server.AuthorizePreview
		require.NoError(t, json.Unmarshal(res.body, &preview))
		assert.Equal(t, []oauth2server.Scope{oauth2server.ScopeMCPRead}, preview.Scopes)
		assert.True(t, preview.Scoped)
	})

	t.Run("unscoped request", func(t *testing.T) {
		res := authorizePreview(t, e, userToken, nil)
		require.Equal(t, http.StatusOK, res.code, string(res.body))
		assert.JSONEq(t, `{
			"client": {"id": "test-mcp-client"},
			"redirect_uri": "http://127.0.0.1:33418/callback",
			"scopes": [],
			"scoped": false
		}`, stripSchema(t, res.body))
	})

	t.Run("rejects an invalid redirect_uri", func(t *testing.T) {
		res := authorizePreview(t, e, userToken, map[string]string{"redirect_uri": "https://evil.example/callback"})
		require.Equal(t, http.StatusBadRequest, res.code, string(res.body))
		assert.Equal(t, models.ErrCodeOAuthInvalidRedirectURI, res.errorCode(t))
	})

	t.Run("requires authentication", func(t *testing.T) {
		res := authorizePreview(t, e, "", map[string]string{"scope": "mcp:read"})
		assert.Equal(t, http.StatusUnauthorized, res.code, string(res.body))
	})

	t.Run("rejects API tokens", func(t *testing.T) {
		for name, token := range map[string]string{
			"legacy oauth token": insertLegacyOAuthScopedToken(t),
			"mcp oauth token":    insertOAuthToken(t, mcpResourceURL, projectsReadPermissions()),
		} {
			res := authorizePreview(t, e, token, map[string]string{"scope": "mcp:read"})
			assert.Equal(t, http.StatusUnauthorized, res.code, "%s: %s", name, res.body)
			assert.NotContains(t, string(res.body), `"client"`, name)
		}
	})
}

func stripSchema(t *testing.T, body []byte) string {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal(body, &m))
	delete(m, "$schema")
	out, err := json.Marshal(m)
	require.NoError(t, err)
	return string(out)
}
