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
	"code.vikunja.io/api/pkg/web"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

// Whole-day relative reminders count calendar days in the saver's zone.
// Instants are UTC literals with their New York reading beside them.

// TestUpdateRelativeReminderDates: the stored reminder instant of each relation is the calendar-day
// one (elapsed would give Sat 19:00 EST, Sun 21:00 EDT and Sat 17:00 EST).
func TestUpdateRelativeReminderDates(t *testing.T) {
	ny := mustZone(t, "America/New_York")
	due := utc(2027, 3, 15, 0, 0)   // Sun 2027-03-14 20:00 EDT
	start := utc(2027, 3, 14, 1, 0) // Sat 2027-03-13 20:00 EST
	task := &Task{
		DueDate:   due,
		StartDate: start,
		EndDate:   due,
		Reminders: []*TaskReminder{
			{RelativeTo: ReminderRelationDueDate, RelativePeriod: -86400},
			{RelativeTo: ReminderRelationStartDate, RelativePeriod: 86400},
			{RelativeTo: ReminderRelationEndDate, RelativePeriod: -93600},
		},
	}
	require.NoError(t, updateRelativeReminderDates(task, ny))

	want := []time.Time{
		utc(2027, 3, 14, 1, 0),  // Sat 20:00 EST
		utc(2027, 3, 15, 0, 0),  // Sun 20:00 EDT
		utc(2027, 3, 13, 23, 0), // Sat 18:00 EST
	}
	for i, w := range want {
		assert.True(t, task.Reminders[i].Reminder.Equal(w), "reminder %d: got %v, want %v", i, task.Reminders[i].Reminder.UTC(), w)
	}

	t.Run("a reminder without its base date is cleared", func(t *testing.T) {
		task := &Task{
			DueDate:   due,
			Reminders: []*TaskReminder{{RelativeTo: ReminderRelationStartDate, RelativePeriod: -86400, Reminder: due}},
		}
		require.NoError(t, updateRelativeReminderDates(task, ny))
		assert.True(t, task.Reminders[0].Reminder.IsZero())
	})
	t.Run("a period without a relation is refused", func(t *testing.T) {
		task := &Task{DueDate: due, Reminders: []*TaskReminder{{RelativePeriod: -86400}}}
		assert.True(t, IsErrReminderRelativeToMissing(updateRelativeReminderDates(task, ny)))
	})
}

// TestUpdateRelativeReminderDatesMatchesRangeGuard: the storable-range guard judges exactly the
// instant updateRelativeReminderDates stores.
func TestUpdateRelativeReminderDatesMatchesRangeGuard(t *testing.T) {
	ny := mustZone(t, "America/New_York")
	cases := append(newYorkRelativeReminderCases(), subDayRelativeReminderCases()...)
	cases = append(cases,
		// due 9999-12-30 20:00 UTC: one day after is storable, two days after is the year 10000.
		relativeReminderCase{name: "one day before the storable end", base: utc(9999, 12, 30, 20, 0), period: 86400, want: utc(9999, 12, 31, 20, 0)},
		relativeReminderCase{name: "past the storable end", base: utc(9999, 12, 30, 20, 0), period: 2 * 86400, want: time.Date(10000, 1, 1, 20, 0, 0, 0, time.UTC)},
	)
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			newTask := func() *Task {
				return &Task{DueDate: c.base, Reminders: []*TaskReminder{{RelativeTo: ReminderRelationDueDate, RelativePeriod: c.period}}}
			}
			stored := newTask()
			require.NoError(t, updateRelativeReminderDates(stored, ny))
			judged := storableRangeDates(newTask(), ny)

			stamp := stored.Reminders[0].Reminder
			assert.True(t, stamp.Equal(c.want), "stored %v, want %v", stamp.UTC(), c.want)
			assert.True(t, judged[len(judged)-1].Equal(stamp), "the guard judged %v, the store wrote %v", judged[len(judged)-1].UTC(), stamp.UTC())
			assert.Equal(t, stamp.UTC().Year() > 9999, datesOutsideStorableRange(newTask(), ny))
		})
	}
}

// countReminderZoneResolutions swaps the reminder zone seam for a counting wrapper.
func countReminderZoneResolutions(t *testing.T) *int {
	t.Helper()
	calls := 0
	saved := resolveReminderTaskZone
	t.Cleanup(func() { resolveReminderTaskZone = saved })
	resolveReminderTaskZone = func(s *xorm.Session, a web.Auth, projectID int64) *time.Location {
		calls++
		return saved(s, a, projectID)
	}
	return &calls
}

// storedReminders reads a task's reminder instants back from task_reminders, earliest first.
func storedReminders(t *testing.T, s *xorm.Session, taskID int64) []time.Time {
	t.Helper()
	rows := []*TaskReminder{}
	require.NoError(t, s.Where("task_id = ?", taskID).OrderBy("reminder asc").Find(&rows))
	out := make([]time.Time, len(rows))
	for i, r := range rows {
		out[i] = r.Reminder
	}
	return out
}

func assertStoredReminder(t *testing.T, s *xorm.Session, taskID int64, want time.Time) {
	t.Helper()
	got := storedReminders(t, s, taskID)
	if assert.Len(t, got, 1) {
		assert.True(t, got[0].Equal(want), "stored %v, want %v", got[0].UTC(), want)
	}
}

