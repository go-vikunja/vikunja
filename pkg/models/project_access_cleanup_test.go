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
	"fmt"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

// GHSA-hjx8-qv73-f7cm
type accessLossFixture struct {
	s        *xorm.Session
	owner    *user.User
	member   *user.User
	parent   *Project
	child    *Project
	task     *Task
	webhookP int64
	webhookC int64
	shareP   int64
	shareC   int64
}

func setupAccessLoss(t *testing.T) *accessLossFixture {
	t.Helper()
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	t.Cleanup(func() { s.Close() })

	f := &accessLossFixture{
		s:      s,
		owner:  &user.User{ID: 1, Username: "user1"},
		member: &user.User{ID: 2, Username: "user2"},
	}

	f.parent = &Project{Title: "access loss parent"}
	require.NoError(t, f.parent.Create(s, f.owner))
	parentID := f.parent.ID
	f.child = &Project{Title: "access loss child", ParentProjectID: &parentID}
	require.NoError(t, f.child.Create(s, f.owner))

	f.task = &Task{Title: "access loss", ProjectID: f.child.ID}
	require.NoError(t, f.task.Create(s, f.owner))
	_, err := s.Insert(&TaskAssginee{TaskID: f.task.ID, UserID: f.member.ID})
	require.NoError(t, err)
	_, err = s.Insert(&Subscription{EntityType: SubscriptionEntityTask, EntityID: f.task.ID, UserID: f.member.ID})
	require.NoError(t, err)
	_, err = s.Insert(&Subscription{EntityType: SubscriptionEntityProject, EntityID: f.parent.ID, UserID: f.member.ID})
	require.NoError(t, err)

	f.webhookP = f.insertWebhook(t, f.parent.ID, f.member.ID)
	f.webhookC = f.insertWebhook(t, f.child.ID, f.member.ID)

	insertLinkShare := func(projectID int64) int64 {
		share := &LinkSharing{
			Hash:        fmt.Sprintf("access-loss-%d", projectID),
			ProjectID:   projectID,
			Permission:  PermissionRead,
			SharingType: SharingTypeWithoutPassword,
			SharedByID:  f.member.ID,
		}
		_, err := s.Insert(share)
		require.NoError(t, err)
		return share.ID
	}
	f.shareP = insertLinkShare(f.parent.ID)
	f.shareC = insertLinkShare(f.child.ID)
	return f
}

func (f *accessLossFixture) insertWebhook(t *testing.T, projectID, createdByID int64) int64 {
	t.Helper()
	w := &Webhook{
		TargetURL:   "https://example.com/access-loss",
		Events:      []string{"task.updated"},
		ProjectID:   projectID,
		CreatedByID: createdByID,
	}
	_, err := f.s.Insert(w)
	require.NoError(t, err)
	return w.ID
}

func (f *accessLossFixture) otherUserWithWebhook(t *testing.T) (other *user.User, webhookID int64) {
	t.Helper()
	other = &user.User{ID: 3, Username: "user3"}
	return other, f.insertWebhook(t, f.child.ID, other.ID)
}

func (f *accessLossFixture) addOtherMemberWithWebhook(t *testing.T, team *Team) (other *user.User, webhookID int64) {
	t.Helper()
	other, webhookID = f.otherUserWithWebhook(t)
	require.NoError(t, (&TeamMember{TeamID: team.ID, Username: other.Username}).Create(f.s, f.owner))
	return other, webhookID
}

func (f *accessLossFixture) shareDirectly(t *testing.T, perm Permission) {
	t.Helper()
	pu := &ProjectUser{ProjectID: f.parent.ID, Username: f.member.Username, Permission: perm}
	require.NoError(t, pu.Create(f.s, f.owner))
}

func (f *accessLossFixture) shareViaTeam(t *testing.T) *Team {
	t.Helper()
	team := &Team{Name: "access loss team"}
	require.NoError(t, team.Create(f.s, f.owner))
	require.NoError(t, (&TeamMember{TeamID: team.ID, Username: f.member.Username}).Create(f.s, f.owner))
	require.NoError(t, (&TeamProject{TeamID: team.ID, ProjectID: f.parent.ID, Permission: PermissionWrite}).Create(f.s, f.owner))
	return team
}

func (f *accessLossFixture) giveParentTo(t *testing.T, u *user.User) {
	t.Helper()
	_, err := f.s.ID(f.parent.ID).Cols("owner_id").Update(&Project{OwnerID: u.ID})
	require.NoError(t, err)
}

