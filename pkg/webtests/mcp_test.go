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
	// Token 10 has only the mcp:access scope, owned by user 1.
	mcpOnlyToken = "tk_mcp_access_token_test_0000000000mcp0001"
	// Token 1 has only {tasks:[read_all, update]} — no mcp scope. Owner: user 1.
	// Token 11 (mcp + projects:{read_one, read_all}) lives in mcp_scopes_test.go.
	noMCPToken = "tk_2eef46f40ebab3304919ab2e7e39993f75f29d2e"
)

// The streamable-HTTP transport requires both Accept types.
func mcpRequest(method, body string) *http.Request {
	req := httptest.NewRequest(method, "/api/v2/mcp", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	return req
}

// The SDK returns either application/json or a single-event SSE stream, depending on negotiation.
func readMCPJSON(t *testing.T, body string) map[string]any {
	t.Helper()
	body = strings.TrimSpace(body)
	// SSE framing — find the first "data: " line.
	if strings.HasPrefix(body, "event:") || strings.Contains(body, "data:") {
		for _, line := range strings.Split(body, "\n") {
			if strings.HasPrefix(line, "data:") {
				body = strings.TrimSpace(strings.TrimPrefix(line, "data:"))
				break
			}
		}
	}
	var out map[string]any
	require.NoError(t, json.Unmarshal([]byte(body), &out), "body was: %s", body)
	return out
}

// Token 12 (owner user 1) has mcp:access plus every projects_* scope.
const mcpFullProjectsToken = "tk_mcp_full_projects_token_test_0fullp003"

// mcpClient does the initialize / notifications handshake against the live Echo server.
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
	c.initialize()
	c.notifyInitialized()
	return c
}

func (c *mcpClient) initialize() {
	c.t.Helper()
	body := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"0.1"}}}`
	req := mcpRequest(http.MethodPost, body)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+c.token)
	rec := httptest.NewRecorder()
	c.e.ServeHTTP(rec, req)
	require.Equal(c.t, http.StatusOK, rec.Code, "initialize body: %s", rec.Body.String())
	c.sessionID = rec.Header().Get("Mcp-Session-Id")
	require.NotEmpty(c.t, c.sessionID, "no session id on initialize response")
}

func (c *mcpClient) notifyInitialized() {
	c.t.Helper()
	req := mcpRequest(http.MethodPost, `{"jsonrpc":"2.0","method":"notifications/initialized"}`)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+c.token)
	req.Header.Set("Mcp-Session-Id", c.sessionID)
	rec := httptest.NewRecorder()
	c.e.ServeHTTP(rec, req)
	require.Less(c.t, rec.Code, 400, "notifications/initialized: %s", rec.Body.String())
}

// Each call uses a fresh request id so the SDK doesn't confuse responses.
func (c *mcpClient) rpc(method string, params any) map[string]any {
	c.t.Helper()
	c.nextID++
	paramsJSON, err := json.Marshal(params)
	require.NoError(c.t, err)
	body := fmt.Sprintf(`{"jsonrpc":"2.0","id":%d,"method":%q,"params":%s}`, c.nextID, method, paramsJSON)
	req := mcpRequest(http.MethodPost, body)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+c.token)
	req.Header.Set("Mcp-Session-Id", c.sessionID)
	rec := httptest.NewRecorder()
	c.e.ServeHTTP(rec, req)
	require.Equal(c.t, http.StatusOK, rec.Code, "rpc %s body: %s", method, rec.Body.String())
	return readMCPJSON(c.t, rec.Body.String())
}

// callTool returns the raw "result" payload; success/failure lives in result["isError"].
func (c *mcpClient) callTool(name string, args map[string]any) map[string]any {
	c.t.Helper()
	resp := c.rpc("tools/call", map[string]any{
		"name":      name,
		"arguments": args,
	})
	result, ok := resp["result"].(map[string]any)
	require.Truef(c.t, ok, "missing result for %s: %v", name, resp)
	return result
}

// toolResultText extracts the first TextContent entry from a tools/call result.
func toolResultText(t *testing.T, result map[string]any) string {
	t.Helper()
	content, ok := result["content"].([]any)
	require.Truef(t, ok, "no content in result: %v", result)
	require.NotEmpty(t, content, "empty content array: %v", result)
	first, ok := content[0].(map[string]any)
	require.True(t, ok, "first content not an object: %v", content[0])
	text, ok := first["text"].(string)
	require.Truef(t, ok, "first content text missing or not a string: %v", first)
	return text
}

// read_all returns the {items, total, page, per_page, total_pages} envelope, not a bare array.
func readAllItems(t *testing.T, result map[string]any, dest any) {
	t.Helper()
	text := toolResultText(t, result)
	var env struct {
		Items json.RawMessage `json:"items"`
	}
	require.NoError(t, json.Unmarshal([]byte(text), &env), "text was: %s", text)
	require.NoError(t, json.Unmarshal(env.Items, dest), "items were: %s", env.Items)
}

func TestMCP_AnonymousRejected(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)

	req := mcpRequest(http.MethodPost, `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestMCP_JWTRejected(t *testing.T) {
	// MCP is a token-only endpoint. JWT bypasses CanDoAPIRoute entirely, so
	// without an explicit rejection the scope gate would be moot.
	e, err := setupTestEnv()
	require.NoError(t, err)

	s := db.NewSession()
	defer s.Close()
	u, err := user.GetUserByID(s, 1)
	require.NoError(t, err)
	jwt, err := auth.NewUserJWTAuthtoken(u, "test-session-id")
	require.NoError(t, err)

	req := mcpRequest(http.MethodPost, `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+jwt)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestMCP_TokenWithoutMCPScopeRejected(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)

	req := mcpRequest(http.MethodPost, `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+noMCPToken)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestMCP_InitializeWithMCPToken(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)

	req := mcpRequest(http.MethodPost, `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"0.1"}}}`)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+mcpOnlyToken)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
	payload := readMCPJSON(t, rec.Body.String())
	result, ok := payload["result"].(map[string]any)
	require.True(t, ok, "response missing result: %s", rec.Body.String())
	assert.NotEmpty(t, result["protocolVersion"])
	serverInfo, ok := result["serverInfo"].(map[string]any)
	require.True(t, ok, "response missing serverInfo: %s", rec.Body.String())
	assert.Equal(t, "vikunja", serverInfo["name"])

	// The SDK exposes the session ID via the Mcp-Session-Id header.
	assert.NotEmpty(t, rec.Header().Get("Mcp-Session-Id"))
}

