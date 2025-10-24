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
	"math"
	"net/http"
	"strconv"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"
	apiv1 "code.vikunja.io/api/pkg/routes/api/v1"
	"code.vikunja.io/api/pkg/services"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/labstack/echo/v4"
)

// ProjectRoutes defines all v2 project API routes with explicit permission scopes
var ProjectRoutes = []apiv1.APIRoute{
	{
		Method:          "GET",
		Path:            "/projects",
		Handler:         GetAllProjects,
		PermissionScope: "read_all",
	},
	{
		Method:          "POST",
		Path:            "/projects",
		Handler:         CreateProject,
		PermissionScope: "create",
	},
	{
		Method:          "GET",
		Path:            "/projects/:id",
		Handler:         GetProject,
		PermissionScope: "read_one",
	},
	{
		Method:          "PUT",
		Path:            "/projects/:id",
		Handler:         UpdateProject,
		PermissionScope: "update",
	},
	{
		Method:          "DELETE",
		Path:            "/projects/:id",
		Handler:         DeleteProject,
		PermissionScope: "delete",
	},
	{
		Method:          "POST",
		Path:            "/projects/:id/duplicate",
		Handler:         DuplicateProject,
		PermissionScope: "create", // Duplicating creates a new project
	},
}

// RegisterProjects registers all project routes
func RegisterProjects(a *echo.Group) {
	// Register main project routes with explicit permissions
	apiv1.RegisterRoutes(a, ProjectRoutes, "v2")

	// Register sub-resource routes (users, teams, tasks) separately
	// These need special handling as they're nested under /projects/:id
	projects := a.Group("/projects")

	// Project Users - these will use the legacy registration for now
	projectUsersHandler := &ProjectUsers{}
	projectUsersGroup := projects.Group("/:id/users")
	projectUsersGroup.GET("", projectUsersHandler.Get)
	projectUsersGroup.POST("", projectUsersHandler.Post)
	projectUsersGroup.PUT("/:userid", projectUsersHandler.Put)
	projectUsersGroup.DELETE("/:userid", projectUsersHandler.Delete)

	// Project Teams - these will use the legacy registration for now
	projectTeamsHandler := &ProjectTeams{}
	projectTeams := projects.Group("/:id/teams")
	projectTeams.GET("", projectTeamsHandler.Get)
	projectTeams.POST("", projectTeamsHandler.Post)
	projectTeams.PUT("/:teamid", projectTeamsHandler.Put)
	projectTeams.DELETE("/:teamid", projectTeamsHandler.Delete)

	// Project Tasks - these will use the legacy registration for now
	projects.GET("/:id/tasks", GetProjectTasks)
	projects.POST("/:id/tasks", CreateProjectTask)
}

type ProjectLinks struct {
	Self  string `json:"self"`
	Tasks string `json:"tasks"`
	Users string `json:"users"`
	Teams string `json:"teams"`
}

type ProjectResponse struct {
	*models.Project
	Links *ProjectLinks `json:"_links"`
}

// GetAllProjects handles retrieving all projects for a user
func GetAllProjects(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	auth, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return handler.HandleHTTPError(err)
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

	search := c.QueryParam("s")
	isArchived, _ := strconv.ParseBool(c.QueryParam("is_archived"))

	u, err := user.GetFromAuth(auth)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	opts := &models.ProjectOptions{
		User:        u,
		Search:      search,
		Page:        page,
		PerPage:     perPage,
		GetArchived: isArchived,
	}

	projects, total, err := models.GetAllProjectsForUser(s, auth.GetID(), opts)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	err = models.AddProjectDetails(s, projects, auth)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	resultCount := len(projects)

	var numberOfPages = math.Ceil(float64(total) / float64(perPage))
	if page < 0 {
		numberOfPages = 1
	}
	if resultCount == 0 {
		numberOfPages = 0
	}

	c.Response().Header().Set("x-pagination-total-pages", strconv.FormatFloat(numberOfPages, 'f', 0, 64))
	c.Response().Header().Set("x-pagination-result-count", strconv.Itoa(resultCount))
	c.Response().Header().Set("Access-Control-Expose-Headers", "x-pagination-total-pages, x-pagination-result-count")

	projectsResponse := make([]*ProjectResponse, len(projects))
	for i, p := range projects {
		projectsResponse[i] = &ProjectResponse{
			Project: p,
			Links: &ProjectLinks{
				Self:  fmt.Sprintf("/api/v2/projects/%d", p.ID),
				Tasks: fmt.Sprintf("/api/v2/projects/%d/tasks", p.ID),
				Users: fmt.Sprintf("/api/v2/projects/%d/users", p.ID),
				Teams: fmt.Sprintf("/api/v2/projects/%d/teams", p.ID),
			},
		}
	}

	return c.JSON(http.StatusOK, projectsResponse)
}

