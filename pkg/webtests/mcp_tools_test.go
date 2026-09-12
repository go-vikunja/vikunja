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
	"fmt"
	"net/http"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMCP_Tools_TaskLifecycle(t *testing.T) {
	c := newMCPClient(t, mcpFullToken)
	var task map[string]any
	toolResultJSON(t, c.callTool("tasks_create", map[string]any{
		"project":     1,
		"title":       "mcp task",
		"description": "**keep me**",
		"priority":    4,
		"hex_color":   "ff8800",
	}), &task)
	id := int64(task["id"].(float64))
	assert.Equal(t, "**keep me**", task["description"])
	require.NotContains(t, c.callTool("tasks_update", map[string]any{
		"projecttask": id,
		"done":        true,
	}), "isError")
	toolResultJSON(t, c.callTool("tasks_read", map[string]any{"projecttask": id}), &task)
	assert.Equal(t, true, task["done"])
	assert.Equal(t, "mcp task", task["title"])
	assert.Equal(t, "**keep me**", task["description"])
	assert.InDelta(t, 4, task["priority"], 0.001)
	assert.Equal(t, "ff8800", task["hex_color"])
	require.NotContains(t, c.callTool("tasks_update", map[string]any{
		"projecttask": id,
		"done":        false,
	}), "isError")
	toolResultJSON(t, c.callTool("tasks_read", map[string]any{"projecttask": id}), &task)
	assert.Equal(t, false, task["done"])
	require.NotContains(t, c.callTool("tasks_delete", map[string]any{"projecttask": id}), "isError")
	gone := c.callTool("tasks_read", map[string]any{"projecttask": id})
	assert.Equal(t, true, gone["isError"])
	assert.Contains(t, toolResultText(t, gone), "404")
}
func TestMCP_Tools_UnchangedUpdateIsNotAnError(t *testing.T) {
	c := newMCPClient(t, mcpFullToken)
	args := map[string]any{
		"projecttask": 1,
		"title":       "same title",
	}
	first := c.callTool("tasks_update", args)
	require.NotContains(t, first, "isError", toolResultText(t, first))
	second := c.callTool("tasks_update", args)
	require.NotContains(t, second, "isError", toolResultText(t, second))
	assert.Contains(t, toolResultText(t, second), `"unchanged":true`)
}
func TestMCP_Tools_NullClearsAndUnsets(t *testing.T) {
	c := newMCPClient(t, mcpFullToken)
	var task map[string]any
	toolResultJSON(t, c.callTool("tasks_create", map[string]any{
		"project":  1,
		"title":    "mcp due",
		"due_date": "2030-01-01T12:00:00Z",
	}), &task)
	id := int64(task["id"].(float64))
	cleared := c.callTool("tasks_update", map[string]any{
		"projecttask": id,
		"due_date":    nil,
	})
	require.NotContains(t, cleared, "isError", toolResultText(t, cleared))
	toolResultJSON(t, c.callTool("tasks_read", map[string]any{"projecttask": id}), &task)
	assert.NotContains(t, task["due_date"], "2030")
	listed := c.callTool("tasks_list", map[string]any{"filter": nil})
	assert.NotContains(t, listed, "isError", toolResultText(t, listed))
}
func TestMCP_Tools_RecordsTokenUsagePerCall(t *testing.T) {
	config.AuditEnabled.Set(true)
	t.Cleanup(func() { config.AuditEnabled.Set(false) })
	c := newMCPClient(t, mcpFullToken)
	events.ClearDispatchedEvents()
	rec := c.post(fmt.Sprintf(`[%s,%s]`,
		`{"jsonrpc":"2.0","id":200,"method":"tools/call","params":{"name":"projects_read","arguments":{"id":1}}}`,
		`{"jsonrpc":"2.0","id":201,"method":"tools/call","params":{"name":"projects_read","arguments":{"id":2}}}`,
	))
	require.Equal(t, http.StatusOK, rec.Code, "%s", rec.Body.String())
	// One usage for the MCP request itself plus one per tool call.
	assert.Equal(t, 3, events.CountDispatchedEvents((&models.APITokenUsedEvent{}).Name()))
}
func TestMCP_Tools_ListEnvelopeAndFilter(t *testing.T) {
	c := newMCPClient(t, mcpFullToken)
	var env map[string]any
	toolResultJSON(t, c.callTool("tasks_list", map[string]any{
		"filter":   "done = false",
		"per_page": 2,
		"page":     1,
	}), &env)
	for _, k := range []string{
		"items",
		"total",
		"page",
		"per_page",
		"total_pages",
	} {
		assert.Contains(t, env, k)
	}
	assert.LessOrEqual(t, len(env["items"].([]any)), 2)
}
func TestMCP_Tools_CommentLifecycle(t *testing.T) {
	c := newMCPClient(t, mcpFullToken)
	var comment map[string]any
	toolResultJSON(t, c.callTool("task_comments_create", map[string]any{
		"task":    1,
		"comment": "from mcp",
	}), &comment)
	assert.Equal(t, "from mcp", comment["comment"])
	assert.NotContains(t, comment["author"].(map[string]any), "email")
	id := int64(comment["id"].(float64))
	res := c.callTool("task_comments_update", map[string]any{
		"task":      1,
		"commentid": id,
		"comment":   "edited",
	})
	require.NotContains(t, res, "isError", toolResultText(t, res))
	toolResultJSON(t, c.callTool("task_comments_read", map[string]any{
		"task":      1,
		"commentid": id,
	}), &comment)
	assert.Equal(t, "edited", comment["comment"])
}
func assigneeIDs(t *testing.T, c *mcpClient) []int64 {
	t.Helper()
	var assignees []struct {
		ID int64 `json:"id"`
	}
	readAllItems(t, c.callTool("task_assignees_list", map[string]any{"projecttask": 1}), &assignees)
	ids := make([]int64, 0, len(assignees))
	for _, a := range assignees {
		ids = append(ids, a.ID)
	}
	return ids
}
func TestMCP_Tools_AssigneeAddRemove(t *testing.T) {
	c := newMCPClient(t, mcpFullToken)
	require.NotContains(t, c.callTool("task_assignees_create", map[string]any{
		"projecttask": 1,
		"user_id":     1,
	}), "isError")
	assert.Contains(t, assigneeIDs(t, c), int64(1))
	require.NotContains(t, c.callTool("task_assignees_delete", map[string]any{
		"projecttask": 1,
		"user":        1,
	}), "isError")
	assert.NotContains(t, assigneeIDs(t, c), int64(1))
}
func TestMCP_Tools_UsersSearchStripsEmail(t *testing.T) {
	c := newMCPClient(t, mcpFullToken)
	var users []map[string]any
	readAllItems(t, c.callTool("users_search", map[string]any{"q": "user2"}), &users)
	require.NotEmpty(t, users)
	for _, u := range users {
		assert.Empty(t, u["email"])
	}
}
func TestMCP_Tools_ForbiddenIsError(t *testing.T) {
	c := newMCPClient(t, mcpFullToken)
	res := c.callTool("projects_read", map[string]any{"id": 20})
	assert.Equal(t, true, res["isError"])
	assert.Contains(t, toolResultText(t, res), "403")
}
func TestMCP_Tools_ScopeDeniedIsError(t *testing.T) {
	c := newMCPClient(t, mcpFullToken)
	// do_action checks scopes per call, so actions outside the token's scopes still fail.
	res := c.callTool("do_action", map[string]any{
		"action": "project_views_update",
		"arguments": map[string]any{
			"project": 1,
			"view":    1,
			"title":   "nope",
		},
	})
	assert.Equal(t, true, res["isError"])
	assert.Contains(t, toolResultText(t, res), "not authorized")
}
func TestMCP_Tools_ValidationIsError(t *testing.T) {
	c := newMCPClient(t, mcpFullToken)
	res := c.callTool("tasks_create", map[string]any{
		"project": 1,
		"title":   "",
	})
	assert.Equal(t, true, res["isError"])
	text := toolResultText(t, res)
	assert.Contains(t, text, "invalid arguments")
	assert.Contains(t, text, "title")
}
func TestMCP_Tools_UnknownArgumentIsError(t *testing.T) {
	c := newMCPClient(t, mcpFullToken)
	res := c.callTool("tasks_read", map[string]any{
		"projecttask": 1,
		"bogus":       true,
	})
	assert.Equal(t, true, res["isError"])
	assert.Contains(t, toolResultText(t, res), "bogus")
}
func TestMCP_Tools_CreateSchemaMarksRequired(t *testing.T) {
	c := newMCPClient(t, mcpFullToken)
	resp := c.rpc("tools/list", map[string]any{})
	for _, raw := range resp["result"].(map[string]any)["tools"].([]any) {
		tl := raw.(map[string]any)
		if tl["name"] != "tasks_create" {
			continue
		}
		schema := tl["inputSchema"].(map[string]any)
		assert.ElementsMatch(t, []any{
			"project",
			"title",
		}, schema["required"])
		props := schema["properties"].(map[string]any)
		assert.NotContains(t, props, "created_by")
		assert.NotContains(t, props, "id")
		assert.NotContains(t, props, "project_id")
		assert.Contains(t, props, "due_date")
		return
	}
	t.Fatal("tasks_create not listed")
}

