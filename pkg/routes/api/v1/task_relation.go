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

// RegisterTaskRelations registers all task relation routes
func RegisterTaskRelations(a *echo.Group) {
	a.PUT("/tasks/:task/relations", handler.WithDBAndUser(createTaskRelationLogic, true))
	a.DELETE("/tasks/:task/relations/:relationKind/:otherTask", handler.WithDBAndUser(deleteTaskRelationLogic, true))
}

// createTaskRelationLogic creates a new relation between two tasks.
//
// @Summary Create a task relation
// @Description Creates a new relation between two tasks. The user needs to have write access to both tasks.
// @tags task
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param task path int true "Task ID"
// @Param relation body models.TaskRelation true "The relation object with the kind and the other task ID."
// @Success 201 {object} models.TaskRelation "The created task relation object."
// @Failure 400 {object} web.HTTPError "Invalid task relation object provided."
// @Failure 403 {object} web.HTTPError "The user does not have access to the task"
// @Failure 404 {object} web.HTTPError "The task does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{task}/relations [put]
func createTaskRelationLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	// Parse task ID
	taskID, err := strconv.ParseInt(c.Param("task"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid task ID")
	}

	// Parse relation from request body
	var relation models.TaskRelation
	if err := c.Bind(&relation); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid task relation object")
	}

	relation.TaskID = taskID

	// Use model's Create method (which delegates to service)
	err = relation.Create(s, u)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, relation)
}

// deleteTaskRelationLogic deletes a relation between two tasks.
//
// @Summary Delete a task relation
// @Description Deletes a relation between two tasks. The user needs to have write access to both tasks.
// @tags task
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param task path int true "Task ID"
// @Param relationKind path string true "The kind of the relation to delete."
// @Param otherTask path int true "The other task ID"
// @Success 200 {object} models.Message "The task relation was successfully deleted."
// @Failure 400 {object} web.HTTPError "Invalid task relation object provided."
// @Failure 403 {object} web.HTTPError "The user does not have access to the task"
// @Failure 404 {object} web.HTTPError "The task relation does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{task}/relations/{relationKind}/{otherTask} [delete]
func deleteTaskRelationLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	// Parse task ID
	taskID, err := strconv.ParseInt(c.Param("task"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid task ID")
	}

	// Parse other task ID
	otherTaskID, err := strconv.ParseInt(c.Param("otherTask"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid other task ID")
	}

	// Parse relation kind
	relationKind := models.RelationKind(c.Param("relationKind"))

	// Create relation object for deletion
	relation := &models.TaskRelation{
		TaskID:       taskID,
		OtherTaskID:  otherTaskID,
		RelationKind: relationKind,
	}

	// Use model's Delete method (which delegates to service)
	err = relation.Delete(s, u)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, models.Message{Message: "The task relation was successfully deleted."})
}
