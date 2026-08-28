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
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Token 11 has mcp:access plus a partial projects scope: read_one and read_all only.
const mcpMixedScopeToken = "tk_mcp_mixed_scope_token_test_00mcpmixed02"

// toolNamesFromList extracts the "name" of every tool in a tools/list result.
func toolNamesFromList(t *testing.T, resp map[string]any) map[string]bool {
	t.Helper()
	result, ok := resp["result"].(map[string]any)
	require.True(t, ok, "response missing result: %v", resp)
	tools, ok := result["tools"].([]any)
	require.True(t, ok, "response missing tools array: %v", result)
	names := make(map[string]bool, len(tools))
	for _, raw := range tools {
		tool, isMap := raw.(map[string]any)
		require.Truef(t, isMap, "tool entry not an object: %v", raw)
		name, _ := tool["name"].(string)
		names[name] = true
	}
	return names
}

func TestMCP_Scopes_ToolsListMixed(t *testing.T) {
	// Token 11 must see exactly its two project tools and no others.
	c := newMCPClient(t, mcpMixedScopeToken)
	resp := c.rpc("tools/list", map[string]any{})
	names := toolNamesFromList(t, resp)

	assert.Truef(t, names["projects_read_one"], "expected projects_read_one in: %v", names)
	assert.Truef(t, names["projects_read_all"], "expected projects_read_all in: %v", names)

	assert.Falsef(t, names["projects_create"], "projects_create must be filtered out: %v", names)
	assert.Falsef(t, names["projects_update"], "projects_update must be filtered out: %v", names)
	assert.Falsef(t, names["projects_delete"], "projects_delete must be filtered out: %v", names)
}

func TestMCP_Scopes_ToolsListMcpOnly(t *testing.T) {
	// Token 10 has no project scopes, so no project tools may show up.
	c := newMCPClient(t, mcpOnlyToken)
	resp := c.rpc("tools/list", map[string]any{})
	names := toolNamesFromList(t, resp)

	for _, want := range []string{
		"projects_create",
		"projects_read_one",
		"projects_read_all",
		"projects_update",
		"projects_delete",
	} {
		assert.Falsef(t, names[want], "%s must be filtered out for an mcp-only token: %v", want, names)
	}
}

func TestMCP_Scopes_ToolsListFullScopes(t *testing.T) {
	// Token 12: mcp:access + projects:* — should see all five project tools.
	c := newMCPClient(t, mcpFullProjectsToken)
	resp := c.rpc("tools/list", map[string]any{})
	names := toolNamesFromList(t, resp)

	for _, want := range []string{
		"projects_create",
		"projects_read_one",
		"projects_read_all",
		"projects_update",
		"projects_delete",
	} {
		assert.Truef(t, names[want], "expected %s in: %v", want, names)
	}
}

func TestMCP_Scopes_CallCreateForbidden(t *testing.T) {
	// Out-of-scope tools are never registered on the session, so the SDK
	// rejects the call before the dispatcher sees it.
	projectsBefore := countProjects(t)

	c := newMCPClient(t, mcpMixedScopeToken)
	resp := c.rpc("tools/call", map[string]any{
		"name":      "projects_create",
		"arguments": map[string]any{"title": "should not be created"},
	})
	requireUnknownToolError(t, resp, "projects_create")

	assert.Equal(t, projectsBefore, countProjects(t), "no project should be created when scope is denied")
}

func TestMCP_Scopes_CallNonexistentTool(t *testing.T) {
	c := newMCPClient(t, mcpFullProjectsToken)
	resp := c.rpc("tools/call", map[string]any{
		"name":      "nonexistent_tool",
		"arguments": map[string]any{},
	})
	requireUnknownToolError(t, resp, "nonexistent_tool")
}

// requireUnknownToolError pins the SDK's response for a tool the session did
// not register: a JSON-RPC "invalid params" error, not a tool result.
func requireUnknownToolError(t *testing.T, resp map[string]any, tool string) {
	t.Helper()
	require.NotContains(t, resp, "result", "unregistered tool must not produce a tool result: %v", resp)
	errObj, ok := resp["error"].(map[string]any)
	require.Truef(t, ok, "missing JSON-RPC error: %v", resp)
	assert.InDelta(t, float64(-32602), errObj["code"], 0.0001)
	assert.Contains(t, errObj["message"], `unknown tool "`+tool+`"`)
}

// countProjects verifies a denied-scope tool call did not mutate the database.
func countProjects(t *testing.T) int64 {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	n, err := s.Count(&models.Project{})
	require.NoError(t, err)
	return n
}
