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
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/modules/auth/ldap"
	"code.vikunja.io/api/pkg/modules/keyvalue"
	user2 "code.vikunja.io/api/pkg/user"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v5"
)

// Login is the login handler
// @Summary Login
// @Description Logs a user in. Returns a JWT-Token to authenticate further requests.
// @tags auth
// @Accept json
// @Produce json
// @Param credentials body user.Login true "The login credentials"
// @Success 200 {object} auth.Token
// @Failure 400 {object} models.Message "Invalid user password model."
// @Failure 412 {object} models.Message "Invalid totp passcode."
// @Failure 403 {object} models.Message "Invalid username or password."
// @Router /login [post]
func Login(c *echo.Context) (err error) {
	u := user2.Login{}
	if err := c.Bind(&u); err != nil {
		return c.JSON(http.StatusBadRequest, models.Message{Message: "Please provide a username and password."})
	}

	s := db.NewSession()
	defer s.Close()

	var user *user2.User
	if config.AuthLdapEnabled.GetBool() {
		user, err = ldap.AuthenticateUserInLDAP(s, u.Username, u.Password, config.AuthLdapGroupSyncEnabled.GetBool(), config.AuthLdapAvatarSyncAttribute.GetString())
		if err != nil && !user2.IsErrWrongUsernameOrPassword(err) {
			_ = s.Rollback()
			return err
		}
	}

	if user == nil {
		// This allows us to still have local users while ldap is enabled
		user, err = user2.CheckUserCredentials(s, &u)
		if err != nil {
			_ = s.Rollback()
			return err
		}
	}

	if user.Status == user2.StatusDisabled {
		_ = s.Rollback()
		return &user2.ErrAccountDisabled{UserID: user.ID}
	}

	totpEnabled, err := user2.TOTPEnabledForUser(s, user)
	if err != nil {
		_ = s.Rollback()
		return err
	}

	if totpEnabled {
		if u.TOTPPasscode == "" {
			_ = s.Rollback()
			return user2.ErrInvalidTOTPPasscode{}
		}

		_, err = user2.ValidateTOTPPasscode(s, &user2.TOTPPasscode{
			User:     user,
			Passcode: u.TOTPPasscode,
		})
		if err != nil {
			if user2.IsErrInvalidTOTPPasscode(err) {
				user2.HandleFailedTOTPAuth(s, user)
			}
			_ = s.Rollback()
			return err
		}
	}

	if err := keyvalue.Del(user.GetFailedTOTPAttemptsKey()); err != nil {
		return err
	}
	if err := keyvalue.Del(user.GetFailedPasswordAttemptsKey()); err != nil {
		return err
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return err
	}

	// Create token
	return auth.NewUserAuthTokenResponse(user, c, u.LongToken)
}

// RenewToken renews a link share token only. User tokens must use
// POST /user/token/refresh with a refresh token instead.
// @Summary Renew link share token
// @Description Returns a new valid jwt link share token. Only works for link share tokens.
// @tags auth
// @Accept json
// @Produce json
// @Success 200 {object} auth.Token
// @Failure 400 {object} models.Message "Only link share tokens can be renewed."
// @Router /user/token [post]
func RenewToken(c *echo.Context) (err error) {
	jwtinf := c.Get("user").(*jwt.Token)
	claims := jwtinf.Claims.(jwt.MapClaims)
	typFloat, is := claims["type"].(float64)
	if !is {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid JWT token.")
	}
	typ := int(typFloat)

	if typ == auth.AuthTypeUser {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"User tokens cannot be renewed via this endpoint. Use POST /user/token/refresh with a refresh token.",
		)
	}

	if typ != auth.AuthTypeLinkShare {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid token type.")
	}

	s := db.NewSession()
	defer s.Close()

	share := &models.LinkSharing{}
	idFloat, is := claims["id"].(float64)
	if !is {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid JWT token.")
	}
	share.ID = int64(idFloat)
	err = share.ReadOne(s, share)
	if err != nil {
		_ = s.Rollback()
		return err
	}
	t, err := auth.NewLinkShareJWTAuthtoken(share)
	if err != nil {
		_ = s.Rollback()
		return err
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return err
	}

	return c.JSON(http.StatusOK, auth.Token{Token: t})
}

