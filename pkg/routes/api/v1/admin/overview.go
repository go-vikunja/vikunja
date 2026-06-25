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

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"

	"github.com/labstack/echo/v5"
)

// GetOverview returns aggregate instance counts and metadata.
// @Summary Admin overview
// @Description Returns per-instance counts (users, projects, shares) plus version and license info. Instance-admin only, gated by the admin_panel feature.
// @tags admin
// @Produce json
// @Security JWTKeyAuth
// @Success 200 {object} models.Overview
// @Failure 404 {object} web.HTTPError
// @Router /admin/overview [get]
func GetOverview(c *echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	overview, err := models.BuildOverview(s)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, overview)
}
