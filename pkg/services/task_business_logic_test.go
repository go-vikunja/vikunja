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

	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		ts := NewTaskService(testEngine)
		task := &models.Task{
			Title:       "Lorem",
			Description: "Lorem Ipsum Dolor",
			ProjectID:   1,
		}

		createdTask, err := ts.CreateWithOptions(s, task, u, true, true, false)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		// Verify UID was generated
		assert.NotEmpty(t, createdTask.UID)
		// Verify index was assigned
		assert.NotEmpty(t, createdTask.Index)
		assert.Equal(t, int64(18), createdTask.Index)

		// Verify task was created in database
		db.AssertExists(t, "tasks", map[string]interface{}{
			"id":            createdTask.ID,
			"title":         "Lorem",
			"description":   "Lorem Ipsum Dolor",
			"project_id":    1,
			"created_by_id": 1,
		}, false)

		// Verify task was placed in default bucket
		db.AssertExists(t, "task_buckets", map[string]interface{}{
			"task_id":   createdTask.ID,
			"bucket_id": 1,
		}, false)
	})

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

	t.Run("nonexistant user should fail", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		ts := NewTaskService(testEngine)
		nUser := &user.User{ID: 99999999}
		task := &models.Task{
			Title:       "Test",
			Description: "Lorem Ipsum Dolor",
			ProjectID:   1,
		}

		_, err := ts.CreateWithOptions(s, task, nUser, true, true, false)
		require.Error(t, err)
		// Service layer performs permission check first, which returns ErrAccessDenied
		// This is more secure than the original model layer which checked user existence first
		assert.True(t, models.IsErrGenericForbidden(err), "Expected ErrAccessDenied (better security), got: %v", err)
	})

	t.Run("default bucket different", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		ts := NewTaskService(testEngine)
		// Project 6 is owned by user 6, so we need to use that user
		u6 := &user.User{ID: 6}
		task := &models.Task{
			Title:       "Lorem",
			Description: "Lorem Ipsum Dolor",
			ProjectID:   6, // Project 6 has bucket 22 as default with position 2
		}

		createdTask, err := ts.CreateWithOptions(s, task, u6, true, true, false)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		// Verify task was placed in project 6's default bucket (bucket 22)
		db.AssertExists(t, "task_buckets", map[string]interface{}{
			"task_id":   createdTask.ID,
			"bucket_id": 22, // default bucket of project 6 but with a position of 2
		}, false)
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

	t.Run("nonexistent task should fail", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		ts := NewTaskService(testEngine)
		task := &models.Task{
			ID: 9999999,
		}
		err := ts.Delete(s, task, u)
		require.Error(t, err)
		// Service layer performs permission check first, which fails for nonexistent tasks
		// This is more secure than revealing whether a task exists
		assert.True(t, models.IsErrGenericForbidden(err), "Expected ErrAccessDenied (better security), got: %v", err)
	})
}

// ============================================================================
// ENHANCEMENT TESTS (Beyond Original Model Tests)
// These tests provide additional coverage for features not explicitly tested
// in the original model test suite, improving comprehensive test coverage.
// ============================================================================

func TestTaskService_Create_WithAssignees(t *testing.T) {
	u := &user.User{ID: 1}

	t.Run("create task with assignees", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		ts := NewTaskService(testEngine)
		task := &models.Task{
			Title:     "Task with Assignees",
			ProjectID: 1,
			Assignees: []*user.User{
				{ID: 1}, // User 1 owns the project, so has access
			},
		}

		createdTask, err := ts.CreateWithOptions(s, task, u, true, true, false)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		// Verify assignees were created
		db.AssertExists(t, "task_assignees", map[string]interface{}{
			"task_id": createdTask.ID,
			"user_id": 1,
		}, false)

		// Verify the returned task includes assignees
		assert.Len(t, createdTask.Assignees, 1)
	})
}

