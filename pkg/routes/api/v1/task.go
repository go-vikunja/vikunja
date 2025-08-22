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

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/services"
	"code.vikunja.io/api/pkg/web/handler"
	"github.com/labstack/echo/v4"
)

// RegisterTasks registers all task routes
func RegisterTasks(a *echo.Group) {
	a.GET("/projects/:project/tasks", GetAllTasksByProject)
	a.PUT("/projects/:project/tasks", CreateTask)
	a.GET("/tasks/:task", GetTask)
	a.POST("/tasks/:task", UpdateTask)
	a.DELETE("/tasks/:task", DeleteTask)
	a.GET("/tasks", GetAllTasks)

	a.GET("/projects/:project/views/:view/tasks", GetTasksByProjectView)
	a.GET("/tasks/all", GetAllTasks) // This already exists, but the old route was `/tasks/all`. The new one is `/tasks`. I'll add the old one back for compatibility.
	a.POST("/tasks/:task/position", UpdateTaskPosition)
	a.POST("/tasks/bulk", BulkUpdateTasks)
	a.PUT("/tasks/:task/assignees", AddAssignee)
	a.DELETE("/tasks/:task/assignees/:user", RemoveAssignee)
	a.GET("/tasks/:task/assignees", GetAssignees)
	a.POST("/tasks/:task/assignees/bulk", BulkAddAssignees)
	a.PUT("/tasks/:task/labels", AddLabelToTask)
	a.DELETE("/tasks/:task/labels/:label", RemoveLabelFromTask)
	a.GET("/tasks/:task/labels", GetTaskLabels)
	a.POST("/tasks/:task/labels/bulk", BulkAddLabelsToTask)
	a.PUT("/tasks/:task/relations", AddTaskRelation)
	a.DELETE("/tasks/:task/relations/:relationKind/:otherTask", DeleteTaskRelation)
	a.GET("/tasks/:task/attachments", GetTaskAttachments)
	a.DELETE("/tasks/:task/attachments/:attachment", DeleteTaskAttachment)
	a.GET("/tasks/:task/comments", GetTaskComments)
	a.PUT("/tasks/:task/comments", CreateTaskComment)
	a.DELETE("/tasks/:task/comments/:commentid", DeleteTaskComment)
	a.POST("/tasks/:task/comments/:commentid", UpdateTaskComment)
	a.GET("/tasks/:task/comments/:commentid", GetTaskComment)
}

func GetTask(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	auth, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	taskID, err := strconv.ParseInt(c.Param("task"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid task ID").SetInternal(err)
	}

	ts := services.NewTaskService()
	t, err := ts.Get(s, taskID, auth)
	if err != nil {
		return handler.HandleHTTPError(err)
	}
	return c.JSON(http.StatusOK, t)
}

func GetAllTasksByProject(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	auth, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	projectID, err := strconv.ParseInt(c.Param("project"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID").SetInternal(err)
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
	if perPage < 1 {
		perPage = 20
	}
	search := c.QueryParam("s")

	ts := services.NewTaskService()
	tasks, _, _, err := ts.GetByProject(s, projectID, auth, search, page, perPage, services.TaskOptions{})
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	return c.JSON(http.StatusOK, tasks)
}

func GetAllTasks(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	auth, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
	if perPage < 1 {
		perPage = 20
	}
	search := c.QueryParam("s")

	ts := services.NewTaskService()
	tasks, _, _, err := ts.GetAll(s, auth, search, page, perPage, services.TaskOptions{})
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	return c.JSON(http.StatusOK, tasks)
}

func CreateTask(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	auth, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

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

	ts := services.NewTaskService()
	t, err = ts.Create(s, t, auth)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	if err := s.Commit(); err != nil {
		return handler.HandleHTTPError(err)
	}

	return c.JSON(http.StatusCreated, t)
}

func UpdateTask(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	auth, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	taskID, err := strconv.ParseInt(c.Param("task"), 10, 64)
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

	ts := services.NewTaskService()
	t, err := ts.Update(s, updatePayload, auth)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	if err := s.Commit(); err != nil {
		return handler.HandleHTTPError(err)
	}

	return c.JSON(http.StatusOK, t)
}

func DeleteTask(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	auth, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	taskID, err := strconv.ParseInt(c.Param("task"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid task ID").SetInternal(err)
	}

	ts := services.NewTaskService()
	if err := ts.Delete(s, taskID, auth); err != nil {
		return handler.HandleHTTPError(err)
	}

	if err := s.Commit(); err != nil {
		return handler.HandleHTTPError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

func GetTasksByProjectView(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, "Not implemented")
}

func UpdateTaskPosition(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, "Not implemented")
}

func BulkUpdateTasks(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, "Not implemented")
}

func AddAssignee(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, "Not implemented")
}

func RemoveAssignee(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, "Not implemented")
}

func GetAssignees(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, "Not implemented")
}

func BulkAddAssignees(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, "Not implemented")
}

func AddLabelToTask(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, "Not implemented")
}

func RemoveLabelFromTask(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, "Not implemented")
}

func GetTaskLabels(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, "Not implemented")
}

func BulkAddLabelsToTask(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, "Not implemented")
}

func AddTaskRelation(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, "Not implemented")
}

func DeleteTaskRelation(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, "Not implemented")
}

func GetTaskAttachments(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, "Not implemented")
}

func DeleteTaskAttachment(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, "Not implemented")
}

func GetTaskComments(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, "Not implemented")
}

func CreateTaskComment(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, "Not implemented")
}

func DeleteTaskComment(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, "Not implemented")
}

func UpdateTaskComment(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, "Not implemented")
}

func GetTaskComment(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, "Not implemented")
}
