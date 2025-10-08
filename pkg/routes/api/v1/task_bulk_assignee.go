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
	"strconv"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web/handler"
	"github.com/labstack/echo/v4"
	"xorm.io/xorm"
)

// RegisterBulkAssignees registers the bulk assignee route
func RegisterBulkAssignees(a *echo.Group) {
	a.POST("/tasks/:projecttask/assignees/bulk", handler.WithDBAndUser(bulkAssigneeLogic, true))
}

// bulkAssigneeLogic assigns or removes multiple users from a task at once.
//
// @Summary Bulk assign users to a task
// @Description Assigns or removes multiple users from a task. The user needs to have write access to the task.
// @tags assignees
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param projecttask path int true "Task ID"
// @Param assignees body models.BulkAssignees true "The assignees to update. Contains arrays of users to assign or remove."
// @Success 201 {object} models.BulkAssignees "The updated bulk assignees object."
// @Failure 400 {object} web.HTTPError "Invalid task ID or bulk assignee object"
// @Failure 403 {object} web.HTTPError "The user does not have access to the task"
// @Failure 404 {object} web.HTTPError "The task does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{projecttask}/assignees/bulk [post]
func bulkAssigneeLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	// Parse task ID
	taskID, err := strconv.ParseInt(c.Param("projecttask"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid task ID")
	}

	// Parse bulk assignees from request body
	var bulkAssignees models.BulkAssignees
	if err := c.Bind(&bulkAssignees); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid bulk assignees object")
	}

	bulkAssignees.TaskID = taskID

	// Use model's Create method (which delegates to service)
	err = bulkAssignees.Create(s, u)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, bulkAssignees)
}
