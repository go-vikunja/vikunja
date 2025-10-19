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

// ===== Permission Tests =====
// T-PERM-007: Task permission method tests

func TestTaskService_CanRead(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()
	ts := NewTaskService(db.GetEngine())

	t.Run("Owner_CanRead", func(t *testing.T) {
		// User 1 owns project 1, which contains task 1
		u := &user.User{ID: 1}
		canRead, maxRight, err := ts.CanRead(s, 1, u)

		require.NoError(t, err)
		assert.True(t, canRead)
		assert.Equal(t, int(models.PermissionAdmin), maxRight)
	})

	t.Run("ReadUser_CanRead", func(t *testing.T) {
		// User 1 has read permission on project 3, which contains tasks
		u := &user.User{ID: 1}
		canRead, maxRight, err := ts.CanRead(s, 15, u) // Task 15 is in project 3

		require.NoError(t, err)
		assert.True(t, canRead)
		assert.GreaterOrEqual(t, maxRight, int(models.PermissionRead))
	})

	t.Run("NoPermission_CannotRead", func(t *testing.T) {
		// User 13 has no permission on project 1 or its tasks
		u := &user.User{ID: 13}
		canRead, maxRight, err := ts.CanRead(s, 1, u)

		require.NoError(t, err)
		assert.False(t, canRead)
		assert.Equal(t, 0, maxRight)
	})

	t.Run("NonexistentTask_Error", func(t *testing.T) {
		u := &user.User{ID: 1}
		canRead, maxRight, err := ts.CanRead(s, 99999, u)

		assert.Error(t, err)
		assert.False(t, canRead)
		assert.Equal(t, 0, maxRight)
		assert.True(t, models.IsErrTaskDoesNotExist(err))
	})
}

func TestTaskService_CanWrite(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()
	ts := NewTaskService(db.GetEngine())

	t.Run("Owner_CanWrite", func(t *testing.T) {
		// User 1 owns project 1, which contains task 1
		u := &user.User{ID: 1}
		canWrite, err := ts.CanWrite(s, 1, u)

		require.NoError(t, err)
		assert.True(t, canWrite)
	})

	t.Run("WriteUser_CanWrite", func(t *testing.T) {
		// User 1 has write permission on project 10
		u := &user.User{ID: 1}
		canWrite, err := ts.CanWrite(s, 25, u) // Task 25 is in project 10

		require.NoError(t, err)
		assert.True(t, canWrite)
	})

	t.Run("ReadUser_CannotWrite", func(t *testing.T) {
		// User 1 has only read permission on project 9
		u := &user.User{ID: 1}
		canWrite, err := ts.CanWrite(s, 24, u) // Task 24 is in project 9

		require.NoError(t, err)
		assert.False(t, canWrite)
	})

	t.Run("NoPermission_CannotWrite", func(t *testing.T) {
		// User 13 has no permission on project 1
		u := &user.User{ID: 13}
		canWrite, err := ts.CanWrite(s, 1, u)

		require.NoError(t, err)
		assert.False(t, canWrite)
	})
}

func TestTaskService_CanUpdate(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()
	ts := NewTaskService(db.GetEngine())

	t.Run("Owner_CanUpdate", func(t *testing.T) {
		// User 1 owns project 1, which contains task 1
		u := &user.User{ID: 1}
		task := &models.Task{ID: 1, Title: "Updated"}
		canUpdate, err := ts.CanUpdate(s, 1, task, u)

		require.NoError(t, err)
		assert.True(t, canUpdate)
	})

	t.Run("WriteUser_CanUpdate", func(t *testing.T) {
		// User 1 has write permission on project 10
		u := &user.User{ID: 1}
		task := &models.Task{ID: 25, Title: "Updated"}
		canUpdate, err := ts.CanUpdate(s, 25, task, u)

		require.NoError(t, err)
		assert.True(t, canUpdate)
	})

	t.Run("ReadUser_CannotUpdate", func(t *testing.T) {
		// User 1 has only read permission on project 9
		u := &user.User{ID: 1}
		task := &models.Task{ID: 24, Title: "Updated"}
		canUpdate, err := ts.CanUpdate(s, 24, task, u)

		require.NoError(t, err)
		assert.False(t, canUpdate)
	})

	t.Run("MoveTask_BothProjectsPermissionRequired", func(t *testing.T) {
		// User 1 owns project 1, has write on project 10
		// Should be able to move task 1 (project 1) to project 10
		u := &user.User{ID: 1}
		task := &models.Task{ID: 1, ProjectID: 10} // Moving to project 10
		canUpdate, err := ts.CanUpdate(s, 1, task, u)

		require.NoError(t, err)
		assert.True(t, canUpdate)
	})

	t.Run("MoveTask_NoPermissionOnNewProject_Fails", func(t *testing.T) {
		// User 6 has read on project 1, but no permission on project 10
		// Should not be able to move task 1 to project 10
		u := &user.User{ID: 6}
		task := &models.Task{ID: 1, ProjectID: 10} // Try moving to project 10
		canUpdate, err := ts.CanUpdate(s, 1, task, u)

		// Should either return false or an error
		if err != nil {
			assert.True(t, models.IsErrGenericForbidden(err))
		} else {
			assert.False(t, canUpdate)
		}
	})
}

