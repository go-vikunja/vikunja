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

package models

import (
	"code.vikunja.io/api/pkg/license"

	"xorm.io/xorm"
)

type ShareCounts struct {
	LinkShares int64 `json:"link_shares" readOnly:"true" doc:"Number of link shares across all projects."`
	TeamShares int64 `json:"team_shares" readOnly:"true" doc:"Number of team-project shares."`
	UserShares int64 `json:"user_shares" readOnly:"true" doc:"Number of user-project shares."`
}

type Overview struct {
	Users    int64        `json:"users" readOnly:"true" doc:"Total number of user accounts."`
	Projects int64        `json:"projects" readOnly:"true" doc:"Total number of projects."`
	Tasks    int64        `json:"tasks" readOnly:"true" doc:"Total number of tasks."`
	Teams    int64        `json:"teams" readOnly:"true" doc:"Total number of teams."`
	Shares   ShareCounts  `json:"shares" readOnly:"true" doc:"Aggregate share counts."`
	License  license.Info `json:"license" readOnly:"true" doc:"Snapshot of the instance license state."`
}

// BuildOverview returns aggregate instance counts plus the current license snapshot.
func BuildOverview(s *xorm.Session) (*Overview, error) {
	users, err := s.Table("users").Count()
	if err != nil {
		return nil, err
	}
	projects, err := s.Table("projects").Count()
	if err != nil {
		return nil, err
	}
	tasks, err := s.Table("tasks").Count()
	if err != nil {
		return nil, err
	}
	teams, err := s.Table("teams").Count()
	if err != nil {
		return nil, err
	}
	linkShares, err := s.Table("link_shares").Count()
	if err != nil {
		return nil, err
	}
	teamShares, err := s.Table("team_projects").Count()
	if err != nil {
		return nil, err
	}
	userShares, err := s.Table("users_projects").Count()
	if err != nil {
		return nil, err
	}

	return &Overview{
		Users:    users,
		Projects: projects,
		Tasks:    tasks,
		Teams:    teams,
		Shares: ShareCounts{
			LinkShares: linkShares,
			TeamShares: teamShares,
			UserShares: userShares,
		},
		License: license.CurrentInfo(),
	}, nil
}
