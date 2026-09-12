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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func findActions(t *testing.T, c *mcpClient, args map[string]any) []map[string]any {
	t.Helper()
	var out struct {
		Actions []map[string]any `json:"actions"`
	}
	toolResultJSON(t, c.callTool("find_action", args), &out)
	return out.Actions
}
func TestMCP_Catalog_FindActionFollowsScopes(t *testing.T) {
	var names []string
	for _, a := range findActions(t, newMCPClient(t, mcpFullToken), map[string]any{}) {
		names = append(names, a["name"].(string))
		assert.Nil(t, a["input_schema"])
	}
	assert.Contains(t, names, "task_labels_create")
	assert.Contains(t, names, "project_views_list")
	for _, name := range []string{
		"teams_members_add",
		"tasks_create",
		"tokens_create",
	} {
		assert.NotContains(t, names, name)
	}
	assert.NotContains(t, newMCPClient(t, mcpOnlyToken).toolNames(), "find_action")
}
func TestMCP_Catalog_FindActionDescriptionListsAuthorizedAreas(t *testing.T) {
	desc := newMCPClient(t, mcpFullToken).toolDescription("find_action")
	assert.Contains(t, desc, "project_views")
	assert.NotContains(t, desc, "teams")
}
func TestMCP_Catalog_FindActionReturnsSchemas(t *testing.T) {
	c := newMCPClient(t, mcpFullToken)
	actions := findActions(t, c, map[string]any{"resource": "task_labels"})
	require.Len(t, actions, 3)
	for _, a := range actions {
		assert.NotNil(t, a["input_schema"])
	}
	one := findActions(t, c, map[string]any{"action": "project_views_list"})
	require.Len(t, one, 1)
	assert.Contains(t, one[0]["input_schema"].(map[string]any)["properties"], "project")
}
func TestMCP_Catalog_DoActionRoundTrip(t *testing.T) {
	c := newMCPClient(t, mcpFullToken)
	res := c.callTool("do_action", map[string]any{
		"action": "task_labels_create",
		"arguments": map[string]any{
			"projecttask": 1,
			"label_id":    7,
		},
	})
	require.NotContains(t, res, "isError", toolResultText(t, res))
	var env map[string]any
	toolResultJSON(t, c.callTool("do_action", map[string]any{
		"action":    "task_labels_list",
		"arguments": map[string]any{"projecttask": 1},
	}), &env)
	assert.NotEmpty(t, env["items"])
	res = c.callTool("do_action", map[string]any{
		"action": "task_labels_delete",
		"arguments": map[string]any{
			"projecttask": 1,
			"label":       7,
		},
	})
	require.NotContains(t, res, "isError", toolResultText(t, res))
}
func TestMCP_Catalog_DoActionCannotEscalate(t *testing.T) {
	c := newMCPClient(t, mcpFullToken)
	for _, action := range []string{
		"teams_members_add",
		"tokens_create",
		"nope",
	} {
		res := c.callTool("do_action", map[string]any{
			"action":    action,
			"arguments": map[string]any{},
		})
		assert.Equal(t, true, res["isError"], action)
	}
	res := c.callTool("do_action", map[string]any{
		"action": "task_labels_create",
		"arguments": map[string]any{
			"projecttask": 1,
			"bogus":       1,
		},
	})
	assert.Equal(t, true, res["isError"])
	assert.Contains(t, toolResultText(t, res), "bogus")
}
