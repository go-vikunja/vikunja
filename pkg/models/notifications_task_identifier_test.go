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

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/notifications"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// assertNotifiedTaskIdentifier reads back the notification persisted for the mentioned
// user and checks the task identifier in its payload - this is what the web client renders.
func assertNotifiedTaskIdentifier(t *testing.T, notificationName string) {
	t.Helper()

	s := db.NewSession()
	defer s.Close()

	dbn := &notifications.DatabaseNotification{}
	has, err := s.
		Where("notifiable_id = ? AND name = ?", notifiedUserID, notificationName).
		Desc("id").
		Get(dbn)
	require.NoError(t, err)
	require.True(t, has, "expected a %s notification for user %d", notificationName, notifiedUserID)

	payload, err := json.Marshal(dbn.Notification)
	require.NoError(t, err)

	var parsed struct {
		Task struct {
			Identifier string `json:"identifier"`
		} `json:"task"`
	}
	require.NoError(t, json.Unmarshal(payload, &parsed))
	assert.Equal(t, taskIdentifierTestWant, parsed.Task.Identifier)
}

// Task 32 lives in project 3 which has the identifier TEST3 and holds the task at index 1.
// User 2 has access to that project.
const taskIdentifierTestTaskID = 32
const taskIdentifierTestWant = "TEST3-1"
const notifiedUserID = 2

func TestNotificationsIncludeProjectPrefix(t *testing.T) {
	doer := &user.User{ID: 1}

	t.Run("task comment", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()

		task, err := GetTaskByIDSimple(s, taskIdentifierTestTaskID)
		require.NoError(t, err)
		require.Empty(t, task.Identifier, "the event payload carries a task without an identifier")

		tc := &TaskComment{
			Comment: `<p>Lorem Ipsum <mention-user data-id="user2">@user2</mention-user></p>`,
			TaskID:  taskIdentifierTestTaskID,
		}
		require.NoError(t, tc.Create(s, doer))
		require.NoError(t, s.Commit())
		_ = s.Close()

		events.TestListener(t, &TaskCommentCreatedEvent{
			Task:    &task,
			Doer:    doer,
			Comment: tc,
		}, &SendTaskCommentNotification{})

		assertNotifiedTaskIdentifier(t, (&TaskCommentNotification{}).Name())
	})

	t.Run("task assigned", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()

		task, err := GetTaskByIDSimple(s, taskIdentifierTestTaskID)
		require.NoError(t, err)

		assignee, err := user.GetUserByID(s, notifiedUserID)
		require.NoError(t, err)

		sub := &Subscription{
			UserID:     notifiedUserID,
			EntityType: SubscriptionEntityTask,
			EntityID:   taskIdentifierTestTaskID,
		}
		require.NoError(t, sub.Create(s, assignee))
		require.NoError(t, s.Commit())
		_ = s.Close()

		events.TestListener(t, &TaskAssigneeCreatedEvent{
			Task:     &task,
			Assignee: assignee,
			Doer:     doer,
		}, &SendTaskAssignedNotification{})

		assertNotifiedTaskIdentifier(t, (&TaskAssignedNotification{}).Name())
	})

	t.Run("task deleted", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()

		task, err := GetTaskByIDSimple(s, taskIdentifierTestTaskID)
		require.NoError(t, err)

		subscriber, err := user.GetUserByID(s, notifiedUserID)
		require.NoError(t, err)

		sub := &Subscription{
			UserID:     notifiedUserID,
			EntityType: SubscriptionEntityTask,
			EntityID:   taskIdentifierTestTaskID,
		}
		require.NoError(t, sub.Create(s, subscriber))
		require.NoError(t, s.Commit())
		_ = s.Close()

		events.TestListener(t, &TaskDeletedEvent{
			Task: &task,
			Doer: doer,
		}, &SendTaskDeletedNotification{})

		assertNotifiedTaskIdentifier(t, (&TaskDeletedNotification{}).Name())
	})

	t.Run("mentioned in task description", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()

		task, err := GetTaskByIDSimple(s, taskIdentifierTestTaskID)
		require.NoError(t, err)
		task.Description = `<p><mention-user data-id="user2">@user2</mention-user></p>`
		require.NoError(t, s.Commit())
		_ = s.Close()

		events.TestListener(t, &TaskUpdatedEvent{
			Task: &task,
			Doer: doer,
		}, &HandleTaskUpdatedMentions{})

		assertNotifiedTaskIdentifier(t, (&UserMentionedInTaskNotification{}).Name())
	})
}