func TestTaskService_Create_WithLabels(t *testing.T) {
	u := &user.User{ID: 1}

	t.Run("create task with labels", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		ts := NewTaskService(testEngine)
		task := &models.Task{
			Title:     "Task with Labels",
			ProjectID: 1,
			Labels: []*models.Label{
				{ID: 1},
				{ID: 4},
			},
		}

		createdTask, err := ts.CreateWithOptions(s, task, u, true, true, false)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		// Verify labels were associated (table is label_tasks not label_task)
		db.AssertExists(t, "label_tasks", map[string]interface{}{
			"task_id":  createdTask.ID,
			"label_id": 1,
		}, false)
		db.AssertExists(t, "label_tasks", map[string]interface{}{
			"task_id":  createdTask.ID,
			"label_id": 4,
		}, false)

		// Verify the returned task includes labels
		assert.Len(t, createdTask.Labels, 2)
	})
}

func TestTaskService_Update_Assignees(t *testing.T) {
	u := &user.User{ID: 1}

	t.Run("update task assignees", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		ts := NewTaskService(testEngine)
		task := &models.Task{
			ID:        1,
			ProjectID: 1,
			Assignees: []*user.User{
				{ID: 1}, // User 1 has access to project 1
			},
		}

		updatedTask, err := ts.Update(s, task, u)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		// Verify assignees were updated
		db.AssertExists(t, "task_assignees", map[string]interface{}{
			"task_id": 1,
			"user_id": 1,
		}, false)

		// Verify the returned task includes updated assignees
		assert.Len(t, updatedTask.Assignees, 1)
	})
}

func TestTaskService_Update_Labels(t *testing.T) {
	u := &user.User{ID: 1}

	t.Run("update task labels", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		ts := NewTaskService(testEngine)
		task := &models.Task{
			ID:        1,
			ProjectID: 1,
			Labels: []*models.Label{
				{ID: 1}, // User 1 has access to label 1
				{ID: 4}, // User 1 has access to label 4
			},
		}

		updatedTask, err := ts.Update(s, task, u)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		// Verify labels were updated
		db.AssertExists(t, "label_tasks", map[string]interface{}{
			"task_id":  1,
			"label_id": 1,
		}, false)
		db.AssertExists(t, "label_tasks", map[string]interface{}{
			"task_id":  1,
			"label_id": 4,
		}, false)

		// Verify the returned task includes updated labels
		assert.Len(t, updatedTask.Labels, 2)
	})

	t.Run("remove labels from task", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// First add labels to task
		ts := NewTaskService(testEngine)
		task := &models.Task{
			ID:        1,
			ProjectID: 1,
			Labels: []*models.Label{
				{ID: 1},
				{ID: 4},
			},
		}

		_, err := ts.Update(s, task, u)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		// Now remove all labels
		s2 := db.NewSession()
		defer s2.Close()
		task.Labels = []*models.Label{}
		updatedTask, err := ts.Update(s2, task, u)
		require.NoError(t, err)
		require.NoError(t, s2.Commit())

		// Verify labels were removed
		db.AssertMissing(t, "label_tasks", map[string]interface{}{
			"task_id": 1,
		})

		// Verify the returned task has no labels
		assert.Len(t, updatedTask.Labels, 0)
	})
}

