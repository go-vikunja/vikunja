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
	"fmt"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/notifications"
	"code.vikunja.io/api/pkg/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotificationsService_GetNotificationsForUser(t *testing.T) {
	t.Run("Get notifications with pagination", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		service := NewNotificationsService(s)
		u := &user.User{ID: 1}

		// Create test notifications using Notify method
		for i := 0; i < 15; i++ {
			testNotif := &testNotification{name: fmt.Sprintf("test.pagination.%d", i)}
			err := service.Notify(u, testNotif)
			require.NoError(t, err)
		}
		_ = s.Commit()

		// Start new session to retrieve
		s = db.NewSession()
		defer s.Close()
		service = NewNotificationsService(s)

		// Get first page
		notifs, resultCount, total, err := service.GetNotificationsForUser(u.ID, 10, 0)
		require.NoError(t, err)
		assert.Greater(t, total, int64(0), "should have notifications")
		assert.Equal(t, resultCount, len(notifs))
		assert.LessOrEqual(t, resultCount, 10)
		assert.Equal(t, int64(15), total, "should have 15 total notifications")
	})

	t.Run("Get notifications with offset", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		service := NewNotificationsService(s)
		u := &user.User{ID: 1}

		// Create test notifications using Notify method
		for i := 0; i < 15; i++ {
			testNotif := &testNotification{name: fmt.Sprintf("test.offset.%d", i)}
			err := service.Notify(u, testNotif)
			require.NoError(t, err)
		}
		_ = s.Commit()

		// Start new session to retrieve
		s = db.NewSession()
		defer s.Close()
		service = NewNotificationsService(s)

		// Get second page
		notifs, resultCount, total, err := service.GetNotificationsForUser(u.ID, 5, 5)
		require.NoError(t, err)
		assert.Greater(t, total, int64(0))
		assert.Equal(t, resultCount, len(notifs))
		assert.Equal(t, 5, resultCount, "should have 5 notifications on second page")
	})

	t.Run("User with no notifications", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		service := NewNotificationsService(s)
		userID := int64(999999) // Non-existent user

		notifs, resultCount, total, err := service.GetNotificationsForUser(userID, 10, 0)
		require.NoError(t, err)
		assert.Equal(t, int64(0), total)
		assert.Equal(t, 0, resultCount)
		assert.Empty(t, notifs)
	})
}

func TestNotificationsService_GetNotificationsForNameAndUser(t *testing.T) {
	t.Run("Get notifications by event name", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		service := NewNotificationsService(s)

		// Create notifications with subject ID
		u := &user.User{ID: 1}
		subjectID := int64(123)
		testNotif := &testNotificationWithSubject{
			testNotification: testNotification{name: "test.event"},
			subjectID:        subjectID,
		}

		err := service.Notify(u, testNotif)
		require.NoError(t, err)

		_ = s.Commit()
		s = db.NewSession()
		defer s.Close()
		service = NewNotificationsService(s)

		// Now retrieve it by event name and subject ID
		notifs, err := service.GetNotificationsForNameAndUser(u.ID, "test.event", subjectID)
		require.NoError(t, err)
		assert.NotEmpty(t, notifs, "should have notifications for event and subject")
		assert.Equal(t, 1, len(notifs), "should have exactly 1 notification")
		assert.Equal(t, "test.event", notifs[0].Name)
		assert.Equal(t, subjectID, notifs[0].SubjectID)
	})

	t.Run("Event not found", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		service := NewNotificationsService(s)
		userID := int64(1)

		notifs, err := service.GetNotificationsForNameAndUser(userID, "nonexistent.event", 0)
		require.NoError(t, err)
		assert.Empty(t, notifs)
	})
}

