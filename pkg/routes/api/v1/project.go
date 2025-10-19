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

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/services"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/labstack/echo/v4"
	"xorm.io/xorm"
)

// ProjectRoutes defines all project API routes with their explicit permission scopes.
// This replaces the old implicit permission detection with explicit declarations.
var ProjectRoutes = []APIRoute{
	{Method: "GET", Path: "/projects", Handler: handler.WithDBAndUser(getAllProjectsLogic, false), PermissionScope: "read_all"},
	{Method: "GET", Path: "/projects/:project", Handler: handler.WithDBAndUser(getProjectLogic, false), PermissionScope: "read_one"},
	{Method: "PUT", Path: "/projects", Handler: handler.WithDBAndUser(createProjectLogic, true), PermissionScope: "create"},
	{Method: "POST", Path: "/projects/:project", Handler: handler.WithDBAndUser(updateProjectLogic, true), PermissionScope: "update"},
	{Method: "DELETE", Path: "/projects/:project", Handler: handler.WithDBAndUser(deleteProjectLogic, true), PermissionScope: "delete"},
	{Method: "GET", Path: "/projects/:project/tasks", Handler: handler.WithDBAndUser(getProjectTasksLogic, false), PermissionScope: "read_all"},
}

// RegisterProjects registers all project routes
func RegisterProjects(a *echo.Group) {
	registerRoutes(a, ProjectRoutes)
}

// getAllProjectsLogic handles retrieving all projects for a user
func getAllProjectsLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	// Parse pagination parameters
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
		perPageStr = "50"
	}
	perPage, err := strconv.Atoi(perPageStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid per_page number").SetInternal(err)
	}

	// Parse archived parameter
	archivedStr := c.QueryParam("is_archived")
	isArchived := archivedStr == "true"

	projectService := services.NewProjectService(s.Engine())
	projects, resultCount, totalItems, err := projectService.GetAllForUser(s, u, c.QueryParam("s"), page, perPage, isArchived)
	if err != nil {
		return err
	}

	// Set pagination headers
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

	return c.JSON(http.StatusOK, projects)
}

// getProjectLogic retrieves a single project by its ID
func getProjectLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	projectID, err := strconv.ParseInt(c.Param("project"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID").SetInternal(err)
	}

	projectService := services.NewProjectService(s.Engine())
	project, err := projectService.GetByID(s, projectID, u)
	if err != nil {
		return err
	}

	// Set max permission header for frontend compatibility
	c.Response().Header().Set("x-max-permission", strconv.FormatInt(int64(project.MaxPermission), 10))
	c.Response().Header().Set("Access-Control-Expose-Headers", "x-max-permission")

	return c.JSON(http.StatusOK, project)
}

// createProjectLogic creates a new project
func createProjectLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	p := new(models.Project)
	if err := c.Bind(p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project object provided.").SetInternal(err)
	}

	if err := c.Validate(p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error()).SetInternal(err)
	}

	projectService := services.NewProjectService(s.Engine())
	createdProject, err := projectService.Create(s, p, u)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, createdProject)
}

// updateProjectLogic handles updating a project
func updateProjectLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	projectID, err := strconv.ParseInt(c.Param("project"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID").SetInternal(err)
	}

	updatePayload := new(models.Project)
	if err := c.Bind(updatePayload); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project object provided.").SetInternal(err)
	}
	updatePayload.ID = projectID

	if err := c.Validate(updatePayload); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error()).SetInternal(err)
	}

	projectService := services.NewProjectService(s.Engine())
	updatedProject, err := projectService.Update(s, updatePayload, u)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, updatedProject)
}

// deleteProjectLogic handles deleting a project
func deleteProjectLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	projectID, err := strconv.ParseInt(c.Param("project"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID").SetInternal(err)
	}

	projectService := services.NewProjectService(s.Engine())
	if err := projectService.Delete(s, projectID, u); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

// getProjectTasksLogic handles retrieving all tasks in a project
func getProjectTasksLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	projectID, err := strconv.ParseInt(c.Param("project"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID").SetInternal(err)
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

	taskService := services.NewTaskService(s.Engine())
	tasks, resultCount, totalItems, err := taskService.GetAllByProject(s, projectID, u, page, perPage, c.QueryParam("s"))
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
