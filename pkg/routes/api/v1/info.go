//   Vikunja is a todo-list application to facilitate your life.
//   Copyright 2019 Vikunja and contributors. All rights reserved.
//
//   This program is free software: you can redistribute it and/or modify
//   it under the terms of the GNU General Public License as published by
//   the Free Software Foundation, either version 3 of the License, or
//   (at your option) any later version.
//
//   This program is distributed in the hope that it will be useful,
//   but WITHOUT ANY WARRANTY; without even the implied warranty of
//   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//   GNU General Public License for more details.
//
//   You should have received a copy of the GNU General Public License
//   along with this program.  If not, see <https://www.gnu.org/licenses/>.

package v1

import (
	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/version"
	"github.com/labstack/echo/v4"
	"net/http"
)

type vikunjaInfos struct {
	Version            string `json:"version"`
	FrontendURL        string `json:"frontend_url"`
	Motd               string `json:"motd"`
	LinkSharingEnabled bool   `json:"link_sharing_enabled"`
}

// Info is the handler to get infos about this vikunja instance
// @Summary Info
// @Description Returns the version, frontendurl, motd and various settings of Vikunja
// @tags service
// @Produce json
// @Success 200 {object} v1.vikunjaInfos
// @Router /info [get]
func Info(c echo.Context) error {
	return c.JSON(http.StatusOK, vikunjaInfos{
		Version:            version.Version,
		FrontendURL:        config.ServiceFrontendurl.GetString(),
		Motd:               config.ServiceMotd.GetString(),
		LinkSharingEnabled: config.ServiceEnableLinkSharing.GetBool(),
	})
}
