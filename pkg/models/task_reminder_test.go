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
	"encoding/json"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/mail"
	"code.vikunja.io/api/pkg/notifications"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReminderGetTasksInTheNextMinute(t *testing.T) {
	t.Run("Found Tasks", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		now, err := time.Parse(time.RFC3339Nano, "2018-12-01T01:12:00Z")
		require.NoError(t, err)
		notifications, err := getTasksWithRemindersDueAndTheirUsers(s, now)
		require.NoError(t, err)
		assert.Len(t, notifications, 1)
		assert.Equal(t, int64(27), notifications[0].Task.ID)
		assert.Equal(t, "TEST1-18", notifications[0].Task.Identifier)
	})
	t.Run("reminder at window start", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		now := time.Date(2018, 12, 1, 1, 12, 0, 0, time.UTC)
		_, err := s.ID(1).Cols("reminder").Update(&TaskReminder{Reminder: now})
		require.NoError(t, err)

		notifications, err := getTasksWithRemindersDueAndTheirUsers(s, now)
		require.NoError(t, err)
		require.Len(t, notifications, 1)
		assert.Equal(t, int64(27), notifications[0].Task.ID)
	})
	t.Run("completed task is excluded", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := s.ID(27).Cols("done").Update(&Task{Done: true})
		require.NoError(t, err)

		now := time.Date(2018, 12, 1, 1, 12, 0, 0, time.UTC)
		notifications, err := getTasksWithRemindersDueAndTheirUsers(s, now)
		require.NoError(t, err)
		assert.Empty(t, notifications)
	})
	t.Run("Found No Tasks", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		now, err := time.Parse(time.RFC3339Nano, "2018-12-02T01:13:00Z")
		require.NoError(t, err)
		taskIDs, err := getTasksWithRemindersDueAndTheirUsers(s, now)
		require.NoError(t, err)
		assert.Empty(t, taskIDs)
	})
}

