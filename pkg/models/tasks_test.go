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

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teambition/rrule-go"
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

		events.DispatchPending(context.Background(), s)
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
			updateDone(oldTask, newTask)

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
			updateDone(oldTask, newTask)

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
			t.Run("future due date searches from now (regression)", func(t *testing.T) {
				// Repeat-from-current-date with a due date in the FUTURE must advance
				// from now, not from the future due date. See plan §4.4.
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
				updateDone(oldTask, newTask)

				// Next occurrence is ~tomorrow (measured from now), well before the
				// original future due date — not futureDue+1day.
				assert.True(t, newTask.DueDate.After(time.Now()))
				assert.True(t, newTask.DueDate.Before(futureDue))
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
			t.Run("preserves start-to-due gap (regression)", func(t *testing.T) {
				// With both start and due dates, repeat-from-current-date must keep
				// the start->due offset; it previously collapsed start onto the due
				// date. See plan §4.6.
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
				updateDone(oldTask, newTask)

				assert.False(t, newTask.Done)
				// Due date moves to ~now+1day (measured from completion).
				assert.True(t, newTask.DueDate.After(time.Now()))
				// The 3-day start->due gap is preserved, not collapsed to zero.
				assert.Equal(t, oldDue.Sub(oldStart), newTask.DueDate.Sub(newTask.StartDate))
			})
		})
		t.Run("finite recurrence (COUNT) decrements and terminates", func(t *testing.T) {
			// COUNT=3 yields exactly 3 completable occurrences: the stored rule is
			// decremented on each completion and the task stays done once exhausted,
			// instead of repeating forever. RepeatsFromCurrentDate keeps the rule's
			// count window anchored at now so each completion reschedules. See §4.5.
			now := time.Now()

			// Completion 1: COUNT 3 -> 2, reschedules (task reopens).
			t1 := &Task{Repeats: "FREQ=DAILY;COUNT=3", RepeatsFromCurrentDate: true, DueDate: now}
			n1 := &Task{Done: true}
			updateDone(t1, n1)
			require.False(t, n1.Done, "should reopen after completing occurrence 1 of 3")
			o1, err := rrule.StrToROption(n1.Repeats)
			require.NoError(t, err)
			assert.Equal(t, 2, o1.Count)

			// Completion 2: COUNT 2 -> 1, reschedules.
			t2 := &Task{Repeats: n1.Repeats, RepeatsFromCurrentDate: true, DueDate: n1.DueDate}
			n2 := &Task{Done: true}
			updateDone(t2, n2)
			require.False(t, n2.Done, "should reopen after completing occurrence 2 of 3")
			o2, err := rrule.StrToROption(n2.Repeats)
			require.NoError(t, err)
			assert.Equal(t, 1, o2.Count)

			// Completion 3: COUNT == 1, exhausted -> stays done, no reschedule.
			t3 := &Task{Repeats: n2.Repeats, RepeatsFromCurrentDate: true, DueDate: n2.DueDate}
			n3 := &Task{Done: true}
			updateDone(t3, n3)
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

			updateDone(oldTask, newTask)

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

			updateDone(oldTask, newTask)

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
		// Frequencies with a fixed second-count map to a Duration.
		{"hourly interval 1", rrule.HOURLY, 1, time.Hour, true},
		{"daily interval 1", rrule.DAILY, 1, 24 * time.Hour, true},
		{"weekly interval 2", rrule.WEEKLY, 2, 14 * 24 * time.Hour, true},
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
	updateDone(oldTask, newTask)
	elapsed := time.Since(start)

	require.Less(t, elapsed, time.Second, "updateDone must not take seconds for ancient due dates")
	assert.True(t, newTask.DueDate.After(start), "new due date must be strictly after now")
	assert.True(t, newTask.StartDate.After(start), "new start date must be strictly after now")
	assert.True(t, newTask.EndDate.After(start), "new end date must be strictly after now")
	assert.False(t, newTask.Done, "repeating task should be unmarked as done")
}

func TestUpdateDone_RRuleAncientDueDate_SlowPath(t *testing.T) {
	// A daily rule with a By-part takes the rrule.After slow path (no fixed
	// duration). Anchored in year 1000, rrule.After hits its internal iteration
	// cap and returns the zero time, so without clampRRuleAnchor advancing the
	// DTSTART the task would never reschedule.
	oldTask := &Task{
		Done:    false,
		Repeats: "FREQ=DAILY;BYHOUR=9",
		DueDate: time.Date(1000, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	newTask := &Task{Done: true}

	start := time.Now()
	updateDone(oldTask, newTask)
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
	updateDone(oldTask, newTask)

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
