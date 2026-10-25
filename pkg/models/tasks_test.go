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
	"context"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teambition/rrule-go"
	"xorm.io/builder"
	"xorm.io/xorm"
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
		// The soft-deleted task 51 holds index 34, which must not be reused
		assert.NotEmpty(t, task.Index)
		assert.Equal(t, int64(35), task.Index)
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
		db.AssertExists(t, "subscriptions", map[string]interface{}{
			"entity_type": SubscriptionEntityTask,
			"entity_id":   task.ID,
			"user_id":     usr.ID,
		}, false)

		events.DispatchPending(context.Background(), s)
		events.AssertDispatched(t, &TaskCreatedEvent{})
	})
	t.Run("already subscribed to the project", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// user6 is subscribed to project 12 (fixture), so the inherited subscription must not
		// be duplicated with a task level one
		usr6 := &user.User{ID: 6, Username: "user6"}
		task := &Task{
			Title:     "Lorem",
			ProjectID: 12,
		}
		err := task.Create(s, usr6)
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)

		db.AssertMissing(t, "subscriptions", map[string]interface{}{
			"entity_type": SubscriptionEntityTask,
			"entity_id":   task.ID,
		})
	})
	t.Run("created by link share is not subscribed", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		linkShare := &LinkSharing{ID: 2, ProjectID: 2, Permission: PermissionWrite}
		task := &Task{
			Title:     "Lorem",
			ProjectID: 2,
		}
		err := task.Create(s, linkShare)
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)

		db.AssertMissing(t, "subscriptions", map[string]interface{}{
			"entity_type": SubscriptionEntityTask,
			"entity_id":   task.ID,
		})
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
		require.NoError(t, s.Commit())
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
	t.Run("hex color with leading #", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task := &Task{
			ID:        1,
			Title:     "test",
			ProjectID: 1,
			HexColor:  "#f0f0f0",
		}
		err := task.Update(s, u)
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)

		assert.Equal(t, "f0f0f0", task.HexColor)
		db.AssertExists(t, "tasks", map[string]interface{}{
			"id":        1,
			"hex_color": "f0f0f0",
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

		events.DispatchPending(context.Background(), s)
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
	t.Run("repeating tasks marked done from a non-default bucket are moved to the default bucket", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Pre-position task 28 in bucket 2 (non-default, non-done) via a
		// raw update to bypass the bucket-limit check.
		_, err := s.Where("task_id = ? AND project_view_id = ?", 28, 4).
			Cols("bucket_id").
			Update(&TaskBucket{BucketID: 2})
		require.NoError(t, err)

		// Mark the repeating task as done via Task.Update (same code path
		// the frontend hits when the user clicks "Done" in the task
		// detail pane).
		task := &Task{
			ID:      28,
			Done:    true,
			Repeats: "FREQ=HOURLY;INTERVAL=1",
		}
		err = task.Update(s, u)
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)

		// updateDone should have re-opened the task for the next iteration.
		assert.False(t, task.Done)

		// And the task should now be sitting in the default bucket (1),
		// not left in bucket 2 or moved to the done bucket (3).
		db.AssertExists(t, "task_buckets", map[string]interface{}{
			"task_id":         28,
			"project_view_id": 4,
			"bucket_id":       1,
		}, false)
		db.AssertMissing(t, "task_buckets", map[string]interface{}{
			"task_id":         28,
			"project_view_id": 4,
			"bucket_id":       2,
		})
		db.AssertMissing(t, "task_buckets", map[string]interface{}{
			"task_id":         28,
			"project_view_id": 4,
			"bucket_id":       3,
		})
	})
	t.Run("repeating tasks marked done when no default bucket is configured stay in their bucket", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// View 4 has default_bucket_id: 1. Remove it to hit the
		// no-default branch of moveTaskToDefaultBuckets.
		_, err := s.ID(4).Cols("default_bucket_id").Update(&ProjectView{DefaultBucketID: 0})
		require.NoError(t, err)

		// Pre-position task 28 in bucket 2 (non-default, non-done) via a
		// raw update to bypass the bucket-limit check.
		_, err = s.Where("task_id = ? AND project_view_id = ?", 28, 4).
			Cols("bucket_id").
			Update(&TaskBucket{BucketID: 2})
		require.NoError(t, err)

		task := &Task{
			ID:      28,
			Done:    true,
			Repeats: "FREQ=HOURLY;INTERVAL=1",
		}
		err = task.Update(s, u)
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)

		// updateDone should have re-opened the task for the next iteration.
		assert.False(t, task.Done)

		// The task stays in bucket 2 — not moved to the first bucket (1)
		// and not into the done bucket (3).
		db.AssertExists(t, "task_buckets", map[string]interface{}{
			"task_id":         28,
			"project_view_id": 4,
			"bucket_id":       2,
		}, false)
		db.AssertMissing(t, "task_buckets", map[string]interface{}{
			"task_id":         28,
			"project_view_id": 4,
			"bucket_id":       1,
		})
		db.AssertMissing(t, "task_buckets", map[string]interface{}{
			"task_id":         28,
			"project_view_id": 4,
			"bucket_id":       3,
		})
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
	t.Run("passing fields restricts the write to the given columns", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Priority stays zero to simulate a caller not parsing that column (fixture has 100).
		task := &Task{
			ID:       3,
			Title:    "updated with a field list",
			Priority: 0,
		}

		err := task.updateSingleTask(s, u, []string{"title"})

		require.NoError(t, err)
		require.NoError(t, s.Commit())

		updatedTask := &Task{ID: 3}
		err = updatedTask.ReadOne(s, u)
		require.NoError(t, err)
		assert.Equal(t, "updated with a field list", updatedTask.Title)
		assert.Equal(t, int64(100), updatedTask.Priority, "priority must survive since it wasn't in the fields list")
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

		events.DispatchPending(context.Background(), s)
		events.AssertDispatched(t, &TaskDeletedEvent{})

		s2 := db.NewSession()
		defer s2.Close()

		// The row is still there, only marked as deleted
		deletedTask := &Task{}
		has, err := s2.Unscoped().Where("id = ?", 1).Get(deletedTask)
		require.NoError(t, err)
		require.True(t, has)
		assert.False(t, deletedTask.DeletedAt.IsZero())

		readTask := &Task{ID: 1}
		err = readTask.ReadOne(s2, &user.User{ID: 1})
		require.Error(t, err)
		assert.True(t, IsErrTaskDoesNotExist(err))

		// Position and bucket rows are removed right away because bucket counts
		// don't join the tasks table
		db.AssertMissing(t, "task_positions", map[string]interface{}{"task_id": 1})
		db.AssertMissing(t, "task_buckets", map[string]interface{}{"task_id": 1})

		// Everything else is kept for a possible restore
		db.AssertExists(t, "task_comments", map[string]interface{}{"task_id": 1}, false)
		db.AssertExists(t, "task_attachments", map[string]interface{}{"task_id": 1}, false)
		db.AssertExists(t, "label_tasks", map[string]interface{}{"task_id": 1}, false)
		db.AssertExists(t, "task_relations", map[string]interface{}{"task_id": 1}, false)
		db.AssertExists(t, "favorites", map[string]interface{}{"entity_id": 1, "kind": FavoriteKindTask}, false)
		db.AssertExists(t, "reactions", map[string]interface{}{"entity_id": 1, "entity_kind": ReactionKindTask}, false)

		// The project's updated timestamp is bumped — regression for the delete
		// receiver only carrying the task id, not the project id
		project := &Project{}
		has, err = s2.Where("id = ?", 1).Get(project)
		require.NoError(t, err)
		require.True(t, has)
		assert.True(t, project.Updated.After(deletedTask.Created))
	})
}

