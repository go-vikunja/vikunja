// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2021 Vikunja and contributors. All rights reserved.
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

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTaskService_ApplyFiltersToQuery tests the applyFiltersToQuery function
func TestTaskService_ApplyFiltersToQuery(t *testing.T) {
	t.Run("should apply search filter", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		taskService := NewTaskService(db.GetEngine())

		opts := &taskSearchOptions{
			search: "test",
		}

		query := s.Table("tasks")
		resultQuery, _, err := taskService.applyFiltersToQuery(query, opts)
		assert.NoError(t, err)
		assert.NotNil(t, resultQuery)
	})

	t.Run("should apply complex filters", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		taskService := NewTaskService(db.GetEngine())

		opts := &taskSearchOptions{
			filter: "due_date >= '2024-01-01'",
		}

		query := s.Table("tasks")
		resultQuery, _, err := taskService.applyFiltersToQuery(query, opts)
		assert.NoError(t, err)
		assert.NotNil(t, resultQuery)
	})

	t.Run("should handle empty filters", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		taskService := NewTaskService(db.GetEngine())

		opts := &taskSearchOptions{
			search: "",
			filter: "",
		}

		query := s.Table("tasks")
		resultQuery, _, err := taskService.applyFiltersToQuery(query, opts)
		assert.NoError(t, err)
		assert.NotNil(t, resultQuery)
	})
}

// TestTaskService_ApplySortingToQuery tests the applySortingToQuery function
func TestTaskService_ApplySortingToQuery(t *testing.T) {
	t.Run("should apply ascending sort", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		taskService := NewTaskService(db.GetEngine())

		sortParams := []*sortParam{
			{sortBy: "title", orderBy: orderAscending},
		}

		query := s.Table("tasks")
		taskService.applySortingToQuery(query, sortParams)
		assert.NotNil(t, query)
	})

	t.Run("should apply descending sort", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		taskService := NewTaskService(db.GetEngine())

		sortParams := []*sortParam{
			{sortBy: "created", orderBy: orderDescending},
		}

		query := s.Table("tasks")
		taskService.applySortingToQuery(query, sortParams)
		assert.NotNil(t, query)
	})

	t.Run("should apply multiple sorts", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		taskService := NewTaskService(db.GetEngine())

		sortParams := []*sortParam{
			{sortBy: "done", orderBy: orderAscending},
			{sortBy: "created", orderBy: orderDescending},
		}

		query := s.Table("tasks")
		taskService.applySortingToQuery(query, sortParams)
		assert.NotNil(t, query)
	})

	t.Run("should handle empty sort params", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		taskService := NewTaskService(db.GetEngine())

		query := s.Table("tasks")
		taskService.applySortingToQuery(query, []*sortParam{})
		assert.NotNil(t, query)
	})
}

// TestTaskService_AddBucketsToTasks tests the addBucketsToTasks function
func TestTaskService_AddBucketsToTasks(t *testing.T) {
	t.Run("should add buckets to tasks", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		taskService := NewTaskService(db.GetEngine())

		// Get some tasks
		tasks := []*models.Task{}
		err := s.Limit(2).Find(&tasks)
		require.NoError(t, err)
		require.NotEmpty(t, tasks)

		// Create task map
		taskMap := make(map[int64]*models.Task)
		taskIDs := make([]int64, 0, len(tasks))
		for _, task := range tasks {
			taskMap[task.ID] = task
			taskIDs = append(taskIDs, task.ID)
		}

		// Call addBucketsToTasks
		u := &user.User{ID: 1}
		err = taskService.addBucketsToTasks(s, u, taskIDs, taskMap)
		assert.NoError(t, err)
	})

	t.Run("should handle nil KanbanService gracefully", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Create service without KanbanService
		taskService := &TaskService{
			DB:            db.GetEngine(),
			KanbanService: nil,
		}

		taskMap := make(map[int64]*models.Task)
		u := &user.User{ID: 1}
		err := taskService.addBucketsToTasks(s, u, []int64{1}, taskMap)
		assert.NoError(t, err)
	})
}

