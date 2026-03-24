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

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskSoftDelete(t *testing.T) {
	t.Run("delete moves task to trash", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()

		task := &Task{ID: 1}
		u := &user.User{ID: 1}
		err := task.Delete(s, u)
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)
		s.Close()

		// Verify task still exists in DB with deleted_at set
		s2 := db.NewSession()
		var found Task
		exists, err := s2.ID(1).Get(&found)
		s2.Close()
		require.NoError(t, err)
		assert.True(t, exists)
		assert.NotNil(t, found.DeletedAt)
	})

	t.Run("delete preserves relationships", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()

		task := &Task{ID: 1}
		u := &user.User{ID: 1}
		err := task.Delete(s, u)
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)
		s.Close()

		// Labels should still exist
		db.AssertExists(t, "label_tasks", map[string]interface{}{
			"task_id": 1,
		}, false)
	})
}

func TestTaskHardDelete(t *testing.T) {
	t.Run("hard delete removes task and relationships", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()

		task := &Task{ID: 1}
		u := &user.User{ID: 1}
		err := task.HardDelete(s, u)
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)
		s.Close()

		db.AssertMissing(t, "tasks", map[string]interface{}{
			"id": 1,
		})
	})
}

func TestTaskRestore(t *testing.T) {
	t.Run("restore clears deleted_at", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()

		task := &Task{ID: 49}
		u := &user.User{ID: 1}
		err := task.Restore(s, u)
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)
		s.Close()

		s2 := db.NewSession()
		var found Task
		exists, err := s2.ID(49).Get(&found)
		s2.Close()
		require.NoError(t, err)
		assert.True(t, exists)
		assert.Nil(t, found.DeletedAt)
	})

	t.Run("restore non-trashed task returns error", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()

		task := &Task{ID: 1}
		u := &user.User{ID: 1}
		err := task.Restore(s, u)
		require.Error(t, err)
		assert.True(t, IsErrTaskIsNotTrashed(err))
		s.Close()
	})
}

func TestTrashedTasksExcludedFromQueries(t *testing.T) {
	t.Run("trashed task not found by GetTaskSimple", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()

		_, err := GetTaskSimple(s, &Task{ID: 49})
		require.Error(t, err)
		assert.True(t, IsErrTaskDoesNotExist(err))
		s.Close()
	})
}

func TestGetTrashedTasks(t *testing.T) {
	t.Run("returns only trashed tasks", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()

		u := &user.User{ID: 1}
		tasks, count, err := GetTrashedTasks(s, u, 0, 1, 50)
		require.NoError(t, err)
		assert.Greater(t, count, int64(0))
		for _, task := range tasks {
			assert.NotNil(t, task.DeletedAt)
		}
		s.Close()
	})

	t.Run("filter by project", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()

		u := &user.User{ID: 1}
		tasks, _, err := GetTrashedTasks(s, u, 1, 1, 50)
		require.NoError(t, err)
		for _, task := range tasks {
			assert.Equal(t, int64(1), task.ProjectID)
		}
		s.Close()
	})
}
