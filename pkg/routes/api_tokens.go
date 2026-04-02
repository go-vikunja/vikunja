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

package routes

import (
	"net/http"
	"strings"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/web"

	echojwt "github.com/labstack/echo-jwt/v5"
	"github.com/labstack/echo/v5"
)

// ErrCodeInvalidToken is the error code returned when the JWT is missing,
// malformed, or expired. The frontend uses this to distinguish "token expired,
// try refreshing" from other 401s (disabled account, wrong API token, etc.).
const ErrCodeInvalidToken = 11
const apiTokenAuthErrorContextKey = "api_token_auth_error"

func SetupTokenMiddleware() echo.MiddlewareFunc {
	return echojwt.WithConfig(echojwt.Config{
		SigningKey: []byte(config.ServiceSecret.GetString()),
		Skipper: func(c *echo.Context) bool {
			authHeader := c.Request().Header.Values("Authorization")
			if len(authHeader) == 0 {
				return false // let the jwt middleware handle invalid headers
			}

			for _, s := range authHeader {
				if strings.HasPrefix(s, "Bearer "+models.APITokenPrefix) {
					if c.Request().URL.Path == "/api/v1/token/test" {
						return true
					}

					err := checkAPITokenAndPutItInContext(s, c)
					if err != nil {
						c.Set(apiTokenAuthErrorContextKey, err)
					}
					return err == nil
				}
			}

			return false
		},
		ErrorHandler: func(c *echo.Context, err error) error {
			if apiTokenErr, ok := c.Get(apiTokenAuthErrorContextKey).(*echo.HTTPError); ok {
				statusCode := http.StatusUnauthorized
				if code, isCode := apiTokenErr.Code.(int); isCode {
					statusCode = code
				}

				message := http.StatusText(statusCode)
				if msg, hasMsg := apiTokenErr.Message.(string); hasMsg && msg != "" {
					message = msg
				}

				errorCode := models.ErrCodeAPITokenInvalid
				if statusCode == http.StatusForbidden {
					errorCode = models.ErrCodeInvalidAPITokenPermission
				}

				return c.JSON(statusCode, web.HTTPError{
					HTTPCode: statusCode,
					Code:     errorCode,
					Message:  message,
				})
			}

			if err != nil {
				return c.JSON(http.StatusUnauthorized, web.HTTPError{
					HTTPCode: http.StatusUnauthorized,
					Code:     ErrCodeInvalidToken,
					Message:  "missing, malformed, expired or otherwise invalid token provided",
				})
			}

			return nil
		},
	})
}

func checkAPITokenAndPutItInContext(tokenHeaderValue string, c *echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	token, u, err := models.ValidateTokenAndGetOwner(s, strings.TrimPrefix(tokenHeaderValue, "Bearer "))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal Server Error").Wrap(err)
	}
	if token == nil || u == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or expired API token.")
	}

	if !models.CanDoAPIRoute(c, token) {
		log.Debugf("[auth] Tried authenticating with token %d but it does not have permission to do this route", token.ID)
		return echo.NewHTTPError(http.StatusForbidden, "API token does not have permission for this route.")
	}

	c.Set("api_token", token)
	c.Set("api_user", u)

	return nil
}
