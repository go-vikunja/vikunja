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
	"bytes"
	"context"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/notifications"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

func TestDeleteUser(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		notifications.Fake()
		t.Cleanup(notifications.Unfake)

		u := &user.User{ID: 6}
		err := DeleteUser(s, u)

		require.NoError(t, err)
		require.NoError(t, s.Commit())
		db.AssertMissing(t, "users", map[string]interface{}{"id": u.ID})
		db.AssertMissing(t, "projects", map[string]interface{}{"id": 24}) // only user6 had access to this project
		db.AssertExists(t, "projects", map[string]interface{}{"id": 6}, false)
		db.AssertExists(t, "projects", map[string]interface{}{"id": 7}, false)
		db.AssertExists(t, "projects", map[string]interface{}{"id": 8}, false)
		db.AssertExists(t, "projects", map[string]interface{}{"id": 9}, false)
		db.AssertExists(t, "projects", map[string]interface{}{"id": 10}, false)
		db.AssertExists(t, "projects", map[string]interface{}{"id": 11}, false)
	})
	t.Run("user with no projects", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		notifications.Fake()
		t.Cleanup(notifications.Unfake)

		u := &user.User{ID: 4}
		err := DeleteUser(s, u)

		require.NoError(t, err)
		// No assertions for deleted projects since that user doesn't have any
	})
	t.Run("user with a default project", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		notifications.Fake()
		t.Cleanup(notifications.Unfake)

		u := &user.User{ID: 16}
		err := DeleteUser(s, u)

		require.NoError(t, err)
		require.NoError(t, s.Commit())
		db.AssertMissing(t, "users", map[string]interface{}{"id": u.ID})
		db.AssertMissing(t, "projects", map[string]interface{}{"id": 37}) // only user16 had access to this project, and it was their default
	})
	t.Run("disabled user", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		notifications.Fake()
		t.Cleanup(notifications.Unfake)

		u := &user.User{ID: 17}
		err := DeleteUser(s, u)

		require.NoError(t, err)
		require.NoError(t, s.Commit())
		db.AssertMissing(t, "users", map[string]interface{}{"id": u.ID})
	})
	t.Run("disabled user with task attachment they created", func(t *testing.T) {
		// Regression test: the cascade calls TaskAttachment.Delete -> ReadOne,
		// which loads the attachment's creator via user.GetUserByID. If the
		// creator is the disabled user being deleted, that lookup must not
		// surface ErrAccountDisabled out of the cascade.
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		notifications.Fake()
		t.Cleanup(notifications.Unfake)

		project := &Project{
			Title:   "disabled user project",
			OwnerID: 17,
		}
		insertTestProject(t, s, project)

		task := &Task{
			Title:       "disabled user task",
			ProjectID:   project.ID,
			CreatedByID: 17,
			Index:       1,
		}
		_, err := s.Insert(task)
		require.NoError(t, err)

		_, err = s.Insert(&TaskAttachment{
			TaskID:      task.ID,
			FileID:      1,
			CreatedByID: 17,
		})
		require.NoError(t, err)

		err = DeleteUser(s, &user.User{ID: 17})
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		db.AssertMissing(t, "users", map[string]interface{}{"id": 17})
		db.AssertMissing(t, "projects", map[string]interface{}{"id": project.ID})
		db.AssertMissing(t, "tasks", map[string]interface{}{"id": task.ID})
	})
	t.Run("cleans up task assignments and subscriptions", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		notifications.Fake()
		t.Cleanup(notifications.Unfake)

		task := &Task{
			Title:       "user cleanup",
			ProjectID:   19,
			CreatedByID: 7,
			Index:       4,
		}
		_, err := s.Insert(task)
		require.NoError(t, err)

		_, err = s.Insert(&TaskAssginee{TaskID: task.ID, UserID: 4})
		require.NoError(t, err)

		_, err = s.Insert(&Subscription{EntityType: SubscriptionEntityTask, EntityID: task.ID, UserID: 4})
		require.NoError(t, err)

		_, err = s.Insert(&Subscription{EntityType: SubscriptionEntityProject, EntityID: 19, UserID: 4})
		require.NoError(t, err)

		_, err = s.Insert(&TeamMember{TeamID: 9, UserID: 4})
		require.NoError(t, err)

		err = DeleteUser(s, &user.User{ID: 4})
		require.NoError(t, err)

		require.NoError(t, s.Commit())
		db.AssertMissing(t, "task_assignees", map[string]interface{}{"user_id": 4})
		db.AssertMissing(t, "subscriptions", map[string]interface{}{"user_id": 4})
		db.AssertMissing(t, "team_members", map[string]interface{}{"user_id": 4})
	})
}

// GHSA-4hv6-xc92-j86g
func TestDeleteUser_RevokesSessions(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()
	events.ClearDispatchedEvents()

	u, err := user.GetUserByID(s, 1)
	require.NoError(t, err)
	require.NoError(t, DeleteUser(s, u))
	assert.Zero(t, events.CountDispatchedEvents((&SessionsRevokedEvent{}).Name()))
	require.NoError(t, s.Commit())
	events.DispatchPending(context.Background(), s)

	db.AssertMissing(t, "sessions", map[string]interface{}{"user_id": 1})
	revoked := events.GetDispatchedEvents((&SessionsRevokedEvent{}).Name())
	require.Len(t, revoked, 1)
	require.Equal(t, &SessionsRevokedEvent{UserID: 1}, revoked[0])
}

