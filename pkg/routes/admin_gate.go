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

// RequireInstanceAdmin serves 404 to make admin routes indistinguishable from unregistered ones,
// and re-reads is_admin from the DB so demotion takes effect before the JWT expires.
func RequireInstanceAdmin() echo.MiddlewareFunc {
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

			// Close before next() — SQLite deadlocks when a read session is held across a write on users.
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
