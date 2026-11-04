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

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/user"

	"xorm.io/builder"
	"xorm.io/xorm"
)

// AccessLossCleanup decides what happens to the webhooks and link shares of users who lose access to a project.
// The zero value deletes their webhooks and keeps their link shares.
type AccessLossCleanup struct {
	// Kept webhooks move to the project owner: delivery checks the creator's access, so they'd stop firing otherwise.
	KeepWebhooks     bool
	RevokeLinkShares bool
	// Confirm fails with ErrAccessLossNeedsConfirmation instead of cleaning up when any webhook or link share is affected.
	Confirm bool
}

// No admin is around to ask during a login-time sync.
func teamSyncAccessLossCleanup() AccessLossCleanup {
	return AccessLossCleanup{
		KeepWebhooks:     config.AuthTeamSyncTransferWebhooks.GetBool(),
		RevokeLinkShares: config.AuthTeamSyncRevokeLinkShares.GetBool(),
	}
}

func cleanupAfterProjectAccessLoss(s *xorm.Session, userID int64, projectIDs []int64, cleanup AccessLossCleanup) error {
	return cleanupAfterProjectAccessLossForUsers(s, []int64{userID}, projectIDs, cleanup)
}

func cleanupAfterProjectAccessLossForUsers(s *xorm.Session, userIDs []int64, projectIDs []int64, cleanup AccessLossCleanup) error {
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

	// Unscoped: trashed projects keep their access until they are purged.
	existing := map[int64]*Project{}
	if err := s.Unscoped().In("id", candidates).Find(&existing); err != nil {
		return err
	}

	deletedIDs := make([]int64, 0, len(candidates))
	liveIDs := make([]int64, 0, len(existing))
	trashedIDs := []int64{}
	for _, id := range candidates {
		project, has := existing[id]
		switch {
		case !has:
			deletedIDs = append(deletedIDs, id)
		case project.DeletedAt.IsZero():
			liveIDs = append(liveIDs, id)
		default:
			trashedIDs = append(trashedIDs, id)
		}
	}

	lostByUser := make(map[int64][]int64, len(userIDs))
	for _, userID := range userIDs {
		permissions, err := checkReadPermissionsForProjects(s, &user.User{ID: userID}, liveIDs)
		if err != nil {
			return err
		}
		lost := []int64{}
		for id, permission := range permissions {
			if !permission.canRead {
				lost = append(lost, id)
			}
		}
		if len(trashedIDs) > 0 {
			access, err := getProjectAccessIncludingTrash(s, userID)
			if err != nil {
				return err
			}
			for _, id := range trashedIDs {
				if _, has := access.permission(id); !has {
					lost = append(lost, id)
				}
			}
		}
		lostByUser[userID] = lost
	}

	if cleanup.Confirm {
		if err := checkAccessLossConfirmed(s, lostByUser); err != nil {
			return err
		}
	}

	for _, userID := range userIDs {
		lost := lostByUser[userID]
		err := deleteUserDataInProjects(s, userID, slices.Concat(lost, deletedIDs))
		if err != nil {
			return err
		}
		// Nothing left to keep in deleted projects.
		err = deleteUserWebhooksAndLinkShares(s, userID, deletedIDs, AccessLossCleanup{RevokeLinkShares: true})
		if err != nil {
			return err
		}
		err = deleteUserWebhooksAndLinkShares(s, userID, lost, cleanup)
		if err != nil {
			return err
		}
		if cleanup.KeepWebhooks {
			if err := transferWebhooksToProjectOwners(s, userID, lost, existing); err != nil {
				return err
			}
		}
	}

	return nil
}

func checkAccessLossConfirmed(s *xorm.Session, projectIDsByUser map[int64][]int64) error {
	affected := &ErrAccessLossNeedsConfirmation{}
	for userID, projectIDs := range projectIDsByUser {
		if len(projectIDs) == 0 {
			continue
		}
		webhooks, err := s.In("project_id", projectIDs).
			And("created_by_id = ?", userID).
			Count(&Webhook{})
		if err != nil {
			return err
		}
		linkShares, err := s.In("project_id", projectIDs).
			And("shared_by_id = ?", userID).
			Count(&LinkSharing{})
		if err != nil {
			return err
		}
		affected.Webhooks += webhooks
		affected.LinkShares += linkShares
	}
	if affected.Webhooks > 0 || affected.LinkShares > 0 {
		return affected
	}
	return nil
}

