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

// RegisterTaskPositions registers the task position route
func RegisterTaskPositions(a *echo.Group) {
	a.POST("/tasks/:task/position", handler.WithDBAndUser(updateTaskPositionLogic, true))
}

// updateTaskPositionLogic updates the position of a task within a project view.
//
// @Summary Updates a task position
// @Description Updates a task position within a project view. This is used for kanban boards and other visual task organization.
// @tags task
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param task path int true "Task ID"
// @Param position body models.TaskPosition true "The task position with updated values you want to change."
// @Success 200 {object} models.TaskPosition "The updated task position."
// @Failure 400 {object} web.HTTPError "Invalid task position object provided."
// @Failure 403 {object} web.HTTPError "The user does not have access to the task"
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{task}/position [post]
func updateTaskPositionLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	// Parse task ID
	taskID, err := strconv.ParseInt(c.Param("task"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid task ID")
	}

	// Parse position from request body
	var position models.TaskPosition
	if err := c.Bind(&position); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid task position object")
	}

	position.TaskID = taskID

	// Use model's Update method (which delegates to service)
	err = position.Update(s, u)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, position)
}
