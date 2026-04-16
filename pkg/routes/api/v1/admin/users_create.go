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

package admin

import (
	"net/http"
	"net/mail"
	"strings"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/utils"
	"github.com/labstack/echo/v5"
)

// CreateRequest is the body for POST /admin/users.
type CreateRequest struct {
	Username         string `json:"username"`
	Email            string `json:"email"`
	Name             string `json:"name"`
	Password         string `json:"password"`
	Language         string `json:"language"`
	IsAdmin          bool   `json:"is_admin"`
	SkipEmailConfirm bool   `json:"skip_email_confirm"`
}

// CreateUser creates a new local user from the admin panel. The admin panel bypasses
// ServiceEnableRegistration — site admins should always be able to add accounts.
//
// Behavior matrix:
//   - password + mailer on: the built-in email-confirm flow runs (status stays EmailConfirmationRequired),
//     unless skip_email_confirm is true in which case status is forced to Active.
//   - password + mailer off: status is forced to Active (no confirm mail to send).
//   - no password + mailer on: a random placeholder password is generated, status is forced to Active,
//     and a password-reset email is dispatched so the user sets their own password.
//   - no password + mailer off: request is rejected — there is no way for the user to log in.
//
// @Summary Create a user (admin)
// @Description Create a new local user. Bypasses the public registration toggle. Password is optional when the mailer is configured (the user receives a set-password email).
// @tags admin
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param body body admin.CreateRequest true "New user"
// @Success 200 {object} admin.User
// @Failure 400 {object} web.HTTPError
// @Failure 404 {object} web.HTTPError
// @Router /admin/users [post]
func CreateUser(c *echo.Context) error {
	body := &CreateRequest{}
	if err := c.Bind(body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}

	username := strings.TrimSpace(body.Username)
	email := strings.TrimSpace(body.Email)
	if username == "" || strings.Contains(username, " ") {
		return echo.NewHTTPError(http.StatusBadRequest, "username is required and must not contain spaces")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "a valid email is required")
	}

	mailerOn := config.MailerEnabled.GetBool()
	passwordProvided := body.Password != ""
	if !passwordProvided && !mailerOn {
		return echo.NewHTTPError(http.StatusBadRequest, "password is required when mailer is disabled")
	}

	password := body.Password
	if !passwordProvided {
		// Placeholder secret. The user sets their real password via the reset-token email below.
		random, err := utils.CryptoRandomString(32)
		if err != nil {
			return err
		}
		password = random
	}

	s := db.NewSession()
	defer s.Close()

	newUser, err := user.CreateUser(s, &user.User{
		Username: username,
		Email:    email,
		Name:     body.Name,
		Password: password,
		Language: body.Language,
		Issuer:   user.IssuerLocal,
	})
	if err != nil {
		_ = s.Rollback()
		return err
	}

	if err := models.CreateNewProjectForUser(s, newUser); err != nil {
		_ = s.Rollback()
		return err
	}

	if body.IsAdmin {
		if _, err := s.ID(newUser.ID).Cols("is_admin").Update(&user.User{IsAdmin: true}); err != nil {
			_ = s.Rollback()
			return err
		}
		newUser.IsAdmin = true
	}

	// Force status to Active whenever we don't want the user to go through the email-confirmation
	// flow: mailer is off (no mail can be sent), the admin asked to skip it, or we generated the
	// password and will send a reset email instead.
	if body.SkipEmailConfirm || !mailerOn || !passwordProvided {
		if err := user.SetUserStatus(s, newUser, user.StatusActive); err != nil {
			_ = s.Rollback()
			return err
		}
		newUser.Status = user.StatusActive
	} else {
		// CreateUser flipped the status in a separate update — reflect it on our local copy
		// for the response.
		newUser.Status = user.StatusEmailConfirmationRequired
	}

	if !passwordProvided && mailerOn {
		if err := user.RequestUserPasswordResetToken(s, newUser); err != nil {
			_ = s.Rollback()
			return err
		}
	}

	if err := s.Commit(); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, &User{User: newUser, IsAdmin: newUser.IsAdmin, Status: newUser.Status})
}
