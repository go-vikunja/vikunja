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

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/labstack/echo/v4"
)

type UserPasswordConfirmation struct {
	Password string `json:"password" valid:"required"`
}

type UserDeletionRequestConfirm struct {
	Token string `json:"token" valid:"required"`
}

// UserRequestDeletion is the handler to request a user deletion process (sends a mail)
// @Summary Request the deletion of the user
// @Description Requests the deletion of the current user. It will trigger an email which has to be confirmed to start the deletion.
// @tags user
// @Accept json
// @Produce json
// @Param credentials body v1.UserPasswordConfirmation true "The user password."
// @Success 200 {object} models.Message
// @Failure 412 {object} web.HTTPError "Bad password provided."
// @Failure 500 {object} models.Message "Internal error"
// @Router /user/deletion/request [post]
func UserRequestDeletion(c echo.Context) error {

	s := db.NewSession()
	defer s.Close()

	err := s.Begin()
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	u, err := user.GetCurrentUserFromDB(s, c)
	if err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err)
	}

	if u.IsLocalUser() {
		var deletionRequest UserPasswordConfirmation
		if err := c.Bind(&deletionRequest); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "No password provided.").SetInternal(err)
		}

		err = c.Validate(deletionRequest)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err).SetInternal(err)
		}

		err = user.CheckUserPassword(u, deletionRequest.Password)
		if err != nil {
			_ = s.Rollback()
			return handler.HandleHTTPError(err).SetInternal(err)
		}
	}

	err = user.RequestDeletion(s, u)
	if err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err)
	}

	err = s.Commit()
	if err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err)
	}

	return c.JSON(http.StatusOK, models.Message{Message: "Successfully requested deletion."})
}

// UserConfirmDeletion is the handler to confirm a user deletion process and start it
// @Summary Confirm a user deletion request
// @Description Confirms the deletion request of a user sent via email.
// @tags user
// @Accept json
// @Produce json
// @Param credentials body v1.UserDeletionRequestConfirm true "The token."
// @Success 200 {object} models.Message
// @Failure 412 {object} web.HTTPError "Bad token provided."
// @Failure 500 {object} models.Message "Internal error"
// @Router /user/deletion/confirm [post]
func UserConfirmDeletion(c echo.Context) error {
	var deleteConfirmation UserDeletionRequestConfirm
	if err := c.Bind(&deleteConfirmation); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "No token provided.").SetInternal(err)
	}

	err := c.Validate(deleteConfirmation)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err).SetInternal(err)
	}

	s := db.NewSession()
	defer s.Close()

	err = s.Begin()
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	u, err := user.GetCurrentUserFromDB(s, c)
	if err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err)
	}

	err = user.ConfirmDeletion(s, u, deleteConfirmation.Token)
	if err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err)
	}

	err = s.Commit()
	if err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err)
	}

	return c.JSON(http.StatusNoContent, models.Message{Message: "Successfully confirmed the deletion request."})
}

// UserCancelDeletion is the handler to abort a user deletion process
// @Summary Abort a user deletion request
// @Description Aborts an in-progress user deletion.
// @tags user
// @Accept json
// @Produce json
// @Param credentials body v1.UserPasswordConfirmation true "The user password to confirm."
// @Success 200 {object} models.Message
// @Failure 412 {object} web.HTTPError "Bad password provided."
// @Failure 500 {object} models.Message "Internal error"
// @Router /user/deletion/cancel [post]
func UserCancelDeletion(c echo.Context) error {

	s := db.NewSession()
	defer s.Close()

	err := s.Begin()
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	u, err := user.GetCurrentUserFromDB(s, c)
	if err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err)
	}

	if u.IsLocalUser() {
		var deletionRequest UserPasswordConfirmation
		if err := c.Bind(&deletionRequest); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "No password provided.").SetInternal(err)
		}

		err = c.Validate(deletionRequest)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err).SetInternal(err)
		}

		err = user.CheckUserPassword(u, deletionRequest.Password)
		if err != nil {
			_ = s.Rollback()
			return handler.HandleHTTPError(err)
		}
	}

	err = user.CancelDeletion(s, u)
	if err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err)
	}

	err = s.Commit()
	if err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err)
	}

	return c.JSON(http.StatusNoContent, models.Message{Message: "Successfully confirmed the deletion request."})
}
