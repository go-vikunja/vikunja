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
	"github.com/labstack/echo"
	"net/http"
)

// UserPassword holds a user password. Used to update it.
type UserPassword struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// UserChangePassword is the handler to change a users password
// @Summary Change password
// @Description Lets the current user change its password.
// @tags user
// @Accept json
// @Produce json
// @Param userPassword body v1.UserPassword true "The current and new password."
// @Security ApiKeyAuth
// @Success 200 {object} models.Message
// @Failure 400 {object} models.HTTPError "Something's invalid."
// @Failure 404 {object} models.HTTPError "User does not exist."
// @Failure 500 {object} models.Message "Internal server error."
// @Router /user/password [post]
func UserChangePassword(c echo.Context) error {
	// Check if the user is itself
	doer, err := models.GetCurrentUser(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Error getting current user.")
	}

	// Check for Request Content
	var newPW UserPassword
	if err := c.Bind(&newPW); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "No password provided.")
	}

	// Check the current password
	if _, err = models.CheckUserCredentials(&models.UserLogin{Username: doer.Username, Password: newPW.OldPassword}); err != nil {
		return handler.HandleHTTPError(err, c)
	}

	// Update the password
	if err = models.UpdateUserPassword(doer, newPW.NewPassword); err != nil {
		return handler.HandleHTTPError(err, c)
	}

	return c.JSON(http.StatusOK, models.Message{"The password was updated successfully."})
}
