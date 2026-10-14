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
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func trashTaskForTest(t *testing.T, taskID int64, a web.Auth) {
	s := db.NewSession()
	defer s.Close()
	require.NoError(t, (&Task{ID: taskID}).Delete(s, a))
	require.NoError(t, s.Commit())
}

func setTaskColumnsForTest(t *testing.T, taskID int64, cols map[string]any) {
	s := db.NewAutocommitSession()
	defer s.Close()
	_, err := s.Table("tasks").Where("id = ?", taskID).Update(cols)
	require.NoError(t, err)
}

func restoreTaskForTest(t *testing.T, taskID int64, a web.Auth) (*Task, error) {
	s := db.NewSession()
	defer s.Close()

	task := &Task{ID: taskID}
	can, err := task.CanRestore(s, a)
	if err != nil {
		return nil, err
	}
	if !can {
		return nil, ErrGenericForbidden{}
	}
	if err := task.Restore(s, a); err != nil {
		return nil, err
	}
	require.NoError(t, s.Commit())
	events.DispatchPending(context.Background(), s)
	return task, nil
}

func TestTask_Restore(t *testing.T) {
	u1 := &user.User{ID: 1}

	t.Run("restores into the default bucket of every view", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		task, err := restoreTaskForTest(t, 51, u1)
		require.NoError(t, err)
		events.AssertDispatched(t, &TaskRestoredEvent{})

		db.AssertExists(t, "tasks", map[string]interface{}{"id": 51, "deleted_at": nil, "deleted_by_id": nil}, false)
		assert.True(t, task.Updated.After(time.Date(2018, 12, 1, 1, 12, 4, 0, time.UTC)))

		// View 4 is project 1's manual kanban with default bucket 1
		db.AssertExists(t, "task_positions", map[string]interface{}{"task_id": 51, "project_view_id": 1}, false)
		db.AssertExists(t, "task_positions", map[string]interface{}{"task_id": 51, "project_view_id": 4}, false)
		db.AssertExists(t, "task_buckets", map[string]interface{}{"task_id": 51, "project_view_id": 4, "bucket_id": 1}, false)

		s := db.NewSession()
		defer s.Close()
		read := &Task{ID: 51}
		require.NoError(t, read.ReadOne(s, u1))
	})
	t.Run("done task restores into the done bucket", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		setTaskColumnsForTest(t, 51, map[string]any{"done": true})

		_, err := restoreTaskForTest(t, 51, u1)
		require.NoError(t, err)

		db.AssertExists(t, "task_buckets", map[string]interface{}{"task_id": 51, "project_view_id": 4, "bucket_id": 3}, false)
	})
	t.Run("round trip through delete keeps the deleter cleared", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		trashTaskForTest(t, 1, u1)

		_, err := restoreTaskForTest(t, 1, u1)
		require.NoError(t, err)

		db.AssertExists(t, "tasks", map[string]interface{}{"id": 1, "deleted_at": nil, "deleted_by_id": nil}, false)
		db.AssertExists(t, "task_buckets", map[string]interface{}{"task_id": 1, "project_view_id": 4}, false)
	})
	t.Run("write link share can restore in its project", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		share := &LinkSharing{ID: 2, ProjectID: 2, Permission: PermissionWrite}
		trashTaskForTest(t, 13, share)

		_, err := restoreTaskForTest(t, 13, share)
		require.NoError(t, err)
		db.AssertExists(t, "tasks", map[string]interface{}{"id": 13, "deleted_at": nil}, false)
	})
	t.Run("read-only link share cannot restore", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		_, err := restoreTaskForTest(t, 51, &LinkSharing{ID: 1, ProjectID: 1, Permission: PermissionRead})
		require.ErrorIs(t, err, ErrGenericForbidden{})
	})
	t.Run("unrelated user cannot restore", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		_, err := restoreTaskForTest(t, 51, &user.User{ID: 2})
		require.ErrorIs(t, err, ErrGenericForbidden{})
		db.AssertExists(t, "tasks", map[string]interface{}{"id": 51, "deleted_by_id": nil}, false)
	})
	t.Run("refused in an archived project", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		// Task 36 lives in archived project 22, which a delete would refuse
		setTaskColumnsForTest(t, 36, map[string]any{"deleted_at": time.Now()})

		_, err := restoreTaskForTest(t, 36, u1)
		require.Error(t, err)
		assert.True(t, IsErrProjectIsArchived(err))
	})
	t.Run("live task is not in the trash", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		_, err := restoreTaskForTest(t, 1, u1)
		require.Error(t, err)
		assert.True(t, IsErrTaskDoesNotExist(err))
	})
}