// RefreshToken exchanges a valid refresh token (sent as an HttpOnly cookie) for
// a new short-lived JWT. The refresh token is rotated on every call.
// @Summary Refresh user token
// @Description Exchanges the refresh token cookie for a new short-lived JWT.
// @tags auth
// @Produce json
// @Success 200 {object} auth.Token
// @Failure 401 {object} models.Message "Invalid or expired refresh token."
// @Router /user/token/refresh [post]
func RefreshToken(c *echo.Context) (err error) {
	cookie, err := c.Cookie(auth.RefreshTokenCookieName)
	if err != nil || cookie.Value == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "No refresh token provided.")
	}
	rawToken := cookie.Value

	s := db.NewSession()
	defer s.Close()

	session, err := models.GetSessionByRefreshToken(s, rawToken)
	if err != nil {
		_ = s.Rollback()
		if models.IsErrSessionNotFound(err) {
			// Don't clear the cookie here — another tab may have already
			// rotated the token, and clearing would overwrite the new cookie.
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or expired refresh token.")
		}
		return err
	}

	// Check if the session has expired based on its type
	maxAge := time.Duration(config.ServiceJWTTTL.GetInt64()) * time.Second
	if session.IsLongSession {
		maxAge = time.Duration(config.ServiceJWTTTLLong.GetInt64()) * time.Second
	}
	if time.Since(session.LastActive) > maxAge {
		if _, err := s.Where("id = ?", session.ID).Delete(&models.Session{}); err != nil {
			_ = s.Rollback()
			return err
		}
		if err := s.Commit(); err != nil {
			return err
		}
		auth.ClearRefreshTokenCookie(c)
		return echo.NewHTTPError(http.StatusUnauthorized, "Session expired.")
	}

	if err := models.UpdateSessionLastActive(s, session.ID); err != nil {
		_ = s.Rollback()
		return err
	}

	newRawToken, err := models.RotateRefreshToken(s, session)
	if err != nil {
		_ = s.Rollback()
		if models.IsErrSessionNotFound(err) {
			// Don't clear the cookie — a concurrent request in another tab
			// may have already rotated the token successfully.
			return echo.NewHTTPError(http.StatusUnauthorized, "Refresh token already used.")
		}
		return err
	}

	u, err := user2.GetUserWithEmail(s, &user2.User{ID: session.UserID})
	if err != nil {
		_ = s.Rollback()
		return err
	}

	if u.Status == user2.StatusDisabled {
		if _, err := s.Where("id = ?", session.ID).Delete(&models.Session{}); err != nil {
			_ = s.Rollback()
			return err
		}
		if err := s.Commit(); err != nil {
			return err
		}
		auth.ClearRefreshTokenCookie(c)
		return &user2.ErrAccountDisabled{UserID: u.ID}
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return err
	}

	t, err := auth.NewUserJWTAuthtoken(u, session.ID)
	if err != nil {
		return err
	}

	cookieMaxAge := int(config.ServiceJWTTTL.GetInt64())
	if session.IsLongSession {
		cookieMaxAge = int(config.ServiceJWTTTLLong.GetInt64())
	}
	auth.SetRefreshTokenCookie(c, newRawToken, cookieMaxAge)

	c.Response().Header().Set("Cache-Control", "no-store")
	return c.JSON(http.StatusOK, auth.Token{Token: t})
}

// Logout deletes the current session from the server.
// @Summary Logout
// @Description Destroys the current session and clears the refresh token cookie.
// @tags auth
// @Produce json
// @Success 200 {object} models.Message "Successfully logged out."
// @Router /user/logout [post]
func Logout(c *echo.Context) (err error) {
	auth.ClearRefreshTokenCookie(c)

	var sid string
	if raw := c.Get("user"); raw != nil {
		if jwtinf, ok := raw.(*jwt.Token); ok {
			if claims, ok := jwtinf.Claims.(jwt.MapClaims); ok {
				sid, _ = claims["sid"].(string)
			}
		}
	}

	if sid == "" {
		return c.JSON(http.StatusOK, models.Message{Message: "Successfully logged out."})
	}

	s := db.NewSession()
	defer s.Close()

	_, err = s.Where("id = ?", sid).Delete(&models.Session{})
	if err != nil {
		_ = s.Rollback()
		return err
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return err
	}

	return c.JSON(http.StatusOK, models.Message{Message: "Successfully logged out."})
}
