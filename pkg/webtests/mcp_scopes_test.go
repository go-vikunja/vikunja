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

// Token 10 has {mcp:access, projects:[read_one, read_all]} — a partial scope
// for the scope-filtered tools/list and tools/call tests.
const mcpMixedScopeToken = "tk_mcp_mixed_scope_token_test_00mcpmixed02"

// toolNamesFromList extracts the "name" field from every tool in a tools/list
// result payload.
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
	// Token 10: projects:[read_one, read_all] — should see exactly those two
	// project tools and no others.
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
	// Token 9: only {mcp:access} — no project scopes, so no project tools
	// must show in tools/list.
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
	// Token 11: mcp:access + projects:* — should see all five project tools.
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
	// Token 10 lacks projects:create. Calling projects_create must come back
	// as an error response without writing to the database. The SDK may
	// return either a JSON-RPC protocol error (tool not found, because the
	// tool wasn't registered for this session's server) or a tool result
	// with isError=true (if the dispatcher's defensive scope check ran).
	// Both are valid — what matters is that no DB write happened.
	projectsBefore := countProjects(t)

	c := newMCPClient(t, mcpMixedScopeToken)
	resp := c.rpc("tools/call", map[string]any{
		"name":      "projects_create",
		"arguments": map[string]any{"title": "should not be created"},
	})

	// Either a JSON-RPC error or a tool result with isError=true is
	// acceptable; what matters is no DB write.
	if _, hasErr := resp["error"]; !hasErr {
		result, ok := resp["result"].(map[string]any)
		require.Truef(t, ok, "missing result: %v", resp)
		isErr, _ := result["isError"].(bool)
		assert.Truef(t, isErr, "expected isError for forbidden create: %v", result)
	}

	projectsAfter := countProjects(t)
	assert.Equal(t, projectsBefore, projectsAfter, "no project should be created when scope is denied")
}

func TestMCP_Scopes_CallNonexistentTool(t *testing.T) {
	// An unknown tool name must result in an error tool call result (or a
	// JSON-RPC error from the SDK saying "tool not found"). Either way, the
	// caller sees a failure, not a JSON-parse 500.
	c := newMCPClient(t, mcpFullProjectsToken)
	resp := c.rpc("tools/call", map[string]any{
		"name":      "nonexistent_tool",
		"arguments": map[string]any{},
	})

	if _, hasErr := resp["error"]; hasErr {
		return // SDK returned a JSON-RPC error — acceptable.
	}
	result, ok := resp["result"].(map[string]any)
	require.Truef(t, ok, "missing both error and result: %v", resp)
	isErr, _ := result["isError"].(bool)
	assert.Truef(t, isErr, "expected isError for nonexistent tool: %v", result)
}

// countProjects returns the number of rows in the projects table. Used to
// verify that a denied-scope tool call did not mutate the database.
func countProjects(t *testing.T) int64 {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	n, err := s.Count(&models.Project{})
	require.NoError(t, err)
	return n
}
