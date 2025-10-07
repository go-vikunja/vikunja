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

package services

import (
	"testing"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/builder"
)

// ============================================================================
// SERVICE LAYER BUSINESS LOGIC TESTS (Migrated from pkg/models/tasks_test.go)
// These tests validate the business logic that was moved from the model layer
// to the service layer as part of T015A and T015B.
// ============================================================================

func TestTaskService_Create_WithBusinessLogic(t *testing.T) {
	u := &user.User{
		ID:       1,
		Username: "user1",
		Email:    "user1@example.com",
	}

	t.Run("create task with reminders", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		ts := NewTaskService(testEngine)
		task := &models.Task{
			Title:       "Task with Reminders",
			Description: "Testing relative reminders",
			ProjectID:   1,
			DueDate:     time.Date(2023, time.March, 7, 22, 5, 0, 0, time.UTC),
			StartDate:   time.Date(2023, time.March, 7, 22, 5, 10, 0, time.UTC),
			EndDate:     time.Date(2023, time.March, 7, 22, 5, 20, 0, time.UTC),
			Reminders: []*models.TaskReminder{
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
			},
		}

		createdTask, err := ts.CreateWithOptions(s, task, u, true, true, false)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		// Verify relative reminders were calculated correctly
		assert.Equal(t, time.Date(2023, time.March, 7, 22, 5, 1, 0, time.UTC), createdTask.Reminders[0].Reminder)
		assert.Equal(t, int64(1), createdTask.Reminders[0].RelativePeriod)
		assert.Equal(t, models.ReminderRelationDueDate, createdTask.Reminders[0].RelativeTo)
		assert.Equal(t, time.Date(2023, time.March, 7, 22, 5, 8, 0, time.UTC), createdTask.Reminders[1].Reminder)
		assert.Equal(t, models.ReminderRelationStartDate, createdTask.Reminders[1].RelativeTo)
		assert.Equal(t, time.Date(2023, time.March, 7, 22, 5, 19, 0, time.UTC), createdTask.Reminders[2].Reminder)
		assert.Equal(t, models.ReminderRelationEndDate, createdTask.Reminders[2].RelativeTo)
		assert.Equal(t, time.Date(2023, time.March, 7, 23, 0, 0, 0, time.UTC), createdTask.Reminders[3].Reminder)
	})

	t.Run("empty title should fail", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		ts := NewTaskService(testEngine)
		task := &models.Task{
			Title:       "",
			Description: "Lorem Ipsum Dolor",
			ProjectID:   1,
		}

		_, err := ts.CreateWithOptions(s, task, u, true, true, false)
		require.Error(t, err)
		assert.True(t, models.IsErrTaskCannotBeEmpty(err))
	})

	t.Run("nonexistant project should fail", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		ts := NewTaskService(testEngine)
		task := &models.Task{
			Title:       "Test",
			Description: "Lorem Ipsum Dolor",
			ProjectID:   9999999,
		}

		_, err := ts.CreateWithOptions(s, task, u, true, true, false)
		require.Error(t, err)
		assert.True(t, models.IsErrProjectDoesNotExist(err))
	})
}