func TestGetTaskUsersForTasks(t *testing.T) {
	t.Run("task owner", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Task 1 is owned by user 1 (created_by_id: 1) in project 1 (owned by user 1)
		taskUsers, err := getTaskUsersForTasks(s, []int64{1}, nil)
		require.NoError(t, err)
		require.NotEmpty(t, taskUsers)

		// Should include the task creator
		hasUser1 := false
		for _, tu := range taskUsers {
			if tu.User.ID == 1 && tu.Task.ID == 1 {
				hasUser1 = true
				break
			}
		}
		assert.True(t, hasUser1, "task owner should be included in task users")
	})

	t.Run("project shared directly with user", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Task 32 is in project 3, which is shared directly with user 1 (users_projects id: 1)
		taskUsers, err := getTaskUsersForTasks(s, []int64{32}, nil)
		require.NoError(t, err)
		require.NotEmpty(t, taskUsers)

		// Should include user 1 who has direct share
		hasUser1 := false
		for _, tu := range taskUsers {
			if tu.User.ID == 1 && tu.Task.ID == 32 {
				hasUser1 = true
				break
			}
		}
		assert.True(t, hasUser1, "user with direct project share should be included")
	})

	t.Run("creator who lost project access", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Task 1 is in project 1 (owned by user 1)
		// Task 1 was created by user 1 (created_by_id: 1)
		// User 13 has no access to project 1
		// Create a scenario by pretending user 13 created the task but has no access

		_, err := s.
			Cols("created_by_id").
			Where("id = ?", 1).
			Update(&Task{CreatedByID: 13})
		require.NoError(t, err)

		taskUsers, err := getTaskUsersForTasks(s, []int64{1}, nil)
		require.NoError(t, err)

		// Should only include users with access
		// User 13 should not be in the results (no access to project 1)
		hasUser13 := false
		for _, tu := range taskUsers {
			if tu.User.ID == 13 && tu.Task.ID == 1 {
				hasUser13 = true
				break
			}
		}
		assert.False(t, hasUser13, "creator without project access should be filtered out")
	})

	t.Run("subscriber who lost project access", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Task 2 is in project 1 (owned by user 1)
		// Create a subscription for user 13 who has no access to project 1
		subscription := &Subscription{
			EntityType: SubscriptionEntityTask,
			EntityID:   2,
			UserID:     13,
		}
		_, err := s.Insert(subscription)
		require.NoError(t, err)

		taskUsers, err := getTaskUsersForTasks(s, []int64{2}, nil)
		require.NoError(t, err)

		// User 13 should NOT be in the results (subscribed but no access to project 1)
		hasUser13 := false
		for _, tu := range taskUsers {
			if tu.User.ID == 13 && tu.Task.ID == 2 {
				hasUser13 = true
				break
			}
		}
		assert.False(t, hasUser13, "subscriber without project access should be filtered out")
	})

	t.Run("assignees - with and without project access", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Task 30 has assignees: user 1 and user 2 (task_assignees)
		// Task 30 is in project 1, owned by user 1
		// User 1 has access (owner), user 2 does NOT have access to project 1
		taskUsers, err := getTaskUsersForTasks(s, []int64{30}, nil)
		require.NoError(t, err)
		require.NotEmpty(t, taskUsers)

		// Should include user 1 (assignee WITH project access)
		// Should NOT include user 2 (assignee WITHOUT project access)
		hasUser1 := false
		hasUser2 := false
		for _, tu := range taskUsers {
			if tu.Task.ID == 30 {
				if tu.User.ID == 1 {
					hasUser1 = true
				}
				if tu.User.ID == 2 {
					hasUser2 = true
				}
			}
		}
		assert.True(t, hasUser1, "assignee with project access should be included")
		assert.False(t, hasUser2, "assignee without project access should be filtered out")
	})

	t.Run("subscribers - with project access", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Task 2 has subscription from user 1 (subscriptions id: 1)
		// Task 2 is in project 1, owned by user 1
		// User 1 has access as the owner
		taskUsers, err := getTaskUsersForTasks(s, []int64{2}, nil)
		require.NoError(t, err)
		require.NotEmpty(t, taskUsers)

		// Should include the subscriber who has access
		hasUser1 := false
		for _, tu := range taskUsers {
			if tu.User.ID == 1 && tu.Task.ID == 2 {
				hasUser1 = true
				break
			}
		}
		assert.True(t, hasUser1, "subscriber with project access should be included")
	})

	t.Run("no duplicate users", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Task 30: user 1 is both creator and assignee
		taskUsers, err := getTaskUsersForTasks(s, []int64{30}, nil)
		require.NoError(t, err)
		require.NotEmpty(t, taskUsers)

		// Count how many times user 1 appears for task 30
		user1Count := 0
		for _, tu := range taskUsers {
			if tu.User.ID == 1 && tu.Task.ID == 30 {
				user1Count++
			}
		}
		assert.Equal(t, 1, user1Count, "each user should appear only once per task")
	})

	t.Run("empty task list", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		taskUsers, err := getTaskUsersForTasks(s, []int64{}, nil)
		require.NoError(t, err)
		assert.Empty(t, taskUsers)
	})

	t.Run("multiple tasks with various relationships", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Task 1: user 1 is creator and owner
		// Task 2: user 1 is subscriber and owner
		// Task 30: user 1 is assignee and owner, user 2 is assignee without access
		taskUsers, err := getTaskUsersForTasks(s, []int64{1, 2, 30}, nil)
		require.NoError(t, err)
		require.NotEmpty(t, taskUsers)

		// Count unique task IDs in results
		taskIDs := make(map[int64]bool)
		for _, tu := range taskUsers {
			taskIDs[tu.Task.ID] = true
		}
		assert.True(t, taskIDs[1], "should include users for task 1")
		assert.True(t, taskIDs[2], "should include users for task 2")
		assert.True(t, taskIDs[30], "should include users for task 30")

		// Verify user 2 is NOT included for any task (no access to project 1)
		hasUser2 := false
		for _, tu := range taskUsers {
			if tu.User.ID == 2 {
				hasUser2 = true
				break
			}
		}
		assert.False(t, hasUser2, "user without project access should not be included for any task")
	})
}