// CreateProject creates a new project
func CreateProject(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	auth, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	p := new(models.Project)
	if err := c.Bind(p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project object provided.").SetInternal(err)
	}

	if err := c.Validate(p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error()).SetInternal(err)
	}

	if err := p.Create(s, auth); err != nil {
		return handler.HandleHTTPError(err)
	}

	if err := s.Commit(); err != nil {
		return handler.HandleHTTPError(err)
	}

	response := &ProjectResponse{
		Project: p,
		Links: &ProjectLinks{
			Self:  fmt.Sprintf("/api/v2/projects/%d", p.ID),
			Tasks: fmt.Sprintf("/api/v2/projects/%d/tasks", p.ID),
			Users: fmt.Sprintf("/api/v2/projects/%d/users", p.ID),
			Teams: fmt.Sprintf("/api/v2/projects/%d/teams", p.ID),
		},
	}

	return c.JSON(http.StatusCreated, response)
}

// GetProject retrieves a single project by its ID
func GetProject(c echo.Context) error {
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

	// The CanRead method is responsible for checking if the user has read permissions
	// and for loading the actual project data into the struct.
	can, _, err := p.CanRead(s, auth)
	if err != nil {
		return handler.HandleHTTPError(err)
	}
	if !can {
		return echo.ErrForbidden
	}

	// Now that the project is loaded and permissions are checked,
	// we can populate the rest of the details.
	if err = p.ReadOne(s, auth); err != nil {
		return handler.HandleHTTPError(err)
	}

	response := &ProjectResponse{
		Project: p,
		Links: &ProjectLinks{
			Self:  fmt.Sprintf("/api/v2/projects/%d", p.ID),
			Tasks: fmt.Sprintf("/api/v2/projects/%d/tasks", p.ID),
			Users: fmt.Sprintf("/api/v2/projects/%d/users", p.ID),
			Teams: fmt.Sprintf("/api/v2/projects/%d/teams", p.ID),
		},
	}

	return c.JSON(http.StatusOK, response)
}

// UpdateProject handles updating a project
func UpdateProject(c echo.Context) error {
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

	updatePayload := new(models.Project)
	if err := c.Bind(updatePayload); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project object provided.").SetInternal(err)
	}

	// Set the ID from the URL param, not the payload
	updatePayload.ID = projectID

	if err := c.Validate(updatePayload); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error()).SetInternal(err)
	}

	// The CanUpdate method checks permissions and loads the project.
	can, err := updatePayload.CanUpdate(s, auth)
	if err != nil {
		return handler.HandleHTTPError(err)
	}
	if !can {
		return echo.ErrForbidden
	}

	if err := models.UpdateProject(s, updatePayload, auth, false); err != nil {
		return handler.HandleHTTPError(err)
	}

	if err := s.Commit(); err != nil {
		return handler.HandleHTTPError(err)
	}

	response := &ProjectResponse{
		Project: updatePayload,
		Links: &ProjectLinks{
			Self:  fmt.Sprintf("/api/v2/projects/%d", updatePayload.ID),
			Tasks: fmt.Sprintf("/api/v2/projects/%d/tasks", updatePayload.ID),
			Users: fmt.Sprintf("/api/v2/projects/%d/users", updatePayload.ID),
			Teams: fmt.Sprintf("/api/v2/projects/%d/teams", updatePayload.ID),
		},
	}

	return c.JSON(http.StatusOK, response)
}

// DeleteProject handles deleting a project
func DeleteProject(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	auth, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	// Get the user from auth
	u, err := user.GetFromAuth(auth)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID").SetInternal(err)
	}

	// Use the new Project service Delete method
	projectService := services.NewProjectService(s.Engine())
	if err := projectService.Delete(s, projectID, u); err != nil {
		return handler.HandleHTTPError(err)
	}

	if err := s.Commit(); err != nil {
		return handler.HandleHTTPError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

// DuplicateProject duplicates a project and all its related data
// @Summary Duplicate an existing project
// @Description Copies the project, tasks, files, kanban data, assignees, comments, attachments, labels, relations, backgrounds, user/team permissions and link shares from one project to a new one. The user needs read access in the project and write access in the parent of the new project.
// @Tags project
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "The project ID to duplicate"
// @Param project body models.ProjectDuplicateRequest true "The target parent project which should hold the copied project."
// @Success 201 {object} models.Project "The created project."
// @Failure 400 {object} web.HTTPError "Invalid project duplicate request provided."
// @Failure 403 {object} web.HTTPError "The user does not have access to the project or its parent."
// @Failure 500 {object} web.HTTPError "Internal error"
// @Router /api/v2/projects/{id}/duplicate [post]
func DuplicateProject(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	// Get project ID from path
	projectIDStr := c.Param("id")
	projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID").SetInternal(err)
	}

	// Parse request body for parent project ID
	var req models.ProjectDuplicate
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body").SetInternal(err)
	}

	// Get user from context
	u, err := user.GetCurrentUser(c)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	// Use the ProjectDuplicateService
	service := services.NewProjectDuplicateService(db.GetEngine())
	duplicatedProject, err := service.Duplicate(s, projectID, req.ParentProjectID, u)
	if err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err)
	}

	if err := s.Commit(); err != nil {
		return handler.HandleHTTPError(err)
	}

	return c.JSON(http.StatusCreated, duplicatedProject)
}
