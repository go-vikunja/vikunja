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
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

// read_all returns {items, result_count, total_items, page, per_page}, not a bare array.
func readAllItems(t *testing.T, result map[string]any, dest any) {
	t.Helper()
	text := toolResultText(t, result)
	var env struct {
		Items json.RawMessage `json:"items"`
	}
	require.NoError(t, json.Unmarshal([]byte(text), &env), "text was: %s", text)
	require.NoError(t, json.Unmarshal(env.Items, dest), "items were: %s", env.Items)
}

func TestMCP_Projects_ToolsListAll(t *testing.T) {
	// Token 12 also carries non-project scopes, so assert presence rather
	// than an exact tool count.
	c := newMCPClient(t, mcpFullProjectsToken)
	resp := c.rpc("tools/list", map[string]any{})
	result, ok := resp["result"].(map[string]any)
	require.True(t, ok)
	tools, ok := result["tools"].([]any)
	require.True(t, ok)

	names := make(map[string]bool, len(tools))
	for _, raw := range tools {
		tool := raw.(map[string]any)
		names[tool["name"].(string)] = true
	}
	for _, want := range []string{
		"projects_create",
		"projects_read_one",
		"projects_read_all",
		"projects_update",
		"projects_delete",
	} {
		assert.Truef(t, names[want], "missing tool %q in %v", want, names)
	}
}

func TestMCP_Projects_Create(t *testing.T) {
	c := newMCPClient(t, mcpFullProjectsToken)
	result := c.callTool("projects_create", map[string]any{
		"title":       "MCP created project",
		"description": "Created by mcp_projects_test",
	})
	require.NotContains(t, result, "isError", "create unexpectedly errored: %v", result)

	text := toolResultText(t, result)
	var project map[string]any
	require.NoError(t, json.Unmarshal([]byte(text), &project), "text was: %s", text)
	assert.Equal(t, "MCP created project", project["title"])
	assert.Equal(t, "Created by mcp_projects_test", project["description"])
	id, ok := project["id"].(float64)
	require.Truef(t, ok, "id missing or not a number: %v", project)
	assert.Positive(t, int(id))
}

func TestMCP_Projects_CreateMissingTitle(t *testing.T) {
	// "title" has no omitempty, so the SDK's schema validation rejects the
	// call inside the tool handler — an isError result, not an RPC error.
	c := newMCPClient(t, mcpFullProjectsToken)
	result := c.callTool("projects_create", map[string]any{})
	isErr, _ := result["isError"].(bool)
	require.Truef(t, isErr, "expected isError for missing required title: %v", result)
	assert.Contains(t, toolResultText(t, result), "mcp: invalid arguments for projects_create")
}

func TestMCP_Projects_ReadOneOwned(t *testing.T) {
	c := newMCPClient(t, mcpFullProjectsToken)
	result := c.callTool("projects_read_one", map[string]any{"id": 1})
	require.NotContains(t, result, "isError")

	text := toolResultText(t, result)
	var project map[string]any
	require.NoError(t, json.Unmarshal([]byte(text), &project))
	assert.InDelta(t, float64(1), project["id"], 0.0001)
	assert.Equal(t, "Test1", project["title"])
}

func TestMCP_Projects_ReadOneForbidden(t *testing.T) {
	// Project 20 belongs to user 13, so user 1's read must be denied.
	c := newMCPClient(t, mcpFullProjectsToken)
	result := c.callTool("projects_read_one", map[string]any{"id": 20})
	isErr, _ := result["isError"].(bool)
	require.True(t, isErr, "expected isError for forbidden project, got: %v", result)
}

func TestMCP_Projects_ReadOneNonexistent(t *testing.T) {
	c := newMCPClient(t, mcpFullProjectsToken)
	result := c.callTool("projects_read_one", map[string]any{"id": 999999})
	isErr, _ := result["isError"].(bool)
	require.True(t, isErr, "expected isError for nonexistent project, got: %v", result)
}

