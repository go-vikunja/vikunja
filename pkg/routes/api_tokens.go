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
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
)

func SetupTokenMiddleware() echo.MiddlewareFunc {
	return echojwt.WithConfig(echojwt.Config{
		SigningKey: []byte(config.ServiceJWTSecret.GetString()),
		Skipper: func(c echo.Context) bool {
			authHeader := c.Request().Header.Values("Authorization")
			if len(authHeader) == 0 {
				return false // let the jwt middleware handle invalid headers
			}

			for _, s := range authHeader {
				if strings.HasPrefix(s, "Bearer "+models.APITokenPrefix) {
					// If the route does not exist, skip the current handling and let the rest of echo's logic handle it
					findCtx := c.Echo().NewContext(c.Request(), c.Response())
					c.Echo().Router().Find(c.Request().Method, echo.GetPath(c.Request()), findCtx)
					if findCtx.Path() == "/api/v1/*" {
						return true
					}

					if c.Request().URL.Path == "/api/v1/token/test" {
						return true
					}

					err := checkAPITokenAndPutItInContext(s, c)
					return err == nil
				}
			}

			return false
		},
		ErrorHandler: func(_ echo.Context, err error) error {
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing, malformed, expired or otherwise invalid token provided").SetInternal(err)
			}

			return nil
		},
	})
}

func checkAPITokenAndPutItInContext(tokenHeaderValue string, c echo.Context) error {
	s := db.NewSession()
	defer s.Close()
	token, err := models.GetTokenFromTokenString(s, strings.TrimPrefix(tokenHeaderValue, "Bearer "))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).SetInternal(err)
	}

	if time.Now().After(token.ExpiresAt) {
		log.Debugf("[auth] Tried authenticating with token %d but it expired on %s", token.ID, token.ExpiresAt.String())
		return echo.NewHTTPError(http.StatusUnauthorized)
	}

	if !models.CanDoAPIRoute(c, token) {
		log.Debugf("[auth] Tried authenticating with token %d but it does not have permission to do this route", token.ID)
		return echo.NewHTTPError(http.StatusUnauthorized)
	}

	u, err := user.GetUserByID(s, token.OwnerID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).SetInternal(err)
	}

	c.Set("api_token", token)
	c.Set("api_user", u)

	return nil
}