func (f *accessLossFixture) assertCleanedUp(t *testing.T) {
	t.Helper()
	require.NoError(t, f.s.Commit())
	db.AssertMissing(t, "webhooks", map[string]interface{}{"id": f.webhookP})
	db.AssertMissing(t, "webhooks", map[string]interface{}{"id": f.webhookC})
	db.AssertMissing(t, "task_assignees", map[string]interface{}{"task_id": f.task.ID, "user_id": f.member.ID})
	db.AssertMissing(t, "subscriptions", map[string]interface{}{"entity_type": SubscriptionEntityTask, "entity_id": f.task.ID, "user_id": f.member.ID})
	db.AssertMissing(t, "subscriptions", map[string]interface{}{"entity_type": SubscriptionEntityProject, "entity_id": f.parent.ID, "user_id": f.member.ID})
	db.AssertExists(t, "link_shares", map[string]interface{}{"id": f.shareP}, false)
	db.AssertExists(t, "link_shares", map[string]interface{}{"id": f.shareC}, false)
}

func (f *accessLossFixture) assertLinkSharesRevoked(t *testing.T) {
	t.Helper()
	db.AssertMissing(t, "link_shares", map[string]interface{}{"id": f.shareP})
	db.AssertMissing(t, "link_shares", map[string]interface{}{"id": f.shareC})
}

func (f *accessLossFixture) assertWebhooksDeleted(t *testing.T) {
	t.Helper()
	db.AssertMissing(t, "webhooks", map[string]interface{}{"id": f.webhookP})
	db.AssertMissing(t, "webhooks", map[string]interface{}{"id": f.webhookC})
}

func (f *accessLossFixture) assertWebhooksTransferredTo(t *testing.T, u *user.User) {
	t.Helper()
	db.AssertExists(t, "webhooks", map[string]interface{}{"id": f.webhookP, "created_by_id": u.ID}, false)
	db.AssertExists(t, "webhooks", map[string]interface{}{"id": f.webhookC, "created_by_id": u.ID}, false)
}

func (f *accessLossFixture) moveChildToTopLevel(t *testing.T, cleanup AccessLossCleanup, askBeforeAccessLoss bool) error {
	t.Helper()
	return f.moveChildToTopLevelAs(t, f.owner, cleanup, askBeforeAccessLoss)
}

func (f *accessLossFixture) moveChildToTopLevelAs(t *testing.T, doer web.Auth, cleanup AccessLossCleanup, askBeforeAccessLoss bool) error {
	t.Helper()
	detached := int64(0)
	child := &Project{
		ID:                  f.child.ID,
		Title:               f.child.Title,
		ParentProjectID:     &detached,
		AccessLossCleanup:   cleanup,
		AskBeforeAccessLoss: askBeforeAccessLoss,
	}
	return child.Update(f.s, doer)
}

func (f *accessLossFixture) assertWebhooksKept(t *testing.T) {
	t.Helper()
	require.NoError(t, f.s.Commit())
	db.AssertExists(t, "webhooks", map[string]interface{}{"id": f.webhookP}, false)
	db.AssertExists(t, "webhooks", map[string]interface{}{"id": f.webhookC}, false)
}

