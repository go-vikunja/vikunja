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
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/builder"
)

func TestTask_Create(t *testing.T) {
	usr := &user.User{
		ID:       1,
		Username: "user1",
		Email:    "user1@example.com",
	}

	// We only test creating a task here, the permissions are all well tested in the web tests.

	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task := &Task{
			Title:       "Lorem",
			Description: "Lorem Ipsum Dolor",
			ProjectID:   1,
		}
		err := task.Create(s, usr)
		require.NoError(t, err)
		// Assert getting a uid
		assert.NotEmpty(t, task.UID)
		// Assert getting a new index
		assert.NotEmpty(t, task.Index)
		assert.Equal(t, int64(18), task.Index)
		err = s.Commit()
		require.NoError(t, err)

		db.AssertExists(t, "tasks", map[string]interface{}{
			"id":            task.ID,
			"title":         "Lorem",
			"description":   "Lorem Ipsum Dolor",
			"project_id":    1,
			"created_by_id": 1,
		}, false)
		db.AssertExists(t, "task_buckets", map[string]interface{}{
			"task_id":   task.ID,
			"bucket_id": 1,
		}, false)

		events.AssertDispatched(t, &TaskCreatedEvent{})
	})
	t.Run("with reminders", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task := &Task{
			Title:       "Lorem",
			Description: "Lorem Ipsum Dolor",
			ProjectID:   1,
			DueDate:     time.Date(2023, time.March, 7, 22, 5, 0, 0, time.UTC),
			StartDate:   time.Date(2023, time.March, 7, 22, 5, 10, 0, time.UTC),
			EndDate:     time.Date(2023, time.March, 7, 22, 5, 20, 0, time.UTC),
			Reminders: []*TaskReminder{
				{
					RelativeTo:     "due_date",
					RelativePeriod: 1,
				},
				{
					RelativeTo:     "start_date",
					RelativePeriod: -2,
				},
				{
					RelativeTo:     "end_date",
					RelativePeriod: -1,
				},
				{
					Reminder: time.Date(2023, time.March, 7, 23, 0, 0, 0, time.UTC),
				},
			}}
		err := task.Create(s, usr)
		require.NoError(t, err)
		assert.Equal(t, time.Date(2023, time.March, 7, 22, 5, 1, 0, time.UTC), task.Reminders[0].Reminder)
		assert.Equal(t, int64(1), task.Reminders[0].RelativePeriod)
		assert.Equal(t, ReminderRelationDueDate, task.Reminders[0].RelativeTo)
		assert.Equal(t, time.Date(2023, time.March, 7, 22, 5, 8, 0, time.UTC), task.Reminders[1].Reminder)
		assert.Equal(t, ReminderRelationStartDate, task.Reminders[1].RelativeTo)
		assert.Equal(t, time.Date(2023, time.March, 7, 22, 5, 19, 0, time.UTC), task.Reminders[2].Reminder)
		assert.Equal(t, ReminderRelationEndDate, task.Reminders[2].RelativeTo)
		assert.Equal(t, time.Date(2023, time.March, 7, 23, 0, 0, 0, time.UTC), task.Reminders[3].Reminder)
		err = s.Commit()
		require.NoError(t, err)
	})
	t.Run("empty title", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task := &Task{
			Title:       "",
			Description: "Lorem Ipsum Dolor",
			ProjectID:   1,
		}
		err := task.Create(s, usr)
		require.Error(t, err)
		assert.True(t, IsErrTaskCannotBeEmpty(err))
	})
	t.Run("nonexistant project", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task := &Task{
			Title:       "Test",
			Description: "Lorem Ipsum Dolor",
			ProjectID:   9999999,
		}
		err := task.Create(s, usr)
		require.Error(t, err)
		assert.True(t, IsErrProjectDoesNotExist(err))
	})
	t.Run("nonexistant user", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		nUser := &user.User{ID: 99999999}
		task := &Task{
			Title:       "Test",
			Description: "Lorem Ipsum Dolor",
			ProjectID:   1,
		}
		err := task.Create(s, nUser)
		require.Error(t, err)
		assert.True(t, user.IsErrUserDoesNotExist(err))
	})
	t.Run("default bucket different", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task := &Task{
			Title:       "Lorem",
			Description: "Lorem Ipsum Dolor",
			ProjectID:   6,
		}
		err := task.Create(s, usr)
		require.NoError(t, err)
		db.AssertExists(t, "task_buckets", map[string]interface{}{
			"task_id":   task.ID,
			"bucket_id": 22, // default bucket of project 6 but with a position of 2
		}, false)
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
			ProjectID:   1,
		}
		err := task.Update(s, u)
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)

		db.AssertExists(t, "tasks", map[string]interface{}{
			"id":          1,
			"title":       "test10000",
			"description": "Lorem Ipsum Dolor",
			"project_id":  1,
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
			ProjectID:   1,
		}
		err := task.Update(s, u)
		require.Error(t, err)
		assert.True(t, IsErrTaskDoesNotExist(err))
	})
	t.Run("default bucket when moving a task between projects", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task := &Task{
			ID:        1,
			ProjectID: 2,
		}
		err := task.Update(s, u)
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)

		db.AssertExists(t, "task_buckets", map[string]interface{}{
			"task_id": task.ID,
			// bucket 40 is the default bucket on project 2
			"bucket_id": 40,
		}, false)
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
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)
		assert.True(t, task.Done)

		db.AssertExists(t, "tasks", map[string]interface{}{
			"id":   1,
			"done": true,
		}, false)
		db.AssertExists(t, "task_buckets", map[string]interface{}{
			"task_id":   1,
			"bucket_id": 3,
		}, false)
	})
	t.Run("marking a task as done should fire exactly ONE task.updated event", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Clear any events from previous operations
		events.ClearDispatchedEvents()

		task := &Task{
			ID:   1,
			Done: true,
		}
		err := task.Update(s, u)
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)

		// Verify exactly ONE task.updated event was dispatched
		count := events.CountDispatchedEvents("task.updated")
		assert.Equal(t, 1, count, "Expected exactly 1 task.updated event, got %d", count)
	})
	t.Run("move task to another project should use the default bucket", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task := &Task{
			ID:        1,
			ProjectID: 2,
		}
		err := task.Update(s, u)
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)

		db.AssertExists(t, "tasks", map[string]interface{}{
			"id":         1,
			"project_id": 2,
		}, false)
		db.AssertExists(t, "task_buckets", map[string]interface{}{
			"task_id":   1,
			"bucket_id": 40,
		}, false)
	})
	t.Run("move done task to another project with a done bucket", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task := &Task{
			ID:        2,
			Done:      true,
			ProjectID: 2,
		}
		err := task.Update(s, u)
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)

		db.AssertExists(t, "tasks", map[string]interface{}{
			"id":         task.ID,
			"project_id": 2,
			"done":       true,
		}, false)
		db.AssertExists(t, "task_buckets", map[string]interface{}{
			"task_id":   task.ID,
			"bucket_id": 4, // 4 is the done bucket
		}, false)
	})
	t.Run("repeating tasks should not be moved to the done bucket", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task := &Task{
			ID:      28,
			Done:    true,
			Repeats: "FREQ=HOURLY;INTERVAL=1",
		}
		err := task.Update(s, u)
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)
		assert.False(t, task.Done)

		db.AssertExists(t, "tasks", map[string]interface{}{
			"id":   28,
			"done": false,
		}, false)
		db.AssertExists(t, "task_buckets", map[string]interface{}{
			"task_id":   28,
			"bucket_id": 1,
		}, false)
	})
	t.Run("repeating tasks should set done_at when marked done", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Get the task before updating to check done_at was empty
		taskBefore := &Task{ID: 28}
		err := taskBefore.ReadOne(s, u)
		require.NoError(t, err)
		assert.True(t, taskBefore.DoneAt.IsZero())
		assert.False(t, taskBefore.Done)

		// Mark the repeating task as done
		task := &Task{
			ID:      28,
			Done:    true,
			Repeats: "FREQ=HOURLY;INTERVAL=1",
		}
		err = task.Update(s, u)
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)

		// Task should be reset to not done (because it repeats) but done_at should be set
		assert.False(t, task.Done)
		assert.False(t, task.DoneAt.IsZero(), "done_at should be set for repeating tasks when marked as done")

		// Verify in database
		updatedTask := &Task{ID: 28}
		err = updatedTask.ReadOne(s, u)
		require.NoError(t, err)
		assert.False(t, updatedTask.Done)
		assert.False(t, updatedTask.DoneAt.IsZero(), "done_at should be persisted in database for repeating tasks")
	})
	t.Run("moving a task between projects should give it a correct index", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task := &Task{
			ID:        12,
			ProjectID: 2, // From project 1
		}
		err := task.Update(s, u)
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)
		assert.Equal(t, int64(3), task.Index)
	})

	t.Run("reminders will be updated", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task := &Task{
			ID:        1,
			ProjectID: 1,
			Title:     "test",
			DueDate:   time.Date(2023, time.March, 7, 22, 5, 0, 0, time.UTC),
			StartDate: time.Date(2023, time.March, 7, 22, 5, 10, 0, time.UTC),
			EndDate:   time.Date(2023, time.March, 7, 22, 5, 20, 0, time.UTC),
			Reminders: []*TaskReminder{
				{
					RelativeTo:     "due_date",
					RelativePeriod: 1,
				},
				{
					RelativeTo:     "start_date",
					RelativePeriod: -2,
				},
				{
					RelativeTo:     "end_date",
					RelativePeriod: -1,
				},
				{
					Reminder: time.Date(2023, time.March, 7, 23, 0, 0, 0, time.UTC),
				},
			}}
		err := task.Update(s, u)
		require.NoError(t, err)
		assert.Equal(t, time.Date(2023, time.March, 7, 22, 5, 1, 0, time.UTC), task.Reminders[0].Reminder)
		assert.Equal(t, int64(1), task.Reminders[0].RelativePeriod)
		assert.Equal(t, ReminderRelationDueDate, task.Reminders[0].RelativeTo)
		assert.Equal(t, time.Date(2023, time.March, 7, 22, 5, 8, 0, time.UTC), task.Reminders[1].Reminder)
		assert.Equal(t, ReminderRelationStartDate, task.Reminders[1].RelativeTo)
		assert.Equal(t, time.Date(2023, time.March, 7, 22, 5, 19, 0, time.UTC), task.Reminders[2].Reminder)
		assert.Equal(t, ReminderRelationEndDate, task.Reminders[2].RelativeTo)
		assert.Equal(t, time.Date(2023, time.March, 7, 23, 0, 0, 0, time.UTC), task.Reminders[3].Reminder)
		err = s.Commit()
		require.NoError(t, err)
		db.AssertCount(t, "task_reminders", builder.Eq{"task_id": 1}, 4)
	})
	t.Run("the same reminder multiple times should be saved once", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task := &Task{
			ID:    1,
			Title: "test",
			Reminders: []*TaskReminder{
				{
					Reminder: time.Unix(1674745156, 0),
				},
				{
					Reminder: time.Unix(1674745156, 223),
				},
			},
			ProjectID: 1,
		}
		err := task.Update(s, u)
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)
		db.AssertCount(t, "task_reminders", builder.Eq{"task_id": 1}, 1)
	})
	t.Run("update relative reminder when start_date changes", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// given task with start_date and relative reminder for start_date
		taskBefore := &Task{
			Title:     "test",
			ProjectID: 1,
			StartDate: time.Date(2022, time.March, 8, 8, 5, 20, 0, time.UTC),
			Reminders: []*TaskReminder{
				{
					RelativeTo:     "start_date",
					RelativePeriod: -60,
				},
			}}
		err := taskBefore.Create(s, u)
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)
		assert.Equal(t, time.Date(2022, time.March, 8, 8, 4, 20, 0, time.UTC), taskBefore.Reminders[0].Reminder)

		// when start_date is modified
		task := taskBefore
		task.StartDate = time.Date(2023, time.March, 8, 8, 5, 0, 0, time.UTC)
		err = task.Update(s, u)
		require.NoError(t, err)

		// then reminder time is updated
		assert.Equal(t, time.Date(2023, time.March, 8, 8, 4, 0, 0, time.UTC), task.Reminders[0].Reminder)
		err = s.Commit()
		require.NoError(t, err)
	})
	t.Run("don't allow done_at change when passing fields", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task := &Task{
			ID:     1,
			DoneAt: time.Now(),
		}

		err := task.updateSingleTask(s, u, []string{"done_at"})

		require.Error(t, err)
		assert.Contains(t, err.Error(), `Task column done_at is invalid`)
		require.NoError(t, s.Commit())
	})
	t.Run("ignore done_at when updating unrelated values", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task := &Task{
			ID:     1,
			Title:  "updated",
			DoneAt: time.Now(),
		}

		err := task.Update(s, u)

		require.NoError(t, err)
		require.NoError(t, s.Commit())

		updatedTask := &Task{ID: 1}
		err = updatedTask.ReadOne(s, u)
		require.NoError(t, err)
		assert.Equal(t, "updated", updatedTask.Title)
		assert.True(t, updatedTask.DoneAt.IsZero())
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
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)

		db.AssertMissing(t, "tasks", map[string]interface{}{
			"id": 1,
		})
	})
}

