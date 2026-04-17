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
	"code.vikunja.io/api/pkg/db"
	auth2 "code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/user"

	"github.com/labstack/echo/v5"
)

// RequireSiteAdmin returns a middleware that serves 404 when the caller is not
// a site admin. Same 404 treatment as RequireFeature — the route should look
// identical to an unregistered one from the outside.
//
// The is_admin claim on the JWT is only a hint: a demoted or deleted admin
// keeps the claim until their token expires (up to ServiceJWTTTLShort). We
// therefore re-read is_admin from the DB on every admin-gated request so
// revocation takes effect immediately. A disabled/locked/missing user fails
// the gate as well (GetUserByID surfaces those as errors).
func RequireSiteAdmin() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			a, err := auth2.GetAuthFromClaims(c)
			if err != nil {
				return echo.ErrNotFound
			}
			u, ok := a.(*user.User)
			if !ok {
				return echo.ErrNotFound
			}

			// Close the session before handing off to the downstream handler.
			// On SQLite (used in tests) keeping a read session open while the
			// next handler opens its own write session deadlocks on the users
			// table. The admin check is a single PK lookup — we do not need
			// to hold the session.
			s := db.NewSession()
			fresh, err := user.GetUserByID(s, u.ID)
			_ = s.Close()
			if err != nil || !fresh.IsAdmin {
				return echo.ErrNotFound
			}
			return next(c)
		}
	}
}