func TestTaskService_Update_WithBusinessLogic(t *testing.T) {
	u := &user.User{ID: 1}

	t.Run("update basic task fields", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		ts := NewTaskService(testEngine)
		task := &models.Task{
			ID:          1,
			Title:       "Updated Title",
			Description: "Updated Description",
			ProjectID:   1,
		}

		updatedTask, err := ts.Update(s, task, u)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		db.AssertExists(t, "tasks", map[string]interface{}{
			"id":          1,
			"title":       "Updated Title",
			"description": "Updated Description",
			"project_id":  1,
		}, false)

		assert.Equal(t, "Updated Title", updatedTask.Title)
		assert.Equal(t, "Updated Description", updatedTask.Description)
	})

	t.Run("move task to different project should reassign bucket", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		ts := NewTaskService(testEngine)
		task := &models.Task{
			ID:        1,
			ProjectID: 2,
		}

		updatedTask, err := ts.Update(s, task, u)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		// Verify task moved to project 2
		db.AssertExists(t, "tasks", map[string]interface{}{
			"id":         1,
			"project_id": 2,
		}, false)

		// Verify task was placed in default bucket of new project (bucket 40 for project 2)
		db.AssertExists(t, "task_buckets", map[string]interface{}{
			"task_id":   1,
			"bucket_id": 40,
		}, false)

		assert.Equal(t, int64(2), updatedTask.ProjectID)
	})

	t.Run("marking task as done should move to done bucket", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		ts := NewTaskService(testEngine)
		task := &models.Task{
			ID:   1,
			Done: true,
		}

		updatedTask, err := ts.Update(s, task, u)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		// Verify task marked as done
		db.AssertExists(t, "tasks", map[string]interface{}{
			"id":   1,
			"done": true,
		}, false)

		// Verify task moved to done bucket (bucket 3 for project 1)
		db.AssertExists(t, "task_buckets", map[string]interface{}{
			"task_id":   1,
			"bucket_id": 3,
		}, false)

		assert.True(t, updatedTask.Done)
	})

	t.Run("move done task to different project with done bucket", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		ts := NewTaskService(testEngine)
		task := &models.Task{
			ID:        2,
			Done:      true,
			ProjectID: 2,
		}

		updatedTask, err := ts.Update(s, task, u)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		// Verify task moved and still done
		db.AssertExists(t, "tasks", map[string]interface{}{
			"id":         2,
			"project_id": 2,
			"done":       true,
		}, false)

		// Verify task moved to done bucket of new project (bucket 4 for project 2)
		db.AssertExists(t, "task_buckets", map[string]interface{}{
			"task_id":   2,
			"bucket_id": 4,
		}, false)

		assert.True(t, updatedTask.Done)
		assert.Equal(t, int64(2), updatedTask.ProjectID)
	})

	t.Run("repeating tasks should not move to done bucket", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		ts := NewTaskService(testEngine)
		task := &models.Task{
			ID:          28,
			Done:        true,
			RepeatAfter: 3600,
		}

		updatedTask, err := ts.Update(s, task, u)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		// Repeating task should NOT be done after update
		db.AssertExists(t, "tasks", map[string]interface{}{
			"id":   28,
			"done": false,
		}, false)

		// Should stay in original bucket
		db.AssertExists(t, "task_buckets", map[string]interface{}{
			"task_id":   28,
			"bucket_id": 1,
		}, false)

		assert.False(t, updatedTask.Done)
	})

	t.Run("moving task between projects should recalculate index", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		ts := NewTaskService(testEngine)
		task := &models.Task{
			ID:        12,
			ProjectID: 2, // From project 1
		}

		updatedTask, err := ts.Update(s, task, u)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		// Task should get correct index in new project
		assert.Equal(t, int64(3), updatedTask.Index)
	})

	t.Run("update task reminders", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		ts := NewTaskService(testEngine)
		task := &models.Task{
			ID:        1,
			ProjectID: 1,
			Title:     "test",
			DueDate:   time.Date(2023, time.March, 7, 22, 5, 0, 0, time.UTC),
			StartDate: time.Date(2023, time.March, 7, 22, 5, 10, 0, time.UTC),
			EndDate:   time.Date(2023, time.March, 7, 22, 5, 20, 0, time.UTC),
			Reminders: []*models.TaskReminder{
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
			},
		}

		updatedTask, err := ts.Update(s, task, u)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		// Verify reminders were calculated correctly
		assert.Equal(t, time.Date(2023, time.March, 7, 22, 5, 1, 0, time.UTC), updatedTask.Reminders[0].Reminder)
		assert.Equal(t, int64(1), updatedTask.Reminders[0].RelativePeriod)
		assert.Equal(t, models.ReminderRelationDueDate, updatedTask.Reminders[0].RelativeTo)
		assert.Equal(t, time.Date(2023, time.March, 7, 22, 5, 8, 0, time.UTC), updatedTask.Reminders[1].Reminder)
		assert.Equal(t, models.ReminderRelationStartDate, updatedTask.Reminders[1].RelativeTo)
		assert.Equal(t, time.Date(2023, time.March, 7, 22, 5, 19, 0, time.UTC), updatedTask.Reminders[2].Reminder)
		assert.Equal(t, models.ReminderRelationEndDate, updatedTask.Reminders[2].RelativeTo)
		assert.Equal(t, time.Date(2023, time.March, 7, 23, 0, 0, 0, time.UTC), updatedTask.Reminders[3].Reminder)
	})

	t.Run("duplicate reminders should be saved once", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		ts := NewTaskService(testEngine)
		task := &models.Task{
			ID:    1,
			Title: "test",
			Reminders: []*models.TaskReminder{
				{
					Reminder: time.Unix(1674745156, 0),
				},
				{
					Reminder: time.Unix(1674745156, 223),
				},
			},
			ProjectID: 1,
		}

		_, err := ts.Update(s, task, u)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		db.AssertCount(t, "task_reminders", builder.Eq{"task_id": 1}, 1)
	})

	t.Run("update relative reminder when start date changes", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		ts := NewTaskService(testEngine)

		// given task with start_date and relative reminder for start_date
		taskBefore := &models.Task{
			Title:     "test",
			ProjectID: 1,
			StartDate: time.Date(2022, time.March, 8, 8, 5, 20, 0, time.UTC),
			Reminders: []*models.TaskReminder{
				{
					RelativeTo:     "start_date",
					RelativePeriod: -60,
				},
			},
		}
		createdTask, err := ts.CreateWithOptions(s, taskBefore, u, true, true, false)
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		assert.Equal(t, time.Date(2022, time.March, 8, 8, 4, 20, 0, time.UTC), createdTask.Reminders[0].Reminder)

		// when start_date is modified
		task := createdTask
		task.StartDate = time.Date(2023, time.March, 8, 8, 5, 0, 0, time.UTC)
		updatedTask, err := ts.Update(s, task, u)
		require.NoError(t, err)

		// then reminder time is updated
		assert.Equal(t, time.Date(2023, time.March, 8, 8, 4, 0, 0, time.UTC), updatedTask.Reminders[0].Reminder)
		require.NoError(t, s.Commit())
	})

	t.Run("nonexistent task should fail", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		ts := NewTaskService(testEngine)
		task := &models.Task{
			ID:          9999999,
			Title:       "test10000",
			Description: "Lorem Ipsum Dolor",
			ProjectID:   1,
		}

		_, err := ts.Update(s, task, u)
		require.Error(t, err)
		assert.True(t, models.IsErrTaskDoesNotExist(err))
	})
}

func TestTaskService_Delete_WithBusinessLogic(t *testing.T) {
	u := &user.User{ID: 1}

	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		ts := NewTaskService(testEngine)
		task := &models.Task{
			ID: 1,
		}
		err := ts.Delete(s, task, u)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		db.AssertMissing(t, "tasks", map[string]interface{}{
			"id": 1,
		})
	})
}
