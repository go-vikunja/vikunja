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
	"context"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The admin model actions dispatch their audit events on commit; these tests
// mirror the handler flow (action → commit → DispatchPending) and assert the
// event payload carries the acting admin and the affected user/project.

func adminActionsSetup(t *testing.T) {
	t.Helper()
	db.LoadAndAssertFixtures(t)
	events.ClearDispatchedEvents()
}

func singleDispatchedEvent[T events.Event](t *testing.T) T {
	t.Helper()
	var zero T
	dispatched := events.GetDispatchedEvents(zero.Name())
	require.Len(t, dispatched, 1)
	return dispatched[0].(T)
}

func TestSetUserAdminFlag(t *testing.T) {
	doer := &user.User{ID: 1}

	t.Run("granting dispatches admin.user.admin.granted", func(t *testing.T) {
		adminActionsSetup(t)
		s := db.NewSession()
		defer s.Close()

		target, err := SetUserAdminFlag(s, doer, 2, true)
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		events.DispatchPending(context.Background(), s)

		assert.True(t, target.IsAdmin)
		evt := singleDispatchedEvent[*AdminUserAdminGrantedEvent](t)
		assert.Equal(t, int64(1), evt.Doer.ID)
		assert.Equal(t, int64(2), evt.User.ID)
	})

	t.Run("revoking dispatches admin.user.admin.revoked", func(t *testing.T) {
		adminActionsSetup(t)
		s := db.NewSession()
		defer s.Close()

		// Two admins so the demotion passes the last-admin guard.
		_, err := s.Cols("is_admin").In("id", 2, 3).Update(&user.User{IsAdmin: true})
		require.NoError(t, err)

		target, err := SetUserAdminFlag(s, doer, 2, false)
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		events.DispatchPending(context.Background(), s)

		assert.False(t, target.IsAdmin)
		evt := singleDispatchedEvent[*AdminUserAdminRevokedEvent](t)
		assert.Equal(t, int64(1), evt.Doer.ID)
		assert.Equal(t, int64(2), evt.User.ID)
	})

	t.Run("refused demotion of the last admin dispatches nothing", func(t *testing.T) {
		adminActionsSetup(t)
		s := db.NewSession()
		defer s.Close()

		_, err := s.ID(2).Cols("is_admin").Update(&user.User{IsAdmin: true})
		require.NoError(t, err)

		_, err = SetUserAdminFlag(s, doer, 2, false)
		require.Error(t, err)
		events.DispatchPending(context.Background(), s)

		assert.Zero(t, events.CountDispatchedEvents((&AdminUserAdminRevokedEvent{}).Name()))
	})
}

func TestSetUserStatusAsAdmin_Events(t *testing.T) {
	doer := &user.User{ID: 1}

	t.Run("dispatches admin.user.status.changed with old and new status", func(t *testing.T) {
		adminActionsSetup(t)
		s := db.NewSession()
		defer s.Close()

		_, err := SetUserStatusAsAdmin(s, doer, 2, user.StatusDisabled)
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		events.DispatchPending(context.Background(), s)

		evt := singleDispatchedEvent[*AdminUserStatusChangedEvent](t)
		assert.Equal(t, int64(1), evt.Doer.ID)
		assert.Equal(t, int64(2), evt.User.ID)
		assert.Equal(t, user.StatusActive, evt.OldStatus)
		assert.Equal(t, user.StatusDisabled, evt.NewStatus)
	})

	t.Run("nonexistent user dispatches nothing", func(t *testing.T) {
		adminActionsSetup(t)
		s := db.NewSession()
		defer s.Close()

		_, err := SetUserStatusAsAdmin(s, doer, 99999, user.StatusDisabled)
		require.Error(t, err)
		events.DispatchPending(context.Background(), s)

		assert.Zero(t, events.CountDispatchedEvents((&AdminUserStatusChangedEvent{}).Name()))
	})
}

func TestSetUserPasswordAsAdmin_Events(t *testing.T) {
	doer := &user.User{ID: 1}

	t.Run("dispatches admin.user.password.set", func(t *testing.T) {
		adminActionsSetup(t)
		s := db.NewSession()
		defer s.Close()

		_, err := SetUserPasswordAsAdmin(s, doer, 2, "averyl0ngpassword")
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		events.DispatchPending(context.Background(), s)

		evt := singleDispatchedEvent[*AdminUserPasswordSetEvent](t)
		assert.Equal(t, int64(1), evt.Doer.ID)
		assert.Equal(t, int64(2), evt.User.ID)
	})

	t.Run("refused non-local account dispatches nothing", func(t *testing.T) {
		adminActionsSetup(t)
		s := db.NewSession()
		defer s.Close()

		// User 19 is an OpenID account.
		_, err := SetUserPasswordAsAdmin(s, doer, 19, "averyl0ngpassword")
		require.Error(t, err)
		events.DispatchPending(context.Background(), s)

		assert.Zero(t, events.CountDispatchedEvents((&AdminUserPasswordSetEvent{}).Name()))
	})
}