func TestHardDeleteTask(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	files.InitTestFileFixtures(t)
	s := db.NewSession()
	defer s.Close()

	// Add the child rows task 1 doesn't have in the fixtures so the sweep covers
	// every related table
	_, err := s.Insert(&TaskAssginee{TaskID: 1, UserID: 2})
	require.NoError(t, err)
	_, err = s.Insert(&TaskReminder{TaskID: 1, Reminder: time.Now()})
	require.NoError(t, err)
	_, err = s.Insert(&TaskUnreadStatus{TaskID: 1, UserID: 2})
	require.NoError(t, err)
	_, err = s.Insert(
		&TaskIndexAlias{ProjectID: 2, Index: 99, TaskID: 1},
		&TaskIndexAlias{ProjectID: 3, Index: 99, TaskID: 1},
	)
	require.NoError(t, err)
	_, err = s.Insert(&Subscription{EntityType: SubscriptionEntityTask, EntityID: 1, UserID: 2})
	require.NoError(t, err)
	// Comment 1 belongs to task 1
	_, err = s.Insert(&Reaction{EntityID: 1, EntityKind: ReactionKindComment, UserID: 1, Value: "👍"})
	require.NoError(t, err)

	err = hardDeleteTask(s, &Task{ID: 1})
	require.NoError(t, err)
	require.NoError(t, s.Commit())

	db.AssertMissing(t, "tasks", map[string]interface{}{"id": 1})
	for _, table := range []string{
		"task_assignees",
		"task_comments",
		"task_attachments",
		"label_tasks",
		"task_relations",
		"task_reminders",
		"task_positions",
		"task_buckets",
		"task_unread_statuses",
	} {
		db.AssertMissing(t, table, map[string]interface{}{"task_id": 1})
	}
	db.AssertMissing(t, "task_relations", map[string]interface{}{"other_task_id": 1})
	db.AssertMissing(t, "favorites", map[string]interface{}{"entity_id": 1, "kind": FavoriteKindTask})
	db.AssertMissing(t, "subscriptions", map[string]interface{}{"entity_id": 1, "entity_type": SubscriptionEntityTask})
	db.AssertMissing(t, "task_index_aliases", map[string]interface{}{"task_id": 1})
	db.AssertMissing(t, "reactions", map[string]interface{}{"entity_id": 1, "entity_kind": ReactionKindTask})
	db.AssertMissing(t, "reactions", map[string]interface{}{"entity_id": 1, "entity_kind": ReactionKindComment})
	// The attachment files are gone too
	db.AssertMissing(t, "files", map[string]interface{}{"id": 1})
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
		updateDone(oldTask, newTask, config.GetTimeZone())
		assert.NotEqual(t, time.Time{}, newTask.DoneAt)
	})
	t.Run("unmarking a task as done", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		oldTask := &Task{Done: true}
		newTask := &Task{Done: false}
		updateDone(oldTask, newTask, config.GetTimeZone())
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
		updateDone(oldTask, newTask, config.GetTimeZone())

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
			updateDone(oldTask, newTask, config.GetTimeZone())

			// A 2-day-overdue daily rule skips the two past days and lands exactly
			// one day in the future: oldDueDate + 3 days. Weekly would be +7 days,
			// so an exact delta guards against the frequency being misinterpreted.
			assert.Equal(t, oldDueDate.Add(72*time.Hour).Unix(), newTask.DueDate.Unix())
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
			updateDone(oldTask, newTask, config.GetTimeZone())
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
			updateDone(oldTask, newTask, config.GetTimeZone())

			assert.Len(t, newTask.Reminders, 2)
			// New reminders should be in the future
			assert.True(t, newTask.Reminders[0].Reminder.After(oldReminder1))
			assert.True(t, newTask.Reminders[1].Reminder.After(oldReminder2))
			assert.False(t, newTask.Done)
		})
		t.Run("update start date", func(t *testing.T) {
			// Start and due share the same 2-day-overdue anchor, so the due date
			// shifts by exactly 72h and start must shift by the same delta to keep
			// its offset from the due date intact.
			base := time.Now().Add(-48 * time.Hour)
			oldTask := &Task{
				Done:      false,
				Repeats:   "FREQ=DAILY;INTERVAL=1",
				DueDate:   base,
				StartDate: base,
			}
			newTask := &Task{
				Done: true,
			}
			updateDone(oldTask, newTask, config.GetTimeZone())

			assert.Equal(t, base.Add(72*time.Hour).Unix(), newTask.StartDate.Unix())
			assert.False(t, newTask.Done)
		})
		t.Run("update end date", func(t *testing.T) {
			base := time.Now().Add(-48 * time.Hour)
			oldTask := &Task{
				Done:    false,
				Repeats: "FREQ=DAILY;INTERVAL=1",
				DueDate: base,
				EndDate: base,
			}
			newTask := &Task{
				Done: true,
			}
			updateDone(oldTask, newTask, config.GetTimeZone())

			assert.Equal(t, base.Add(72*time.Hour).Unix(), newTask.EndDate.Unix())
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
			updateDone(oldTask, newTask, config.GetTimeZone())
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
				updateDone(oldTask, newTask, config.GetTimeZone())

				// Should calculate from now, so new due date should be tomorrow or later
				assert.True(t, newTask.DueDate.After(time.Now()))
				assert.False(t, newTask.Done)
			})
			t.Run("future due date searches from now (regression)", func(t *testing.T) {
				// Repeat-from-current-date with a due date in the FUTURE must advance
				// from now, not from the future due date.
				futureDue := time.Now().Add(30 * 24 * time.Hour)
				oldTask := &Task{
					Done:                   false,
					Repeats:                "FREQ=DAILY;INTERVAL=1",
					RepeatsFromCurrentDate: true,
					DueDate:                futureDue,
				}
				newTask := &Task{
					Done: true,
				}
				updateDone(oldTask, newTask, config.GetTimeZone())

				// Next occurrence is ~tomorrow (measured from now), well before the
				// original future due date — not futureDue+1day.
				assert.True(t, newTask.DueDate.After(time.Now()))
				assert.True(t, newTask.DueDate.Before(futureDue))
				assert.False(t, newTask.Done)
			})
			t.Run("reminders spanning more than 292 years", func(t *testing.T) {
				// time.Duration saturates at ~292 years; the offset between reminders
				// must not be computed as a single Duration.
				oldTask := &Task{
					Done:                   false,
					Repeats:                "FREQ=DAILY;INTERVAL=3650",
					RepeatsFromCurrentDate: true,
					Reminders: []*TaskReminder{
						{Reminder: time.Date(1734, 1, 1, 0, 0, 0, 0, time.UTC)},
						{Reminder: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
					}}
				newTask := &Task{
					Done: true,
				}
				updateDone(oldTask, newTask, config.GetTimeZone())

				assert.Len(t, newTask.Reminders, 2)
				expectedFirst := time.Now().Add(3650 * 24 * time.Hour)
				assert.Equal(t, expectedFirst.Unix(), newTask.Reminders[0].Reminder.Unix())
				assert.Equal(t, expectedFirst.Year()+292, newTask.Reminders[1].Reminder.Year())
				assert.True(t, newTask.Reminders[1].Reminder.After(newTask.Reminders[0].Reminder))
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
				updateDone(oldTask, newTask, config.GetTimeZone())

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
				updateDone(oldTask, newTask, config.GetTimeZone())

				assert.True(t, newTask.EndDate.After(time.Now()))
				assert.False(t, newTask.Done)
			})
			t.Run("preserves start-to-due gap (regression)", func(t *testing.T) {
				// With both start and due dates, repeat-from-current-date must keep
				// the start->due offset; it previously collapsed start onto the due
				// date.
				oldStart := time.Unix(1550000000, 0)
				oldDue := oldStart.Add(72 * time.Hour) // 3-day gap
				oldTask := &Task{
					Done:                   false,
					Repeats:                "FREQ=DAILY;INTERVAL=1",
					RepeatsFromCurrentDate: true,
					StartDate:              oldStart,
					DueDate:                oldDue,
				}
				newTask := &Task{
					Done: true,
				}
				updateDone(oldTask, newTask, config.GetTimeZone())

				assert.False(t, newTask.Done)
				// Due date moves to ~now+1day (measured from completion).
				assert.True(t, newTask.DueDate.After(time.Now()))
				// The 3-day start->due gap is preserved, not collapsed to zero. The start moves by
				// the due date's whole local days and keeps its own clock time: three calendar days
				// before the new due date, at the old start's time (the zone is the server's, GMT,
				// here), rather than exactly 72 elapsed hours before it.
				gmt := config.GetTimeZone()
				newDue, start := newTask.DueDate.In(gmt), oldStart.In(gmt)
				wantStart := time.Date(newDue.Year(), newDue.Month(), newDue.Day()-3, start.Hour(), start.Minute(), start.Second(), 0, gmt)
				assert.True(t, newTask.StartDate.Equal(wantStart), "start: got %v, want %v", newTask.StartDate, wantStart)
			})
		})
		t.Run("finite recurrence (COUNT) decrements and terminates", func(t *testing.T) {
			// COUNT=3 yields exactly 3 completable occurrences: the stored rule is
			// decremented on each completion and the task stays done once exhausted,
			// instead of repeating forever. RepeatsFromCurrentDate keeps the rule's
			// count window anchored at now so each completion reschedules.
			now := time.Now()

			// Completion 1: COUNT 3 -> 2, reschedules (task reopens).
			t1 := &Task{Repeats: "FREQ=DAILY;COUNT=3", RepeatsFromCurrentDate: true, DueDate: now}
			n1 := &Task{Done: true}
			updateDone(t1, n1, config.GetTimeZone())
			require.False(t, n1.Done, "should reopen after completing occurrence 1 of 3")
			o1, err := rrule.StrToROption(n1.Repeats)
			require.NoError(t, err)
			assert.Equal(t, 2, o1.Count)

			// Completion 2: COUNT 2 -> 1, reschedules.
			t2 := &Task{Repeats: n1.Repeats, RepeatsFromCurrentDate: true, DueDate: n1.DueDate}
			n2 := &Task{Done: true}
			updateDone(t2, n2, config.GetTimeZone())
			require.False(t, n2.Done, "should reopen after completing occurrence 2 of 3")
			o2, err := rrule.StrToROption(n2.Repeats)
			require.NoError(t, err)
			assert.Equal(t, 1, o2.Count)

			// Completion 3: COUNT == 1, exhausted -> stays done, no reschedule.
			t3 := &Task{Repeats: n2.Repeats, RepeatsFromCurrentDate: true, DueDate: n2.DueDate}
			n3 := &Task{Done: true}
			updateDone(t3, n3, config.GetTimeZone())
			assert.True(t, n3.Done, "should stay done once the finite recurrence is exhausted")
			assert.Empty(t, n3.Repeats, "exhausted recurrence should clear the rule so the UI stops showing it")
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

				updateDone(oldTask, newTask, config.GetTimeZone())

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

				updateDone(oldTask, newTask, config.GetTimeZone())

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

				updateDone(oldTask, newTask, config.GetTimeZone())

				assert.True(t, newTask.EndDate.After(oldEndDate))
				assert.False(t, newTask.Done)
			})
		})
		t.Run("reset checklist on recurrence", func(t *testing.T) {
			const checked = `before<ul data-type="taskList"><li data-checked="true" data-type="taskItem"><label><input type="checkbox" checked="checked"><span></span></label><div><p>Item</p></li></ul>after`
			const unchecked = `before<ul data-type="taskList"><li data-checked="false" data-type="taskItem"><label><input type="checkbox"><span></span></label><div><p>Item</p></li></ul>after`

			oldTask := &Task{
				Done:    false,
				Repeats: "FREQ=DAILY;INTERVAL=1",
				DueDate: time.Unix(1550000000, 0),
			}
			newTask := &Task{
				Done:        true,
				Description: checked,
			}

			updateDone(oldTask, newTask, config.GetTimeZone())

			assert.False(t, newTask.Done)
			assert.True(t, newTask.DueDate.After(oldTask.DueDate))
			assert.Equal(t, unchecked, newTask.Description)
		})
		t.Run("non-recurring description untouched", func(t *testing.T) {
			const checked = `before<ul data-type="taskList"><li data-checked="true" data-type="taskItem"><label><input type="checkbox" checked="checked"><span></span></label><div><p>Item</p></li></ul>after`

			oldTask := &Task{
				Done:    false,
				DueDate: time.Unix(1550000000, 0),
			}
			newTask := &Task{
				Done:        true,
				Description: checked,
			}

			updateDone(oldTask, newTask, config.GetTimeZone())

			assert.True(t, newTask.Done)
			assert.Equal(t, checked, newTask.Description)
		})
	})
}

