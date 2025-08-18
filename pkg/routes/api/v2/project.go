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
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/labstack/echo/v4"
)

// RegisterProjects registers all project routes
func RegisterProjects(a *echo.Group) {
	projects := a.Group("/projects")
	projects.GET("", GetAllProjects)
	projects.POST("", CreateProject)
	projects.GET("/:id", GetProject)
	projects.PUT("/:id", UpdateProject)
	projects.DELETE("/:id", DeleteProject)
	projects.POST("/:id/duplicate", DuplicateProject)

	// Project Users
	projectUsers := projects.Group("/:id/users")
	projectUsers.GET("", GetProjectUsers)
	projectUsers.POST("", AddProjectUser)
	projectUsers.PUT("/:userid", UpdateProjectUser)
	projectUsers.DELETE("/:userid", RemoveProjectUser)

	// Project Teams
	projectTeams := projects.Group("/:id/teams")
	projectTeams.GET("", GetProjectTeams)
	projectTeams.POST("", AddProjectTeam)
	projectTeams.PUT("/:teamid", UpdateProjectTeam)
	projectTeams.DELETE("/:teamid", RemoveProjectTeam)
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
		return err
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
		return err
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

	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID").SetInternal(err)
	}

	p := &models.Project{ID: projectID}

	// The CanDelete method checks permissions and loads the project.
	can, err := p.CanDelete(s, auth)
	if err != nil {
		return handler.HandleHTTPError(err)
	}
	if !can {
		return echo.ErrForbidden
	}

	if err := p.Delete(s, auth); err != nil {
		return handler.HandleHTTPError(err)
	}

	if err := s.Commit(); err != nil {
		return handler.HandleHTTPError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

// DuplicateProject is a stub handler
func DuplicateProject(c echo.Context) error {
	return c.String(http.StatusNotImplemented, "Not Implemented")
}

// GetProjectUsers is a stub handler
func GetProjectUsers(c echo.Context) error {
	return c.String(http.StatusNotImplemented, "Not Implemented")
}

// AddProjectUser is a stub handler
func AddProjectUser(c echo.Context) error {
	return c.String(http.StatusNotImplemented, "Not Implemented")
}

// UpdateProjectUser is a stub handler
func UpdateProjectUser(c echo.Context) error {
	return c.String(http.StatusNotImplemented, "Not Implemented")
}

// RemoveProjectUser is a stub handler
func RemoveProjectUser(c echo.Context) error {
	return c.String(http.StatusNotImplemented, "Not Implemented")
}

// GetProjectTeams handles getting all teams in a project
func GetProjectTeams(c echo.Context) error {
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

	tp := &models.TeamProject{ProjectID: projectID}
	teams, resultCount, total, err := tp.ReadAll(s, auth, search, page, perPage)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	c.Response().Header().Set("x-pagination-total-pages", strconv.FormatFloat(math.Ceil(float64(total)/float64(perPage)), 'f', 0, 64))
	c.Response().Header().Set("x-pagination-result-count", strconv.Itoa(resultCount))
	c.Response().Header().Set("Access-Control-Expose-Headers", "x-pagination-total-pages, x-pagination-result-count")

	return c.JSON(http.StatusOK, teams)
}

// AddProjectTeam adds a team to a project
func AddProjectTeam(c echo.Context) error {
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

	tp := new(models.TeamProject)
	if err := c.Bind(tp); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid team project object provided.").SetInternal(err)
	}
	tp.ProjectID = projectID

	if err := c.Validate(tp); err != nil {
		return err
	}

	if err := tp.Create(s, auth); err != nil {
		return handler.HandleHTTPError(err)
	}

	if err := s.Commit(); err != nil {
		return handler.HandleHTTPError(err)
	}

	return c.JSON(http.StatusCreated, tp)
}

// UpdateProjectTeam updates a team on a project
func UpdateProjectTeam(c echo.Context) error {
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

	teamID, err := strconv.ParseInt(c.Param("teamid"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid team ID").SetInternal(err)
	}

	tp := new(models.TeamProject)
	if err := c.Bind(tp); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid team project object provided.").SetInternal(err)
	}
	tp.ProjectID = projectID
	tp.TeamID = teamID

	if err := c.Validate(tp); err != nil {
		return err
	}

	if err := tp.Update(s, auth); err != nil {
		return handler.HandleHTTPError(err)
	}

	if err := s.Commit(); err != nil {
		return handler.HandleHTTPError(err)
	}

	return c.JSON(http.StatusOK, tp)
}

// RemoveProjectTeam removes a team from a project
func RemoveProjectTeam(c echo.Context) error {
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

	teamID, err := strconv.ParseInt(c.Param("teamid"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid team ID").SetInternal(err)
	}

	tp := &models.TeamProject{
		ProjectID: projectID,
		TeamID:    teamID,
	}

	if err := tp.Delete(s, auth); err != nil {
		return handler.HandleHTTPError(err)
	}

	if err := s.Commit(); err != nil {
		return handler.HandleHTTPError(err)
	}

	return c.NoContent(http.StatusNoContent)
}
