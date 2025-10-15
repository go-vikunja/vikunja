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

// TestPermissionBaseline_Project tests current Project permission behavior
// This baseline test captures the EXACT current behavior before migration to service layer
func TestPermissionBaseline_Project(t *testing.T) {
	t.Run("CanRead", func(t *testing.T) {
		t.Run("Owner_CanRead", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			u := &user.User{ID: 1}
			project := &models.Project{ID: 1}

			canRead, maxRight, err := project.CanRead(s, u)
			require.NoError(t, err)
			assert.True(t, canRead)
			assert.Equal(t, int(models.PermissionAdmin), maxRight)
		})

		t.Run("UserWithReadPermission_CanRead", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			// User 2 has read permission (permission=0) on project 3
			u := &user.User{ID: 2}
			project := &models.Project{ID: 3}

			canRead, maxRight, err := project.CanRead(s, u)
			require.NoError(t, err)
			assert.True(t, canRead)
			assert.Equal(t, int(models.PermissionRead), maxRight)
		})

		t.Run("UserWithWritePermission_CanRead", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			// User 1 has write permission (permission=1) on project 10
			u := &user.User{ID: 1}
			project := &models.Project{ID: 10}

			canRead, maxRight, err := project.CanRead(s, u)
			require.NoError(t, err)
			assert.True(t, canRead)
			assert.Equal(t, int(models.PermissionWrite), maxRight)
		})

		t.Run("UserWithAdminPermission_CanRead", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			// User 1 has admin permission (permission=2) on project 3
			u := &user.User{ID: 1}
			project := &models.Project{ID: 3}

			canRead, maxRight, err := project.CanRead(s, u)
			require.NoError(t, err)
			assert.True(t, canRead)
			assert.Equal(t, int(models.PermissionAdmin), maxRight)
		})

		t.Run("UserWithoutPermission_CannotRead", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			// User 2 has no permission on project 1
			u := &user.User{ID: 2}
			project := &models.Project{ID: 1}

			canRead, maxRight, err := project.CanRead(s, u)
			require.NoError(t, err)
			assert.False(t, canRead)
			assert.Equal(t, 0, maxRight)
		})

		t.Run("LinkShareWithReadPermission_CanRead", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			// Link share ID 1 has read permission (permission=0) on project 1
			linkShare := &models.LinkSharing{ID: 1, ProjectID: 1, Permission: models.PermissionRead}
			project := &models.Project{ID: 1}

			canRead, maxRight, err := project.CanRead(s, linkShare)
			require.NoError(t, err)
			assert.True(t, canRead)
			assert.Equal(t, int(models.PermissionRead), maxRight)
		})

		t.Run("LinkShareWithWritePermission_CanRead", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			// Link share ID 2 has write permission (permission=1) on project 2
			linkShare := &models.LinkSharing{ID: 2, ProjectID: 2, Permission: models.PermissionWrite}
			project := &models.Project{ID: 2}

			canRead, maxRight, err := project.CanRead(s, linkShare)
			require.NoError(t, err)
			assert.True(t, canRead)
			assert.Equal(t, int(models.PermissionWrite), maxRight)
		})

		t.Run("LinkShareWithAdminPermission_CanRead", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			// Link share ID 3 has admin permission (permission=2) on project 3
			linkShare := &models.LinkSharing{ID: 3, ProjectID: 3, Permission: models.PermissionAdmin}
			project := &models.Project{ID: 3}

			canRead, maxRight, err := project.CanRead(s, linkShare)
			require.NoError(t, err)
			assert.True(t, canRead)
			assert.Equal(t, int(models.PermissionAdmin), maxRight)
		})

		t.Run("NonexistentProject_ReturnsError", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			u := &user.User{ID: 1}
			project := &models.Project{ID: 99999}

			canRead, maxRight, err := project.CanRead(s, u)
			assert.Error(t, err)
			assert.False(t, canRead)
			assert.Equal(t, 0, maxRight)
		})
	})

	t.Run("CanWrite", func(t *testing.T) {
		t.Run("Owner_CanWrite", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			u := &user.User{ID: 1}
			project := &models.Project{ID: 1}

			canWrite, err := project.CanWrite(s, u)
			require.NoError(t, err)
			assert.True(t, canWrite)
		})

		t.Run("UserWithWritePermission_CanWrite", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			// User 1 has write permission on project 10
			u := &user.User{ID: 1}
			project := &models.Project{ID: 10}

			canWrite, err := project.CanWrite(s, u)
			require.NoError(t, err)
			assert.True(t, canWrite)
		})

		t.Run("UserWithAdminPermission_CanWrite", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			// User 1 has admin permission on project 3
			u := &user.User{ID: 1}
			project := &models.Project{ID: 3}

			canWrite, err := project.CanWrite(s, u)
			require.NoError(t, err)
			assert.True(t, canWrite)
		})

		t.Run("UserWithReadPermission_CannotWrite", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			// User 2 has read permission on project 3
			u := &user.User{ID: 2}
			project := &models.Project{ID: 3}

			canWrite, err := project.CanWrite(s, u)
			require.NoError(t, err)
			assert.False(t, canWrite)
		})

		t.Run("UserWithoutPermission_CannotWrite", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			u := &user.User{ID: 2}
			project := &models.Project{ID: 1}

			canWrite, err := project.CanWrite(s, u)
			require.NoError(t, err)
			assert.False(t, canWrite)
		})

		t.Run("LinkShareWithWritePermission_CanWrite", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			// Link share with write permission
			linkShare := &models.LinkSharing{ID: 2, ProjectID: 2, Permission: models.PermissionWrite}
			project := &models.Project{ID: 2}

			canWrite, err := project.CanWrite(s, linkShare)
			require.NoError(t, err)
			assert.True(t, canWrite)
		})

		t.Run("LinkShareWithReadPermission_CannotWrite", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			// Link share with read permission
			linkShare := &models.LinkSharing{ID: 1, ProjectID: 1, Permission: models.PermissionRead}
			project := &models.Project{ID: 1}

			canWrite, err := project.CanWrite(s, linkShare)
			require.NoError(t, err)
			assert.False(t, canWrite)
		})
	})

	t.Run("CanUpdate", func(t *testing.T) {
		t.Run("Owner_CanUpdate", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			u := &user.User{ID: 1}
			project := &models.Project{ID: 1}

			canUpdate, err := project.CanUpdate(s, u)
			require.NoError(t, err)
			assert.True(t, canUpdate)
		})

		t.Run("UserWithWritePermission_CanUpdate", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			// User 1 has write permission on project 10
			u := &user.User{ID: 1}
			project := &models.Project{ID: 10}

			canUpdate, err := project.CanUpdate(s, u)
			require.NoError(t, err)
			assert.True(t, canUpdate)
		})

		t.Run("UserWithReadPermission_CannotUpdate", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			// User 2 has read permission on project 3
			u := &user.User{ID: 2}
			project := &models.Project{ID: 3}

			canUpdate, err := project.CanUpdate(s, u)
			require.NoError(t, err)
			assert.False(t, canUpdate)
		})

		t.Run("UserWithoutPermission_CannotUpdate", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			u := &user.User{ID: 2}
			project := &models.Project{ID: 1}

			canUpdate, err := project.CanUpdate(s, u)
			require.NoError(t, err)
			assert.False(t, canUpdate)
		})
	})

	t.Run("CanDelete", func(t *testing.T) {
		t.Run("Owner_CanDelete", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			u := &user.User{ID: 1}
			project := &models.Project{ID: 1}

			canDelete, err := project.CanDelete(s, u)
			require.NoError(t, err)
			assert.True(t, canDelete)
		})

		t.Run("UserWithAdminPermission_CanDelete", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			// User 1 has admin permission on project 3
			u := &user.User{ID: 1}
			project := &models.Project{ID: 3}

			canDelete, err := project.CanDelete(s, u)
			require.NoError(t, err)
			assert.True(t, canDelete)
		})

		t.Run("UserWithWritePermission_CannotDelete", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			// User 1 has write permission on project 10
			u := &user.User{ID: 1}
			project := &models.Project{ID: 10}

			canDelete, err := project.CanDelete(s, u)
			require.NoError(t, err)
			assert.False(t, canDelete)
		})

		t.Run("UserWithReadPermission_CannotDelete", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			// User 2 has read permission on project 3
			u := &user.User{ID: 2}
			project := &models.Project{ID: 3}

			canDelete, err := project.CanDelete(s, u)
			require.NoError(t, err)
			assert.False(t, canDelete)
		})

		t.Run("UserWithoutPermission_CannotDelete", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			u := &user.User{ID: 2}
			project := &models.Project{ID: 1}

			canDelete, err := project.CanDelete(s, u)
			require.NoError(t, err)
			assert.False(t, canDelete)
		})
	})

	t.Run("CanCreate", func(t *testing.T) {
		t.Run("RegularUser_CanCreateTopLevel", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			u := &user.User{ID: 1}
			project := &models.Project{ParentProjectID: 0}

			canCreate, err := project.CanCreate(s, u)
			require.NoError(t, err)
			assert.True(t, canCreate)
		})

		t.Run("RegularUser_CanCreateInOwnedProject", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			u := &user.User{ID: 1}
			project := &models.Project{ParentProjectID: 1} // Parent owned by user 1

			canCreate, err := project.CanCreate(s, u)
			require.NoError(t, err)
			assert.True(t, canCreate)
		})

		t.Run("RegularUser_CanCreateInProjectWithWritePermission", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			// User 1 has write permission on project 10
			u := &user.User{ID: 1}
			project := &models.Project{ParentProjectID: 10}

			canCreate, err := project.CanCreate(s, u)
			require.NoError(t, err)
			assert.True(t, canCreate)
		})

		t.Run("RegularUser_CannotCreateInProjectWithoutPermission", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			// User 2 has no permission on project 1
			u := &user.User{ID: 2}
			project := &models.Project{ParentProjectID: 1}

			canCreate, err := project.CanCreate(s, u)
			require.NoError(t, err)
			assert.False(t, canCreate)
		})

		t.Run("LinkShare_CannotCreate", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			linkShare := &models.LinkSharing{ID: 1}
			project := &models.Project{ParentProjectID: 0}

			canCreate, err := project.CanCreate(s, linkShare)
			require.NoError(t, err)
			assert.False(t, canCreate)
		})
	})
}