func TestValidateRRule(t *testing.T) {
	t.Run("empty string is valid", func(t *testing.T) {
		assert.NoError(t, validateRRule(""))
	})
	t.Run("valid daily rule", func(t *testing.T) {
		assert.NoError(t, validateRRule("FREQ=DAILY;INTERVAL=1"))
	})
	t.Run("valid weekly with byday", func(t *testing.T) {
		assert.NoError(t, validateRRule("FREQ=WEEKLY;INTERVAL=2;BYDAY=MO,WE,FR"))
	})
	t.Run("valid monthly with bymonthday", func(t *testing.T) {
		assert.NoError(t, validateRRule("FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15"))
	})
	t.Run("malformed string rejected", func(t *testing.T) {
		err := validateRRule("not a valid rrule")
		require.Error(t, err)
		assert.True(t, IsErrInvalidData(err))
	})
	t.Run("missing freq rejected", func(t *testing.T) {
		err := validateRRule("INTERVAL=2")
		require.Error(t, err)
		assert.True(t, IsErrInvalidData(err))
	})
	t.Run("plain sub-daily rules stay valid", func(t *testing.T) {
		// The legacy migration and the UI only author plain sub-daily interval
		// rules, so these must keep passing.
		assert.NoError(t, validateRRule("FREQ=SECONDLY;INTERVAL=7200"))
		assert.NoError(t, validateRRule("FREQ=MINUTELY;INTERVAL=30"))
		assert.NoError(t, validateRRule("FREQ=HOURLY;INTERVAL=2"))
	})
	t.Run("sub-daily with by-part rejected", func(t *testing.T) {
		// FREQ=SECONDLY;BYSECOND=0 is a time-only filter that defeats the rrule
		// library's day-skip optimization and iterates second-by-second (DoS).
		for _, r := range []string{
			"FREQ=SECONDLY;BYSECOND=0",
			"FREQ=MINUTELY;BYMINUTE=0",
			"FREQ=HOURLY;BYHOUR=9",
			"FREQ=SECONDLY;BYDAY=MO",
		} {
			err := validateRRule(r)
			require.Error(t, err, r)
			assert.True(t, IsErrInvalidData(err), r)
		}
	})
}

func TestFixedDurationForRRule(t *testing.T) {
	cases := []struct {
		name         string
		freq         rrule.Frequency
		interval     int
		wantDuration time.Duration
		wantOK       bool
	}{
		// Sub-daily frequencies step by elapsed time, so they map to a Duration.
		{"secondly interval 30", rrule.SECONDLY, 30, 30 * time.Second, true},
		{"minutely interval 15", rrule.MINUTELY, 15, 15 * time.Minute, true},
		{"hourly interval 1", rrule.HOURLY, 1, time.Hour, true},
		// DAILY and WEEKLY step on the local wall clock,
		// so they are no longer a fixed number of seconds. Their no-due-date policy is tested by
		// TestNextOccurrenceRuleAttributes.
		{"daily interval 1", rrule.DAILY, 1, 0, false},
		{"weekly interval 2", rrule.WEEKLY, 2, 0, false},
		// Months and years have variable length, so the caller falls back to
		// rrule.After instead of using a Duration.
		{"monthly is not fixed", rrule.MONTHLY, 1, 0, false},
		{"yearly is not fixed", rrule.YEARLY, 1, 0, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			d, ok := fixedDurationForRRule(&rrule.ROption{Freq: c.freq, Interval: c.interval})
			assert.Equal(t, c.wantOK, ok)
			assert.Equal(t, c.wantDuration, d)
		})
	}
}

// TestUpdateDone_UnmovedDueKeepsACompanionEdit: when the next occurrence is the stored due date
// itself (from completion, due exactly one step ahead), completion moves no companion date, so a
// start or end the same update edits or clears is kept, as before completion became a
// nextOccurrence call.
func TestUpdateDone_UnmovedDueKeepsACompanionEdit(t *testing.T) {
	now := time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC)
	withFixedNow(t, now)
	due := time.Date(2026, 10, 23, 12, 0, 0, 0, time.UTC)
	oldTask := &Task{
		Repeats: "FREQ=DAILY", RepeatsFromCurrentDate: true, DueDate: due,
		StartDate: time.Date(2026, 10, 23, 10, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 10, 23, 14, 0, 0, 0, time.UTC),
	}
	editedStart := time.Date(2026, 10, 23, 17, 0, 0, 0, time.UTC)
	newTask := &Task{Done: true, Repeats: oldTask.Repeats, DueDate: due, StartDate: editedStart} // end cleared

	updateDone(oldTask, newTask, time.UTC)
	assert.False(t, newTask.Done)
	assert.True(t, newTask.DueDate.Equal(due), "got %v", newTask.DueDate)
	assert.True(t, newTask.StartDate.Equal(editedStart), "the edited start is kept, got %v", newTask.StartDate)
	assert.True(t, newTask.EndDate.IsZero(), "the cleared end stays cleared, got %v", newTask.EndDate)
}

func TestUpdateDone_RRuleAncientDueDate(t *testing.T) {
	oldTask := &Task{
		Done:      false,
		Repeats:   "FREQ=SECONDLY;INTERVAL=1",
		DueDate:   time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC),
		StartDate: time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(1900, 1, 2, 0, 0, 0, 0, time.UTC),
		Reminders: []*TaskReminder{
			{Reminder: time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)},
		},
	}
	newTask := &Task{Done: true}

	start := time.Now()
	updateDone(oldTask, newTask, config.GetTimeZone())
	elapsed := time.Since(start)

	require.Less(t, elapsed, time.Second, "updateDone must not take seconds for ancient due dates")
	assert.True(t, newTask.DueDate.After(start), "new due date must be strictly after now")
	assert.True(t, newTask.StartDate.After(start), "new start date must be strictly after now")
	assert.True(t, newTask.EndDate.After(start), "new end date must be strictly after now")
	assert.False(t, newTask.Done, "repeating task should be unmarked as done")
}

func TestUpdateDone_RRuleAncientDueDate_SlowPath(t *testing.T) {
	// A daily rule with a By-part takes the rrule.After slow path (no fixed
	// duration). Anchored at the oldest due date the horizon still evaluates
	// (200 years back), clampRRuleAnchor advancing the DTSTART keeps it cheap.
	oldTask := &Task{
		Done:    false,
		Repeats: "FREQ=DAILY;BYHOUR=9",
		DueDate: func() time.Time {
			n := time.Now().UTC()
			return time.Date(n.Year()-200, n.Month(), n.Day()+1, 0, 0, 0, 0, time.UTC)
		}(),
	}
	newTask := &Task{Done: true}

	start := time.Now()
	updateDone(oldTask, newTask, config.GetTimeZone())
	elapsed := time.Since(start)

	// FREQ=DAILY;BYHOUR=9 fires every day at 09:00:00 UTC, so the next occurrence
	// is the first 09:00 UTC strictly after now.
	su := start.UTC()
	expected := time.Date(su.Year(), su.Month(), su.Day(), 9, 0, 0, 0, time.UTC)
	if !expected.After(start) {
		expected = expected.AddDate(0, 0, 1)
	}

	require.Less(t, elapsed, time.Second, "slow-path reschedule must stay bounded for ancient due dates")
	assert.True(t, newTask.DueDate.Equal(expected), "expected next due date %v, got %v", expected, newTask.DueDate.UTC())
	assert.False(t, newTask.Done, "repeating task should be unmarked as done")
}

func TestUpdateDone_RRuleSlowPathPreservesPhase(t *testing.T) {
	// clampRRuleAnchor advances the DTSTART by whole periods; the resulting next
	// occurrence must match what the un-advanced (recent enough to iterate
	// correctly) rule yields for the same phase. INTERVAL=3 exercises the period
	// arithmetic; the 08:30 anchor minute proves wall-clock time is preserved.
	const repeats = "FREQ=DAILY;INTERVAL=3;BYHOUR=9"
	anchor := time.Date(2000, 3, 7, 8, 30, 0, 0, time.UTC)

	rule, err := rrule.StrToRRule(repeats)
	require.NoError(t, err)
	rule.DTStart(anchor)
	expected := rule.After(time.Now(), false)
	require.False(t, expected.IsZero())

	oldTask := &Task{Done: false, Repeats: repeats, DueDate: anchor}
	newTask := &Task{Done: true}
	updateDone(oldTask, newTask, config.GetTimeZone())

	assert.True(t, newTask.DueDate.Equal(expected), "clamped anchor changed the occurrence: expected %v, got %v", expected.UTC(), newTask.DueDate.UTC())
	assert.False(t, newTask.Done)
}

func TestUpdateDone_FieldScopedRepeatPersists(t *testing.T) {
	// Completing a repeating task through a field-scoped update (e.g. the bulk
	// endpoint sending only `done`) must still persist the reschedule: the new
	// due date and the reopened done state, not just done=false at the old date.
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()
	u := &user.User{ID: 1}

	// Fixture task 28 repeats hourly with a due date back in 2018.
	orig := &Task{}
	_, err := s.ID(28).Get(orig)
	require.NoError(t, err)
	require.False(t, orig.Done)

	task := &Task{ID: 28, ProjectID: 1, Done: true}
	require.NoError(t, task.updateSingleTask(s, u, []string{"done"}))

	stored := &Task{}
	_, err = s.ID(28).Get(stored)
	require.NoError(t, err)
	assert.False(t, stored.Done, "repeating task reopens for its next occurrence")
	assert.True(t, stored.DueDate.After(time.Now()), "rescheduled due date must persist, not stay at the old date")
	assert.True(t, stored.DueDate.After(orig.DueDate))
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

func TestGetTasksByUIDs(t *testing.T) {
	t.Run("returns task for authorized user", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		owner := &user.User{ID: 15}
		tasks, err := GetTasksByUIDs(s, []string{"uid-caldav-test"}, owner)
		require.NoError(t, err)
		require.Len(t, tasks, 1)
		assert.Equal(t, int64(40), tasks[0].ID)
		assert.Equal(t, "Title Caldav Test", tasks[0].Title)
	})

	t.Run("does not return task for unauthorized user", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// user 6 has no access to project 36 where uid-caldav-test lives
		outsider := &user.User{ID: 6}
		tasks, err := GetTasksByUIDs(s, []string{"uid-caldav-test"}, outsider)
		require.NoError(t, err)
		assert.Empty(t, tasks, "unauthorized user must not receive tasks by UID")
	})

	t.Run("mixed authorized and unauthorized UIDs returns only authorized", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Give task 1 (project 1, owned by user 1) a UID so we can look it up.
		_, err := s.ID(1).Cols("uid").Update(&Task{UID: "uid-user1-test"})
		require.NoError(t, err)

		user1 := &user.User{ID: 1}
		tasks, err := GetTasksByUIDs(s, []string{"uid-user1-test", "uid-caldav-test"}, user1)
		require.NoError(t, err)
		require.Len(t, tasks, 1)
		assert.Equal(t, int64(1), tasks[0].ID, "only user 1's task should be returned")
	})
}

