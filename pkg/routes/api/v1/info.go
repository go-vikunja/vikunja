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

package v1

import (
	"net/http"

	"code.vikunja.io/api/pkg/routes/api/shared"

	"github.com/labstack/echo/v5"
)

// Info is the handler to get infos about this vikunja instance
// @Summary Info
// @Description Returns the version, frontendurl, motd and various settings of Vikunja
// @tags service
// @Produce json
// @Success 200 {object} shared.VikunjaInfos
// @Router /info [get]
func Info(c *echo.Context) error {
	return c.JSON(http.StatusOK, shared.BuildInfo())
}
