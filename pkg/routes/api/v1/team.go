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

// RegisterTeams registers all team management routes
func RegisterTeams(a *echo.Group) {
	a.GET("/teams", handler.WithDBAndUser(getAllTeamsLogic, false))
	a.GET("/teams/:team", handler.WithDBAndUser(getTeamLogic, false))
	a.PUT("/teams", handler.WithDBAndUser(createTeamLogic, true))
	a.POST("/teams/:team", handler.WithDBAndUser(updateTeamLogic, true))
	a.DELETE("/teams/:team", handler.WithDBAndUser(deleteTeamLogic, true))
	a.PUT("/teams/:team/members", handler.WithDBAndUser(addTeamMemberLogic, true))
	a.DELETE("/teams/:team/members/:user", handler.WithDBAndUser(removeTeamMemberLogic, true))
	a.POST("/teams/:team/members/:user/admin", handler.WithDBAndUser(updateTeamMemberLogic, true))
}

// getAllTeamsLogic retrieves all teams the user has access to.
//
// @Summary Get all teams
// @Description Returns all teams the user has access to.
// @tags teams
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param page query int false "The page number for pagination. Used for pagination. If not provided, the first page of results is returned."
// @Param per_page query int false "The maximum number of items per page. Note this parameter is limited by the configured maximum of items per page."
// @Param s query string false "Search teams by name."
// @Success 200 {array} models.Team "All teams."
// @Failure 500 {object} models.Message "Internal error"
// @Router /teams [get]
func getAllTeamsLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	// Parse pagination and search parameters
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}

	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
	if perPage < 1 {
		perPage = 50
	}

	search := c.QueryParam("s")

	// Create team object for ReadAll
	team := &models.Team{}

	// Use model's ReadAll method (which delegates to service)
	result, resultCount, totalItems, err := team.ReadAll(s, u, search, page, perPage)
	if err != nil {
		return err
	}

	// Set pagination headers
	totalPages := totalItems / int64(perPage)
	if totalItems%int64(perPage) > 0 {
		totalPages++
	}
	c.Response().Header().Set("x-pagination-total-pages", strconv.FormatInt(totalPages, 10))
	c.Response().Header().Set("x-pagination-result-count", strconv.Itoa(resultCount))

	return c.JSON(http.StatusOK, result)
}

// getTeamLogic retrieves a single team by ID.
//
// @Summary Get a team
// @Description Returns a team by its ID.
// @tags teams
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param team path int true "Team ID"
// @Success 200 {object} models.Team "The team"
// @Failure 400 {object} web.HTTPError "Invalid team ID"
// @Failure 403 {object} web.HTTPError "The user does not have access to the team"
// @Failure 404 {object} web.HTTPError "The team does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /teams/{team} [get]
func getTeamLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	// Parse team ID
	teamID, err := strconv.ParseInt(c.Param("team"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid team ID")
	}

	// Create team object
	team := &models.Team{ID: teamID}

	// Use model's ReadOne method (which delegates to service)
	err = team.ReadOne(s, u)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, team)
}

// createTeamLogic creates a new team.
//
// @Summary Create a team
// @Description Creates a new team in a given namespace. The user needs write access to the namespace.
// @tags teams
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param team body models.Team true "The team object"
// @Success 201 {object} models.Team "The created team."
// @Failure 400 {object} web.HTTPError "Invalid team object"
// @Failure 500 {object} models.Message "Internal error"
// @Router /teams [put]
func createTeamLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	// Parse team from request body
	var team models.Team
	if err := c.Bind(&team); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid team object")
	}

	// Use model's Create method (which delegates to service)
	err := team.Create(s, u)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, team)
}

// updateTeamLogic updates a team.
//
// @Summary Update a team
// @Description Updates a team. The user needs write access to the team.
// @tags teams
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param team path int true "Team ID"
// @Param teamData body models.Team true "The team with updated values"
// @Success 200 {object} models.Team "The updated team."
// @Failure 400 {object} web.HTTPError "Invalid team object"
// @Failure 403 {object} web.HTTPError "The user does not have access to the team"
// @Failure 404 {object} web.HTTPError "The team does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /teams/{team} [post]
func updateTeamLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	// Parse team ID
	teamID, err := strconv.ParseInt(c.Param("team"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid team ID")
	}

	// Parse team from request body
	var team models.Team
	if err := c.Bind(&team); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid team object")
	}

	team.ID = teamID

	// Use model's Update method (which delegates to service)
	err = team.Update(s, u)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, team)
}