func TestGetTaskByProjectAndIndex(t *testing.T) {
	t.Run("existing task", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task, err := GetTaskByProjectAndIndex(s, 1, 1)
		require.NoError(t, err)
		assert.Equal(t, int64(1), task.ID)
		assert.Equal(t, "task #1", task.Title)
	})

	t.Run("nonexistent index", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := GetTaskByProjectAndIndex(s, 1, 99999)
		require.Error(t, err)
		assert.True(t, IsErrTaskDoesNotExist(err))
	})

	t.Run("wrong project", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Project 4 has no tasks at all.
		_, err := GetTaskByProjectAndIndex(s, 4, 1)
		require.Error(t, err)
		assert.True(t, IsErrTaskDoesNotExist(err))
	})

	t.Run("index exists only in another project", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Project 2 has indexes 1 and 2; index 5 lives under project 1 (task 5).
		// A non-scoped WHERE clause would leak task 5 here.
		_, err := GetTaskByProjectAndIndex(s, 2, 5)
		require.Error(t, err)
		assert.True(t, IsErrTaskDoesNotExist(err))
	})

	t.Run("invalid input", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := GetTaskByProjectAndIndex(s, 0, 1)
		require.Error(t, err)
		assert.True(t, IsErrTaskDoesNotExist(err))

		_, err = GetTaskByProjectAndIndex(s, 1, 0)
		require.Error(t, err)
		assert.True(t, IsErrTaskDoesNotExist(err))
	})
}

func TestTaskIndexUniqueConstraint(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	// (project_id=1, index=1) is already taken by task 1 in fixtures.
	_, err := s.Insert(&Task{
		Title:       "duplicate index",
		ProjectID:   1,
		Index:       1,
		CreatedByID: 1,
	})
	require.Error(t, err, "unique constraint on (project_id, index) must reject duplicates")
}

func TestGetTaskByIDSimpleMemo(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	t.Cleanup(func() { db.LoadAndAssertFixtures(t) })
	s := db.NewSession()
	defer s.Close()
	require.NoError(t, s.Commit())

	first, err := GetTaskByIDSimple(s, 1)
	require.NoError(t, err)
	assert.Equal(t, "task #1", first.Title)
	first.Title = "mutated"

	updateTitleBehindTheBack(t, 1, &Task{Title: behindTheBackTitle})

	second, err := GetTaskByIDSimple(s, 1)
	require.NoError(t, err)
	assert.Equal(t, "task #1", second.Title, "a re-read that reaches the db would see the other session's write")

	_, err = s.ID(2).Cols("title").Update(&Task{Title: "any write invalidates the memo"})
	require.NoError(t, err)

	afterWrite, err := GetTaskByIDSimple(s, 1)
	require.NoError(t, err)
	assert.Equal(t, behindTheBackTitle, afterWrite.Title)
}

// withFixedNow pins the recurrence engine's clock for one test.
func withFixedNow(t *testing.T, now time.Time) {
	t.Helper()
	orig := timeNow
	timeNow = func() time.Time { return now }
	t.Cleanup(func() { timeNow = orig })
}

// completeRepeating marks a copy of old done, as Update does, and returns the result. The rule is
// copied onto the new task first, as the write path resolves it before completion runs.
func completeRepeating(old *Task) *Task { return completeRepeatingIn(old, config.GetTimeZone()) }

// completeRepeatingIn is completeRepeating with an explicit completion zone.
func completeRepeatingIn(old *Task, zone *time.Location) *Task {
	n := &Task{Done: true, Repeats: old.Repeats}
	updateDone(old, n, zone)
	return n
}

func reminderAt(at time.Time) *TaskReminder {
	return &TaskReminder{Reminder: at}
}

// reminderTime returns the i-th reminder of n, failing the test when it is missing.
func reminderTime(t *testing.T, n *Task, i int) time.Time {
	t.Helper()
	require.Greater(t, len(n.Reminders), i, "reminder %d is missing", i)
	return n.Reminders[i].Reminder
}

