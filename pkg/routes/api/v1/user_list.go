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
	auth2 "code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web/handler"
	"github.com/labstack/echo/v4"
)

// UserList gets all information about a list of users
// @Summary Get users
// @Description Search for a user by its username, name or full email. Name (not username) or email require that the user has enabled this in their settings.
// @tags user
// @Accept json
// @Produce json
// @Param s query string false "The search criteria."
// @Security JWTKeyAuth
// @Success 200 {array} user.User "All (found) users."
// @Failure 400 {object} web.HTTPError "Something's invalid."
// @Failure 500 {object} models.Message "Internal server error."
// @Router /users [get]
func UserList(c echo.Context) error {
	search := c.QueryParam("s")

	s := db.NewSession()
	defer s.Close()

	currentUser, err := user.GetCurrentUser(c)
	if err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err)
	}

	users, err := user.ListUsers(s, search, currentUser, nil)
	if err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err)
	}

	// Obfuscate the mailadresses
	for in := range users {
		users[in].Email = ""
	}

	return c.JSON(http.StatusOK, users)
}

// ListUsersForProject returns a list with all users who have access to a project, regardless of the method the project was shared with them.
// @Summary Get users
// @Description Lists all users (without emailadresses). Also possible to search for a specific user.
// @tags project
// @Accept json
// @Produce json
// @Param s query string false "Search for a user by its name."
// @Security JWTKeyAuth
// @Param id path int true "Project ID"
// @Success 200 {array} user.User "All (found) users."
// @Failure 400 {object} web.HTTPError "Something's invalid."
// @Failure 401 {object} web.HTTPError "The user does not have the permission to see the project."
// @Failure 500 {object} models.Message "Internal server error."
// @Router /projects/{id}/projectusers [get]
func ListUsersForProject(c echo.Context) error {
	projectID, err := strconv.ParseInt(c.Param("project"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": "Invalid project ID",
			"details": err.Error(),
		})
	}

	project := models.Project{ID: projectID}
	auth, err := auth2.GetAuthFromClaims(c)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	s := db.NewSession()
	defer s.Close()

	canRead, _, err := project.CanRead(s, auth)
	if err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err)
	}
	if !canRead {
		return echo.ErrForbidden
	}

	currentUser, err := user.GetCurrentUser(c)
	if err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err)
	}

	search := c.QueryParam("s")
	users, err := models.ListUsersFromProject(s, &project, currentUser, search)
	if err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err)
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err)
	}

	return c.JSON(http.StatusOK, users)
}
