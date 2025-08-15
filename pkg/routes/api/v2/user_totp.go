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
	"bytes"
	"image/jpeg"
	"net/http"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	v2 "code.vikunja.io/api/pkg/models/v2"
	"code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web/handler"
	"github.com/labstack/echo/v4"
)

// GetTOTP returns the current totp implementation if any is enabled.
func GetTOTP(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	aut, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return err
	}
	u, err := models.GetUserOrLinkShareUser(s, aut)
	if err != nil {
		return err
	}

	if !u.IsLocalUser() {
		return handler.HandleHTTPError(&user.ErrAccountIsNotLocal{UserID: u.ID})
	}

	t, err := user.GetTOTPForUser(s, u)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	return c.JSON(http.StatusOK, &v2.TOTP{
		Secret: t.Secret,
		URL:    t.URL,
		Links: &v2.TOTPLinks{
			Self:   &v2.Link{Href: "/api/v2/user/totp"},
			QRCode: &v2.Link{Href: "/api/v2/user/totp/qrcode"},
		},
	})
}

// EnrollTOTP is the handler to enroll a user into totp
func EnrollTOTP(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	aut, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return err
	}
	u, err := models.GetUserOrLinkShareUser(s, aut)
	if err != nil {
		return err
	}

	if !u.IsLocalUser() {
		return handler.HandleHTTPError(&user.ErrAccountIsNotLocal{UserID: u.ID})
	}

	t, err := user.EnrollTOTP(s, u)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	return c.JSON(http.StatusOK, &v2.TOTP{
		Secret: t.Secret,
		URL:    t.URL,
		Links: &v2.TOTPLinks{
			Self:   &v2.Link{Href: "/api/v2/user/totp"},
			QRCode: &v2.Link{Href: "/api/v2/user/totp/qrcode"},
		},
	})
}

// GetTOTPQrCode is the handler to show a qr code to enroll the user into totp
func GetTOTPQrCode(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	aut, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return err
	}
	u, err := models.GetUserOrLinkShareUser(s, aut)
	if err != nil {
		return err
	}

	if !u.IsLocalUser() {
		return handler.HandleHTTPError(&user.ErrAccountIsNotLocal{UserID: u.ID})
	}

	qrcode, err := user.GetTOTPQrCodeForUser(s, u)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	buff := &bytes.Buffer{}
	err = jpeg.Encode(buff, qrcode, nil)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	return c.Blob(http.StatusOK, "image/jpeg", buff.Bytes())
}

// EnableTOTP is the handler to enable totp for a user
func EnableTOTP(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	aut, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return err
	}
	u, err := models.GetUserOrLinkShareUser(s, aut)
	if err != nil {
		return err
	}

	if !u.IsLocalUser() {
		return handler.HandleHTTPError(&user.ErrAccountIsNotLocal{UserID: u.ID})
	}

	var passcode v2.TOTPPasscode
	if err := c.Bind(&passcode); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid model provided.")
	}

	err = user.EnableTOTP(s, &user.TOTPPasscode{
		User:     u,
		Passcode: passcode.Passcode,
	})
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

// DisableTOTP disables totp settings for the current user.
func DisableTOTP(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	aut, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return err
	}
	u, err := models.GetUserOrLinkShareUser(s, aut)
	if err != nil {
		return err
	}

	if !u.IsLocalUser() {
		return handler.HandleHTTPError(&user.ErrAccountIsNotLocal{UserID: u.ID})
	}

	err = user.DisableTOTP(s, u)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	return c.NoContent(http.StatusNoContent)
}