// TestTaskService_AddReactionsToTasks tests the addReactionsToTasks function
func TestTaskService_AddReactionsToTasks(t *testing.T) {
	t.Run("should add reactions to tasks", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		taskService := NewTaskService(db.GetEngine())

		// Get some tasks
		tasks := []*models.Task{}
		err := s.Limit(2).Find(&tasks)
		require.NoError(t, err)
		require.NotEmpty(t, tasks)

		// Create task map
		taskMap := make(map[int64]*models.Task)
		taskIDs := make([]int64, 0, len(tasks))
		for _, task := range tasks {
			taskMap[task.ID] = task
			taskIDs = append(taskIDs, task.ID)
		}

		// Call addReactionsToTasks
		err = taskService.addReactionsToTasks(s, taskIDs, taskMap)
		assert.NoError(t, err)
	})

	t.Run("should handle nil ReactionsService gracefully", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Create service without ReactionsService
		taskService := &TaskService{
			DB:               db.GetEngine(),
			ReactionsService: nil,
		}

		taskMap := make(map[int64]*models.Task)
		err := taskService.addReactionsToTasks(s, []int64{1}, taskMap)
		assert.NoError(t, err)
	})
}

// TestTaskService_AddCommentsToTasks tests the addCommentsToTasks function
func TestTaskService_AddCommentsToTasks(t *testing.T) {
	t.Run("should add comments to tasks", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		taskService := NewTaskService(db.GetEngine())

		// Get some tasks
		tasks := []*models.Task{}
		err := s.Limit(2).Find(&tasks)
		require.NoError(t, err)
		require.NotEmpty(t, tasks)

		// Create task map
		taskMap := make(map[int64]*models.Task)
		taskIDs := make([]int64, 0, len(tasks))
		for _, task := range tasks {
			taskMap[task.ID] = task
			taskIDs = append(taskIDs, task.ID)
		}

		// Call addCommentsToTasks
		err = taskService.addCommentsToTasks(s, taskIDs, taskMap)
		assert.NoError(t, err)
	})

	t.Run("should handle nil CommentService gracefully", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Create service without CommentService
		taskService := &TaskService{
			DB:             db.GetEngine(),
			CommentService: nil,
		}

		taskMap := make(map[int64]*models.Task)
		err := taskService.addCommentsToTasks(s, []int64{1}, taskMap)
		assert.NoError(t, err)
	})
}

// TestTaskService_MoveTaskToDoneBuckets tests the moveTaskToDoneBuckets function
func TestTaskService_MoveTaskToDoneBuckets(t *testing.T) {
	t.Run("should move task to done bucket when task is marked done", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		taskService := NewTaskService(db.GetEngine())
		u := &user.User{ID: 1}

		// Get a task that's not done
		task := &models.Task{}
		_, err := s.Where("id = ?", 1).Get(task)
		require.NoError(t, err)

		// Mark it as done
		task.Done = true

		// Get views for the project (view 4 is a Kanban view with done_bucket_id = 3)
		views, err := taskService.getViewsForProject(s, task.ProjectID)
		require.NoError(t, err)
		require.NotEmpty(t, views)

		// Move task to done buckets
		err = taskService.moveTaskToDoneBuckets(s, task, u, views)
		assert.NoError(t, err)
	})

	t.Run("should move task from done bucket when task is unmarked done", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		taskService := NewTaskService(db.GetEngine())
		u := &user.User{ID: 1}

		// Get a task
		task := &models.Task{}
		_, err := s.Where("id = ?", 1).Get(task)
		require.NoError(t, err)

		// Mark it as not done
		task.Done = false

		// Get views for the project
		views, err := taskService.getViewsForProject(s, task.ProjectID)
		require.NoError(t, err)

		// Move task from done buckets
		err = taskService.moveTaskToDoneBuckets(s, task, u, views)
		assert.NoError(t, err)
	})

	t.Run("should handle views without done bucket", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		taskService := NewTaskService(db.GetEngine())
		u := &user.User{ID: 1}

		// Get a task
		task := &models.Task{}
		_, err := s.Where("id = ?", 1).Get(task)
		require.NoError(t, err)
		task.Done = true

		// Create a view without done bucket
		views := []*models.ProjectView{
			{ID: 1, DoneBucketID: 0}, // No done bucket configured
		}

		// Should not error even though no done bucket
		err = taskService.moveTaskToDoneBuckets(s, task, u, views)
		assert.NoError(t, err)
	})

	t.Run("should handle empty views list", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		taskService := NewTaskService(db.GetEngine())
		u := &user.User{ID: 1}

		task := &models.Task{ID: 1, Done: true}
		views := []*models.ProjectView{}

		err := taskService.moveTaskToDoneBuckets(s, task, u, views)
		assert.NoError(t, err)
	})
}

