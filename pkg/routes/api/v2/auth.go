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

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	v2 "code.vikunja.io/api/pkg/models/v2"
	"code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/modules/auth/ldap"
	"code.vikunja.io/api/pkg/modules/keyvalue"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// Login is the login handler
func Login(c echo.Context) (err error) {
	var loginInfo v2.Login
	if err := c.Bind(&loginInfo); err != nil {
		return c.JSON(http.StatusBadRequest, models.Message{Message: "Please provide a username and password."})
	}

	s := db.NewSession()
	defer s.Close()

	var u *user.User
	if config.AuthLdapEnabled.GetBool() {
		u, err = ldap.AuthenticateUserInLDAP(s, loginInfo.Username, loginInfo.Password, config.AuthLdapGroupSyncEnabled.GetBool(), config.AuthLdapAvatarSyncAttribute.GetString())
		if err != nil && !user.IsErrWrongUsernameOrPassword(err) {
			_ = s.Rollback()
			return handler.HandleHTTPError(err)
		}
	}

	if u == nil {
		// This allows us to still have local users while ldap is enabled
		u, err = user.CheckUserCredentials(s, &user.Login{
			Username: loginInfo.Username,
			Password: loginInfo.Password,
		})
		if err != nil {
			_ = s.Rollback()
			return handler.HandleHTTPError(err)
		}
	}

	if u.Status == user.StatusDisabled {
		_ = s.Rollback()
		return handler.HandleHTTPError(&user.ErrAccountDisabled{UserID: u.ID})
	}

	totpEnabled, err := user.TOTPEnabledForUser(s, u)
	if err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err)
	}

	if totpEnabled {
		if loginInfo.TOTPPasscode == "" {
			_ = s.Rollback()
			return handler.HandleHTTPError(user.ErrInvalidTOTPPasscode{})
		}

		_, err = user.ValidateTOTPPasscode(s, &user.TOTPPasscode{
			User:     u,
			Passcode: loginInfo.TOTPPasscode,
		})
		if err != nil {
			if user.IsErrInvalidTOTPPasscode(err) {
				user.HandleFailedTOTPAuth(s, u)
			}
			_ = s.Rollback()
			return handler.HandleHTTPError(err)
		}
	}

	if err := keyvalue.Del(u.GetFailedTOTPAttemptsKey()); err != nil {
		return err
	}
	if err := keyvalue.Del(u.GetFailedPasswordAttemptsKey()); err != nil {
		return err
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err)
	}

	return auth.NewUserAuthTokenResponse(u, c, loginInfo.LongToken)
}

// RenewToken gives a new token to every user with a valid token
func RenewToken(c echo.Context) (err error) {
	s := db.NewSession()
	defer s.Close()

	jwtinf := c.Get("user").(*jwt.Token)
	claims := jwtinf.Claims.(jwt.MapClaims)
	typ := int(claims["type"].(float64))
	if typ == auth.AuthTypeLinkShare {
		share := &models.LinkSharing{}
		share.ID = int64(claims["id"].(float64))
		err := share.ReadOne(s, share)
		if err != nil {
			_ = s.Rollback()
			return handler.HandleHTTPError(err)
		}
		t, err := auth.NewLinkShareJWTAuthtoken(share)
		if err != nil {
			_ = s.Rollback()
			return handler.HandleHTTPError(err)
		}
		return c.JSON(http.StatusOK, &v2.Token{Token: t})
	}

	u, err := user.GetUserFromClaims(claims)
	if err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err)
	}

	user, err := user.GetUserWithEmail(s, &user.User{ID: u.ID})
	if err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err)
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err)
	}

	var long bool
	lng, has := claims["long"]
	if has {
		long = lng.(bool)
	}

	return auth.NewUserAuthTokenResponse(user, c, long)
}
