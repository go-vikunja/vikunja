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

	"code.vikunja.io/api/pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMCP_TaskWebLinks(t *testing.T) {
	c := newMCPClient(t, mcpFullToken)
	var task map[string]any
	toolResultJSON(t, c.callTool("tasks_read", map[string]any{
		"task":   1,
		"fields": []string{"title"},
		"web_ui_bases": []map[string]string{
			{"label": "public", "url": "https://example.org/vikunja/"},
			{"label": "vpn", "url": "http://100.64.0.1:3456/"},
		},
	}), &task)
	links := task["web_links"].(map[string]any)
	assert.InDelta(t, 1, links["id"], 0.001)
	urls := links["urls"].([]any)
	require.Len(t, urls, 2)
	assert.Equal(t, "https://example.org/vikunja/tasks/1", urls[0].(map[string]any)["url"])
	assert.NotContains(t, task, "id")
	toolResultJSON(t, c.callTool("tasks_read", map[string]any{"task": 1}), &task)
	assert.Equal(t, "missing_base_url", task["web_links"].(map[string]any)["status"])
	denied := newMCPClientOn(t, c.e, mcpProjectsReadToken)
	assert.NotContains(t, denied.toolNames(), "tasks_read")
	assert.Contains(t, denied.rpc("tools/call", map[string]any{
		"name":      "tasks_read",
		"arguments": map[string]any{"task": 1},
	}), "error")
}

func TestMCP_ShareLinks(t *testing.T) {
	c := newAllScopesMCPClient(t)
	actions := findActions(t, c, map[string]any{"resource": "shares"})
	require.Len(t, actions, 4)
	denied := newMCPClientOn(t, c.e, mcpFullToken)
	assert.Empty(t, findActions(t, denied, map[string]any{"resource": "shares"}))
	for _, action := range actions {
		name := action["name"].(string)
		assert.NotNil(t, action["annotations"])
		assert.Equal(t, true, denied.callTool("do_action", map[string]any{
			"action":    name,
			"arguments": map[string]any{"project": 1},
		})["isError"])
	}
	call := func(name string, args map[string]any) map[string]any {
		return c.callTool("do_action", map[string]any{
			"action":    name,
			"arguments": args,
		})
	}
	var share map[string]any
	toolResultJSON(t, call("shares_create", map[string]any{
		"project":      1,
		"permission":   0,
		"password":     "test-share-password",
		"web_ui_bases": []map[string]string{{"label": "public", "url": "https://example.org/prefix/"}},
	}), &share)
	id := share["id"]
	assert.Empty(t, share["password"])
	assert.NotEmpty(t, share["hash"])
	links := share["web_links"].(map[string]any)
	assert.Equal(t, "project_share", links["kind"])
	assert.InDelta(t, 1, links["project_id"], 0.001)
	urls := links["urls"].([]any)
	assert.Equal(t, "https://example.org/prefix/#share-auth-token="+share["hash"].(string), urls[0].(map[string]any)["url"])
	toolResultJSON(t, call("shares_read", map[string]any{"project": 1, "share": id}), &share)
	assert.Equal(t, "missing_base_url", share["web_links"].(map[string]any)["status"])
	assert.Equal(t, true, call("shares_read", map[string]any{"project": 2, "share": id})["isError"])
	require.NotContains(t, call("shares_delete", map[string]any{"project": 1, "share": id}), "isError")
	assert.Equal(t, true, call("shares_read", map[string]any{"project": 1, "share": id})["isError"])
}

func TestMCP_SharesDisabled(t *testing.T) {
	config.ServiceEnableLinkSharing.Set(false)
	t.Cleanup(func() { config.ServiceEnableLinkSharing.Set(true) })
	c := newAllScopesMCPClient(t)
	assert.Empty(t, findActions(t, c, map[string]any{"resource": "shares"}))
	result := c.callTool("do_action", map[string]any{
		"action":    "shares_create",
		"arguments": map[string]any{"project": 1, "permission": 0},
	})
	assert.Equal(t, true, result["isError"])
}

func TestMCP_SharesRequireProjectAdmin(t *testing.T) {
	c := newAllScopesMCPClient(t)
	for _, project := range []int64{9, 10, 20} {
		for _, action := range []string{"shares_list", "shares_create"} {
			result := c.callTool("do_action", map[string]any{
				"action":    action,
				"arguments": map[string]any{"project": project},
			})
			assert.Equal(t, true, result["isError"])
			assert.Contains(t, toolResultText(t, result), "403")
		}
	}
}
