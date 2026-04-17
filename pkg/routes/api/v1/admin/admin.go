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

package admin

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

// Ping is a smoke-test handler for the admin route group.
// @Summary Admin ping
// @Description Returns ok when the admin panel is licensed and the caller is an instance admin. Used to verify gate wiring.
// @tags admin
// @Produce json
// @Security JWTKeyAuth
// @Success 200 {object} map[string]string
// @Failure 404 {object} web.HTTPError
// @Router /admin/ping [get]
func Ping(c *echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}
