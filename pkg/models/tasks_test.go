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

	"code.vikunja.io/api/pkg/events"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"
	"github.com/stretchr/testify/assert"
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
		s := db.NewSession()
		defer s.Close()

		task := &Task{
			Title:       "Lorem",
			Description: "Lorem Ipsum Dolor",
			ListID:      1,
		}
		err := task.Create(s, usr)
		assert.NoError(t, err)
		// Assert getting a uid
		assert.NotEmpty(t, task.UID)
		// Assert getting a new index
		assert.NotEmpty(t, task.Index)
		assert.Equal(t, int64(18), task.Index)
		// Assert moving it into the default bucket
		assert.Equal(t, int64(1), task.BucketID)
		err = s.Commit()
		assert.NoError(t, err)

		db.AssertExists(t, "tasks", map[string]interface{}{
			"id":            task.ID,
			"title":         "Lorem",
			"description":   "Lorem Ipsum Dolor",
			"list_id":       1,
			"created_by_id": 1,
			"bucket_id":     1,
		}, false)

		events.AssertDispatched(t, &TaskCreatedEvent{})
	})
	t.Run("empty title", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task := &Task{
			Title:       "",
			Description: "Lorem Ipsum Dolor",
			ListID:      1,
		}
		err := task.Create(s, usr)
		assert.Error(t, err)
		assert.True(t, IsErrTaskCannotBeEmpty(err))
	})
	t.Run("nonexistant list", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task := &Task{
			Title:       "Test",
			Description: "Lorem Ipsum Dolor",
			ListID:      9999999,
		}
		err := task.Create(s, usr)
		assert.Error(t, err)
		assert.True(t, IsErrListDoesNotExist(err))
	})
	t.Run("noneixtant user", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		nUser := &user.User{ID: 99999999}
		task := &Task{
			Title:       "Test",
			Description: "Lorem Ipsum Dolor",
			ListID:      1,
		}
		err := task.Create(s, nUser)
		assert.Error(t, err)
		assert.True(t, user.IsErrUserDoesNotExist(err))
	})
	t.Run("full bucket", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task := &Task{
			Title:       "Lorem",
			Description: "Lorem Ipsum Dolor",
			ListID:      1,
			BucketID:    2, // Bucket 2 already has 3 tasks and a limit of 3
		}
		err := task.Create(s, usr)
		assert.Error(t, err)
		assert.True(t, IsErrBucketLimitExceeded(err))
	})
}

func TestTask_Update(t *testing.T) {
	u := &user.User{ID: 1}

	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task := &Task{
			ID:          1,
			Title:       "test10000",
			Description: "Lorem Ipsum Dolor",
			ListID:      1,
		}
		err := task.Update(s, u)
		assert.NoError(t, err)
		err = s.Commit()
		assert.NoError(t, err)

		db.AssertExists(t, "tasks", map[string]interface{}{
			"id":          1,
			"title":       "test10000",
			"description": "Lorem Ipsum Dolor",
			"list_id":     1,
		}, false)
	})
	t.Run("nonexistant task", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task := &Task{
			ID:          9999999,
			Title:       "test10000",
			Description: "Lorem Ipsum Dolor",
			ListID:      1,
		}
		err := task.Update(s, u)
		assert.Error(t, err)
		assert.True(t, IsErrTaskDoesNotExist(err))
	})
	t.Run("full bucket", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task := &Task{
			ID:          1,
			Title:       "test10000",
			Description: "Lorem Ipsum Dolor",
			ListID:      1,
			BucketID:    2, // Bucket 2 already has 3 tasks and a limit of 3
		}
		err := task.Update(s, u)
		assert.Error(t, err)
		assert.True(t, IsErrBucketLimitExceeded(err))
	})
	t.Run("full bucket but not changing the bucket", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task := &Task{
			ID:          4,
			Title:       "test10000",
			Description: "Lorem Ipsum Dolor",
			Position:    10,
			ListID:      1,
			BucketID:    2, // Bucket 2 already has 3 tasks and a limit of 3
		}
		err := task.Update(s, u)
		assert.NoError(t, err)
	})
	t.Run("bucket on other list", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task := &Task{
			ID:          1,
			Title:       "test10000",
			Description: "Lorem Ipsum Dolor",
			ListID:      1,
			BucketID:    4, // Bucket 4 belongs to list 2
		}
		err := task.Update(s, u)
		assert.Error(t, err)
		assert.True(t, IsErrBucketDoesNotBelongToList(err))
	})
	t.Run("moving a task to the done bucket", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task := &Task{
			ID:       1,
			Title:    "test",
			ListID:   1,
			BucketID: 3, // Bucket 3 is the done bucket
		}
		err := task.Update(s, u)
		assert.NoError(t, err)
		err = s.Commit()
		assert.NoError(t, err)
		assert.True(t, task.Done)

		db.AssertExists(t, "tasks", map[string]interface{}{
			"id":        1,
			"done":      true,
			"title":     "test",
			"list_id":   1,
			"bucket_id": 3,
		}, false)
	})
	t.Run("default bucket when moving a task between lists", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task := &Task{
			ID:     1,
			ListID: 2,
		}
		err := task.Update(s, u)
		assert.NoError(t, err)
		err = s.Commit()
		assert.NoError(t, err)

		assert.Equal(t, int64(4), task.BucketID) // bucket 4 is the default bucket on list 2
		assert.True(t, task.Done)                // bucket 4 is the done bucket, so the task should be marked as done as well
	})
	t.Run("marking a task as done should move it to the done bucket", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task := &Task{
			ID:   1,
			Done: true,
		}
		err := task.Update(s, u)
		assert.NoError(t, err)
		err = s.Commit()
		assert.NoError(t, err)
		assert.True(t, task.Done)
		assert.Equal(t, int64(3), task.BucketID)

		db.AssertExists(t, "tasks", map[string]interface{}{
			"id":        1,
			"done":      true,
			"bucket_id": 3,
		}, false)
	})
	t.Run("move task to another list", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task := &Task{
			ID:     1,
			ListID: 2,
		}
		err := task.Update(s, u)
		assert.NoError(t, err)
		err = s.Commit()
		assert.NoError(t, err)

		db.AssertExists(t, "tasks", map[string]interface{}{
			"id":        1,
			"list_id":   2,
			"bucket_id": 4,
		}, false)
	})
	t.Run("repeating tasks should not be moved to the done bucket", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task := &Task{
			ID:   28,
			Done: true,
		}
		err := task.Update(s, u)
		assert.NoError(t, err)
		err = s.Commit()
		assert.NoError(t, err)
		assert.False(t, task.Done)
		assert.Equal(t, int64(1), task.BucketID)

		db.AssertExists(t, "tasks", map[string]interface{}{
			"id":        28,
			"done":      false,
			"bucket_id": 1,
		}, false)
	})
}

