// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package routes

import (
	"net/http"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"

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
					err := checkAPITokenAndPutItInContext(s, c)
					if err != nil {
						return false
					}
					return true
				}
			}

			return false
		},
		ErrorHandler: func(c echo.Context, err error) error {
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
		return echo.NewHTTPError(http.StatusUnauthorized)
	}

	if !models.CanDoAPIRoute(c, token) {
		return echo.NewHTTPError(http.StatusUnauthorized)
	}

	c.Set("api_token", token)

	return nil
}
