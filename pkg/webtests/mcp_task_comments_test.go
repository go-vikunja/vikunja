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
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMCP_TaskComments_ToolsListAll(t *testing.T) {
	c := newMCPClient(t, mcpFullProjectsToken)
	resp := c.rpc("tools/list", map[string]any{})
	names := toolNamesFromList(t, resp)

	for _, want := range []string{
		"tasks_comments_create",
		"tasks_comments_read_one",
		"tasks_comments_read_all",
		"tasks_comments_update",
		"tasks_comments_delete",
	} {
		assert.Truef(t, names[want], "missing %s in tools/list: %v", want, names)
	}
}

func TestMCP_TaskComments_Create(t *testing.T) {
	c := newMCPClient(t, mcpFullProjectsToken)
	result := c.callTool("tasks_comments_create", map[string]any{
		"task_id": 1,
		"comment": "mcp comment",
	})
	require.NotContains(t, result, "isError", "create errored: %v", result)

	text := toolResultText(t, result)
	var comment map[string]any
	require.NoError(t, json.Unmarshal([]byte(text), &comment))
	assert.Equal(t, "mcp comment", comment["comment"])
	id, ok := comment["id"].(float64)
	require.Truef(t, ok, "id missing: %v", comment)
	assert.Positive(t, int(id))
}

func TestMCP_TaskComments_CreateMissingTaskID(t *testing.T) {
	// task_id has no omitempty in TaskCommentCreateInput, so schema
	// validation rejects the call before it can dereference task 0.
	c := newMCPClient(t, mcpFullProjectsToken)
	result := c.callTool("tasks_comments_create", map[string]any{"comment": "missing task id"})
	isErr, _ := result["isError"].(bool)
	require.Truef(t, isErr, "expected isError for missing task_id: %v", result)
	assert.Contains(t, toolResultText(t, result), "mcp: invalid arguments for tasks_comments_create")
}

func TestMCP_TaskComments_ReadAll(t *testing.T) {
	c := newMCPClient(t, mcpFullProjectsToken)
	result := c.callTool("tasks_comments_read_all", map[string]any{"task_id": 1})
	require.NotContains(t, result, "isError")

	var comments []map[string]any
	readAllItems(t, result, &comments)
	// Fixture task 1 has at least one comment.
	require.NotEmpty(t, comments)
}

func TestMCP_TaskComments_ReadAllForbidden(t *testing.T) {
	// Task 34 belongs to project 20, which only user 13 can access.
	c := newMCPClient(t, mcpFullProjectsToken)
	result := c.callTool("tasks_comments_read_all", map[string]any{"task_id": 34})
	isErr, _ := result["isError"].(bool)
	require.True(t, isErr, "expected isError for forbidden task comments, got: %v", result)
}

func TestMCP_TaskComments_DisabledByConfig(t *testing.T) {
	// The gate is read per session, so a new client must not see the comment tools.
	original := config.ServiceEnableTaskComments.GetBool()
	config.ServiceEnableTaskComments.Set(false)
	t.Cleanup(func() { config.ServiceEnableTaskComments.Set(original) })

	c := newMCPClient(t, mcpFullProjectsToken)
	resp := c.rpc("tools/list", map[string]any{})
	names := toolNamesFromList(t, resp)

	for name := range names {
		assert.Falsef(t, strings.HasPrefix(name, "tasks_comments_"),
			"tasks_comments_* tool must be absent when comments are disabled: %s", name)
	}

	// do_action names tools directly, bypassing tools/list, so the gate has
	// to be re-checked in the dispatcher.
	result := c.callTool("do_action", map[string]any{
		"action":    "tasks_comments_create",
		"arguments": map[string]any{"task_id": 1, "comment": "must not be created"},
	})
	require.Equal(t, true, result["isError"], "do_action must not reach a disabled resource: %v", result)
	assert.Contains(t, toolResultText(t, result), "mcp: tool not found: tasks_comments_create")
}

func TestMCP_TaskComments_CreateRejectsEmptyComment(t *testing.T) {
	// TaskComment.Comment is valid:"required"; MCP must enforce it like the REST layer does.
	c := newMCPClient(t, mcpFullProjectsToken)
	result := c.callTool("tasks_comments_create", map[string]any{
		"task_id": 1,
		"comment": "",
	})
	require.Equal(t, true, result["isError"], "expected isError: %v", result)
	assert.Contains(t, toolResultText(t, result), "comment")
}