func TestNotificationsService_CanMarkNotificationAsRead(t *testing.T) {
	t.Run("User can mark their own notification", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		service := NewNotificationsService(s)

		// Create a notification
		u := &user.User{ID: 1}
		testNotif := &testNotification{name: "test.mark"}

		err := service.Notify(u, testNotif)
		require.NoError(t, err)
		_ = s.Commit()

		// Retrieve the notification
		s = db.NewSession()
		defer s.Close()
		service = NewNotificationsService(s)
		notifs, _, _, err := service.GetNotificationsForUser(u.ID, 1, 0)
		require.NoError(t, err)
		require.NotEmpty(t, notifs)

		// Check if user can mark it
		notification := &notifications.DatabaseNotification{ID: notifs[0].ID}
		can, err := service.CanMarkNotificationAsRead(notification, u.ID)
		require.NoError(t, err)
		assert.True(t, can)
	})

	t.Run("User cannot mark another user's notification", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		service := NewNotificationsService(s)

		// Create a notification for user 1
		u := &user.User{ID: 1}
		testNotif := &testNotification{name: "test.other"}

		err := service.Notify(u, testNotif)
		require.NoError(t, err)
		_ = s.Commit()

		// Try to access as user 2
		s = db.NewSession()
		defer s.Close()
		service = NewNotificationsService(s)
		notifs, _, _, err := service.GetNotificationsForUser(u.ID, 1, 0)
		require.NoError(t, err)
		require.NotEmpty(t, notifs)

		notification := &notifications.DatabaseNotification{ID: notifs[0].ID}
		can, err := service.CanMarkNotificationAsRead(notification, 2)
		require.NoError(t, err)
		assert.False(t, can)
	})
}

func TestNotificationsService_MarkNotificationAsRead(t *testing.T) {
	t.Run("Mark notification as read", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		service := NewNotificationsService(s)

		// Create a notification
		u := &user.User{ID: 1}
		testNotif := &testNotification{name: "test.read"}

		err := service.Notify(u, testNotif)
		require.NoError(t, err)
		_ = s.Commit()

		// Get the notification
		s = db.NewSession()
		defer s.Close()
		service = NewNotificationsService(s)
		notifs, _, _, err := service.GetNotificationsForUser(u.ID, 1, 0)
		require.NoError(t, err)
		require.NotEmpty(t, notifs)

		// Mark as read
		notification := notifs[0]
		assert.True(t, notification.ReadAt.IsZero(), "should start unread")

		err = service.MarkNotificationAsRead(notification, true)
		require.NoError(t, err)
		_ = s.Commit()

		// Verify it's marked
		s = db.NewSession()
		defer s.Close()
		service = NewNotificationsService(s)
		updated := &notifications.DatabaseNotification{ID: notification.ID}
		exists, err := s.Get(updated)
		require.NoError(t, err)
		require.True(t, exists)
		assert.False(t, updated.ReadAt.IsZero(), "should be marked as read")
	})

	t.Run("Mark notification as unread", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		service := NewNotificationsService(s)

		// Create a notification
		u := &user.User{ID: 1}
		testNotif := &testNotification{name: "test.unread"}

		err := service.Notify(u, testNotif)
		require.NoError(t, err)
		_ = s.Commit()

		// Get and mark as read first
		s = db.NewSession()
		defer s.Close()
		service = NewNotificationsService(s)
		notifs, _, _, err := service.GetNotificationsForUser(u.ID, 1, 0)
		require.NoError(t, err)
		require.NotEmpty(t, notifs)

		notification := notifs[0]
		err = service.MarkNotificationAsRead(notification, true)
		require.NoError(t, err)
		_ = s.Commit()

		// Now mark as unread
		s = db.NewSession()
		defer s.Close()
		service = NewNotificationsService(s)
		notification.ReadAt = time.Now() // Simulate it being read
		err = service.MarkNotificationAsRead(notification, false)
		require.NoError(t, err)
		_ = s.Commit()

		// Verify it's unread
		s = db.NewSession()
		defer s.Close()
		updated := &notifications.DatabaseNotification{ID: notification.ID}
		exists, err := s.Get(updated)
		require.NoError(t, err)
		require.True(t, exists)
		assert.True(t, updated.ReadAt.IsZero(), "should be marked as unread")
	})
}

