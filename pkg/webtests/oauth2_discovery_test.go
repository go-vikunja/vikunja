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
	"net/http/httptest"
	"net/url"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/modules/auth/oauth2server"
	"code.vikunja.io/api/pkg/routes"

	"github.com/labstack/echo/v5"
	"github.com/modelcontextprotocol/go-sdk/oauthex"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const metadataPath = "/.well-known/oauth-authorization-server"

// discoveryEnv rebuilds the routes after configure, for settings read at registration.
func discoveryEnv(t *testing.T, configure func()) *echo.Echo {
	t.Helper()
	_, err := setupTestEnv()
	require.NoError(t, err)
	configure()
	e := routes.NewEcho()
	routes.RegisterRoutes(e)
	return e
}

func getMetadata(e *echo.Echo, path string, header http.Header) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	for k, v := range header {
		req.Header[k] = v
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func TestOAuthAuthorizationServerMetadata(t *testing.T) {
	t.Run("describes the authorization server", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		rec := getMetadata(e, metadataPath, nil)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		assert.Contains(t, rec.Header().Get(echo.HeaderContentType), "application/json")
		assert.Equal(t, "public, max-age=3600", rec.Header().Get(echo.HeaderCacheControl))
		assert.JSONEq(t, `{
			"issuer": "https://localhost",
			"authorization_endpoint": "https://localhost/oauth/authorize",
			"token_endpoint": "https://localhost/api/v2/oauth/token",
			"response_types_supported": ["code"],
			"grant_types_supported": ["authorization_code"],
			"code_challenge_methods_supported": ["S256"],
			"token_endpoint_auth_methods_supported": ["none"],
			"scopes_supported": ["mcp:read", "mcp:write"],
			"client_id_metadata_document_supported": true,
			"authorization_response_iss_parameter_supported": true
		}`, rec.Body.String())
	})

	t.Run("advertises a working token endpoint", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		var metadata oauth2server.AuthorizationServerMetadata
		require.NoError(t, json.Unmarshal(getMetadata(e, metadataPath, nil).Body.Bytes(), &metadata))
		tokenEndpoint, err := url.Parse(metadata.TokenEndpoint)
		require.NoError(t, err)

		authz := scopedAuthorize(t, e, "v2", map[string]string{"scope": "mcp:read"})
		require.Equal(t, http.StatusOK, authz.code, string(authz.body))
		var authorized oauth2server.AuthorizeResponse
		require.NoError(t, json.Unmarshal(authz.body, &authorized))
		form := url.Values{
			"grant_type":    {"authorization_code"},
			"code":          {authorized.Code},
			"client_id":     {scopedClientID},
			"redirect_uri":  {scopedRedirectURI},
			"code_verifier": {scopedVerifier},
		}
		rec := humaRequest(t, e, http.MethodPost, tokenEndpoint.Path, form.Encode(), "", "application/x-www-form-urlencoded")
		assert.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	})

	t.Run("passes the MCP SDK's validation", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		server := httptest.NewServer(e)
		t.Cleanup(server.Close)
		config.ServicePublicURL.Set(server.URL + "/")

		metadata, err := oauthex.GetAuthServerMeta(t.Context(), server.URL+metadataPath, server.URL, server.Client())
		require.NoError(t, err)
		require.NotNil(t, metadata)
		assert.Equal(t, server.URL+"/oauth/authorize", metadata.AuthorizationEndpoint)
	})

	t.Run("is not found without a public URL", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		withoutPublicURL(t)

		rec := getMetadata(e, metadataPath, nil)
		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.Empty(t, rec.Body.String(), "must not fall back to the frontend's index.html")
	})

	for name, caldav := range map[string]bool{"caldav enabled": true, "caldav disabled": false} {
		t.Run("needs no authentication with "+name, func(t *testing.T) {
			e := discoveryEnv(t, func() { config.ServiceEnableCaldav.Set(caldav) })

			rec := getMetadata(e, metadataPath, nil)
			assert.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			assert.Contains(t, rec.Body.String(), `"issuer":"https://localhost"`)

			if caldav {
				principal := getMetadata(e, "/.well-known/caldav", nil)
				assert.Equal(t, http.StatusUnauthorized, principal.Code, "the caldav group must keep its basic auth")
			}
		})
	}

	t.Run("is readable from any origin", func(t *testing.T) {
		e := discoveryEnv(t, func() {
			config.CorsEnable.Set(true)
			config.CorsOrigins.Set([]string{"https://allowed.example"})
		})

		for _, origin := range []string{"https://mcp-client.example", "https://allowed.example"} {
			rec := getMetadata(e, metadataPath, http.Header{"Origin": {origin}})
			require.Equal(t, http.StatusOK, rec.Code, origin)
			assert.Equal(t, "*", rec.Header().Get(echo.HeaderAccessControlAllowOrigin), origin)
			assert.Empty(t, rec.Header().Get(echo.HeaderAccessControlAllowCredentials), origin)
		}

		req := httptest.NewRequest(http.MethodOptions, metadataPath, nil)
		req.Header.Set("Origin", "https://mcp-client.example")
		req.Header.Set(echo.HeaderAccessControlRequestMethod, http.MethodGet)
		req.Header.Set(echo.HeaderAccessControlRequestHeaders, "mcp-protocol-version")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNoContent, rec.Code)
		assert.Equal(t, "*", rec.Header().Get(echo.HeaderAccessControlAllowOrigin))
		assert.Contains(t, rec.Header().Get(echo.HeaderAccessControlAllowMethods), http.MethodGet)
		assert.Equal(t, "mcp-protocol-version", rec.Header().Get(echo.HeaderAccessControlAllowHeaders))
		assert.Empty(t, rec.Header().Get(echo.HeaderAccessControlAllowCredentials))
	})

	t.Run("is served below the issuer path of a subpath deployment", func(t *testing.T) {
		e := discoveryEnv(t, func() { config.ServicePublicURL.Set("https://example.com/vikunja/") })

		for _, path := range []string{metadataPath + "/vikunja", metadataPath} {
			rec := getMetadata(e, path, nil)
			require.Equal(t, http.StatusOK, rec.Code, path)
			var metadata oauth2server.AuthorizationServerMetadata
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &metadata))
			assert.Equal(t, "https://example.com/vikunja", metadata.Issuer, path)
			assert.Equal(t, "https://example.com/vikunja/api/v2/oauth/token", metadata.TokenEndpoint, path)
		}
	})
}

func TestOAuthAuthorizeIssuer(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)

	for name, tc := range map[string]struct {
		version string
		params  map[string]string
	}{
		"v2 scoped":   {version: "v2", params: map[string]string{"scope": "mcp:read"}},
		"v2 unscoped": {version: "v2"},
		"v1 scoped":   {version: "v1", params: map[string]string{"scope": "mcp:read"}},
	} {
		t.Run(name, func(t *testing.T) {
			res := scopedAuthorize(t, e, tc.version, tc.params)
			require.Equal(t, http.StatusOK, res.code, string(res.body))
			var authorized oauth2server.AuthorizeResponse
			require.NoError(t, json.Unmarshal(res.body, &authorized))
			assert.Equal(t, "https://localhost", authorized.Issuer)
		})
	}

	t.Run("omitted without a public URL", func(t *testing.T) {
		withoutPublicURL(t)
		res := scopedAuthorize(t, e, "v2", nil)
		require.Equal(t, http.StatusOK, res.code, string(res.body))
		assert.NotContains(t, string(res.body), `"iss"`)
	})
}
