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
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/web/handler"
	"github.com/labstack/echo/v4"
	"net/http"
)

// UserConfirmEmail is the handler to confirm a user email
// @Summary Confirm the email of a new user
// @Description Confirms the email of a newly registered user.
// @tags user
// @Accept json
// @Produce json
// @Param credentials body models.EmailConfirm true "The token."
// @Success 200 {object} models.Message
// @Failure 412 {object} code.vikunja.io/web.HTTPError "Bad token provided."
// @Failure 500 {object} models.Message "Internal error"
// @Router /user/confirm [post]
func UserConfirmEmail(c echo.Context) error {
	// Check for Request Content
	var emailConfirm user.EmailConfirm
	if err := c.Bind(&emailConfirm); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "No token provided.")
	}

	err := user.ConfirmEmail(&emailConfirm)
	if err != nil {
		return handler.HandleHTTPError(err, c)
	}

	return c.JSON(http.StatusOK, models.Message{"The email was confirmed successfully."})
}