func TestNotificationsService_MarkAllNotificationsAsRead(t *testing.T) {
	t.Run("Mark all notifications as read", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		service := NewNotificationsService(s)

		// Create multiple notifications
		u := &user.User{ID: 1}
		for i := 0; i < 3; i++ {
			testNotif := &testNotification{name: "test.bulk"}
			err := service.Notify(u, testNotif)
			require.NoError(t, err)
		}
		_ = s.Commit()

		// Mark all as read
		s = db.NewSession()
		defer s.Close()
		service = NewNotificationsService(s)
		err := service.MarkAllNotificationsAsRead(u.ID)
		require.NoError(t, err)
		_ = s.Commit()

		// Verify all are read
		s = db.NewSession()
		defer s.Close()
		service = NewNotificationsService(s)
		notifs, _, _, err := service.GetNotificationsForUser(u.ID, 100, 0)
		require.NoError(t, err)

		for _, notif := range notifs {
			assert.False(t, notif.ReadAt.IsZero(), "all notifications should be marked as read")
		}
	})
}

func TestNotificationsService_Notify(t *testing.T) {
	notifications.Fake()
	defer notifications.Unfake()

	t.Run("Send notification to user", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		service := NewNotificationsService(s)

		u := &user.User{ID: 1}
		testNotif := &testNotification{name: "test.notify"}

		err := service.Notify(u, testNotif)
		require.NoError(t, err)
		_ = s.Commit()

		// Verify notification was saved
		s = db.NewSession()
		defer s.Close()
		service = NewNotificationsService(s)
		notifs, _, _, err := service.GetNotificationsForUser(u.ID, 1, 0)
		require.NoError(t, err)
		assert.NotEmpty(t, notifs)
		assert.Equal(t, "test.notify", notifs[0].Name)
	})

	t.Run("Notification with subject ID", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		service := NewNotificationsService(s)

		u := &user.User{ID: 1}
		testNotif := &testNotificationWithSubject{
			testNotification: testNotification{name: "test.subject"},
			subjectID:        456,
		}

		err := service.Notify(u, testNotif)
		require.NoError(t, err)
		_ = s.Commit()

		// Verify subject ID was saved
		s = db.NewSession()
		defer s.Close()
		service = NewNotificationsService(s)
		notifs, err := service.GetNotificationsForNameAndUser(u.ID, "test.subject", 456)
		require.NoError(t, err)
		assert.NotEmpty(t, notifs)
		assert.Equal(t, int64(456), notifs[0].SubjectID)
	})

	t.Run("Notification with nil mail", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		service := NewNotificationsService(s)

		u := &user.User{ID: 1}
		testNotif := &testNotificationNoMail{name: "test.nomail"}

		err := service.Notify(u, testNotif)
		require.NoError(t, err)
		_ = s.Commit()

		// Should still save to DB
		s = db.NewSession()
		defer s.Close()
		service = NewNotificationsService(s)
		notifs, _, _, err := service.GetNotificationsForUser(u.ID, 100, 0)
		require.NoError(t, err)

		found := false
		for _, notif := range notifs {
			if notif.Name == "test.nomail" {
				found = true
				break
			}
		}
		assert.True(t, found, "notification should be saved even without mail")
	})

	t.Run("Notification with nil DB content", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		service := NewNotificationsService(s)

		u := &user.User{ID: 1}
		testNotif := &testNotificationNoDB{name: "test.nodb"}

		err := service.Notify(u, testNotif)
		require.NoError(t, err)
		_ = s.Commit()

		// Should not save to DB
		s = db.NewSession()
		defer s.Close()
		service = NewNotificationsService(s)
		notifs, err := service.GetNotificationsForNameAndUser(u.ID, "test.nodb", 0)
		require.NoError(t, err)
		assert.Empty(t, notifs, "notification should not be saved to DB")
	})
}

