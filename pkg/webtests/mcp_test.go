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
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/user"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	mcpOnlyToken         = "tk_mcp_access_token_test_0000000000mcp0001"
	mcpProjectsReadToken = "tk_mcp_mixed_scope_token_test_00mcpmixed02"
	mcpFullToken         = "tk_mcp_full_projects_token_test_0fullp003"
	noMCPToken           = "tk_2eef46f40ebab3304919ab2e7e39993f75f29d2e"
	initializeBody       = `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"0.1"}}}`
)

func mcpRequest(method, body string) *http.Request {
	req := httptest.NewRequest(method, "/api/v2/mcp", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	return req
}
func readMCPJSON(t *testing.T, body string) map[string]any {
	t.Helper()
	body = strings.TrimSpace(body)
	if strings.HasPrefix(body, "event:") || strings.Contains(body, "data:") {
		for _, line := range strings.Split(body, "\n") {
			if strings.HasPrefix(line, "data:") {
				body = strings.TrimSpace(strings.TrimPrefix(line, "data:"))
				break
			}
		}
	}
	var out map[string]any
	require.NoError(t, json.Unmarshal([]byte(body), &out), "body: %s", body)
	return out
}

type mcpClient struct {
	t         *testing.T
	e         *echo.Echo
	token     string
	sessionID string
	nextID    int
}

func newMCPClient(t *testing.T, token string) *mcpClient {
	t.Helper()
	e, err := setupTestEnv()
	require.NoError(t, err)
	c := &mcpClient{t: t, e: e, token: token, nextID: 1}
	rec := c.post(initializeBody)
	require.Equal(t, http.StatusOK, rec.Code, "initialize: %s", rec.Body.String())
	c.sessionID = rec.Header().Get("Mcp-Session-Id")
	require.NotEmpty(t, c.sessionID)
	rec = c.post(`{"jsonrpc":"2.0","method":"notifications/initialized"}`)
	require.Less(t, rec.Code, 400, "initialized: %s", rec.Body.String())
	return c
}
func (c *mcpClient) post(body string) *httptest.ResponseRecorder {
	c.t.Helper()
	req := mcpRequest(http.MethodPost, body)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+c.token)
	if c.sessionID != "" {
		req.Header.Set("Mcp-Session-Id", c.sessionID)
	}
	rec := httptest.NewRecorder()
	c.e.ServeHTTP(rec, req)
	return rec
}
func (c *mcpClient) rpc(method string, params any) map[string]any {
	c.t.Helper()
	c.nextID++
	raw, err := json.Marshal(params)
	require.NoError(c.t, err)
	rec := c.post(fmt.Sprintf(`{"jsonrpc":"2.0","id":%d,"method":%q,"params":%s}`, c.nextID, method, raw))
	require.Equal(c.t, http.StatusOK, rec.Code, "rpc %s: %s", method, rec.Body.String())
	return readMCPJSON(c.t, rec.Body.String())
}
func (c *mcpClient) callTool(name string, args map[string]any) map[string]any {
	c.t.Helper()
	resp := c.rpc("tools/call", map[string]any{"name": name, "arguments": args})
	result, ok := resp["result"].(map[string]any)
	require.True(c.t, ok, "missing result for %s: %v", name, resp)
	return result
}
func (c *mcpClient) toolNames() map[string]bool {
	c.t.Helper()
	resp := c.rpc("tools/list", map[string]any{})
	result, ok := resp["result"].(map[string]any)
	require.True(c.t, ok, "%v", resp)
	tools, ok := result["tools"].([]any)
	require.True(c.t, ok, "%v", resp)
	names := map[string]bool{}
	for _, raw := range tools {
		names[raw.(map[string]any)["name"].(string)] = true
	}
	return names
}
func TestMCP_AnonymousRejected(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, mcpRequest(http.MethodPost, initializeBody))
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
func TestMCP_JWTRejected(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	s := db.NewSession()
	defer s.Close()
	u, err := user.GetUserByID(s, 1)
	require.NoError(t, err)
	jwt, err := auth.NewUserJWTAuthtoken(u, "test-session-id")
	require.NoError(t, err)
	req := mcpRequest(http.MethodPost, initializeBody)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+jwt)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