func TestMCP_ToolsListReturnsRegisteredResources(t *testing.T) {
	// An mcp-only token (no projects scope) sees zero project
	// tools in tools/list — the per-session tool registration filters by
	// the requesting token's (group, permission) scopes. Tools/list visibility
	// for tokens with project scopes is covered in mcp_scopes_test.go.
	e, err := setupTestEnv()
	require.NoError(t, err)

	// Step 1: initialize so the SDK opens a session.
	initReq := mcpRequest(http.MethodPost, `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"0.1"}}}`)
	initReq.Header.Set(echo.HeaderAuthorization, "Bearer "+mcpOnlyToken)
	initRec := httptest.NewRecorder()
	e.ServeHTTP(initRec, initReq)
	require.Equal(t, http.StatusOK, initRec.Code, "body: %s", initRec.Body.String())
	sessionID := initRec.Header().Get("Mcp-Session-Id")
	require.NotEmpty(t, sessionID)

	// Step 2: send the required "notifications/initialized" client message.
	initNotifyReq := mcpRequest(http.MethodPost, `{"jsonrpc":"2.0","method":"notifications/initialized"}`)
	initNotifyReq.Header.Set(echo.HeaderAuthorization, "Bearer "+mcpOnlyToken)
	initNotifyReq.Header.Set("Mcp-Session-Id", sessionID)
	initNotifyRec := httptest.NewRecorder()
	e.ServeHTTP(initNotifyRec, initNotifyReq)
	require.Less(t, initNotifyRec.Code, 400, "body: %s", initNotifyRec.Body.String())

	// Step 3: ask for tools.
	listReq := mcpRequest(http.MethodPost, `{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}`)
	listReq.Header.Set(echo.HeaderAuthorization, "Bearer "+mcpOnlyToken)
	listReq.Header.Set("Mcp-Session-Id", sessionID)
	listRec := httptest.NewRecorder()
	e.ServeHTTP(listRec, listReq)

	require.Equal(t, http.StatusOK, listRec.Code, "body: %s", listRec.Body.String())
	payload := readMCPJSON(t, listRec.Body.String())
	result, ok := payload["result"].(map[string]any)
	require.True(t, ok, "response missing result: %s", listRec.Body.String())
	tools, ok := result["tools"].([]any)
	require.True(t, ok, "response missing tools array: %s", listRec.Body.String())

	// No project tools because the token has no projects:* scopes.
	projectToolCount := 0
	for _, raw := range tools {
		tool, isMap := raw.(map[string]any)
		require.True(t, isMap, "tool entry should be an object: %v", raw)
		name, _ := tool["name"].(string)
		if strings.HasPrefix(name, "projects_") {
			projectToolCount++
		}
	}
	assert.Zero(t, projectToolCount, "mcp-only token must see zero project tools, got %v", tools)
}