// TestPermissionBaseline_Task tests current Task permission behavior
func TestPermissionBaseline_Task(t *testing.T) {
	t.Run("CanRead", func(t *testing.T) {
		t.Run("ProjectOwner_CanReadTask", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			// User 1 owns project 1, which contains task 1
			u := &user.User{ID: 1}
			task := &models.Task{ID: 1}

			canRead, maxRight, err := task.CanRead(s, u)
			require.NoError(t, err)
			assert.True(t, canRead)
			assert.Equal(t, int(models.PermissionAdmin), maxRight)
		})

		t.Run("UserWithProjectReadPermission_CanReadTask", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			// User 2 has read permission on project 3
			// Task 32 is in project 3
			u := &user.User{ID: 2}
			task := &models.Task{ID: 32}

			canRead, maxRight, err := task.CanRead(s, u)
			require.NoError(t, err)
			assert.True(t, canRead)
			assert.Equal(t, int(models.PermissionRead), maxRight)
		})

		t.Run("UserWithoutPermission_CannotReadTask", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			// User 2 has no permission on project 1
			u := &user.User{ID: 2}
			task := &models.Task{ID: 1}

			canRead, maxRight, err := task.CanRead(s, u)
			require.NoError(t, err)
			assert.False(t, canRead)
			assert.Equal(t, 0, maxRight)
		})

		t.Run("NonexistentTask_ReturnsError", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			u := &user.User{ID: 1}
			task := &models.Task{ID: 99999}

			canRead, maxRight, err := task.CanRead(s, u)
			assert.Error(t, err)
			assert.False(t, canRead)
			assert.Equal(t, 0, maxRight)
		})
	})

	t.Run("CanWrite", func(t *testing.T) {
		t.Run("ProjectOwner_CanWriteTask", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			u := &user.User{ID: 1}
			task := &models.Task{ID: 1}

			canWrite, err := task.CanWrite(s, u)
			require.NoError(t, err)
			assert.True(t, canWrite)
		})

		t.Run("UserWithProjectWritePermission_CanWriteTask", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			// User 1 has write permission on project 10
			// Task 19 is in project 10
			u := &user.User{ID: 1}
			task := &models.Task{ID: 19}

			canWrite, err := task.CanWrite(s, u)
			require.NoError(t, err)
			assert.True(t, canWrite)
		})

		t.Run("UserWithProjectReadPermission_CannotWriteTask", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			// User 2 has read permission on project 3
			// Task 32 is in project 3
			u := &user.User{ID: 2}
			task := &models.Task{ID: 32}

			canWrite, err := task.CanWrite(s, u)
			require.NoError(t, err)
			assert.False(t, canWrite)
		})

		t.Run("UserWithoutPermission_CannotWriteTask", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			u := &user.User{ID: 2}
			task := &models.Task{ID: 1}

			canWrite, err := task.CanWrite(s, u)
			require.NoError(t, err)
			assert.False(t, canWrite)
		})
	})

	t.Run("CanUpdate", func(t *testing.T) {
		t.Run("ProjectOwner_CanUpdateTask", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			u := &user.User{ID: 1}
			task := &models.Task{ID: 1}

			canUpdate, err := task.CanUpdate(s, u)
			require.NoError(t, err)
			assert.True(t, canUpdate)
		})

		t.Run("UserWithProjectReadPermission_CannotUpdateTask", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			// User 2 has read permission on project 3
			// Task 32 is in project 3
			u := &user.User{ID: 2}
			task := &models.Task{ID: 32}

			canUpdate, err := task.CanUpdate(s, u)
			require.NoError(t, err)
			assert.False(t, canUpdate)
		})

		t.Run("UserWithoutPermission_CannotUpdateTask", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			u := &user.User{ID: 2}
			task := &models.Task{ID: 1}

			canUpdate, err := task.CanUpdate(s, u)
			require.NoError(t, err)
			assert.False(t, canUpdate)
		})
	})

	t.Run("CanDelete", func(t *testing.T) {
		t.Run("ProjectOwner_CanDeleteTask", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			u := &user.User{ID: 1}
			task := &models.Task{ID: 1}

			canDelete, err := task.CanDelete(s, u)
			require.NoError(t, err)
			assert.True(t, canDelete)
		})

		t.Run("UserWithProjectWritePermission_CanDeleteTask", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			// User 1 has write permission on project 10
			// Task 19 is in project 10
			u := &user.User{ID: 1}
			task := &models.Task{ID: 19}

			canDelete, err := task.CanDelete(s, u)
			require.NoError(t, err)
			assert.True(t, canDelete)
		})

		t.Run("UserWithProjectReadPermission_CannotDeleteTask", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			// User 2 has read permission on project 3
			// Task 32 is in project 3
			u := &user.User{ID: 2}
			task := &models.Task{ID: 32}

			canDelete, err := task.CanDelete(s, u)
			require.NoError(t, err)
			assert.False(t, canDelete)
		})

		t.Run("UserWithoutPermission_CannotDeleteTask", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			u := &user.User{ID: 2}
			task := &models.Task{ID: 1}

			canDelete, err := task.CanDelete(s, u)
			require.NoError(t, err)
			assert.False(t, canDelete)
		})
	})

	t.Run("CanCreate", func(t *testing.T) {
		t.Run("UserWithWritePermissionOnProject_CanCreateTask", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			// User 1 has write permission on project 10
			u := &user.User{ID: 1}
			task := &models.Task{ProjectID: 10}

			canCreate, err := task.CanCreate(s, u)
			require.NoError(t, err)
			assert.True(t, canCreate)
		})

		t.Run("ProjectOwner_CanCreateTask", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			u := &user.User{ID: 1}
			task := &models.Task{ProjectID: 1}

			canCreate, err := task.CanCreate(s, u)
			require.NoError(t, err)
			assert.True(t, canCreate)
		})

		t.Run("UserWithReadPermissionOnProject_CannotCreateTask", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			// User 2 has read permission on project 3
			u := &user.User{ID: 2}
			task := &models.Task{ProjectID: 3}

			canCreate, err := task.CanCreate(s, u)
			require.NoError(t, err)
			assert.False(t, canCreate)
		})

		t.Run("UserWithoutPermissionOnProject_CannotCreateTask", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			u := &user.User{ID: 2}
			task := &models.Task{ProjectID: 1}

			canCreate, err := task.CanCreate(s, u)
			require.NoError(t, err)
			assert.False(t, canCreate)
		})
	})
}

