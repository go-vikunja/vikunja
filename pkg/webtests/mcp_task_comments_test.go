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
	// task_id has no omitempty in TaskCommentCreateInput, so omitting it
	// must surface as either a schema-level error or a tool result with
	// isError=true (the task_id=0 path would dereference an invalid task).
	c := newMCPClient(t, mcpFullProjectsToken)
	resp := c.rpc("tools/call", map[string]any{
		"name":      "tasks_comments_create",
		"arguments": map[string]any{"comment": "missing task id"},
	})
	if _, hasErr := resp["error"]; hasErr {
		return
	}
	result, ok := resp["result"].(map[string]any)
	require.Truef(t, ok, "missing result: %v", resp)
	isErr, _ := result["isError"].(bool)
	assert.Truef(t, isErr, "expected isError for missing task_id: %v", result)
}

func TestMCP_TaskComments_ReadAll(t *testing.T) {
	c := newMCPClient(t, mcpFullProjectsToken)
	result := c.callTool("tasks_comments_read_all", map[string]any{"task_id": 1})
	require.NotContains(t, result, "isError")

	text := toolResultText(t, result)
	var comments []map[string]any
	require.NoError(t, json.Unmarshal([]byte(text), &comments))
	// Fixture task 1 has at least one comment.
	require.NotEmpty(t, comments)
}

func TestMCP_TaskComments_ReadAllForbidden(t *testing.T) {
	// Task 34 belongs to project 20, only user 13 has access. User 1
	// cannot see its comments.
	c := newMCPClient(t, mcpFullProjectsToken)
	result := c.callTool("tasks_comments_read_all", map[string]any{"task_id": 34})
	isErr, _ := result["isError"].(bool)
	require.True(t, isErr, "expected isError for forbidden task comments, got: %v", result)
}

func TestMCP_TaskComments_DisabledByConfig(t *testing.T) {
	// Flip ServiceEnableTaskComments off, build a new session, ensure the
	// comment tools disappear from tools/list.
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
}