func TestProjectAccessLossCleanup(t *testing.T) {
	t.Run("removing a direct share cleans up the subtree", func(t *testing.T) {
		f := setupAccessLoss(t)
		f.shareDirectly(t, PermissionWrite)

		require.NoError(t, (&ProjectUser{ProjectID: f.parent.ID, Username: f.member.Username}).Delete(f.s, f.owner))
		f.assertCleanedUp(t)
	})
	t.Run("removing the team from the project", func(t *testing.T) {
		f := setupAccessLoss(t)
		team := f.shareViaTeam(t)

		require.NoError(t, (&TeamProject{TeamID: team.ID, ProjectID: f.parent.ID}).Delete(f.s, f.owner))
		f.assertCleanedUp(t)
	})
	t.Run("removing the team from the project spares members with a direct share", func(t *testing.T) {
		f := setupAccessLoss(t)
		team := f.shareViaTeam(t)
		other, othersWebhook := f.addOtherMemberWithWebhook(t, team)
		require.NoError(t, (&ProjectUser{ProjectID: f.parent.ID, Username: other.Username, Permission: PermissionRead}).Create(f.s, f.owner))

		require.NoError(t, (&TeamProject{TeamID: team.ID, ProjectID: f.parent.ID}).Delete(f.s, f.owner))
		f.assertCleanedUp(t)
		db.AssertExists(t, "webhooks", map[string]interface{}{"id": othersWebhook}, false)
	})
	t.Run("removing the member from the team", func(t *testing.T) {
		f := setupAccessLoss(t)
		team := f.shareViaTeam(t)

		require.NoError(t, (&TeamMember{TeamID: team.ID, Username: f.member.Username}).Delete(f.s, f.owner))
		f.assertCleanedUp(t)
	})
	t.Run("deleting the team", func(t *testing.T) {
		f := setupAccessLoss(t)
		team := f.shareViaTeam(t)

		require.NoError(t, team.Delete(f.s, f.owner))
		f.assertCleanedUp(t)
	})
	t.Run("removing the member through external team sync", func(t *testing.T) {
		f := setupAccessLoss(t)
		team := f.shareViaTeam(t)

		require.NoError(t, removeUserFromTeamsByIDs(f.s, f.member, []int64{team.ID}))
		require.NoError(t, f.s.Commit())
		for _, id := range []int64{f.webhookP, f.webhookC} {
			db.AssertExists(t, "webhooks", map[string]interface{}{"id": id, "created_by_id": f.member.ID, "target_url": "https://example.com/access-loss", "is_disabled": true}, false)
		}
		db.AssertMissing(t, "task_assignees", map[string]interface{}{"task_id": f.task.ID, "user_id": f.member.ID})
		db.AssertMissing(t, "subscriptions", map[string]interface{}{"entity_type": SubscriptionEntityTask, "entity_id": f.task.ID, "user_id": f.member.ID})
		db.AssertMissing(t, "subscriptions", map[string]interface{}{"entity_type": SubscriptionEntityProject, "entity_id": f.parent.ID, "user_id": f.member.ID})
		db.AssertExists(t, "link_shares", map[string]interface{}{"id": f.shareP}, false)
		db.AssertExists(t, "link_shares", map[string]interface{}{"id": f.shareC}, false)
	})
	t.Run("external team sync leaves teams the user was never in alone", func(t *testing.T) {
		f := setupAccessLoss(t)
		team := &Team{
			Name:        "access loss issuer team",
			CreatedByID: f.owner.ID,
			ExternalID:  "access-loss",
			Issuer:      "https://access-loss.issuer",
		}
		_, err := f.s.Insert(team)
		require.NoError(t, err)
		_, err = f.s.Insert(&TeamProject{
			TeamID:     team.ID,
			ProjectID:  f.parent.ID,
			Permission: PermissionWrite,
		})
		require.NoError(t, err)

		require.NoError(t, SyncExternalTeamsForUser(f.s, f.member, nil, team.Issuer, ""))
		f.assertWebhooksKept(t)
	})
	t.Run("reassigning the project away from its owner", func(t *testing.T) {
		f := setupAccessLoss(t)
		f.giveParentTo(t, f.member)

		_, err := ReassignProjectOwner(f.s, f.owner, f.parent.ID, f.owner.ID)
		require.NoError(t, err)
		f.assertCleanedUp(t)
	})
	t.Run("keeps webhooks of a reassigned owner who still has a share", func(t *testing.T) {
		f := setupAccessLoss(t)
		f.shareDirectly(t, PermissionRead)
		f.giveParentTo(t, f.member)

		_, err := ReassignProjectOwner(f.s, f.owner, f.parent.ID, f.owner.ID)
		require.NoError(t, err)
		f.assertWebhooksKept(t)
	})
	t.Run("keeps webhooks while a team still grants access", func(t *testing.T) {
		f := setupAccessLoss(t)
		f.shareDirectly(t, PermissionWrite)
		f.shareViaTeam(t)

		require.NoError(t, (&ProjectUser{ProjectID: f.parent.ID, Username: f.member.Username}).Delete(f.s, f.owner))
		f.assertWebhooksKept(t)
	})
	t.Run("keeps webhooks while a direct share still grants access", func(t *testing.T) {
		f := setupAccessLoss(t)
		f.shareDirectly(t, PermissionRead)
		team := f.shareViaTeam(t)

		require.NoError(t, (&TeamMember{TeamID: team.ID, Username: f.member.Username}).Delete(f.s, f.owner))
		f.assertWebhooksKept(t)
	})
	t.Run("keeps webhooks on a downgrade to read", func(t *testing.T) {
		f := setupAccessLoss(t)
		f.shareDirectly(t, PermissionWrite)

		require.NoError(t, (&ProjectUser{ProjectID: f.parent.ID, Username: f.member.Username, Permission: PermissionRead}).Update(f.s, f.owner))
		f.assertWebhooksKept(t)
	})
	t.Run("removes orphaned data for deleted project", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		const orphanProjectID int64 = 54321

		s := db.NewSession()
		defer s.Close()

		_, err := s.Insert(&TeamProject{TeamID: 9, ProjectID: orphanProjectID, Permission: PermissionRead})
		require.NoError(t, err)

		task := &Task{
			Title:       "orphan cleanup",
			ProjectID:   orphanProjectID,
			CreatedByID: 7,
			Index:       5,
		}
		_, err = s.Insert(task)
		require.NoError(t, err)

		_, err = s.Insert(&TaskAssginee{TaskID: task.ID, UserID: 2})
		require.NoError(t, err)

		_, err = s.Insert(&Subscription{EntityType: SubscriptionEntityTask, EntityID: task.ID, UserID: 2})
		require.NoError(t, err)

		_, err = s.Insert(&Subscription{EntityType: SubscriptionEntityProject, EntityID: orphanProjectID, UserID: 2})
		require.NoError(t, err)

		_, err = s.Where("team_id = ? AND user_id = ?", 9, 2).Delete(&TeamMember{})
		require.NoError(t, err)

		_, err = s.Where("project_id = ? AND user_id = ?", orphanProjectID, 2).Delete(&ProjectUser{})
		require.NoError(t, err)

		require.NoError(t, s.Commit())

		projectIDs, err := teamProjectIDs(s, 9)
		require.NoError(t, err)
		err = cleanupAfterProjectAccessLoss(s, nil, 2, projectIDs, AccessLossCleanup{})
		require.NoError(t, err)

		db.AssertMissing(t, "task_assignees", map[string]interface{}{"task_id": task.ID, "user_id": 2})
		db.AssertMissing(t, "subscriptions", map[string]interface{}{"entity_type": SubscriptionEntityTask, "entity_id": task.ID, "user_id": 2})
		db.AssertMissing(t, "subscriptions", map[string]interface{}{"entity_type": SubscriptionEntityProject, "entity_id": orphanProjectID, "user_id": 2})
	})
	t.Run("deleting the user deletes their webhooks on foreign projects", func(t *testing.T) {
		f := setupAccessLoss(t)
		f.shareDirectly(t, PermissionWrite)

		u, err := user.GetUserByID(f.s, f.member.ID)
		require.NoError(t, err)
		require.NoError(t, DeleteUser(f.s, u))
		require.NoError(t, f.s.Commit())
		db.AssertMissing(t, "webhooks", map[string]interface{}{"id": f.webhookP})
		db.AssertMissing(t, "webhooks", map[string]interface{}{"id": f.webhookC})
	})
}

