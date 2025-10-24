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

package v1

import (
	"net/http"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web/handler"
	"github.com/labstack/echo/v4"
	"xorm.io/xorm"
)

// BulkTaskRoutes defines all bulk task API routes with their explicit permission scopes.
// This enables API tokens to be scoped for bulk task operations.
// Note: Uses 'bulk_update' scope instead of 'update' to avoid conflicting with single-task update route.
var BulkTaskRoutes = []APIRoute{
	{Method: "POST", Path: "/tasks/bulk", Handler: handler.WithDBAndUser(bulkUpdateTasksLogic, true), PermissionScope: "bulk_update"},
}

// RegisterBulkTasks registers the bulk task update route
func RegisterBulkTasks(a *echo.Group) {
	registerRoutes(a, BulkTaskRoutes)

	// Bulk routes are registered in a separate group (tasks_bulk),
	// but permission checking strips _bulk suffix. Copy bulk route registration
	// to base tasks group so CanDoAPIRoute() can find it.
	routes := models.GetAPITokenRoutes()
	if routes["v1"]["tasks"] == nil {
		routes["v1"]["tasks"] = make(models.APITokenRoute)
	}
	if bulkRoute, ok := routes["v1"]["tasks_bulk"]["bulk_update"]; ok {
		routes["v1"]["tasks"]["bulk_update"] = bulkRoute
	}
}

// bulkUpdateTasksLogic updates multiple tasks at once.
//
// @Summary Bulk update tasks
// @Description Updates multiple tasks at once. All tasks must be in the same project. The user needs to have write access to the project.
// @tags task
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param tasks body models.BulkTask true "The tasks to update with the same properties."
// @Success 200 {object} models.BulkTask "The updated tasks object."
// @Failure 400 {object} web.HTTPError "Invalid bulk task object provided."
// @Failure 403 {object} web.HTTPError "The user does not have access to the tasks"
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/bulk [post]
func bulkUpdateTasksLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	// Parse bulk task from request body
	var bulkTask models.BulkTask
	if err := c.Bind(&bulkTask); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid bulk task object")
	}

	// Use model's Update method (which delegates to service)
	err := bulkTask.Update(s, u)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, bulkTask)
}
