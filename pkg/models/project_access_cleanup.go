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
	"slices"

	"code.vikunja.io/api/pkg/user"

	"xorm.io/builder"
	"xorm.io/xorm"
)

func cleanupAfterProjectAccessLoss(s *xorm.Session, userID int64, projectIDs []int64) error {
	return cleanupAfterProjectAccessLossForUsers(s, []int64{userID}, projectIDs)
}

func cleanupAfterProjectAccessLossForUsers(s *xorm.Session, userIDs []int64, projectIDs []int64) error {
	if len(userIDs) == 0 || len(projectIDs) == 0 {
		return nil
	}

	descendantIDs := []int64{}
	err := s.Table("project_ancestors").
		Cols("project_id").
		In("ancestor_id", projectIDs).
		Find(&descendantIDs)
	if err != nil {
		return err
	}

	// Deleted projects have no ancestor rows but may still hold the user's data.
	candidates := slices.Concat(projectIDs, descendantIDs)
	slices.Sort(candidates)
	candidates = slices.Compact(candidates)

	existing, err := GetProjectsMapByIDs(s, candidates)
	if err != nil {
		return err
	}

	deletedIDs := make([]int64, 0, len(candidates))
	existingIDs := make([]int64, 0, len(existing))
	for _, id := range candidates {
		if _, has := existing[id]; has {
			existingIDs = append(existingIDs, id)
			continue
		}
		deletedIDs = append(deletedIDs, id)
	}

	for _, userID := range userIDs {
		permissions, err := checkReadPermissionsForProjects(s, &user.User{ID: userID}, existingIDs)
		if err != nil {
			return err
		}
		lost := slices.Clone(deletedIDs)
		for id, permission := range permissions {
			if !permission.canRead {
				lost = append(lost, id)
			}
		}

		if err := deleteUserDataInProjects(s, userID, lost); err != nil {
			return err
		}
	}

	return nil
}

// Keep in sync with ProjectAccessCleanupList.vue.
func deleteUserDataInProjects(s *xorm.Session, userID int64, projectIDs []int64) error {
	if len(projectIDs) == 0 {
		return nil
	}

	// Subquery, not an id list: one bind param per task can exceed the DB limit.
	taskIDs := builder.Select("id").From("tasks").Where(builder.In("project_id", projectIDs))

	_, err := s.Where(builder.In("task_id", taskIDs)).
		And("user_id = ?", userID).
		Delete(&TaskAssginee{})
	if err != nil {
		return err
	}

	_, err = s.Where(builder.In("entity_id", taskIDs)).
		And("entity_type = ? AND user_id = ?", SubscriptionEntityTask, userID).
		Delete(&Subscription{})
	if err != nil {
		return err
	}

	_, err = s.In("entity_id", projectIDs).
		And("entity_type = ? AND user_id = ?", SubscriptionEntityProject, userID).
		Delete(&Subscription{})
	if err != nil {
		return err
	}

	_, err = s.In("project_id", projectIDs).
		And("created_by_id = ?", userID).
		Delete(&Webhook{})
	return err
}

func teamProjectIDs(s *xorm.Session, teamIDs ...int64) (projectIDs []int64, err error) {
	err = s.Table("team_projects").
		Cols("project_id").
		In("team_id", teamIDs).
		Find(&projectIDs)
	return
}

func teamMemberIDs(s *xorm.Session, teamID int64) (userIDs []int64, err error) {
	err = s.Table("team_members").
		Cols("user_id").
		Where("team_id = ?", teamID).
		Find(&userIDs)
	return
}
