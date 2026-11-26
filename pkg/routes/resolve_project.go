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
	"strconv"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	auth2 "code.vikunja.io/api/pkg/modules/auth"

	"github.com/labstack/echo/v5"
)

// ResolveProjectIdentifier accepts either a numeric project id or a project
// identifier (e.g. "PROJ") in the :project path param and rewrites it to the
// numeric id so downstream handlers can bind it as an int64. Identifiers are
// resolved among the projects the caller can read. Pure-digit values are always treated as ids, which means identifiers consisting solely of
// digits are unreachable via this route.
func ResolveProjectIdentifier() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			raw := c.Param("project")
			if raw == "" {
				return next(c)
			}
			if _, err := strconv.ParseInt(raw, 10, 64); err == nil {
				return next(c)
			}

			a, err := auth2.GetAuthFromClaims(c)
			if err != nil {
				return err
			}

			s := db.NewSession()
			project, err := models.GetProjectSimpleByIdentifier(s, a, raw)
			_ = s.Close()
			if models.IsErrProjectDoesNotExist(err) {
				return echo.NewHTTPError(http.StatusNotFound, "Project not found")
			}
			if err != nil {
				return err
			}

			values := c.PathValues()
			for i, v := range values {
				if v.Name == "project" {
					values[i].Value = strconv.FormatInt(project.ID, 10)
					break
				}
			}
			c.SetPathValues(values)

			return next(c)
		}
	}
}
