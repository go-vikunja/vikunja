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
	"strings"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func remindersForTest(count int) []*TaskReminder {
	reminders := make([]*TaskReminder, count)
	for i := range reminders {
		reminders[i] = &TaskReminder{Reminder: time.Unix(int64(1_700_000_000+i*60), 0)}
	}
	return reminders
}

func assertInvalidField(t *testing.T, err error, field string) {
	t.Helper()
	require.Error(t, err)
	var verr ValidationHTTPError
	require.ErrorAs(t, err, &verr)
	assert.Contains(t, verr.InvalidFields, field)
}

func TestTaskLimits(t *testing.T) {
	u := &user.User{ID: 1}

	t.Run("create with too many reminders", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task := &Task{Title: "x", ProjectID: 1, Reminders: remindersForTest(MaxRemindersPerTask + 1)}
		assertInvalidField(t, task.Create(s, u), "reminders")
	})

	t.Run("create with the maximum reminders", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task := &Task{Title: "x", ProjectID: 1, Reminders: remindersForTest(MaxRemindersPerTask)}
		require.NoError(t, task.Create(s, u))
	})

	t.Run("create with too long title", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task := &Task{Title: strings.Repeat("a", MaxTaskTitleLength+1), ProjectID: 1}
		assertInvalidField(t, task.Create(s, u), "title")
	})

	t.Run("title length counts characters", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task := &Task{Title: strings.Repeat("ä", MaxTaskTitleLength), ProjectID: 1}
		require.NoError(t, task.Create(s, u))
	})

	t.Run("bulk create checks every task", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		btc := &BulkTaskCreation{ProjectID: 1, Tasks: []*Task{
			{Title: "ok"},
			{Title: "x", Reminders: remindersForTest(MaxRemindersPerTask + 1)},
		}}
		err := btc.Create(s, u)
		var berr ErrInvalidTaskInBulkCreation
		require.ErrorAs(t, err, &berr)
		assert.Equal(t, 1, berr.Index)
	})

	t.Run("update with too many reminders", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task := &Task{ID: 1, Title: "task #1", ProjectID: 1, Reminders: remindersForTest(MaxRemindersPerTask + 1)}
		assertInvalidField(t, task.Update(s, u), "reminders")
	})

	t.Run("update to too long title", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task := &Task{ID: 1, Title: strings.Repeat("a", MaxTaskTitleLength+1), ProjectID: 1}
		assertInvalidField(t, task.Update(s, u), "title")
	})

	t.Run("update keeps an existing long title", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		long := strings.Repeat("a", MaxTaskTitleLength+1)
		_, err := s.ID(1).Cols("title").Update(&Task{Title: long})
		require.NoError(t, err)

		task := &Task{ID: 1, Title: long, ProjectID: 1, Done: true}
		require.NoError(t, task.Update(s, u))
	})

	t.Run("update keeps existing reminders over the limit", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		stored := remindersForTest(MaxRemindersPerTask + 1)
		for _, r := range stored {
			r.TaskID = 1
		}
		_, err := s.Insert(stored)
		require.NoError(t, err)

		task := &Task{ID: 1, Done: true}
		require.NoError(t, task.updateSingleTask(s, u, []string{"done"}))

		task = &Task{ID: 1, Title: "task #1", ProjectID: 1, Reminders: remindersForTest(MaxRemindersPerTask + 2)}
		assertInvalidField(t, task.Update(s, u), "reminders")
	})
}
