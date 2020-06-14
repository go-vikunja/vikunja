// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2020 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/timeutil"
	"code.vikunja.io/api/pkg/user"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestTask_Create(t *testing.T) {
	usr := &user.User{
		ID:       1,
		Username: "user1",
		Email:    "user1@example.com",
	}

	// We only test creating a task here, the rights are all well tested in the integration tests.

	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		task := &Task{
			Title:       "Lorem",
			Description: "Lorem Ipsum Dolor",
			ListID:      1,
		}
		err := task.Create(usr)
		assert.NoError(t, err)
		// Assert getting a uid
		assert.NotEmpty(t, task.UID)
		// Assert getting a new index
		assert.NotEmpty(t, task.Index)
		assert.Equal(t, int64(18), task.Index)
		// Assert moving it into the default bucket
		assert.Equal(t, int64(1), task.BucketID)

	})
	t.Run("empty title", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		task := &Task{
			Title:       "",
			Description: "Lorem Ipsum Dolor",
			ListID:      1,
		}
		err := task.Create(usr)
		assert.Error(t, err)
		assert.True(t, IsErrTaskCannotBeEmpty(err))
	})
	t.Run("nonexistant list", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		task := &Task{
			Title:       "Test",
			Description: "Lorem Ipsum Dolor",
			ListID:      9999999,
		}
		err := task.Create(usr)
		assert.Error(t, err)
		assert.True(t, IsErrListDoesNotExist(err))
	})
	t.Run("noneixtant user", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		nUser := &user.User{ID: 99999999}
		task := &Task{
			Title:       "Test",
			Description: "Lorem Ipsum Dolor",
			ListID:      1,
		}
		err := task.Create(nUser)
		assert.Error(t, err)
		assert.True(t, user.IsErrUserDoesNotExist(err))
	})
}

func TestTask_Update(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		task := &Task{
			ID:          1,
			Title:       "test10000",
			Description: "Lorem Ipsum Dolor",
			ListID:      1,
		}
		err := task.Update()
		assert.NoError(t, err)
	})
	t.Run("nonexistant task", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		task := &Task{
			ID:          9999999,
			Title:       "test10000",
			Description: "Lorem Ipsum Dolor",
			ListID:      1,
		}
		err := task.Update()
		assert.Error(t, err)
		assert.True(t, IsErrTaskDoesNotExist(err))
	})
}

func TestTask_Delete(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		task := &Task{
			ID: 1,
		}
		err := task.Delete()
		assert.NoError(t, err)
	})
}

