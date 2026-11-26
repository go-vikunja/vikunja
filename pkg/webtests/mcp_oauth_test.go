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

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const mcpResourceURL = "https://localhost/api/v2/mcp"

// insertOAuthToken mints an OAuth-issued token for user 1 and returns its cleartext.
func insertOAuthToken(t *testing.T, resource string, permissions models.APIPermissions) string {
	t.Helper()
	return insertOAuthTokenExpiring(t, resource, permissions, time.Now().Add(time.Hour))
}

func insertOAuthTokenExpiring(t *testing.T, resource string, permissions models.APIPermissions, expiresAt time.Time) string {
	t.Helper()
	token := &models.APIToken{
		Title:          "oauth client",
		APIPermissions: permissions,
		ExpiresAt:      expiresAt,
		OwnerID:        1,
		OAuthClientID:  "test-client",
		OAuthResource:  resource,
	}
	s := db.NewSession()
	defer s.Close()
	require.NoError(t, token.CreateForOAuthClient(s))
	require.NoError(t, s.Commit())
	return token.Token
}

func projectsReadPermissions() models.APIPermissions {
	return models.APIPermissions{
		"mcp": {
			"access",
		},
		"projects": {
			"read_one",
			"read_all",
		},
	}
}

func serveWithToken(e *echo.Echo, method, path, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func TestMCP_OAuthTokenOnlyWorksForMCP(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	token := insertOAuthToken(t, mcpResourceURL, projectsReadPermissions())

	for _, path := range []string{
		"/api/v2/projects",
		"/api/v1/projects",
		"/api/v2/token/test",
		"/api/v1/token/test",
	} {
		rec := serveWithToken(e, http.MethodGet, path, token)
		assert.Equal(t, http.StatusUnauthorized, rec.Code, "%s: %s", path, rec.Body.String())
	}

	c := newMCPClientOn(t, e, token)
	var projects []map[string]any
	readAllItems(t, c.callTool("projects_list", map[string]any{}), &projects)
	assert.NotEmpty(t, projects)
}

func TestMCP_HandMadeTokenStillWorksOnREST(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	rec := serveWithToken(e, http.MethodGet, "/api/v2/projects", mcpProjectsReadToken)
	assert.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
}

func TestMCP_OAuthTokenAudience(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)

	for resource, want := range map[string]int{
		mcpResourceURL:                     http.StatusOK,
		"HTTPS://LOCALHOST/api/v2/mcp/":    http.StatusOK,
		"https://other.example/api/v2/mcp": http.StatusUnauthorized,
		"https://localhost/api/v2/other":   http.StatusUnauthorized,
	} {
		token := insertOAuthToken(t, resource, projectsReadPermissions())
		req := mcpRequest(http.MethodPost, initializeBody)
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, want, rec.Code, "%s: %s", resource, rec.Body.String())
	}
}
