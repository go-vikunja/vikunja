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

package migration

import (
	"testing"

	"code.vikunja.io/api/pkg/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/builder"
)

func TestBackfillProjectAncestors20260908202544(t *testing.T) {
	x, err := db.CreateTestEngine()
	require.NoError(t, err)

	tables := []interface{}{projects20260908202544{}, projectAncestors20260908202544{}}
	// x is the process-global test engine; the projects table here is a minimal stub.
	t.Cleanup(func() {
		require.NoError(t, x.DropTables(tables...))
	})
	require.NoError(t, x.DropTables(tables...))
	require.NoError(t, x.Sync2(tables...))

	parent := func(id int64) *int64 { return &id }

	_, err = x.Insert(
		// chain of depth 3
		&projects20260908202544{ID: 1},
		&projects20260908202544{ID: 2, ParentProjectID: parent(1)},
		&projects20260908202544{ID: 3, ParentProjectID: parent(2)},
		&projects20260908202544{ID: 4, ParentProjectID: parent(3)},
		// cycle
		&projects20260908202544{ID: 20, ParentProjectID: parent(21)},
		&projects20260908202544{ID: 21, ParentProjectID: parent(20)},
		// root stored as 0, root stored as null
		&projects20260908202544{ID: 30, ParentProjectID: parent(0)},
		&projects20260908202544{ID: 31},
		// child of a project which does not exist
		&projects20260908202544{ID: 32, ParentProjectID: parent(999999)},
	)
	require.NoError(t, err)

	const bigRootID = 1000
	bigChildrenCount := projectAncestorsBatch20260908202544 + 2
	bigTree := make([]interface{}, 0, bigChildrenCount+1)
	bigTree = append(bigTree, &projects20260908202544{ID: bigRootID})
	for i := range bigChildrenCount {
		bigTree = append(bigTree, &projects20260908202544{ID: bigRootID + 1 + int64(i), ParentProjectID: parent(bigRootID)})
	}
	_, err = x.Insert(bigTree...)
	require.NoError(t, err)

	require.NoError(t, backfillProjectAncestors20260908202544(x))

	ancestorsOf := func(projectIDs ...int64) []projectAncestors20260908202544 {
		t.Helper()
		rows := []projectAncestors20260908202544{}
		require.NoError(t, x.Where(builder.In("project_id", projectIDs)).Find(&rows))
		return rows
	}

	assert.ElementsMatch(t, []projectAncestors20260908202544{
		{ProjectID: 1, AncestorID: 1, Depth: 0},
		{ProjectID: 2, AncestorID: 2, Depth: 0},
		{ProjectID: 2, AncestorID: 1, Depth: 1},
		{ProjectID: 3, AncestorID: 3, Depth: 0},
		{ProjectID: 3, AncestorID: 2, Depth: 1},
		{ProjectID: 3, AncestorID: 1, Depth: 2},
		{ProjectID: 4, AncestorID: 4, Depth: 0},
		{ProjectID: 4, AncestorID: 3, Depth: 1},
		{ProjectID: 4, AncestorID: 2, Depth: 2},
		{ProjectID: 4, AncestorID: 1, Depth: 3},
	}, ancestorsOf(1, 2, 3, 4))

	assert.ElementsMatch(t, []projectAncestors20260908202544{
		{ProjectID: 20, AncestorID: 20, Depth: 0},
		{ProjectID: 20, AncestorID: 21, Depth: 1},
		{ProjectID: 21, AncestorID: 21, Depth: 0},
		{ProjectID: 21, AncestorID: 20, Depth: 1},
	}, ancestorsOf(20, 21))

	assert.ElementsMatch(t, []projectAncestors20260908202544{
		{ProjectID: 30, AncestorID: 30, Depth: 0},
		{ProjectID: 31, AncestorID: 31, Depth: 0},
		{ProjectID: 32, AncestorID: 32, Depth: 0},
	}, ancestorsOf(30, 31, 32))

	bigCount, err := x.Where(builder.Gte{"project_id": bigRootID}).Count(&projectAncestors20260908202544{})
	require.NoError(t, err)
	// one self row per project plus one row per child pointing at the root
	assert.Equal(t, int64(bigChildrenCount*2+1), bigCount)

	// re-run: drifted rows go, missing rows come back, nothing collides
	_, err = x.Where(builder.In("project_id", []int64{1, 2, 3, 4})).Delete(&projectAncestors20260908202544{})
	require.NoError(t, err)
	_, err = x.Insert(
		&projectAncestors20260908202544{ProjectID: 1, AncestorID: 1, Depth: 0},
		&projectAncestors20260908202544{ProjectID: 2, AncestorID: 2, Depth: 0},
		&projectAncestors20260908202544{ProjectID: 3, AncestorID: 3, Depth: 0},
		&projectAncestors20260908202544{ProjectID: 4, AncestorID: 4, Depth: 0},
		// stale: an ancestor which is gone, and a depth from a parent link which changed
		&projectAncestors20260908202544{ProjectID: 3, AncestorID: 999999, Depth: 7},
		&projectAncestors20260908202544{ProjectID: 4, AncestorID: 1, Depth: 9},
	)
	require.NoError(t, err)

	require.NoError(t, backfillProjectAncestors20260908202544(x))

	assert.ElementsMatch(t, []projectAncestors20260908202544{
		{ProjectID: 1, AncestorID: 1, Depth: 0},
		{ProjectID: 2, AncestorID: 2, Depth: 0},
		{ProjectID: 2, AncestorID: 1, Depth: 1},
		{ProjectID: 3, AncestorID: 3, Depth: 0},
		{ProjectID: 3, AncestorID: 2, Depth: 1},
		{ProjectID: 3, AncestorID: 1, Depth: 2},
		{ProjectID: 4, AncestorID: 4, Depth: 0},
		{ProjectID: 4, AncestorID: 3, Depth: 1},
		{ProjectID: 4, AncestorID: 2, Depth: 2},
		{ProjectID: 4, AncestorID: 1, Depth: 3},
	}, ancestorsOf(1, 2, 3, 4))
}
