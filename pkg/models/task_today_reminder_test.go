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
	"code.vikunja.io/api/pkg/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetTasksForDailyReminder(t *testing.T) {
	t.Run("no undone tasks", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		now, err := time.Parse(time.RFC3339Nano, "2018-01-01T01:13:00Z")
		require.NoError(t, err)
		tasks, err := getTasksForDailyReminder(s, now)
		require.NoError(t, err)
		assert.Empty(t, tasks)
	})
	t.Run("overdue and due today", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		now, err := time.Parse(time.RFC3339Nano, "2018-12-01T09:00:00Z")
		require.NoError(t, err)
		uts, err := getTasksForDailyReminder(s, now)
		require.NoError(t, err)
		assert.Len(t, uts, 1)
		assert.Len(t, uts[1].overdue, 2)
		assert.Len(t, uts[1].dueToday, 1)
		_, ok := uts[1].dueToday[47]
		assert.True(t, ok)

		// Disable today reminders and ensure the task is not included
		_, err = s.Where("id = ?", 1).Cols("today_tasks_reminders_enabled").Update(&user.User{TodayTasksRemindersEnabled: false})
		require.NoError(t, err)
		uts, err = getTasksForDailyReminder(s, now)
		require.NoError(t, err)
		assert.Len(t, uts[1].overdue, 2)
		assert.Empty(t, uts[1].dueToday)
	})
	t.Run("only due today", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// disable overdue reminders, keep today reminders enabled
		_, err := s.Where("id = ?", 1).Cols("overdue_tasks_reminders_enabled").Update(&user.User{OverdueTasksRemindersEnabled: false})
		require.NoError(t, err)

		now, err := time.Parse(time.RFC3339Nano, "2018-12-01T09:00:00Z")
		require.NoError(t, err)
		uts, err := getTasksForDailyReminder(s, now)
		require.NoError(t, err)
		assert.Len(t, uts, 1)
		assert.Empty(t, uts[1].overdue)
		assert.Len(t, uts[1].dueToday, 1)
	})
	t.Run("done overdue", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		now, err := time.Parse(time.RFC3339Nano, "2018-11-01T01:13:00Z")
		require.NoError(t, err)
		tasks, err := getTasksForDailyReminder(s, now)
		require.NoError(t, err)
		assert.Empty(t, tasks)
	})
}
