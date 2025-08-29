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
)

func TestTaskService_Update(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	ts := NewTaskService(testEngine)
	u := &user.User{ID: 1}

	t.Run("should update a task", func(t *testing.T) {
		task := &models.Task{
			Title:       "Test Task",
			Created:     time.Now(),
			Updated:     time.Now(),
			CreatedByID: 1,
			ProjectID:   1,
		}
		_, err := s.Insert(task)
		assert.NoError(t, err)

		task.Title = "Updated Task Title"
		updatedTask, err := ts.Update(s, task, u)
		assert.NoError(t, err)

		var fromDB models.Task
		has, err := s.ID(updatedTask.ID).Get(&fromDB)
		assert.NoError(t, err)
		assert.True(t, has)
		assert.Equal(t, "Updated Task Title", fromDB.Title)
	})

	t.Run("should not update a task without access", func(t *testing.T) {
		otherUser := &user.User{ID: 2}
		taskToUpdate := &models.Task{
			ID:          1,
			Title:       "Updated Title by other user",
			CreatedByID: 1,
			ProjectID:   1,
		}
		_, err := ts.Update(s, taskToUpdate, otherUser)
		assert.Error(t, err, "should not be able to update task")
	})
}

func TestTaskService_GetByID(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	ts := NewTaskService(testEngine)
	u := &user.User{ID: 1}

	t.Run("should get a task by id", func(t *testing.T) {
		task, err := ts.GetByID(s, 1, u)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), task.ID)
	})

	t.Run("should not get a task without access", func(t *testing.T) {
		otherUser := &user.User{ID: 2}
		_, err := ts.GetByID(s, 1, otherUser)
		assert.ErrorIs(t, err, ErrAccessDenied)
	})

	t.Run("should return an error for a non-existent task", func(t *testing.T) {
		_, err := ts.GetByID(s, 9999, u)
		assert.Error(t, err)
	})
}

func TestTaskService_GetAllByProject(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	ts := NewTaskService(testEngine)
	u := &user.User{ID: 1}

	t.Run("should get all tasks in a project", func(t *testing.T) {
		tasks, err := ts.GetAllByProject(s, 1, u)
		assert.NoError(t, err)
		assert.Len(t, tasks, 3)
	})

	t.Run("should not get tasks without access", func(t *testing.T) {
		otherUser := &user.User{ID: 2}
		_, err := ts.GetAllByProject(s, 1, otherUser)
		assert.ErrorIs(t, err, ErrAccessDenied)
	})

	t.Run("should return an empty slice for a project with no tasks", func(t *testing.T) {
		// Project 2 has no tasks
		tasks, err := ts.GetAllByProject(s, 2, u)
		assert.NoError(t, err)
		assert.Len(t, tasks, 0)
	})
}
