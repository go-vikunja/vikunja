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
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/notifications"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

// Fixtures: project 3 is shared with user1 (directly and via team 1), project 2
// belongs to user3 alone, project 1 to user1.
func TestDatabaseNotifications_ReadAll_ProjectPermissions(t *testing.T) {
	t.Run("keeps a notification about a project the user can still read", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		insertStoredNotification(t, s, 1, taskCommentNotificationForProject(3))

		assert.Contains(t, readNotificationNames(t, s, 1), "task.comment")
	})

	t.Run("hides the same notification once project access is revoked", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		insertStoredNotification(t, s, 1, taskCommentNotificationForProject(3))
		require.Contains(t, readNotificationNames(t, s, 1), "task.comment",
			"the notification must be visible before access is revoked, otherwise the check below proves nothing")

		unshareProject(t, s, 3, 1)

		assert.NotContains(t, readNotificationNames(t, s, 1), "task.comment")
	})

	t.Run("hides a notification about a project the user never had access to", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		insertStoredNotification(t, s, 1, taskCommentNotificationForProject(2))

		assert.NotContains(t, readNotificationNames(t, s, 1), "task.comment")
	})

	t.Run("keeps account-level notifications regardless of project access", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		insertStoredNotification(t, s, 1, &TeamMemberAddedNotification{
			Member: &user.User{ID: 1},
			Doer:   &user.User{ID: 3},
			Team:   &Team{ID: 1, Name: "Testteam1"},
		})
		insertStoredNotification(t, s, 1, &APITokenExpiringWeekNotification{
			User:  &user.User{ID: 1},
			Token: &APIToken{ID: 1, Title: "test"},
		})
		insertStoredNotification(t, s, 1, taskCommentNotificationForProject(2))

		names := readNotificationNames(t, s, 1)
		assert.Contains(t, names, "team.member.added")
		assert.Contains(t, names, "api_token.expiring.week")
		assert.NotContains(t, names, "task.comment",
			"only the project-scoped notification may be filtered out")
	})

	t.Run("takes the project from the task when the payload has none", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Task 1 lives in project 1, which user1 owns and user2 cannot read.
		deleted := &TaskDeletedNotification{Doer: &user.User{ID: 1}, Task: &Task{ID: 1, ProjectID: 1}}
		require.EqualValues(t, 1, notifications.ProjectIDOf(deleted))
		insertStoredNotification(t, s, 1, deleted)
		insertStoredNotification(t, s, 2, deleted)

		assert.Contains(t, readNotificationNames(t, s, 1), "task.deleted")
		assert.NotContains(t, readNotificationNames(t, s, 2), "task.deleted")
	})

	t.Run("keeps notifications whose type is not registered", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Fixtures #1 and #2 belong to user1 and carry the unregistered name "test.notification".
		names := readNotificationNames(t, s, 1)
		assert.Equal(t, []string{"test.notification", "test.notification"}, names)
	})

	t.Run("hides a notification whose project could not be resolved, from everyone", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		orphan := &TaskCommentNotification{
			Doer:    &user.User{ID: 3},
			Task:    &Task{},
			Comment: &TaskComment{ID: 1, Comment: "a secret comment"},
		}
		require.Equal(t, notifications.ProjectIDUnresolved, notifications.ProjectIDOf(orphan))

		insertStoredNotification(t, s, 1, orphan)
		insertStoredNotification(t, s, 3, orphan)

		assert.NotContains(t, readNotificationNames(t, s, 1), "task.comment")
		assert.NotContains(t, readNotificationNames(t, s, 3), "task.comment",
			"not even the owner of every project may read a row that cannot be scoped")
	})

	t.Run("pages over the filtered set, so a page comes back full", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Hidden rows are inserted last so they sort first: filtering after the
		// query would leave the first page empty.
		for i := 0; i < 5; i++ {
			insertStoredNotification(t, s, 1, taskCommentNotificationForProject(3))
		}
		for i := 0; i < 10; i++ {
			insertStoredNotification(t, s, 1, taskCommentNotificationForProject(2))
		}

		result, resultCount, total, err := (&DatabaseNotifications{}).ReadAll(s, &user.User{ID: 1}, "", 1, 5)
		require.NoError(t, err)
		assert.Len(t, result, 5, "the first page must be full, not emptied by the filter")
		assert.Equal(t, 5, resultCount)
		assert.EqualValues(t, 7, total, "total must count only the notifications the user can see")

		result, resultCount, total, err = (&DatabaseNotifications{}).ReadAll(s, &user.User{ID: 1}, "", 2, 5)
		require.NoError(t, err)
		assert.Len(t, result, 2, "the last page holds the remaining visible notifications")
		assert.Equal(t, 2, resultCount)
		assert.EqualValues(t, 7, total)
	})
}

