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

// OutsideWebhooks decides what happens to the webhooks of users who lose access to a project.
type OutsideWebhooks int

const (
	// OutsideWebhooksDelete is the zero value so v1 and internal callers keep deleting them.
	OutsideWebhooksDelete OutsideWebhooks = iota
	OutsideWebhooksKeep
	// OutsideWebhooksConfirm fails with ErrOutsideWebhooksNeedConfirmation when any webhook would be deleted.
	OutsideWebhooksConfirm
)

// OutsideWebhooksFromChoice maps the API's nullable clear_outside_webhooks onto an option.
func OutsideWebhooksFromChoice(clear *bool) OutsideWebhooks {
	switch {
	case clear == nil:
		return OutsideWebhooksConfirm
	case *clear:
		return OutsideWebhooksDelete
	default:
		return OutsideWebhooksKeep
	}
}

func cleanupAfterProjectAccessLoss(s *xorm.Session, userID int64, projectIDs []int64, webhooks OutsideWebhooks) error {
	return cleanupAfterProjectAccessLossForUsers(s, []int64{userID}, projectIDs, webhooks)
}

func cleanupAfterProjectAccessLossForUsers(s *xorm.Session, userIDs []int64, projectIDs []int64, webhooks OutsideWebhooks) error {
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

	lostByUser := make(map[int64][]int64, len(userIDs))
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
		lostByUser[userID] = lost
	}

	if webhooks == OutsideWebhooksConfirm {
		count, err := countWebhooksInProjects(s, lostByUser)
		if err != nil {
			return err
		}
		if count > 0 {
			return &ErrOutsideWebhooksNeedConfirmation{Webhooks: count}
		}
	}

	for _, userID := range userIDs {
		if err := deleteUserDataInProjects(s, userID, lostByUser[userID], webhooks); err != nil {
			return err
		}
	}

	return nil
}

func countWebhooksInProjects(s *xorm.Session, projectIDsByUser map[int64][]int64) (total int64, err error) {
	for userID, projectIDs := range projectIDsByUser {
		if len(projectIDs) == 0 {
			continue
		}
		count, err := s.In("project_id", projectIDs).
			And("created_by_id = ?", userID).
			Count(&Webhook{})
		if err != nil {
			return 0, err
		}
		total += count
	}
	return total, nil
}

// Moving a project can drop access inherited from its old ancestors, with no share removed to trigger cleanup.
func cleanupAfterProjectMove(s *xorm.Session, projectID int64, webhooks OutsideWebhooks) error {
	subtreeIDs := builder.Select("project_id").From("project_ancestors").Where(builder.Eq{"ancestor_id": projectID})
	taskIDs := builder.Select("id").From("tasks").Where(builder.In("project_id", subtreeIDs))

	userIDs := []int64{}
	err := s.Table("webhooks").
		Cols("created_by_id").
		Where(builder.In("project_id", subtreeIDs)).
		Find(&userIDs)
	if err != nil {
		return err
	}

	subscriberIDs := []int64{}
	err = s.Table("subscriptions").
		Cols("user_id").
		Where(builder.Or(
			builder.Eq{"entity_type": SubscriptionEntityProject}.And(builder.In("entity_id", subtreeIDs)),
			builder.Eq{"entity_type": SubscriptionEntityTask}.And(builder.In("entity_id", taskIDs)),
		)).
		Find(&subscriberIDs)
	if err != nil {
		return err
	}

	assigneeIDs := []int64{}
	err = s.Table("task_assignees").
		Cols("user_id").
		Where(builder.In("task_id", taskIDs)).
		Find(&assigneeIDs)
	if err != nil {
		return err
	}

	userIDs = slices.Concat(userIDs, subscriberIDs, assigneeIDs)
	slices.Sort(userIDs)
	userIDs = slices.Compact(userIDs)

	return cleanupAfterProjectAccessLossForUsers(s, userIDs, []int64{projectID}, webhooks)
}

// Keep in sync with ProjectAccessCleanupList.vue.
func deleteUserDataInProjects(s *xorm.Session, userID int64, projectIDs []int64, webhooks OutsideWebhooks) error {
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

	if webhooks != OutsideWebhooksDelete {
		return nil
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
