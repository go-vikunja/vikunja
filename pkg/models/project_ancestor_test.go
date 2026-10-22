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
	"slices"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/builder"
	"xorm.io/xorm"
)

func allProjectAncestors(t *testing.T, s *xorm.Session) []ProjectAncestor {
	t.Helper()
	rows := []ProjectAncestor{}
	require.NoError(t, s.Find(&rows))
	return rows
}

func projectAncestorsOf(t *testing.T, s *xorm.Session, projectIDs ...int64) []ProjectAncestor {
	t.Helper()
	rows := []ProjectAncestor{}
	require.NoError(t, s.Where(builder.In("project_id", projectIDs)).Find(&rows))
	return rows
}

func TestProjectAncestorsFixtureMatchesProjects(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	fromFixture := allProjectAncestors(t, s)
	require.NoError(t, RebuildProjectAncestors(s))
	rebuilt := allProjectAncestors(t, s)

	assert.ElementsMatch(t, fromFixture, rebuilt)
}

func TestProjectAncestorsOnCreate(t *testing.T) {
	usr := &user.User{ID: 6, Username: "user6"}

	t.Run("top level project", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		project := &Project{Title: "new top level"}
		require.NoError(t, project.Create(s, usr))

		assert.ElementsMatch(t, []ProjectAncestor{
			{ProjectID: project.ID, AncestorID: project.ID, Depth: 0},
		}, projectAncestorsOf(t, s, project.ID))
	})

	t.Run("child of a project with ancestors", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// 26 -> 25 -> 12 -> 27
		project := &Project{Title: "new child", ParentProjectID: Ptr(int64(26))}
		require.NoError(t, project.Create(s, usr))

		assert.ElementsMatch(t, []ProjectAncestor{
			{ProjectID: project.ID, AncestorID: project.ID, Depth: 0},
			{ProjectID: project.ID, AncestorID: 26, Depth: 1},
			{ProjectID: project.ID, AncestorID: 25, Depth: 2},
			{ProjectID: project.ID, AncestorID: 12, Depth: 3},
			{ProjectID: project.ID, AncestorID: 27, Depth: 4},
		}, projectAncestorsOf(t, s, project.ID))
	})
}

func TestProjectAncestorsOnMove(t *testing.T) {
	usr := &user.User{ID: 6, Username: "user6"}

	reparent := func(t *testing.T, s *xorm.Session, projectID, newParentID int64) {
		t.Helper()
		project, err := GetProjectSimpleByID(s, projectID)
		require.NoError(t, err)
		project.ParentProjectID = Ptr(newParentID)
		require.NoError(t, UpdateProject(s, project, usr, false))
	}

	t.Run("subtree under another parent", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// 26 -> 25 -> 12 -> 27, moving 12 under 28
		reparent(t, s, 12, 28)

		assert.ElementsMatch(t, []ProjectAncestor{
			{ProjectID: 12, AncestorID: 12, Depth: 0},
			{ProjectID: 12, AncestorID: 28, Depth: 1},
			{ProjectID: 25, AncestorID: 25, Depth: 0},
			{ProjectID: 25, AncestorID: 12, Depth: 1},
			{ProjectID: 25, AncestorID: 28, Depth: 2},
			{ProjectID: 26, AncestorID: 26, Depth: 0},
			{ProjectID: 26, AncestorID: 25, Depth: 1},
			{ProjectID: 26, AncestorID: 12, Depth: 2},
			{ProjectID: 26, AncestorID: 28, Depth: 3},
		}, projectAncestorsOf(t, s, 12, 25, 26))
	})

	t.Run("top level project under a parent", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// 42 -> 41 at the top level, moving 41 under 26 -> 25 -> 12 -> 27
		reparent(t, s, 41, 26)

		assert.ElementsMatch(t, []ProjectAncestor{
			{ProjectID: 41, AncestorID: 41, Depth: 0},
			{ProjectID: 41, AncestorID: 26, Depth: 1},
			{ProjectID: 41, AncestorID: 25, Depth: 2},
			{ProjectID: 41, AncestorID: 12, Depth: 3},
			{ProjectID: 41, AncestorID: 27, Depth: 4},
			{ProjectID: 42, AncestorID: 42, Depth: 0},
			{ProjectID: 42, AncestorID: 41, Depth: 1},
			{ProjectID: 42, AncestorID: 26, Depth: 2},
			{ProjectID: 42, AncestorID: 25, Depth: 3},
			{ProjectID: 42, AncestorID: 12, Depth: 4},
			{ProjectID: 42, AncestorID: 27, Depth: 5},
		}, projectAncestorsOf(t, s, 41, 42))
	})

	t.Run("subtree to the top level", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		reparent(t, s, 12, 0)

		assert.ElementsMatch(t, []ProjectAncestor{
			{ProjectID: 12, AncestorID: 12, Depth: 0},
			{ProjectID: 25, AncestorID: 25, Depth: 0},
			{ProjectID: 25, AncestorID: 12, Depth: 1},
			{ProjectID: 26, AncestorID: 26, Depth: 0},
			{ProjectID: 26, AncestorID: 25, Depth: 1},
			{ProjectID: 26, AncestorID: 12, Depth: 2},
		}, projectAncestorsOf(t, s, 12, 25, 26))
	})
}