func TestShiftTime(t *testing.T) {
	t.Run("short span moves by the exact offset", func(t *testing.T) {
		from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		got := shiftTime(from.Add(90*time.Minute), from, from.Add(36*time.Hour))
		assert.True(t, got.Equal(from.Add(90*time.Minute+36*time.Hour)), "got %v", got)
	})
	t.Run("span over 292 years forward does not saturate", func(t *testing.T) {
		from := time.Date(1700, 1, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
		got := shiftTime(time.Date(1700, 1, 2, 0, 0, 0, 0, time.UTC), from, to)
		assert.True(t, got.Equal(time.Date(2026, 1, 2, 10, 0, 0, 0, time.UTC)), "got %v", got)
	})
	t.Run("span over 292 years backward does not saturate", func(t *testing.T) {
		from := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
		to := time.Date(1700, 1, 1, 0, 0, 0, 0, time.UTC)
		got := shiftTime(time.Date(2026, 1, 2, 10, 0, 0, 0, time.UTC), from, to)
		assert.True(t, got.Equal(time.Date(1700, 1, 2, 0, 0, 0, 0, time.UTC)), "got %v", got)
	})
}

// TestSetTaskDatesRRule pins the completion engine against upstream's legacy setTaskDatesDefault /
// setTaskDatesMonthRepeat / setTaskDatesFromCurrentDateRepeat. Their parity still governs which
// dates move without a due date (the no-due-date policy); DAILY and WEEKLY rules now step on the
// completion zone's wall clock, not by 24 elapsed hours. The "fixed-duration" subtest names
// predate that change and are kept: they are the policy ratchet.
func TestSetTaskDatesRRule(t *testing.T) {
	// Thursday, Oct 22 2026, 12:00 UTC.
	now := time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC)
	at := func(month time.Month, day, hour, minute int) time.Time {
		return time.Date(2026, month, day, hour, minute, 0, 0, time.UTC)
	}

	t.Run("normal mode, no due date", func(t *testing.T) {
		t.Run("fixed-duration: each date independently jumps past now", func(t *testing.T) {
			withFixedNow(t, now)
			n := completeRepeating(&Task{
				Repeats:   "FREQ=DAILY;INTERVAL=1",
				StartDate: at(10, 20, 9, 0),
				EndDate:   at(10, 23, 10, 0),
				Reminders: []*TaskReminder{reminderAt(at(10, 21, 8, 0))},
			})
			assert.False(t, n.Done)
			assert.True(t, n.DueDate.IsZero(), "no due date is invented")
			assert.True(t, n.StartDate.Equal(at(10, 23, 9, 0)), "start: got %v", n.StartDate)
			assert.True(t, n.EndDate.Equal(at(10, 24, 10, 0)), "end (already future) moves one step: got %v", n.EndDate)
			assert.True(t, reminderTime(t, n, 0).Equal(at(10, 23, 8, 0)), "reminder: got %v", reminderTime(t, n, 0))
		})
		t.Run("calendar: each date takes exactly one rule step, end keeps its gap to start", func(t *testing.T) {
			withFixedNow(t, now)
			n := completeRepeating(&Task{
				Repeats:   "FREQ=MONTHLY;INTERVAL=1",
				StartDate: at(1, 10, 9, 0),
				EndDate:   at(1, 12, 17, 0),
				Reminders: []*TaskReminder{reminderAt(at(1, 15, 8, 0))},
			})
			assert.False(t, n.Done)
			// One step only, even though the result is still before now (upstream's monthly path).
			assert.True(t, n.StartDate.Equal(at(2, 10, 9, 0)), "start: got %v", n.StartDate)
			assert.True(t, n.EndDate.Equal(at(2, 12, 17, 0)), "end: got %v", n.EndDate)
			assert.True(t, reminderTime(t, n, 0).Equal(at(2, 15, 8, 0)), "reminder: got %v", reminderTime(t, n, 0))
		})
		t.Run("calendar: the end keeps its gap to the start instead of stepping on its own", func(t *testing.T) {
			withFixedNow(t, now)
			// MONTHLY from Jan 30 skips February (no 30th) to Mar 30; on its own the end would
			// step Feb 1 -> Mar 1, but it keeps its two-day gap to the start instead.
			n := completeRepeating(&Task{Repeats: "FREQ=MONTHLY", StartDate: at(1, 30, 9, 0), EndDate: at(2, 1, 9, 0)})
			assert.True(t, n.StartDate.Equal(at(3, 30, 9, 0)), "start: got %v", n.StartDate)
			assert.True(t, n.EndDate.Equal(at(4, 1, 9, 0)), "end: got %v", n.EndDate)
		})
		t.Run("calendar: migrated month-end rule steps a Jan 31 reminder to Feb 28", func(t *testing.T) {
			withFixedNow(t, now)
			n := completeRepeating(&Task{
				Repeats:   "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=28,29,30,31;BYSETPOS=-1",
				Reminders: []*TaskReminder{reminderAt(at(1, 31, 10, 0))},
			})
			assert.False(t, n.Done)
			assert.True(t, reminderTime(t, n, 0).Equal(at(2, 28, 10, 0)), "reminder: got %v", reminderTime(t, n, 0))
		})
		t.Run("calendar: only an end date takes one step", func(t *testing.T) {
			withFixedNow(t, now)
			n := completeRepeating(&Task{Repeats: "FREQ=MONTHLY;INTERVAL=1", EndDate: at(3, 5, 9, 0)})
			assert.True(t, n.EndDate.Equal(at(4, 5, 9, 0)), "end: got %v", n.EndDate)
			assert.True(t, n.StartDate.IsZero())
		})
		t.Run("finite: a COUNT rule takes one rule step and is never caught up like a fixed duration", func(t *testing.T) {
			withFixedNow(t, now)
			n := completeRepeating(&Task{Repeats: "FREQ=DAILY;COUNT=3", StartDate: at(10, 17, 12, 0)})
			assert.False(t, n.Done)
			assert.True(t, n.StartDate.Equal(at(10, 18, 12, 0)), "start: got %v", n.StartDate)
			o, err := rrule.StrToROption(n.Repeats)
			require.NoError(t, err)
			assert.Equal(t, 2, o.Count)
		})
		t.Run("no dates at all reopens the task with nothing to move (upstream parity)", func(t *testing.T) {
			withFixedNow(t, now)
			n := completeRepeating(&Task{Repeats: "FREQ=DAILY;INTERVAL=1"})
			assert.False(t, n.Done)
			assert.True(t, n.DueDate.IsZero())
			assert.True(t, n.StartDate.IsZero())
			assert.True(t, n.EndDate.IsZero())
		})
	})

	t.Run("from current date, no due date", func(t *testing.T) {
		t.Run("earliest absolute reminder lands on now + one step, the others keep their spacing", func(t *testing.T) {
			withFixedNow(t, now)
			n := completeRepeating(&Task{
				Repeats:                "FREQ=DAILY;INTERVAL=1",
				RepeatsFromCurrentDate: true,
				StartDate:              at(10, 1, 9, 0),
				EndDate:                at(10, 3, 9, 0),
				// The later reminder comes first: the earliest is found, not assumed.
				Reminders: []*TaskReminder{reminderAt(at(10, 1, 10, 0)), reminderAt(at(10, 1, 8, 0))},
			})
			assert.False(t, n.Done)
			next := now.Add(24 * time.Hour)
			assert.True(t, reminderTime(t, n, 1).Equal(next), "earliest reminder: got %v", reminderTime(t, n, 1))
			// A companion moves by the reference's whole local days (Oct 1 -> Oct 23) and keeps its
			// own clock time, 10:00, instead of keeping its elapsed spacing to the re-based reminder
			// (it was next + 2 h).
			assert.True(t, reminderTime(t, n, 0).Equal(at(10, 23, 10, 0)), "other reminder: got %v", reminderTime(t, n, 0))
			assert.True(t, n.StartDate.Equal(next), "start: got %v", n.StartDate)
			// Likewise, the end moves by the start's 22 days (Oct 1 -> Oct 23) and keeps its
			// 09:00 (it was next + 48 h, at the start's new clock).
			assert.True(t, n.EndDate.Equal(at(10, 25, 9, 0)), "end keeps its whole-day gap to start: got %v", n.EndDate)
		})
		t.Run("calendar rule: next is the first occurrence after now", func(t *testing.T) {
			withFixedNow(t, now)
			n := completeRepeating(&Task{
				Repeats:                "FREQ=WEEKLY;BYDAY=MO",
				RepeatsFromCurrentDate: true,
				Reminders:              []*TaskReminder{reminderAt(at(9, 1, 8, 0))},
			})
			assert.True(t, reminderTime(t, n, 0).Equal(at(10, 26, 12, 0)), "next Monday from now: got %v", reminderTime(t, n, 0))
		})
		t.Run("only an end date moves to now + one step", func(t *testing.T) {
			withFixedNow(t, now)
			n := completeRepeating(&Task{Repeats: "FREQ=DAILY", RepeatsFromCurrentDate: true, EndDate: at(9, 1, 8, 0)})
			assert.True(t, n.EndDate.Equal(now.Add(24*time.Hour)), "end: got %v", n.EndDate)
		})
		t.Run("relative reminders don't take part in the anchoring", func(t *testing.T) {
			withFixedNow(t, now)
			relative := &TaskReminder{RelativeTo: ReminderRelationStartDate, RelativePeriod: -3600, Reminder: at(9, 1, 8, 0)}
			n := completeRepeating(&Task{
				Repeats:                "FREQ=DAILY",
				RepeatsFromCurrentDate: true,
				StartDate:              at(10, 1, 9, 0),
				Reminders:              []*TaskReminder{relative, reminderAt(at(10, 1, 10, 0))},
			})
			// The stale value of the relative reminder is earlier, but only absolute ones anchor.
			assert.True(t, reminderTime(t, n, 1).Equal(now.Add(24*time.Hour)), "absolute reminder: got %v", reminderTime(t, n, 1))
		})
		t.Run("no dates at all reopens the task with nothing to move", func(t *testing.T) {
			withFixedNow(t, now)
			n := completeRepeating(&Task{Repeats: "FREQ=DAILY", RepeatsFromCurrentDate: true})
			assert.False(t, n.Done)
			assert.True(t, n.DueDate.IsZero())
			assert.True(t, n.StartDate.IsZero())
			assert.True(t, n.EndDate.IsZero())
			assert.Empty(t, n.Reminders)
		})
	})

	t.Run("with a due date", func(t *testing.T) {
		// The overflow coverage is TestShiftTime and the from-current-date test of reminders
		// more than 292 years apart; this one checks that dates far before a recent due date keep
		// their offset.
		t.Run("start and reminder far before the due date keep their offset", func(t *testing.T) {
			withFixedNow(t, now)
			due := at(10, 20, 9, 0)
			start := time.Date(1700, 1, 1, 9, 0, 0, 0, time.UTC)
			reminder := time.Date(1700, 1, 1, 7, 0, 0, 0, time.UTC)
			n := completeRepeating(&Task{
				Repeats:   "FREQ=DAILY;BYHOUR=9",
				DueDate:   due,
				StartDate: start,
				Reminders: []*TaskReminder{reminderAt(reminder)},
			})
			require.False(t, n.Done)
			assert.True(t, n.DueDate.Equal(at(10, 23, 9, 0)), "due: got %v", n.DueDate)
			assert.True(t, n.StartDate.Equal(start.AddDate(0, 0, 3)), "start keeps its offset: got %v", n.StartDate)
			assert.True(t, reminderTime(t, n, 0).Equal(reminder.AddDate(0, 0, 3)), "reminder keeps its offset: got %v", reminderTime(t, n, 0))
		})
		t.Run("a due date more than 200 years ago is beyond the horizon: the task stays done", func(t *testing.T) {
			// Nothing looks back further than 200 years.
			withFixedNow(t, now)
			n := completeRepeating(&Task{Repeats: "FREQ=DAILY;INTERVAL=3;BYHOUR=9", DueDate: time.Date(1700, 1, 1, 9, 0, 0, 0, time.UTC)})
			assert.True(t, n.Done)
			assert.True(t, n.DueDate.IsZero(), "dates unchanged")
		})
		t.Run("a DAILY rule anchored in 1830 reschedules quickly and to the right phase", func(t *testing.T) {
			withFixedNow(t, now)
			anchor := time.Date(1830, 1, 1, 9, 0, 0, 0, time.UTC)
			// INTERVAL=3: the next occurrence is the first 09:00 strictly after now whose day
			// distance from the anchor is a multiple of 3.
			daysSinceAnchor := int(time.Date(2026, 10, 22, 0, 0, 0, 0, time.UTC).Unix()-time.Date(1830, 1, 1, 0, 0, 0, 0, time.UTC).Unix()) / 86400
			k := (daysSinceAnchor + 1 + 2) / 3 * 3
			expected := time.Date(1830, 1, 1+k, 9, 0, 0, 0, time.UTC)

			start := time.Now()
			n := completeRepeating(&Task{Repeats: "FREQ=DAILY;INTERVAL=3;BYHOUR=9", DueDate: anchor})
			elapsed := time.Since(start)

			assert.Less(t, elapsed, 50*time.Millisecond)
			assert.True(t, n.DueDate.Equal(expected), "expected %v, got %v", expected, n.DueDate)
		})
		t.Run("migrated month-end rule from Jan 31 goes to Feb 28, then Mar 31", func(t *testing.T) {
			withFixedNow(t, at(1, 15, 12, 0))
			const monthEnd = "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=28,29,30,31;BYSETPOS=-1"
			n1 := completeRepeating(&Task{Repeats: monthEnd, DueDate: at(1, 31, 10, 0)})
			assert.True(t, n1.DueDate.Equal(at(2, 28, 10, 0)), "first: got %v", n1.DueDate)
			n2 := completeRepeating(&Task{Repeats: monthEnd, DueDate: n1.DueDate})
			assert.True(t, n2.DueDate.Equal(at(3, 31, 10, 0)), "second: got %v", n2.DueDate)
		})
		t.Run("new FREQ=MONTHLY from Jan 31 skips short months", func(t *testing.T) {
			withFixedNow(t, at(1, 15, 12, 0))
			n := completeRepeating(&Task{Repeats: "FREQ=MONTHLY", DueDate: at(1, 31, 10, 0)})
			assert.True(t, n.DueDate.Equal(at(3, 31, 10, 0)), "got %v", n.DueDate)
		})
		t.Run("COUNT=1 stays done and clears the rule", func(t *testing.T) {
			withFixedNow(t, now)
			n := completeRepeating(&Task{Repeats: "FREQ=DAILY;COUNT=1", DueDate: now.Add(time.Hour)})
			assert.True(t, n.Done)
			assert.Empty(t, n.Repeats)
		})
		t.Run("COUNT=3 becomes COUNT=2 after completion", func(t *testing.T) {
			withFixedNow(t, now)
			n := completeRepeating(&Task{Repeats: "FREQ=DAILY;COUNT=3", DueDate: now.Add(time.Hour)})
			assert.False(t, n.Done)
			assert.True(t, n.DueDate.Equal(now.Add(25*time.Hour)), "got %v", n.DueDate)
			o, err := rrule.StrToROption(n.Repeats)
			require.NoError(t, err)
			assert.Equal(t, 2, o.Count)
		})
		t.Run("UNTIL exactly at the next occurrence reschedules", func(t *testing.T) {
			withFixedNow(t, now)
			n := completeRepeating(&Task{Repeats: "FREQ=DAILY;UNTIL=20261023T130000Z", DueDate: at(10, 22, 13, 0)})
			assert.False(t, n.Done)
			assert.True(t, n.DueDate.Equal(at(10, 23, 13, 0)), "got %v", n.DueDate)
		})
		t.Run("UNTIL one second before the next occurrence stays done", func(t *testing.T) {
			withFixedNow(t, now)
			old := &Task{Repeats: "FREQ=DAILY;UNTIL=20261023T125959Z", DueDate: at(10, 22, 13, 0)}
			n := completeRepeating(old)
			assert.True(t, n.Done)
			assert.Equal(t, old.Repeats, n.Repeats, "the rule is kept")
			assert.True(t, n.DueDate.IsZero(), "nothing is rescheduled")
		})
		t.Run("DST: a calendar DAILY rule keeps the wall-clock time across spring-forward", func(t *testing.T) {
			// The rule is evaluated in the completion's zone, so Berlin is passed explicitly; the
			// expected value is unchanged.
			berlin, err := time.LoadLocation("Europe/Berlin")
			require.NoError(t, err)
			withFixedNow(t, time.Date(2026, 3, 28, 8, 0, 0, 0, berlin))
			n := completeRepeatingIn(&Task{Repeats: "FREQ=DAILY;BYHOUR=9", DueDate: time.Date(2026, 3, 28, 9, 0, 0, 0, berlin)}, berlin)
			assert.True(t, n.DueDate.Equal(time.Date(2026, 3, 29, 9, 0, 0, 0, berlin)), "got %v", n.DueDate.In(berlin))
		})
		t.Run("DST: a plain DAILY rule keeps the wall-clock time across spring-forward", func(t *testing.T) {
			// FREQ=DAILY without a BY* part steps on the completion zone's wall clock, not by 24
			// elapsed hours, so 09:00 CET is followed by 09:00 CEST (it used to be 10:00 CEST).
			// Sub-daily rules stay elapsed.
			berlin, err := time.LoadLocation("Europe/Berlin")
			require.NoError(t, err)
			withFixedNow(t, time.Date(2026, 3, 28, 8, 0, 0, 0, berlin))
			n := completeRepeatingIn(&Task{Repeats: "FREQ=DAILY", DueDate: time.Date(2026, 3, 28, 9, 0, 0, 0, berlin)}, berlin)
			assert.True(t, n.DueDate.Equal(time.Date(2026, 3, 29, 9, 0, 0, 0, berlin)), "got %v", n.DueDate.In(berlin))
		})
		t.Run("leap day: YEARLY from Feb 29 skips to the next Feb 29", func(t *testing.T) {
			// rrule-go follows RFC 5545: an invalid date (Feb 29 in a common year) is skipped.
			withFixedNow(t, time.Date(2028, 2, 1, 12, 0, 0, 0, time.UTC))
			n := completeRepeating(&Task{Repeats: "FREQ=YEARLY", DueDate: time.Date(2028, 2, 29, 9, 0, 0, 0, time.UTC)})
			assert.True(t, n.DueDate.Equal(time.Date(2032, 2, 29, 9, 0, 0, 0, time.UTC)), "got %v", n.DueDate)
		})
	})

	t.Run("UNTIL already passed leaves a task without due date done", func(t *testing.T) {
		withFixedNow(t, now)
		old := &Task{Repeats: "FREQ=DAILY;UNTIL=20260101T000000Z", StartDate: time.Date(2025, 11, 1, 9, 0, 0, 0, time.UTC)}
		n := completeRepeating(old)
		assert.True(t, n.Done)
		assert.Equal(t, old.Repeats, n.Repeats, "the rule is kept")
		assert.True(t, n.StartDate.IsZero(), "nothing is rescheduled")
	})
}