func TestRequestPasswordResetAsAdmin_Events(t *testing.T) {
	doer := &user.User{ID: 1}

	t.Run("dispatches admin.user.password_reset.sent", func(t *testing.T) {
		adminActionsSetup(t)
		s := db.NewSession()
		defer s.Close()

		require.NoError(t, RequestPasswordResetAsAdmin(s, doer, 2))
		require.NoError(t, s.Commit())
		events.DispatchPending(context.Background(), s)

		evt := singleDispatchedEvent[*AdminUserPasswordResetSentEvent](t)
		assert.Equal(t, int64(1), evt.Doer.ID)
		assert.Equal(t, int64(2), evt.User.ID)
	})

	t.Run("refused bot account dispatches nothing", func(t *testing.T) {
		adminActionsSetup(t)
		s := db.NewSession()
		defer s.Close()

		// User 23 is a bot.
		err := RequestPasswordResetAsAdmin(s, doer, 23)
		require.Error(t, err)
		events.DispatchPending(context.Background(), s)

		assert.Zero(t, events.CountDispatchedEvents((&AdminUserPasswordResetSentEvent{}).Name()))
	})
}

func TestDeleteUserAsAdmin_Events(t *testing.T) {
	doer := &user.User{ID: 1}

	t.Run("mode=now dispatches admin.user.deleted", func(t *testing.T) {
		adminActionsSetup(t)
		s := db.NewSession()
		defer s.Close()

		require.NoError(t, DeleteUserAsAdmin(s, doer, 15, "now"))
		require.NoError(t, s.Commit())
		events.DispatchPending(context.Background(), s)

		evt := singleDispatchedEvent[*AdminUserDeletedEvent](t)
		assert.Equal(t, int64(1), evt.Doer.ID)
		assert.Equal(t, int64(15), evt.User.ID)
		assert.Equal(t, "now", evt.Mode)
	})

	t.Run("mode=scheduled dispatches admin.user.deleted", func(t *testing.T) {
		adminActionsSetup(t)
		s := db.NewSession()
		defer s.Close()

		require.NoError(t, DeleteUserAsAdmin(s, doer, 16, "scheduled"))
		require.NoError(t, s.Commit())
		events.DispatchPending(context.Background(), s)

		evt := singleDispatchedEvent[*AdminUserDeletedEvent](t)
		assert.Equal(t, int64(16), evt.User.ID)
		assert.Equal(t, "scheduled", evt.Mode)
	})

	t.Run("refused deletion of the last admin dispatches nothing", func(t *testing.T) {
		adminActionsSetup(t)
		s := db.NewSession()
		defer s.Close()

		_, err := s.ID(2).Cols("is_admin").Update(&user.User{IsAdmin: true})
		require.NoError(t, err)

		err = DeleteUserAsAdmin(s, doer, 2, "now")
		require.Error(t, err)
		events.DispatchPending(context.Background(), s)

		assert.Zero(t, events.CountDispatchedEvents((&AdminUserDeletedEvent{}).Name()))
	})
}

func TestReassignProjectOwner_Events(t *testing.T) {
	doer := &user.User{ID: 1}

	t.Run("dispatches admin.project.owner.changed with both owner ids", func(t *testing.T) {
		adminActionsSetup(t)
		s := db.NewSession()
		defer s.Close()

		p, err := ReassignProjectOwner(s, doer, 1, 2)
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		events.DispatchPending(context.Background(), s)

		assert.Equal(t, int64(2), p.OwnerID)
		evt := singleDispatchedEvent[*AdminProjectOwnerChangedEvent](t)
		assert.Equal(t, int64(1), evt.Doer.ID)
		assert.Equal(t, int64(1), evt.Project.ID)
		assert.Equal(t, int64(1), evt.OldOwnerID)
		assert.Equal(t, int64(2), evt.NewOwnerID)
	})

	t.Run("refused owner scheduled for deletion dispatches nothing", func(t *testing.T) {
		adminActionsSetup(t)
		s := db.NewSession()
		defer s.Close()

		_, err := s.ID(2).Cols("deletion_scheduled_at").Update(&user.User{DeletionScheduledAt: time.Now()})
		require.NoError(t, err)

		_, err = ReassignProjectOwner(s, doer, 1, 2)
		require.Error(t, err)
		events.DispatchPending(context.Background(), s)

		assert.Zero(t, events.CountDispatchedEvents((&AdminProjectOwnerChangedEvent{}).Name()))
	})
}

func TestCreateUserAsAdmin_Events(t *testing.T) {
	doer := &user.User{ID: 1}

	t.Run("dispatches admin.user.created alongside user.created", func(t *testing.T) {
		adminActionsSetup(t)
		s := db.NewSession()
		defer s.Close()

		newUser, err := CreateUserAsAdmin(s, doer, &CreateUserBody{
			APIUserPassword: user.APIUserPassword{
				Username: "admin-created-1",
				Password: "averyl0ngpassword",
				Email:    "admin-created-1@example.com",
			},
		})
		require.NoError(t, err)
		events.DispatchPending(context.Background(), s)

		evt := singleDispatchedEvent[*AdminUserCreatedEvent](t)
		assert.Equal(t, int64(1), evt.Doer.ID)
		assert.Equal(t, newUser.ID, evt.User.ID)
		// The regular self-registration event stays untouched (actor = the new user).
		events.AssertDispatched(t, &user.CreatedEvent{})
	})

	t.Run("failed creation dispatches nothing", func(t *testing.T) {
		adminActionsSetup(t)
		s := db.NewSession()
		defer s.Close()

		// user1 already exists.
		_, err := CreateUserAsAdmin(s, doer, &CreateUserBody{
			APIUserPassword: user.APIUserPassword{
				Username: "user1",
				Password: "averyl0ngpassword",
				Email:    "duplicate@example.com",
			},
		})
		require.Error(t, err)
		events.CleanupPending(s)

		assert.Zero(t, events.CountDispatchedEvents((&AdminUserCreatedEvent{}).Name()))
	})
}
