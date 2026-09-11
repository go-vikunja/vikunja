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

	"code.vikunja.io/api/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJalaliConversionVectors(t *testing.T) {
	tests := []struct {
		name string
		gy   int
		gm   time.Month
		gd   int
		jy   int
		jm   int
		jd   int
	}{
		{name: "nowruz 1405", gy: 2026, gm: time.March, gd: 21, jy: 1405, jm: 1, jd: 1},
		{name: "shahrivar 1405", gy: 2026, gm: time.September, gd: 7, jy: 1405, jm: 6, jd: 16},
		{name: "nowruz 1404", gy: 2025, gm: time.March, gd: 21, jy: 1404, jm: 1, jd: 1},
		{name: "esfand 30 leap 1403", gy: 2025, gm: time.March, gd: 20, jy: 1403, jm: 12, jd: 30},
		{name: "esfand 29 common 1404", gy: 2026, gm: time.March, gd: 20, jy: 1404, jm: 12, jd: 29},
		{name: "shahrivar 31", gy: 2025, gm: time.September, gd: 22, jy: 1404, jm: 6, jd: 31},
		{name: "ordibehesht 1 1405", gy: 2026, gm: time.April, gd: 21, jy: 1405, jm: 2, jd: 1},
	}

	for _, tt := range tests {
		t.Run("gregorian to jalali/"+tt.name, func(t *testing.T) {
			jy, jm, jd, ok := gregorianToJalali(tt.gy, tt.gm, tt.gd)
			require.True(t, ok)
			assert.Equal(t, tt.jy, jy)
			assert.Equal(t, tt.jm, jm)
			assert.Equal(t, tt.jd, jd)
		})
		t.Run("jalali to gregorian/"+tt.name, func(t *testing.T) {
			gy, gm, gd, ok := jalaliToGregorian(tt.jy, tt.jm, tt.jd)
			require.True(t, ok)
			assert.Equal(t, tt.gy, gy)
			assert.Equal(t, tt.gm, gm)
			assert.Equal(t, tt.gd, gd)
		})
	}
}

func TestJalaliMonthLength(t *testing.T) {
	for m := 1; m <= 6; m++ {
		assert.Equal(t, 31, jalaliMonthLength(1404, m), "month %d", m)
	}
	for m := 7; m <= 11; m++ {
		assert.Equal(t, 30, jalaliMonthLength(1404, m), "month %d", m)
	}
	assert.Equal(t, 30, jalaliMonthLength(1403, 12), "esfand 1403 is leap")
	assert.Equal(t, 29, jalaliMonthLength(1404, 12), "esfand 1404 is common")
}

func TestAddJalaliMonthsToDate(t *testing.T) {
	utc := time.UTC
	at := func(y int, m time.Month, d, h, mi, s int) time.Time {
		return time.Date(y, m, d, h, mi, s, 0, utc)
	}

	tests := []struct {
		name string
		in   time.Time
		want time.Time
	}{
		{
			name: "shahrivar 31 clamps to mehr 30",
			in:   at(2025, time.September, 22, 12, 0, 0),
			want: at(2025, time.October, 22, 12, 0, 0),
		},
		{
			name: "bahman 30 clamps to esfand 29 in common year",
			in:   at(2026, time.February, 19, 12, 0, 0),
			want: at(2026, time.March, 20, 12, 0, 0),
		},
		{
			name: "bahman 30 needs no clamp into leap esfand",
			in:   at(2025, time.February, 18, 12, 0, 0), // 1403/11/30
			want: at(2025, time.March, 20, 12, 0, 0),    // 1403/12/30
		},
		{
			name: "esfand wraps to farvardin",
			in:   at(2026, time.March, 20, 12, 0, 0), // 1404/12/29
			want: at(2026, time.April, 18, 12, 0, 0), // 1405/01/29
		},
		{
			name: "no clamp between 31-day months",
			in:   at(2025, time.August, 22, 12, 0, 0), // 1404/05/31
			want: at(2025, time.September, 22, 12, 0, 0),
		},
		{
			name: "preserves time of day",
			in:   at(2025, time.September, 22, 8, 45, 33),
			want: at(2025, time.October, 22, 8, 45, 33),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.True(t, addJalaliMonthsToDate(tt.in, 1, utc).Equal(tt.want),
				"got %s, want %s", addJalaliMonthsToDate(tt.in, 1, utc), tt.want)
		})
	}

	t.Run("nil loc falls back to service timezone", func(t *testing.T) {
		in := at(2025, time.September, 22, 12, 0, 0)
		assert.True(t, addJalaliMonthsToDate(in, 1, nil).Equal(addJalaliMonthsToDate(in, 1, config.GetTimeZone())))
	})
}

