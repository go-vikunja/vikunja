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

	"code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/user"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v5"
)

// passwordChangeAllowedPaths are the only authenticated routes a user who has to change their
// password may call: reading themselves, changing the password and logging out. They are matched
// on the route template, like pathScoped.
var passwordChangeAllowedPaths = map[string]bool{
	"/api/v1/user":          true,
	"/api/v1/user/password": true,
	"/api/v1/user/logout":   true,
	"/api/v2/user":          true,
	"/api/v2/user/password": true,
	"/api/v2/logout":        true,
}

// RequirePasswordChange refuses every request that carries a user token flagged with
// must_change_password, except the routes a user needs to get out of that state. The claim is
// minted from the database each time a token is issued, and setting a password ends all sessions,
// so a token without the claim can never outlive a forced change. API tokens and link shares never
// carry the claim. It has to run after SetupTokenMiddleware.
func RequirePasswordChange() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if !tokenRequiresPasswordChange(c) || passwordChangeAllowedPaths[c.Path()] {
				return next(c)
			}

			return c.JSON(http.StatusForbidden, user.ErrPasswordChangeRequired{}.HTTPError())
		}
	}
}

func tokenRequiresPasswordChange(c *echo.Context) bool {
	if c.Get("api_token") != nil {
		return false
	}
	token, ok := c.Get("user").(*jwt.Token)
	if !ok || token == nil {
		return false
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return false
	}
	return auth.ClaimsRequirePasswordChange(claims)
}
