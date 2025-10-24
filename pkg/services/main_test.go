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

package services

import (
	"os"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/i18n"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"xorm.io/xorm"
)

var testEngine *xorm.Engine

func TestMain(m *testing.M) {
	// Initialize logger for tests
	log.InitLogger()

	// Set default config
	config.InitDefaultConfig()
	// We need to set the root path even if we're not using the config, otherwise fixtures are not loaded correctly
	config.ServiceRootpath.Set(os.Getenv("VIKUNJA_SERVICE_ROOTPATH"))

	i18n.Init()

	// Some tests use the file engine, so we'll need to initialize that
	files.InitTests()

	user.InitTests()

	models.SetupTests()
	events.Fake()

	// Initialize service dependency injection in the correct order
	// First, wire up model/service dependencies
	InitializeDependencies()
	// Then initialize service instances
	InitUserService()
	InitSavedFilterService()
	InitTaskService()
	InitProjectService()
	InitKanbanService()
	InitProjectDuplicateService()
	InitAttachmentService()
	InitCommentService() // T-PERM-011: Initialize comment service for tests

	// Initialize testEngine for service tests
	testEngine = db.GetEngine()

	// Register API routes for permission validation in tests
	// This populates the apiTokenRoutes map needed for PermissionsAreValid()
	registerTestAPIRoutes()

	os.Exit(m.Run())
}

