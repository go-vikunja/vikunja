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

func TestMCP_Teams_ToolsListAll(t *testing.T) {
	c := newMCPClient(t, mcpFullProjectsToken)
	resp := c.rpc("tools/list", map[string]any{})
	names := toolNamesFromList(t, resp)

	for _, want := range []string{
		"teams_create",
		"teams_read_one",
		"teams_read_all",
		"teams_update",
		"teams_delete",
	} {
		assert.Truef(t, names[want], "missing %s in tools/list: %v", want, names)
	}
}

func TestMCP_Teams_Create(t *testing.T) {
	c := newMCPClient(t, mcpFullProjectsToken)
	result := c.callTool("teams_create", map[string]any{
		"name":        "mcp team",
		"description": "Team created via mcp",
	})
	require.NotContains(t, result, "isError", "create errored: %v", result)

	text := toolResultText(t, result)
	var team map[string]any
	require.NoError(t, json.Unmarshal([]byte(text), &team))
	assert.Equal(t, "mcp team", team["name"])
	id, ok := team["id"].(float64)
	require.Truef(t, ok, "id missing: %v", team)
	assert.Positive(t, int(id))
}

func TestMCP_Teams_ReadAll(t *testing.T) {
	c := newMCPClient(t, mcpFullProjectsToken)
	result := c.callTool("teams_read_all", map[string]any{})
	require.NotContains(t, result, "isError")

	text := toolResultText(t, result)
	var teams []map[string]any
	require.NoError(t, json.Unmarshal([]byte(text), &teams))
	// User 1 created several testteam* teams (fixtures).
	require.NotEmpty(t, teams)
}

func TestMCP_Teams_ReadOneForbidden(t *testing.T) {
	// User 1 is a member of teams 1..8 (see team_members.yml fixture).
	// Team 9 is owned by user 7 with no user-1 membership row, so user 1
	// must not be able to read it.
	c := newMCPClient(t, mcpFullProjectsToken)
	result := c.callTool("teams_read_one", map[string]any{"id": 9})
	isErr, _ := result["isError"].(bool)
	require.True(t, isErr, "expected isError for inaccessible team, got: %v", result)
}