func TestUpdateDone(t *testing.T) {
	t.Run("marking a task as done", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		oldTask := &Task{Done: false}
		newTask := &Task{Done: true}
		updateDone(oldTask, newTask)
		assert.NotEqual(t, timeutil.TimeStamp(0), newTask.DoneAt)
	})
	t.Run("unmarking a task as done", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		oldTask := &Task{Done: true}
		newTask := &Task{Done: false}
		updateDone(oldTask, newTask)
		assert.Equal(t, timeutil.TimeStamp(0), newTask.DoneAt)
	})
	t.Run("repeating interval", func(t *testing.T) {
		t.Run("normal", func(t *testing.T) {
			oldTask := &Task{
				Done:        false,
				RepeatAfter: 8600,
				DueDate:     timeutil.TimeStamp(1550000000),
			}
			newTask := &Task{
				Done: true,
			}
			updateDone(oldTask, newTask)

			var expected int64 = 1550008600
			for expected < time.Now().Unix() {
				expected += oldTask.RepeatAfter
			}

			assert.Equal(t, timeutil.TimeStamp(expected), newTask.DueDate)
		})
		t.Run("don't update if due date is zero", func(t *testing.T) {
			oldTask := &Task{
				Done:        false,
				RepeatAfter: 8600,
				DueDate:     timeutil.TimeStamp(0),
			}
			newTask := &Task{
				Done:    true,
				DueDate: timeutil.TimeStamp(1543626724),
			}
			updateDone(oldTask, newTask)
			assert.Equal(t, timeutil.TimeStamp(1543626724), newTask.DueDate)
		})
		t.Run("update reminders", func(t *testing.T) {
			oldTask := &Task{
				Done:        false,
				RepeatAfter: 8600,
				Reminders: []timeutil.TimeStamp{
					1550000000,
					1555000000,
				},
			}
			newTask := &Task{
				Done: true,
			}
			updateDone(oldTask, newTask)

			var expected1 int64 = 1550008600
			var expected2 int64 = 1555008600
			for expected1 < time.Now().Unix() {
				expected1 += oldTask.RepeatAfter
			}
			for expected2 < time.Now().Unix() {
				expected2 += oldTask.RepeatAfter
			}

			assert.Len(t, newTask.Reminders, 2)
			assert.Equal(t, timeutil.TimeStamp(expected1), newTask.Reminders[0])
			assert.Equal(t, timeutil.TimeStamp(expected2), newTask.Reminders[1])
		})
		t.Run("update start date", func(t *testing.T) {
			oldTask := &Task{
				Done:        false,
				RepeatAfter: 8600,
				StartDate:   timeutil.TimeStamp(1550000000),
			}
			newTask := &Task{
				Done: true,
			}
			updateDone(oldTask, newTask)

			var expected int64 = 1550008600
			for expected < time.Now().Unix() {
				expected += oldTask.RepeatAfter
			}

			assert.Equal(t, timeutil.TimeStamp(expected), newTask.StartDate)
		})
		t.Run("update end date", func(t *testing.T) {
			oldTask := &Task{
				Done:        false,
				RepeatAfter: 8600,
				EndDate:     timeutil.TimeStamp(1550000000),
			}
			newTask := &Task{
				Done: true,
			}
			updateDone(oldTask, newTask)

			var expected int64 = 1550008600
			for expected < time.Now().Unix() {
				expected += oldTask.RepeatAfter
			}

			assert.Equal(t, timeutil.TimeStamp(expected), newTask.EndDate)
		})
		t.Run("ensure due date is repeated even if the original one is in the future", func(t *testing.T) {
			oldTask := &Task{
				Done:        false,
				RepeatAfter: 8600,
				DueDate:     timeutil.FromTime(time.Now().Add(time.Hour)),
			}
			newTask := &Task{
				Done: true,
			}
			updateDone(oldTask, newTask)
			expected := int64(oldTask.DueDate) + oldTask.RepeatAfter
			assert.Equal(t, timeutil.TimeStamp(expected), newTask.DueDate)
		})
		t.Run("repeat from current date", func(t *testing.T) {
			t.Run("due date", func(t *testing.T) {
				oldTask := &Task{
					Done:                  false,
					RepeatAfter:           8600,
					RepeatFromCurrentDate: true,
					DueDate:               timeutil.TimeStamp(1550000000),
				}
				newTask := &Task{
					Done: true,
				}
				updateDone(oldTask, newTask)

				assert.Equal(t, timeutil.FromTime(time.Now().Add(time.Duration(oldTask.RepeatAfter)*time.Second)), newTask.DueDate)
			})
			t.Run("reminders", func(t *testing.T) {
				oldTask := &Task{
					Done:                  false,
					RepeatAfter:           8600,
					RepeatFromCurrentDate: true,
					Reminders: []timeutil.TimeStamp{
						1550000000,
						1555000000,
					},
				}
				newTask := &Task{
					Done: true,
				}
				updateDone(oldTask, newTask)

				diff := time.Duration(oldTask.Reminders[1]-oldTask.Reminders[0]) * time.Second

				assert.Len(t, newTask.Reminders, 2)
				assert.Equal(t, timeutil.FromTime(time.Now().Add(time.Duration(oldTask.RepeatAfter)*time.Second)), newTask.Reminders[0])
				assert.Equal(t, timeutil.FromTime(time.Now().Add(diff+time.Duration(oldTask.RepeatAfter)*time.Second)), newTask.Reminders[1])
			})
			t.Run("start date", func(t *testing.T) {
				oldTask := &Task{
					Done:                  false,
					RepeatAfter:           8600,
					RepeatFromCurrentDate: true,
					StartDate:             timeutil.TimeStamp(1550000000),
				}
				newTask := &Task{
					Done: true,
				}
				updateDone(oldTask, newTask)

				assert.Equal(t, timeutil.FromTime(time.Now().Add(time.Duration(oldTask.RepeatAfter)*time.Second)), newTask.StartDate)
			})
			t.Run("end date", func(t *testing.T) {
				oldTask := &Task{
					Done:                  false,
					RepeatAfter:           8600,
					RepeatFromCurrentDate: true,
					EndDate:               timeutil.TimeStamp(1560000000),
				}
				newTask := &Task{
					Done: true,
				}
				updateDone(oldTask, newTask)

				assert.Equal(t, timeutil.FromTime(time.Now().Add(time.Duration(oldTask.RepeatAfter)*time.Second)), newTask.EndDate)
			})
			t.Run("start and end date", func(t *testing.T) {
				oldTask := &Task{
					Done:                  false,
					RepeatAfter:           8600,
					RepeatFromCurrentDate: true,
					StartDate:             timeutil.TimeStamp(1550000000),
					EndDate:               timeutil.TimeStamp(1560000000),
				}
				newTask := &Task{
					Done: true,
				}
				updateDone(oldTask, newTask)

				diff := time.Duration(oldTask.EndDate-oldTask.StartDate) * time.Second

				assert.Equal(t, timeutil.FromTime(time.Now().Add(time.Duration(oldTask.RepeatAfter)*time.Second)), newTask.StartDate)
				assert.Equal(t, timeutil.FromTime(time.Now().Add(diff+time.Duration(oldTask.RepeatAfter)*time.Second)), newTask.EndDate)
			})
		})
	})
}

func TestTask_ReadOne(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		task := &Task{ID: 1}
		err := task.ReadOne()
		assert.NoError(t, err)
		assert.Equal(t, "task #1", task.Title)
	})
	t.Run("nonexisting", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		task := &Task{ID: 99999}
		err := task.ReadOne()
		assert.Error(t, err)
		assert.True(t, IsErrTaskDoesNotExist(err))
	})
}
