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

// UserResetPassword is the handler to change a users password
// @Summary Resets a password
// @Description Resets a user email with a previously reset token.
// @tags user
// @Accept json
// @Produce json
// @Param credentials body models.PasswordReset true "The token with the new password."
// @Success 200 {object} models.Message
// @Failure 400 {object} code.vikunja.io/web.HTTPError "Bad token provided."
// @Failure 500 {object} models.Message "Internal error"
// @Router /user/password/reset [post]
func UserResetPassword(c echo.Context) error {
	// Check for Request Content
	var pwReset user.PasswordReset
	if err := c.Bind(&pwReset); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "No password provided.")
	}

	err := user.ResetPassword(&pwReset)
	if err != nil {
		return handler.HandleHTTPError(err, c)
	}

	return c.JSON(http.StatusOK, models.Message{"The password was updated successfully."})
}

// UserRequestResetPasswordToken is the handler to change a users password
// @Summary Request password reset token
// @Description Requests a token to reset a users password. The token is sent via email.
// @tags user
// @Accept json
// @Produce json
// @Param credentials body models.PasswordTokenRequest true "The username of the user to request a token for."
// @Success 200 {object} models.Message
// @Failure 404 {object} code.vikunja.io/web.HTTPError "The user does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /user/password/token [post]
func UserRequestResetPasswordToken(c echo.Context) error {
	// Check for Request Content
	var pwTokenReset user.PasswordTokenRequest
	if err := c.Bind(&pwTokenReset); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "No username provided.")
	}

	if err := c.Validate(pwTokenReset); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	err := user.RequestUserPasswordResetToken(&pwTokenReset)
	if err != nil {
		return handler.HandleHTTPError(err, c)
	}

	return c.JSON(http.StatusOK, models.Message{"Token was sent."})
}
