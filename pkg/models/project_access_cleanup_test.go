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
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/user"

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

	insertWebhook := func(projectID int64) int64 {
		w := &Webhook{
			TargetURL:   "https://example.com/access-loss",
			Events:      []string{"task.updated"},
			ProjectID:   projectID,
			CreatedByID: f.member.ID,
		}
		_, err := s.Insert(w)
		require.NoError(t, err)
		return w.ID
	}
	f.webhookP = insertWebhook(f.parent.ID)
	f.webhookC = insertWebhook(f.child.ID)
	return f
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
		other := &user.User{ID: 3, Username: "user3"}
		require.NoError(t, (&TeamMember{TeamID: team.ID, Username: other.Username}).Create(f.s, f.owner))
		require.NoError(t, (&ProjectUser{ProjectID: f.parent.ID, Username: other.Username, Permission: PermissionRead}).Create(f.s, f.owner))
		othersWebhook := &Webhook{
			TargetURL:   "https://example.com/access-loss",
			Events:      []string{"task.updated"},
			ProjectID:   f.child.ID,
			CreatedByID: other.ID,
		}
		_, err := f.s.Insert(othersWebhook)
		require.NoError(t, err)

		require.NoError(t, (&TeamProject{TeamID: team.ID, ProjectID: f.parent.ID}).Delete(f.s, f.owner))
		f.assertCleanedUp(t)
		db.AssertExists(t, "webhooks", map[string]interface{}{"id": othersWebhook.ID}, false)
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
		f.assertCleanedUp(t)
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
		err = cleanupAfterProjectAccessLoss(s, 2, projectIDs)
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

func TestWebhookListener_SkipsCreatorWithoutAccess(t *testing.T) {
	f := setupAccessLoss(t)
	f.shareDirectly(t, PermissionWrite)

	// A move drops access without triggering the cleanup.
	_, err := f.s.Insert(&Webhook{
		TargetURL:   "https://example.com/owner",
		Events:      []string{"task.updated"},
		ProjectID:   f.child.ID,
		CreatedByID: f.owner.ID,
	})
	require.NoError(t, err)
	detached := int64(0)
	f.child.ParentProjectID = &detached
	require.NoError(t, f.child.Update(f.s, f.owner))
	require.NoError(t, f.s.Commit())

	events.ClearDispatchedEvents()
	events.TestListener(t, &TaskUpdatedEvent{Task: f.task, Doer: f.owner}, &WebhookListener{EventName: "task.updated"})

	delivered := map[int64]bool{}
	for _, e := range events.GetDispatchedEvents((&WebhookDeliveryEvent{}).Name()) {
		delivered[e.(*WebhookDeliveryEvent).WebhookID] = true
	}
	assert.False(t, delivered[f.webhookC], "webhook of a creator without access was delivered")
	assert.Len(t, delivered, 1, "the owner's webhook should still be delivered")
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

			delivered := map[int64]bool{}
			for _, e := range events.GetDispatchedEvents((&WebhookDeliveryEvent{}).Name()) {
				delivered[e.(*WebhookDeliveryEvent).WebhookID] = true
			}
			assert.True(t, delivered[owners.ID], "the active owner's webhook should still be delivered")
			assert.Equal(t, tc.delivered, delivered[f.webhookP])
			assert.Equal(t, tc.delivered, delivered[f.webhookC])
		})
	}
}
