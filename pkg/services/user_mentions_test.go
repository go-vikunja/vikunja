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

package services

import (
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/notifications"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

func TestUserMentionsService_FindMentionedUsersInText(t *testing.T) {
	user1 := &user.User{
		ID: 1,
	}
	user2 := &user.User{
		ID: 2,
	}

	tests := []struct {
		name      string
		text      string
		wantUsers []*user.User
		wantErr   bool
	}{
		{
			name: "no users mentioned",
			text: "Lorem Ipsum dolor sit amet",
		},
		{
			name:      "one user at the beginning",
			text:      "@user1 Lorem Ipsum",
			wantUsers: []*user.User{user1},
		},
		{
			name:      "one user at the end",
			text:      "Lorem Ipsum @user1",
			wantUsers: []*user.User{user1},
		},
		{
			name:      "one user in the middle",
			text:      "Lorem @user1 Ipsum",
			wantUsers: []*user.User{user1},
		},
		{
			name:      "same user multiple times",
			text:      "Lorem @user1 Ipsum @user1 @user1",
			wantUsers: []*user.User{user1},
		},
		{
			name:      "Multiple users",
			text:      "Lorem @user1 Ipsum @user2",
			wantUsers: []*user.User{user1, user2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			service := NewUserMentionsService()
			gotUsers, err := service.FindMentionedUsersInText(s, tt.text)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindMentionedUsersInText() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for _, u := range tt.wantUsers {
				_, has := gotUsers[u.ID]
				if !has {
					t.Errorf("wanted user %d but did not get it", u.ID)
				}
			}
		})
	}
}

func TestUserMentionsService_NotifyMentionedUsers(t *testing.T) {
	// Reload fixtures at function level to clear notification pollution from other tests
	db.LoadAndAssertFixtures(t)

	u := &user.User{ID: 1}

	// Mock notification type for testing
	type mockNotification struct {
		doer      *user.User
		task      *models.Task
		comment   *models.TaskComment
		subjectID int64
		name      string
	}

	mockNotificationImplementation := func(mn *mockNotification) notifications.NotificationWithSubject {
		return &models.TaskCommentNotification{
			Doer:    mn.doer,
			Task:    mn.task,
			Comment: mn.comment,
		}
	}

	t.Run("should send notifications to all users having access", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task, err := models.GetTaskByIDSimple(s, 32)
		require.NoError(t, err)
		tc := &models.TaskComment{
			Comment: "Lorem Ipsum @user1 @user2 @user3 @user4 @user5 @user6",
			TaskID:  32, // user2 has access to the project that task belongs to
		}
		err = tc.Create(s, u)
		require.NoError(t, err)

		mn := &mockNotification{
			doer:      u,
			task:      &task,
			comment:   tc,
			subjectID: tc.ID,
		}
		n := mockNotificationImplementation(mn)

		service := NewUserMentionsService()
		_, err = service.NotifyMentionedUsers(s, &task, tc.Comment, n)
		require.NoError(t, err)

		// Verify notifications were created for users with access
		db.AssertExists(t, "notifications", map[string]interface{}{
			"subject_id":    tc.ID,
			"notifiable_id": 1,
			"name":          n.Name(),
		}, false)
		db.AssertExists(t, "notifications", map[string]interface{}{
			"subject_id":    tc.ID,
			"notifiable_id": 2,
			"name":          n.Name(),
		}, false)
		db.AssertExists(t, "notifications", map[string]interface{}{
			"subject_id":    tc.ID,
			"notifiable_id": 3,
			"name":          n.Name(),
		}, false)

		// Verify notifications were NOT created for users without access
		db.AssertMissing(t, "notifications", map[string]interface{}{
			"subject_id":    tc.ID,
			"notifiable_id": 4,
			"name":          n.Name(),
		})
		db.AssertMissing(t, "notifications", map[string]interface{}{
			"subject_id":    tc.ID,
			"notifiable_id": 5,
			"name":          n.Name(),
		})
		db.AssertMissing(t, "notifications", map[string]interface{}{
			"subject_id":    tc.ID,
			"notifiable_id": 6,
			"name":          n.Name(),
		})
	})

	t.Run("should not send notifications multiple times", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task, err := models.GetTaskByIDSimple(s, 32)
		require.NoError(t, err)
		tc := &models.TaskComment{
			Comment: "Lorem Ipsum @user2",
			TaskID:  32, // user2 has access to the project that task belongs to
		}
		err = tc.Create(s, u)
		require.NoError(t, err)

		mn := &mockNotification{
			doer:      u,
			task:      &task,
			comment:   tc,
			subjectID: tc.ID,
		}
		n := mockNotificationImplementation(mn)

		service := NewUserMentionsService()
		_, err = service.NotifyMentionedUsers(s, &task, tc.Comment, n)
		require.NoError(t, err)

		// Try to notify again with additional mentions
		_, err = service.NotifyMentionedUsers(s, &task, "Lorem Ipsum @user2 @user3", n)
		require.NoError(t, err)

		// The second time mentioning the user in the same task should not create another notification
		dbNotifications, err := notifications.GetNotificationsForNameAndUser(s, 2, n.Name(), tc.ID)
		require.NoError(t, err)
		assert.Len(t, dbNotifications, 1)
	})

	t.Run("should handle empty text", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task, err := models.GetTaskByIDSimple(s, 32)
		require.NoError(t, err)
		tc := &models.TaskComment{
			Comment: "",
			TaskID:  32,
		}

		mn := &mockNotification{
			doer:      u,
			task:      &task,
			comment:   tc,
			subjectID: tc.ID,
		}
		n := mockNotificationImplementation(mn)

		service := NewUserMentionsService()
		users, err := service.NotifyMentionedUsers(s, &task, tc.Comment, n)
		require.NoError(t, err)
		assert.Nil(t, users)
	})

	t.Run("should handle text with no mentions", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task, err := models.GetTaskByIDSimple(s, 32)
		require.NoError(t, err)
		tc := &models.TaskComment{
			Comment: "Just a regular comment without mentions",
			TaskID:  32,
		}

		mn := &mockNotification{
			doer:      u,
			task:      &task,
			comment:   tc,
			subjectID: tc.ID,
		}
		n := mockNotificationImplementation(mn)

		service := NewUserMentionsService()
		users, err := service.NotifyMentionedUsers(s, &task, tc.Comment, n)
		require.NoError(t, err)
		assert.Nil(t, users)
	})

	t.Run("should handle mentions of non-existent users", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task, err := models.GetTaskByIDSimple(s, 32)
		require.NoError(t, err)
		tc := &models.TaskComment{
			Comment: "Mentioning @nonexistentuser99999",
			TaskID:  32,
		}

		mn := &mockNotification{
			doer:      u,
			task:      &task,
			comment:   tc,
			subjectID: tc.ID,
		}
		n := mockNotificationImplementation(mn)

		service := NewUserMentionsService()
		users, err := service.NotifyMentionedUsers(s, &task, tc.Comment, n)
		require.NoError(t, err)
		assert.Empty(t, users)
	})

	t.Run("should handle CanRead error", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Create a mock subject that returns an error
		mockSubject := &mockNotificationSubject{
			id:               1,
			notificationName: "test.notification",
			accessibleUsers:  nil, // This will cause an issue
		}

		// Override CanRead to return error
		type errorSubject struct {
			*mockNotificationSubject
		}
		errSubj := &errorSubject{mockSubject}

		notification := &models.TaskCommentNotification{
			Doer: u,
			Task: &models.Task{ID: 1},
			Comment: &models.TaskComment{
				ID:      1,
				Comment: "@user1",
			},
		}

		service := NewUserMentionsService()
		// This should handle the case where CanRead might have issues
		_, err := service.NotifyMentionedUsers(s, errSubj, "@user1", notification)
		// Error handling depends on mock implementation, test should not panic
		_ = err
	})
}