func TestMCP_TokenWithoutMCPScopeRejected(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	req := mcpRequest(http.MethodPost, initializeBody)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+noMCPToken)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}
func TestMCP_Initialize(t *testing.T) {
	c := newMCPClient(t, mcpOnlyToken)
	rec := c.post(initializeBody)
	require.Equal(t, http.StatusOK, rec.Code)
	result := readMCPJSON(t, rec.Body.String())["result"].(map[string]any)
	assert.Equal(t, "vikunja", result["serverInfo"].(map[string]any)["name"])
}
func TestMCP_ToolsListMatchesScopes(t *testing.T) {
	assert.Equal(t, map[string]bool{"find_action": true, "do_action": true}, newMCPClient(t, mcpOnlyToken).toolNames())
	assert.Equal(t, map[string]bool{"find_action": true, "do_action": true, "projects_read": true, "projects_list": true}, newMCPClient(t, mcpProjectsReadToken).toolNames())
	full := newMCPClient(t, mcpFullToken).toolNames()
	for _, name := range []string{"projects_list", "projects_read", "projects_create", "projects_update", "projects_delete", "tasks_list", "tasks_read", "tasks_create", "tasks_update", "tasks_delete", "labels_list", "labels_read", "labels_create", "labels_update", "labels_delete", "task_comments_list", "task_comments_read", "task_comments_create", "task_comments_update", "task_comments_delete", "task_assignees_list", "task_assignees_create", "task_assignees_delete", "users_search", "find_action", "do_action"} {
		assert.True(t, full[name], "missing %s", name)
	}
	assert.Len(t, full, 26)
	assert.False(t, full["task_labels_create"])
}
func TestMCP_SessionIDDoesNotCarryIdentity(t *testing.T) {
	full := newMCPClient(t, mcpFullToken)
	weak := &mcpClient{t: t, e: full.e, token: mcpOnlyToken, sessionID: full.sessionID, nextID: 50}
	assert.Equal(t, map[string]bool{"find_action": true, "do_action": true}, weak.toolNames())
	result := weak.callTool("do_action", map[string]any{"action": "projects_read", "arguments": map[string]any{"id": 1}})
	assert.Equal(t, true, result["isError"])
}
func pingBatch(n int) string {
	msgs := make([]string, n)
	for i := range msgs {
		msgs[i] = fmt.Sprintf(`{"jsonrpc":"2.0","id":%d,"method":"ping","params":{}}`, i+100)
	}
	return "[" + strings.Join(msgs, ",") + "]"
}
func TestMCP_BatchLimits(t *testing.T) {
	c := newMCPClient(t, mcpOnlyToken)
	rec := c.post(pingBatch(21))
	require.Equal(t, http.StatusBadRequest, rec.Code, "%s", rec.Body.String())
	assert.Contains(t, rec.Body.String(), "20")
	rec = c.post(pingBatch(3))
	assert.Equal(t, http.StatusOK, rec.Code, "%s", rec.Body.String())
}
func TestMCP_OversizedBodyRejected(t *testing.T) {
	c := newMCPClient(t, mcpOnlyToken)
	rec := c.post(fmt.Sprintf(`{"jsonrpc":"2.0","id":1,"method":"ping","params":{"pad":%q}}`, strings.Repeat("a", 5<<20)))
	assert.Equal(t, http.StatusRequestEntityTooLarge, rec.Code)
}
func TestMCP_NonLoopbackHostAccepted(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	req := mcpRequest(http.MethodPost, initializeBody)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+mcpOnlyToken)
	req.Host = "vikunja.example.com"
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:3456")
	require.NoError(t, err)
	req = req.WithContext(context.WithValue(req.Context(), http.LocalAddrContextKey, addr))
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code, "%s", rec.Body.String())
}
