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

func withPublicURL(t *testing.T, publicURL string) {
	t.Helper()
	previous := config.ServicePublicURL.GetString()
	config.ServicePublicURL.Set(publicURL)
	t.Cleanup(func() { config.ServicePublicURL.Set(previous) })
}
