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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMCP_TaskAssignees_ToolsList(t *testing.T) {
	// Only three tools: create / read_all / delete. No read_one, no update.
	c := newMCPClient(t, mcpFullProjectsToken)
	resp := c.rpc("tools/list", map[string]any{})
	names := toolNamesFromList(t, resp)

	for _, want := range []string{
		"tasks_assignees_create",
		"tasks_assignees_read_all",
		"tasks_assignees_delete",
	} {
		assert.Truef(t, names[want], "missing %s in tools/list: %v", want, names)
	}

	for name := range names {
		if strings.HasPrefix(name, "tasks_assignees_") {
			assert.NotEqual(t, "tasks_assignees_read_one", name, "task_assignees has no read_one op")
			assert.NotEqual(t, "tasks_assignees_update", name, "task_assignees has no update op")
		}
	}
}

func TestMCP_TaskAssignees_ReadAllAccess(t *testing.T) {
	c := newMCPClient(t, mcpFullProjectsToken)
	result := c.callTool("tasks_assignees_read_all", map[string]any{"task_id": 30})
	require.NotContains(t, result, "isError", "read_all errored: %v", result)

	var assignees []map[string]any
	readAllItems(t, result, &assignees)

	ids := make([]float64, 0, len(assignees))
	for _, a := range assignees {
		ids = append(ids, a["id"].(float64))
		assert.Empty(t, a["email"], "read_all must not leak assignee email addresses: %v", a)
	}
	assert.ElementsMatch(t, []float64{1, 2}, ids)
}

func TestMCP_TaskAssignees_CreateAndDelete(t *testing.T) {
	// Project 34 is shared with team 1, which holds both user 1 and user 2, so user 2 passes CanRead.
	c := newMCPClient(t, mcpFullProjectsToken)

	taskRes := c.callTool("tasks_create", map[string]any{
		"title":      "task for assignee test",
		"project_id": 34,
	})
	require.NotContains(t, taskRes, "isError")
	var task map[string]any
	require.NoError(t, json.Unmarshal([]byte(toolResultText(t, taskRes)), &task))
	tid := int64(task["id"].(float64))

	assignRes := c.callTool("tasks_assignees_create", map[string]any{
		"task_id": tid,
		"user_id": 2,
	})
	require.NotContains(t, assignRes, "isError", "assign errored: %v", assignRes)

	delRes := c.callTool("tasks_assignees_delete", map[string]any{
		"task_id": tid,
		"user_id": 2,
	})
	require.NotContains(t, delRes, "isError", "delete errored: %v", delRes)

	readRes := c.callTool("tasks_assignees_read_all", map[string]any{"task_id": tid})
	require.NotContains(t, readRes, "isError")
	var assignees []map[string]any
	readAllItems(t, readRes, &assignees)
	assert.Empty(t, assignees, "assignee should be gone after delete")
}

func TestMCP_TaskAssignees_ReadAllForbidden(t *testing.T) {
	// Task 34 is in project 20, user 13's private project.
	c := newMCPClient(t, mcpFullProjectsToken)
	result := c.callTool("tasks_assignees_read_all", map[string]any{"task_id": 34})
	isErr, _ := result["isError"].(bool)
	require.True(t, isErr, "expected isError for forbidden task assignees")
}