func TestClampRRuleAnchor(t *testing.T) {
	t.Run("the day gap does not saturate past 292 years", func(t *testing.T) {
		a := time.Date(1700, 1, 1, 9, 0, 0, 0, time.UTC)
		b := time.Date(2025, 10, 22, 12, 0, 0, 0, time.UTC)
		// 325 years from 1700-01-01 to 2025-01-01 hold 79 leap days (1700, 1800 and 1900 are not leap years).
		expected := 325*365 + 79 + (time.Date(2025, 10, 22, 0, 0, 0, 0, time.UTC).YearDay() - 1)
		assert.Equal(t, expected, daysBetween(a, b))
	})
	t.Run("an anchor two millennia back lands within one period below the cutoff, quickly", func(t *testing.T) {
		// The clamp runs on the floating timeline: the zone's wall clock read as
		// UTC, so New York's DST never reaches it and its day count is exact.
		ny, err := time.LoadLocation("America/New_York")
		require.NoError(t, err)
		now := float(time.Date(2026, 10, 22, 12, 0, 0, 0, ny), ny)
		cutoff := now.AddDate(-1, 0, 0)
		anchor := float(time.Date(1, 1, 1, 9, 0, 0, 0, ny), ny)
		require.Equal(t, time.UTC, anchor.Location(), "a floating value is UTC-located")
		opt := &rrule.ROption{Freq: rrule.WEEKLY, Interval: 1}

		start := time.Now()
		got := clampRRuleAnchor(opt, anchor, now)
		elapsed := time.Since(start)

		// A loop over the ~105k skipped weeks takes far longer; 50ms leaves room for a loaded CI host.
		assert.Less(t, elapsed, 50*time.Millisecond)
		assert.False(t, got.After(cutoff), "got %v", got)
		assert.True(t, got.AddDate(0, 0, 7).After(cutoff), "got %v", got)
		assert.Equal(t, anchor.Weekday(), got.Weekday(), "phase is preserved")
		assert.Equal(t, 9, got.Hour())
	})
}

// repeatingFixtureTask is fixture task #28, which repeats hourly.
const repeatingFixtureTask int64 = 28

// storeRepeats writes a rule straight onto the repeating fixture task, bypassing the write path's
// validation.
func storeRepeats(t *testing.T, s *xorm.Session, repeats string) {
	t.Helper()
	_, err := s.ID(repeatingFixtureTask).Cols("repeats").NoAutoTime().Update(&Task{Repeats: repeats})
	require.NoError(t, err)
}

func storedRepeats(t *testing.T, s *xorm.Session, taskID int64) string {
	t.Helper()
	stored := &Task{}
	has, err := s.ID(taskID).Get(stored)
	require.NoError(t, err)
	require.True(t, has)
	return stored.Repeats
}

// readForEcho reads the repeating fixture task as the API does and drops the raw rule, as JSON
// decoding would (Repeats is json:"-"), so the result is what a client sends back.
func readForEcho(t *testing.T, s *xorm.Session, u *user.User) *Task {
	t.Helper()
	read := &Task{ID: repeatingFixtureTask}
	require.NoError(t, read.ReadOne(s, u))
	read.Repeats = ""
	return read
}