// Moving a project can drop access inherited from its old ancestors, with no share removed to trigger cleanup.
func cleanupAfterProjectMove(s *xorm.Session, projectID int64, cleanup AccessLossCleanup) error {
	subtreeIDs := builder.Select("project_id").From("project_ancestors").Where(builder.Eq{"ancestor_id": projectID})
	inSubtree := builder.Or(builder.Eq{"project_id": projectID}, builder.In("project_id", subtreeIDs))
	taskIDs := builder.Select("id").From("tasks").Where(inSubtree)

	webhookCreatorIDs := []int64{}
	err := s.Table("webhooks").Cols("created_by_id").Where(inSubtree).Find(&webhookCreatorIDs)
	if err != nil {
		return err
	}

	linkShareCreatorIDs := []int64{}
	err = s.Table("link_shares").Cols("shared_by_id").Where(inSubtree).Find(&linkShareCreatorIDs)
	if err != nil {
		return err
	}

	subscriberIDs := []int64{}
	err = s.Table("subscriptions").
		Cols("user_id").
		Where(builder.Or(
			builder.Eq{"entity_type": SubscriptionEntityProject}.And(
				builder.Or(builder.Eq{"entity_id": projectID}, builder.In("entity_id", subtreeIDs)),
			),
			builder.Eq{"entity_type": SubscriptionEntityTask}.And(builder.In("entity_id", taskIDs)),
		)).
		Find(&subscriberIDs)
	if err != nil {
		return err
	}

	assigneeIDs := []int64{}
	err = s.Table("task_assignees").Cols("user_id").Where(builder.In("task_id", taskIDs)).Find(&assigneeIDs)
	if err != nil {
		return err
	}

	userIDs := slices.Concat(webhookCreatorIDs, linkShareCreatorIDs, subscriberIDs, assigneeIDs)
	slices.Sort(userIDs)
	userIDs = slices.Compact(userIDs)
	if len(userIDs) == 0 {
		return nil
	}

	// DeleteUser keeps link shares, and a missing user would count as losing access.
	existingUserIDs := []int64{}
	err = s.Table("users").In("id", userIDs).Cols("id").Find(&existingUserIDs)
	if err != nil {
		return err
	}

	return cleanupAfterProjectAccessLossForUsers(s, existingUserIDs, []int64{projectID}, cleanup)
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
	return err
}

func deleteUserWebhooksAndLinkShares(s *xorm.Session, userID int64, projectIDs []int64, cleanup AccessLossCleanup) error {
	if len(projectIDs) == 0 {
		return nil
	}

	if !cleanup.KeepWebhooks {
		_, err := s.In("project_id", projectIDs).
			And("created_by_id = ?", userID).
			Delete(&Webhook{})
		if err != nil {
			return err
		}
	}

	if cleanup.RevokeLinkShares {
		_, err := s.In("project_id", projectIDs).
			And("shared_by_id = ?", userID).
			Delete(&LinkSharing{})
		if err != nil {
			return err
		}
	}
	return nil
}

func transferWebhooksToProjectOwners(s *xorm.Session, userID int64, projectIDs []int64, projects map[int64]*Project) error {
	projectIDsByOwner := map[int64][]int64{}
	for _, id := range projectIDs {
		owner := projects[id].OwnerID
		projectIDsByOwner[owner] = append(projectIDsByOwner[owner], id)
	}

	for ownerID, ids := range projectIDsByOwner {
		_, err := s.In("project_id", ids).
			And("created_by_id = ?", userID).
			Cols("created_by_id").
			NoAutoTime().
			Update(&Webhook{CreatedByID: ownerID})
		if err != nil {
			return err
		}
	}
	return nil
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
