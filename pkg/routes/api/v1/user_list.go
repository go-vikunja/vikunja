//  Vikunja is a todo-list application to facilitate your life.
//  Copyright 2018 Vikunja and contributors. All rights reserved.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU General Public License as published by
//  the Free Software Foundation, either version 3 of the License, or
//  (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU General Public License for more details.
//
//  You should have received a copy of the GNU General Public License
//  along with this program.  If not, see <https://www.gnu.org/licenses/>.

package v1

import (
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/web/handler"
	"github.com/labstack/echo/v4"
	"net/http"
)

// UserList gets all information about a user
// @Summary Get users
// @Description Lists all users (without emailadresses). Also possible to search for a specific user.
// @tags user
// @Accept json
// @Produce json
// @Param s query string false "Search for a user by its name."
// @Security JWTKeyAuth
// @Success 200 {array} models.User "All (found) users."
// @Failure 400 {object} code.vikunja.io/web.HTTPError "Something's invalid."
// @Failure 500 {object} models.Message "Internal server error."
// @Router /users [get]
func UserList(c echo.Context) error {
	s := c.QueryParam("s")
	users, err := models.ListUsers(s)
	if err != nil {
		return handler.HandleHTTPError(err, c)
	}

	// Obfuscate the mailadresses
	for in := range users {
		users[in].Email = ""
	}

	return c.JSON(http.StatusOK, users)
}