func TestProjectAncestorsWithoutClosureRows(t *testing.T) {
	// 12 -> 27, with 25 and 26 below 12
	const broken = int64(12)

	setup := func(t *testing.T) *xorm.Session {
		t.Helper()
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		t.Cleanup(func() {
			require.NoError(t, s.Close())
		})

		_, err := s.Where(builder.Eq{"project_id": broken}).Delete(&ProjectAncestor{})
		require.NoError(t, err)
		return s
	}

	t.Run("create a child below it", func(t *testing.T) {
		s := setup(t)

		err := insertProjectAncestors(s, 4711, broken)
		require.ErrorContains(t, err, "project 12 has no ancestor rows")
		assert.Empty(t, projectAncestorsOf(t, s, 4711))
	})

	t.Run("move it below another project", func(t *testing.T) {
		s := setup(t)

		err := moveProjectAncestors(s, broken, 28)
		require.ErrorContains(t, err, "project 12 has no self row in project_ancestors")

		assert.ElementsMatch(t, []ProjectAncestor{
			{ProjectID: 25, AncestorID: 25, Depth: 0},
			{ProjectID: 25, AncestorID: 12, Depth: 1},
			{ProjectID: 25, AncestorID: 27, Depth: 2},
			{ProjectID: 26, AncestorID: 26, Depth: 0},
			{ProjectID: 26, AncestorID: 25, Depth: 1},
			{ProjectID: 26, AncestorID: 12, Depth: 2},
			{ProjectID: 26, AncestorID: 27, Depth: 3},
		}, projectAncestorsOf(t, s, 25, 26))
	})

	t.Run("move another project below it", func(t *testing.T) {
		s := setup(t)

		err := moveProjectAncestors(s, 26, broken)
		require.ErrorContains(t, err, "project 12 has no ancestor rows")

		assert.ElementsMatch(t, []ProjectAncestor{
			{ProjectID: 26, AncestorID: 26, Depth: 0},
			{ProjectID: 26, AncestorID: 25, Depth: 1},
			{ProjectID: 26, AncestorID: 12, Depth: 2},
			{ProjectID: 26, AncestorID: 27, Depth: 3},
		}, projectAncestorsOf(t, s, 26))
	})
}

func TestProjectAncestorsOnDelete(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	usr := &user.User{ID: 6, Username: "user6"}

	// 12 has 25 as child and 26 as grandchild
	removed := []int64{12, 25, 26}
	survivors := []ProjectAncestor{}
	for _, row := range allProjectAncestors(t, s) {
		if slices.Contains(removed, row.ProjectID) || slices.Contains(removed, row.AncestorID) {
			continue
		}
		survivors = append(survivors, row)
	}
	require.NotEmpty(t, survivors)

	project, err := GetProjectSimpleByID(s, 12)
	require.NoError(t, err)
	require.NoError(t, project.Delete(s, usr))

	assert.Empty(t, projectAncestorsOf(t, s, removed...))
	assert.ElementsMatch(t, survivors, allProjectAncestors(t, s))
}

