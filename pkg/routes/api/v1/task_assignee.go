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

// RegisterTaskAssignees registers all task assignee routes
func RegisterTaskAssignees(a *echo.Group) {
	a.PUT("/tasks/:projecttask/assignees", handler.WithDBAndUser(addAssigneeLogic, true))
	a.DELETE("/tasks/:projecttask/assignees/:user", handler.WithDBAndUser(removeAssigneeLogic, true))
	a.GET("/tasks/:projecttask/assignees", handler.WithDBAndUser(getTaskAssigneesLogic, false))
}

// addAssigneeLogic adds a user as an assignee to a task.
//
// @Summary Add a user to a task
// @Description Adds a user as an assignee to a task. The user needs to have write access to the task and the user being assigned needs to have at least read access.
// @tags assignees
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param projecttask path int true "Task ID"
// @Param user body user.User true "The user to add as an assignee. Must contain at least a user_id field."
// @Success 201 {object} models.TaskAssginee "The created assignee object."
// @Failure 400 {object} web.HTTPError "Invalid task ID or user object"
// @Failure 403 {object} web.HTTPError "The user does not have access to the task"
// @Failure 404 {object} web.HTTPError "The task does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{projecttask}/assignees [put]
func addAssigneeLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	// Parse task ID
	taskID, err := strconv.ParseInt(c.Param("projecttask"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid task ID")
	}

	// Parse assignee from request body
	var assignee models.TaskAssginee
	if err := c.Bind(&assignee); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid assignee object")
	}

	assignee.TaskID = taskID

	// Use model's Create method (which delegates to service)
	err = assignee.Create(s, u)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, assignee)
}

// removeAssigneeLogic removes a user from a task's assignees.
//
// @Summary Remove a user from a task
// @Description Removes a user from a task's assignees. The user needs to have write access to the task.
// @tags assignees
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param projecttask path int true "Task ID"
// @Param user path int true "User ID"
// @Success 200 {object} models.Message "The user was successfully removed from the task."
// @Failure 400 {object} web.HTTPError "Invalid task ID or user ID"
// @Failure 403 {object} web.HTTPError "The user does not have access to the task"
// @Failure 404 {object} web.HTTPError "The task or assignee does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{projecttask}/assignees/{user} [delete]
func removeAssigneeLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	// Parse task ID
	taskID, err := strconv.ParseInt(c.Param("projecttask"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid task ID")
	}

	// Parse user ID
	userID, err := strconv.ParseInt(c.Param("user"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID")
	}

	// Create assignee object for deletion
	assignee := &models.TaskAssginee{
		TaskID: taskID,
		UserID: userID,
	}

	// Use model's Delete method (which delegates to service)
	err = assignee.Delete(s, u)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, models.Message{Message: "The user was successfully removed from the task."})
}

// getTaskAssigneesLogic retrieves all assignees for a task.
//
// @Summary Get all assignees for a task
// @Description Returns all users assigned to a task.
// @tags assignees
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param projecttask path int true "Task ID"
// @Param page query int false "The page number. Used for pagination. If not provided, the first page of results is returned."
// @Param per_page query int false "The maximum number of items per page. Note this parameter is limited by the configured maximum of items per page."
// @Success 200 {array} models.TaskAssginee "The assignees for the task."
// @Failure 400 {object} web.HTTPError "Invalid task ID"
// @Failure 403 {object} web.HTTPError "The user does not have access to the task"
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{projecttask}/assignees [get]
func getTaskAssigneesLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	// Parse task ID
	taskID, err := strconv.ParseInt(c.Param("projecttask"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid task ID")
	}

	// Create assignee object for ReadAll
	assignee := &models.TaskAssginee{
		TaskID: taskID,
	}

	// Use model's ReadAll method (which delegates to service)
	result, _, _, err := assignee.ReadAll(s, u, "", 1, 50)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, result)
}