func TestNotificationsService_DeleteNotification(t *testing.T) {
	t.Run("Delete own notification", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		service := NewNotificationsService(s)
		u := &user.User{ID: 1}

		// Create a test notification
		testNotif := &testNotification{name: "test.delete"}
		err := service.Notify(u, testNotif)
		require.NoError(t, err)
		_ = s.Commit()

		// Get the notification ID
		s = db.NewSession()
		defer s.Close()
		service = NewNotificationsService(s)
		notifs, _, _, err := service.GetNotificationsForUser(u.ID, 1, 0)
		require.NoError(t, err)
		require.Greater(t, len(notifs), 0, "should have created notification")
		notifID := notifs[0].ID

		// Delete the notification
		err = service.DeleteNotification(notifID, u.ID)
		require.NoError(t, err)
		_ = s.Commit()

		// Verify deletion
		s = db.NewSession()
		defer s.Close()
		service = NewNotificationsService(s)
		deleted := &notifications.DatabaseNotification{ID: notifID}
		exists, err := s.Get(deleted)
		require.NoError(t, err)
		assert.False(t, exists, "notification should be deleted")
	})

	t.Run("Cannot delete other user's notification", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		service := NewNotificationsService(s)
		u1 := &user.User{ID: 1}
		u2 := &user.User{ID: 2}

		// Create notification for user 1
		testNotif := &testNotification{name: "test.delete.other"}
		err := service.Notify(u1, testNotif)
		require.NoError(t, err)
		_ = s.Commit()

		// Get the notification ID
		s = db.NewSession()
		defer s.Close()
		service = NewNotificationsService(s)
		notifs, _, _, err := service.GetNotificationsForUser(u1.ID, 1, 0)
		require.NoError(t, err)
		require.Greater(t, len(notifs), 0)
		notifID := notifs[0].ID

		// Try to delete as user 2 (should not delete)
		err = service.DeleteNotification(notifID, u2.ID)
		require.NoError(t, err) // No error, just doesn't delete
		_ = s.Commit()

		// Verify notification still exists
		s = db.NewSession()
		defer s.Close()
		notif := &notifications.DatabaseNotification{ID: notifID}
		exists, err := s.Get(notif)
		require.NoError(t, err)
		assert.True(t, exists, "notification should still exist (user 2 can't delete user 1's notification)")
	})
}