func deliveredWebhookIDs() map[int64]bool {
	delivered := map[int64]bool{}
	for _, e := range events.GetDispatchedEvents((&WebhookDeliveryEvent{}).Name()) {
		delivered[e.(*WebhookDeliveryEvent).WebhookID] = true
	}
	return delivered
}

func TestProjectAccessLossCleanupOptions(t *testing.T) {
	removeShare := func(t *testing.T, f *accessLossFixture, cleanup AccessLossCleanup) error {
		t.Helper()
		pu := &ProjectUser{
			ProjectID:         f.parent.ID,
			Username:          f.member.Username,
			AccessLossCleanup: cleanup,
		}
		return pu.Delete(f.s, f.owner)
	}

	t.Run("revokes link shares", func(t *testing.T) {
		f := setupAccessLoss(t)
		f.shareDirectly(t, PermissionWrite)

		require.NoError(t, removeShare(t, f, AccessLossCleanup{RevokeLinkShares: true}))
		require.NoError(t, f.s.Commit())
		f.assertLinkSharesRevoked(t)
		db.AssertMissing(t, "webhooks", map[string]interface{}{"id": f.webhookP})
	})
	t.Run("keeps webhooks by transferring them to the acting user", func(t *testing.T) {
		f := setupAccessLoss(t)
		f.shareDirectly(t, PermissionWrite)

		require.NoError(t, removeShare(t, f, AccessLossCleanup{KeepWebhooks: true}))
		require.NoError(t, f.s.Commit())
		f.assertWebhooksTransferredTo(t, f.owner)
		db.AssertMissing(t, "task_assignees", map[string]interface{}{"task_id": f.task.ID, "user_id": f.member.ID})
	})
	t.Run("transfers subtree webhooks to the doer despite different project owners", func(t *testing.T) {
		f := setupAccessLoss(t)
		f.shareDirectly(t, PermissionWrite)
		childOwner := &user.User{ID: 3}
		_, err := f.s.ID(f.child.ID).Cols("owner_id").Update(&Project{OwnerID: childOwner.ID})
		require.NoError(t, err)

		require.NoError(t, removeShare(t, f, AccessLossCleanup{KeepWebhooks: true}))
		require.NoError(t, f.s.Commit())
		db.AssertExists(t, "webhooks", map[string]interface{}{"id": f.webhookP, "created_by_id": f.owner.ID}, false)
		db.AssertExists(t, "webhooks", map[string]interface{}{"id": f.webhookC, "created_by_id": f.owner.ID}, false)
	})
	t.Run("team removal applies the options to every member", func(t *testing.T) {
		f := setupAccessLoss(t)
		team := f.shareViaTeam(t)

		tp := &TeamProject{
			TeamID:    team.ID,
			ProjectID: f.parent.ID,
			AccessLossCleanup: AccessLossCleanup{
				KeepWebhooks:     true,
				RevokeLinkShares: true,
			},
		}
		require.NoError(t, tp.Delete(f.s, f.owner))
		require.NoError(t, f.s.Commit())
		f.assertWebhooksTransferredTo(t, f.owner)
		f.assertLinkSharesRevoked(t)
	})
	t.Run("member removal applies the options", func(t *testing.T) {
		f := setupAccessLoss(t)
		team := f.shareViaTeam(t)

		tm := &TeamMember{
			TeamID:            team.ID,
			Username:          f.member.Username,
			AccessLossCleanup: AccessLossCleanup{RevokeLinkShares: true},
		}
		require.NoError(t, tm.Delete(f.s, f.owner))
		require.NoError(t, f.s.Commit())
		f.assertLinkSharesRevoked(t)
	})
	t.Run("a member leaving their team cannot keep their webhooks", func(t *testing.T) {
		f := setupAccessLoss(t)
		team := f.shareViaTeam(t)

		tm := &TeamMember{
			TeamID:            team.ID,
			Username:          f.member.Username,
			AccessLossCleanup: AccessLossCleanup{KeepWebhooks: true},
		}
		require.NoError(t, tm.Delete(f.s, f.member))
		require.NoError(t, f.s.Commit())
		f.assertWebhooksDeleted(t)
	})
	t.Run("removing your own share cannot keep your webhooks", func(t *testing.T) {
		f := setupAccessLoss(t)
		f.shareDirectly(t, PermissionAdmin)

		pu := &ProjectUser{
			ProjectID:         f.parent.ID,
			Username:          f.member.Username,
			AccessLossCleanup: AccessLossCleanup{KeepWebhooks: true},
		}
		require.NoError(t, pu.Delete(f.s, f.member))
		require.NoError(t, f.s.Commit())
		f.assertWebhooksDeleted(t)
	})
	t.Run("team deletion cleans the doer's hooks before transferring other members' hooks", func(t *testing.T) {
		f := setupAccessLoss(t)
		team := f.shareViaTeam(t)
		doer, doersWebhook := f.addOtherMemberWithWebhook(t, team)
		_, err := f.s.Where("team_id = ? AND user_id = ?", team.ID, doer.ID).Cols("admin").Update(&TeamMember{Admin: true})
		require.NoError(t, err)
		canDelete, err := team.CanDelete(f.s, doer)
		require.NoError(t, err)
		require.True(t, canDelete)
		team.AccessLossCleanup = AccessLossCleanup{KeepWebhooks: true}

		require.NoError(t, team.Delete(f.s, doer))
		require.NoError(t, f.s.Commit())
		f.assertWebhooksTransferredTo(t, doer)
		db.AssertMissing(t, "webhooks", map[string]interface{}{"id": doersWebhook})

		events.ClearDispatchedEvents()
		events.TestListener(t, &TaskUpdatedEvent{
			Task: f.task,
			Doer: f.owner,
		}, &WebhookListener{EventName: "task.updated"})
		assert.Empty(t, events.GetDispatchedEvents((&WebhookDeliveryEvent{}).Name()))
	})
	t.Run("removing your own team from the project cannot keep your webhooks", func(t *testing.T) {
		f := setupAccessLoss(t)
		team := f.shareViaTeam(t)
		_, othersWebhook := f.addOtherMemberWithWebhook(t, team)

		tp := &TeamProject{
			TeamID:            team.ID,
			ProjectID:         f.parent.ID,
			AccessLossCleanup: AccessLossCleanup{KeepWebhooks: true},
		}
		require.NoError(t, tp.Delete(f.s, f.member))
		require.NoError(t, f.s.Commit())
		f.assertWebhooksDeleted(t)
		db.AssertExists(t, "webhooks", map[string]interface{}{"id": othersWebhook, "created_by_id": f.member.ID}, false)
	})
	t.Run("team deletion applies the options", func(t *testing.T) {
		f := setupAccessLoss(t)
		team := f.shareViaTeam(t)
		team.AccessLossCleanup = AccessLossCleanup{KeepWebhooks: true}

		require.NoError(t, team.Delete(f.s, f.owner))
		require.NoError(t, f.s.Commit())
		f.assertWebhooksTransferredTo(t, f.owner)
		db.AssertExists(t, "link_shares", map[string]interface{}{"id": f.shareP}, false)
		db.AssertExists(t, "link_shares", map[string]interface{}{"id": f.shareC}, false)
	})
	t.Run("team sync disables webhooks and can revoke link shares", func(t *testing.T) {
		f := setupAccessLoss(t)
		team := f.shareViaTeam(t)
		config.AuthTeamSyncRevokeLinkShares.Set(true)
		t.Cleanup(func() {
			config.AuthTeamSyncRevokeLinkShares.Set(false)
		})

		require.NoError(t, removeUserFromTeamsByIDs(f.s, f.member, []int64{team.ID}))
		require.NoError(t, f.s.Commit())
		for _, id := range []int64{f.webhookP, f.webhookC} {
			db.AssertExists(t, "webhooks", map[string]interface{}{"id": id, "created_by_id": f.member.ID, "is_disabled": true}, false)
		}
		f.assertLinkSharesRevoked(t)
	})
}

