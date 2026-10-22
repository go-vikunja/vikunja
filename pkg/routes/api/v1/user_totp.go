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
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/labstack/echo/v5"
	"xorm.io/xorm"
)

// getLocalUserFromContext is a helper function to get the current local user and database session
func getLocalUserFromContext(c *echo.Context) (*user.User, *xorm.Session, error) {
	s := db.NewSession()

	u, err := user.GetCurrentUserFromDB(s, c)
	if err != nil {
		s.Close()
		return nil, nil, err
	}

	if !u.IsLocalUser() {
		s.Close()
		return nil, nil, &user.ErrAccountIsNotLocal{UserID: u.ID}
	}

	return u, s, nil
}

// UserTOTPEnableRequest holds the body of the totp enable request.
type UserTOTPEnableRequest struct {
	// A passcode generated from the enrolled secret.
	Passcode string `json:"passcode"`
	// The current user's password.
	Password string `json:"password"`
}

func bindInvalidModel(c *echo.Context, v any) error {
	if err := c.Bind(v); err != nil {
		log.Debugf("Invalid model error. Internal error was: %s", err.Error())
		var he *echo.HTTPError
		if errors.As(err, &he) {
			return models.ErrInvalidModel{Message: fmt.Sprintf("%v", he.Message), Err: err}
		}
		return models.ErrInvalidModel{Err: err}
	}
	return nil
}

// UserTOTPEnroll is the handler to enroll a user into totp
// @Summary Enroll a user into totp
// @Description Creates an initial setup for the user in the db. After this step, the user needs to verify they have a working totp setup with the "enable totp" endpoint. Requires the current password.
// @tags user
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param password body v1.UserPasswordConfirmation true "The current user's password."
// @Success 200 {object} user.TOTP
// @Failure 400 {object} web.HTTPError "Something's invalid."
// @Failure 403 {object} web.HTTPError "Wrong password."
// @Failure 404 {object} web.HTTPError "User does not exist."
// @Failure 429 {object} web.HTTPError "Too many failed attempts."
// @Failure 500 {object} models.Message "Internal server error."
// @Router /user/settings/totp/enroll [post]
func UserTOTPEnroll(c *echo.Context) error {
	var pass UserPasswordConfirmation
	if err := bindInvalidModel(c, &pass); err != nil {
		return err
	}

	u, s, err := getLocalUserFromContext(c)
	if err != nil {
		return err
	}
	defer s.Close()

	t, err := user.EnrollTOTP(s, u, pass.Password)
	if err != nil {
		_ = s.Rollback()
		return err
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return err
	}

	return c.JSON(http.StatusOK, t)
}

// UserTOTPEnable is the handler to enable totp for a user
// @Summary Enable a previously enrolled totp setting.
// @Description Enables a previously enrolled totp setting by providing a totp passcode and the current password.
// @tags user
// @Accept json
// @Produce json
// @Param totp body v1.UserTOTPEnableRequest true "The totp passcode and the current password."
// @Security JWTKeyAuth
// @Success 200 {object} models.Message "Successfully enabled"
// @Failure 400 {object} web.HTTPError "Something's invalid."
// @Failure 403 {object} web.HTTPError "Wrong password."
// @Failure 404 {object} web.HTTPError "User does not exist."
// @Failure 412 {object} web.HTTPError "TOTP is not enrolled."
// @Failure 429 {object} web.HTTPError "Too many failed attempts."
// @Failure 500 {object} models.Message "Internal server error."
// @Router /user/settings/totp/enable [post]
func UserTOTPEnable(c *echo.Context) error {
	var body UserTOTPEnableRequest
	if err := bindInvalidModel(c, &body); err != nil {
		return err
	}

	u, s, err := getLocalUserFromContext(c)
	if err != nil {
		return err
	}
	defer s.Close()
	defer events.CleanupPending(s)

	err = user.EnableTOTP(s, u, body.Password, body.Passcode)
	if err != nil {
		_ = s.Rollback()
		return err
	}

	if err := models.DeleteAllUserSessions(s, u.ID); err != nil {
		_ = s.Rollback()
		return err
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return err
	}
	events.DispatchPending(c.Request().Context(), s)

	return c.JSON(http.StatusOK, models.Message{Message: "TOTP was enabled successfully."})
}

// UserTOTPDisable disables totp settings for the current user.
// @Summary Disable totp settings
// @Description Disables any totp settings for the current user. Requires the current password and a current totp passcode.
// @tags user
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param totp body user.Login true "The current user's password and a current totp passcode (totp_passcode)."
// @Success 200 {object} models.Message "Successfully disabled"
// @Failure 400 {object} web.HTTPError "Something's invalid."
// @Failure 403 {object} web.HTTPError "Wrong password."
// @Failure 404 {object} web.HTTPError "User does not exist."
// @Failure 412 {object} web.HTTPError "Invalid totp passcode."
// @Failure 429 {object} web.HTTPError "Too many failed attempts."
// @Failure 500 {object} models.Message "Internal server error."
// @Router /user/settings/totp/disable [post]
func UserTOTPDisable(c *echo.Context) error {
	login := &user.Login{}
	if err := bindInvalidModel(c, login); err != nil {
		return err
	}

	u, s, err := getLocalUserFromContext(c)
	if err != nil {
		return err
	}
	defer s.Close()

	err = user.DisableTOTPWithReauth(s, u, login.Password, login.TOTPPasscode)
	if err != nil {
		_ = s.Rollback()
		return err
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return err
	}

	return c.JSON(http.StatusOK, models.Message{Message: "TOTP was disabled successfully."})
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
func UserTOTPQrCode(c *echo.Context) error {
	u, s, err := getLocalUserFromContext(c)
	if err != nil {
		return err
	}
	defer s.Close()

	qrcode, err := user.GetTOTPQrCodeAsJpegForUser(s, u)
	if err != nil {
		_ = s.Rollback()
		return err
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return err
	}

	return c.Blob(http.StatusOK, "image/jpeg", qrcode)
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
func UserTOTP(c *echo.Context) error {
	u, s, err := getLocalUserFromContext(c)
	if err != nil {
		return err
	}
	defer s.Close()

	t, err := user.GetTOTPForUser(s, u)
	if err != nil {
		_ = s.Rollback()
		return err
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return err
	}

	return c.JSON(http.StatusOK, t.APICopy())
}
