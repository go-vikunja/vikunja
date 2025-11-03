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
	"errors"
	"fmt"
	"net/http"

	"code.vikunja.io/api/pkg/db"

	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web/handler"
	"github.com/labstack/echo/v4"
)

// UpdateUserEmail is the handler to let a user update their email address.
// @Summary Update email address
// @Description Lets the current user change their email address.
// @tags user
// @Accept json
// @Produce json
// @Param userEmailUpdate body user.EmailUpdate true "The new email address and current password."
// @Security JWTKeyAuth
// @Success 200 {object} models.Message
// @Failure 400 {object} web.HTTPError "Something's invalid."
// @Failure 404 {object} web.HTTPError "User does not exist."
// @Failure 500 {object} models.Message "Internal server error."
// @Router /user/settings/email [post]
func UpdateUserEmail(c echo.Context) (err error) {

	var emailUpdate = &user.EmailUpdate{}
	if err := c.Bind(emailUpdate); err != nil {
		log.Debugf("Invalid model error. Internal error was: %s", err.Error())
		var he *echo.HTTPError
		if errors.As(err, &he) {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid model provided. Error was: %s", he.Message)).SetInternal(err)
		}
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid model provided.").SetInternal(err)
	}

	emailUpdate.User, err = user.GetCurrentUser(c)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	s := db.NewSession()
	defer s.Close()

	emailUpdate.User, err = user.CheckUserCredentials(s, &user.Login{
		Username: emailUpdate.User.Username,
		Password: emailUpdate.Password,
	})
	if err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err)
	}

	err = user.UpdateEmail(s, emailUpdate)
	if err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err)
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err)
	}

	return c.JSON(http.StatusOK, models.Message{Message: "We sent you email with a link to confirm your email address."})
}
