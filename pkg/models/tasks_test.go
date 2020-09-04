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
	t.Run("full bucket", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		task := &Task{
			Title:       "Lorem",
			Description: "Lorem Ipsum Dolor",
			ListID:      1,
			BucketID:    2, // Bucket 2 already has 3 tasks and a limit of 3
		}
		err := task.Create(usr)
		assert.Error(t, err)
		assert.True(t, IsErrBucketLimitExceeded(err))
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
	t.Run("full bucket", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		task := &Task{
			ID:          1,
			Title:       "test10000",
			Description: "Lorem Ipsum Dolor",
			ListID:      1,
			BucketID:    2, // Bucket 2 already has 3 tasks and a limit of 3
		}
		err := task.Update()
		assert.Error(t, err)
		assert.True(t, IsErrBucketLimitExceeded(err))
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
		assert.NotEqual(t, time.Time{}, newTask.DoneAt)
	})
	t.Run("unmarking a task as done", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		oldTask := &Task{Done: true}
		newTask := &Task{Done: false}
		updateDone(oldTask, newTask)
		assert.Equal(t, time.Time{}, newTask.DoneAt)
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
			updateDone(oldTask, newTask)

			var expected = time.Unix(1550008600, 0)
			for time.Since(expected) > 0 {
				expected = expected.Add(time.Second * time.Duration(oldTask.RepeatAfter))
			}

			assert.Equal(t, expected, newTask.DueDate)
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
			updateDone(oldTask, newTask)
			assert.Equal(t, time.Unix(1543626724, 0), newTask.DueDate)
		})
		t.Run("update reminders", func(t *testing.T) {
			oldTask := &Task{
				Done:        false,
				RepeatAfter: 8600,
				Reminders: []time.Time{
					time.Unix(1550000000, 0),
					time.Unix(1555000000, 0),
				},
			}
			newTask := &Task{
				Done: true,
			}
			updateDone(oldTask, newTask)

			var expected1 = time.Unix(1550008600, 0)
			var expected2 = time.Unix(1555008600, 0)
			for time.Since(expected1) > 0 {
				expected1 = expected1.Add(time.Duration(oldTask.RepeatAfter) * time.Second)
			}
			for time.Since(expected2) > 0 {
				expected2 = expected2.Add(time.Duration(oldTask.RepeatAfter) * time.Second)
			}

			assert.Len(t, newTask.Reminders, 2)
			assert.Equal(t, expected1, newTask.Reminders[0])
			assert.Equal(t, expected2, newTask.Reminders[1])
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
			updateDone(oldTask, newTask)

			var expected = time.Unix(1550008600, 0)
			for time.Since(expected) > 0 {
				expected = expected.Add(time.Second * time.Duration(oldTask.RepeatAfter))
			}

			assert.Equal(t, expected, newTask.StartDate)
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
			updateDone(oldTask, newTask)

			var expected = time.Unix(1550008600, 0)
			for time.Since(expected) > 0 {
				expected = expected.Add(time.Second * time.Duration(oldTask.RepeatAfter))
			}

			assert.Equal(t, expected, newTask.EndDate)
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
			updateDone(oldTask, newTask)
			expected := oldTask.DueDate.Add(time.Duration(oldTask.RepeatAfter) * time.Second)
			assert.Equal(t, expected, newTask.DueDate)
		})
		t.Run("repeat from current date", func(t *testing.T) {
			t.Run("due date", func(t *testing.T) {
				oldTask := &Task{
					Done:                  false,
					RepeatAfter:           8600,
					RepeatFromCurrentDate: true,
					DueDate:               time.Unix(1550000000, 0),
				}
				newTask := &Task{
					Done: true,
				}
				updateDone(oldTask, newTask)

				// Only comparing unix timestamps because time.Time use nanoseconds which can't ever possibly have the same value
				assert.Equal(t, time.Now().Add(time.Duration(oldTask.RepeatAfter)*time.Second).Unix(), newTask.DueDate.Unix())
			})
			t.Run("reminders", func(t *testing.T) {
				oldTask := &Task{
					Done:                  false,
					RepeatAfter:           8600,
					RepeatFromCurrentDate: true,
					Reminders: []time.Time{
						time.Unix(1550000000, 0),
						time.Unix(1555000000, 0),
					},
				}
				newTask := &Task{
					Done: true,
				}
				updateDone(oldTask, newTask)

				diff := oldTask.Reminders[1].Sub(oldTask.Reminders[0])

				assert.Len(t, newTask.Reminders, 2)
				// Only comparing unix timestamps because time.Time use nanoseconds which can't ever possibly have the same value
				assert.Equal(t, time.Now().Add(time.Duration(oldTask.RepeatAfter)*time.Second).Unix(), newTask.Reminders[0].Unix())
				assert.Equal(t, time.Now().Add(diff+time.Duration(oldTask.RepeatAfter)*time.Second).Unix(), newTask.Reminders[1].Unix())
			})
			t.Run("start date", func(t *testing.T) {
				oldTask := &Task{
					Done:                  false,
					RepeatAfter:           8600,
					RepeatFromCurrentDate: true,
					StartDate:             time.Unix(1550000000, 0),
				}
				newTask := &Task{
					Done: true,
				}
				updateDone(oldTask, newTask)

				// Only comparing unix timestamps because time.Time use nanoseconds which can't ever possibly have the same value
				assert.Equal(t, time.Now().Add(time.Duration(oldTask.RepeatAfter)*time.Second).Unix(), newTask.StartDate.Unix())
			})
			t.Run("end date", func(t *testing.T) {
				oldTask := &Task{
					Done:                  false,
					RepeatAfter:           8600,
					RepeatFromCurrentDate: true,
					EndDate:               time.Unix(1560000000, 0),
				}
				newTask := &Task{
					Done: true,
				}
				updateDone(oldTask, newTask)

				// Only comparing unix timestamps because time.Time use nanoseconds which can't ever possibly have the same value
				assert.Equal(t, time.Now().Add(time.Duration(oldTask.RepeatAfter)*time.Second).Unix(), newTask.EndDate.Unix())
			})
			t.Run("start and end date", func(t *testing.T) {
				oldTask := &Task{
					Done:                  false,
					RepeatAfter:           8600,
					RepeatFromCurrentDate: true,
					StartDate:             time.Unix(1550000000, 0),
					EndDate:               time.Unix(1560000000, 0),
				}
				newTask := &Task{
					Done: true,
				}
				updateDone(oldTask, newTask)

				diff := oldTask.EndDate.Sub(oldTask.StartDate)

				// Only comparing unix timestamps because time.Time use nanoseconds which can't ever possibly have the same value
				assert.Equal(t, time.Now().Add(time.Duration(oldTask.RepeatAfter)*time.Second).Unix(), newTask.StartDate.Unix())
				assert.Equal(t, time.Now().Add(diff+time.Duration(oldTask.RepeatAfter)*time.Second).Unix(), newTask.EndDate.Unix())
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