// TestTask_RecurrenceWritePaths pins the recurrence write rules at model level: create, bulk
// create and update all decide the stored rule through resolveIncomingRepeat.
func TestTask_RecurrenceWritePaths(t *testing.T) {
	u := &user.User{ID: 1}

	t.Run("update echoing the read repeat keeps the stored text byte-identical", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		// Non-canonical part order: re-serialising the projection would rewrite it.
		storeRepeats(t, s, "INTERVAL=1;FREQ=HOURLY")

		upd := readForEcho(t, s, u)
		require.NotNil(t, upd.Repeat)
		upd.Title = "renamed"
		require.NoError(t, upd.Update(s, u))

		assert.Equal(t, "INTERVAL=1;FREQ=HOURLY", storedRepeats(t, s, repeatingFixtureTask))
		assert.Equal(t, "renamed", upd.Title)
	})
	t.Run("fixture task 28 survives an echo unchanged", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		upd := readForEcho(t, s, u)
		upd.Title = "renamed"
		require.NoError(t, upd.Update(s, u))

		assert.Equal(t, "FREQ=HOURLY;INTERVAL=1", storedRepeats(t, s, repeatingFixtureTask))
	})
	t.Run("a read-only rule is kept on echo, refused on edit and cleared on null", func(t *testing.T) {
		const easter = "FREQ=YEARLY;BYEASTER=0"

		t.Run("echo keeps it", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()
			storeRepeats(t, s, easter)

			upd := readForEcho(t, s, u)
			require.NotNil(t, upd.Repeat)
			require.True(t, upd.Repeat.ReadOnly)
			upd.Title = "renamed"
			require.NoError(t, upd.Update(s, u))

			assert.Equal(t, easter, storedRepeats(t, s, repeatingFixtureTask))
			require.NotNil(t, upd.Repeat, "the response carries the projection")
			assert.True(t, upd.Repeat.ReadOnly, "the response marks it read-only")
		})
		t.Run("changing the interval is refused", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()
			storeRepeats(t, s, easter)

			upd := readForEcho(t, s, u)
			upd.Repeat.Interval = 2
			err := upd.Update(s, u)
			require.Error(t, err)
			assert.True(t, IsErrInvalidData(err), "got %v", err)
			assert.Equal(t, easter, storedRepeats(t, s, repeatingFixtureTask))
		})
		t.Run("repeat nil clears it", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()
			storeRepeats(t, s, easter)

			upd := readForEcho(t, s, u)
			upd.Repeat = nil
			require.NoError(t, upd.Update(s, u))

			assert.Empty(t, storedRepeats(t, s, repeatingFixtureTask))
			assert.Nil(t, upd.Repeat)
		})
	})
	t.Run("a different valid repeat replaces the rule", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		upd := readForEcho(t, s, u)
		upd.Repeat = &TaskRepeat{Freq: "weekly", Interval: 2}
		require.NoError(t, upd.Update(s, u))

		assert.Equal(t, "FREQ=WEEKLY;INTERVAL=2", storedRepeats(t, s, repeatingFixtureTask))
	})
	t.Run("masked update excluding recurrence never re-validates the stored rule", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		// Rejected by today's policy (sub-daily with a BY* part); it may predate the policy.
		const legacy = "FREQ=HOURLY;BYHOUR=9"
		storeRepeats(t, s, legacy)

		upd := &Task{ID: repeatingFixtureTask, Title: "renamed only"}
		require.NoError(t, upd.updateSingleTask(s, u, []string{"title"}))

		assert.Equal(t, legacy, storedRepeats(t, s, repeatingFixtureTask))
		assert.Equal(t, legacy, upd.Repeats, "the response reflects the stored rule")
	})
	t.Run("masked update ignores a repeat in the body when the mask excludes it", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		upd := &Task{ID: repeatingFixtureTask, Title: "renamed only", Repeat: &TaskRepeat{Freq: "daily", Interval: 1}}
		require.NoError(t, upd.updateSingleTask(s, u, []string{"title"}))

		assert.Equal(t, "FREQ=HOURLY;INTERVAL=1", storedRepeats(t, s, repeatingFixtureTask))
		assert.Equal(t, "FREQ=HOURLY;INTERVAL=1", upd.Repeats, "the response must not show a rule that was not stored")
	})
	t.Run("masked update including repeat applies the write rules", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		storeRepeats(t, s, "INTERVAL=1;FREQ=HOURLY")

		upd := &Task{ID: repeatingFixtureTask, Repeat: &TaskRepeat{Freq: "hourly", Interval: 1}}
		require.NoError(t, upd.updateSingleTask(s, u, []string{"repeat"}))
		assert.Equal(t, "INTERVAL=1;FREQ=HOURLY", storedRepeats(t, s, repeatingFixtureTask), "an equal echo keeps the text")

		upd = &Task{ID: repeatingFixtureTask}
		require.NoError(t, upd.updateSingleTask(s, u, []string{"repeat"}))
		assert.Empty(t, storedRepeats(t, s, repeatingFixtureTask), "repeat absent from the body but in the mask clears it")
	})
	t.Run("bulk create stores a raw rule verbatim", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		tasks := []*Task{
			{Title: "easter", Repeats: "FREQ=YEARLY;BYEASTER=0"},
			{Title: "reordered", Repeats: "INTERVAL=2;FREQ=DAILY"},
			{Title: "structured", Repeat: &TaskRepeat{Freq: "weekly", Interval: 1}},
		}
		require.NoError(t, createTasks(s, 1, tasks, u, true, true, false))

		assert.Equal(t, "FREQ=YEARLY;BYEASTER=0", storedRepeats(t, s, tasks[0].ID))
		assert.Equal(t, "INTERVAL=2;FREQ=DAILY", storedRepeats(t, s, tasks[1].ID))
		assert.Equal(t, "FREQ=WEEKLY;INTERVAL=1", storedRepeats(t, s, tasks[2].ID))
		require.NotNil(t, tasks[0].Repeat, "the response carries the projection")
		assert.True(t, tasks[0].Repeat.ReadOnly)
	})
	t.Run("bulk create reports the index of a raw rule its structured repeat contradicts", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		tasks := []*Task{
			{Title: "fine", Repeats: "FREQ=DAILY"},
			{Title: "contradicting", Repeats: "FREQ=DAILY", Repeat: &TaskRepeat{Freq: "weekly", Interval: 1}},
		}
		err := createTasks(s, 1, tasks, u, true, true, false)
		require.Error(t, err)
		var berr ErrInvalidTaskInBulkCreation
		require.ErrorAs(t, err, &berr)
		assert.Equal(t, 1, berr.Index)
		assert.True(t, IsErrInvalidData(berr.Err), "got %v", berr.Err)
	})
	t.Run("completion without a due date moves absolute reminders and recomputes relative ones", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		withFixedNow(t, time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC))

		task := &Task{
			Title:     "no due date",
			ProjectID: 1,
			Repeat:    &TaskRepeat{Freq: "daily", Interval: 1},
			StartDate: time.Date(2026, 10, 20, 9, 0, 0, 0, time.UTC),
			Reminders: []*TaskReminder{
				{RelativeTo: ReminderRelationStartDate, RelativePeriod: -3600},
				{Reminder: time.Date(2026, 10, 21, 7, 30, 0, 0, time.UTC)},
			},
		}
		require.NoError(t, task.Create(s, u))
		require.NoError(t, s.Commit())

		done := &Task{ID: task.ID, Done: true}
		require.NoError(t, done.updateSingleTask(s, u, []string{"done"}))
		require.NoError(t, s.Commit())

		read := &Task{ID: task.ID}
		require.NoError(t, read.ReadOne(s, u))
		assert.False(t, read.Done, "the task reopens")
		newStart := time.Date(2026, 10, 23, 9, 0, 0, 0, time.UTC)
		assert.True(t, read.StartDate.Equal(newStart), "start: got %v", read.StartDate)
		require.Len(t, read.Reminders, 2)
		byKind := map[bool]time.Time{}
		for _, r := range read.Reminders {
			byKind[r.RelativeTo != ""] = r.Reminder
		}
		assert.True(t, byKind[true].Equal(newStart.Add(-time.Hour)), "relative reminder follows the new start: got %v", byKind[true])
		assert.True(t, byKind[false].Equal(time.Date(2026, 10, 23, 7, 30, 0, 0, time.UTC)), "absolute reminder: got %v", byKind[false])
	})
}

// seedRepeatingTask writes a rule and a due date straight onto task 1, as a CalDAV client or
// an earlier completion would have stored them.
func seedRepeatingTask(t *testing.T, s *xorm.Session, rule string, due time.Time) {
	t.Helper()
	_, err := s.ID(1).Cols("repeats", "due_date", "done").Update(&Task{Repeats: rule, DueDate: due})
	require.NoError(t, err)
}

// storedTask1 reads task 1 as stored.
func storedTask1(t *testing.T, s *xorm.Session) *Task {
	t.Helper()
	task := &Task{}
	has, err := s.ID(1).Get(task)
	require.NoError(t, err)
	require.True(t, has)
	return task
}

// Both completion call sites compute the next occurrence in the completing user's zone,
// re-fetched from the database (the auth object is JWT-shaped, with an empty Timezone). The
// MO,WE,FR rule, due Mon 2026-10-26 20:00 EDT, completed at 20:05 EDT, lands on Wed 2026-10-28
// 20:00 EDT; in the server zone (GMT, the fixture default) it would be Wed 2026-10-28 00:00Z, a
// day early.
func TestCompletionZone(t *testing.T) {
	const rule = "FREQ=WEEKLY;BYDAY=MO,WE,FR"
	due := time.Date(2026, 10, 27, 0, 0, 0, 0, time.UTC)
	jwtUser := &user.User{ID: 1, Username: "user1"}
	wantNY := time.Date(2026, 10, 29, 0, 0, 0, 0, time.UTC)
	wantServer := time.Date(2026, 10, 28, 0, 0, 0, 0, time.UTC)

	setup := func(t *testing.T, zone string) *xorm.Session {
		t.Helper()
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		t.Cleanup(func() { s.Close() })
		withFixedNow(t, time.Date(2026, 10, 27, 0, 5, 0, 0, time.UTC))
		setUserZone(t, s, 1, zone)
		seedRepeatingTask(t, s, rule, due)
		return s
	}

	t.Run("update by a JWT-shaped user completes in the stored zone", func(t *testing.T) {
		s := setup(t, "America/New_York")
		require.NoError(t, (&Task{ID: 1, Done: true}).Update(s, jwtUser))

		task := storedTask1(t, s)
		assert.False(t, task.Done, "reopened")
		assert.Equal(t, wantNY, task.DueDate.UTC())
	})
	t.Run("field-masked update completes in the stored zone", func(t *testing.T) {
		s := setup(t, "America/New_York")
		require.NoError(t, (&Task{ID: 1, Done: true}).updateSingleTask(s, jwtUser, []string{"done"}))

		task := storedTask1(t, s)
		assert.False(t, task.Done)
		assert.Equal(t, wantNY, task.DueDate.UTC())
		assert.Equal(t, rule, task.Repeats)
	})
	t.Run("kanban done bucket completes in the stored zone", func(t *testing.T) {
		s := setup(t, "America/New_York")
		tb := &TaskBucket{TaskID: 1, BucketID: 3, ProjectViewID: 4, ProjectID: 1}
		require.NoError(t, tb.Update(s, jwtUser))

		task := storedTask1(t, s)
		assert.False(t, task.Done)
		assert.Equal(t, wantNY, task.DueDate.UTC())
	})
	t.Run("a user without a zone completes in the server zone", func(t *testing.T) {
		s := setup(t, "")
		require.NoError(t, (&Task{ID: 1, Done: true}).updateSingleTask(s, jwtUser, []string{"done"}))

		assert.Equal(t, wantServer, storedTask1(t, s).DueDate.UTC(), "positive control: the fixture discriminates")
	})
	t.Run("kanban by a user without a zone completes in the server zone", func(t *testing.T) {
		s := setup(t, "")
		tb := &TaskBucket{TaskID: 1, BucketID: 3, ProjectViewID: 4, ProjectID: 1}
		require.NoError(t, tb.Update(s, jwtUser))

		assert.Equal(t, wantServer, storedTask1(t, s).DueDate.UTC())
	})
}

func intPtr(v int) *int { return &v }

// readTaskForEcho reads a task as a client would before sending it back in a full update.
func readTaskForEcho(t *testing.T, s *xorm.Session, taskID int64, u *user.User) *Task {
	t.Helper()
	read := &Task{ID: taskID}
	require.NoError(t, read.ReadOne(s, u))
	read.Repeats = ""
	return read
}

// seedIntendedDueTime writes intended_due_time straight onto task 1, as an earlier completion would.
func seedIntendedDueTime(t *testing.T, s *xorm.Session, v *int) {
	t.Helper()
	_, err := s.ID(1).Cols("intended_due_time").Update(&Task{IntendedDueTime: v})
	require.NoError(t, err)
}