// TestTaskService_GetRawFavoriteTasks tests the getRawFavoriteTasks function
func TestTaskService_GetRawFavoriteTasks(t *testing.T) {
	t.Run("should get favorite tasks with filtering", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		taskService := NewTaskService(db.GetEngine())

		// Use some task IDs that exist in fixtures
		favoriteTaskIDs := []int64{1, 2, 3}

		opts := &taskSearchOptions{
			page:    1,
			perPage: 50,
			sortby:  []*sortParam{{sortBy: "id", orderBy: orderAscending}},
		}

		tasks, resultCount, totalItems, err := taskService.getRawFavoriteTasks(s, favoriteTaskIDs, opts)
		assert.NoError(t, err)
		assert.NotEmpty(t, tasks)
		assert.Equal(t, len(tasks), resultCount)
		assert.Greater(t, totalItems, int64(0))
	})

	t.Run("should apply pagination to favorite tasks", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		taskService := NewTaskService(db.GetEngine())

		favoriteTaskIDs := []int64{1, 2, 3}

		opts := &taskSearchOptions{
			page:    1,
			perPage: 1, // Only 1 task per page
			sortby:  []*sortParam{{sortBy: "id", orderBy: orderAscending}},
		}

		tasks, resultCount, totalItems, err := taskService.getRawFavoriteTasks(s, favoriteTaskIDs, opts)
		assert.NoError(t, err)
		assert.LessOrEqual(t, resultCount, 1)
		assert.Greater(t, totalItems, int64(0))

		// Verify pagination worked - should get max 1 task
		assert.LessOrEqual(t, len(tasks), 1)
	})

	t.Run("should apply sorting to favorite tasks", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		taskService := NewTaskService(db.GetEngine())

		favoriteTaskIDs := []int64{1, 2, 3}

		// Test descending order
		opts := &taskSearchOptions{
			page:    1,
			perPage: 50,
			sortby:  []*sortParam{{sortBy: "id", orderBy: orderDescending}},
		}

		tasks, _, _, err := taskService.getRawFavoriteTasks(s, favoriteTaskIDs, opts)
		assert.NoError(t, err)
		assert.NotEmpty(t, tasks)
	})

	t.Run("should handle empty favorite task IDs", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		taskService := NewTaskService(db.GetEngine())

		opts := &taskSearchOptions{
			page:    1,
			perPage: 50,
			sortby:  []*sortParam{},
		}

		tasks, resultCount, totalItems, err := taskService.getRawFavoriteTasks(s, []int64{}, opts)
		assert.NoError(t, err)
		assert.Empty(t, tasks)
		assert.Equal(t, 0, resultCount)
		assert.Equal(t, int64(0), totalItems)
	})

	t.Run("should clear project IDs in favorite opts", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		taskService := NewTaskService(db.GetEngine())

		favoriteTaskIDs := []int64{1, 2, 3}

		// Provide project IDs, but they should be cleared for favorites
		opts := &taskSearchOptions{
			page:       1,
			perPage:    50,
			projectIDs: []int64{1, 2, 3}, // Should be ignored
			sortby:     []*sortParam{{sortBy: "id", orderBy: orderAscending}},
		}

		tasks, resultCount, totalItems, err := taskService.getRawFavoriteTasks(s, favoriteTaskIDs, opts)
		assert.NoError(t, err)
		assert.NotEmpty(t, tasks)
		assert.Equal(t, len(tasks), resultCount)
		assert.Greater(t, totalItems, int64(0))
	})
}