func TestTask_Delete(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task := &Task{
			ID: 1,
		}
		err := task.Delete(s, &user.User{ID: 1})
		assert.NoError(t, err)
		err = s.Commit()
		assert.NoError(t, err)

		db.AssertMissing(t, "tasks", map[string]interface{}{
			"id": 1,
		})
	})
}

func TestUpdateDone(t *testing.T) {
	t.Run("marking a task as done", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		oldTask := &Task{Done: false}
		newTask := &Task{Done: true}
		updateDone(oldTask, newTask)
		assert.NotEqual(t, time.Time{}, newTask.DoneAt)
	})
	t.Run("unmarking a task as done", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		oldTask := &Task{Done: true}
		newTask := &Task{Done: false}
		updateDone(oldTask, newTask)
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
		updateDone(oldTask, newTask)

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
			updateDone(oldTask, newTask)

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
			updateDone(oldTask, newTask)
			assert.Equal(t, time.Unix(1543626724, 0), newTask.DueDate)
			assert.False(t, newTask.Done)
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
			updateDone(oldTask, newTask)

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
			updateDone(oldTask, newTask)

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
			updateDone(oldTask, newTask)
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
				updateDone(oldTask, newTask)

				// Only comparing unix timestamps because time.Time use nanoseconds which can't ever possibly have the same value
				assert.Equal(t, time.Now().Add(time.Duration(oldTask.RepeatAfter)*time.Second).Unix(), newTask.DueDate.Unix())
				assert.False(t, newTask.Done)
			})
			t.Run("reminders", func(t *testing.T) {
				oldTask := &Task{
					Done:        false,
					RepeatAfter: 8600,
					RepeatMode:  TaskRepeatModeFromCurrentDate,
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
				updateDone(oldTask, newTask)

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
				updateDone(oldTask, newTask)

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
				updateDone(oldTask, newTask)

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

				updateDone(oldTask, newTask)

				assert.True(t, newTask.DueDate.After(oldDueDate))
				assert.NotEqual(t, oldDueDate.Month(), newTask.DueDate.Month())
				assert.False(t, newTask.Done)
			})
			t.Run("reminders", func(t *testing.T) {
				oldTask := &Task{
					Done:       false,
					RepeatMode: TaskRepeatModeMonth,
					Reminders: []time.Time{
						time.Unix(1550000000, 0),
						time.Unix(1555000000, 0),
					},
				}
				newTask := &Task{
					Done: true,
				}
				oldReminders := make([]time.Time, len(oldTask.Reminders))
				copy(oldReminders, oldTask.Reminders)

				updateDone(oldTask, newTask)

				assert.Len(t, newTask.Reminders, len(oldReminders))
				for i, r := range newTask.Reminders {
					assert.True(t, r.After(oldReminders[i]))
					assert.NotEqual(t, oldReminders[i].Month(), r.Month())
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

				updateDone(oldTask, newTask)

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

				updateDone(oldTask, newTask)

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

				updateDone(oldTask, newTask)

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

func TestTask_ReadOne(t *testing.T) {
	u := &user.User{ID: 1}

	t.Run("default", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task := &Task{ID: 1}
		err := task.ReadOne(s, u)
		assert.NoError(t, err)
		assert.Equal(t, "task #1", task.Title)
	})
	t.Run("nonexisting", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task := &Task{ID: 99999}
		err := task.ReadOne(s, u)
		assert.Error(t, err)
		assert.True(t, IsErrTaskDoesNotExist(err))
	})
	t.Run("with subscription", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task := &Task{ID: 22}
		err := task.ReadOne(s, &user.User{ID: 6})
		assert.NoError(t, err)
		assert.NotNil(t, task.Subscription)
	})
	t.Run("created by link share", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task := &Task{ID: 37}
		err := task.ReadOne(s, u)
		assert.NoError(t, err)
		assert.Equal(t, "task #37", task.Title)
		assert.Equal(t, int64(-2), task.CreatedByID)
		assert.NotNil(t, task.CreatedBy)
		assert.Equal(t, int64(-2), task.CreatedBy.ID)
	})
	t.Run("favorite", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task := &Task{ID: 1}
		err := task.ReadOne(s, u)
		assert.NoError(t, err)
		assert.True(t, task.IsFavorite)
	})
	t.Run("favorite for a different user", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task := &Task{ID: 1}
		err := task.ReadOne(s, &user.User{ID: 2})
		assert.NoError(t, err)
		assert.False(t, task.IsFavorite)
	})
}
