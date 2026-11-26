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
	"net/url"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"

	"github.com/labstack/echo/v5"
	"github.com/modelcontextprotocol/go-sdk/oauthex"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	resourceMetadataRoot = "/.well-known/oauth-protected-resource"
	resourceMetadataPath = resourceMetadataRoot + "/api/v2/mcp"
)

func TestMCPProtectedResourceMetadata(t *testing.T) {
	t.Run("describes the MCP endpoint", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		for _, path := range []string{resourceMetadataPath, resourceMetadataRoot} {
			rec := getMetadata(e, path, nil)
			require.Equal(t, http.StatusOK, rec.Code, "%s: %s", path, rec.Body.String())
			assert.Contains(t, rec.Header().Get(echo.HeaderContentType), "application/json", path)
			assert.Equal(t, "public, max-age=3600", rec.Header().Get(echo.HeaderCacheControl), path)
			assert.JSONEq(t, `{
				"resource": "https://localhost/api/v2/mcp",
				"authorization_servers": ["https://localhost"],
				"scopes_supported": ["mcp:read", "mcp:write"],
				"bearer_methods_supported": ["header"],
				"resource_name": "Vikunja"
			}`, rec.Body.String(), path)
		}
	})

	t.Run("passes the MCP SDK's validation", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		server := httptest.NewServer(e)
		t.Cleanup(server.Close)
		withPublicURL(t, server.URL+"/")

		metadata, err := oauthex.GetProtectedResourceMetadata(t.Context(), server.URL+resourceMetadataPath, server.URL+"/api/v2/mcp", server.Client())
		require.NoError(t, err)
		assert.Equal(t, []string{server.URL}, metadata.AuthorizationServers)
	})

	t.Run("is not found without a public URL", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		withoutPublicURL(t)

		for _, path := range []string{resourceMetadataPath, resourceMetadataRoot} {
			rec := getMetadata(e, path, nil)
			assert.Equal(t, http.StatusNotFound, rec.Code, path)
			assert.Empty(t, rec.Body.String(), "%s must not fall back to the frontend's index.html", path)
		}
	})

	for name, caldav := range map[string]bool{"caldav enabled": true, "caldav disabled": false} {
		t.Run("needs no authentication with "+name, func(t *testing.T) {
			e := discoveryEnv(t, map[config.Key]any{config.ServiceEnableCaldav: caldav})

			for _, path := range []string{resourceMetadataPath, resourceMetadataRoot} {
				rec := getMetadata(e, path, nil)
				assert.Equal(t, http.StatusOK, rec.Code, "%s: %s", path, rec.Body.String())
				assert.Contains(t, rec.Body.String(), `"resource":"https://localhost/api/v2/mcp"`, path)
			}
			if caldav {
				principal := getMetadata(e, "/.well-known/caldav", nil)
				assert.Equal(t, http.StatusUnauthorized, principal.Code, "the caldav group must keep its basic auth")
			}
		})
	}

	t.Run("is readable from any origin", func(t *testing.T) {
		e := discoveryEnv(t, map[config.Key]any{
			config.CorsEnable:  true,
			config.CorsOrigins: []string{"https://allowed.example"},
		})

		for _, path := range []string{resourceMetadataPath, resourceMetadataRoot} {
			for _, origin := range []string{"https://mcp-client.example", "https://allowed.example"} {
				rec := getMetadata(e, path, http.Header{"Origin": {origin}})
				require.Equal(t, http.StatusOK, rec.Code, "%s %s", path, origin)
				assert.Equal(t, "*", rec.Header().Get(echo.HeaderAccessControlAllowOrigin), "%s %s", path, origin)
				assert.Empty(t, rec.Header().Get(echo.HeaderAccessControlAllowCredentials), "%s %s", path, origin)
			}

			req := httptest.NewRequest(http.MethodOptions, path, nil)
			req.Header.Set("Origin", "https://mcp-client.example")
			req.Header.Set(echo.HeaderAccessControlRequestMethod, http.MethodGet)
			req.Header.Set(echo.HeaderAccessControlRequestHeaders, "mcp-protocol-version")
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			assert.Equal(t, http.StatusNoContent, rec.Code, path)
			assert.Equal(t, "*", rec.Header().Get(echo.HeaderAccessControlAllowOrigin), path)
			assert.Equal(t, "mcp-protocol-version", rec.Header().Get(echo.HeaderAccessControlAllowHeaders), path)
			assert.Empty(t, rec.Header().Get(echo.HeaderAccessControlAllowCredentials), path)
		}
	})

	t.Run("names the subpath endpoint in a subpath deployment", func(t *testing.T) {
		e := discoveryEnv(t, map[config.Key]any{config.ServicePublicURL: "https://example.com/vikunja/"})

		for _, path := range []string{
			resourceMetadataRoot + "/vikunja/api/v2/mcp",
			resourceMetadataPath,
			resourceMetadataRoot,
		} {
			rec := getMetadata(e, path, nil)
			require.Equal(t, http.StatusOK, rec.Code, path)
			assert.JSONEq(t, `{
				"resource": "https://example.com/vikunja/api/v2/mcp",
				"authorization_servers": ["https://example.com/vikunja"],
				"scopes_supported": ["mcp:read", "mcp:write"],
				"bearer_methods_supported": ["header"],
				"resource_name": "Vikunja"
			}`, rec.Body.String(), path)
		}
	})
}

const (
	resourceMetadataURL   = "https://localhost" + resourceMetadataPath
	missingTokenChallenge = `Bearer resource_metadata="` + resourceMetadataURL + `"`
	invalidTokenChallenge = `Bearer resource_metadata="` + resourceMetadataURL + `", error="invalid_token", ` +
		`error_description="The access token is expired, revoked or not valid for this resource"`
)