// TestTaskRelativeReminder: creating and updating a task through the model stores a whole-day
// relative reminder on the saver's calendar.
func TestTaskRelativeReminder(t *testing.T) {
	newYorker := &user.User{ID: 1}
	due := utc(2027, 3, 15, 0, 0) // Sun 2027-03-14 20:00 EDT
	dayBefore := func() []*TaskReminder {
		return []*TaskReminder{{RelativeTo: ReminderRelationDueDate, RelativePeriod: -86400}}
	}

	t.Run("create, update, due change and last saver", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		setUserZone(t, s, 1, "America/New_York")
		calls := countReminderZoneResolutions(t)

		task := &Task{Title: "relative reminder", ProjectID: 1, DueDate: due, Reminders: dayBefore()}
		require.NoError(t, task.Create(s, newYorker))
		assert.Equal(t, 1, *calls, "a whole-day relative reminder resolves the zone once")
		assertStoredReminder(t, s, task.ID, utc(2027, 3, 14, 1, 0)) // Sat 20:00 EST, not elapsed 19:00

		// Re-saved with a new title: still Sat 20:00 EST.
		update := &Task{ID: task.ID, Title: "renamed", ProjectID: 1, DueDate: due, Reminders: dayBefore()}
		require.NoError(t, update.Update(s, newYorker))
		assertStoredReminder(t, s, task.ID, utc(2027, 3, 14, 1, 0))

		// The due date moves to Sun 2027-03-21 20:00 EDT: the reminder is Sat 03-20 20:00 EDT.
		update = &Task{ID: task.ID, Title: "renamed", ProjectID: 1, DueDate: utc(2027, 3, 22, 0, 0), Reminders: dayBefore()}
		require.NoError(t, update.Update(s, newYorker))
		assertStoredReminder(t, s, task.ID, utc(2027, 3, 21, 0, 0))

		// Last saver wins: a Berlin user re-saves the original due date unchanged. One day
		// before Mon 03-15 01:00 CET is Sun 03-14 01:00 CET (00:00Z), Berlin's calendar, which reads
		// Sat 19:00 EST in New York.
		setUserZone(t, s, 2, "Europe/Berlin")
		update = &Task{ID: task.ID, Title: "renamed", ProjectID: 1, DueDate: due, Reminders: dayBefore()}
		require.NoError(t, update.Update(s, &user.User{ID: 2}))
		assertStoredReminder(t, s, task.ID, utc(2027, 3, 14, 0, 0))
		// The same save by the New Yorker puts it back on New York's calendar.
		require.NoError(t, update.Update(s, newYorker))
		assertStoredReminder(t, s, task.ID, utc(2027, 3, 14, 1, 0))
	})

	t.Run("a sub-day period resolves no zone", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		setUserZone(t, s, 1, "America/New_York")
		calls := countReminderZoneResolutions(t)

		task := &Task{Title: "two hours before", ProjectID: 1, DueDate: due,
			Reminders: []*TaskReminder{{RelativeTo: ReminderRelationDueDate, RelativePeriod: -7200}}}
		require.NoError(t, task.Create(s, newYorker))
		assert.Equal(t, 0, *calls)
		assertStoredReminder(t, s, task.ID, utc(2027, 3, 14, 22, 0)) // Sun 18:00 EDT
	})

	t.Run("a reminder without its base date resolves no zone and is cleared", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		setUserZone(t, s, 1, "America/New_York")
		calls := countReminderZoneResolutions(t)

		task := &Task{Title: "no start date", ProjectID: 1, DueDate: due,
			Reminders: []*TaskReminder{{RelativeTo: ReminderRelationStartDate, RelativePeriod: -86400}}}
		require.NoError(t, task.Create(s, newYorker))
		assert.Equal(t, 0, *calls)
		got := storedReminders(t, s, task.ID)
		if assert.Len(t, got, 1) {
			assert.True(t, got[0].IsZero(), "stored %v", got[0])
		}
	})
}

// TestTaskNeedsReminderZone: only a relative reminder of a whole day or more whose base date is
// set needs a zone, so every other write resolves none.
func TestTaskNeedsReminderZone(t *testing.T) {
	due := utc(2027, 3, 15, 0, 0)
	rel := func(to ReminderRelation, period int64) *TaskReminder {
		return &TaskReminder{RelativeTo: to, RelativePeriod: period}
	}
	cases := []struct {
		name      string
		reminders []*TaskReminder
		want      bool
	}{
		{"no reminders", nil, false},
		{"absolute only", []*TaskReminder{{Reminder: due}}, false},
		{"relative 2 hours before", []*TaskReminder{rel(ReminderRelationDueDate, -7200)}, false},
		{"relative 1 day before with its base", []*TaskReminder{rel(ReminderRelationDueDate, -86400)}, true},
		{"relative 1 day before without its base", []*TaskReminder{rel(ReminderRelationStartDate, -86400)}, false},
		{"relative just under a day after", []*TaskReminder{rel(ReminderRelationDueDate, 86399)}, false},
		{"relative 2 days before", []*TaskReminder{rel(ReminderRelationEndDate, -172800)}, true},
		{"one of several", []*TaskReminder{{Reminder: due}, rel(ReminderRelationDueDate, -60), rel(ReminderRelationDueDate, 86400)}, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			task := &Task{DueDate: due, EndDate: due, Reminders: c.reminders}
			assert.Equal(t, c.want, taskNeedsReminderZone(task))
		})
	}
}
