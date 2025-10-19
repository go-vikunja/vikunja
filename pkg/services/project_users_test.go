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

func TestProjectUserService_Create(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	service := NewProjectUserService(db.GetEngine())

	t.Run("create normally", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		pu := &models.ProjectUser{
			Username:  "user1",
			ProjectID: 2,
		}
		doer := &user.User{ID: 1}

		err := service.Create(s, pu, doer)
		assert.NoError(t, err)
		assert.NotEqual(t, int64(0), pu.ID, "ID should be set")

		// Verify it was created
		s.Commit()
		db.AssertExists(t, "users_projects", map[string]interface{}{
			"user_id":    1,
			"project_id": 2,
		}, false)
	})

	t.Run("create for duplicate", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		pu := &models.ProjectUser{
			Username:  "user1",
			ProjectID: 3, // user1 already has access to project 3
		}
		doer := &user.User{ID: 1}

		err := service.Create(s, pu, doer)
		assert.Error(t, err)
		assert.True(t, models.IsErrUserAlreadyHasAccess(err))
	})

	t.Run("create with invalid permission", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		pu := &models.ProjectUser{
			Username:   "user1",
			ProjectID:  2,
			Permission: 500, // Invalid permission
		}
		doer := &user.User{ID: 1}

		err := service.Create(s, pu, doer)
		assert.Error(t, err)
		assert.True(t, models.IsErrInvalidPermission(err))
	})

	t.Run("create with nonexistent project", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		pu := &models.ProjectUser{
			Username:  "user1",
			ProjectID: 9999,
		}
		doer := &user.User{ID: 1}

		err := service.Create(s, pu, doer)
		assert.Error(t, err)
		assert.True(t, models.IsErrProjectDoesNotExist(err))
	})

	t.Run("create with nonexistent user", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		pu := &models.ProjectUser{
			Username:  "nonexistent",
			ProjectID: 2,
		}
		doer := &user.User{ID: 1}

		err := service.Create(s, pu, doer)
		assert.Error(t, err)
		assert.True(t, user.IsErrUserDoesNotExist(err))
	})

	t.Run("create with owner as shared user", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		pu := &models.ProjectUser{
			Username:  "user3", // user3 owns project 4
			ProjectID: 4,
		}
		doer := &user.User{ID: 1}

		err := service.Create(s, pu, doer)
		assert.Error(t, err)
		assert.True(t, models.IsErrUserAlreadyHasAccess(err))
	})
}

func TestProjectUserService_Delete(t *testing.T) {
	db.LoadAndAssertFixtures(t)

	service := NewProjectUserService(db.GetEngine())

	t.Run("delete normally", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		pu := &models.ProjectUser{
			Username:  "user1",
			ProjectID: 3, // user1 has access to project 3
		}

		err := service.Delete(s, pu)
		assert.NoError(t, err)

		s.Commit()
		db.AssertMissing(t, "users_projects", map[string]interface{}{
			"user_id":    1,
			"project_id": 3,
		})
	})

	t.Run("delete nonexistent user", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		pu := &models.ProjectUser{
			Username:  "nonexistent",
			ProjectID: 3,
		}

		err := service.Delete(s, pu)
		assert.Error(t, err)
		assert.True(t, user.IsErrUserDoesNotExist(err))
	})

	t.Run("delete user without access", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		pu := &models.ProjectUser{
			Username:  "user1",
			ProjectID: 4, // user1 doesn't have access to project 4
		}

		err := service.Delete(s, pu)
		assert.Error(t, err)
		assert.True(t, models.IsErrUserDoesNotHaveAccessToProject(err))
	})
}