// TestPermissionBaseline_LinkSharing tests current LinkSharing permission behavior
func TestPermissionBaseline_LinkSharing(t *testing.T) {
	t.Run("CanRead", func(t *testing.T) {
		t.Run("ProjectOwner_CanReadLinkShare", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			// User 1 owns project 1, link share 1 has hash "test" for project 1
			u := &user.User{ID: 1}
			linkShare := &models.LinkSharing{ID: 1, Hash: "test", ProjectID: 1}

			canRead, _, err := linkShare.CanRead(s, u)
			require.NoError(t, err)
			assert.True(t, canRead)
		})

		t.Run("UserWithoutProjectPermission_CannotReadLinkShare", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			// User 2 has no permission on project 1
			u := &user.User{ID: 2}
			linkShare := &models.LinkSharing{ID: 1, Hash: "test", ProjectID: 1}

			canRead, _, err := linkShare.CanRead(s, u)
			require.NoError(t, err)
			assert.False(t, canRead)
		})
	})

	t.Run("CanUpdate", func(t *testing.T) {
		t.Run("ProjectOwner_CanUpdateLinkShare", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			// User 1 owns project 1, link share 1 is for project 1
			u := &user.User{ID: 1}
			linkShare := &models.LinkSharing{ID: 1, ProjectID: 1}

			canUpdate, err := linkShare.CanUpdate(s, u)
			require.NoError(t, err)
			assert.True(t, canUpdate)
		})

		t.Run("UserWithoutProjectPermission_CannotUpdateLinkShare", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			// User 2 has no permission on project 1
			u := &user.User{ID: 2}
			linkShare := &models.LinkSharing{ID: 1, ProjectID: 1}

			canUpdate, err := linkShare.CanUpdate(s, u)
			require.NoError(t, err)
			assert.False(t, canUpdate)
		})
	})

	t.Run("CanDelete", func(t *testing.T) {
		t.Run("ProjectOwner_CanDeleteLinkShare", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			// User 1 owns project 1, link share 1 is for project 1
			u := &user.User{ID: 1}
			linkShare := &models.LinkSharing{ID: 1, ProjectID: 1}

			canDelete, err := linkShare.CanDelete(s, u)
			require.NoError(t, err)
			assert.True(t, canDelete)
		})

		t.Run("UserWithoutProjectPermission_CannotDeleteLinkShare", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			// User 2 has no permission on project 1
			u := &user.User{ID: 2}
			linkShare := &models.LinkSharing{ID: 1, ProjectID: 1}

			canDelete, err := linkShare.CanDelete(s, u)
			require.NoError(t, err)
			assert.False(t, canDelete)
		})
	})

	t.Run("CanCreate", func(t *testing.T) {
		t.Run("UserWithWritePermissionOnProject_CanCreateLinkShare", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			// User 1 has write permission on project 10
			u := &user.User{ID: 1}
			linkShare := &models.LinkSharing{ProjectID: 10}

			canCreate, err := linkShare.CanCreate(s, u)
			require.NoError(t, err)
			assert.True(t, canCreate)
		})

		t.Run("ProjectOwner_CanCreateLinkShare", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			// User 1 owns project 1
			u := &user.User{ID: 1}
			linkShare := &models.LinkSharing{ProjectID: 1}

			canCreate, err := linkShare.CanCreate(s, u)
			require.NoError(t, err)
			assert.True(t, canCreate)
		})

		t.Run("UserWithReadPermissionOnProject_CannotCreateLinkShare", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			// User 2 has read permission on project 3
			u := &user.User{ID: 2}
			linkShare := &models.LinkSharing{ProjectID: 3}

			canCreate, err := linkShare.CanCreate(s, u)
			require.NoError(t, err)
			assert.False(t, canCreate)
		})

		t.Run("UserWithoutPermissionOnProject_CannotCreateLinkShare", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			// User 2 has no permission on project 1
			u := &user.User{ID: 2}
			linkShare := &models.LinkSharing{ProjectID: 1}

			canCreate, err := linkShare.CanCreate(s, u)
			require.NoError(t, err)
			assert.False(t, canCreate)
		})
	})
}