func TestProjectMoveAccessLossCleanup(t *testing.T) {
	t.Run("cleans up the moved subtree", func(t *testing.T) {
		f := setupAccessLoss(t)
		f.shareDirectly(t, PermissionWrite)

		require.NoError(t, f.moveChildToTopLevel(t, AccessLossCleanup{}, false))
		require.NoError(t, f.s.Commit())
		db.AssertMissing(t, "webhooks", map[string]interface{}{"id": f.webhookC})
		db.AssertMissing(t, "task_assignees", map[string]interface{}{"task_id": f.task.ID, "user_id": f.member.ID})
		db.AssertMissing(t, "subscriptions", map[string]interface{}{"entity_type": SubscriptionEntityTask, "entity_id": f.task.ID, "user_id": f.member.ID})
		db.AssertExists(t, "link_shares", map[string]interface{}{"id": f.shareC}, false)
		db.AssertExists(t, "webhooks", map[string]interface{}{"id": f.webhookP}, false)
		db.AssertExists(t, "subscriptions", map[string]interface{}{"entity_type": SubscriptionEntityProject, "entity_id": f.parent.ID, "user_id": f.member.ID}, false)
	})
	t.Run("applies the options", func(t *testing.T) {
		f := setupAccessLoss(t)
		f.shareDirectly(t, PermissionWrite)

		require.NoError(t, f.moveChildToTopLevel(t, AccessLossCleanup{
			KeepWebhooks:     true,
			RevokeLinkShares: true,
		}, false))
		require.NoError(t, f.s.Commit())
		db.AssertExists(t, "webhooks", map[string]interface{}{"id": f.webhookC, "created_by_id": f.owner.ID}, false)
		db.AssertMissing(t, "link_shares", map[string]interface{}{"id": f.shareC})
		db.AssertExists(t, "link_shares", map[string]interface{}{"id": f.shareP}, false)
	})
	t.Run("moving the project out of your inherited access cannot keep your webhooks", func(t *testing.T) {
		f := setupAccessLoss(t)
		f.shareDirectly(t, PermissionAdmin)
		other, othersWebhook := f.otherUserWithWebhook(t)
		pu := &ProjectUser{
			ProjectID:  f.parent.ID,
			Username:   other.Username,
			Permission: PermissionWrite,
		}
		require.NoError(t, pu.Create(f.s, f.owner))

		require.NoError(t, f.moveChildToTopLevelAs(t, f.member, AccessLossCleanup{KeepWebhooks: true}, false))
		require.NoError(t, f.s.Commit())
		db.AssertMissing(t, "webhooks", map[string]interface{}{"id": f.webhookC})
		db.AssertExists(t, "webhooks", map[string]interface{}{"id": f.webhookP, "created_by_id": f.member.ID}, false)
		db.AssertExists(t, "webhooks", map[string]interface{}{"id": othersWebhook, "created_by_id": f.member.ID}, false)
	})
	t.Run("a link share moving the project gets the default cleanup", func(t *testing.T) {
		f := setupAccessLoss(t)
		f.shareDirectly(t, PermissionWrite)
		adminShare := &LinkSharing{
			Hash:        "access-loss-admin-share",
			ProjectID:   f.child.ID,
			Permission:  PermissionAdmin,
			SharingType: SharingTypeWithoutPassword,
			SharedByID:  f.owner.ID,
		}
		_, err := f.s.Insert(adminShare)
		require.NoError(t, err)

		require.NoError(t, f.moveChildToTopLevelAs(t, adminShare, AccessLossCleanup{
			KeepWebhooks:     true,
			RevokeLinkShares: true,
		}, true))
		require.NoError(t, f.s.Commit())
		db.AssertExists(t, "link_shares", map[string]interface{}{"id": f.shareC}, false)
		db.AssertMissing(t, "webhooks", map[string]interface{}{"id": f.webhookC})
	})
	t.Run("covers grandchildren", func(t *testing.T) {
		f := setupAccessLoss(t)
		f.shareDirectly(t, PermissionWrite)
		childID := f.child.ID
		grandchild := &Project{
			Title:           "access loss grandchild",
			ParentProjectID: &childID,
		}
		require.NoError(t, grandchild.Create(f.s, f.owner))
		share := &LinkSharing{
			Hash:        "access-loss-grandchild",
			ProjectID:   grandchild.ID,
			SharingType: SharingTypeWithoutPassword,
			SharedByID:  f.member.ID,
		}
		_, err := f.s.Insert(share)
		require.NoError(t, err)

		require.NoError(t, f.moveChildToTopLevel(t, AccessLossCleanup{RevokeLinkShares: true}, false))
		require.NoError(t, f.s.Commit())
		db.AssertMissing(t, "link_shares", map[string]interface{}{"id": share.ID})
	})
	t.Run("ask fails with the affected counts", func(t *testing.T) {
		f := setupAccessLoss(t)
		f.shareDirectly(t, PermissionWrite)

		err := f.moveChildToTopLevel(t, AccessLossCleanup{}, true)
		assert.Equal(t, &ErrAccessLossNeedsConfirmation{
			Webhooks:   1,
			LinkShares: 1,
		}, err)
	})
	t.Run("ask passes when everyone keeps access", func(t *testing.T) {
		f := setupAccessLoss(t)
		f.shareDirectly(t, PermissionWrite)
		pu := &ProjectUser{
			ProjectID:  f.child.ID,
			Username:   f.member.Username,
			Permission: PermissionRead,
		}
		require.NoError(t, pu.Create(f.s, f.owner))

		require.NoError(t, f.moveChildToTopLevel(t, AccessLossCleanup{}, true))
		require.NoError(t, f.s.Commit())
		db.AssertExists(t, "webhooks", map[string]interface{}{"id": f.webhookC, "created_by_id": f.member.ID}, false)
		db.AssertExists(t, "link_shares", map[string]interface{}{"id": f.shareC, "shared_by_id": f.member.ID}, false)
		db.AssertExists(t, "projects", map[string]interface{}{"id": f.child.ID, "parent_project_id": nil}, false)
	})
	t.Run("ask ignores link shares of deleted users", func(t *testing.T) {
		f := setupAccessLoss(t)
		f.shareDirectly(t, PermissionWrite)
		pu := &ProjectUser{
			ProjectID:  f.child.ID,
			Username:   f.member.Username,
			Permission: PermissionRead,
		}
		require.NoError(t, pu.Create(f.s, f.owner))
		share := &LinkSharing{
			Hash:        "access-loss-deleted-sharer",
			ProjectID:   f.child.ID,
			SharingType: SharingTypeWithoutPassword,
			SharedByID:  9999,
		}
		_, err := f.s.Insert(share)
		require.NoError(t, err)

		require.NoError(t, f.moveChildToTopLevel(t, AccessLossCleanup{}, true))
		require.NoError(t, f.s.Commit())
		db.AssertExists(t, "link_shares", map[string]interface{}{"id": share.ID}, false)
	})
	t.Run("a plain edit does not clean up", func(t *testing.T) {
		f := setupAccessLoss(t)
		f.shareDirectly(t, PermissionWrite)
		_, err := f.s.Where("project_id = ? AND user_id = ?", f.parent.ID, f.member.ID).Delete(&ProjectUser{})
		require.NoError(t, err)

		child := &Project{
			ID:                  f.child.ID,
			Title:               "renamed",
			AskBeforeAccessLoss: true,
		}
		require.NoError(t, child.Update(f.s, f.owner))
		require.NoError(t, f.s.Commit())
		db.AssertExists(t, "webhooks", map[string]interface{}{"id": f.webhookC}, false)
	})
}