func TestProjectAncestorsOnRepair(t *testing.T) {
	t.Run("repair", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// 12 -> 27, dropping 27 orphans the whole subtree below 12
		_, err := s.ID(27).Delete(&Project{})
		require.NoError(t, err)

		result, err := RepairOrphanedProjects(s, false)
		require.NoError(t, err)
		// 12 plus fixture project 39, which points at a parent that never existed
		assert.Equal(t, 2, result.Found)
		assert.Equal(t, 2, result.Repaired)

		assert.ElementsMatch(t, []ProjectAncestor{
			{ProjectID: 12, AncestorID: 12, Depth: 0},
			{ProjectID: 25, AncestorID: 25, Depth: 0},
			{ProjectID: 25, AncestorID: 12, Depth: 1},
			{ProjectID: 26, AncestorID: 26, Depth: 0},
			{ProjectID: 26, AncestorID: 25, Depth: 1},
			{ProjectID: 26, AncestorID: 12, Depth: 2},
		}, projectAncestorsOf(t, s, 12, 25, 26))

		ofDeleted := []ProjectAncestor{}
		require.NoError(t, s.
			Where(builder.Or(builder.Eq{"project_id": 27}, builder.Eq{"ancestor_id": 27})).
			Find(&ofDeleted))
		assert.Empty(t, ofDeleted)
	})

	t.Run("dry run", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		before := allProjectAncestors(t, s)

		_, err := s.ID(27).Delete(&Project{})
		require.NoError(t, err)

		result, err := RepairOrphanedProjects(s, true)
		require.NoError(t, err)
		assert.Equal(t, 2, result.Found)
		assert.Equal(t, 0, result.Repaired)

		assert.ElementsMatch(t, before, allProjectAncestors(t, s))
	})
}

func TestProjectAncestorsRebuildTerminatesOnParentCycle(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	// Two concurrent reparents can each see an acyclic tree and both commit, so the rebuild
	// has to cope with 21 -> 22 -> 21. Both are owned by user 1.
	parent := int64(21)
	_, err := s.ID(22).Cols("parent_project_id").Update(&Project{ParentProjectID: &parent})
	require.NoError(t, err)

	require.NoError(t, RebuildProjectAncestors(s))

	// The visited guard ends each chain after one hop into the cycle.
	assert.ElementsMatch(t, []ProjectAncestor{
		{ProjectID: 21, AncestorID: 21, Depth: 0},
		{ProjectID: 21, AncestorID: 22, Depth: 1},
		{ProjectID: 22, AncestorID: 22, Depth: 0},
		{ProjectID: 22, AncestorID: 21, Depth: 1},
	}, projectAncestorsOf(t, s, 21, 22))

	access, err := getProjectAccessForUser(s, 1)
	require.NoError(t, err)
	for _, projectID := range []int64{21, 22} {
		got, has := access.permission(projectID)
		assert.True(t, has)
		assert.Equal(t, PermissionAdmin, got)
	}
}

func TestProjectAccessFollowsMove(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	// 26 -> 25 -> 12 -> 27; user 7 reads 27, user 4 reads 28, neither has any other grant on them.
	_, err := s.Insert(&ProjectUser{ProjectID: 27, UserID: 7, Permission: PermissionRead})
	require.NoError(t, err)
	_, err = s.Insert(&ProjectUser{ProjectID: 28, UserID: 4, Permission: PermissionWrite})
	require.NoError(t, err)

	permissionOf := func(t *testing.T, userID, projectID int64) (Permission, bool) {
		t.Helper()
		access, err := getProjectAccessForUser(s, userID)
		require.NoError(t, err)
		return access.permission(projectID)
	}
	assertSubtree := func(t *testing.T, userID int64, want Permission, has bool) {
		t.Helper()
		for _, projectID := range []int64{12, 25, 26} {
			got, gotHas := permissionOf(t, userID, projectID)
			assert.Equal(t, has, gotHas, "user %d project %d", userID, projectID)
			if has {
				assert.Equal(t, want, got, "user %d project %d", userID, projectID)
			}
		}
	}

	assertSubtree(t, 7, PermissionRead, true)
	assertSubtree(t, 4, PermissionUnknown, false)

	project, err := GetProjectSimpleByID(s, 12)
	require.NoError(t, err)
	project.ParentProjectID = Ptr(int64(28))
	require.NoError(t, UpdateProject(s, project, &user.User{ID: 6, Username: "user6"}, false))

	assertSubtree(t, 7, PermissionUnknown, false)
	assertSubtree(t, 4, PermissionWrite, true)

	project.ParentProjectID = Ptr(int64(0))
	require.NoError(t, UpdateProject(s, project, &user.User{ID: 6, Username: "user6"}, false))

	assertSubtree(t, 7, PermissionUnknown, false)
	assertSubtree(t, 4, PermissionUnknown, false)
	assertSubtree(t, 6, PermissionAdmin, true)
}