// TestPermissionBaseline_Label tests current Label permission behavior
func TestPermissionBaseline_Label(t *testing.T) {
	t.Run("CanRead", func(t *testing.T) {
		t.Run("LabelCreator_CanReadLabel", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			// Label 1 created by user 1
			u := &user.User{ID: 1}
			label := &models.Label{ID: 1}

			canRead, _, err := label.CanRead(s, u)
			require.NoError(t, err)
			assert.True(t, canRead)
		})

		t.Run("OtherUser_CannotReadPrivateLabel", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			// Label 1 created by user 1
			u := &user.User{ID: 2}
			label := &models.Label{ID: 1}

			canRead, _, err := label.CanRead(s, u)
			require.NoError(t, err)
			assert.False(t, canRead)
		})
	})

	t.Run("CanUpdate", func(t *testing.T) {
		t.Run("LabelCreator_CanUpdateLabel", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			u := &user.User{ID: 1}
			label := &models.Label{ID: 1}

			canUpdate, err := label.CanUpdate(s, u)
			require.NoError(t, err)
			assert.True(t, canUpdate)
		})

		t.Run("OtherUser_CannotUpdateLabel", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			u := &user.User{ID: 2}
			label := &models.Label{ID: 1}

			canUpdate, err := label.CanUpdate(s, u)
			require.NoError(t, err)
			assert.False(t, canUpdate)
		})
	})

	t.Run("CanDelete", func(t *testing.T) {
		t.Run("LabelCreator_CanDeleteLabel", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			u := &user.User{ID: 1}
			label := &models.Label{ID: 1}

			canDelete, err := label.CanDelete(s, u)
			require.NoError(t, err)
			assert.True(t, canDelete)
		})

		t.Run("OtherUser_CannotDeleteLabel", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			u := &user.User{ID: 2}
			label := &models.Label{ID: 1}

			canDelete, err := label.CanDelete(s, u)
			require.NoError(t, err)
			assert.False(t, canDelete)
		})
	})

	t.Run("CanCreate", func(t *testing.T) {
		t.Run("AnyUser_CanCreateLabel", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			u := &user.User{ID: 1}
			label := &models.Label{}

			canCreate, err := label.CanCreate(s, u)
			require.NoError(t, err)
			assert.True(t, canCreate)
		})
	})
}