// deleteTeamLogic deletes a team.
//
// @Summary Delete a team
// @Description Deletes a team. The user needs write access to the team.
// @tags teams
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param team path int true "Team ID"
// @Success 200 {object} models.Message "The team was successfully deleted."
// @Failure 400 {object} web.HTTPError "Invalid team ID"
// @Failure 403 {object} web.HTTPError "The user does not have access to the team"
// @Failure 404 {object} web.HTTPError "The team does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /teams/{team} [delete]
func deleteTeamLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	// Parse team ID
	teamID, err := strconv.ParseInt(c.Param("team"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid team ID")
	}

	// Create team object for deletion
	team := &models.Team{ID: teamID}

	// Use model's Delete method (which delegates to service)
	err = team.Delete(s, u)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, models.Message{Message: "The team was successfully deleted."})
}

// addTeamMemberLogic adds a user to a team.
//
// @Summary Add a user to a team
// @Description Add a user to a team. The user needs write access to the team.
// @tags teams
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param team path int true "Team ID"
// @Param member body models.TeamMember true "The user to add to the team"
// @Success 201 {object} models.TeamMember "The created team member object."
// @Failure 400 {object} web.HTTPError "Invalid team ID or member object"
// @Failure 403 {object} web.HTTPError "The user does not have access to the team"
// @Failure 404 {object} web.HTTPError "The team does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /teams/{team}/members [put]
func addTeamMemberLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	// Parse team ID
	teamID, err := strconv.ParseInt(c.Param("team"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid team ID")
	}

	// Parse member from request body
	var member models.TeamMember
	if err := c.Bind(&member); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid member object")
	}

	member.TeamID = teamID

	// Use model's Create method (which delegates to service)
	err = member.Create(s, u)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, member)
}

// removeTeamMemberLogic removes a user from a team.
//
// @Summary Remove a user from a team
// @Description Removes a user from a team. The user needs write access to the team.
// @tags teams
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param team path int true "Team ID"
// @Param user path int true "User ID"
// @Success 200 {object} models.Message "The user was successfully removed from the team."
// @Failure 400 {object} web.HTTPError "Invalid team ID or user ID"
// @Failure 403 {object} web.HTTPError "The user does not have access to the team"
// @Failure 404 {object} web.HTTPError "The team member does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /teams/{team}/members/{user} [delete]
func removeTeamMemberLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	// Parse team ID
	teamID, err := strconv.ParseInt(c.Param("team"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid team ID")
	}

	// Parse user ID
	userID, err := strconv.ParseInt(c.Param("user"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID")
	}

	// Create member object for deletion
	member := &models.TeamMember{
		TeamID:   teamID,
		Username: strconv.FormatInt(userID, 10),
	}

	// Use model's Delete method (which delegates to service)
	err = member.Delete(s, u)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, models.Message{Message: "The user was successfully removed from the team."})
}

// updateTeamMemberLogic updates a team member's admin status.
//
// @Summary Update a team member
// @Description Update a team member. Primarily used to update the admin flag.
// @tags teams
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param team path int true "Team ID"
// @Param user path int true "User ID"
// @Param member body models.TeamMember true "The team member with updated values"
// @Success 200 {object} models.TeamMember "The updated team member."
// @Failure 400 {object} web.HTTPError "Invalid team ID or user ID"
// @Failure 403 {object} web.HTTPError "The user does not have access to the team"
// @Failure 404 {object} web.HTTPError "The team member does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /teams/{team}/members/{user}/admin [post]
func updateTeamMemberLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	// Parse team ID
	teamID, err := strconv.ParseInt(c.Param("team"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid team ID")
	}

	// Parse user ID
	userID, err := strconv.ParseInt(c.Param("user"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID")
	}

	// Parse member from request body
	var member models.TeamMember
	if err := c.Bind(&member); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid member object")
	}

	member.TeamID = teamID
	member.Username = strconv.FormatInt(userID, 10)

	// Use model's Update method (which delegates to service)
	err = member.Update(s, u)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, member)
}
