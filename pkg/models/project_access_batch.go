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
	"code.vikunja.io/api/pkg/user"

	"xorm.io/builder"
	"xorm.io/xorm"
)

type projectReadAccess map[int64]map[int64]bool

func getProjectReadAccessForUsers(s *xorm.Session, userIDs []int64, projects map[int64]*Project) (projectReadAccess, error) {
	access := make(projectReadAccess, len(userIDs))
	if len(userIDs) == 0 || len(projects) == 0 {
		return access, nil
	}
	projectIDs := make([]int64, 0, len(projects))
	for id := range projects {
		projectIDs = append(projectIDs, id)
	}
	type grant struct {
		ProjectID  int64
		UserID     int64
		Depth      int
		Permission Permission
	}
	owners := []grant{}
	err := s.Table("projects").
		Select("project_ancestors.project_id, projects.owner_id AS user_id, project_ancestors.depth").
		Join("INNER", "project_ancestors", "project_ancestors.ancestor_id = projects.id").
		In("project_ancestors.project_id", projectIDs).
		In("projects.owner_id", userIDs).
		Find(&owners)
	if err != nil {
		return nil, err
	}
	for i := range owners {
		owners[i].Permission = PermissionAdmin
	}
	direct := []grant{}
	err = s.Table("users_projects").
		Select("project_ancestors.project_id, users_projects.user_id, project_ancestors.depth, users_projects.permission").
		Join("INNER", "project_ancestors", "project_ancestors.ancestor_id = users_projects.project_id").
		In("project_ancestors.project_id", projectIDs).
		In("users_projects.user_id", userIDs).
		Find(&direct)
	if err != nil {
		return nil, err
	}
	teams := []grant{}
	err = s.Table("team_projects").
		Select("project_ancestors.project_id, team_members.user_id, project_ancestors.depth, team_projects.permission").
		Join("INNER", "team_members", "team_members.team_id = team_projects.team_id").
		Join("INNER", "project_ancestors", "project_ancestors.ancestor_id = team_projects.project_id").
		In("project_ancestors.project_id", projectIDs).
		In("team_members.user_id", userIDs).
		Find(&teams)
	if err != nil {
		return nil, err
	}
	type effectiveGrant struct {
		nearest grant
		maximum Permission
	}
	grants := make(map[int64]map[int64]*effectiveGrant, len(userIDs))
	for _, userID := range userIDs {
		access[userID] = map[int64]bool{}
		grants[userID] = map[int64]*effectiveGrant{}
	}
	for _, rows := range [][]grant{
		owners,
		direct,
		teams,
	} {
		for _, row := range rows {
			effective, exists := grants[row.UserID][row.ProjectID]
			if !exists {
				grants[row.UserID][row.ProjectID] = &effectiveGrant{
					nearest: row,
					maximum: row.Permission,
				}
				continue
			}
			effective.maximum = max(effective.maximum, row.Permission)
			if row.Depth < effective.nearest.Depth || (row.Depth == effective.nearest.Depth && row.Permission > effective.nearest.Permission) {
				effective.nearest = row
			}
		}
	}
	for userID, byProject := range grants {
		for projectID, effective := range byProject {
			// Match projectAccessQuery: max at each ancestor, nearest wins unless the maximum is admin.
			permission := effective.nearest.Permission
			if effective.maximum == PermissionAdmin {
				permission = PermissionAdmin
			}
			if permission.isValid() == nil {
				access[userID][projectID] = true
			}
		}
	}
	if license.IsFeatureEnabled(license.FeatureAdminPanel) {
		adminIDs := []int64{}
		err = s.Table("users").Cols("id").
			In("id", userIDs).
			Where(builder.Eq{"is_admin": true}).
			NotIn("status", user.StatusDisabled, user.StatusAccountLocked).
			Find(&adminIDs)
		if err != nil {
			return nil, err
		}
		for _, userID := range adminIDs {
			for _, projectID := range projectIDs {
				access[userID][projectID] = true
			}
		}
	}
	return access, nil
}
