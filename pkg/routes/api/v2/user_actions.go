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

package v2

import (
	"net/http"

	"code.vikunja.io/api/pkg/db"
	v2 "code.vikunja.io/api/pkg/models/v2"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web/handler"
	"github.com/labstack/echo/v4"
)

// ConfirmEmail is the handler to confirm a user email
func ConfirmEmail(c echo.Context) error {
	var emailConfirm v2.EmailConfirm
	if err := c.Bind(&emailConfirm); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "No token provided.")
	}

	s := db.NewSession()
	defer s.Close()

	err := user.ConfirmEmail(s, &user.EmailConfirm{
		Token: emailConfirm.Token,
	})
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

// RequestPasswordResetToken is the handler to change a users password
func RequestPasswordResetToken(c echo.Context) error {
	var pwTokenReset v2.PasswordTokenRequest
	if err := c.Bind(&pwTokenReset); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "No username provided.")
	}

	s := db.NewSession()
	defer s.Close()

	err := user.RequestUserPasswordResetTokenByEmail(s, &user.PasswordTokenRequest{
		Email: pwTokenReset.Username,
	})
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

// ResetPassword is the handler to change a users password
func ResetPassword(c echo.Context) error {
	var pwReset v2.PasswordReset
	if err := c.Bind(&pwReset); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "No password provided.")
	}

	s := db.NewSession()
	defer s.Close()

	err := user.ResetPassword(s, &user.PasswordReset{
		Token:       pwReset.Token,
		NewPassword: pwReset.Password,
	})
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	return c.NoContent(http.StatusNoContent)
}