func TestWebhookListener_DeliversTransferredWebhook(t *testing.T) {
	f := setupAccessLoss(t)
	f.shareDirectly(t, PermissionWrite)
	doer := &user.User{
		ID:       3,
		Username: "user3",
	}
	doerShare := &ProjectUser{
		ProjectID:  f.parent.ID,
		Username:   doer.Username,
		Permission: PermissionAdmin,
	}
	require.NoError(t, doerShare.Create(f.s, f.owner))
	pu := &ProjectUser{
		ProjectID:         f.parent.ID,
		Username:          f.member.Username,
		AccessLossCleanup: AccessLossCleanup{KeepWebhooks: true},
	}
	require.NoError(t, pu.Delete(f.s, doer))
	require.NoError(t, f.s.Commit())
	f.assertWebhooksTransferredTo(t, doer)

	events.ClearDispatchedEvents()
	events.TestListener(t, &TaskUpdatedEvent{
		Task: f.task,
		Doer: f.owner,
	}, &WebhookListener{EventName: "task.updated"})

	delivered := map[int64]bool{}
	for _, e := range events.GetDispatchedEvents((&WebhookDeliveryEvent{}).Name()) {
		delivered[e.(*WebhookDeliveryEvent).WebhookID] = true
	}
	assert.True(t, delivered[f.webhookC], "the kept webhook should still be delivered")

	require.NoError(t, doerShare.Delete(f.s, f.owner))
	require.NoError(t, f.s.Commit())
	f.assertWebhooksDeleted(t)
	events.ClearDispatchedEvents()
	events.TestListener(t, &TaskUpdatedEvent{
		Task: f.task,
		Doer: f.owner,
	}, &WebhookListener{EventName: "task.updated"})
	assert.Empty(t, events.GetDispatchedEvents((&WebhookDeliveryEvent{}).Name()))
}

