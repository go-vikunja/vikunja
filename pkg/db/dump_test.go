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

package db

import (
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseDumpTime(t *testing.T) {
	t.Run("rfc3339 utc", func(t *testing.T) {
		parsed, err := parseDumpTime("2026-03-27T11:27:01Z")
		require.NoError(t, err)
		assert.Equal(t, time.Date(2026, 3, 27, 11, 27, 1, 0, time.UTC), parsed.UTC())
	})

	t.Run("rfc3339 with offset", func(t *testing.T) {
		parsed, err := parseDumpTime("2024-02-06T08:32:51+01:00")
		require.NoError(t, err)
		assert.Equal(t, time.Date(2024, 2, 6, 7, 32, 51, 0, time.UTC), parsed.UTC())
	})

	t.Run("rfc3339 with nanoseconds", func(t *testing.T) {
		parsed, err := parseDumpTime("2026-03-27T11:27:01.123456789Z")
		require.NoError(t, err)
		assert.Equal(t, time.Date(2026, 3, 27, 11, 27, 1, 123456789, time.UTC), parsed.UTC())
	})

	t.Run("space separated without timezone", func(t *testing.T) {
		parsed, err := parseDumpTime("2026-03-27 11:27:01")
		require.NoError(t, err)
		assert.Equal(t, time.Date(2026, 3, 27, 11, 27, 1, 0, time.UTC), parsed.UTC())
	})

	t.Run("t separated without timezone", func(t *testing.T) {
		parsed, err := parseDumpTime("2026-03-27T11:27:01")
		require.NoError(t, err)
		assert.Equal(t, time.Date(2026, 3, 27, 11, 27, 1, 0, time.UTC), parsed.UTC())
	})

	t.Run("date only", func(t *testing.T) {
		parsed, err := parseDumpTime("2026-03-27")
		require.NoError(t, err)
		assert.Equal(t, time.Date(2026, 3, 27, 0, 0, 0, 0, time.UTC), parsed.UTC())
	})

	t.Run("invalid", func(t *testing.T) {
		_, err := parseDumpTime("not-a-date")
		require.Error(t, err)
	})
}

type dumpRestoreTest struct {
	ID      int64     `xorm:"autoincr not null unique pk"`
	Title   string    `xorm:"varchar(250)"`
	Created time.Time `xorm:"not null"`
	Updated time.Time `xorm:""`
}

func TestRestore(t *testing.T) {
	config.InitDefaultConfig()
	engine, err := CreateTestEngine()
	require.NoError(t, err)
	require.NoError(t, engine.Sync(&dumpRestoreTest{}))

	t.Run("time values from a dump", func(t *testing.T) {
		err := RestoreAndTruncate("dump_restore_test", []map[string]interface{}{
			{
				"id":      int64(1),
				"title":   "test",
				"created": "2026-03-27T11:27:01Z",
				"updated": "0001-01-01T00:00:00Z",
			},
		})
		require.NoError(t, err)

		row := &dumpRestoreTest{ID: 1}
		has, err := engine.Get(row)
		require.NoError(t, err)
		require.True(t, has)
		assert.Equal(t, time.Date(2026, 3, 27, 11, 27, 1, 0, time.UTC).Unix(), row.Created.Unix())
		assert.True(t, row.Updated.IsZero())
	})

	t.Run("unknown columns are dropped", func(t *testing.T) {
		err := RestoreAndTruncate("dump_restore_test", []map[string]interface{}{
			{
				"id":             int64(2),
				"title":          "test2",
				"created":        "2026-03-27 11:27:01",
				"does_not_exist": "some value",
			},
		})
		require.NoError(t, err)

		row := &dumpRestoreTest{ID: 2}
		has, err := engine.Get(row)
		require.NoError(t, err)
		require.True(t, has)
		assert.Equal(t, "test2", row.Title)
	})

	t.Run("invalid time value", func(t *testing.T) {
		err := RestoreAndTruncate("dump_restore_test", []map[string]interface{}{
			{
				"id":      int64(3),
				"created": "definitely-not-a-date",
			},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "could not parse time value")
	})
}