func TestGetTaskUsersForTasksIsAssignee(t *testing.T) {
	findTaskUser := func(taskUsers []*taskUser, taskID, userID int64) *taskUser {
		for _, tu := range taskUsers {
			if tu.Task.ID == taskID && tu.User.ID == userID {
				return tu
			}
		}
		return nil
	}

	t.Run("assignee", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Task 30: user 1 is creator and assignee
		taskUsers, err := getTaskUsersForTasks(s, []int64{30}, nil)
		require.NoError(t, err)

		tu := findTaskUser(taskUsers, 30, 1)
		require.NotNil(t, tu)
		assert.True(t, tu.IsAssignee)
	})

	t.Run("creator", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Task 1: user 1 is creator only
		taskUsers, err := getTaskUsersForTasks(s, []int64{1}, nil)
		require.NoError(t, err)

		tu := findTaskUser(taskUsers, 1, 1)
		require.NotNil(t, tu)
		assert.False(t, tu.IsAssignee)
	})

	t.Run("subscriber", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task := &Task{
			Title:       "Subscribed only",
			CreatedByID: 1,
			ProjectID:   1,
		}
		err := task.Create(s, &user.User{ID: 1})
		require.NoError(t, err)

		_, err = s.Insert(&ProjectUser{UserID: 2, ProjectID: 1, Permission: PermissionRead})
		require.NoError(t, err)
		_, err = s.Insert(&Subscription{EntityType: SubscriptionEntityTask, EntityID: task.ID, UserID: 2})
		require.NoError(t, err)

		taskUsers, err := getTaskUsersForTasks(s, []int64{task.ID}, nil)
		require.NoError(t, err)

		tu := findTaskUser(taskUsers, task.ID, 2)
		require.NotNil(t, tu)
		assert.False(t, tu.IsAssignee)
	})

	t.Run("assigned and subscribed", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task := &Task{
			Title:       "Assigned and subscribed",
			CreatedByID: 1,
			ProjectID:   1,
		}
		err := task.Create(s, &user.User{ID: 1})
		require.NoError(t, err)

		_, err = s.Insert(&ProjectUser{UserID: 2, ProjectID: 1, Permission: PermissionRead})
		require.NoError(t, err)
		_, err = s.Insert(&TaskAssginee{TaskID: task.ID, UserID: 2})
		require.NoError(t, err)
		_, err = s.Insert(&Subscription{EntityType: SubscriptionEntityTask, EntityID: task.ID, UserID: 2})
		require.NoError(t, err)

		taskUsers, err := getTaskUsersForTasks(s, []int64{task.ID}, nil)
		require.NoError(t, err)

		count := 0
		for _, tu := range taskUsers {
			if tu.Task.ID == task.ID && tu.User.ID == 2 {
				count++
			}
		}
		assert.Equal(t, 1, count, "user should appear once per task")

		tu := findTaskUser(taskUsers, task.ID, 2)
		require.NotNil(t, tu)
		assert.True(t, tu.IsAssignee)
	})
}