func TestTaskService_CanDelete(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()
	ts := NewTaskService(db.GetEngine())

	t.Run("Owner_CanDelete", func(t *testing.T) {
		// User 1 owns project 1, which contains task 1
		u := &user.User{ID: 1}
		canDelete, err := ts.CanDelete(s, 1, u)

		require.NoError(t, err)
		assert.True(t, canDelete)
	})

	t.Run("WriteUser_CanDelete", func(t *testing.T) {
		// User 1 has write permission on project 10
		u := &user.User{ID: 1}
		canDelete, err := ts.CanDelete(s, 25, u)

		require.NoError(t, err)
		assert.True(t, canDelete)
	})

	t.Run("ReadUser_CannotDelete", func(t *testing.T) {
		// User 1 has only read permission on project 9
		u := &user.User{ID: 1}
		canDelete, err := ts.CanDelete(s, 24, u)

		require.NoError(t, err)
		assert.False(t, canDelete)
	})

	t.Run("NoPermission_CannotDelete", func(t *testing.T) {
		// User 13 has no permission on project 1
		u := &user.User{ID: 13}
		canDelete, err := ts.CanDelete(s, 1, u)

		require.NoError(t, err)
		assert.False(t, canDelete)
	})
}

func TestTaskService_CanCreate(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()
	ts := NewTaskService(db.GetEngine())

	t.Run("Owner_CanCreate", func(t *testing.T) {
		// User 1 owns project 1
		u := &user.User{ID: 1}
		task := &models.Task{ProjectID: 1, Title: "New Task"}
		canCreate, err := ts.CanCreate(s, task, u)

		require.NoError(t, err)
		assert.True(t, canCreate)
	})

	t.Run("WriteUser_CanCreate", func(t *testing.T) {
		// User 1 has write permission on project 10
		u := &user.User{ID: 1}
		task := &models.Task{ProjectID: 10, Title: "New Task"}
		canCreate, err := ts.CanCreate(s, task, u)

		require.NoError(t, err)
		assert.True(t, canCreate)
	})

	t.Run("ReadUser_CannotCreate", func(t *testing.T) {
		// User 1 has only read permission on project 9
		u := &user.User{ID: 1}
		task := &models.Task{ProjectID: 9, Title: "New Task"}
		canCreate, err := ts.CanCreate(s, task, u)

		require.NoError(t, err)
		assert.False(t, canCreate)
	})

	t.Run("NoPermission_CannotCreate", func(t *testing.T) {
		// User 13 has no permission on project 1
		u := &user.User{ID: 13}
		task := &models.Task{ProjectID: 1, Title: "New Task"}
		canCreate, err := ts.CanCreate(s, task, u)

		require.NoError(t, err)
		assert.False(t, canCreate)
	})
}

// ===== Task Relation Permission Tests (T-PERM-010) =====

func TestTaskService_CanCreateAssignee(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()
	ts := NewTaskService(db.GetEngine())

	t.Run("WriteUser_CanCreateAssignee", func(t *testing.T) {
		// User 1 can write to task 1 (owns project 1)
		u := &user.User{ID: 1}
		can, err := ts.CanCreateAssignee(s, 1, u)

		require.NoError(t, err)
		assert.True(t, can)
	})

	t.Run("ReadUser_CannotCreateAssignee", func(t *testing.T) {
		// User 1 has only read permission on project 6
		u := &user.User{ID: 1}
		can, err := ts.CanCreateAssignee(s, 15, u) // Task 15 is in project 6

		require.NoError(t, err)
		assert.False(t, can)
	})

	t.Run("NoPermission_CannotCreateAssignee", func(t *testing.T) {
		// User 13 has no permission on project 1
		u := &user.User{ID: 13}
		can, err := ts.CanCreateAssignee(s, 1, u)

		require.NoError(t, err)
		assert.False(t, can)
	})
}

