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
	"code.vikunja.io/api/pkg/notifications"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

// TestDatabaseNotifications_ReadAll_RefreshesUsers guards #2720 for notifications
// already in the database: those were serialized with a partial doer (id +
// username, no display Name), so reading them must reload the embedded users so
// the display name is shown. The fix in the dispatch path only helps new
// notifications; old rows are healed here at read time.
func TestDatabaseNotifications_ReadAll_RefreshesUsers(t *testing.T) {
	t.Run("fills in the display name from the database", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// user12 has the display name "Name with spaces" in the fixtures.
		insertStoredNotification(t, s, 1, &TaskAssignedNotification{
			Doer:     &user.User{ID: 12, Username: "user12"},
			Assignee: &user.User{ID: 12, Username: "user12"},
			Task:     &Task{ID: 1},
		})

		got := readAssignedNotification(t, s, 1)
		require.Equal(t, "Name with spaces", got.Doer.GetName())
		require.Equal(t, "Name with spaces", got.Assignee.GetName())
	})

	t.Run("keeps the stored value when the user no longer exists", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		insertStoredNotification(t, s, 1, &TaskAssignedNotification{
			Doer: &user.User{ID: 999999, Username: "ghost"},
			Task: &Task{ID: 1},
		})

		got := readAssignedNotification(t, s, 1)
		require.Equal(t, "ghost", got.Doer.Username)
	})

	t.Run("refreshes a disabled user", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// user17 is disabled in the fixtures; the reload must still win over the
		// stale stored value.
		insertStoredNotification(t, s, 1, &TaskAssignedNotification{
			Doer: &user.User{ID: 17, Username: "stale"},
			Task: &Task{ID: 1},
		})

		got := readAssignedNotification(t, s, 1)
		require.Equal(t, "user17", got.Doer.Username)
	})
}

func insertStoredNotification(t *testing.T, s *xorm.Session, notifiableID int64, n notifications.Notification) {
	t.Helper()
	content, err := json.Marshal(n)
	require.NoError(t, err)
	_, err = s.Insert(&notifications.DatabaseNotification{
		NotifiableID: notifiableID,
		Notification: json.RawMessage(content),
		Name:         n.Name(),
	})
	require.NoError(t, err)
}

func readAssignedNotification(t *testing.T, s *xorm.Session, notifiableID int64) *TaskAssignedNotification {
	t.Helper()
	result, _, _, err := (&DatabaseNotifications{}).ReadAll(s, &user.User{ID: notifiableID}, "", 1, 50)
	require.NoError(t, err)

	for _, dbn := range result.([]*notifications.DatabaseNotification) {
		if n, is := dbn.Notification.(*TaskAssignedNotification); is {
			return n
		}
	}
	t.Fatal("no task.assigned notification was returned")
	return nil
}