func TestMCP_Tools_RESTValidationIsError(t *testing.T) {
	c := newMCPClient(t, mcpFullToken)
	res := c.callTool("tasks_create", map[string]any{
		"project":      1,
		"title":        "invalid repeat",
		"repeat_after": -1,
	})
	assert.Equal(t, true, res["isError"])
	assert.Contains(t, toolResultText(t, res), "422")
	assert.Contains(t, toolResultText(t, res), "repeat_after")
}
func TestMCP_Tools_ExpandRequiresScopeOnLoopback(t *testing.T) {
	c := newMCPClient(t, mcpFullToken)
	allowed := c.callTool("tasks_read", map[string]any{
		"projecttask": 1,
		"expand":      []string{"comments"},
	})
	require.NotContains(t, allowed, "isError", toolResultText(t, allowed))
	denied := c.callTool("tasks_read", map[string]any{
		"projecttask": 1,
		"expand":      []string{"reactions"},
	})
	assert.Equal(t, true, denied["isError"])
	assert.Equal(t, `401: {"code":11,"message":"missing, malformed, expired or otherwise invalid token provided"} — the API token lacks a scope required by this call (check route and expand scopes)`, toolResultText(t, denied))
}
func TestMCP_Tools_ProjectAndLabelLifecycle(t *testing.T) {
	for _, resource := range []string{
		"projects",
		"labels",
	} {
		t.Run(resource, func(t *testing.T) {
			c := newMCPClient(t, mcpFullToken)
			var item map[string]any
			toolResultJSON(t, c.callTool(resource+"_create", map[string]any{
				"title":     "mcp item",
				"hex_color": "ff8800",
			}), &item)
			id := item["id"]
			updated := c.callTool(resource+"_update", map[string]any{
				"id":    id,
				"title": "edited",
			})
			require.NotContains(t, updated, "isError", toolResultText(t, updated))
			toolResultJSON(t, c.callTool(resource+"_read", map[string]any{"id": id}), &item)
			assert.Equal(t, "edited", item["title"])
			assert.Equal(t, "ff8800", item["hex_color"])
			require.NotContains(t, c.callTool(resource+"_delete", map[string]any{"id": id}), "isError")
		})
	}
}

func TestMCP_Tools_LoopbackUsesTheAuthorisedToken(t *testing.T) {
	c := newMCPClient(t, mcpFullToken)
	s := db.NewSession()
	defer s.Close()
	u, err := user.GetUserByID(s, 1)
	require.NoError(t, err)
	jwt, err := auth.NewUserJWTAuthtoken(u, "test-session-id")
	require.NoError(t, err)
	c.authorizations = []string{
		"Bearer " + jwt,
		"Bearer " + mcpFullToken,
	}
	denied := c.callTool("tasks_read", map[string]any{
		"projecttask": 1,
		"expand":      []string{"reactions"},
	})
	assert.Equal(t, true, denied["isError"])
	assert.Equal(t, `401: {"code":11,"message":"missing, malformed, expired or otherwise invalid token provided"} — the API token lacks a scope required by this call (check route and expand scopes)`, toolResultText(t, denied))
}
