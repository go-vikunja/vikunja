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
	"math"
	"net/http"
	"strconv"
	"strings"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/services"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web/handler"
	"github.com/labstack/echo/v4"
	"xorm.io/xorm"
)

// TaskRoutes defines all task API routes with their explicit permission scopes.
var TaskRoutes = []APIRoute{
	// {Method: "GET", Path: "/tasks/all", Handler: handler.WithDBAndUser(getAllTasksLogic, false), PermissionScope: "read_all"},
	// read_all is handled by TaskCollection ?
	{Method: "PUT", Path: "/projects/:project/tasks", Handler: handler.WithDBAndUser(createTaskLogic, true), PermissionScope: "create"},
	{Method: "GET", Path: "/tasks/:taskid", Handler: handler.WithDBAndUser(getTaskLogic, false), PermissionScope: "read_one"},
	{Method: "POST", Path: "/tasks/:taskid", Handler: handler.WithDBAndUser(updateTaskLogic, true), PermissionScope: "update"},
	{Method: "DELETE", Path: "/tasks/:taskid", Handler: handler.WithDBAndUser(deleteTaskLogic, true), PermissionScope: "delete"},
}

// RegisterTasks registers all task routes
func RegisterTasks(a *echo.Group) {
	registerRoutes(a, TaskRoutes)
}

// createTaskLogic is the handler to create a task.
func createTaskLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	projectID, err := strconv.ParseInt(c.Param("project"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID").SetInternal(err)
	}

	t := new(models.Task)
	if err := c.Bind(t); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid task object provided.").SetInternal(err)
	}
	t.ProjectID = projectID

	if err := c.Validate(t); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error()).SetInternal(err)
	}

	taskService := services.NewTaskService(s.Engine())
	createdTask, err := taskService.Create(s, t, u)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, createdTask)
}

// getTaskLogic is the handler to get a single task.
func getTaskLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	taskID, err := strconv.ParseInt(c.Param("taskid"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid task ID").SetInternal(err)
	}

	// Parse expand parameters from query string
	expand := []models.TaskCollectionExpandable{}
	expandParam := c.QueryParam("expand")
	if expandParam != "" {
		expandValues := strings.Split(expandParam, ",")
		for _, expandValue := range expandValues {
			expandValue = strings.TrimSpace(expandValue)
			expandable := models.TaskCollectionExpandable(expandValue)
			if err := expandable.Validate(); err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, "Invalid expand parameter: "+expandValue).SetInternal(err)
			}
			expand = append(expand, expandable)
		}
	}

	taskService := services.NewTaskService(s.Engine())
	task, err := taskService.GetByIDWithExpansion(s, taskID, u, expand)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, task)
}

// getAllTasksLogic is the handler to get all tasks.
func getAllTasksLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	pageStr := c.QueryParam("page")
	if pageStr == "" {
		pageStr = "1"
	}
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid page number").SetInternal(err)
	}

	perPageStr := c.QueryParam("per_page")
	if perPageStr == "" {
		perPageStr = "50" // A reasonable default
	}
	perPage, err := strconv.Atoi(perPageStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid per_page number").SetInternal(err)
	}

	collection := new(models.TaskCollection)
	if err := c.Bind(collection); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid filter object provided.").SetInternal(err)
	}

	taskService := services.NewTaskService(s.Engine())
	search := c.QueryParam("s")

	tasks, resultCount, totalItems, err := taskService.GetAllWithFilters(s, collection, u, search, page, perPage)
	if err != nil {
		return err
	}

	var numberOfPages = math.Ceil(float64(totalItems) / float64(perPage))
	if page < 0 {
		numberOfPages = 1
	}
	if resultCount == 0 {
		numberOfPages = 0
	}

	c.Response().Header().Set("x-pagination-total-pages", strconv.FormatFloat(numberOfPages, 'f', 0, 64))
	c.Response().Header().Set("x-pagination-result-count", strconv.Itoa(resultCount))
	c.Response().Header().Set("Access-Control-Expose-Headers", "x-pagination-total-pages, x-pagination-result-count")

	return c.JSON(http.StatusOK, tasks)
}

// updateTaskLogic handles updating a task
func updateTaskLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	taskID, err := strconv.ParseInt(c.Param("taskid"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid task ID").SetInternal(err)
	}

	updatePayload := new(models.Task)
	if err := c.Bind(updatePayload); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid task object provided.").SetInternal(err)
	}
	updatePayload.ID = taskID

	if err := c.Validate(updatePayload); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error()).SetInternal(err)
	}

	taskService := services.NewTaskService(s.Engine())
	updatedTask, err := taskService.Update(s, updatePayload, u)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, updatedTask)
}

// deleteTaskLogic is the handler to delete a task.
func deleteTaskLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	taskID, err := strconv.ParseInt(c.Param("taskid"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid task ID").SetInternal(err)
	}

	task := &models.Task{ID: taskID}

	taskService := services.NewTaskService(s.Engine())
	if err := taskService.Delete(s, task, u); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}