func TestProjectUserService_GetAll(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	service := NewProjectUserService(db.GetEngine())

	t.Run("get all users for project", func(t *testing.T) {
		users, count, total, err := service.GetAll(s, 3, &user.User{ID: 1}, "", 0, 50)
		assert.NoError(t, err)
		assert.Greater(t, count, 0, "Should have users")
		assert.Greater(t, total, int64(0), "Should have total count")
		assert.NotNil(t, users)

		// Check that emails are obfuscated
		for _, u := range users {
			assert.Empty(t, u.Email, "Email should be obfuscated")
		}
	})

	t.Run("pagination", func(t *testing.T) {
		users1, count1, total, err := service.GetAll(s, 3, &user.User{ID: 1}, "", 1, 1)
		assert.NoError(t, err)
		assert.Equal(t, 1, count1, "Should return 1 user per page")

		if total > 1 {
			users2, count2, _, err := service.GetAll(s, 3, &user.User{ID: 1}, "", 2, 1)
			assert.NoError(t, err)
			assert.Equal(t, 1, count2, "Should return 1 user per page")

			// Pages should have different users
			if len(users1) > 0 && len(users2) > 0 {
				assert.NotEqual(t, users1[0].ID, users2[0].ID, "Different pages should have different users")
			}
		}
	})

	t.Run("search by username", func(t *testing.T) {
		users, count, total, err := service.GetAll(s, 3, &user.User{ID: 1}, "user1", 0, 50)
		assert.NoError(t, err)

		// Should find user1 if they have access to project 3
		for _, u := range users {
			assert.Contains(t, u.Username, "user1", "Should match search term")
		}

		if count > 0 {
			assert.Greater(t, total, int64(0), "Should have total when results found")
		}
	})

	t.Run("no access to project", func(t *testing.T) {
		// User 2 doesn't have access to project 1 (owned by user 1)
		_, _, _, err := service.GetAll(s, 1, &user.User{ID: 2}, "", 0, 50)
		assert.Error(t, err)
		assert.True(t, models.IsErrNeedToHaveProjectReadAccess(err))
	})
}

func TestProjectUserService_Update(t *testing.T) {
	db.LoadAndAssertFixtures(t)

	service := NewProjectUserService(db.GetEngine())

	t.Run("update normally", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		pu := &models.ProjectUser{
			Username:   "user1",
			ProjectID:  3,
			Permission: models.PermissionAdmin, // Upgrade to admin
		}

		err := service.Update(s, pu)
		assert.NoError(t, err)

		s.Commit()
		db.AssertExists(t, "users_projects", map[string]interface{}{
			"user_id":    1,
			"project_id": 3,
			"permission": 2, // PermissionAdmin
		}, false)
	})

	t.Run("update with invalid permission", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		pu := &models.ProjectUser{
			Username:   "user1",
			ProjectID:  3,
			Permission: 999, // Invalid
		}

		err := service.Update(s, pu)
		assert.Error(t, err)
		assert.True(t, models.IsErrInvalidPermission(err))
	})

	t.Run("update nonexistent user", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		pu := &models.ProjectUser{
			Username:   "nonexistent",
			ProjectID:  3,
			Permission: models.PermissionAdmin,
		}

		err := service.Update(s, pu)
		assert.Error(t, err)
		assert.True(t, user.IsErrUserDoesNotExist(err))
	})
}

func TestProjectUserService_HasAccess(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	service := NewProjectUserService(db.GetEngine())

	t.Run("user has access", func(t *testing.T) {
		has, err := service.HasAccess(s, 3, 1) // user1 has access to project 3
		assert.NoError(t, err)
		assert.True(t, has, "User should have access")
	})

	t.Run("user doesn't have access", func(t *testing.T) {
		has, err := service.HasAccess(s, 1, 2) // user2 doesn't have access to project 1
		assert.NoError(t, err)
		assert.False(t, has, "User should not have access")
	})
}

func TestProjectUserService_GetPermission(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	service := NewProjectUserService(db.GetEngine())

	t.Run("get existing permission", func(t *testing.T) {
		perm, err := service.GetPermission(s, 3, 1) // user1 has access to project 3
		assert.NoError(t, err)
		assert.NotEqual(t, models.PermissionUnknown, perm, "Should have a valid permission")
	})

	t.Run("get permission for user without access", func(t *testing.T) {
		perm, err := service.GetPermission(s, 1, 2) // user2 doesn't have access to project 1
		assert.NoError(t, err)
		assert.Equal(t, models.Permission(models.PermissionUnknown), perm, "Should return PermissionUnknown")
	})
}
