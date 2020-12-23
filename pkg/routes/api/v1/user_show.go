// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2020 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package v1

import (
	"net/http"

	"code.vikunja.io/api/pkg/db"

	user2 "code.vikunja.io/api/pkg/user"
	"code.vikunja.io/web/handler"
	"github.com/labstack/echo/v4"
)

// UserShow gets all informations about the current user
// @Summary Get user information
// @Description Returns the current user object.
// @tags user
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Success 200 {object} user.User
// @Failure 404 {object} web.HTTPError "User does not exist."
// @Failure 500 {object} models.Message "Internal server error."
// @Router /user [get]
func UserShow(c echo.Context) error {
	userInfos, err := user2.GetCurrentUser(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Error getting current user.")
	}

	s := db.NewSession()
	defer s.Close()

	user, err := user2.GetUserByID(s, userInfos.ID)
	if err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err, c)
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err, c)
	}

	return c.JSON(http.StatusOK, user)
}