// A daily task due Sat 2027-03-13 02:30 EST, completed in New York,
// lands on Sun 03:30 EDT (02:30 does not exist that day) and remembers 02:30 as 9000 s; the next
// completion returns to Mon 02:30 EDT and stores NULL. Each chain exercises both save sites.
func TestIntendedDueTime(t *testing.T) {
	due := time.Date(2027, 3, 13, 7, 30, 0, 0, time.UTC)
	shifted := time.Date(2027, 3, 14, 7, 30, 0, 0, time.UTC)  // Sun 03:30 EDT
	restored := time.Date(2027, 3, 15, 6, 30, 0, 0, time.UTC) // Mon 02:30 EDT
	jwtUser := &user.User{ID: 1, Username: "user1"}

	setup := func(t *testing.T) *xorm.Session {
		t.Helper()
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		t.Cleanup(func() { s.Close() })
		withFixedNow(t, time.Date(2027, 3, 13, 8, 0, 0, 0, time.UTC))
		setUserZone(t, s, 1, "America/New_York")
		seedRepeatingTask(t, s, "FREQ=DAILY;INTERVAL=1", due)
		return s
	}
	patch := func(t *testing.T, s *xorm.Session) {
		t.Helper()
		require.NoError(t, (&Task{ID: 1, Done: true}).updateSingleTask(s, jwtUser, []string{"done"}))
	}
	kanban := func(t *testing.T, s *xorm.Session) {
		t.Helper()
		require.NoError(t, (&TaskBucket{TaskID: 1, BucketID: 3, ProjectViewID: 4, ProjectID: 1}).Update(s, jwtUser))
	}
	steps := map[string]func(*testing.T, *xorm.Session){"update": patch, "kanban": kanban}

	for _, chain := range [][2]string{{"update", "update"}, {"update", "kanban"}, {"kanban", "update"}, {"kanban", "kanban"}} {
		t.Run(chain[0]+" then "+chain[1], func(t *testing.T) {
			s := setup(t)

			steps[chain[0]](t, s)
			task := storedTask1(t, s)
			assert.False(t, task.Done)
			assert.Equal(t, shifted, task.DueDate.UTC())
			require.NotNil(t, task.IntendedDueTime, "the gap shift is remembered")
			assert.Equal(t, 9000, *task.IntendedDueTime)

			steps[chain[1]](t, s)
			task = storedTask1(t, s)
			assert.Equal(t, restored, task.DueDate.UTC(), "the intended 02:30 comes back")
			assert.Nil(t, task.IntendedDueTime, "NULL once the due date is on its intended time")
		})
	}

	t.Run("a stored value outside a day is ignored", func(t *testing.T) {
		s := setup(t)
		seedIntendedDueTime(t, s, intPtr(86400))
		patch(t, s)

		task := storedTask1(t, s)
		assert.Equal(t, shifted, task.DueDate.UTC())
		require.NotNil(t, task.IntendedDueTime)
		assert.Equal(t, 9000, *task.IntendedDueTime)
	})
}

// Only completion writes intended_due_time, and an explicit due-date edit clears it. The
// gap check cannot tell an edit from a shift within its bound: from the shifted state (Sun 03:30
// EDT, 9000), an edit to Sun 03:20 EDT is a 50-minute displacement, inside the 60-minute gap, so
// without the clear the next completion would land on Mon 02:30 EDT instead of 03:20.
func TestIntendedClear(t *testing.T) {
	shifted := time.Date(2027, 3, 14, 7, 30, 0, 0, time.UTC) // Sun 03:30 EDT
	edited := time.Date(2027, 3, 14, 7, 20, 0, 0, time.UTC)  // Sun 03:20 EDT
	jwtUser := &user.User{ID: 1, Username: "user1"}
	const rule = "FREQ=DAILY;INTERVAL=1"

	setup := func(t *testing.T) *xorm.Session {
		t.Helper()
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		t.Cleanup(func() { s.Close() })
		withFixedNow(t, time.Date(2027, 3, 14, 7, 0, 0, 0, time.UTC))
		setUserZone(t, s, 1, "America/New_York")
		seedRepeatingTask(t, s, rule, shifted)
		seedIntendedDueTime(t, s, intPtr(9000))
		return s
	}
	complete := func(t *testing.T, s *xorm.Session) {
		t.Helper()
		require.NoError(t, (&Task{ID: 1, Done: true}).updateSingleTask(s, jwtUser, []string{"done"}))
	}

	t.Run("a field-masked due-date edit clears it", func(t *testing.T) {
		s := setup(t)
		require.NoError(t, (&Task{ID: 1, DueDate: edited}).updateSingleTask(s, jwtUser, []string{"due_date"}))
		assert.Nil(t, storedTask1(t, s).IntendedDueTime)

		complete(t, s)
		assert.Equal(t, time.Date(2027, 3, 15, 7, 20, 0, 0, time.UTC), storedTask1(t, s).DueDate.UTC(), "Mon 03:20 EDT")
	})
	t.Run("a full update with a new due date clears it", func(t *testing.T) {
		s := setup(t)
		upd := readTaskForEcho(t, s, 1, jwtUser)
		upd.DueDate = edited
		require.NoError(t, upd.Update(s, jwtUser))
		assert.Nil(t, storedTask1(t, s).IntendedDueTime)
	})
	t.Run("clearing the due date clears it", func(t *testing.T) {
		s := setup(t)
		require.NoError(t, (&Task{ID: 1}).updateSingleTask(s, jwtUser, []string{"due_date"}))
		assert.Nil(t, storedTask1(t, s).IntendedDueTime)
	})

	keeps := func(t *testing.T, s *xorm.Session) {
		t.Helper()
		task := storedTask1(t, s)
		require.NotNil(t, task.IntendedDueTime)
		assert.Equal(t, 9000, *task.IntendedDueTime)
	}
	t.Run("a title-only update keeps it", func(t *testing.T) {
		s := setup(t)
		require.NoError(t, (&Task{ID: 1, Title: "renamed"}).updateSingleTask(s, jwtUser, []string{"title"}))
		keeps(t, s)
	})
	t.Run("a full update echoing every date, in another offset, keeps it", func(t *testing.T) {
		s := setup(t)
		upd := readTaskForEcho(t, s, 1, jwtUser)
		upd.DueDate = upd.DueDate.In(time.FixedZone("UTC+5", 5*60*60))
		upd.Title = "renamed"
		require.NoError(t, upd.Update(s, jwtUser))
		keeps(t, s)
	})
	t.Run("a start-date edit keeps it", func(t *testing.T) {
		s := setup(t)
		require.NoError(t, (&Task{ID: 1, StartDate: edited}).updateSingleTask(s, jwtUser, []string{"start_date"}))
		keeps(t, s)
	})
	t.Run("a rule change keeps it", func(t *testing.T) {
		s := setup(t)
		require.NoError(t, (&Task{ID: 1, Repeat: &TaskRepeat{Freq: "daily", Interval: 2}}).updateSingleTask(s, jwtUser, []string{"repeats"}))
		assert.Contains(t, storedTask1(t, s).Repeats, "INTERVAL=2")
		keeps(t, s)
	})
	t.Run("the incoming task never supplies it", func(t *testing.T) {
		s := setup(t)
		require.NoError(t, (&Task{ID: 1, Title: "renamed", IntendedDueTime: intPtr(1)}).updateSingleTask(s, jwtUser, []string{"title"}))
		keeps(t, s)
	})
}

// A refused completion (here: a relative reminder would land past 9999, so the dates are put
// back) leaves the task done with its dates and its intended_due_time unchanged; so does a rule
// whose UNTIL has passed.
func TestIntendedRefusedCompletion(t *testing.T) {
	jwtUser := &user.User{ID: 1, Username: "user1"}

	t.Run("outside the storable range", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		withFixedNow(t, time.Date(2026, 10, 23, 12, 0, 0, 0, time.UTC))
		due := time.Date(9999, 12, 29, 12, 0, 0, 0, time.UTC)
		seedRepeatingTask(t, s, "FREQ=DAILY;INTERVAL=1", due)
		seedIntendedDueTime(t, s, intPtr(9000))
		_, err := s.Insert(&TaskReminder{TaskID: 1, Reminder: due.Add(48 * time.Hour), RelativeTo: ReminderRelationDueDate, RelativePeriod: 172800})
		require.NoError(t, err)

		require.NoError(t, (&Task{ID: 1, Done: true}).updateSingleTask(s, jwtUser, []string{"done"}))

		task := storedTask1(t, s)
		assert.True(t, task.Done, "not advanced")
		assert.Equal(t, due, task.DueDate.UTC())
		require.NotNil(t, task.IntendedDueTime)
		assert.Equal(t, 9000, *task.IntendedDueTime)
	})
	t.Run("UNTIL passed", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		withFixedNow(t, time.Date(2027, 3, 20, 12, 0, 0, 0, time.UTC))
		due := time.Date(2027, 3, 14, 7, 30, 0, 0, time.UTC)
		seedRepeatingTask(t, s, "FREQ=DAILY;INTERVAL=1;UNTIL=20270316T000000Z", due)
		seedIntendedDueTime(t, s, intPtr(9000))

		require.NoError(t, (&TaskBucket{TaskID: 1, BucketID: 3, ProjectViewID: 4, ProjectID: 1}).Update(s, jwtUser))

		task := storedTask1(t, s)
		assert.True(t, task.Done)
		assert.Equal(t, due, task.DueDate.UTC())
		require.NotNil(t, task.IntendedDueTime)
		assert.Equal(t, 9000, *task.IntendedDueTime)
	})
}

// A saved-filter kanban view carries a negative synthetic project id. A Done move there must
// resolve the zone with the task's own project, so the owner fallback still applies. The
// completing user (user1) has no zone; the owner of project 1 is made user2, in New York, and
// user1 keeps write access through a share.
func TestCompletionZone_SavedFilterKanban(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()
	withFixedNow(t, time.Date(2026, 10, 27, 0, 5, 0, 0, time.UTC))
	u := &user.User{ID: 1, Username: "user1"}
	setUserZone(t, s, 1, "")
	setUserZone(t, s, 2, "America/New_York")
	setProjectOwner(t, s, 1, 2)
	_, err := s.Insert(&ProjectUser{UserID: 1, ProjectID: 1, Permission: PermissionWrite})
	require.NoError(t, err)
	seedRepeatingTask(t, s, "FREQ=WEEKLY;BYDAY=MO,WE,FR", time.Date(2026, 10, 27, 0, 0, 0, 0, time.UTC))

	sf := &SavedFilter{Title: "zone filter", Filters: &TaskCollection{Filter: "done = false"}}
	require.NoError(t, sf.Create(s, u))
	filterProjectID := getProjectIDFromSavedFilterID(sf.ID)
	require.Negative(t, filterProjectID)

	view := &ProjectView{}
	has, err := s.Where("project_id = ? AND view_kind = ?", filterProjectID, ProjectViewKindKanban).Get(view)
	require.NoError(t, err)
	require.True(t, has)
	done := &Bucket{Title: "done", ProjectViewID: view.ID, ProjectID: filterProjectID}
	require.NoError(t, done.Create(s, u))
	view.DoneBucketID = done.ID
	_, err = s.ID(view.ID).Cols("done_bucket_id").Update(view)
	require.NoError(t, err)

	tb := &TaskBucket{TaskID: 1, BucketID: done.ID, ProjectViewID: view.ID, ProjectID: filterProjectID}
	require.NoError(t, tb.Update(s, u))

	task := storedTask1(t, s)
	assert.False(t, task.Done, "reopened")
	assert.Equal(t, time.Date(2026, 10, 29, 0, 0, 0, 0, time.UTC), task.DueDate.UTC(),
		"the owner's New York calendar, not the server's (2026-10-28)")
}
