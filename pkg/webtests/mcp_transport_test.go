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
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func serveMCP(t *testing.T, req *http.Request) *httptest.ResponseRecorder {
	t.Helper()
	e, err := setupTestEnv()
	require.NoError(t, err)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+mcpOnlyToken)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func TestMCP_GetNotAllowed(t *testing.T) {
	rec := serveMCP(t, mcpRequest(http.MethodGet, ""))
	require.Equal(t, http.StatusMethodNotAllowed, rec.Code, "%s", rec.Body.String())
	assert.Contains(t, rec.Header().Get(echo.HeaderAllow), http.MethodPost)
}

func TestMCP_DeleteNotAllowed(t *testing.T) {
	rec := serveMCP(t, mcpRequest(http.MethodDelete, ""))
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code, "%s", rec.Body.String())
}

func TestMCP_OversizedBodyKeepsMCPMessage(t *testing.T) {
	c := newMCPClient(t, mcpOnlyToken)
	rec := c.post(fmt.Sprintf(`{"jsonrpc":"2.0","id":1,"method":"ping","params":{"pad":%q}}`, strings.Repeat("a", 4<<20)))
	require.Equal(t, http.StatusRequestEntityTooLarge, rec.Code)
	assert.Contains(t, rec.Body.String(), "MCP request body")
}

func TestMCP_CrossOriginRejected(t *testing.T) {
	req := mcpRequest(http.MethodPost, initializeBody)
	req.Header.Set("Origin", "https://evil.example")
	rec := serveMCP(t, req)
	assert.Equal(t, http.StatusForbidden, rec.Code, "%s", rec.Body.String())

	rec = serveMCP(t, mcpRequest(http.MethodPost, initializeBody))
	assert.Equal(t, http.StatusOK, rec.Code, "%s", rec.Body.String())
}

func TestMCP_SubPathNotFound(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v2/mcp/anything", strings.NewReader(initializeBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	rec := serveMCP(t, req)
	assert.Equal(t, http.StatusNotFound, rec.Code, "%s", rec.Body.String())
}