// TestPermissionBaseline_TaskComment tests current TaskComment permission behavior
func TestPermissionBaseline_TaskComment(t *testing.T) {
	t.Run("CanRead", func(t *testing.T) {
		t.Run("UserWithTaskReadPermission_CanReadComment", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			// User 1 owns project 1, task 1 is in project 1, comment 1 is on task 1
			u := &user.User{ID: 1}
			comment := &models.TaskComment{ID: 1, TaskID: 1}

			canRead, _, err := comment.CanRead(s, u)
			require.NoError(t, err)
			assert.True(t, canRead)
		})

		t.Run("UserWithoutTaskPermission_CannotReadComment", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			// User 2 has no permission on project 1, so no permission on task 1
			u := &user.User{ID: 2}
			comment := &models.TaskComment{ID: 1, TaskID: 1}

			canRead, _, err := comment.CanRead(s, u)
			require.NoError(t, err)
			assert.False(t, canRead)
		})
	})

	t.Run("CanUpdate", func(t *testing.T) {
		t.Run("CommentAuthor_CanUpdateComment", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			// Comment 1 is created by user 1 on task 1 (which user 1 owns)
			u := &user.User{ID: 1}
			comment := &models.TaskComment{ID: 1, TaskID: 1}

			canUpdate, err := comment.CanUpdate(s, u)
			require.NoError(t, err)
			assert.True(t, canUpdate)
		})

		t.Run("OtherUser_CannotUpdateComment", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			// Comment 1 is created by user 1, trying to update as user 3
			// User 3 owns project 2,3,4 but comment 1 is on task 1 in project 1
			u := &user.User{ID: 3}
			comment := &models.TaskComment{ID: 1, TaskID: 1}

			canUpdate, err := comment.CanUpdate(s, u)
			require.NoError(t, err)
			// User 3 doesn't have write access to task 1
			assert.False(t, canUpdate)
		})
	})

	t.Run("CanDelete", func(t *testing.T) {
		t.Run("CommentAuthor_CanDeleteComment", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			// Comment 1 is created by user 1 on task 1 (which user 1 owns)
			u := &user.User{ID: 1}
			comment := &models.TaskComment{ID: 1, TaskID: 1}

			canDelete, err := comment.CanDelete(s, u)
			require.NoError(t, err)
			assert.True(t, canDelete)
		})

		t.Run("OtherUser_CannotDeleteComment", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			// Comment 1 is created by user 1, trying to delete as user 3
			u := &user.User{ID: 3}
			comment := &models.TaskComment{ID: 1, TaskID: 1}

			canDelete, err := comment.CanDelete(s, u)
			require.NoError(t, err)
			// User 3 doesn't have write access to task 1
			assert.False(t, canDelete)
		})
	})

	t.Run("CanCreate", func(t *testing.T) {
		t.Run("UserWithTaskWritePermission_CanCreateComment", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			// User 1 owns project 1, task 1 is in project 1
			u := &user.User{ID: 1}
			comment := &models.TaskComment{TaskID: 1}

			canCreate, err := comment.CanCreate(s, u)
			require.NoError(t, err)
			assert.True(t, canCreate)
		})

		t.Run("UserWithoutTaskPermission_CannotCreateComment", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			u := &user.User{ID: 2}
			comment := &models.TaskComment{TaskID: 1}

			canCreate, err := comment.CanCreate(s, u)
			require.NoError(t, err)
			assert.False(t, canCreate)
		})
	})
}