func TestNotificationsService_DeleteAllReadNotifications(t *testing.T) {
	t.Run("Delete all read notifications", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		service := NewNotificationsService(s)
		u := &user.User{ID: 1}

		// Create multiple notifications
		for i := 0; i < 5; i++ {
			testNotif := &testNotification{name: fmt.Sprintf("test.delete.read.%d", i)}
			err := service.Notify(u, testNotif)
			require.NoError(t, err)
		}
		_ = s.Commit()

		// Mark first 3 as read
		s = db.NewSession()
		defer s.Close()
		service = NewNotificationsService(s)
		notifs, _, _, err := service.GetNotificationsForUser(u.ID, 5, 0)
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(notifs), 5, "should have at least 5 notifications")

		for i := 0; i < 3; i++ {
			err = service.MarkNotificationAsRead(notifs[i], true)
			require.NoError(t, err)
		}
		_ = s.Commit()

		// Delete all read notifications
		s = db.NewSession()
		defer s.Close()
		service = NewNotificationsService(s)
		err = service.DeleteAllReadNotifications(u.ID)
		require.NoError(t, err)
		_ = s.Commit()

		// Verify only unread notifications remain
		s = db.NewSession()
		defer s.Close()
		service = NewNotificationsService(s)
		remaining, _, _, err := service.GetNotificationsForUser(u.ID, 10, 0)
		require.NoError(t, err)

		// Count unread notifications
		unreadCount := 0
		for _, n := range remaining {
			if n.ReadAt.IsZero() || n.ReadAt.Equal(time.Time{}) {
				unreadCount++
			}
		}

		assert.GreaterOrEqual(t, unreadCount, 2, "should have at least 2 unread notifications remaining")
	})

	t.Run("Delete all read notifications doesn't affect unread", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		service := NewNotificationsService(s)
		u := &user.User{ID: 1}

		// Create notifications
		for i := 0; i < 3; i++ {
			testNotif := &testNotification{name: fmt.Sprintf("test.delete.unread.%d", i)}
			err := service.Notify(u, testNotif)
			require.NoError(t, err)
		}
		_ = s.Commit()

		// Get initial count
		s = db.NewSession()
		defer s.Close()
		service = NewNotificationsService(s)
		before, _, _, err := service.GetNotificationsForUser(u.ID, 10, 0)
		require.NoError(t, err)
		unreadBefore := 0
		for _, n := range before {
			if n.ReadAt.IsZero() || n.ReadAt.Equal(time.Time{}) {
				unreadBefore++
			}
		}

		// Delete all read (should not affect unread)
		err = service.DeleteAllReadNotifications(u.ID)
		require.NoError(t, err)
		_ = s.Commit()

		// Verify unread count unchanged
		s = db.NewSession()
		defer s.Close()
		service = NewNotificationsService(s)
		after, _, _, err := service.GetNotificationsForUser(u.ID, 10, 0)
		require.NoError(t, err)
		unreadAfter := 0
		for _, n := range after {
			if n.ReadAt.IsZero() || n.ReadAt.Equal(time.Time{}) {
				unreadAfter++
			}
		}

		assert.GreaterOrEqual(t, unreadAfter, 3, "should have at least 3 unread notifications (none should be deleted)")
	})

	t.Run("Delete all read notifications for different users", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		service := NewNotificationsService(s)
		u1 := &user.User{ID: 1}
		u2 := &user.User{ID: 2}

		// Create and mark as read for user 1
		testNotif1 := &testNotification{name: "test.delete.user1"}
		err := service.Notify(u1, testNotif1)
		require.NoError(t, err)

		// Create and mark as read for user 2
		testNotif2 := &testNotification{name: "test.delete.user2"}
		err = service.Notify(u2, testNotif2)
		require.NoError(t, err)
		_ = s.Commit()

		// Mark both as read
		s = db.NewSession()
		defer s.Close()
		service = NewNotificationsService(s)
		notifs1, _, _, _ := service.GetNotificationsForUser(u1.ID, 1, 0)
		notifs2, _, _, _ := service.GetNotificationsForUser(u2.ID, 1, 0)
		if len(notifs1) > 0 {
			_ = service.MarkNotificationAsRead(notifs1[0], true)
		}
		if len(notifs2) > 0 {
			_ = service.MarkNotificationAsRead(notifs2[0], true)
		}
		_ = s.Commit()

		// Delete user 1's read notifications
		s = db.NewSession()
		defer s.Close()
		service = NewNotificationsService(s)
		err = service.DeleteAllReadNotifications(u1.ID)
		require.NoError(t, err)
		_ = s.Commit()

		// Verify user 2's notifications still exist
		s = db.NewSession()
		defer s.Close()
		service = NewNotificationsService(s)
		u2Notifs, _, _, err := service.GetNotificationsForUser(u2.ID, 10, 0)
		require.NoError(t, err)
		assert.Greater(t, len(u2Notifs), 0, "user 2 should still have notifications")
	})
}

// Test notification types

type testNotification struct {
	name string
}

func (t *testNotification) ToMail(lang string) *notifications.Mail {
	return notifications.NewMail().
		Subject("Test notification").
		Line("This is a test notification")
}

func (t *testNotification) ToDB() interface{} {
	return t
}

func (t *testNotification) Name() string {
	return t.name
}

type testNotificationWithSubject struct {
	testNotification
	subjectID int64
}

func (t *testNotificationWithSubject) SubjectID() int64 {
	return t.subjectID
}

type testNotificationNoMail struct {
	name string
}

func (t *testNotificationNoMail) ToMail(lang string) *notifications.Mail {
	return nil
}

func (t *testNotificationNoMail) ToDB() interface{} {
	return t
}

func (t *testNotificationNoMail) Name() string {
	return t.name
}

type testNotificationNoDB struct {
	name string
}

func (t *testNotificationNoDB) ToMail(lang string) *notifications.Mail {
	return notifications.NewMail().
		Subject("Test notification").
		Line("This is a test notification")
}

func (t *testNotificationNoDB) ToDB() interface{} {
	return nil
}

func (t *testNotificationNoDB) Name() string {
	return t.name
}
