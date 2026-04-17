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
	"code.vikunja.io/api/pkg/license"
	"code.vikunja.io/api/pkg/version"
	"github.com/labstack/echo/v5"
)

type ShareCounts struct {
	LinkShares int64 `json:"link_shares"`
	TeamShares int64 `json:"team_shares"`
	UserShares int64 `json:"user_shares"`
}

type Overview struct {
	Users    int64        `json:"users"`
	Projects int64        `json:"projects"`
	Tasks    int64        `json:"tasks"`
	Teams    int64        `json:"teams"`
	Shares   ShareCounts  `json:"shares"`
	Version  string       `json:"version"`
	License  license.Info `json:"license"`
}

// GetOverview returns aggregate instance counts and metadata.
// @Summary Admin overview
// @Description Returns per-instance counts (users, projects, shares) plus version and license info. Instance-admin only, gated by the admin_panel feature.
// @tags admin
// @Produce json
// @Security JWTKeyAuth
// @Success 200 {object} admin.Overview
// @Failure 404 {object} web.HTTPError
// @Router /admin/overview [get]
func GetOverview(c *echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	users, err := s.Table("users").Count()
	if err != nil {
		return err
	}
	projects, err := s.Table("projects").Count()
	if err != nil {
		return err
	}
	tasks, err := s.Table("tasks").Count()
	if err != nil {
		return err
	}
	teams, err := s.Table("teams").Count()
	if err != nil {
		return err
	}
	linkShares, err := s.Table("link_shares").Count()
	if err != nil {
		return err
	}
	teamShares, err := s.Table("team_projects").Count()
	if err != nil {
		return err
	}
	userShares, err := s.Table("users_projects").Count()
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, Overview{
		Users:    users,
		Projects: projects,
		Tasks:    tasks,
		Teams:    teams,
		Shares: ShareCounts{
			LinkShares: linkShares,
			TeamShares: teamShares,
			UserShares: userShares,
		},
		Version: version.Version,
		License: license.CurrentInfo(),
	})
}
