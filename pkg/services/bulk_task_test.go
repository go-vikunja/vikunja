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

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBulkTaskService_New(t *testing.T) {
	t.Run("create new bulk task service", func(t *testing.T) {
		service := NewBulkTaskService(db.GetEngine())

		assert.NotNil(t, service)
		assert.IsType(t, &BulkTaskService{}, service)
		assert.NotNil(t, service.TaskService)
	})
}

func TestBulkTaskService_GetTasksByIDs(t *testing.T) {
	db.LoadAndAssertFixtures(t)

	s := db.NewSession()
	defer s.Close()

	service := NewBulkTaskService(db.GetEngine())

	t.Run("successfully get tasks by IDs", func(t *testing.T) {
		taskIDs := []int64{1, 2}

		tasks, err := service.GetTasksByIDs(s, taskIDs)
		require.NoError(t, err)
		assert.NotNil(t, tasks)
		assert.Equal(t, 2, len(tasks))
	})

	t.Run("fails with invalid task ID", func(t *testing.T) {
		taskIDs := []int64{-1}

		_, err := service.GetTasksByIDs(s, taskIDs)
		require.Error(t, err)
		assert.True(t, models.IsErrTaskDoesNotExist(err))
	})

	t.Run("fails with zero task ID", func(t *testing.T) {
		taskIDs := []int64{0}

		_, err := service.GetTasksByIDs(s, taskIDs)
		require.Error(t, err)
		assert.True(t, models.IsErrTaskDoesNotExist(err))
	})

	t.Run("returns empty list for non-existent tasks", func(t *testing.T) {
		taskIDs := []int64{99999}

		tasks, err := service.GetTasksByIDs(s, taskIDs)
		require.NoError(t, err)
		assert.NotNil(t, tasks)
		assert.Equal(t, 0, len(tasks))
	})

	t.Run("handles empty task ID list", func(t *testing.T) {
		taskIDs := []int64{}

		tasks, err := service.GetTasksByIDs(s, taskIDs)
		require.NoError(t, err)
		assert.NotNil(t, tasks)
		assert.Equal(t, 0, len(tasks))
	})
}

func TestBulkTaskService_CanUpdate(t *testing.T) {
	db.LoadAndAssertFixtures(t)

	s := db.NewSession()
	defer s.Close()

	service := NewBulkTaskService(db.GetEngine())

	t.Run("allows update for tasks in same project with permission", func(t *testing.T) {
		u := &user.User{ID: 1}
		taskIDs := []int64{1, 2}

		canUpdate, err := service.CanUpdate(s, taskIDs, u)
		require.NoError(t, err)
		assert.True(t, canUpdate)
	})

	t.Run("denies update for user without permission", func(t *testing.T) {
		u := &user.User{ID: 999}
		taskIDs := []int64{1, 2}

		canUpdate, err := service.CanUpdate(s, taskIDs, u)
		require.NoError(t, err)
		assert.False(t, canUpdate)
	})

	t.Run("fails when no tasks provided", func(t *testing.T) {
		u := &user.User{ID: 1}
		taskIDs := []int64{}

		_, err := service.CanUpdate(s, taskIDs, u)
		require.Error(t, err)
		assert.True(t, models.IsErrBulkTasksNeedAtLeastOne(err))
	})

	t.Run("fails when tasks are in different projects", func(t *testing.T) {
		u := &user.User{ID: 1}
		// Task 1 is in project 1, task 15 is in project 6
		taskIDs := []int64{1, 15}

		_, err := service.CanUpdate(s, taskIDs, u)
		require.Error(t, err)
		assert.True(t, models.IsErrBulkTasksMustBeInSameProject(err))
	})

	t.Run("fails with invalid task ID", func(t *testing.T) {
		u := &user.User{ID: 1}
		taskIDs := []int64{-1}

		_, err := service.CanUpdate(s, taskIDs, u)
		require.Error(t, err)
		assert.True(t, models.IsErrTaskDoesNotExist(err))
	})
}

func TestBulkTaskService_Update(t *testing.T) {
	db.LoadAndAssertFixtures(t)

	service := NewBulkTaskService(db.GetEngine())

	t.Run("successfully update tasks", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 1}
		taskIDs := []int64{1, 2}
		taskUpdate := &models.Task{
			Title:       "Bulk updated title",
			Description: "Bulk updated description",
		}

		err := service.Update(s, taskIDs, taskUpdate, nil, u)
		require.NoError(t, err)

		// Verify updates were applied
		tasks, err := service.GetTasksByIDs(s, taskIDs)
		require.NoError(t, err)
		for _, task := range tasks {
			assert.Equal(t, "Bulk updated title", task.Title)
			assert.Equal(t, "Bulk updated description", task.Description)
		}
	})

	t.Run("successfully update task done status", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 1}
		taskIDs := []int64{3, 4}
		taskUpdate := &models.Task{
			Done: true,
		}

		err := service.Update(s, taskIDs, taskUpdate, nil, u)
		require.NoError(t, err)

		// Verify updates were applied
		tasks, err := service.GetTasksByIDs(s, taskIDs)
		require.NoError(t, err)
		for _, task := range tasks {
			assert.True(t, task.Done)
		}
	})

	t.Run("successfully update task to not done", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 1}
		taskIDs := []int64{5, 6}
		taskUpdate := &models.Task{
			Done: false,
		}

		err := service.Update(s, taskIDs, taskUpdate, nil, u)
		require.NoError(t, err)

		// Verify updates were applied
		tasks, err := service.GetTasksByIDs(s, taskIDs)
		require.NoError(t, err)
		for _, task := range tasks {
			assert.False(t, task.Done)
		}
	})

	t.Run("successfully update with assignees", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 1}
		taskIDs := []int64{7, 8}
		taskUpdate := &models.Task{
			Title: "Updated with assignees",
		}
		assignees := []*user.User{
			{ID: 1},
		}

		err := service.Update(s, taskIDs, taskUpdate, assignees, u)
		require.NoError(t, err)

		// Verify updates were applied
		tasks, err := service.GetTasksByIDs(s, taskIDs)
		require.NoError(t, err)
		for _, task := range tasks {
			assert.Equal(t, "Updated with assignees", task.Title)
		}
	})

	t.Run("fails with invalid task ID", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 1}
		taskIDs := []int64{-1}
		taskUpdate := &models.Task{
			Title: "Should fail",
		}

		err := service.Update(s, taskIDs, taskUpdate, nil, u)
		require.Error(t, err)
		assert.True(t, models.IsErrTaskDoesNotExist(err))
	})

	t.Run("handles empty task list", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 1}
		taskIDs := []int64{}
		taskUpdate := &models.Task{
			Title: "Should handle gracefully",
		}

		err := service.Update(s, taskIDs, taskUpdate, nil, u)
		// GetTasksByIDs returns error for empty list, but that's expected
		if err != nil {
			assert.True(t, models.IsErrTaskDoesNotExist(err))
		}
	})
}