func TestTaskService_CanDeleteAssignee(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()
	ts := NewTaskService(db.GetEngine())

	t.Run("WriteUser_CanDeleteAssignee", func(t *testing.T) {
		// User 1 can write to task 1
		u := &user.User{ID: 1}
		can, err := ts.CanDeleteAssignee(s, 1, u)

		require.NoError(t, err)
		assert.True(t, can)
	})

	t.Run("ReadUser_CannotDeleteAssignee", func(t *testing.T) {
		// User 1 has only read permission on project 6
		u := &user.User{ID: 1}
		can, err := ts.CanDeleteAssignee(s, 15, u)

		require.NoError(t, err)
		assert.False(t, can)
	})
}

func TestTaskService_CanCreateRelation(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()
	ts := NewTaskService(db.GetEngine())

	t.Run("ValidRelation_CanCreate", func(t *testing.T) {
		// User 1 has write access to task 1 and read access to task 2
		u := &user.User{ID: 1}
		can, err := ts.CanCreateRelation(s, 1, 2, models.RelationKindSubtask, u)

		require.NoError(t, err)
		assert.True(t, can)
	})

	t.Run("InvalidRelationKind_CannotCreate", func(t *testing.T) {
		// Invalid relation kind
		u := &user.User{ID: 1}
		can, err := ts.CanCreateRelation(s, 1, 2, models.RelationKind("invalid"), u)

		assert.Error(t, err)
		assert.False(t, can)
		assert.True(t, models.IsErrInvalidRelationKind(err))
	})

	t.Run("NoWriteAccessToBaseTask_CannotCreate", func(t *testing.T) {
		// User 1 has only read access to task 15 (project 6)
		u := &user.User{ID: 1}
		can, err := ts.CanCreateRelation(s, 15, 1, models.RelationKindSubtask, u)

		require.NoError(t, err)
		assert.False(t, can)
	})

	t.Run("NoReadAccessToOtherTask_CannotCreate", func(t *testing.T) {
		// User 1 has write access to task 1 but no access to task 34 (owned by user 13)
		u := &user.User{ID: 1}
		can, err := ts.CanCreateRelation(s, 1, 34, models.RelationKindSubtask, u) // Task 34 belongs to user 13

		require.NoError(t, err)
		assert.False(t, can)
	})
}

func TestTaskService_CanDeleteRelation(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()
	ts := NewTaskService(db.GetEngine())

	t.Run("WriteUser_CanDeleteRelation", func(t *testing.T) {
		// User 1 can write to task 1
		u := &user.User{ID: 1}
		can, err := ts.CanDeleteRelation(s, 1, u)

		require.NoError(t, err)
		assert.True(t, can)
	})

	t.Run("ReadUser_CannotDeleteRelation", func(t *testing.T) {
		// User 1 has only read permission on project 6
		u := &user.User{ID: 1}
		can, err := ts.CanDeleteRelation(s, 15, u)

		require.NoError(t, err)
		assert.False(t, can)
	})
}

func TestTaskService_CanUpdatePosition(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()
	ts := NewTaskService(db.GetEngine())

	t.Run("WriteUser_CanUpdatePosition", func(t *testing.T) {
		// User 1 can write to task 1
		u := &user.User{ID: 1}
		can, err := ts.CanUpdatePosition(s, 1, u)

		require.NoError(t, err)
		assert.True(t, can)
	})

	t.Run("ReadUser_CannotUpdatePosition", func(t *testing.T) {
		// User 1 has only read permission on project 6
		u := &user.User{ID: 1}
		can, err := ts.CanUpdatePosition(s, 15, u)

		require.NoError(t, err)
		assert.False(t, can)
	})

	t.Run("NoPermission_CannotUpdatePosition", func(t *testing.T) {
		// User 13 has no permission on project 1
		u := &user.User{ID: 13}
		can, err := ts.CanUpdatePosition(s, 1, u)

		require.NoError(t, err)
		assert.False(t, can)
	})
}