func TestMCP_Projects_ReadAll(t *testing.T) {
	c := newMCPClient(t, mcpFullProjectsToken)
	result := c.callTool("projects_read_all", map[string]any{})
	require.NotContains(t, result, "isError", "read_all errored: %v", result)

	var projects []map[string]any
	readAllItems(t, result, &projects)
	require.NotEmpty(t, projects, "expected at least one project")

	// User 1 owns Test1 (project id 1); confirm it's in the response.
	titles := make(map[string]bool, len(projects))
	for _, p := range projects {
		title, _ := p["title"].(string)
		titles[title] = true
	}
	assert.True(t, titles["Test1"], "expected Test1 in: %v", titles)
}

func TestMCP_Projects_ReadAllSearch(t *testing.T) {
	c := newMCPClient(t, mcpFullProjectsToken)
	result := c.callTool("projects_read_all", map[string]any{
		"search":   "Test1",
		"page":     1,
		"per_page": 50,
	})
	require.NotContains(t, result, "isError")

	var projects []map[string]any
	readAllItems(t, result, &projects)
	// At minimum the matching project Test1 should appear.
	require.NotEmpty(t, projects)
	for _, p := range projects {
		title, _ := p["title"].(string)
		assert.NotEmpty(t, title, "project missing title: %v", p)
	}
}

func TestMCP_Projects_Update(t *testing.T) {
	c := newMCPClient(t, mcpFullProjectsToken)

	// Create rather than reuse a fixture: project 1 is referenced from a lot of test data.
	createResult := c.callTool("projects_create", map[string]any{
		"title": "mcp project to update",
	})
	require.NotContains(t, createResult, "isError")
	var created map[string]any
	require.NoError(t, json.Unmarshal([]byte(toolResultText(t, createResult)), &created))
	pid := int64(created["id"].(float64))

	updateResult := c.callTool("projects_update", map[string]any{
		"id":          pid,
		"title":       "mcp project updated",
		"description": "Updated description",
	})
	require.NotContains(t, updateResult, "isError", "update errored: %v", updateResult)

	readResult := c.callTool("projects_read_one", map[string]any{"id": pid})
	require.NotContains(t, readResult, "isError")
	var project map[string]any
	require.NoError(t, json.Unmarshal([]byte(toolResultText(t, readResult)), &project))
	assert.Equal(t, "mcp project updated", project["title"])
	assert.Equal(t, "Updated description", project["description"])
}

// An explicitly sent `is_archived: false` must un-archive, not read as "omitted".
func TestMCP_Projects_UpdateClearsArchived(t *testing.T) {
	c := newMCPClient(t, mcpFullProjectsToken)

	createResult := c.callTool("projects_create", map[string]any{
		"title":       "mcp project to un-archive",
		"is_archived": true,
	})
	require.NotContains(t, createResult, "isError")
	var created map[string]any
	require.NoError(t, json.Unmarshal([]byte(toolResultText(t, createResult)), &created))
	pid := int64(created["id"].(float64))
	require.True(t, created["is_archived"].(bool), "project should have been created archived")

	updateResult := c.callTool("projects_update", map[string]any{
		"id":          pid,
		"is_archived": false,
	})
	require.NotContains(t, updateResult, "isError", "update errored: %v", updateResult)

	readResult := c.callTool("projects_read_one", map[string]any{"id": pid})
	require.NotContains(t, readResult, "isError")
	var project map[string]any
	require.NoError(t, json.Unmarshal([]byte(toolResultText(t, readResult)), &project))
	assert.False(t, project["is_archived"].(bool), "is_archived must be false after explicit clear")
}

func TestMCP_Projects_Delete(t *testing.T) {
	c := newMCPClient(t, mcpFullProjectsToken)

	createResult := c.callTool("projects_create", map[string]any{
		"title": "mcp project to delete",
	})
	require.NotContains(t, createResult, "isError")
	var created map[string]any
	require.NoError(t, json.Unmarshal([]byte(toolResultText(t, createResult)), &created))
	pid := int64(created["id"].(float64))

	deleteResult := c.callTool("projects_delete", map[string]any{"id": pid})
	require.NotContains(t, deleteResult, "isError", "delete errored: %v", deleteResult)

	readResult := c.callTool("projects_read_one", map[string]any{"id": pid})
	isErr, _ := readResult["isError"].(bool)
	require.True(t, isErr, "expected isError for deleted project, got: %v", readResult)
	text := strings.ToLower(toolResultText(t, readResult))
	assert.NotEmpty(t, text)
}
