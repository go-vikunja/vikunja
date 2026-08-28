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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mcpSearchUsers(t *testing.T, c *mcpClient, args map[string]any) []map[string]any {
	t.Helper()
	result := c.callTool("users_search", args)
	require.NotContains(t, result, "isError", "unexpected error: %v", result)
	var users []map[string]any
	require.NoError(t, json.Unmarshal([]byte(toolResultText(t, result)), &users))
	return users
}

func TestMCP_UsersSearch_ToolsList(t *testing.T) {
	t.Run("listed with users:read_all", func(t *testing.T) {
		c := newMCPClient(t, mcpFullProjectsToken)
		names := toolNamesFromList(t, c.rpc("tools/list", map[string]any{}))
		assert.True(t, names["users_search"])
	})
	t.Run("hidden without scope", func(t *testing.T) {
		c := newMCPClient(t, mcpMixedScopeToken)
		names := toolNamesFromList(t, c.rpc("tools/list", map[string]any{}))
		assert.False(t, names["users_search"])
		// unregistered for this session, so the SDK answers with a protocol error
		resp := c.rpc("tools/call", map[string]any{"name": "users_search", "arguments": map[string]any{"query": "user1"}})
		assert.Contains(t, resp, "error")
	})
}

func TestMCP_UsersSearch_ByUsername(t *testing.T) {
	c := newMCPClient(t, mcpFullProjectsToken)
	users := mcpSearchUsers(t, c, map[string]any{"query": "user1"})
	require.Len(t, users, 1)
	assert.Equal(t, "user1", users[0]["username"])
	assert.Empty(t, users[0]["email"], "email must never be returned")
}

func TestMCP_UsersSearch_InProject(t *testing.T) {
	c := newMCPClient(t, mcpFullProjectsToken)
	users := mcpSearchUsers(t, c, map[string]any{"query": "user1", "project_id": 1})
	require.NotEmpty(t, users)
	assert.Equal(t, "user1", users[0]["username"])
	assert.Empty(t, users[0]["email"], "email must never be returned")

	// ListUsers keeps the email when the query matches it exactly
	byEmail := mcpSearchUsers(t, c, map[string]any{"query": "user1@example.com", "project_id": 1})
	require.NotEmpty(t, byEmail)
	assert.Empty(t, byEmail[0]["email"], "email must never be returned")

	// project 20 is not accessible to user 1
	result := c.callTool("users_search", map[string]any{"query": "user1", "project_id": 20})
	assert.Equal(t, true, result["isError"])
}

func TestMCP_UsersSearch_MissingQuery(t *testing.T) {
	c := newMCPClient(t, mcpFullProjectsToken)
	result := c.callTool("users_search", map[string]any{})
	assert.Equal(t, true, result["isError"])
}