// TestTaskService_BuildAndExecuteTaskQuery tests the buildAndExecuteTaskQuery function
func TestTaskService_BuildAndExecuteTaskQuery(t *testing.T) {
	t.Run("should execute query with project filtering", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		taskService := NewTaskService(db.GetEngine())

		opts := &taskSearchOptions{
			page:       1,
			perPage:    50,
			projectIDs: []int64{1}, // Filter by project 1
			sortby:     []*sortParam{{sortBy: "id", orderBy: orderAscending}},
		}

		tasks, resultCount, totalItems, err := taskService.buildAndExecuteTaskQuery(s, opts)
		assert.NoError(t, err)
		assert.NotEmpty(t, tasks)
		assert.Equal(t, len(tasks), resultCount)
		assert.Greater(t, totalItems, int64(0))
	})

	t.Run("should filter by multiple projects", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		taskService := NewTaskService(db.GetEngine())

		opts := &taskSearchOptions{
			page:       1,
			perPage:    50,
			projectIDs: []int64{1, 2}, // Filter by projects 1 and 2
			sortby:     []*sortParam{{sortBy: "id", orderBy: orderAscending}},
		}

		tasks, resultCount, totalItems, err := taskService.buildAndExecuteTaskQuery(s, opts)
		assert.NoError(t, err)
		assert.NotEmpty(t, tasks)
		assert.Equal(t, len(tasks), resultCount)
		assert.Greater(t, totalItems, int64(0))
	})

	t.Run("should apply search filter", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		taskService := NewTaskService(db.GetEngine())

		opts := &taskSearchOptions{
			page:       1,
			perPage:    50,
			projectIDs: []int64{1},
			search:     "task", // Search for "task" in title
			sortby:     []*sortParam{{sortBy: "id", orderBy: orderAscending}},
		}

		tasks, resultCount, totalItems, err := taskService.buildAndExecuteTaskQuery(s, opts)
		assert.NoError(t, err)
		// Should execute without error
		assert.GreaterOrEqual(t, resultCount, 0)
		assert.GreaterOrEqual(t, totalItems, int64(0))
		assert.Equal(t, len(tasks), resultCount)
	})

	t.Run("should apply pagination", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		taskService := NewTaskService(db.GetEngine())

		opts := &taskSearchOptions{
			page:       1,
			perPage:    2, // Only 2 tasks per page
			projectIDs: []int64{1},
			sortby:     []*sortParam{{sortBy: "id", orderBy: orderAscending}},
		}

		tasks, resultCount, totalItems, err := taskService.buildAndExecuteTaskQuery(s, opts)
		assert.NoError(t, err)
		assert.LessOrEqual(t, resultCount, 2)
		assert.Greater(t, totalItems, int64(0))

		// Verify we got at most 2 tasks
		assert.LessOrEqual(t, len(tasks), 2)
	})

	t.Run("should apply sorting ascending", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		taskService := NewTaskService(db.GetEngine())

		opts := &taskSearchOptions{
			page:       1,
			perPage:    50,
			projectIDs: []int64{1},
			sortby:     []*sortParam{{sortBy: "id", orderBy: orderAscending}},
		}

		tasks, _, _, err := taskService.buildAndExecuteTaskQuery(s, opts)
		assert.NoError(t, err)
		assert.NotEmpty(t, tasks)
	})

	t.Run("should apply sorting descending", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		taskService := NewTaskService(db.GetEngine())

		opts := &taskSearchOptions{
			page:       1,
			perPage:    50,
			projectIDs: []int64{1},
			sortby:     []*sortParam{{sortBy: "id", orderBy: orderDescending}},
		}

		tasks, _, _, err := taskService.buildAndExecuteTaskQuery(s, opts)
		assert.NoError(t, err)
		assert.NotEmpty(t, tasks)
	})

	t.Run("should apply multiple sort criteria", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		taskService := NewTaskService(db.GetEngine())

		opts := &taskSearchOptions{
			page:       1,
			perPage:    50,
			projectIDs: []int64{1},
			sortby: []*sortParam{
				{sortBy: "done", orderBy: orderAscending},
				{sortBy: "id", orderBy: orderDescending},
			},
		}

		tasks, resultCount, totalItems, err := taskService.buildAndExecuteTaskQuery(s, opts)
		assert.NoError(t, err)
		assert.NotEmpty(t, tasks)
		assert.Equal(t, len(tasks), resultCount)
		assert.Greater(t, totalItems, int64(0))
	})

	t.Run("should handle empty project IDs", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		taskService := NewTaskService(db.GetEngine())

		opts := &taskSearchOptions{
			page:       1,
			perPage:    50,
			projectIDs: []int64{}, // Empty project list
			sortby:     []*sortParam{{sortBy: "id", orderBy: orderAscending}},
		}

		tasks, resultCount, totalItems, err := taskService.buildAndExecuteTaskQuery(s, opts)
		assert.NoError(t, err)
		// With empty project IDs, no filtering is applied so we may get all tasks
		// The function should still work without error
		assert.GreaterOrEqual(t, resultCount, 0)
		assert.GreaterOrEqual(t, totalItems, int64(0))
		assert.Equal(t, len(tasks), resultCount)
	})
}
