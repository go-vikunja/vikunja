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
	"strconv"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/builder"
	"xorm.io/xorm"
)

func TestGetProjectAccessForUser(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	access, err := getProjectAccessForUser(s, 1)
	require.NoError(t, err)

	tests := []struct {
		name      string
		projectID int64
		want      Permission
		has       bool
	}{
		{"owner is admin", 1, PermissionAdmin, true},
		{"direct read share", 9, PermissionRead, true},
		{"direct write share", 10, PermissionWrite, true},
		{"direct admin share", 11, PermissionAdmin, true},
		{"team read share", 6, PermissionRead, true},
		{"team admin share", 8, PermissionAdmin, true},
		{"child inherits parent share", 43, PermissionWrite, true},
		{"owned orphan with missing parent", 39, PermissionAdmin, true},
		{"not shared", 2, PermissionUnknown, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, has := access.permission(tt.projectID)
			assert.Equal(t, tt.has, has)
			if has {
				assert.Equal(t, tt.want, got)
			}
		})
	}

	t.Run("child share via team on parent", func(t *testing.T) {
		access, err := getProjectAccessForUser(s, 14)
		require.NoError(t, err)
		got, has := access.permission(42)
		assert.True(t, has)
		assert.Equal(t, PermissionRead, got)
	})

	t.Run("team share without membership is not accessible", func(t *testing.T) {
		// User 14 is only a member of team 16; project 6 is shared with team 2.
		access, err := getProjectAccessForUser(s, 14)
		require.NoError(t, err)
		_, has := access.permission(6)
		assert.False(t, has)
	})

	t.Run("memoized per session", func(t *testing.T) {
		again, err := getProjectAccessForUser(s, 1)
		require.NoError(t, err)
		assert.Same(t, access, again)

		other := db.NewSession()
		defer other.Close()
		fresh, err := getProjectAccessForUser(other, 1)
		require.NoError(t, err)
		assert.NotSame(t, access, fresh)
	})
}

// GHSA-pjr3-86v4-5p7w
func TestGetProjectAccessForUser_NearestGrantWins(t *testing.T) {
	// 27 -> 12 -> 25 -> 26, all inheriting user 1's read on 27.
	accessWith := func(t *testing.T, grants ...*ProjectUser) *projectAccess {
		t.Helper()
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		t.Cleanup(func() { s.Close() })
		for _, g := range grants {
			g.UserID = 1
			_, err := s.Insert(g)
			require.NoError(t, err)
		}
		access, err := getProjectAccessForUser(s, 1)
		require.NoError(t, err)
		return access
	}
	assertPermission := func(t *testing.T, access *projectAccess, projectID int64, want Permission) {
		t.Helper()
		got, has := access.permission(projectID)
		assert.True(t, has)
		assert.Equal(t, want, got, "project %d", projectID)
	}

	t.Run("an inherited admin is not lowered by a child grant", func(t *testing.T) {
		// User 1 is admin on project 29 and therefore on its child 14, which user 6 owns.
		access := accessWith(t, &ProjectUser{ProjectID: 14, Permission: PermissionRead})
		assertPermission(t, access, 14, PermissionAdmin)
	})
	t.Run("a child read lowers an inherited write", func(t *testing.T) {
		access := accessWith(t,
			&ProjectUser{ProjectID: 12, Permission: PermissionWrite},
			&ProjectUser{ProjectID: 25, Permission: PermissionRead},
		)
		assertPermission(t, access, 12, PermissionWrite)
		assertPermission(t, access, 25, PermissionRead)
		assertPermission(t, access, 26, PermissionRead)
	})
	t.Run("a child write raises an inherited read", func(t *testing.T) {
		access := accessWith(t, &ProjectUser{ProjectID: 25, Permission: PermissionWrite})
		assertPermission(t, access, 12, PermissionRead)
		assertPermission(t, access, 25, PermissionWrite)
		assertPermission(t, access, 26, PermissionWrite)
	})
	t.Run("the owner stays admin despite a lower share", func(t *testing.T) {
		access := accessWith(t, &ProjectUser{ProjectID: 1, Permission: PermissionRead})
		assertPermission(t, access, 1, PermissionAdmin)
	})
	t.Run("a user and a team grant on the same project take the higher", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		// User 1 is a member of team 1.
		_, err := s.Insert(&ProjectUser{ProjectID: 2, UserID: 1, Permission: PermissionRead})
		require.NoError(t, err)
		_, err = s.Insert(&TeamProject{ProjectID: 2, TeamID: 1, Permission: PermissionWrite})
		require.NoError(t, err)

		access, err := getProjectAccessForUser(s, 1)
		require.NoError(t, err)
		assertPermission(t, access, 2, PermissionWrite)
	})
}

