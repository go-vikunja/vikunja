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
	"code.vikunja.io/api/pkg/metrics"
	"code.vikunja.io/api/pkg/modules/keyvalue"
	"code.vikunja.io/api/pkg/notifications"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/require"
)

func TestDeleteUser(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		notifications.Fake()

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

		project := &Project{
			Title:   "disabled user project",
			OwnerID: 17,
		}
		_, err := s.Insert(project)
		require.NoError(t, err)

		task := &Task{
			Title:       "disabled user task",
			ProjectID:   project.ID,
			CreatedByID: 17,
			Index:       1,
		}
		_, err = s.Insert(task)
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
	t.Run("decrements user count after committed delete is dispatched", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		notifications.Fake()
		events.Unfake()
		t.Cleanup(events.Fake)
		user.RegisterListeners()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ready, err := events.InitEventsForTesting(ctx)
		require.NoError(t, err)
		<-ready

		require.NoError(t, keyvalue.Put(metrics.UserCountKey, int64(2)))
		t.Cleanup(func() {
			_ = keyvalue.Del(metrics.UserCountKey)
		})

		err = DeleteUser(s, &user.User{ID: 4})
		require.NoError(t, err)

		require.NoError(t, s.Commit())
		events.DispatchPending(s)

		require.Eventually(t, func() bool {
			value, exists, err := keyvalue.Get(metrics.UserCountKey)
			if err != nil || !exists {
				return false
			}

			count, ok := value.(int64)
			return ok && count == 1
		}, time.Second, 10*time.Millisecond)
	})
}