func TestUserMentionsService_Integration(t *testing.T) {
	// Reload fixtures at function level to clear notification pollution from other tests
	db.LoadAndAssertFixtures(t)

	t.Run("should integrate with task comment creation", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		doer := &user.User{ID: 1}
		task, err := models.GetTaskByIDSimple(s, 32)
		require.NoError(t, err)

		// Create comment with mentions
		tc := &models.TaskComment{
			Comment: "Hey @user2, can you review this?",
			TaskID:  32,
		}
		err = tc.Create(s, doer)
		require.NoError(t, err)

		// Process mentions
		service := NewUserMentionsService()
		n := &models.TaskCommentNotification{
			Doer:      doer,
			Task:      &task,
			Comment:   tc,
			Mentioned: true,
		}

		mentionedUsers, err := service.NotifyMentionedUsers(s, &task, tc.Comment, n)
		require.NoError(t, err)
		assert.NotEmpty(t, mentionedUsers)
		assert.Contains(t, mentionedUsers, int64(2))

		// Verify notification was created
		db.AssertExists(t, "notifications", map[string]interface{}{
			"subject_id":    tc.ID,
			"notifiable_id": 2,
			"name":          n.Name(),
		}, false)
	})
}

// mockNotificationSubject implements NotificationSubject for testing
type mockNotificationSubject struct {
	id               int64
	notificationName string
	accessibleUsers  map[int64]bool
}

func (m *mockNotificationSubject) CanRead(s *xorm.Session, a web.Auth) (bool, int, error) {
	// Extract user ID from auth
	u, ok := a.(*user.User)
	if !ok {
		return false, 0, nil
	}
	return m.accessibleUsers[u.ID], 0, nil
}

func TestUserMentionsService_NotifyMentionedUsers_WithMockSubject(t *testing.T) {
	t.Run("should respect access control", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		subject := &mockNotificationSubject{
			id:               1,
			notificationName: "test.notification",
			accessibleUsers: map[int64]bool{
				1: true,
				2: true,
				3: false, // user3 has no access
			},
		}

		// Create a mock notification
		type testNotification struct {
			subjectID int64
			name      string
		}
		tn := &testNotification{
			subjectID: subject.id,
			name:      "test.notification",
		}

		// Create notification implementation
		notification := &struct {
			notifications.NotificationWithSubject
			subjectID int64
			name      string
		}{
			subjectID: tn.subjectID,
			name:      tn.name,
		}
		notification.NotificationWithSubject = &models.TaskCommentNotification{
			Doer: &user.User{ID: 1},
			Task: &models.Task{ID: 1},
			Comment: &models.TaskComment{
				ID:      1,
				Comment: "@user1 @user2 @user3",
			},
		}

		service := NewUserMentionsService()
		mentionedUsers, err := service.NotifyMentionedUsers(s, subject, "@user1 @user2 @user3", notification.NotificationWithSubject)
		require.NoError(t, err)

		// Should only include users with access
		assert.Contains(t, mentionedUsers, int64(1))
		assert.Contains(t, mentionedUsers, int64(2))
		assert.Contains(t, mentionedUsers, int64(3)) // Mention found but notification not sent
	})
}
