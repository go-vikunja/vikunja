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

	"github.com/stretchr/testify/assert"
)

// ============================================================================
// NOTE: Business logic tests for Task Create/Update/Delete have been migrated
// to the service layer in pkg/services/task_business_logic_test.go as part of
// T015B. The model layer now only contains thin delegators to the service layer.
//
// Only utility function tests remain here (TestUpdateDone).
// ============================================================================

func TestUpdateDone(t *testing.T) {
	t.Run("marking a task as done", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		oldTask := &Task{Done: false}
		newTask := &Task{Done: true}
		UpdateDone(oldTask, newTask)
		assert.NotEqual(t, time.Time{}, newTask.DoneAt)
	})
	t.Run("unmarking a task as done", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		oldTask := &Task{Done: true}
		newTask := &Task{Done: false}
		UpdateDone(oldTask, newTask)
		assert.Equal(t, time.Time{}, newTask.DoneAt)
	})
	t.Run("no interval set, default repeat mode", func(t *testing.T) {
		dueDate := time.Unix(1550000000, 0)
		oldTask := &Task{
			Done:        false,
			RepeatAfter: 0,
			RepeatMode:  TaskRepeatModeDefault,
			DueDate:     dueDate,
		}
		newTask := &Task{
			Done:    true,
			DueDate: dueDate,
		}
		UpdateDone(oldTask, newTask)

		assert.Equal(t, dueDate.Unix(), newTask.DueDate.Unix())
		assert.True(t, newTask.Done)
	})
	t.Run("repeating interval", func(t *testing.T) {
		t.Run("normal", func(t *testing.T) {
			oldTask := &Task{
				Done:        false,
				RepeatAfter: 8600,
				DueDate:     time.Unix(1550000000, 0),
			}
			newTask := &Task{
				Done: true,
			}
			UpdateDone(oldTask, newTask)

			var expected = time.Unix(1550008600, 0)
			for time.Since(expected) > 0 {
				expected = expected.Add(time.Second * time.Duration(oldTask.RepeatAfter))
			}

			assert.Equal(t, expected, newTask.DueDate)
			assert.False(t, newTask.Done)
		})
		t.Run("don't update if due date is zero", func(t *testing.T) {
			oldTask := &Task{
				Done:        false,
				RepeatAfter: 8600,
				DueDate:     time.Time{},
			}
			newTask := &Task{
				Done:    true,
				DueDate: time.Unix(1543626724, 0),
			}
			UpdateDone(oldTask, newTask)
			assert.Equal(t, time.Unix(1543626724, 0), newTask.DueDate)
			assert.False(t, newTask.Done)
		})
		t.Run("update reminders", func(t *testing.T) {
			oldTask := &Task{
				Done:        false,
				RepeatAfter: 8600,
				Reminders: []*TaskReminder{
					{
						Reminder: time.Unix(1550000000, 0),
					},
					{
						Reminder: time.Unix(1555000000, 0),
					},
				},
			}
			newTask := &Task{
				Done: true,
			}
			UpdateDone(oldTask, newTask)

			var expected1 = time.Unix(1550008600, 0)
			var expected2 = time.Unix(1555008600, 0)
			for time.Since(expected1) > 0 {
				expected1 = expected1.Add(time.Duration(oldTask.RepeatAfter) * time.Second)
			}
			for time.Since(expected2) > 0 {
				expected2 = expected2.Add(time.Duration(oldTask.RepeatAfter) * time.Second)
			}

			assert.Len(t, newTask.Reminders, 2)
			assert.Equal(t, expected1, newTask.Reminders[0].Reminder)
			assert.Equal(t, expected2, newTask.Reminders[1].Reminder)
			assert.False(t, newTask.Done)
		})
		t.Run("update start date", func(t *testing.T) {
			oldTask := &Task{
				Done:        false,
				RepeatAfter: 8600,
				StartDate:   time.Unix(1550000000, 0),
			}
			newTask := &Task{
				Done: true,
			}
			UpdateDone(oldTask, newTask)

			var expected = time.Unix(1550008600, 0)
			for time.Since(expected) > 0 {
				expected = expected.Add(time.Second * time.Duration(oldTask.RepeatAfter))
			}

			assert.Equal(t, expected, newTask.StartDate)
			assert.False(t, newTask.Done)
		})
		t.Run("update end date", func(t *testing.T) {
			oldTask := &Task{
				Done:        false,
				RepeatAfter: 8600,
				EndDate:     time.Unix(1550000000, 0),
			}
			newTask := &Task{
				Done: true,
			}
			UpdateDone(oldTask, newTask)

			var expected = time.Unix(1550008600, 0)
			for time.Since(expected) > 0 {
				expected = expected.Add(time.Second * time.Duration(oldTask.RepeatAfter))
			}

			assert.Equal(t, expected, newTask.EndDate)
			assert.False(t, newTask.Done)
		})
		t.Run("ensure due date is repeated even if the original one is in the future", func(t *testing.T) {
			oldTask := &Task{
				Done:        false,
				RepeatAfter: 8600,
				DueDate:     time.Now().Add(time.Hour),
			}
			newTask := &Task{
				Done: true,
			}
			UpdateDone(oldTask, newTask)
			expected := oldTask.DueDate.Add(time.Duration(oldTask.RepeatAfter) * time.Second)
			assert.Equal(t, expected, newTask.DueDate)
			assert.False(t, newTask.Done)
		})
		t.Run("repeat from current date", func(t *testing.T) {
			t.Run("due date", func(t *testing.T) {
				oldTask := &Task{
					Done:        false,
					RepeatAfter: 8600,
					RepeatMode:  TaskRepeatModeFromCurrentDate,
					DueDate:     time.Unix(1550000000, 0),
				}
				newTask := &Task{
					Done: true,
				}
				UpdateDone(oldTask, newTask)

				// Only comparing unix timestamps because time.Time use nanoseconds which can't ever possibly have the same value
				assert.Equal(t, time.Now().Add(time.Duration(oldTask.RepeatAfter)*time.Second).Unix(), newTask.DueDate.Unix())
				assert.False(t, newTask.Done)
			})
			t.Run("reminders", func(t *testing.T) {
				oldTask := &Task{
					Done:        false,
					RepeatAfter: 8600,
					RepeatMode:  TaskRepeatModeFromCurrentDate,
					Reminders: []*TaskReminder{
						{
							Reminder: time.Unix(1550000000, 0),
						},
						{
							Reminder: time.Unix(1555000000, 0),
						},
					}}
				newTask := &Task{
					Done: true,
				}
				UpdateDone(oldTask, newTask)

				diff := oldTask.Reminders[1].Reminder.Sub(oldTask.Reminders[0].Reminder)

				assert.Len(t, newTask.Reminders, 2)
				// Only comparing unix timestamps because time.Time use nanoseconds which can't ever possibly have the same value
				assert.Equal(t, time.Now().Add(time.Duration(oldTask.RepeatAfter)*time.Second).Unix(), newTask.Reminders[0].Reminder.Unix())
				assert.Equal(t, time.Now().Add(diff+time.Duration(oldTask.RepeatAfter)*time.Second).Unix(), newTask.Reminders[1].Reminder.Unix())
				assert.False(t, newTask.Done)
			})
			t.Run("start date", func(t *testing.T) {
				oldTask := &Task{
					Done:        false,
					RepeatAfter: 8600,
					RepeatMode:  TaskRepeatModeFromCurrentDate,
					StartDate:   time.Unix(1550000000, 0),
				}
				newTask := &Task{
					Done: true,
				}
				UpdateDone(oldTask, newTask)

				// Only comparing unix timestamps because time.Time use nanoseconds which can't ever possibly have the same value
				assert.Equal(t, time.Now().Add(time.Duration(oldTask.RepeatAfter)*time.Second).Unix(), newTask.StartDate.Unix())
				assert.False(t, newTask.Done)
			})
			t.Run("end date", func(t *testing.T) {
				oldTask := &Task{
					Done:        false,
					RepeatAfter: 8600,
					RepeatMode:  TaskRepeatModeFromCurrentDate,
					EndDate:     time.Unix(1560000000, 0),
				}
				newTask := &Task{
					Done: true,
				}
				UpdateDone(oldTask, newTask)

				// Only comparing unix timestamps because time.Time use nanoseconds which can't ever possibly have the same value
				assert.Equal(t, time.Now().Add(time.Duration(oldTask.RepeatAfter)*time.Second).Unix(), newTask.EndDate.Unix())
				assert.False(t, newTask.Done)
			})
			t.Run("start and end date", func(t *testing.T) {
				oldTask := &Task{
					Done:        false,
					RepeatAfter: 8600,
					RepeatMode:  TaskRepeatModeFromCurrentDate,
					StartDate:   time.Unix(1550000000, 0),
					EndDate:     time.Unix(1560000000, 0),
				}
				newTask := &Task{
					Done: true,
				}
				UpdateDone(oldTask, newTask)

				diff := oldTask.EndDate.Sub(oldTask.StartDate)

				// Only comparing unix timestamps because time.Time use nanoseconds which can't ever possibly have the same value
				assert.Equal(t, time.Now().Add(time.Duration(oldTask.RepeatAfter)*time.Second).Unix(), newTask.StartDate.Unix())
				assert.Equal(t, time.Now().Add(diff+time.Duration(oldTask.RepeatAfter)*time.Second).Unix(), newTask.EndDate.Unix())
				assert.False(t, newTask.Done)
			})
		})
		t.Run("repeat each month", func(t *testing.T) {
			t.Run("due date", func(t *testing.T) {
				oldTask := &Task{
					Done:       false,
					RepeatMode: TaskRepeatModeMonth,
					DueDate:    time.Unix(1550000000, 0),
				}
				newTask := &Task{
					Done: true,
				}
				oldDueDate := oldTask.DueDate

				UpdateDone(oldTask, newTask)

				assert.True(t, newTask.DueDate.After(oldDueDate))
				assert.NotEqual(t, oldDueDate.Month(), newTask.DueDate.Month())
				assert.False(t, newTask.Done)
			})
			t.Run("reminders", func(t *testing.T) {
				oldTask := &Task{
					Done:       false,
					RepeatMode: TaskRepeatModeMonth,
					Reminders: []*TaskReminder{
						{
							Reminder: time.Unix(1550000000, 0),
						},
						{
							Reminder: time.Unix(1555000000, 0),
						},
					}}
				newTask := &Task{
					Done: true,
				}
				oldReminders := make([]time.Time, len(oldTask.Reminders))
				for i, r := range newTask.Reminders {
					oldReminders[i] = r.Reminder
				}

				UpdateDone(oldTask, newTask)

				assert.Len(t, newTask.Reminders, len(oldReminders))
				for i, r := range newTask.Reminders {
					assert.True(t, r.Reminder.After(oldReminders[i]))
					assert.NotEqual(t, oldReminders[i].Month(), r.Reminder.Month())
				}
				assert.False(t, newTask.Done)
			})
			t.Run("start date", func(t *testing.T) {
				oldTask := &Task{
					Done:       false,
					RepeatMode: TaskRepeatModeMonth,
					StartDate:  time.Unix(1550000000, 0),
				}
				newTask := &Task{
					Done: true,
				}
				oldStartDate := oldTask.StartDate

				UpdateDone(oldTask, newTask)

				assert.True(t, newTask.StartDate.After(oldStartDate))
				assert.NotEqual(t, oldStartDate.Month(), newTask.StartDate.Month())
				assert.False(t, newTask.Done)
			})
			t.Run("end date", func(t *testing.T) {
				oldTask := &Task{
					Done:       false,
					RepeatMode: TaskRepeatModeMonth,
					EndDate:    time.Unix(1560000000, 0),
				}
				newTask := &Task{
					Done: true,
				}
				oldEndDate := oldTask.EndDate

				UpdateDone(oldTask, newTask)

				assert.True(t, newTask.EndDate.After(oldEndDate))
				assert.NotEqual(t, oldEndDate.Month(), newTask.EndDate.Month())
				assert.False(t, newTask.Done)
			})
			t.Run("start and end date", func(t *testing.T) {
				oldTask := &Task{
					Done:       false,
					RepeatMode: TaskRepeatModeMonth,
					StartDate:  time.Unix(1550000000, 0),
					EndDate:    time.Unix(1560000000, 0),
				}
				newTask := &Task{
					Done: true,
				}
				oldStartDate := oldTask.StartDate
				oldEndDate := oldTask.EndDate
				oldDiff := oldTask.EndDate.Sub(oldTask.StartDate)

				UpdateDone(oldTask, newTask)

				assert.True(t, newTask.StartDate.After(oldStartDate))
				assert.NotEqual(t, oldStartDate.Month(), newTask.StartDate.Month())
				assert.True(t, newTask.EndDate.After(oldEndDate))
				assert.NotEqual(t, oldEndDate.Month(), newTask.EndDate.Month())
				assert.Equal(t, oldDiff, newTask.EndDate.Sub(newTask.StartDate))
				assert.False(t, newTask.Done)
			})
		})
	})
}
