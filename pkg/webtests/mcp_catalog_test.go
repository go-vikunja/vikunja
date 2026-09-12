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
	"slices"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/license"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
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

// Grants every scope the /routes endpoint offers, so the tool lists it sees are the complete ones.
func newAllScopesMCPClient(t *testing.T) *mcpClient {
	t.Helper()
	e, err := setupTestEnv()
	require.NoError(t, err)
	license.SetForTests([]license.Feature{
		license.FeatureAdminPanel,
		license.FeatureTimeTracking,
	})
	t.Cleanup(license.ResetForTests)
	permissions := models.APIPermissions{}
	for group, perms := range models.GetAPITokenRoutes() {
		for perm := range perms {
			permissions[group] = append(permissions[group], perm)
		}
	}
	s := db.NewSession()
	defer s.Close()
	owner, err := user.GetUserByID(s, 1)
	require.NoError(t, err)
	token := &models.APIToken{
		Title:          "all scopes",
		APIPermissions: permissions,
		ExpiresAt:      time.Now().Add(time.Hour),
	}
	require.NoError(t, token.Create(s, owner))
	require.NoError(t, s.Commit())
	return newMCPClientOn(t, e, token.Token)
}

// A newly registered v2 operation must be classified as typed or catalog here.
func TestMCP_Catalog_ToolListsArePinned(t *testing.T) {
	c := newAllScopesMCPClient(t)
	var catalog []string
	for _, a := range findActions(t, c, map[string]any{}) {
		catalog = append(catalog, a["name"].(string))
	}
	slices.Sort(catalog)
	assert.Equal(t, []string{
		"buckets_create",
		"buckets_delete",
		"buckets_list",
		"buckets_update",
		"filters_create",
		"filters_delete",
		"filters_read",
		"filters_update",
		"notifications_delete_all",
		"notifications_list",
		"notifications_mark_all_read",
		"notifications_mark_read",
		"project_tasks_list",
		"project_teams_create",
		"project_teams_delete",
		"project_teams_list",
		"project_teams_update",
		"project_time_entries_list",
		"project_users_create",
		"project_users_delete",
		"project_users_list",
		"project_users_update",
		"project_view_buckets_tasks_list",
		"project_view_tasks_list",
		"project_views_create",
		"project_views_delete",
		"project_views_list",
		"project_views_read",
		"project_views_update",
		"projects_duplicate",
		"projects_users_search",
		"reactions_create",
		"reactions_delete",
		"reactions_list",
		"task_assignees_bulk",
		"task_attachments_delete",
		"task_attachments_list",
		"task_bucket_update",
		"task_labels_bulk_replace",
		"task_labels_create",
		"task_labels_delete",
		"task_labels_list",
		"task_time_entries_list",
		"tasks_bulk_create",
		"tasks_bulk_update",
		"tasks_duplicate",
		"tasks_mark_read",
		"tasks_position_update",
		"tasks_read_by_index",
		"tasks_relations_create",
		"tasks_relations_delete",
		"teams_create",
		"teams_delete",
		"teams_list",
		"teams_members_add",
		"teams_members_remove",
		"teams_members_toggle_admin",
		"teams_read",
		"teams_update",
		"time_entries_create",
		"time_entries_delete",
		"time_entries_list",
		"time_entries_read",
		"time_entries_timer_stop",
		"time_entries_update",
	}, catalog)
	var typed []string
	for name := range c.toolNames() {
		if name == "find_action" || name == "do_action" {
			continue
		}
		typed = append(typed, name)
	}
	slices.Sort(typed)
	assert.Equal(t, []string{
		"labels_create",
		"labels_delete",
		"labels_list",
		"labels_read",
		"labels_update",
		"projects_create",
		"projects_delete",
		"projects_list",
		"projects_read",
		"projects_update",
		"task_assignees_create",
		"task_assignees_delete",
		"task_assignees_list",
		"task_comments_create",
		"task_comments_delete",
		"task_comments_list",
		"task_comments_read",
		"task_comments_update",
		"tasks_create",
		"tasks_delete",
		"tasks_list",
		"tasks_read",
		"tasks_update",
		"users_search",
	}, typed)
}
