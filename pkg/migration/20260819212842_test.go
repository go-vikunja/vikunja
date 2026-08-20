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

	"github.com/stretchr/testify/require"
)

func TestBackfillArchivedDescendants20260819212842(t *testing.T) {
	x, err := db.CreateTestEngine()
	require.NoError(t, err)

	tables := []interface{}{projects20260819212842{}}
	// x is the process-global test engine; the projects table here is a minimal stub.
	t.Cleanup(func() {
		require.NoError(t, x.DropTables(tables...))
	})
	require.NoError(t, x.DropTables(tables...))
	require.NoError(t, x.Sync2(tables...))

	_, err = x.Insert(
		// archived root with unflagged child + grandchild
		&projects20260819212842{ID: 1, IsArchived: true},
		&projects20260819212842{ID: 2, ParentProjectID: 1},
		&projects20260819212842{ID: 3, ParentProjectID: 2},
		// non-archived tree
		&projects20260819212842{ID: 10},
		&projects20260819212842{ID: 11, ParentProjectID: 10},
		// archived child under non-archived parent, with its own child
		&projects20260819212842{ID: 12, ParentProjectID: 10, IsArchived: true},
		&projects20260819212842{ID: 13, ParentProjectID: 12},
		// cycle
		&projects20260819212842{ID: 20, ParentProjectID: 21, IsArchived: true},
		&projects20260819212842{ID: 21, ParentProjectID: 20},
	)
	require.NoError(t, err)

	// enough children to span more than one update batch
	const bigRootID = 1000
	bigChildrenCount := archivedBackfillBatch20260819212842 + 2
	bigTree := make([]interface{}, 0, bigChildrenCount+1)
	bigTree = append(bigTree, &projects20260819212842{ID: bigRootID, IsArchived: true})
	for i := range bigChildrenCount {
		bigTree = append(bigTree, &projects20260819212842{ID: bigRootID + 1 + int64(i), ParentProjectID: bigRootID})
	}
	_, err = x.Insert(bigTree...)
	require.NoError(t, err)

	require.NoError(t, backfillArchivedDescendants20260819212842(x))

	assertArchived := func(id int64, want bool) {
		t.Helper()
		p := &projects20260819212842{}
		has, err := x.ID(id).Get(p)
		require.NoError(t, err)
		require.True(t, has)
		require.Equal(t, want, p.IsArchived, "project %d", id)
	}
	assertArchived(1, true)
	assertArchived(2, true)
	assertArchived(3, true)
	assertArchived(10, false)
	assertArchived(11, false)
	assertArchived(12, true)
	assertArchived(13, true)
	assertArchived(20, true)
	assertArchived(21, true)
	assertArchived(bigRootID+1, true)
	assertArchived(bigRootID+int64(bigChildrenCount/2), true)
	assertArchived(bigRootID+int64(bigChildrenCount), true)

	// idempotent
	require.NoError(t, backfillArchivedDescendants20260819212842(x))
	assertArchived(11, false)
	assertArchived(3, true)
}
