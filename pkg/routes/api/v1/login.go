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

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/routes/api/shared"
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

	user, err := shared.AuthenticateUserCredentials(c.Request().Context(), &u)
	if err != nil {
		return err
	}

	// Create token
	return auth.NewUserAuthTokenResponse(user, c, u.LongToken, nil)
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

	result, err := auth.RefreshSession(cookie.Value)
	if err != nil {
		if user2.IsErrUserStatusError(err) {
			auth.ClearRefreshTokenCookie(c)
		}
		return err
	}

	cookieMaxAge := int(config.ServiceJWTTTL.GetInt64())
	if result.IsLongSession {
		cookieMaxAge = int(config.ServiceJWTTTLLong.GetInt64())
	}
	auth.SetRefreshTokenCookie(c, result.NewRefreshToken, cookieMaxAge)

	c.Response().Header().Set("Cache-Control", "no-store")
	return c.JSON(http.StatusOK, auth.Token{Token: result.AccessToken})
}

// LogoutResponse confirms a successful logout and, for sessions created via
// OpenID Connect, carries the provider's RP-Initiated Logout URL the frontend
// should redirect the user agent to so the OP session is ended too.
type LogoutResponse struct {
	Message string `json:"message"`
	// OIDCLogoutURL is the fully-built end_session_endpoint URL (with
	// id_token_hint, post_logout_redirect_uri and client_id). Empty for non-OIDC
	// sessions.
	OIDCLogoutURL string `json:"oidc_logout_url,omitempty"`
}

// Logout deletes the current session from the server.
// @Summary Logout
// @Description Destroys the current session and clears the refresh token cookie. For OpenID Connect sessions the response includes an `oidc_logout_url` the client should redirect to so the provider session is ended too.
// @tags auth
// @Produce json
// @Success 200 {object} v1.LogoutResponse "Successfully logged out."
// @Router /user/logout [post]
func Logout(c *echo.Context) (err error) {
	auth.ClearRefreshTokenCookie(c)

	var sid string
	var userID int64
	if raw := c.Get("user"); raw != nil {
		if jwtinf, ok := raw.(*jwt.Token); ok {
			if claims, ok := jwtinf.Claims.(jwt.MapClaims); ok {
				sid, _ = claims["sid"].(string)
				// Only user tokens carry a sid, but check the type explicitly
				// so a link share id can never be logged as a user id.
				if typ, ok := claims["type"].(float64); ok && int(typ) == auth.AuthTypeUser {
					if id, ok := claims["id"].(float64); ok {
						userID = int64(id)
					}
				}
			}
		}
	}

	oidcLogoutURL, err := shared.LogoutSession(sid)
	if err != nil {
		return err
	}

	if userID != 0 {
		if err := events.DispatchWithContext(c.Request().Context(), &user2.LogoutEvent{UserID: userID}); err != nil {
			log.Errorf("Could not dispatch logout event: %s", err)
		}
	}

	return c.JSON(http.StatusOK, LogoutResponse{
		Message:       "Successfully logged out.",
		OIDCLogoutURL: oidcLogoutURL,
	})
}
