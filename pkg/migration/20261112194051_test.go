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
	"time"

	"code.vikunja.io/api/pkg/db"
	"github.com/stretchr/testify/require"
)

func TestRemoveCompletedJobs20261112194051(t *testing.T) {
	engine, err := db.CreateTestEngine()
	require.NoError(t, err)
	table := jobs20261112124932{}
	require.NoError(t, engine.DropTables(table))
	t.Cleanup(func() { require.NoError(t, engine.DropTables(table)) })
	require.NoError(t, engine.Sync2(table))
	now := time.Now().UTC().Truncate(time.Second)
	rows := make([]jobs20261112124932, 1001)
	for i := range rows {
		rows[i] = jobs20261112124932{Queue: "test", Kind: "done", Version: 1, Payload: `{}`, State: "completed", RunAt: now}
	}
	for offset := 0; offset < len(rows); offset += 50 {
		batch := rows[offset:min(offset+50, len(rows))]
		_, err = engine.Insert(&batch)
		require.NoError(t, err)
	}
	for _, state := range []string{"available", "running", "failed"} {
		key := state
		_, err = engine.Insert(&jobs20261112124932{Queue: "test", Kind: "kept", Version: 1, Payload: `{}`, State: state, RunAt: now, UniqueKey: &key})
		require.NoError(t, err)
	}
	require.NoError(t, removeCompletedJobs20261112194051(engine))
	require.NoError(t, removeCompletedJobs20261112194051(engine))
	var remaining []jobs20261112124932
	require.NoError(t, engine.OrderBy("id").Find(&remaining))
	require.Len(t, remaining, 3)
	for _, row := range remaining {
		require.NotEqual(t, "completed", row.State)
		require.NotNil(t, row.UniqueKey)
		require.Equal(t, row.State, *row.UniqueKey)
	}
	duplicate := remaining[0]
	duplicate.ID = 0
	_, err = engine.Insert(&duplicate)
	require.Error(t, err, "live uniqueness must remain enforced after purging history")
}
