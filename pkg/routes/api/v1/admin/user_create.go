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
	"errors"
	"net/http"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth/openid"
	"code.vikunja.io/api/pkg/user"

	"github.com/labstack/echo/v5"
)

// CreateUserBody wraps user.APIUserPassword with admin-only fields.
type CreateUserBody struct {
	// The full name of the new user. Optional.
	Name string `json:"name"`
	// The language of the new user. Must be a valid IETF BCP 47 language code and exist in Vikunja.
	Language string `json:"language" valid:"language"`
	user.APIUserPassword
	// Mark the new user as an instance admin.
	IsAdmin bool `json:"is_admin"`
	// Activate the new user immediately without email confirmation.
	SkipEmailConfirm bool `json:"skip_email_confirm"`
}

// CreateUser provisions a new account on behalf of an instance admin.
// @Summary Create a user (admin)
// @Description Create a new local user account. Respects the admin-only fields `is_admin` and `skip_email_confirm`. The public registration toggle is bypassed.
// @tags admin
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param body body admin.CreateUserBody true "The user to create"
// @Success 200 {object} admin.User
// @Failure 400 {object} web.HTTPError
// @Router /admin/users [post]
func CreateUser(c *echo.Context) error {
	body := &CreateUserBody{}
	if err := c.Bind(body); err != nil {
		return c.JSON(http.StatusBadRequest, models.Message{Message: "No or invalid user model provided."})
	}
	if err := c.Validate(body); err != nil {
		e := models.ValidationHTTPError{}
		if is := errors.As(err, &e); is {
			return c.JSON(e.HTTPCode, e)
		}
		return err
	}

	s := db.NewSession()
	defer s.Close()

	newUser, err := models.RegisterUser(s, &user.User{
		Username: body.Username,
		Password: body.Password,
		Email:    body.Email,
		Name:     body.Name,
		Language: body.Language,
	})
	if err != nil {
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

	// Force Active when the admin asked to skip, or when no mailer exists to send the confirmation.
	if body.SkipEmailConfirm || !config.MailerEnabled.GetBool() {
		if err := user.SetUserStatus(s, newUser, user.StatusActive); err != nil {
			_ = s.Rollback()
			return err
		}
		newUser.Status = user.StatusActive
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return err
	}

	// Reload the user so the returned status reflects what was actually persisted
	// (e.g. StatusEmailConfirmationRequired on mail-enabled instances).
	rs := db.NewSession()
	defer rs.Close()
	newUser, err = user.GetUserByID(rs, newUser.ID)
	if err != nil {
		return err
	}

	providers, err := openid.GetAllProviders()
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, newAdminUser(newUser, providers))
}
