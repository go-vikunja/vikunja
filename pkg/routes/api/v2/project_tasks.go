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
	"math"
	"net/http"
	"strconv"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/web/handler"
	"github.com/labstack/echo/v4"
)

// GetProjectTasks serves a list of tasks in a project.
// @Summary Get all tasks for a project
// @Description Returns all tasks for a project.
// @tags project
// @Accept  json
// @Produce  json
// @Param id path int64 true "The project id"
// @Param page query int false "The page number"
// @Param per_page query int false "The number of items per page"
// @Param s query string false "The filter string"
// @Success 200 {array} models.Task
// @Failure 400 {object} web.HTTPError
// @Failure 401 {object} web.HTTPError
// @Failure 403 {object} web.HTTPError
// @Failure 404 {object} web.HTTPError
// @Router /projects/{id}/tasks [get]
func GetProjectTasks(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	auth, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID").SetInternal(err)
	}

	p := &models.Project{ID: projectID}
	can, _, err := p.CanRead(s, auth)
	if err != nil {
		return handler.HandleHTTPError(err)
	}
	if !can {
		return echo.ErrForbidden
	}

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
		perPageStr = "20"
	}
	perPage, err := strconv.Atoi(perPageStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid per_page number").SetInternal(err)
	}

	tc := &models.TaskCollection{
		ProjectID: projectID,
	}

	tasks, resultCount, totalItems, err := tc.ReadAll(
		s,
		auth,
		c.QueryParam("s"),
		page,
		perPage,
	)
	if err != nil {
		return handler.HandleHTTPError(err)
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

	for _, t := range tasks.([]*models.Task) {
		t.AddLinks(auth)
	}

	return c.JSON(http.StatusOK, tasks)
}

// CreateProjectTask creates a new task in a project
// @Summary Create a new task in a project
// @Description Creates a new task in a project.
// @tags project
// @Accept  json
// @Produce  json
// @Param id path int64 true "The project id"
// @Param task body models.Task true "The task to create"
// @Success 201 {object} models.Task
// @Failure 400 {object} web.HTTPError
// @Failure 401 {object} web.HTTPError
// @Failure 403 {object} web.HTTPError
// @Failure 404 {object} web.HTTPError
// @Router /projects/{id}/tasks [post]
func CreateProjectTask(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	auth, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID").SetInternal(err)
	}

	p := &models.Project{ID: projectID}
	can, err := p.CanWrite(s, auth)
	if err != nil {
		return handler.HandleHTTPError(err)
	}
	if !can {
		return echo.ErrForbidden
	}

	t := new(models.Task)
	if err := c.Bind(t); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid task object provided.").SetInternal(err)
	}
	t.ProjectID = projectID

	if err := c.Validate(t); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error()).SetInternal(err)
	}

	if err := t.Create(s, auth); err != nil {
		return handler.HandleHTTPError(err)
	}

	if err := s.Commit(); err != nil {
		return handler.HandleHTTPError(err)
	}

	t.AddLinks(auth)

	return c.JSON(http.StatusCreated, t)
}
