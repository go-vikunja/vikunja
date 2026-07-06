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

func TestMCP_Tasks_ToolsListMatchesOps(t *testing.T) {
	c := newMCPClient(t, mcpFullProjectsToken)
	resp := c.rpc("tools/list", map[string]any{})
	names := toolNamesFromList(t, resp)

	for _, want := range []string{
		"tasks_create",
		"tasks_read_one",
		"tasks_read_all",
		"tasks_update",
		"tasks_delete",
	} {
		assert.Truef(t, names[want], "missing %s in tools/list: %v", want, names)
	}
}

func TestMCP_Tasks_ReadAllWithFilter(t *testing.T) {
	// Listing goes through models.TaskCollection, so the REST filter engine
	// is available: filter/sort_by/order_by plus the generic search arg.
	c := newMCPClient(t, mcpFullProjectsToken)

	result := c.callTool("tasks_read_all", map[string]any{
		"filter":  "done = true",
		"sort_by": []string{"id"},
	})
	require.NotContains(t, result, "isError", "read_all errored: %v", result)

	var tasks []map[string]any
	require.NoError(t, json.Unmarshal([]byte(toolResultText(t, result)), &tasks))
	require.NotEmpty(t, tasks, "fixtures contain done tasks")
	for _, task := range tasks {
		assert.Equal(t, true, task["done"], "filter must only return done tasks: %v", task["id"])
	}

	// Scoped to a single project via the optional project_id argument.
	result = c.callTool("tasks_read_all", map[string]any{"project_id": 1})
	require.NotContains(t, result, "isError")
	require.NoError(t, json.Unmarshal([]byte(toolResultText(t, result)), &tasks))
	require.NotEmpty(t, tasks)
	for _, task := range tasks {
		assert.InDelta(t, float64(1), task["project_id"], 0.0001, "task %v outside project 1", task["id"])
	}
}

func TestMCP_Tasks_Create(t *testing.T) {
	c := newMCPClient(t, mcpFullProjectsToken)
	result := c.callTool("tasks_create", map[string]any{
		"title":      "MCP created task",
		"project_id": 1,
	})
	require.NotContains(t, result, "isError", "create errored: %v", result)

	text := toolResultText(t, result)
	var task map[string]any
	require.NoError(t, json.Unmarshal([]byte(text), &task), "text was: %s", text)
	assert.Equal(t, "MCP created task", task["title"])
	id, ok := task["id"].(float64)
	require.Truef(t, ok, "id missing or not a number: %v", task)
	assert.Positive(t, int(id))
}

func TestMCP_Tasks_ReadOneOwned(t *testing.T) {
	c := newMCPClient(t, mcpFullProjectsToken)
	result := c.callTool("tasks_read_one", map[string]any{"id": 1})
	require.NotContains(t, result, "isError")

	text := toolResultText(t, result)
	var task map[string]any
	require.NoError(t, json.Unmarshal([]byte(text), &task))
	assert.InDelta(t, float64(1), task["id"], 0.0001)
}

func TestMCP_Tasks_ReadOneForbidden(t *testing.T) {
	// Task 34 belongs to project 20 (user 13 only); user 1 cannot see it.
	c := newMCPClient(t, mcpFullProjectsToken)
	result := c.callTool("tasks_read_one", map[string]any{"id": 34})
	isErr, _ := result["isError"].(bool)
	require.True(t, isErr, "expected isError for forbidden task, got: %v", result)
}

func TestMCP_Tasks_Update(t *testing.T) {
	c := newMCPClient(t, mcpFullProjectsToken)

	createResult := c.callTool("tasks_create", map[string]any{
		"title":      "mcp task to update",
		"project_id": 1,
	})
	require.NotContains(t, createResult, "isError")
	var created map[string]any
	require.NoError(t, json.Unmarshal([]byte(toolResultText(t, createResult)), &created))
	tid := int64(created["id"].(float64))

	updateResult := c.callTool("tasks_update", map[string]any{
		"id":          tid,
		"title":       "mcp task updated",
		"description": "Updated description",
	})
	require.NotContains(t, updateResult, "isError", "update errored: %v", updateResult)

	readResult := c.callTool("tasks_read_one", map[string]any{"id": tid})
	require.NotContains(t, readResult, "isError")
	var task map[string]any
	require.NoError(t, json.Unmarshal([]byte(toolResultText(t, readResult)), &task))
	assert.Equal(t, "mcp task updated", task["title"])
	assert.Equal(t, "Updated description", task["description"])
}

// TestMCP_Tasks_UpdateClearsDone exercises the pointer-source path of
// copyByJSONTag: a `done: false` explicitly supplied through the JSON
// args must flip a task from done back to undone.
func TestMCP_Tasks_UpdateClearsDone(t *testing.T) {
	c := newMCPClient(t, mcpFullProjectsToken)

	createResult := c.callTool("tasks_create", map[string]any{
		"title":      "mcp task to undo",
		"project_id": 1,
		"done":       true,
	})
	require.NotContains(t, createResult, "isError")
	var created map[string]any
	require.NoError(t, json.Unmarshal([]byte(toolResultText(t, createResult)), &created))
	tid := int64(created["id"].(float64))
	require.True(t, created["done"].(bool), "task should have been created in done state")

	updateResult := c.callTool("tasks_update", map[string]any{
		"id":   tid,
		"done": false,
	})
	require.NotContains(t, updateResult, "isError", "update errored: %v", updateResult)

	readResult := c.callTool("tasks_read_one", map[string]any{"id": tid})
	require.NotContains(t, readResult, "isError")
	var task map[string]any
	require.NoError(t, json.Unmarshal([]byte(toolResultText(t, readResult)), &task))
	assert.False(t, task["done"].(bool), "done must be false after explicit clear")
}

func TestMCP_Tasks_Delete(t *testing.T) {
	c := newMCPClient(t, mcpFullProjectsToken)

	createResult := c.callTool("tasks_create", map[string]any{
		"title":      "mcp task to delete",
		"project_id": 1,
	})
	require.NotContains(t, createResult, "isError")
	var created map[string]any
	require.NoError(t, json.Unmarshal([]byte(toolResultText(t, createResult)), &created))
	tid := int64(created["id"].(float64))

	deleteResult := c.callTool("tasks_delete", map[string]any{"id": tid})
	require.NotContains(t, deleteResult, "isError", "delete errored: %v", deleteResult)

	readResult := c.callTool("tasks_read_one", map[string]any{"id": tid})
	isErr, _ := readResult["isError"].(bool)
	require.True(t, isErr, "expected isError for deleted task")
}
