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

	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

// Task 32 lives on project 3, which user 2 can read.
func TestSendTaskCreatedNotification(t *testing.T) {
	const (
		taskID       int64 = 32
		projectID    int64 = 3
		subscriberID int64 = 2
		doerID       int64 = 1
	)

	notificationName := (&TaskCreatedNotification{}).Name()

	subscribeToProject := func(t *testing.T, s *xorm.Session, userID int64) {
		_, err := s.Insert(&Subscription{
			UserID:     userID,
			EntityType: SubscriptionEntityProject,
			EntityID:   projectID,
		})
		require.NoError(t, err)
	}

	t.Run("notifies project subscribers", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		subscribeToProject(t, s, subscriberID)

		task, err := GetTaskByIDSimple(s, taskID)
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		_ = s.Close()

		events.TestListener(t, &TaskCreatedEvent{
			Task: &task,
			Doer: &user.User{ID: doerID},
		}, &SendTaskCreatedNotification{})

		db.AssertExists(t, "notifications", map[string]interface{}{
			"notifiable_id": subscriberID,
			"name":          notificationName,
		}, false)
	})

	t.Run("does not notify the creator", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		subscribeToProject(t, s, doerID)

		task, err := GetTaskByIDSimple(s, taskID)
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		_ = s.Close()

		events.TestListener(t, &TaskCreatedEvent{
			Task: &task,
			Doer: &user.User{ID: doerID},
		}, &SendTaskCreatedNotification{})

		db.AssertMissing(t, "notifications", map[string]interface{}{
			"notifiable_id": doerID,
			"name":          notificationName,
		})
	})

	// HandleTaskCreateMentions already notifies mentioned users for the same event.
	t.Run("does not notify subscribers mentioned in the description", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		subscribeToProject(t, s, subscriberID)

		task, err := GetTaskByIDSimple(s, taskID)
		require.NoError(t, err)
		task.Description = `<p><mention-user data-id="user2">@user2</mention-user></p>`
		require.NoError(t, s.Commit())
		_ = s.Close()

		events.TestListener(t, &TaskCreatedEvent{
			Task: &task,
			Doer: &user.User{ID: doerID},
		}, &SendTaskCreatedNotification{})

		db.AssertMissing(t, "notifications", map[string]interface{}{
			"notifiable_id": subscriberID,
			"name":          notificationName,
		})
	})
}
