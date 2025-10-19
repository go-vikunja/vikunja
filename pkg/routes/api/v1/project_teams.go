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
	"code.vikunja.io/api/pkg/services"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web/handler"
	"github.com/labstack/echo/v4"
	"xorm.io/xorm"
)

// RegisterProjectTeams registers all project-team routes
func RegisterProjectTeams(a *echo.Group) {
	a.GET("/projects/:project/teams", handler.WithDBAndUser(getAllProjectTeamsLogic, false))
	a.PUT("/projects/:project/teams", handler.WithDBAndUser(createProjectTeamLogic, true))
	a.DELETE("/projects/:project/teams/:team", handler.WithDBAndUser(deleteProjectTeamLogic, true))
	a.POST("/projects/:project/teams/:team", handler.WithDBAndUser(updateProjectTeamLogic, true))
}

// getAllProjectTeamsLogic handles retrieving all teams with access to a project.
//
// @Summary Get all teams with access to a project
// @Description Returns all teams that have access to a project with their permission levels
// @tags sharing
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param project path int true "Project ID"
// @Param page query int false "The page number for pagination. Used for pagination. If not provided, the first page of results is returned."
// @Param per_page query int false "The maximum number of items per page. Note this parameter is limited by the configured maximum of items per page."
// @Param s query string false "Search teams by name."
// @Success 200 {array} models.TeamWithPermission "All teams with their permissions."
// @Failure 403 {object} web.HTTPError "The user does not have access to the project"
// @Failure 404 {object} web.HTTPError "The project does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /projects/{project}/teams [get]
func getAllProjectTeamsLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	// Parse project ID
	projectID, err := strconv.ParseInt(c.Param("project"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID")
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}

	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
	if perPage < 1 {
		perPage = 50
	}

	search := c.QueryParam("s")

	// Get teams from service
	service := services.NewProjectTeamService(s.Engine())
	teams, resultCount, totalItems, err := service.GetAll(s, projectID, u, search, page, perPage)
	if err != nil {
		return err
	}

	// Set pagination headers
	c.Response().Header().Set("x-pagination-total-pages", strconv.FormatInt((totalItems+int64(perPage)-1)/int64(perPage), 10))
	c.Response().Header().Set("x-pagination-result-count", strconv.Itoa(resultCount))

	return c.JSON(http.StatusOK, teams)
}

// createProjectTeamLogic handles adding a team to a project.
//
// @Summary Add a team to a project
// @Description Gives a team access to a project with specified permissions.
// @tags sharing
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param project path int true "Project ID"
// @Param projectTeam body models.TeamProject true "The team you want to add to the project."
// @Success 200 {object} models.TeamProject "The created team-project relation."
// @Failure 400 {object} web.HTTPError "Invalid team project object provided."
// @Failure 403 {object} web.HTTPError "The user does not have access to the project"
// @Failure 404 {object} web.HTTPError "The project does not exist."
// @Failure 409 {object} web.HTTPError "The team already has access to the project."
// @Failure 500 {object} models.Message "Internal error"
// @Router /projects/{project}/teams [put]
func createProjectTeamLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	// Parse project ID
	projectID, err := strconv.ParseInt(c.Param("project"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID")
	}

	// Parse request body
	var teamProject models.TeamProject
	if err := c.Bind(&teamProject); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid team project object")
	}

	teamProject.ProjectID = projectID

	// Create team-project relation via service
	service := services.NewProjectTeamService(s.Engine())
	if err := service.Create(s, &teamProject, u); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, teamProject)
}

// deleteProjectTeamLogic handles removing a team from a project.
//
// @Summary Delete a team from a project
// @Description Removes a team's access to a project.
// @tags sharing
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param project path int true "Project ID"
// @Param team path int true "Team ID"
// @Success 200 {object} models.Message "The team was successfully removed from the project."
// @Failure 403 {object} web.HTTPError "The user does not have access to the project"
// @Failure 404 {object} web.HTTPError "The project or team does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /projects/{project}/teams/{team} [delete]
func deleteProjectTeamLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	// Parse project ID
	projectID, err := strconv.ParseInt(c.Param("project"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID")
	}

	// Parse team ID
	teamID, err := strconv.ParseInt(c.Param("team"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid team ID")
	}

	teamProject := &models.TeamProject{
		ProjectID: projectID,
		TeamID:    teamID,
	}

	// Delete team-project relation via service
	service := services.NewProjectTeamService(s.Engine())
	if err := service.Delete(s, teamProject); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, models.Message{Message: "The team was successfully removed from the project."})
}

// updateProjectTeamLogic handles updating a team's permission level on a project.
//
// @Summary Update a team's permission level
// @Description Update a team's permission level on a project. The user needs to have admin access to the project.
// @tags sharing
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param project path int true "Project ID"
// @Param team path int true "Team ID"
// @Param projectTeam body models.TeamProject true "The team you want to update."
// @Success 200 {object} models.TeamProject "The updated team-project relation."
// @Failure 400 {object} web.HTTPError "Invalid team project object provided."
// @Failure 403 {object} web.HTTPError "The user does not have admin access to the project"
// @Failure 404 {object} web.HTTPError "The project or team does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /projects/{project}/teams/{team} [post]
func updateProjectTeamLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	// Parse project ID
	projectID, err := strconv.ParseInt(c.Param("project"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID")
	}

	// Parse team ID
	teamID, err := strconv.ParseInt(c.Param("team"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid team ID")
	}

	// Parse request body
	var teamProject models.TeamProject
	if err := c.Bind(&teamProject); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid team project object")
	}

	teamProject.ProjectID = projectID
	teamProject.TeamID = teamID

	// Update team-project relation via service
	service := services.NewProjectTeamService(s.Engine())
	if err := service.Update(s, &teamProject); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, teamProject)
}