func TestDeleteUser_RemovesReactions(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()
	notifications.Fake()
	t.Cleanup(notifications.Unfake)

	db.AssertExists(t, "reactions", map[string]interface{}{"user_id": 1}, false)

	u, err := user.GetUserByID(s, 1)
	require.NoError(t, err)
	require.NoError(t, DeleteUser(s, u))
	require.NoError(t, s.Commit())

	db.AssertMissing(t, "reactions", map[string]interface{}{"user_id": 1})
}

// GHSA-hq66-9qg8-jwqg
func TestDeleteUser_DoesNotPromoteSharesOnProjectsNotOwned(t *testing.T) {
	t.Run("user share", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		notifications.Fake()
		t.Cleanup(notifications.Unfake)

		project := &Project{Title: "owned by user1", OwnerID: 1}
		insertTestProject(t, s, project)
		_, err := s.Insert(&ProjectUser{ProjectID: project.ID, UserID: 4, Permission: PermissionRead})
		require.NoError(t, err)

		require.NoError(t, DeleteUser(s, &user.User{ID: 4}))
		require.NoError(t, s.Commit())

		db.AssertExists(t, "projects", map[string]interface{}{"id": project.ID, "owner_id": 1}, false)
		db.AssertMissing(t, "users_projects", map[string]interface{}{"user_id": 4})
	})
	t.Run("team share", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		notifications.Fake()
		t.Cleanup(notifications.Unfake)

		project := &Project{Title: "owned by user1", OwnerID: 1}
		insertTestProject(t, s, project)
		_, err := s.Insert(&TeamProject{ProjectID: project.ID, TeamID: 9, Permission: PermissionRead})
		require.NoError(t, err)
		_, err = s.Insert(&TeamMember{TeamID: 9, UserID: 4})
		require.NoError(t, err)

		require.NoError(t, DeleteUser(s, &user.User{ID: 4}))
		require.NoError(t, s.Commit())

		db.AssertExists(t, "projects", map[string]interface{}{"id": project.ID, "owner_id": 1}, false)
		db.AssertExists(t, "team_projects", map[string]interface{}{"project_id": project.ID, "team_id": 9, "permission": PermissionRead}, false)
	})
}

// GHSA-7gmh-4fw6-8xc5
func TestDeleteUser_RemovesExportAndAvatarFiles(t *testing.T) {
	setup := func(t *testing.T) (s *xorm.Session, avatar *files.File) {
		db.LoadAndAssertFixtures(t)
		files.InitTestFileFixtures(t)
		notifications.Fake()
		t.Cleanup(notifications.Unfake)

		s = db.NewSession()
		t.Cleanup(func() {
			events.CleanupPending(s)
			_ = s.Close()
		})
		avatar, err := files.CreateWithSession(s, bytes.NewReader([]byte("avatar")), "avatar.png", 6, &user.User{ID: 4})
		require.NoError(t, err)
		_, err = s.ID(4).Cols("export_file_id", "avatar_file_id").Update(&user.User{ExportFileID: 1, AvatarFileID: avatar.ID})
		require.NoError(t, err)
		return s, avatar
	}

	t.Run("removes both", func(t *testing.T) {
		s, avatar := setup(t)

		require.NoError(t, DeleteUser(s, &user.User{ID: 4}))
		assertBlobExists(t, 1)
		require.NoError(t, s.Commit())
		events.DispatchPending(context.Background(), s)

		db.AssertMissing(t, "files", map[string]interface{}{"id": 1})
		db.AssertMissing(t, "files", map[string]interface{}{"id": avatar.ID})
		assertBlobMissing(t, 1)
		assertBlobMissing(t, avatar.ID)
	})
	t.Run("keeps the shared default avatar", func(t *testing.T) {
		s, avatar := setup(t)
		config.DefaultSettingsAvatarFileID.Set(avatar.ID)
		t.Cleanup(func() { config.DefaultSettingsAvatarFileID.Set(0) })

		require.NoError(t, DeleteUser(s, &user.User{ID: 4}))
		require.NoError(t, s.Commit())
		events.DispatchPending(context.Background(), s)

		db.AssertExists(t, "files", map[string]interface{}{"id": avatar.ID}, false)
		_, err := files.FileStat(avatar)
		require.NoError(t, err)
	})
	t.Run("keeps files when the deletion rolls back", func(t *testing.T) {
		s, _ := setup(t)

		require.NoError(t, DeleteUser(s, &user.User{ID: 4}))
		require.NoError(t, s.Rollback())
		events.CleanupPending(s)

		db.AssertExists(t, "files", map[string]interface{}{"id": 1}, false)
		assertBlobExists(t, 1)
	})
}

func assertBlobExists(t *testing.T, fileID int64) {
	t.Helper()
	_, err := files.FileStat(&files.File{ID: fileID})
	require.NoError(t, err)
}