func TestWebhookListener_SkipsCreatorWithoutAccess(t *testing.T) {
	f := setupAccessLoss(t)
	f.shareDirectly(t, PermissionWrite)

	_, err := f.s.Insert(&Webhook{
		TargetURL:   "https://example.com/owner",
		Events:      []string{"task.updated"},
		ProjectID:   f.child.ID,
		CreatedByID: f.owner.ID,
	})
	require.NoError(t, err)
	// Drops access while skipping the cleanup.
	_, err = f.s.Where("project_id = ? AND user_id = ?", f.parent.ID, f.member.ID).Delete(&ProjectUser{})
	require.NoError(t, err)
	require.NoError(t, f.s.Commit())

	events.ClearDispatchedEvents()
	events.TestListener(t, &TaskUpdatedEvent{Task: f.task, Doer: f.owner}, &WebhookListener{EventName: "task.updated"})

	delivered := deliveredWebhookIDs()
	assert.False(t, delivered[f.webhookC], "webhook of a creator without access was delivered")
	assert.Len(t, delivered, 1, "the owner's webhook should still be delivered")
}

func TestWebhookListener_SkipsCreatorWithoutAccessToTrashedProject(t *testing.T) {
	f := setupAccessLoss(t)
	f.shareDirectly(t, PermissionWrite)

	owners := &Webhook{
		TargetURL:   "https://example.com/owner",
		Events:      []string{"project.deleted"},
		ProjectID:   f.child.ID,
		CreatedByID: f.owner.ID,
	}
	_, err := f.s.Insert(owners)
	require.NoError(t, err)
	_, err = f.s.ID(f.webhookC).Cols("events").Update(&Webhook{Events: []string{"project.deleted"}})
	require.NoError(t, err)
	// A move drops access without triggering the cleanup.
	detached := int64(0)
	f.child.ParentProjectID = &detached
	require.NoError(t, f.child.Update(f.s, f.owner))
	require.NoError(t, (&Project{ID: f.child.ID}).Delete(f.s, f.owner))
	require.NoError(t, f.s.Commit())

	events.ClearDispatchedEvents()
	events.TestListener(
		t,
		&ProjectDeletedEvent{
			Project: f.child,
			Doer:    f.owner,
		},
		&WebhookListener{EventName: "project.deleted"},
	)

	delivered := deliveredWebhookIDs()
	assert.False(t, delivered[f.webhookC], "webhook of a creator without access was delivered")
	assert.True(t, delivered[owners.ID], "the owner's webhook should still be delivered")
}

