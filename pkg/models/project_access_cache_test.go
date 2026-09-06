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
	"code.vikunja.io/api/pkg/modules/keyvalue"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func resolveAccess(t *testing.T, userID int64) *projectAccess {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	pa, err := getProjectAccessForUser(s, userID)
	require.NoError(t, err)
	return pa
}

func projectPermission(t *testing.T, userID, projectID int64) (Permission, bool) {
	t.Helper()
	return resolveAccess(t, userID).permission(projectID)
}

func primeAccess(t *testing.T, userID int64) {
	t.Helper()
	resolveAccess(t, userID)
}

func hasSharedAccess(t *testing.T, userID int64) bool {
	t.Helper()
	has, err := keyvalue.GetWithValue("shared-cache-"+projectAccessCacheKey(userID), &projectAccess{})
	require.NoError(t, err)
	return has
}

func TestProjectAccessCache(t *testing.T) {
	t.Run("sharing a project with a user shows up after commit", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		_, has := projectPermission(t, 2, 1)
		require.False(t, has)

		s := db.NewSession()
		share := &ProjectUser{ProjectID: 1, Username: "user2", Permission: PermissionRead}
		require.NoError(t, share.Create(s, &user.User{ID: 1}))

		_, has = projectPermission(t, 2, 1)
		assert.False(t, has)

		require.NoError(t, s.Commit())
		s.Close()

		permission, has := projectPermission(t, 2, 1)
		require.True(t, has)
		assert.Equal(t, PermissionRead, permission)
	})

	t.Run("a new child project is visible to everyone with access to the parent", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		// User 2 owns nothing and reaches project 3 through a share only.
		permission, has := projectPermission(t, 2, 3)
		require.True(t, has)
		require.Equal(t, PermissionRead, permission)

		s := db.NewSession()
		parent := int64(3)
		child := &Project{Title: "child of a shared project", ParentProjectID: &parent}
		require.NoError(t, child.Create(s, &user.User{ID: 3}))
		require.NoError(t, s.Commit())
		s.Close()

		permission, has = projectPermission(t, 2, child.ID)
		require.True(t, has)
		assert.Equal(t, PermissionRead, permission)
	})

	t.Run("a new top level project is visible to its owner", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		primeAccess(t, 1)

		s := db.NewSession()
		created := &Project{Title: "a new top level project"}
		require.NoError(t, created.Create(s, &user.User{ID: 1}))
		require.NoError(t, s.Commit())
		s.Close()

		permission, has := projectPermission(t, 1, created.ID)
		require.True(t, has)
		assert.Equal(t, PermissionAdmin, permission)
	})

	t.Run("a session that has written bypasses the cache", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		primeAccess(t, 1)

		s := db.NewSession()
		created := &Project{Title: "written in this session"}
		require.NoError(t, created.Create(s, &user.User{ID: 1}))

		access, err := getProjectAccessForUser(s, 1)
		require.NoError(t, err)
		permission, has := access.permission(created.ID)
		require.True(t, has)
		assert.Equal(t, PermissionAdmin, permission)

		require.NoError(t, s.Commit())
		s.Close()

		permission, has = projectPermission(t, 1, created.ID)
		require.True(t, has)
		assert.Equal(t, PermissionAdmin, permission)
	})

	t.Run("removing a user share revokes access", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		// Project 9 is user 6's and reaches user 1 through the direct share only.
		permission, has := projectPermission(t, 1, 9)
		require.True(t, has)
		require.Equal(t, PermissionRead, permission)

		s := db.NewSession()
		share := &ProjectUser{ProjectID: 9, Username: "user1"}
		require.NoError(t, share.Delete(s, &user.User{ID: 6}))
		require.NoError(t, s.Commit())
		s.Close()

		_, has = projectPermission(t, 1, 9)
		assert.False(t, has)
	})

	t.Run("lowering a user share lowers the permission", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		permission, has := projectPermission(t, 1, 11)
		require.True(t, has)
		require.Equal(t, PermissionAdmin, permission)

		s := db.NewSession()
		share := &ProjectUser{ProjectID: 11, Username: "user1", Permission: PermissionRead}
		require.NoError(t, share.Update(s, &user.User{ID: 6}))
		require.NoError(t, s.Commit())
		s.Close()

		permission, has = projectPermission(t, 1, 11)
		require.True(t, has)
		assert.Equal(t, PermissionRead, permission)
	})

	t.Run("removing a team member revokes what only that team granted", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		// User 2 reaches project 33 through team 1 only, project 3 also directly.
		permission, has := projectPermission(t, 2, 33)
		require.True(t, has)
		require.Equal(t, PermissionWrite, permission)

		s := db.NewSession()
		member := &TeamMember{TeamID: 1, Username: "user2"}
		require.NoError(t, member.Delete(s, &user.User{ID: 1}))
		require.NoError(t, s.Commit())
		s.Close()

		_, has = projectPermission(t, 2, 33)
		assert.False(t, has)

		permission, has = projectPermission(t, 2, 3)
		require.True(t, has)
		assert.Equal(t, PermissionRead, permission)
	})

	t.Run("removing a team share revokes access to the project and its children", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		// Team 16 has user 14 as its only member and read on project 41, parent of 42.
		permission, has := projectPermission(t, 14, 42)
		require.True(t, has)
		require.Equal(t, PermissionRead, permission)

		s := db.NewSession()
		share := &TeamProject{TeamID: 16, ProjectID: 41}
		require.NoError(t, share.Delete(s, &user.User{ID: 6}))
		require.NoError(t, s.Commit())
		s.Close()

		_, has = projectPermission(t, 14, 41)
		assert.False(t, has)
		_, has = projectPermission(t, 14, 42)
		assert.False(t, has)
	})

	t.Run("deleting a team revokes access", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		permission, has := projectPermission(t, 14, 42)
		require.True(t, has)
		require.Equal(t, PermissionRead, permission)

		s := db.NewSession()
		team := &Team{ID: 16}
		require.NoError(t, team.Delete(s, &user.User{ID: 6}))
		require.NoError(t, s.Commit())
		s.Close()

		_, has = projectPermission(t, 14, 41)
		assert.False(t, has)
		_, has = projectPermission(t, 14, 42)
		assert.False(t, has)
	})

	t.Run("detaching a child project revokes the inherited access", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		permission, has := projectPermission(t, 14, 42)
		require.True(t, has)
		require.Equal(t, PermissionRead, permission)

		s := db.NewSession()
		// User 6 owns both 41 and 42.
		child, err := GetProjectSimpleByID(s, 42)
		require.NoError(t, err)
		child.ParentProjectID = noParentProjectID()
		require.NoError(t, UpdateProject(s, child, &user.User{ID: 6}, false))
		require.NoError(t, s.Commit())
		s.Close()

		_, has = projectPermission(t, 14, 42)
		assert.False(t, has)

		permission, has = projectPermission(t, 14, 41)
		require.True(t, has)
		assert.Equal(t, PermissionRead, permission)
	})

	t.Run("transferring ownership moves the project to the new owner", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		// User 1 owns project 1 and reaches it no other way; user 2 not at all.
		permission, has := projectPermission(t, 1, 1)
		require.True(t, has)
		require.Equal(t, PermissionAdmin, permission)
		_, has = projectPermission(t, 2, 1)
		require.False(t, has)

		s := db.NewSession()
		_, err := ReassignProjectOwner(s, &user.User{ID: 1}, 1, 2)
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		s.Close()

		_, has = projectPermission(t, 1, 1)
		assert.False(t, has)

		permission, has = projectPermission(t, 2, 1)
		require.True(t, has)
		assert.Equal(t, PermissionAdmin, permission)
	})

	t.Run("deleting a project revokes access for the users it was shared with", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		permission, has := projectPermission(t, 1, 9)
		require.True(t, has)
		require.Equal(t, PermissionRead, permission)

		s := db.NewSession()
		project := &Project{ID: 9, OwnerID: 6}
		require.NoError(t, project.Delete(s, &user.User{ID: 6}))
		require.NoError(t, s.Commit())
		s.Close()

		_, has = projectPermission(t, 1, 9)
		assert.False(t, has)
	})
	t.Run("touching a project's timestamp keeps every entry", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		primeAccess(t, 1)
		primeAccess(t, 2)

		s := db.NewSession()
		require.NoError(t, updateProjectLastUpdated(s, &Project{ID: 1}))
		require.NoError(t, s.Commit())
		s.Close()

		assert.True(t, hasSharedAccess(t, 1))
		assert.True(t, hasSharedAccess(t, 2))
	})

	t.Run("renaming a project keeps every entry", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		primeAccess(t, 1)
		primeAccess(t, 2)

		s := db.NewSession()
		project, err := GetProjectSimpleByID(s, 1)
		require.NoError(t, err)
		project.Title = "renamed"
		require.NoError(t, UpdateProject(s, project, &user.User{ID: 1}, false))
		require.NoError(t, s.Commit())
		s.Close()

		assert.True(t, hasSharedAccess(t, 1))
		assert.True(t, hasSharedAccess(t, 2))
	})

	t.Run("creating a top-level project drops only the owner's entry", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		primeAccess(t, 1)
		primeAccess(t, 2)

		s := db.NewSession()
		created := &Project{Title: "x"}
		require.NoError(t, created.Create(s, &user.User{ID: 1}))
		require.NoError(t, s.Commit())
		s.Close()

		assert.False(t, hasSharedAccess(t, 1))
		assert.True(t, hasSharedAccess(t, 2))
	})

	t.Run("removing a share by user id drops only that user's entry", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		primeAccess(t, 1)
		primeAccess(t, 2)

		s := db.NewSession()
		_, err := s.Delete(&ProjectUser{ProjectID: 9, UserID: 1})
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		s.Close()

		assert.False(t, hasSharedAccess(t, 1))
		assert.True(t, hasSharedAccess(t, 2))

		_, has := projectPermission(t, 1, 9)
		assert.False(t, has)
	})

	t.Run("reassigning ownership through a bare bean drops every entry", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		primeAccess(t, 1)
		primeAccess(t, 2)

		s := db.NewSession()
		_, err := s.Where("id = ?", 1).Cols("owner_id").Update(&Project{OwnerID: 2})
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		s.Close()

		assert.False(t, hasSharedAccess(t, 1))
		assert.False(t, hasSharedAccess(t, 2))

		permission, has := projectPermission(t, 2, 1)
		require.True(t, has)
		assert.Equal(t, PermissionAdmin, permission)

		_, has = projectPermission(t, 1, 1)
		assert.False(t, has)
	})
}