// TestPermissionBaseline_Subscription tests current Subscription permission behavior
func TestPermissionBaseline_Subscription(t *testing.T) {
	t.Run("CanCreate", func(t *testing.T) {
		t.Run("UserWithTaskReadPermission_CanSubscribe", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			// User 1 owns project 1, task 1 is in project 1
			u := &user.User{ID: 1}
			subscription := &models.Subscription{EntityType: models.SubscriptionEntityTask, EntityID: 1}

			canCreate, err := subscription.CanCreate(s, u)
			require.NoError(t, err)
			assert.True(t, canCreate)
		})

		t.Run("UserWithoutTaskPermission_CannotSubscribe", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			u := &user.User{ID: 2}
			subscription := &models.Subscription{EntityType: models.SubscriptionEntityTask, EntityID: 1}

			canCreate, err := subscription.CanCreate(s, u)
			require.NoError(t, err)
			assert.False(t, canCreate)
		})
	})

	t.Run("CanDelete", func(t *testing.T) {
		t.Run("SubscriptionOwner_CanUnsubscribe", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			// Subscription 1: entity_type=3 (task), entity_id=2, user_id=1
			u := &user.User{ID: 1}
			subscription := &models.Subscription{ID: 1, EntityType: models.SubscriptionEntityTask, EntityID: 2}

			canDelete, err := subscription.CanDelete(s, u)
			require.NoError(t, err)
			assert.True(t, canDelete)
		})

		t.Run("OtherUser_CannotUnsubscribe", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			// Subscription 1 belongs to user 1, trying to delete as user 2
			u := &user.User{ID: 2}
			subscription := &models.Subscription{ID: 1, EntityType: models.SubscriptionEntityTask, EntityID: 2}

			canDelete, err := subscription.CanDelete(s, u)
			require.NoError(t, err)
			assert.False(t, canDelete)
		})
	})
}