func TestAddJalaliYearsToDate(t *testing.T) {
	utc := time.UTC
	at := func(y int, m time.Month, d int) time.Time {
		return time.Date(y, m, d, 12, 0, 0, 0, utc)
	}

	tests := []struct {
		name string
		in   time.Time
		want time.Time
	}{
		{
			name: "esfand 30 leap clamps to 29 in common year",
			in:   at(2025, time.March, 20), // 1403/12/30
			want: at(2026, time.March, 20), // 1404/12/29
		},
		{
			name: "esfand 29 common stays 29",
			in:   at(2026, time.March, 20), // 1404/12/29
			want: at(2027, time.March, 20), // 1405/12/29
		},
		{
			name: "ordinary date advances one jalali year",
			in:   at(2025, time.August, 22), // 1404/05/31
			want: at(2026, time.August, 22), // 1405/05/31
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.True(t, addJalaliYearsToDate(tt.in, 1, utc).Equal(tt.want),
				"got %s, want %s", addJalaliYearsToDate(tt.in, 1, utc), tt.want)
		})
	}
}

func TestUpdateDoneJalaliMonth(t *testing.T) {
	utc := time.UTC

	t.Run("due date steps one jalali month with clamp", func(t *testing.T) {
		oldTask := &Task{
			Done:       false,
			RepeatMode: TaskRepeatModeJalaliMonth,
			DueDate:    time.Date(2025, time.September, 22, 12, 0, 0, 0, utc), // 1404/06/31
		}
		newTask := &Task{Done: true}
		updateDone(oldTask, newTask, utc)

		assert.True(t, newTask.DueDate.Equal(time.Date(2025, time.October, 22, 12, 0, 0, 0, utc)), // 1404/07/30
			"got %s", newTask.DueDate)
		assert.False(t, newTask.Done)
	})
	t.Run("ignores repeat_after", func(t *testing.T) {
		oldTask := &Task{
			Done:        false,
			RepeatAfter: 8600,
			RepeatMode:  TaskRepeatModeJalaliMonth,
			DueDate:     time.Date(2025, time.September, 22, 12, 0, 0, 0, utc),
		}
		newTask := &Task{Done: true}
		updateDone(oldTask, newTask, utc)

		assert.True(t, newTask.DueDate.Equal(time.Date(2025, time.October, 22, 12, 0, 0, 0, utc)),
			"got %s", newTask.DueDate)
		assert.False(t, newTask.Done)
	})
	t.Run("reminders", func(t *testing.T) {
		oldTask := &Task{
			Done:       false,
			RepeatMode: TaskRepeatModeJalaliMonth,
			Reminders: []*TaskReminder{
				{Reminder: time.Date(2025, time.September, 22, 12, 0, 0, 0, utc)},
				{Reminder: time.Date(2026, time.February, 19, 12, 0, 0, 0, utc)}, // 1404/11/30
			},
		}
		newTask := &Task{Done: true}
		updateDone(oldTask, newTask, utc)

		require.Len(t, newTask.Reminders, 2)
		assert.True(t, newTask.Reminders[0].Reminder.Equal(time.Date(2025, time.October, 22, 12, 0, 0, 0, utc)))
		assert.True(t, newTask.Reminders[1].Reminder.Equal(time.Date(2026, time.March, 20, 12, 0, 0, 0, utc)), // 1404/12/29 clamp
			"got %s", newTask.Reminders[1].Reminder)
		assert.False(t, newTask.Done)
	})
	t.Run("start and end date keep duration", func(t *testing.T) {
		oldTask := &Task{
			Done:       false,
			RepeatMode: TaskRepeatModeJalaliMonth,
			StartDate:  time.Date(2025, time.September, 22, 12, 0, 0, 0, utc),
			EndDate:    time.Date(2025, time.September, 25, 12, 0, 0, 0, utc),
		}
		newTask := &Task{Done: true}
		oldDiff := oldTask.EndDate.Sub(oldTask.StartDate)
		updateDone(oldTask, newTask, utc)

		assert.True(t, newTask.StartDate.Equal(time.Date(2025, time.October, 22, 12, 0, 0, 0, utc)))
		assert.Equal(t, oldDiff, newTask.EndDate.Sub(newTask.StartDate))
		assert.False(t, newTask.Done)
	})
	t.Run("esfand wraps to farvardin", func(t *testing.T) {
		oldTask := &Task{
			Done:       false,
			RepeatMode: TaskRepeatModeJalaliMonth,
			DueDate:    time.Date(2026, time.March, 20, 12, 0, 0, 0, utc), // 1404/12/29
		}
		newTask := &Task{Done: true}
		updateDone(oldTask, newTask, utc)

		assert.True(t, newTask.DueDate.Equal(time.Date(2026, time.April, 18, 12, 0, 0, 0, utc)), // 1405/01/29
			"got %s", newTask.DueDate)
		assert.False(t, newTask.Done)
	})
	t.Run("zero due date stays zero", func(t *testing.T) {
		oldTask := &Task{
			Done:       false,
			RepeatMode: TaskRepeatModeJalaliMonth,
		}
		newTask := &Task{Done: true}
		updateDone(oldTask, newTask, utc)

		assert.True(t, newTask.DueDate.IsZero())
		assert.False(t, newTask.Done)
	})
}