func TestTaskService_Delete_WithCascade(t *testing.T) {
	u := &user.User{ID: 1}

	t.Run("delete task with cascade", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Create a task with assignees, labels, reminders, and bucket assignment
		ts := NewTaskService(testEngine)
		task := &models.Task{
			Title:     "Task to Delete",
			ProjectID: 1,
			Assignees: []*user.User{{ID: 1}},
			Labels:    []*models.Label{{ID: 1}},
			Reminders: []*models.TaskReminder{
				{Reminder: time.Date(2023, time.March, 7, 23, 0, 0, 0, time.UTC)},
			},
		}

		createdTask, err := ts.CreateWithOptions(s, task, u, true, true, false)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		// Verify all related records exist
		db.AssertExists(t, "tasks", map[string]interface{}{"id": createdTask.ID}, false)
		db.AssertExists(t, "task_assignees", map[string]interface{}{"task_id": createdTask.ID}, false)
		db.AssertExists(t, "label_tasks", map[string]interface{}{"task_id": createdTask.ID}, false)
		db.AssertExists(t, "task_reminders", map[string]interface{}{"task_id": createdTask.ID}, false)
		db.AssertExists(t, "task_buckets", map[string]interface{}{"task_id": createdTask.ID}, false)

		// Delete the task
		s2 := db.NewSession()
		defer s2.Close()
		err = ts.Delete(s2, createdTask, u)
		require.NoError(t, err)
		require.NoError(t, s2.Commit())

		// Verify task and all related records were deleted (cascade)
		db.AssertMissing(t, "tasks", map[string]interface{}{"id": createdTask.ID})
		db.AssertMissing(t, "task_assignees", map[string]interface{}{"task_id": createdTask.ID})
		db.AssertMissing(t, "label_tasks", map[string]interface{}{"task_id": createdTask.ID})
		db.AssertMissing(t, "task_reminders", map[string]interface{}{"task_id": createdTask.ID})
		db.AssertMissing(t, "task_buckets", map[string]interface{}{"task_id": createdTask.ID})
	})
}

// FI-004: Comprehensive Assignee Validation Tests
func TestTaskService_Assignee_WithoutProjectAccess(t *testing.T) {
	t.Run("assigning user without project access should fail gracefully", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 1}
		ts := NewTaskService(testEngine)

		// User 5 does not have access to project 1 (owned by user 1)
		// See fixtures: user5 only has access to project 5
		task := &models.Task{
			ID:        1,
			ProjectID: 1,
			Assignees: []*user.User{
				{ID: 5}, // User 5 does NOT have access to project 1
			},
		}

		updatedTask, err := ts.Update(s, task, u)

		// Should return an error indicating user doesn't have access
		assert.Error(t, err)
		assert.Nil(t, updatedTask)
		assert.Contains(t, err.Error(), "does not have access")
	})

	t.Run("creating task with assignee without project access should fail", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 1}
		ts := NewTaskService(testEngine)

		// Try to create task with user 5 as assignee (who doesn't have project 1 access)
		task := &models.Task{
			Title:     "Task with Invalid Assignee",
			ProjectID: 1,
			Assignees: []*user.User{
				{ID: 5}, // User 5 does NOT have access to project 1
			},
		}

		createdTask, err := ts.CreateWithOptions(s, task, u, true, true, false)

		// Should return an error
		assert.Error(t, err)
		assert.Nil(t, createdTask)
		assert.Contains(t, err.Error(), "does not have access")
	})
}

func TestTaskService_Assignee_BulkOperations(t *testing.T) {
	t.Run("add multiple assignees at once", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 3} // User 3 owns project 3
		ts := NewTaskService(testEngine)

		// User 1 and 2 both have access to project 3
		// See fixtures: user 1 has permission 2, user 2 has permission 0 on project 3
		task := &models.Task{
			ID:        32, // Task 32 is in project 3 (owned by user 3)
			ProjectID: 3,
			Assignees: []*user.User{
				{ID: 1}, // User 1 has access to project 3 (permission 2)
				{ID: 2}, // User 2 has access to project 3 (permission 0)
			},
		}

		updatedTask, err := ts.Update(s, task, u)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		// Verify both assignees were added
		db.AssertExists(t, "task_assignees", map[string]interface{}{
			"task_id": 32,
			"user_id": 1,
		}, false)
		db.AssertExists(t, "task_assignees", map[string]interface{}{
			"task_id": 32,
			"user_id": 2,
		}, false)

		// Both users should be in the assignees list
		assert.Equal(t, 2, len(updatedTask.Assignees))
	})

	t.Run("remove multiple assignees at once", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 1}
		ts := NewTaskService(testEngine)

		// First add multiple assignees
		task := &models.Task{
			ID:        1,
			ProjectID: 1,
			Assignees: []*user.User{
				{ID: 1},
			},
		}

		_, err := ts.Update(s, task, u)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		// Now remove all assignees by passing empty array
		s2 := db.NewSession()
		defer s2.Close()

		task.Assignees = []*user.User{}
		updatedTask, err := ts.Update(s2, task, u)
		require.NoError(t, err)
		require.NoError(t, s2.Commit())

		// Verify all assignees were removed
		assert.Len(t, updatedTask.Assignees, 0)
		db.AssertMissing(t, "task_assignees", map[string]interface{}{
			"task_id": 1,
		})
	})
}

