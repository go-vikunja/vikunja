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
)

func TestTaskService_Delete(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	ts := NewTaskService(testEngine)
	auth := &user.User{ID: 1}

	t.Run("should delete a task", func(t *testing.T) {
		task := &models.Task{ID: 1}
		err := ts.Delete(s, task, auth)
		assert.NoError(t, err)

		// Check that the task is deleted
		has, err := s.ID(task.ID).Get(&models.Task{})
		assert.NoError(t, err)
		assert.False(t, has)
	})

	t.Run("should not delete a non-existent task", func(t *testing.T) {
		task := &models.Task{ID: 9999}
		err := ts.Delete(s, task, auth)
		assert.Error(t, err)
	})

	t.Run("Permissions check", func(t *testing.T) {
		otherUserAuth := &user.User{ID: 2}
		t.Run("Forbidden", func(t *testing.T) {
			taskToDelete := &models.Task{ID: 1}
			err := ts.Delete(s, taskToDelete, otherUserAuth)
			assert.Error(t, err, "should not be able to delete task")
			assert.Equal(t, ErrAccessDenied, err)
		})
		t.Run("Shared Via Team readonly", func(t *testing.T) {
			taskToDelete := &models.Task{ID: 15}
			err := ts.Delete(s, taskToDelete, auth)
			assert.Error(t, err, "should not be able to delete task")
			assert.Equal(t, ErrAccessDenied, err)
		})
		t.Run("Shared Via Team write", func(t *testing.T) {
			taskToDelete := &models.Task{ID: 16}
			err := ts.Delete(s, taskToDelete, auth)
			assert.NoError(t, err)
		})
		t.Run("Shared Via Team admin", func(t *testing.T) {
			taskToDelete := &models.Task{ID: 17}
			err := ts.Delete(s, taskToDelete, auth)
			assert.NoError(t, err)
		})

		t.Run("Shared Via User readonly", func(t *testing.T) {
			taskToDelete := &models.Task{ID: 18}
			err := ts.Delete(s, taskToDelete, auth)
			assert.Error(t, err, "should not be able to delete task")
			assert.Equal(t, ErrAccessDenied, err)
		})
		t.Run("Shared Via User write", func(t *testing.T) {
			taskToDelete := &models.Task{ID: 19}
			err := ts.Delete(s, taskToDelete, auth)
			assert.NoError(t, err)
		})
		t.Run("Shared Via User admin", func(t *testing.T) {
			taskToDelete := &models.Task{ID: 20}
			err := ts.Delete(s, taskToDelete, auth)
			assert.NoError(t, err)
		})

		t.Run("Shared Via Parent Project Team readonly", func(t *testing.T) {
			taskToDelete := &models.Task{ID: 21}
			err := ts.Delete(s, taskToDelete, auth)
			assert.Error(t, err, "should not be able to delete task")
			assert.Equal(t, ErrAccessDenied, err)
		})
		t.Run("Shared Via Parent Project Team write", func(t *testing.T) {
			taskToDelete := &models.Task{ID: 22}
			err := ts.Delete(s, taskToDelete, auth)
			assert.NoError(t, err)
		})
		t.Run("Shared Via Parent Project Team admin", func(t *testing.T) {
			taskToDelete := &models.Task{ID: 23}
			err := ts.Delete(s, taskToDelete, auth)
			assert.NoError(t, err)
		})

		t.Run("Shared Via Parent Project User readonly", func(t *testing.T) {
			taskToDelete := &models.Task{ID: 24}
			err := ts.Delete(s, taskToDelete, auth)
			assert.Error(t, err, "should not be able to delete task")
			assert.Equal(t, ErrAccessDenied, err)
		})
		t.Run("Shared Via Parent Project User write", func(t *testing.T) {
			taskToDelete := &models.Task{ID: 25}
			err := ts.Delete(s, taskToDelete, auth)
			assert.NoError(t, err)
		})
		t.Run("Shared Via Parent Project User admin", func(t *testing.T) {
			taskToDelete := &models.Task{ID: 26}
			err := ts.Delete(s, taskToDelete, auth)
			assert.NoError(t, err)
		})
	})
}
