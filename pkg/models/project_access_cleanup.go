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
	"code.vikunja.io/api/pkg/web"

	"xorm.io/builder"
	"xorm.io/xorm"
)

// AccessLossCleanup's zero value deletes webhooks and keeps link shares.
type AccessLossCleanup struct {
	// Kept webhooks move to the acting user and follow their access.
	KeepWebhooks     bool
	RevokeLinkShares bool
	// Fails with ErrAccessLossNeedsConfirmation instead of cleaning up if anything is affected.
	Confirm bool
}

// No admin is around to ask during a login-time sync.
func teamSyncAccessLossCleanup() AccessLossCleanup {
	return AccessLossCleanup{
		KeepWebhooks:     true,
		RevokeLinkShares: config.AuthTeamSyncRevokeLinkShares.GetBool(),
	}
}

func cleanupAfterProjectAccessLoss(s *xorm.Session, doer web.Auth, userID int64, projectIDs []int64, cleanup AccessLossCleanup) error {
	return cleanupAfterProjectAccessLossForUsers(s, doer, []int64{userID}, projectIDs, cleanup)
}

func cleanupAfterProjectAccessLossForUsers(s *xorm.Session, doer web.Auth, userIDs []int64, projectIDs []int64, cleanup AccessLossCleanup) error {
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
	for _, id := range candidates {
		if _, has := existing[id]; !has {
			deletedIDs = append(deletedIDs, id)
		}
	}

	return cleanupProjectAccessLoss(s, doer, userIDs, existing, deletedIDs, nil, cleanup)
}

func cleanupProjectAccessLoss(s *xorm.Session, doer web.Auth, userIDs []int64, existing map[int64]*Project, deletedIDs []int64, before projectReadAccess, cleanup AccessLossCleanup) error {
	// Confirm counts and revoked link shares are other users' data a link share can't access.
	if _, isLinkShare := doer.(*LinkSharing); isLinkShare {
		cleanup = AccessLossCleanup{}
	}

	access, err := getProjectReadAccessForUsers(s, userIDs, existing)
	if err != nil {
		return err
	}
	lostByUser := make(map[int64][]int64, len(userIDs))
	for _, userID := range userIDs {
		for id := range existing {
			if !access[userID][id] && (before == nil || before[userID][id]) {
				lostByUser[userID] = append(lostByUser[userID], id)
			}
		}
	}

	if cleanup.Confirm {
		if err := checkAccessLossConfirmed(s, lostByUser); err != nil {
			return err
		}
	}

	doerID := int64(0)
	if u, isUser := doer.(*user.User); isUser {
		doerID = u.ID
	}

	for _, userID := range userIDs {
		lost := lostByUser[userID]
		userCleanup := cleanup
		if userID == doerID {
			userCleanup.KeepWebhooks = false
		}
		err := deleteUserDataInProjects(s, userID, slices.Concat(lost, deletedIDs))
		if err != nil {
			return err
		}
		// Nothing left to keep in deleted projects.
		err = deleteUserWebhooksAndLinkShares(s, userID, deletedIDs, AccessLossCleanup{RevokeLinkShares: true})
		if err != nil {
			return err
		}
		err = deleteUserWebhooksAndLinkShares(s, userID, lost, userCleanup)
		if err != nil {
			return err
		}
	}

	// Delete the doer's original hooks before transferring other users' hooks to them.
	if cleanup.KeepWebhooks {
		for userID, lost := range lostByUser {
			if userID == doerID || len(lost) == 0 {
				continue
			}
			column := "created_by_id"
			webhook := &Webhook{CreatedByID: doerID}
			if doer == nil {
				column = "is_disabled"
				webhook = &Webhook{IsDisabled: true}
			}
			_, err := s.In("project_id", lost).
				And("created_by_id = ?", userID).
				Cols(column).
				NoAutoTime().
				Update(webhook)
			if err != nil {
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

type projectMoveAccess struct {
	userIDs  []int64
	projects map[int64]*Project
	access   projectReadAccess
}

func projectAccessBeforeMove(s *xorm.Session, projectID int64) (*projectMoveAccess, error) {
	subtreeIDs := builder.Select("project_id").From("project_ancestors").Where(builder.Eq{"ancestor_id": projectID})
	inSubtree := builder.Or(builder.Eq{"project_id": projectID}, builder.In("project_id", subtreeIDs))
	taskIDs := builder.Select("id").From("tasks").Where(inSubtree)

	webhookCreatorIDs := []int64{}
	err := s.Table("webhooks").Cols("created_by_id").Where(inSubtree).Find(&webhookCreatorIDs)
	if err != nil {
		return nil, err
	}

	linkShareCreatorIDs := []int64{}
	// DeleteUser keeps link shares, and a missing sharer would count as losing access.
	err = s.Table("link_shares").
		Cols("shared_by_id").
		Where(inSubtree).
		And(builder.In("shared_by_id", builder.Select("id").From("users"))).
		Find(&linkShareCreatorIDs)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	assigneeIDs := []int64{}
	err = s.Table("task_assignees").Cols("user_id").Where(builder.In("task_id", taskIDs)).Find(&assigneeIDs)
	if err != nil {
		return nil, err
	}

	userIDs := slices.Concat(webhookCreatorIDs, linkShareCreatorIDs, subscriberIDs, assigneeIDs)
	slices.Sort(userIDs)
	userIDs = slices.Compact(userIDs)

	projectIDs := []int64{projectID}
	descendants := []int64{}
	if err := s.Table("project_ancestors").Cols("project_id").Where("ancestor_id = ?", projectID).Find(&descendants); err != nil {
		return nil, err
	}
	projects, err := GetProjectsMapByIDs(s, append(projectIDs, descendants...))
	if err != nil {
		return nil, err
	}
	access, err := getProjectReadAccessForUsers(s, userIDs, projects)
	if err != nil {
		return nil, err
	}
	return &projectMoveAccess{
		userIDs:  userIDs,
		projects: projects,
		access:   access,
	}, nil
}

func cleanupAfterProjectMove(s *xorm.Session, doer web.Auth, before *projectMoveAccess, cleanup AccessLossCleanup) error {
	return cleanupProjectAccessLoss(s, doer, before.userIDs, before.projects, nil, before.access, cleanup)
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
