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
	"net/http"
	"strings"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/user"

	"github.com/labstack/echo/v5"
)

type UserRegister struct {
	// The language of the new user. Must be a valid IETF BCP 47 language code and exist in Vikunja.
	Language string `json:"language" valid:"language"`
	user.APIUserPassword
	// Mark the new user as a site admin. Ignored unless the caller is an authenticated site admin.
	IsAdmin bool `json:"is_admin"`
	// Activate the new user immediately without email confirmation. Ignored unless the caller is an authenticated site admin.
	SkipEmailConfirm bool `json:"skip_email_confirm"`
}

// callerIsSiteAdmin detects whether the request carries a valid bearer for a
// site-admin user. The /register route is in the public group so we parse the
// header ourselves. Any parse error is treated as "not authenticated" — we
// never surface it, the absence of a bearer is legitimate.
func callerIsSiteAdmin(c *echo.Context) bool {
	header := c.Request().Header.Get(echo.HeaderAuthorization)
	if header == "" {
		return false
	}
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return false
	}
	tokenStr := strings.TrimSpace(header[len(prefix):])
	if tokenStr == "" {
		return false
	}
	u, err := auth.ParseJWTForOptionalAuth(tokenStr)
	if err != nil || u == nil {
		return false
	}
	return u.IsAdmin && u.ID > 0
}

// RegisterUser is the register handler
// @Summary Register
// @Description Creates a new user account. When called by an authenticated site admin, the public registration toggle is bypassed and the admin-only fields `is_admin` and `skip_email_confirm` are honored.
// @tags auth
// @Accept json
// @Produce json
// @Param credentials body v1.UserRegister true "The user with credentials to create"
// @Success 200 {object} user.User
// @Failure 400 {object} web.HTTPError "No or invalid user register object provided / User already exists."
// @Failure 500 {object} models.Message "Internal error"
// @Router /register [post]
func RegisterUser(c *echo.Context) error {
	isAdmin := callerIsSiteAdmin(c)
	if !isAdmin && !config.ServiceEnableRegistration.GetBool() {
		return echo.ErrNotFound
	}
	// Check for Request Content
	var userIn *UserRegister
	if err := c.Bind(&userIn); err != nil {
		return c.JSON(http.StatusBadRequest, models.Message{Message: "No or invalid user model provided."})
	}
	if err := c.Validate(userIn); err != nil {
		e := models.ValidationHTTPError{}
		if is := errors.As(err, &e); is {
			return c.JSON(e.HTTPCode, e)
		}

		return err
	}
	if userIn == nil {
		return c.JSON(http.StatusBadRequest, models.Message{Message: "No or invalid user model provided."})
	}

	s := db.NewSession()
	defer s.Close()

	// Insert the user
	newUser, err := user.CreateUser(s, &user.User{
		Username: userIn.Username,
		Password: userIn.Password,
		Email:    userIn.Email,
		Language: userIn.Language,
	})
	if err != nil {
		_ = s.Rollback()
		return err
	}

	// Create their initial project
	err = models.CreateNewProjectForUser(s, newUser)
	if err != nil {
		_ = s.Rollback()
		return err
	}

	// Admin-only fields are honored only when the caller is an authenticated site admin.
	// The is_admin / skip_email_confirm body fields are silently ignored otherwise; we
	// never trust them from the public register endpoint.
	if isAdmin {
		if userIn.IsAdmin {
			if _, err := s.ID(newUser.ID).Cols("is_admin").Update(&user.User{IsAdmin: true}); err != nil {
				_ = s.Rollback()
				return err
			}
			newUser.IsAdmin = true
		}
		// CreateUser flips status to EmailConfirmationRequired only when the mailer is
		// enabled. Skipping is meaningful in that case; when the mailer is off the user
		// is already Active and this is a harmless no-op.
		if userIn.SkipEmailConfirm || !config.MailerEnabled.GetBool() {
			if err := user.SetUserStatus(s, newUser, user.StatusActive); err != nil {
				_ = s.Rollback()
				return err
			}
			newUser.Status = user.StatusActive
		}
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return err
	}

	return c.JSON(http.StatusOK, newUser)
}