// registerTestAPIRoutes manually registers API routes for testing
// This is needed because tests don't start the web server, but API token
// permission validation requires the apiTokenRoutes map to be populated
func registerTestAPIRoutes() {
	// Register v1 task routes
	models.CollectRoute("PUT", "/api/v1/projects/:project/tasks", "create")
	models.CollectRoute("GET", "/api/v1/tasks/:taskid", "read_one")
	models.CollectRoute("POST", "/api/v1/tasks/:taskid", "update")
	models.CollectRoute("DELETE", "/api/v1/tasks/:taskid", "delete")

	// Register other commonly used v1 routes for API token tests
	models.CollectRoute("GET", "/api/v1/projects", "read_all")
	models.CollectRoute("PUT", "/api/v1/projects", "create")
	models.CollectRoute("GET", "/api/v1/projects/:project", "read_one")
	models.CollectRoute("POST", "/api/v1/projects/:project", "update")
	models.CollectRoute("DELETE", "/api/v1/projects/:project", "delete")

	// Register v1 label-task routes (T066)
	models.CollectRoute("GET", "/api/v1/tasks/:projecttask/labels", "read_all")
	models.CollectRoute("PUT", "/api/v1/tasks/:projecttask/labels", "create")
	models.CollectRoute("DELETE", "/api/v1/tasks/:projecttask/labels/:label", "delete")
	models.CollectRoute("POST", "/api/v1/tasks/:projecttask/labels/bulk", "update")
	// Bulk routes with /bulk suffix create separate route groups, but CanDoAPIRoute() strips
	// _bulk when checking permissions. Copy the bulk route to the main group for path matching.
	routes := models.GetAPITokenRoutes()
	if bulkRoute, ok := routes["v1"]["tasks_labels_bulk"]["update"]; ok {
		routes["v1"]["tasks_labels"]["update"] = bulkRoute
	}

	// Register v1 task assignee routes (T067)
	models.CollectRoute("GET", "/api/v1/tasks/:projecttask/assignees", "read_all")
	models.CollectRoute("PUT", "/api/v1/tasks/:projecttask/assignees", "create")
	models.CollectRoute("DELETE", "/api/v1/tasks/:projecttask/assignees/:user", "delete")
	models.CollectRoute("POST", "/api/v1/tasks/:projecttask/assignees/bulk", "update")
	// Copy bulk route to main group for permission checking
	if bulkRoute, ok := routes["v1"]["tasks_assignees_bulk"]["update"]; ok {
		routes["v1"]["tasks_assignees"]["update"] = bulkRoute
	}

	// Register v1 task relation routes (T068)
	models.CollectRoute("PUT", "/api/v1/tasks/:task/relations", "create")
	models.CollectRoute("DELETE", "/api/v1/tasks/:task/relations/:relationKind/:otherTask", "delete")

	// Register v1 task position routes (T069)
	models.CollectRoute("POST", "/api/v1/tasks/:task/position", "update")

	// Register v1 bulk task routes (T070)
	// Note: Uses 'bulk_update' scope to avoid conflicting with single-task update route
	models.CollectRoute("POST", "/api/v1/tasks/bulk", "bulk_update")
	// Copy bulk route to tasks group for permission checking
	if bulkRoute, ok := routes["v1"]["tasks_bulk"]["bulk_update"]; ok {
		routes["v1"]["tasks"]["bulk_update"] = bulkRoute
	}

	// Register v1 kanban/bucket routes (T071)
	models.CollectRoute("GET", "/api/v1/projects/:project/views/:view/buckets", "read_all")
	models.CollectRoute("PUT", "/api/v1/projects/:project/views/:view/buckets", "create")
	models.CollectRoute("POST", "/api/v1/projects/:project/views/:view/buckets/:bucket", "update")
	models.CollectRoute("DELETE", "/api/v1/projects/:project/views/:view/buckets/:bucket", "delete")
	models.CollectRoute("POST", "/api/v1/projects/:project/views/:view/buckets/:bucket/tasks", "move_task")
	// Move task route creates projects_views_buckets_tasks group, copy to projects_views_buckets for logical grouping
	if tasksGroup, ok := routes["v1"]["projects_views_buckets_tasks"]; ok {
		if moveRoute, ok := tasksGroup["move_task"]; ok {
			routes["v1"]["projects_views_buckets"]["move_task"] = moveRoute
		}
	}

	// Register v1 project view routes (T072)
	models.CollectRoute("GET", "/api/v1/projects/:project/views", "read_all")
	models.CollectRoute("GET", "/api/v1/projects/:project/views/:view", "read_one")
	models.CollectRoute("PUT", "/api/v1/projects/:project/views", "create")
	models.CollectRoute("POST", "/api/v1/projects/:project/views/:view", "update")
	models.CollectRoute("DELETE", "/api/v1/projects/:project/views/:view", "delete")

	// Register v1 saved filter routes (T073)
	models.CollectRoute("GET", "/api/v1/filters", "read_all")
	models.CollectRoute("GET", "/api/v1/filters/:filter", "read_one")
	models.CollectRoute("PUT", "/api/v1/filters", "create")
	models.CollectRoute("POST", "/api/v1/filters/:filter", "update")
	models.CollectRoute("DELETE", "/api/v1/filters/:filter", "delete")

	// Register v1 webhook routes (T074) - Admin-only routes
	models.CollectRoute("GET", "/api/v1/projects/:project/webhooks", "read_all", true)
	models.CollectRoute("PUT", "/api/v1/projects/:project/webhooks", "create", true)
	models.CollectRoute("POST", "/api/v1/projects/:project/webhooks/:webhook", "update", true)
	models.CollectRoute("DELETE", "/api/v1/projects/:project/webhooks/:webhook", "delete", true)

	// Register v1 team routes (T075) - Admin-only routes
	models.CollectRoute("GET", "/api/v1/teams", "read_all", true)
	models.CollectRoute("GET", "/api/v1/teams/:team", "read_one", true)
	models.CollectRoute("PUT", "/api/v1/teams", "create", true)
	models.CollectRoute("POST", "/api/v1/teams/:team", "update", true)
	models.CollectRoute("DELETE", "/api/v1/teams/:team", "delete", true)
	models.CollectRoute("PUT", "/api/v1/teams/:team/members", "add_member", true)
	models.CollectRoute("DELETE", "/api/v1/teams/:team/members/:user", "remove_member", true)
	models.CollectRoute("POST", "/api/v1/teams/:team/members/:user/admin", "update_member", true)

	// Team member routes create teams_members group, but permission
	// checking expects them in teams group. Copy routes to teams group.
	{
		routes := models.GetAPITokenRoutes()
		// Ensure teams group exists
		if routes["v1"]["teams"] == nil {
			routes["v1"]["teams"] = make(models.APITokenRoute)
		}
		// Copy from teams_members
		if memberRoutes, ok := routes["v1"]["teams_members"]; ok {
			for scope, route := range memberRoutes {
				routes["v1"]["teams"][scope] = route
			}
		}
		// Copy from teams_members_admin (for update_member route)
		if memberAdminRoutes, ok := routes["v1"]["teams_members_admin"]; ok {
			for scope, route := range memberAdminRoutes {
				routes["v1"]["teams"][scope] = route
			}
		}
	}

	// Register v2 routes for API token tests
	// v2 tasks
	models.CollectRoute("GET", "/api/v2/tasks", "read_all")

	// v2 projects
	models.CollectRoute("GET", "/api/v2/projects", "read_all")
	models.CollectRoute("POST", "/api/v2/projects", "create")
	models.CollectRoute("GET", "/api/v2/projects/:id", "read_one")
	models.CollectRoute("PUT", "/api/v2/projects/:id", "update")
	models.CollectRoute("DELETE", "/api/v2/projects/:id", "delete")
	models.CollectRoute("POST", "/api/v2/projects/:id/duplicate", "create")

	// v2 labels
	models.CollectRoute("GET", "/api/v2/labels", "read_all")
	models.CollectRoute("POST", "/api/v2/labels", "create")
	models.CollectRoute("GET", "/api/v2/labels/:id", "read_one")
	models.CollectRoute("PUT", "/api/v2/labels/:id", "update")
	models.CollectRoute("DELETE", "/api/v2/labels/:id", "delete")
}