func TestUpdateTasksHelper(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	u := &user.User{ID: 1}
	updates := &Task{Title: "helper"}
	updated, err := updateTasks(s, u, updates, []int64{10}, []string{"title"})
	require.NoError(t, err)
	require.Len(t, updated, 1)
	assert.Equal(t, "helper", updated[0].Title)
	assert.False(t, updated[0].Done)
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
	t.Run("no repeats set", func(t *testing.T) {
		dueDate := time.Unix(1550000000, 0)
		oldTask := &Task{
			Done:    false,
			Repeats: "",
			DueDate: dueDate,
		}
		newTask := &Task{
			Done:    true,
			DueDate: dueDate,
		}
		updateDone(oldTask, newTask)

		assert.Equal(t, dueDate.Unix(), newTask.DueDate.Unix())
		assert.True(t, newTask.Done)
	})
	t.Run("repeating interval with RRULE", func(t *testing.T) {
		t.Run("daily repeat", func(t *testing.T) {
			oldDueDate := time.Now().Add(-48 * time.Hour) // 2 days ago
			oldTask := &Task{
				Done:    false,
				Repeats: "FREQ=DAILY;INTERVAL=1",
				DueDate: oldDueDate,
			}
			newTask := &Task{
				Done: true,
			}
			updateDone(oldTask, newTask)

			// Next occurrence should be after now
			assert.True(t, newTask.DueDate.After(time.Now()) || newTask.DueDate.Equal(time.Now().Truncate(24*time.Hour)))
			assert.False(t, newTask.Done)
		})
		t.Run("no due date is a no-op", func(t *testing.T) {
			oldTask := &Task{
				Done:    false,
				Repeats: "FREQ=DAILY;INTERVAL=1",
				DueDate: time.Time{},
			}
			newTask := &Task{
				Done: true,
			}
			updateDone(oldTask, newTask)
			// Repeating task without a due date should not get one auto-assigned
			assert.True(t, newTask.DueDate.IsZero(), "due date should remain unset")
		})
		t.Run("update reminders", func(t *testing.T) {
			oldReminder1 := time.Now().Add(-48 * time.Hour)
			oldReminder2 := time.Now().Add(-24 * time.Hour)
			oldTask := &Task{
				Done:    false,
				Repeats: "FREQ=DAILY;INTERVAL=1",
				DueDate: time.Now().Add(-48 * time.Hour),
				Reminders: []*TaskReminder{
					{
						Reminder: oldReminder1,
					},
					{
						Reminder: oldReminder2,
					},
				},
			}
			newTask := &Task{
				Done: true,
			}
			updateDone(oldTask, newTask)

			assert.Len(t, newTask.Reminders, 2)
			// New reminders should be in the future
			assert.True(t, newTask.Reminders[0].Reminder.After(oldReminder1))
			assert.True(t, newTask.Reminders[1].Reminder.After(oldReminder2))
			assert.False(t, newTask.Done)
		})
		t.Run("update start date", func(t *testing.T) {
			oldStartDate := time.Now().Add(-48 * time.Hour)
			oldTask := &Task{
				Done:      false,
				Repeats:   "FREQ=DAILY;INTERVAL=1",
				DueDate:   time.Now().Add(-48 * time.Hour),
				StartDate: oldStartDate,
			}
			newTask := &Task{
				Done: true,
			}
			updateDone(oldTask, newTask)

			assert.True(t, newTask.StartDate.After(oldStartDate))
			assert.False(t, newTask.Done)
		})
		t.Run("update end date", func(t *testing.T) {
			oldEndDate := time.Now().Add(-48 * time.Hour)
			oldTask := &Task{
				Done:    false,
				Repeats: "FREQ=DAILY;INTERVAL=1",
				DueDate: time.Now().Add(-48 * time.Hour),
				EndDate: oldEndDate,
			}
			newTask := &Task{
				Done: true,
			}
			updateDone(oldTask, newTask)

			assert.True(t, newTask.EndDate.After(oldEndDate))
			assert.False(t, newTask.Done)
		})
		t.Run("ensure due date is repeated even if the original one is in the future", func(t *testing.T) {
			oldDueDate := time.Now().Add(time.Hour)
			oldTask := &Task{
				Done:    false,
				Repeats: "FREQ=DAILY;INTERVAL=1",
				DueDate: oldDueDate,
			}
			newTask := &Task{
				Done: true,
			}
			updateDone(oldTask, newTask)
			// Next occurrence should be after the original due date
			assert.True(t, newTask.DueDate.After(oldDueDate))
			assert.False(t, newTask.Done)
		})
		t.Run("repeat from current date", func(t *testing.T) {
			t.Run("due date", func(t *testing.T) {
				oldTask := &Task{
					Done:                   false,
					Repeats:                "FREQ=DAILY;INTERVAL=1",
					RepeatsFromCurrentDate: true,
					DueDate:                time.Unix(1550000000, 0),
				}
				newTask := &Task{
					Done: true,
				}
				updateDone(oldTask, newTask)

				// Should calculate from now, so new due date should be tomorrow or later
				assert.True(t, newTask.DueDate.After(time.Now()))
				assert.False(t, newTask.Done)
			})
			t.Run("start date", func(t *testing.T) {
				oldTask := &Task{
					Done:                   false,
					Repeats:                "FREQ=DAILY;INTERVAL=1",
					RepeatsFromCurrentDate: true,
					StartDate:              time.Unix(1550000000, 0),
				}
				newTask := &Task{
					Done: true,
				}
				updateDone(oldTask, newTask)

				assert.True(t, newTask.StartDate.After(time.Now()))
				assert.False(t, newTask.Done)
			})
			t.Run("end date", func(t *testing.T) {
				oldTask := &Task{
					Done:                   false,
					Repeats:                "FREQ=DAILY;INTERVAL=1",
					RepeatsFromCurrentDate: true,
					EndDate:                time.Unix(1560000000, 0),
				}
				newTask := &Task{
					Done: true,
				}
				updateDone(oldTask, newTask)

				assert.True(t, newTask.EndDate.After(time.Now()))
				assert.False(t, newTask.Done)
			})
		})
		t.Run("repeat each month", func(t *testing.T) {
			t.Run("due date", func(t *testing.T) {
				oldTask := &Task{
					Done:    false,
					Repeats: "FREQ=MONTHLY;INTERVAL=1",
					DueDate: time.Unix(1550000000, 0),
				}
				newTask := &Task{
					Done: true,
				}
				oldDueDate := oldTask.DueDate

				updateDone(oldTask, newTask)

				assert.True(t, newTask.DueDate.After(oldDueDate))
				assert.False(t, newTask.Done)
			})
			t.Run("start date", func(t *testing.T) {
				oldTask := &Task{
					Done:      false,
					Repeats:   "FREQ=MONTHLY;INTERVAL=1",
					DueDate:   time.Unix(1550000000, 0),
					StartDate: time.Unix(1550000000, 0),
				}
				newTask := &Task{
					Done: true,
				}
				oldStartDate := oldTask.StartDate

				updateDone(oldTask, newTask)

				assert.True(t, newTask.StartDate.After(oldStartDate))
				assert.False(t, newTask.Done)
			})
			t.Run("end date", func(t *testing.T) {
				oldTask := &Task{
					Done:    false,
					Repeats: "FREQ=MONTHLY;INTERVAL=1",
					DueDate: time.Unix(1560000000, 0),
					EndDate: time.Unix(1560000000, 0),
				}
				newTask := &Task{
					Done: true,
				}
				oldEndDate := oldTask.EndDate

				updateDone(oldTask, newTask)

				assert.True(t, newTask.EndDate.After(oldEndDate))
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
		require.NoError(t, err)
		assert.Equal(t, "task #1", task.Title)
	})
	t.Run("nonexisting", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task := &Task{ID: 99999}
		err := task.ReadOne(s, u)
		require.Error(t, err)
		assert.True(t, IsErrTaskDoesNotExist(err))
	})
	t.Run("with subscription", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task := &Task{ID: 22}
		err := task.ReadOne(s, &user.User{ID: 6})
		require.NoError(t, err)
		assert.NotNil(t, task.Subscription)
	})
	t.Run("created by link share", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task := &Task{ID: 37}
		err := task.ReadOne(s, u)
		require.NoError(t, err)
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
		require.NoError(t, err)
		assert.True(t, task.IsFavorite)
	})
	t.Run("favorite for a different user", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task := &Task{ID: 1}
		err := task.ReadOne(s, &user.User{ID: 2})
		require.NoError(t, err)
		assert.False(t, task.IsFavorite)
	})
}

func Test_getTaskIndexFromSearchString(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name      string
		args      args
		wantIndex int64
	}{
		{
			name:      "task index in text",
			args:      args{s: "Task #12"},
			wantIndex: 12,
		},
		{
			name:      "no task index",
			args:      args{s: "Task"},
			wantIndex: 0,
		},
		{
			name:      "not numeric but with prefix",
			args:      args{s: "Task #aaaaa"},
			wantIndex: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotIndex := getTaskIndexFromSearchString(tt.args.s); gotIndex != tt.wantIndex {
				t.Errorf("getTaskIndexFromSearchString() = %v, want %v", gotIndex, tt.wantIndex)
			}
		})
	}
}