func TestWebhookListener_TrashedRootReceivesChildDeleted(t *testing.T) {
	f := setupAccessLoss(t)

	onRoot := &Webhook{
		TargetURL:   "https://example.com/root",
		Events:      []string{"project.deleted"},
		ProjectID:   f.parent.ID,
		CreatedByID: f.owner.ID,
	}
	_, err := f.s.Insert(onRoot)
	require.NoError(t, err)
	require.NoError(t, (&Project{ID: f.parent.ID}).Delete(f.s, f.owner))
	require.NoError(t, f.s.Commit())

	events.ClearDispatchedEvents()
	events.TestListener(
		t,
		&ProjectDeletedEvent{
			Project: f.child,
			Doer:    f.owner,
		},
		&WebhookListener{EventName: "project.deleted"},
	)

	delivered := deliveredWebhookIDs()
	assert.True(t, delivered[onRoot.ID], "the trashed root's webhook missed the child's project.deleted")
}

func TestWebhookListener_SkipsInactiveCreator(t *testing.T) {
	for _, tc := range []struct {
		status    user.Status
		delivered bool
	}{
		{status: user.StatusActive, delivered: true},
		{status: user.StatusDisabled, delivered: false},
		{status: user.StatusAccountLocked, delivered: false},
	} {
		t.Run(tc.status.String(), func(t *testing.T) {
			f := setupAccessLoss(t)
			f.shareDirectly(t, PermissionWrite)
			owners := &Webhook{
				TargetURL:   "https://example.com/owner",
				Events:      []string{"task.updated"},
				ProjectID:   f.child.ID,
				CreatedByID: f.owner.ID,
			}
			_, err := f.s.Insert(owners)
			require.NoError(t, err)
			require.NoError(t, user.SetUserStatus(f.s, f.member, tc.status))
			require.NoError(t, f.s.Commit())

			events.ClearDispatchedEvents()
			events.TestListener(t, &TaskUpdatedEvent{Task: f.task, Doer: f.owner}, &WebhookListener{EventName: "task.updated"})

			delivered := deliveredWebhookIDs()
			assert.True(t, delivered[owners.ID], "the active owner's webhook should still be delivered")
			assert.Equal(t, tc.delivered, delivered[f.webhookP])
			assert.Equal(t, tc.delivered, delivered[f.webhookC])
		})
	}
}