func TestTaskService_Assignee_PersistenceAcrossProjectMove(t *testing.T) {
	t.Run("assignees should persist when task moves between projects", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 1}
		ts := NewTaskService(testEngine)

		// Create task in project 1 with assignee
		task := &models.Task{
			Title:     "Task to Move",
			ProjectID: 1,
			Assignees: []*user.User{
				{ID: 1},
			},
		}

		createdTask, err := ts.CreateWithOptions(s, task, u, true, true, false)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		// Verify assignee exists
		db.AssertExists(t, "task_assignees", map[string]interface{}{
			"task_id": createdTask.ID,
			"user_id": 1,
		}, false)

		// Move task to project 3 (user 1 has access to both project 1 and 3)
		s2 := db.NewSession()
		defer s2.Close()

		createdTask.ProjectID = 3
		movedTask, err := ts.Update(s2, createdTask, u)
		require.NoError(t, err)
		require.NoError(t, s2.Commit())

		// Verify assignee still exists after move
		assert.Len(t, movedTask.Assignees, 1)
		db.AssertExists(t, "task_assignees", map[string]interface{}{
			"task_id": createdTask.ID,
			"user_id": 1,
		}, false)
	})
}

func TestTaskService_Assignee_ConcurrentUpdates(t *testing.T) {
	t.Run("concurrent assignee updates should not create duplicates", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		u := &user.User{ID: 1}
		ts := NewTaskService(testEngine)

		// Create a task with assignees
		s1 := db.NewSession()
		defer s1.Close()

		task := &models.Task{
			Title:     "Concurrent Update Test",
			ProjectID: 1,
			Assignees: []*user.User{
				{ID: 1},
			},
		}

		createdTask, err := ts.CreateWithOptions(s1, task, u, true, true, false)
		require.NoError(t, err)
		require.NoError(t, s1.Commit())

		// Simulate concurrent updates from two sessions
		s2 := db.NewSession()
		defer s2.Close()

		s3 := db.NewSession()
		defer s3.Close()

		// Both sessions try to update the same task's assignees
		task1 := &models.Task{
			ID:        createdTask.ID,
			ProjectID: 1,
			Assignees: []*user.User{
				{ID: 1},
			},
		}

		task2 := &models.Task{
			ID:        createdTask.ID,
			ProjectID: 1,
			Assignees: []*user.User{
				{ID: 1},
			},
		}

		// Update from session 2
		_, err = ts.Update(s2, task1, u)
		require.NoError(t, err)
		require.NoError(t, s2.Commit())

		// Update from session 3
		_, err = ts.Update(s3, task2, u)
		require.NoError(t, err)
		require.NoError(t, s3.Commit())

		// Verify no duplicate assignees were created
		s4 := db.NewSession()
		defer s4.Close()

		var count int64
		count, err = s4.Where("task_id = ? AND user_id = ?", createdTask.ID, 1).
			Count(&models.TaskAssginee{})
		require.NoError(t, err)

		// Should only have one assignee record, not duplicates
		assert.Equal(t, int64(1), count, "Should not create duplicate assignee records")
	})
}
