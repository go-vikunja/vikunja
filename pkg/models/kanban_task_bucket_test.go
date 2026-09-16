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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskBucket_Update(t *testing.T) {
	u := &user.User{ID: 1}

	t.Run("full bucket", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		tb := &TaskBucket{
			TaskID:        1,
			BucketID:      2, // Bucket 2 already has 3 tasks and a limit of 3
			ProjectViewID: 4,
			ProjectID:     1, // In actual web requests set via the url
		}

		err := tb.Update(s, u)
		require.Error(t, err)
		assert.True(t, IsErrBucketLimitExceeded(err))
	})
	t.Run("full bucket but not changing the bucket", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		tb := &TaskBucket{
			TaskID:        4,
			BucketID:      2, // Bucket 2 already has 3 tasks and a limit of 3
			ProjectViewID: 4,
			ProjectID:     1, // In actual web requests set via the url
		}
		err := tb.Update(s, u)
		require.NoError(t, err)
	})
	t.Run("bucket on other project view", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		tb := &TaskBucket{
			TaskID:        1,
			BucketID:      4, // Bucket 4 belongs to project 2
			ProjectViewID: 4,
			ProjectID:     1, // In actual web requests set via the url
		}
		err := tb.Update(s, u)
		require.Error(t, err)
		assert.True(t, IsErrBucketDoesNotBelongToProject(err))
	})
	t.Run("moving a task to the done bucket", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		tb := &TaskBucket{
			TaskID:        1,
			BucketID:      3, // Bucket 3 is the done bucket
			ProjectViewID: 4,
			ProjectID:     1, // In actual web requests set via the url
		}
		err := tb.Update(s, u)
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)
		assert.True(t, tb.Task.Done)

		db.AssertExists(t, "tasks", map[string]interface{}{
			"id":   1,
			"done": true,
		}, false)
		db.AssertExists(t, "task_buckets", map[string]interface{}{
			"task_id":   1,
			"bucket_id": 3,
		}, false)
	})
	t.Run("move done task out of done bucket", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		tb := &TaskBucket{
			TaskID:        2,
			BucketID:      1, // Bucket 1 is the default bucket
			ProjectViewID: 4,
			ProjectID:     1, // In actual web requests set via the url
		}
		err := tb.Update(s, u)
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)
		assert.False(t, tb.Task.Done)

		db.AssertExists(t, "tasks", map[string]interface{}{
			"id":   tb.TaskID,
			"done": false,
		}, false)
		db.AssertExists(t, "task_buckets", map[string]interface{}{
			"task_id":   tb.TaskID,
			"bucket_id": 1,
		}, false)
		db.AssertMissing(t, "task_buckets", map[string]interface{}{
			"task_id":   tb.TaskID,
			"bucket_id": 3,
		})
	})
	t.Run("moving a repeating task to the done bucket", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		tb := &TaskBucket{
			TaskID:        28,
			BucketID:      3, // Bucket 3 is the done bucket
			ProjectViewID: 4,
			ProjectID:     1, // In actual web requests set via the url
		}

		// Before running the TaskBucket Update we retrieve the task and execute
		// an updateDone to obtain the task with updated start/end/due dates
		// This way we can later match them with what happens after running TaskBucket Update
		u := &user.User{ID: 1}
		oldTask := &Task{ID: tb.TaskID}
		err := oldTask.ReadOne(s, u)
		require.NoError(t, err)
		updatedTask := &Task{ID: tb.TaskID}
		err = updatedTask.ReadOne(s, u)
		require.NoError(t, err)
		updatedTask.Done = true
		updateDone(oldTask, updatedTask) // updatedTask now contains the updated dates

		err = tb.Update(s, u)
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)
		assert.False(t, tb.Task.Done)
		assert.Equal(t, int64(1), tb.BucketID) // This should be the actual bucket

		db.AssertExists(t, "tasks", map[string]interface{}{
			"id":   1,
			"done": false,
		}, false)
		db.AssertExists(t, "task_buckets", map[string]interface{}{
			"task_id":   1,
			"bucket_id": 1,
		}, false)

		assert.Equal(t, updatedTask.DueDate.Unix(), tb.Task.DueDate.Unix())
		assert.Equal(t, updatedTask.StartDate.Unix(), tb.Task.StartDate.Unix())
		assert.Equal(t, updatedTask.EndDate.Unix(), tb.Task.EndDate.Unix())
	})
	t.Run("moving a repeating task from a non-default bucket to the done bucket moves it to the default bucket", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Task 28 is a repeating task. Fixtures place it in bucket 1 (the
		// default bucket on view 4). Pre-position it in bucket 2 ("Doing")
		// using a raw update so we can bypass the bucket-2 limit check —
		// the limit check would otherwise block this setup step since
		// bucket 2 is already at its limit of 3 tasks.
		_, err := s.Where("task_id = ? AND project_view_id = ?", 28, 4).
			Cols("bucket_id").
			Update(&TaskBucket{BucketID: 2})
		require.NoError(t, err)

		tb := &TaskBucket{
			TaskID:        28,
			BucketID:      3, // Bucket 3 is the done bucket on view 4
			ProjectViewID: 4,
			ProjectID:     1,
		}
		err = tb.Update(s, u)
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)

		// Repeating task should have been re-opened by updateDone...
		assert.False(t, tb.Task.Done)

		// ...and routed to the DEFAULT bucket (1), not left in the source
		// bucket (2) and not placed in the done bucket (3).
		assert.Equal(t, int64(1), tb.BucketID)

		db.AssertExists(t, "tasks", map[string]interface{}{
			"id":   28,
			"done": false,
		}, false)
		db.AssertExists(t, "task_buckets", map[string]interface{}{
			"task_id":   28,
			"bucket_id": 1,
		}, false)
		db.AssertMissing(t, "task_buckets", map[string]interface{}{
			"task_id":   28,
			"bucket_id": 2,
		})
		db.AssertMissing(t, "task_buckets", map[string]interface{}{
			"task_id":   28,
			"bucket_id": 3,
		})
	})

	t.Run("moving a repeating task from a non-default bucket when no default bucket is configured preserves the original bucket", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 1}

		// View 4 has default_bucket_id: 1 and done_bucket_id: 3.
		// Remove the default bucket to test fallback behavior.
		_, err := s.ID(4).Cols("default_bucket_id").Update(&ProjectView{DefaultBucketID: 0})
		require.NoError(t, err)

		// Task 28 is a repeating task. Pre-position it in bucket 2
		// using a raw update so we can bypass the bucket-2 limit check.
		_, err = s.Where("task_id = ? AND project_view_id = ?", 28, 4).
			Cols("bucket_id").
			Update(&TaskBucket{BucketID: 2})
		require.NoError(t, err)

		tb := &TaskBucket{
			TaskID:        28,
			BucketID:      3, // Bucket 3 is the done bucket on view 4
			ProjectViewID: 4,
			ProjectID:     1,
		}
		err = tb.Update(s, u)
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)

		// Repeating task should have been re-opened by updateDone...
		assert.False(t, tb.Task.Done)

		// ...and routed back to the ORIGINAL bucket (2), not the default (1)
		// and not left in the done bucket (3).
		assert.Equal(t, int64(2), tb.BucketID)

		db.AssertExists(t, "tasks", map[string]interface{}{
			"id":   28,
			"done": false,
		}, false)
		db.AssertExists(t, "task_buckets", map[string]interface{}{
			"task_id":   28,
			"bucket_id": 2,
		}, false)
		db.AssertMissing(t, "task_buckets", map[string]interface{}{
			"task_id":   28,
			"bucket_id": 1,
		})
		db.AssertMissing(t, "task_buckets", map[string]interface{}{
			"task_id":   28,
			"bucket_id": 3,
		})
	})

	t.Run("done task already in another view's done bucket", func(t *testing.T) {
		// Regression test: marking a task done syncs it into the done bucket
		// of every kanban view in the project. When the task already sits in
		// such a view's done bucket the sync is a no-op update, but on
		// MySQL/MariaDB an UPDATE that doesn't change the value reports 0
		// affected rows. The upsert then mistook that for "row missing" and
		// inserted, hitting the unique index with ErrTaskAlreadyExistsInBucket.
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// A second manual kanban view on project 1. Creating it auto-generates
		// the To-Do/Doing/Done buckets and sets its done bucket.
		secondView := &ProjectView{
			Title:                   "Second Kanban",
			ProjectID:               1,
			ViewKind:                ProjectViewKindKanban,
			BucketConfigurationMode: BucketConfigurationModeManual,
		}
		err := secondView.Create(s, u)
		require.NoError(t, err)
		require.NotZero(t, secondView.DoneBucketID)

		// Pre-place task 1 in the second view's done bucket without going
		// through the done-sync, so the task itself is still open and view 4
		// still has it in its default bucket.
		_, err = s.Where("task_id = ? AND project_view_id = ?", 1, secondView.ID).
			Cols("bucket_id").
			Update(&TaskBucket{BucketID: secondView.DoneBucketID})
		require.NoError(t, err)

		// Moving task 1 into view 4's done bucket marks it done and triggers
		// the cross-view sync into the second view's done bucket, where it
		// already lives. This must succeed rather than error.
		tb := &TaskBucket{
			TaskID:        1,
			BucketID:      3, // done bucket on view 4
			ProjectViewID: 4,
			ProjectID:     1,
		}
		err = tb.Update(s, u)
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)
		assert.True(t, tb.Task.Done)

		db.AssertExists(t, "task_buckets", map[string]interface{}{
			"task_id":   1,
			"bucket_id": 3,
		}, false)
		db.AssertExists(t, "task_buckets", map[string]interface{}{
			"task_id":         1,
			"project_view_id": secondView.ID,
			"bucket_id":       secondView.DoneBucketID,
		}, false)
	})

	t.Run("saved filter: first task into empty limited bucket is allowed", func(t *testing.T) {
		// Regression test for #2672: on a saved-filter kanban view the bucket
		// limit was checked against the total number of tasks matching the
		// filter instead of the number of tasks actually in the target bucket,
		// so adding the first task to an empty limited bucket was wrongly
		// rejected with ErrBucketLimitExceeded.
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// A saved filter matching many tasks; the filter total is well above
		// the bucket limit we set below.
		sf := &SavedFilter{
			Title:   "limit-filter",
			Filters: &TaskCollection{Filter: "done = false"},
		}
		err := sf.Create(s, u)
		require.NoError(t, err)

		filterProjectID := getProjectIDFromSavedFilterID(sf.ID)

		view := &ProjectView{}
		exists, err := s.Where("project_id = ? AND view_kind = ?", filterProjectID, ProjectViewKindKanban).Get(view)
		require.NoError(t, err)
		require.True(t, exists)

		// All matching tasks are placed in the default bucket on creation;
		// pick three of them to move into a fresh, empty bucket.
		var defaultTasks []*TaskBucket
		err = s.Where("project_view_id = ?", view.ID).Find(&defaultTasks)
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(defaultTasks), 3, "filter must match enough tasks to exceed the bucket limit")

		limitedBucket := &Bucket{
			Title:         "limited",
			ProjectViewID: view.ID,
			ProjectID:     filterProjectID,
			Limit:         2,
		}
		err = limitedBucket.Create(s, u)
		require.NoError(t, err)

		moveTaskToBucket := func(taskID int64) error {
			tb := &TaskBucket{
				TaskID:        taskID,
				BucketID:      limitedBucket.ID,
				ProjectViewID: view.ID,
				ProjectID:     filterProjectID,
			}
			return tb.Update(s, u)
		}

		// Moving the FIRST task into the empty bucket must succeed (0/2 -> 1/2).
		require.NoError(t, moveTaskToBucket(defaultTasks[0].TaskID))
		// The second one fills the bucket up to the limit (1/2 -> 2/2).
		require.NoError(t, moveTaskToBucket(defaultTasks[1].TaskID))
		// The third one would exceed the limit and must be rejected.
		err = moveTaskToBucket(defaultTasks[2].TaskID)
		require.Error(t, err)
		assert.True(t, IsErrBucketLimitExceeded(err))
	})

	t.Run("keep done timestamp when moving task between projects", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		u := &user.User{ID: 1}

		doneAt := time.Now().Round(time.Second)

		// Set a done timestamp on the task
		func() {
			s := db.NewSession()
			defer s.Close()
			_, err := s.ID(2).Cols("done_at").Update(&Task{DoneAt: doneAt})
			require.NoError(t, err)
			err = s.Commit()
			require.NoError(t, err)
		}()

		// Move the task to another project without a done bucket using Task.Update
		func() {
			s := db.NewSession()
			defer s.Close()
			task := &Task{ID: 2, Done: true, ProjectID: 9}
			err := task.Update(s, u)
			require.NoError(t, err)
			err = s.Commit()
			require.NoError(t, err)
		}()

		// Verify the task still has the same done timestamp
		func() {
			s := db.NewSession()
			defer s.Close()
			var task Task
			_, err := s.ID(2).Get(&task)
			require.NoError(t, err)
			assert.True(t, task.Done)
			assert.WithinDuration(t, doneAt, task.DoneAt, time.Second)
		}()

		// Move the task back to the original project with a done bucket using Task.Update
		func() {
			s := db.NewSession()
			defer s.Close()
			task := &Task{ID: 2, Done: true, ProjectID: 1}
			err := task.Update(s, u)
			require.NoError(t, err)
			err = s.Commit()
			require.NoError(t, err)
		}()

		// Verify the done timestamp is still preserved
		func() {
			s := db.NewSession()
			defer s.Close()
			var task Task
			_, err := s.ID(2).Get(&task)
			require.NoError(t, err)
			assert.True(t, task.Done)
			assert.WithinDuration(t, doneAt, task.DoneAt, time.Second)
		}()
	})
}
