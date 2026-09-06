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

func cachedProjectAccess(t *testing.T, userID int64) (projectAccessEntry, bool) {
	t.Helper()
	var e projectAccessEntry
	has, err := keyvalue.GetWithValue("shared-cache-"+projectAccessCacheKey(userID), &e)
	require.NoError(t, err)
	return e, has
}

func TestProjectAccessCache(t *testing.T) {
	t.Run("served across sessions until invalidated", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		first := resolveAccess(t, 1)
		cached, has := cachedProjectAccess(t, 1)
		require.True(t, has)
		assert.Equal(t, first.permissions, cached.Permissions)
		assert.Equal(t, first.permissions, resolveAccess(t, 1).permissions)

		invalidateProjectAccess(1)
		_, has = cachedProjectAccess(t, 1)
		assert.False(t, has)
	})
	t.Run("invalidating everything drops every entry", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		resolveAccess(t, 1)
		resolveAccess(t, 2)
		invalidateAllProjectAccess()
		_, has := cachedProjectAccess(t, 1)
		assert.False(t, has)
		_, has = cachedProjectAccess(t, 2)
		assert.False(t, has)
	})
	t.Run("sharing a project with a user shows up after commit", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		// user 2 has no access to project 1 in the fixtures
		before := resolveAccess(t, 2)
		_, has := before.permission(1)
		require.False(t, has)

		s := db.NewSession()
		share := &ProjectUser{ProjectID: 1, Username: "user2", Permission: PermissionRead}
		require.NoError(t, share.Create(s, &user.User{ID: 1}))
		// Not visible before commit: the hook has not run and the entry is still the old one.
		_, has = cachedProjectAccess(t, 2)
		assert.True(t, has)
		require.NoError(t, s.Commit())
		s.Close()

		_, has = cachedProjectAccess(t, 2)
		assert.False(t, has)
		_, has = resolveAccess(t, 2).permission(1)
		assert.True(t, has)
	})
	t.Run("a new child project is visible to everyone with access to the parent", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		resolveAccess(t, 1)
		s := db.NewSession()
		parent := int64(1)
		child := &Project{Title: "child", ParentProjectID: &parent}
		require.NoError(t, child.Create(s, &user.User{ID: 1}))
		require.NoError(t, s.Commit())
		s.Close()
		_, has := resolveAccess(t, 1).permission(child.ID)
		assert.True(t, has)
	})
	t.Run("a session that has written bypasses the cache", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		resolveAccess(t, 1)

		s := db.NewSession()
		defer s.Close()
		created := &Project{Title: "x"}
		require.NoError(t, created.Create(s, &user.User{ID: 1}))

		access, err := getProjectAccessForUser(s, 1)
		require.NoError(t, err)
		_, has := access.permission(created.ID)
		assert.True(t, has)

		require.NoError(t, s.Rollback())
	})
}
