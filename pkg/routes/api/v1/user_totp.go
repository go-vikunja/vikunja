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
	"bytes"
	"errors"
	"fmt"
	"image/jpeg"
	"net/http"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/labstack/echo/v4"
)

// UserTOTPEnroll is the handler to enroll a user into totp
// @Summary Enroll a user into totp
// @Description Creates an initial setup for the user in the db. After this step, the user needs to verify they have a working totp setup with the "enable totp" endpoint.
// @tags user
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Success 200 {object} user.TOTP
// @Failure 400 {object} web.HTTPError "Something's invalid."
// @Failure 404 {object} web.HTTPError "User does not exist."
// @Failure 500 {object} models.Message "Internal server error."
// @Router /user/settings/totp/enroll [post]
func UserTOTPEnroll(c echo.Context) error {
	u, err := user.GetCurrentUser(c)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	// Check if the user is a local user
	if !u.IsLocalUser() {
		return handler.HandleHTTPError(&user.ErrAccountIsNotLocal{UserID: u.ID})
	}

	s := db.NewSession()
	defer s.Close()

	t, err := user.EnrollTOTP(s, u)
	if err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err)
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err)
	}

	return c.JSON(http.StatusOK, t)
}

// UserTOTPEnable is the handler to enable totp for a user
// @Summary Enable a previously enrolled totp setting.
// @Description Enables a previously enrolled totp setting by providing a totp passcode.
// @tags user
// @Accept json
// @Produce json
// @Param totp body user.TOTPPasscode true "The totp passcode."
// @Security JWTKeyAuth
// @Success 200 {object} models.Message "Successfully enabled"
// @Failure 400 {object} web.HTTPError "Something's invalid."
// @Failure 404 {object} web.HTTPError "User does not exist."
// @Failure 412 {object} web.HTTPError "TOTP is not enrolled."
// @Failure 500 {object} models.Message "Internal server error."
// @Router /user/settings/totp/enable [post]
func UserTOTPEnable(c echo.Context) error {
	u, err := user.GetCurrentUser(c)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	// Check if the user is a local user
	if !u.IsLocalUser() {
		return handler.HandleHTTPError(&user.ErrAccountIsNotLocal{UserID: u.ID})
	}

	passcode := &user.TOTPPasscode{
		User: u,
	}
	if err := c.Bind(passcode); err != nil {
		log.Debugf("Invalid model error. Internal error was: %s", err.Error())
		var he *echo.HTTPError
		if errors.As(err, &he) {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid model provided. Error was: %s", he.Message)).SetInternal(err)
		}
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid model provided.").SetInternal(err)
	}

	s := db.NewSession()
	defer s.Close()

	err = user.EnableTOTP(s, passcode)
	if err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err)
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err)
	}

	return c.JSON(http.StatusOK, models.Message{Message: "TOTP was enabled successfully."})
}

// UserTOTPDisable disables totp settings for the current user.
// @Summary Disable totp settings
// @Description Disables any totp settings for the current user.
// @tags user
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param totp body user.Login true "The current user's password (only password is enough)."
// @Success 200 {object} models.Message "Successfully disabled"
// @Failure 400 {object} web.HTTPError "Something's invalid."
// @Failure 404 {object} web.HTTPError "User does not exist."
// @Failure 500 {object} models.Message "Internal server error."
// @Router /user/settings/totp/disable [post]
func UserTOTPDisable(c echo.Context) error {
	login := &user.Login{}
	if err := c.Bind(login); err != nil {
		log.Debugf("Invalid model error. Internal error was: %s", err.Error())
		var he *echo.HTTPError
		if errors.As(err, &he) {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid model provided. Error was: %s", he.Message)).SetInternal(err)
		}
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid model provided.").SetInternal(err)
	}

	u, err := user.GetCurrentUser(c)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	// Check if the user is a local user
	if !u.IsLocalUser() {
		return handler.HandleHTTPError(&user.ErrAccountIsNotLocal{UserID: u.ID})
	}

	s := db.NewSession()
	defer s.Close()

	u, err = user.GetUserByID(s, u.ID)
	if err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err)
	}

	err = user.CheckUserPassword(u, login.Password)
	if err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err)
	}

	err = user.DisableTOTP(s, u)
	if err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err)
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err)
	}

	return c.JSON(http.StatusOK, models.Message{Message: "TOTP was enabled successfully."})
}

// UserTOTPQrCode is the handler to show a qr code to enroll the user into totp
// @Summary Totp QR Code
// @Description Returns a qr code for easier setup at end user's devices.
// @tags user
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Success 200 {file} blob "The qr code as jpeg image"
// @Failure 500 {object} models.Message "Internal server error."
// @Router /user/settings/totp/qrcode [get]
func UserTOTPQrCode(c echo.Context) error {
	u, err := user.GetCurrentUser(c)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	// Check if the user is a local user
	if !u.IsLocalUser() {
		return handler.HandleHTTPError(&user.ErrAccountIsNotLocal{UserID: u.ID})
	}

	s := db.NewSession()
	defer s.Close()

	qrcode, err := user.GetTOTPQrCodeForUser(s, u)
	if err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err)
	}

	buff := &bytes.Buffer{}
	err = jpeg.Encode(buff, qrcode, nil)
	if err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err)
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err)
	}

	return c.Blob(http.StatusOK, "image/jpeg", buff.Bytes())
}

// UserTOTP returns the current totp implementation if any is enabled.
// @Summary Totp setting for the current user
// @Description Returns the current user totp setting or an error if it is not enabled.
// @tags user
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Success 200 {object} user.TOTP "The totp settings."
// @Failure 500 {object} models.Message "Internal server error."
// @Router /user/settings/totp [get]
func UserTOTP(c echo.Context) error {
	u, err := user.GetCurrentUser(c)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	// Check if the user is a local user
	if !u.IsLocalUser() {
		return handler.HandleHTTPError(&user.ErrAccountIsNotLocal{UserID: u.ID})
	}

	s := db.NewSession()
	defer s.Close()

	t, err := user.GetTOTPForUser(s, u)
	if err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err)
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err)
	}

	return c.JSON(http.StatusOK, t)
}