func postMCP(e *echo.Echo, header http.Header) *httptest.ResponseRecorder {
	return serveMCPRequest(e, mcpRequest(http.MethodPost, initializeBody), header)
}

func serveMCPRequest(e *echo.Echo, req *http.Request, header http.Header) *httptest.ResponseRecorder {
	for k, v := range header {
		req.Header[k] = v
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func bearer(token string) http.Header {
	return http.Header{echo.HeaderAuthorization: {"Bearer " + token}}
}

func TestMCPChallenge(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	sessionJWT := humaTokenFor(t, &testuser1)

	for name, tc := range map[string]struct {
		header http.Header
		code   int
		want   string
	}{
		"no credentials": {
			code: http.StatusUnauthorized,
			want: missingTokenChallenge,
		},
		"unknown api token": {
			header: bearer("tk_0000000000000000000000000000000000000000"),
			code:   http.StatusUnauthorized,
			want:   invalidTokenChallenge,
		},
		"expired oauth token": {
			header: bearer(insertOAuthTokenExpiring(t, mcpResourceURL, projectsReadPermissions(), time.Now().Add(-time.Hour))),
			code:   http.StatusUnauthorized,
			want:   invalidTokenChallenge,
		},
		"malformed token": {
			header: bearer("not-a-token"),
			code:   http.StatusUnauthorized,
			want:   invalidTokenChallenge,
		},
		"basic credentials": {
			header: http.Header{echo.HeaderAuthorization: {"Basic dXNlcjE6MTIzNDU2NzhUZXN0"}},
			code:   http.StatusUnauthorized,
			want:   invalidTokenChallenge,
		},
		"session jwt": {
			header: bearer(sessionJWT),
			code:   http.StatusUnauthorized,
			want:   invalidTokenChallenge,
		},
		"oauth token for another resource": {
			header: bearer(insertOAuthToken(t, "https://other.example/api/v2/mcp", projectsReadPermissions())),
			code:   http.StatusUnauthorized,
			want:   invalidTokenChallenge,
		},
		"api token without mcp access": {
			header: bearer(noMCPToken),
			code:   http.StatusForbidden,
		},
		"hand-made api token": {
			header: bearer(mcpProjectsReadToken),
			code:   http.StatusOK,
		},
		"oauth token for this resource": {
			header: bearer(insertOAuthToken(t, mcpResourceURL, projectsReadPermissions())),
			code:   http.StatusOK,
		},
	} {
		t.Run(name, func(t *testing.T) {
			rec := postMCP(e, tc.header)
			require.Equal(t, tc.code, rec.Code, rec.Body.String())
			assert.Equal(t, tc.want, rec.Header().Get(echo.HeaderWWWAuthenticate))
		})
	}

	t.Run("parses with the MCP SDK", func(t *testing.T) {
		rec := postMCP(e, bearer(sessionJWT))
		challenges, err := oauthex.ParseWWWAuthenticate(rec.Header().Values(echo.HeaderWWWAuthenticate))
		require.NoError(t, err)
		require.Len(t, challenges, 1)
		assert.Equal(t, "bearer", challenges[0].Scheme)
		assert.Equal(t, resourceMetadataURL, challenges[0].Params["resource_metadata"])
		assert.Equal(t, "invalid_token", challenges[0].Params["error"])
	})

	t.Run("never takes the token from the query string", func(t *testing.T) {
		req := mcpRequest(http.MethodPost, initializeBody)
		req.URL.RawQuery = url.Values{"access_token": {mcpProjectsReadToken}}.Encode()
		rec := serveMCPRequest(e, req, nil)
		require.Equal(t, http.StatusUnauthorized, rec.Code, rec.Body.String())
		assert.Equal(t, missingTokenChallenge, rec.Header().Get(echo.HeaderWWWAuthenticate))
	})

	t.Run("is absent without a public URL", func(t *testing.T) {
		withoutPublicURL(t)
		for _, header := range []http.Header{nil, bearer(sessionJWT), bearer(noMCPToken)} {
			rec := postMCP(e, header)
			assert.Contains(t, []int{http.StatusUnauthorized, http.StatusForbidden}, rec.Code, rec.Body.String())
			assert.Empty(t, rec.Header().Values(echo.HeaderWWWAuthenticate))
		}
	})

	t.Run("leaves other routes' 401s alone", func(t *testing.T) {
		oauthToken := insertOAuthToken(t, mcpResourceURL, projectsReadPermissions())
		for _, tc := range []struct {
			path   string
			header http.Header
		}{
			{path: "/api/v2/projects"},
			{path: "/api/v1/projects"},
			{path: "/api/v2/projects", header: bearer("tk_0000000000000000000000000000000000000000")},
			{path: "/api/v2/projects", header: bearer(oauthToken)},
			{path: "/api/v1/projects", header: bearer(oauthToken)},
		} {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			for k, v := range tc.header {
				req.Header[k] = v
			}
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			require.Equal(t, http.StatusUnauthorized, rec.Code, "%s: %s", tc.path, rec.Body.String())
			assert.Empty(t, rec.Header().Values(echo.HeaderWWWAuthenticate), tc.path)
		}
	})
}

func withPublicURL(t *testing.T, publicURL string) {
	t.Helper()
	previous := config.ServicePublicURL.GetString()
	config.ServicePublicURL.Set(publicURL)
	t.Cleanup(func() { config.ServicePublicURL.Set(previous) })
}
