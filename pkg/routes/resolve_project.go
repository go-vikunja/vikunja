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
	"strings"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"

	"github.com/labstack/echo/v5"
)

// ResolveProjectIdentifier accepts either a numeric project id or a project
// identifier (e.g. "PROJ") in the :project path param and rewrites it to the
// numeric id so downstream handlers can bind it as an int64. Pure-digit values
// are always treated as ids, which means identifiers consisting solely of
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

			s := db.NewSession()
			project := &models.Project{}
			has, err := s.Where("identifier = ?", strings.ToUpper(raw)).Get(project)
			_ = s.Close()
			if err != nil {
				return err
			}
			if !has {
				return echo.NewHTTPError(http.StatusNotFound, "Project not found")
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
