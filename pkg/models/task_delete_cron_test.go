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
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/files"
)

func TestDeleteExpiredTasks(t *testing.T) {
	// Task 51 was soft-deleted at this time in the fixtures
	deletedAt := time.Date(2018, 12, 1, 1, 12, 4, 0, time.UTC)

	t.Run("older than the retention period", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		files.InitTestFileFixtures(t)

		deleteExpiredTasks(deletedAt.Add(taskDeleteRetention + 24*time.Hour))

		db.AssertMissing(t, "tasks", map[string]interface{}{"id": 51})
		db.AssertMissing(t, "task_reminders", map[string]interface{}{"task_id": 51})
		db.AssertMissing(t, "label_tasks", map[string]interface{}{"task_id": 51})
		db.AssertMissing(t, "subscriptions", map[string]interface{}{"entity_id": 51, "entity_type": SubscriptionEntityTask})
	})

	t.Run("newer than the retention period", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		deleteExpiredTasks(deletedAt.Add(taskDeleteRetention - 24*time.Hour))

		db.AssertExists(t, "tasks", map[string]interface{}{"id": 51}, false)
		db.AssertExists(t, "task_reminders", map[string]interface{}{"task_id": 51}, false)
	})
}