func TestSendDueReminders(t *testing.T) {
	t.Run("stores in-app notification when mailer is disabled", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		oldMailer := config.MailerEnabled.GetBool()
		config.MailerEnabled.Set(false)
		t.Cleanup(func() { config.MailerEnabled.Set(oldMailer) })

		s := db.NewSession()
		defer s.Close()

		_, err := s.Where("id = ?", 1).Cols("email_reminders_enabled").Update(&user.User{EmailRemindersEnabled: false})
		require.NoError(t, err)

		now, err := time.Parse(time.RFC3339Nano, "2018-12-01T01:12:00Z")
		require.NoError(t, err)
		err = sendDueReminders(s, now, false)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		assertSingleReminderNotification(t, 1, 27, 1)
	})
	t.Run("queues mail only after commit when email reminders are on", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		enableReminderMail(t)

		s := db.NewSession()
		defer s.Close()

		_, err := s.Where("id = ?", 1).Cols("email_reminders_enabled").Update(&user.User{EmailRemindersEnabled: true})
		require.NoError(t, err)

		now, err := time.Parse(time.RFC3339Nano, "2018-12-01T01:12:00Z")
		require.NoError(t, err)
		err = sendDueReminders(s, now, false)
		require.NoError(t, err)
		assert.Empty(t, mail.SentMails())
		require.NoError(t, s.Commit())

		sent := mail.SentMails()
		require.Len(t, sent, 1)
		assert.Equal(t, "user1@example.com", sent[0].To)
		assertSingleReminderNotification(t, 1, 27, 1)
	})
	t.Run("stores in-app notification without mail when the user disabled email reminders", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		enableReminderMail(t)

		s := db.NewSession()
		defer s.Close()

		_, err := s.Where("id = ?", 1).Cols("email_reminders_enabled").Update(&user.User{EmailRemindersEnabled: false})
		require.NoError(t, err)

		now, err := time.Parse(time.RFC3339Nano, "2018-12-01T01:12:00Z")
		require.NoError(t, err)
		err = sendDueReminders(s, now, false)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		assert.Empty(t, mail.SentMails())
		assertSingleReminderNotification(t, 1, 27, 1)
	})
}

func enableReminderMail(t *testing.T) {
	t.Helper()

	oldMailer := config.MailerEnabled.GetBool()
	oldReminders := config.ServiceEnableEmailReminders.GetBool()
	config.MailerEnabled.Set(true)
	config.ServiceEnableEmailReminders.Set(true)
	mail.ResetSent()
	t.Cleanup(func() {
		config.MailerEnabled.Set(oldMailer)
		config.ServiceEnableEmailReminders.Set(oldReminders)
		mail.ResetSent()
	})
}

func assertSingleReminderNotification(t *testing.T, userID, taskID, projectID int64) {
	t.Helper()

	s := db.NewSession()
	defer s.Close()

	rows := []*notifications.DatabaseNotification{}
	err := s.Where("notifiable_id = ? AND name = ?", userID, "task.reminder").Find(&rows)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, projectID, rows[0].ProjectID)

	payload, err := json.Marshal(rows[0].Notification)
	require.NoError(t, err)
	var parsed struct {
		Task struct {
			ID int64 `json:"id"`
		} `json:"task"`
	}
	require.NoError(t, json.Unmarshal(payload, &parsed))
	assert.Equal(t, taskID, parsed.Task.ID)
}

func TestReminderDueNotificationToMail(t *testing.T) {
	n := &ReminderDueNotification{
		User:    &user.User{ID: 1, Name: "alice"},
		Task:    &Task{ID: 27, Title: "task"},
		Project: &Project{ID: 1, Title: "proj"},
	}

	t.Run("nil when user disabled email reminders", func(t *testing.T) {
		n.User.EmailRemindersEnabled = false
		assert.Nil(t, n.ToMail("en"))
	})
	t.Run("nil when email reminders disabled instance-wide", func(t *testing.T) {
		old := config.ServiceEnableEmailReminders.GetBool()
		config.ServiceEnableEmailReminders.Set(false)
		t.Cleanup(func() { config.ServiceEnableEmailReminders.Set(old) })

		n.User.EmailRemindersEnabled = true
		assert.Nil(t, n.ToMail("en"))
	})
	t.Run("mail when enabled", func(t *testing.T) {
		n.User.EmailRemindersEnabled = true
		assert.NotNil(t, n.ToMail("en"))
	})
}
