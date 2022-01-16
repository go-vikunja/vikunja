// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2021 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"testing"
	"time"

	"code.vikunja.io/api/pkg/db"
	"github.com/stretchr/testify/assert"
)

func TestReminderGetTasksInTheNextMinute(t *testing.T) {
	t.Run("Found Tasks", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		now, err := time.Parse(time.RFC3339Nano, "2018-12-01T01:13:00Z")
		assert.NoError(t, err)
		notifications, err := getTasksWithRemindersDueAndTheirUsers(s, now)
		assert.NoError(t, err)
		assert.Len(t, notifications, 1)
		assert.Equal(t, int64(27), notifications[0].Task.ID)
	})
	t.Run("Found No Tasks", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		now, err := time.Parse(time.RFC3339Nano, "2018-12-02T01:13:00Z")
		assert.NoError(t, err)
		taskIDs, err := getTasksWithRemindersDueAndTheirUsers(s, now)
		assert.NoError(t, err)
		assert.Len(t, taskIDs, 0)
	})
}
