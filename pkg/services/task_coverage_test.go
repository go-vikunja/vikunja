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