// The compiler forces ProjectID to exist; nothing forces a project-scoped type
// to actually return a project.
func TestNotificationScopeClassification(t *testing.T) {
	projectScoped := []string{
		"project.created",
		"task.assigned",
		"task.comment",
		"task.deleted",
		"task.mentioned",
		"task.reminder",
	}
	accountScoped := []string{
		"api_token.expiring.day",
		"api_token.expiring.week",
		"team.member.added",
	}

	registered := notifications.RegisteredNames()
	require.Len(t, registered, len(projectScoped)+len(accountScoped),
		"a notification type was registered or removed without updating the lists in this test")

	for _, name := range registered {
		t.Run(name, func(t *testing.T) {
			n, ok := notifications.Lookup(name)
			require.True(t, ok)

			inProject := slices.Contains(projectScoped, name)
			inAccount := slices.Contains(accountScoped, name)
			require.NotEqual(t, inProject, inAccount,
				"%q must be listed in exactly one of projectScoped and accountScoped in this test", name)

			// A fresh instance names no project, so a project-scoped type must fail closed.
			want := int64(0)
			if inProject {
				want = notifications.ProjectIDUnresolved
			}
			assert.Equal(t, want, notifications.ProjectIDOf(n),
				"%q must report 0 if and only if it is account-scoped", name)
		})
	}
}

func TestDatabaseNotifications_CanUpdate_ProjectPermissions(t *testing.T) {
	t.Run("allows marking a readable notification as read", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		dbn := insertStoredNotification(t, s, 1, taskCommentNotificationForProject(3))

		can, err := (&DatabaseNotifications{DatabaseNotification: notifications.DatabaseNotification{ID: dbn.ID}}).
			CanUpdate(s, &user.User{ID: 1})
		require.NoError(t, err)
		assert.True(t, can)
	})

	t.Run("refuses once project access is revoked", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		dbn := insertStoredNotification(t, s, 1, taskCommentNotificationForProject(3))
		unshareProject(t, s, 3, 1)

		can, err := (&DatabaseNotifications{DatabaseNotification: notifications.DatabaseNotification{ID: dbn.ID}}).
			CanUpdate(s, &user.User{ID: 1})
		require.NoError(t, err)
		assert.False(t, can)
	})
}

func taskCommentNotificationForProject(projectID int64) *TaskCommentNotification {
	return &TaskCommentNotification{
		Doer:    &user.User{ID: 3},
		Task:    &Task{ID: 32, ProjectID: projectID},
		Comment: &TaskComment{ID: 1, Comment: "a secret comment"},
		Project: &Project{ID: projectID, Title: "a secret project"},
	}
}

func unshareProject(t *testing.T, s *xorm.Session, projectID, userID int64) {
	t.Helper()
	_, err := s.Where("project_id = ? AND user_id = ?", projectID, userID).Delete(&ProjectUser{})
	require.NoError(t, err)
	_, err = s.Where("project_id = ?", projectID).Delete(&TeamProject{})
	require.NoError(t, err)
}

func readNotificationNames(t *testing.T, s *xorm.Session, notifiableID int64) []string {
	t.Helper()
	result, resultCount, _, err := (&DatabaseNotifications{}).ReadAll(s, &user.User{ID: notifiableID}, "", 1, 50)
	require.NoError(t, err)

	ns, ok := result.([]*notifications.DatabaseNotification)
	require.True(t, ok, "ReadAll returned unexpected type %T", result)
	require.Equal(t, len(ns), resultCount, "filtering happens in the query, so the count must be exact")

	names := make([]string, 0, len(ns))
	for _, dbn := range ns {
		names = append(names, dbn.Name)
	}
	return names
}