func TestGetProjectAccessForUser_OutOfEnumGrantIsNoGrant(t *testing.T) {
	for _, permission := range []Permission{-5, 3} {
		t.Run(strconv.Itoa(int(permission)), func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			_, err := s.Insert(&ProjectUser{ProjectID: 2, UserID: 1, Permission: permission})
			require.NoError(t, err)

			access, err := getProjectAccessForUser(s, 1)
			require.NoError(t, err)

			_, has := access.permission(2)
			assert.False(t, has)
			assert.NotContains(t, access.sortedIDs, int64(2))
		})
	}
}

func projectIDsMatching(t *testing.T, s *xorm.Session, cond builder.Cond) []int64 {
	t.Helper()
	projects := []*Project{}
	require.NoError(t, s.Where(cond).Cols("id").OrderBy("id").Find(&projects))
	ids := make([]int64, 0, len(projects))
	for _, p := range projects {
		ids = append(ids, p.ID)
	}
	return ids
}

func TestProjectAccessCond(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	access, err := getProjectAccessForUser(s, 1)
	require.NoError(t, err)

	t.Run("no access matches nothing", func(t *testing.T) {
		empty := &projectAccess{userID: 1}
		assert.Empty(t, projectIDsMatching(t, s, empty.cond("id")))
	})

	t.Run("subquery fallback matches the same projects with bounded parameters", func(t *testing.T) {
		fallback := access.condViaSubquery("id")
		assert.Equal(t, projectIDsMatching(t, s, access.cond("id")), projectIDsMatching(t, s, fallback))

		_, args, err := builder.ToSQL(fallback)
		require.NoError(t, err)
		assert.Len(t, args, 3)
	})

	t.Run("link share is confined to its project", func(t *testing.T) {
		cond, err := accessibleProjectIDsCond(s, &LinkSharing{ProjectID: 42}, "id")
		require.NoError(t, err)
		assert.Equal(t, []int64{42}, projectIDsMatching(t, s, cond))
	})
}

func TestGetAllParentProjects(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	chain, err := GetAllParentProjects(s, 43)
	require.NoError(t, err)
	assert.Len(t, chain, 2)
	assert.NotNil(t, chain[43])
	assert.NotNil(t, chain[10])

	again, err := GetAllParentProjects(s, 43)
	require.NoError(t, err)
	assert.Equal(t, chain, again)

	t.Run("callers cannot corrupt the memo", func(t *testing.T) {
		require.NotNil(t, chain[43].ParentProjectID)
		*chain[43].ParentProjectID = 999999
		chain[43].Title = "mutated"
		delete(chain, 10)

		fresh, err := GetAllParentProjects(s, 43)
		require.NoError(t, err)
		assert.Len(t, fresh, 2)
		assert.NotEqual(t, "mutated", fresh[43].Title)
		require.NotNil(t, fresh[43].ParentProjectID)
		assert.Equal(t, int64(10), *fresh[43].ParentProjectID)
	})
}

func TestSessionMemoStopsAfterWrite(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	access, err := getProjectAccessForUser(s, 1)
	require.NoError(t, err)
	_, has := access.permission(2)
	assert.False(t, has)

	_, err = s.Insert(&ProjectUser{ProjectID: 2, UserID: 1, Permission: PermissionRead})
	require.NoError(t, err)

	afterWrite, err := getProjectAccessForUser(s, 1)
	require.NoError(t, err)
	assert.NotSame(t, access, afterWrite)
	got, has := afterWrite.permission(2)
	assert.True(t, has)
	assert.Equal(t, PermissionRead, got)

	chain, err := GetAllParentProjects(s, 43)
	require.NoError(t, err)
	assert.Len(t, chain, 2)

	_, err = s.ID(43).Cols("parent_project_id").Update(&Project{})
	require.NoError(t, err)
	require.NoError(t, moveProjectAncestors(s, 43, 0))

	chain, err = GetAllParentProjects(s, 43)
	require.NoError(t, err)
	assert.Len(t, chain, 1)
}

func TestProjectAccess_InvalidatedByShareInSameSession(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	owner := &user.User{ID: 1}
	other := &user.User{ID: 2}
	project := &Project{ID: 1}

	canRead, _, err := project.CanRead(s, other)
	require.NoError(t, err)
	assert.False(t, canRead)

	share := &ProjectUser{ProjectID: 1, Username: "user2", Permission: PermissionRead}
	require.NoError(t, share.Create(s, owner))

	canRead, _, err = project.CanRead(s, other)
	require.NoError(t, err)
	assert.True(t, canRead)

	require.NoError(t, share.Delete(s, owner))

	canRead, _, err = project.CanRead(s, other)
	require.NoError(t, err)
	assert.False(t, canRead)
}
