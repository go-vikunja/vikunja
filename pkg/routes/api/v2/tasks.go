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

package v2

import (
	"fmt"
	"net/http"
	"strconv"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/models/v2"
	"code.vikunja.io/api/pkg/modules/auth"
	"github.com/labstack/echo/v4"
)

// DeleteTask handles deleting a task.
func DeleteTask(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	u, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return err
	}

	taskID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid task ID.")
	}

	t := &models.Task{ID: taskID}
	if err := t.Delete(s, u); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

// UpdateTask handles updating a task.
func UpdateTask(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	u, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return err
	}

	taskID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid task ID.")
	}

	var t models.Task
	if err := c.Bind(&t); err != nil {
		return err
	}
	t.ID = taskID

	if err := t.Update(s, u); err != nil {
		return err
	}

	v2Task := &v2.Task{
		Task: t,
		Links: &v2.TaskLinks{
			Self: &v2.Link{
				Href: fmt.Sprintf("/api/v2/tasks/%d", t.ID),
			},
			Project: &v2.Link{
				Href: fmt.Sprintf("/api/v2/projects/%d", t.ProjectID),
			},
		},
	}

	return c.JSON(http.StatusOK, v2Task)
}

// GetTask handles getting a task by its ID.
func GetTask(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	u, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return err
	}

	taskID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid task ID.")
	}

	t := &models.Task{ID: taskID}
	if err := t.ReadOne(s, u); err != nil {
		return err
	}

	v2Task := &v2.Task{
		Task: *t,
		Links: &v2.TaskLinks{
			Self: &v2.Link{
				Href: fmt.Sprintf("/api/v2/tasks/%d", t.ID),
			},
			Project: &v2.Link{
				Href: fmt.Sprintf("/api/v2/projects/%d", t.ProjectID),
			},
		},
	}

	return c.JSON(http.StatusOK, v2Task)
}

// CreateTask handles creating a new task.
func CreateTask(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	u, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return err
	}

	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID.")
	}

	var t models.Task
	if err := c.Bind(&t); err != nil {
		return err
	}
	t.ProjectID = projectID

	if err := t.Create(s, u); err != nil {
		return err
	}

	v2Task := &v2.Task{
		Task: t,
		Links: &v2.TaskLinks{
			Self: &v2.Link{
				Href: fmt.Sprintf("/api/v2/tasks/%d", t.ID),
			},
			Project: &v2.Link{
				Href: fmt.Sprintf("/api/v2/projects/%d", t.ProjectID),
			},
		},
	}

	return c.JSON(http.StatusCreated, v2Task)
}

// GetTasks handles getting all tasks for a project.
func GetTasks(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	aut, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return err
	}

	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID")
	}

	p, err := models.GetProjectSimpleByID(s, projectID)
	if err != nil {
		return err
	}

	page, perPage := v2.GetPageAndPerPage(c)
	search := c.QueryParam("s")

	tasks, _, _, err := models.GetTasksForProjects(s, []*models.Project{p}, aut, &models.TaskSearchOptions{
		Search:  search,
		Page:    page,
		PerPage: perPage,
	}, nil)
	if err != nil {
		return err
	}

	v2Tasks := make([]*v2.Task, len(tasks))
	for i, t := range tasks {
		v2Tasks[i] = &v2.Task{
			Task: *t,
			Links: &v2.TaskLinks{
				Self: &v2.Link{
					Href: fmt.Sprintf("/api/v2/tasks/%d", t.ID),
				},
			},
		}
	}

	return c.JSON(http.StatusOK, v2Tasks)
}