func TestUpdateDoneJalaliYear(t *testing.T) {
	utc := time.UTC

	t.Run("due date steps one jalali year", func(t *testing.T) {
		oldTask := &Task{
			Done:       false,
			RepeatMode: TaskRepeatModeJalaliYear,
			DueDate:    time.Date(2025, time.August, 22, 12, 0, 0, 0, utc), // 1404/05/31
		}
		newTask := &Task{Done: true}
		updateDone(oldTask, newTask, utc)

		assert.True(t, newTask.DueDate.Equal(time.Date(2026, time.August, 22, 12, 0, 0, 0, utc)), // 1405/05/31
			"got %s", newTask.DueDate)
		assert.False(t, newTask.Done)
	})
	t.Run("esfand 30 leap clamps to 29", func(t *testing.T) {
		oldTask := &Task{
			Done:       false,
			RepeatMode: TaskRepeatModeJalaliYear,
			DueDate:    time.Date(2025, time.March, 20, 12, 0, 0, 0, utc), // 1403/12/30
		}
		newTask := &Task{Done: true}
		updateDone(oldTask, newTask, utc)

		assert.True(t, newTask.DueDate.Equal(time.Date(2026, time.March, 20, 12, 0, 0, 0, utc)), // 1404/12/29
			"got %s", newTask.DueDate)
		assert.False(t, newTask.Done)
	})
	t.Run("start and end date keep duration", func(t *testing.T) {
		oldTask := &Task{
			Done:       false,
			RepeatMode: TaskRepeatModeJalaliYear,
			StartDate:  time.Date(2025, time.August, 22, 12, 0, 0, 0, utc),
			EndDate:    time.Date(2025, time.August, 29, 12, 0, 0, 0, utc),
		}
		newTask := &Task{Done: true}
		oldDiff := oldTask.EndDate.Sub(oldTask.StartDate)
		updateDone(oldTask, newTask, utc)

		assert.True(t, newTask.StartDate.Equal(time.Date(2026, time.August, 22, 12, 0, 0, 0, utc)))
		assert.Equal(t, oldDiff, newTask.EndDate.Sub(newTask.StartDate))
		assert.False(t, newTask.Done)
	})
}

