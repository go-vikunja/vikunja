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

// findActions calls find_action and returns the decoded action list.
func findActions(t *testing.T, c *mcpClient, args map[string]any) []map[string]any {
	t.Helper()
	result := c.callTool("find_action", args)
	require.NotContains(t, result, "isError", "find_action errored: %v", result)
	var payload struct {
		Actions []map[string]any `json:"actions"`
	}
	require.NoError(t, json.Unmarshal([]byte(toolResultText(t, result)), &payload))
	return payload.Actions
}

func TestMCP_Catalog_MetaToolsInToolsList(t *testing.T) {
	c := newMCPClient(t, mcpFullProjectsToken)
	resp := c.rpc("tools/list", map[string]any{})
	names := toolNamesFromList(t, resp)

	assert.True(t, names["find_action"], "find_action missing: %v", names)
	assert.True(t, names["do_action"], "do_action missing: %v", names)
	// Catalog actions must not appear as first-class tools.
	assert.False(t, names["tasks_labels_create"], "catalog actions must stay out of tools/list")
}

func TestMCP_Catalog_FindActionScopeFiltered(t *testing.T) {
	c := newMCPClient(t, mcpFullProjectsToken)
	actions := findActions(t, c, map[string]any{})

	names := map[string]bool{}
	for _, a := range actions {
		names[a["name"].(string)] = true
		assert.NotContains(t, a, "input_schema", "unfiltered find_action must stay schema-free")
	}
	// Exact set, derived from token 12's scopes, so this can't silently drift when the fixture changes.
	assert.Equal(t, map[string]bool{
		"tasks_labels_create":     true,
		"tasks_labels_read_all":   true,
		"tasks_labels_delete":     true,
		"projects_views_read_all": true,
	}, names)
	assert.False(t, names["tasks_create"], "typed tools must not appear in the catalog")
}

func TestMCP_Catalog_FindActionReturnsSchemas(t *testing.T) {
	c := newMCPClient(t, mcpFullProjectsToken)
	actions := findActions(t, c, map[string]any{"resource": "tasks_labels"})
	require.Len(t, actions, 3)

	for _, a := range actions {
		schema, ok := a["input_schema"].(map[string]any)
		require.Truef(t, ok, "action %v missing input_schema", a["name"])
		props, ok := schema["properties"].(map[string]any)
		require.True(t, ok)
		assert.Contains(t, props, "task_id")
	}
}

func TestMCP_Catalog_FindActionEmptyWithoutScopes(t *testing.T) {
	c := newMCPClient(t, mcpOnlyToken)
	actions := findActions(t, c, map[string]any{})
	assert.Empty(t, actions)
}

func TestMCP_Catalog_DoActionLabelRoundTrip(t *testing.T) {
	c := newMCPClient(t, mcpFullProjectsToken)

	// Attach label 1 (owned by user 1) to task 1, then read it back.
	result := c.callTool("do_action", map[string]any{
		"action":    "tasks_labels_create",
		"arguments": map[string]any{"task_id": 1, "label_id": 1},
	})
	require.NotContains(t, result, "isError", "do_action create errored: %v", result)

	result = c.callTool("do_action", map[string]any{
		"action":    "tasks_labels_read_all",
		"arguments": map[string]any{"task_id": 1},
	})
	require.NotContains(t, result, "isError")
	var labels []map[string]any
	readAllItems(t, result, &labels)
	ids := map[float64]bool{}
	for _, l := range labels {
		ids[l["id"].(float64)] = true
	}
	assert.True(t, ids[1], "label 1 should be attached: %v", labels)

	result = c.callTool("do_action", map[string]any{
		"action":    "tasks_labels_delete",
		"arguments": map[string]any{"task_id": 1, "label_id": 1},
	})
	require.NotContains(t, result, "isError", "do_action delete errored: %v", result)
}

func TestMCP_Catalog_DoActionListsProjectViews(t *testing.T) {
	c := newMCPClient(t, mcpFullProjectsToken)
	result := c.callTool("do_action", map[string]any{
		"action":    "projects_views_read_all",
		"arguments": map[string]any{"project_id": 1},
	})
	require.NotContains(t, result, "isError", "do_action projects_views_read_all errored: %v", result)

	var views []map[string]any
	readAllItems(t, result, &views)
	require.NotEmpty(t, views)
	for _, v := range views {
		assert.EqualValues(t, 1, v["project_id"])
	}
}

func TestMCP_Catalog_DoActionScopeDenied(t *testing.T) {
	// Token 12 has no projects_users scope; the per-call re-check inside
	// Dispatch must reject the action even though it exists.
	c := newMCPClient(t, mcpFullProjectsToken)
	result := c.callTool("do_action", map[string]any{
		"action":    "projects_users_create",
		"arguments": map[string]any{"project_id": 1, "username": "user2"},
	})
	isErr, _ := result["isError"].(bool)
	require.True(t, isErr, "expected scope denial: %v", result)
	assert.Contains(t, toolResultText(t, result), "not authorized")
}

func TestMCP_Catalog_DoActionUnknownAction(t *testing.T) {
	c := newMCPClient(t, mcpFullProjectsToken)
	result := c.callTool("do_action", map[string]any{
		"action":    "nonexistent_create",
		"arguments": map[string]any{},
	})
	isErr, _ := result["isError"].(bool)
	require.True(t, isErr, "expected tool-not-found: %v", result)
}

func TestMCP_Catalog_DoActionValidatesArguments(t *testing.T) {
	// Schema validation happens inside Dispatch, so this surfaces as an isError result.
	c := newMCPClient(t, mcpFullProjectsToken)
	result := c.callTool("do_action", map[string]any{
		"action":    "tasks_labels_create",
		"arguments": map[string]any{"task_id": 1},
	})
	isErr, _ := result["isError"].(bool)
	require.True(t, isErr, "expected validation error: %v", result)
}

// Both meta-tools declare additionalProperties:false, so a typo has to be named
// back to the caller instead of silently changing what the call does.
func TestMCP_Catalog_MetaToolsRejectUnknownArguments(t *testing.T) {
	c := newMCPClient(t, mcpFullProjectsToken)

	result := c.callTool("find_action", map[string]any{"resource_name": "tasks_labels"})
	require.Equal(t, true, result["isError"], "expected isError: %v", result)
	assert.Contains(t, toolResultText(t, result), "resource_name")

	result = c.callTool("do_action", map[string]any{"action": "tasks_labels_read_all", "argumentz": map[string]any{}})
	require.Equal(t, true, result["isError"], "expected isError: %v", result)
	assert.Contains(t, toolResultText(t, result), "argumentz")
}