func TestMCP_SessionRoundTrip(t *testing.T) {
	// Verifies that the Mcp-Session-Id round-trip survives the Echo wrapper.
	e, err := setupTestEnv()
	require.NoError(t, err)

	initReq := mcpRequest(http.MethodPost, `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"0.1"}}}`)
	initReq.Header.Set(echo.HeaderAuthorization, "Bearer "+mcpOnlyToken)
	initRec := httptest.NewRecorder()
	e.ServeHTTP(initRec, initReq)
	require.Equal(t, http.StatusOK, initRec.Code, "body: %s", initRec.Body.String())
	sessionID := initRec.Header().Get("Mcp-Session-Id")
	require.NotEmpty(t, sessionID)

	// A follow-up request with a known session id should be accepted (not
	// rejected as "session not found").
	pingReq := mcpRequest(http.MethodPost, `{"jsonrpc":"2.0","id":99,"method":"ping","params":{}}`)
	pingReq.Header.Set(echo.HeaderAuthorization, "Bearer "+mcpOnlyToken)
	pingReq.Header.Set("Mcp-Session-Id", sessionID)
	pingRec := httptest.NewRecorder()
	e.ServeHTTP(pingRec, pingReq)
	require.Equal(t, http.StatusOK, pingRec.Code, "body: %s", pingRec.Body.String())
}

func TestMCP_SessionIDDoesNotCarryIdentity(t *testing.T) {
	// A leaked Mcp-Session-Id must not let a second token inherit the tool set
	// (and thus the scopes) of whoever opened the session.
	e, err := setupTestEnv()
	require.NoError(t, err)

	initReq := mcpRequest(http.MethodPost, `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"0.1"}}}`)
	initReq.Header.Set(echo.HeaderAuthorization, "Bearer "+mcpFullProjectsToken)
	initRec := httptest.NewRecorder()
	e.ServeHTTP(initRec, initReq)
	require.Equal(t, http.StatusOK, initRec.Code, "body: %s", initRec.Body.String())
	sessionID := initRec.Header().Get("Mcp-Session-Id")
	require.NotEmpty(t, sessionID)

	listReq := mcpRequest(http.MethodPost, `{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}`)
	listReq.Header.Set(echo.HeaderAuthorization, "Bearer "+mcpOnlyToken)
	listReq.Header.Set("Mcp-Session-Id", sessionID)
	listRec := httptest.NewRecorder()
	e.ServeHTTP(listRec, listReq)
	require.Equal(t, http.StatusOK, listRec.Code, "body: %s", listRec.Body.String())

	names := toolNamesFromList(t, readMCPJSON(t, listRec.Body.String()))
	assert.Equal(t, map[string]bool{"find_action": true, "do_action": true}, names,
		"mcp-only token must only see the catalog meta-tools, got %v", names)
}

func TestMCP_NonLoopbackHostAccepted(t *testing.T) {
	// The SDK's DNS-rebinding guard would 403 the usual reverse-proxy-to-
	// 127.0.0.1 deployment.
	e, err := setupTestEnv()
	require.NoError(t, err)

	req := mcpRequest(http.MethodPost, `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"0.1"}}}`)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+mcpOnlyToken)
	req.Host = "vikunja.example.com"
	// The guard only fires for a request served on a loopback address, which
	// httptest never sets — without this the test would pass either way.
	localAddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:3456")
	require.NoError(t, err)
	req = req.WithContext(context.WithValue(req.Context(), http.LocalAddrContextKey, localAddr))
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
}