func TestUpdateDoneJalaliTimezoneAnchoring(t *testing.T) {
	utc := time.UTC
	tehran, err := time.LoadLocation("Asia/Tehran")
	require.NoError(t, err)

	// 2026-03-20 21:00Z is 1404/12/29 in UTC but already Nowruz (1405/01/01) in Tehran.
	due := time.Date(2026, time.March, 20, 21, 0, 0, 0, utc)

	oldTask := &Task{Done: false, RepeatMode: TaskRepeatModeJalaliMonth, DueDate: due}

	tehranTask := &Task{Done: true}
	updateDone(oldTask, tehranTask, tehran)
	assert.True(t, tehranTask.DueDate.Equal(time.Date(2026, time.April, 20, 21, 0, 0, 0, utc)),
		"tehran-anchored got %s", tehranTask.DueDate)
	jy, jm, jd, ok := gregorianToJalali(tehranTask.DueDate.In(tehran).Year(), tehranTask.DueDate.In(tehran).Month(), tehranTask.DueDate.In(tehran).Day())
	require.True(t, ok)
	assert.Equal(t, [3]int{1405, 2, 1}, [3]int{jy, jm, jd})

	utcTask := &Task{Done: true}
	updateDone(oldTask, utcTask, utc)
	assert.True(t, utcTask.DueDate.Equal(time.Date(2026, time.April, 18, 21, 0, 0, 0, utc)),
		"utc-anchored got %s", utcTask.DueDate)
	jy, jm, jd, ok = gregorianToJalali(utcTask.DueDate.In(utc).Year(), utcTask.DueDate.In(utc).Month(), utcTask.DueDate.In(utc).Day())
	require.True(t, ok)
	assert.Equal(t, [3]int{1405, 1, 29}, [3]int{jy, jm, jd})

	assert.False(t, tehranTask.DueDate.Equal(utcTask.DueDate), "anchoring must be observable")
	assert.False(t, tehranTask.Done)
	assert.False(t, utcTask.Done)
}

func TestIsRepeatingJalaliModes(t *testing.T) {
	// isRepeating drives the done-bucket routing in updateSingleTask and
	// updateTaskBucket, so Jalali modes must report repeating here.
	assert.True(t, (&Task{RepeatMode: TaskRepeatModeJalaliMonth}).isRepeating())
	assert.True(t, (&Task{RepeatMode: TaskRepeatModeJalaliYear}).isRepeating())
	assert.True(t, (&Task{RepeatMode: TaskRepeatModeJalaliMonth, RepeatAfter: 0}).isRepeating())
	assert.False(t, (&Task{RepeatMode: TaskRepeatMode(5)}).isRepeating())
	assert.False(t, (&Task{RepeatMode: TaskRepeatModeDefault}).isRepeating())
}

func TestUpdateDoneUnknownRepeatModeStaysDone(t *testing.T) {
	// Unknown modes have no case in updateDone and fall through, staying done.
	dueDate := time.Date(2025, time.September, 22, 12, 0, 0, 0, time.UTC)
	oldTask := &Task{
		Done:       false,
		RepeatMode: TaskRepeatMode(5),
		DueDate:    dueDate,
	}
	newTask := &Task{Done: true, DueDate: dueDate}
	updateDone(oldTask, newTask, nil)

	assert.True(t, newTask.Done)
	assert.True(t, newTask.DueDate.Equal(dueDate))
}
