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
	"errors"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/modules/keyvalue"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

func TestIsWriteStatement(t *testing.T) {
	tests := map[string]bool{
		"SELECT * FROM projects": false,
		"\n  select 1":           false,
		"BEGIN TRANSACTION":      false,
		"COMMIT":                 false,
		"ROLLBACK":               false,
		"PREPARE":                false,
		"WITH RECURSIVE tree AS (SELECT 1) SELECT * FROM tree":                        false,
		"WITH t AS (SELECT created, updated, deleted FROM tasks) SELECT * FROM t":     false,
		"INSERT INTO projects (id) VALUES (1)":                                        true,
		"  update projects set title = ?":                                             true,
		"DELETE FROM users_projects":                                                  true,
		"CREATE INDEX foo ON bar (baz)":                                               true,
		"EXPLAIN ANALYZE DELETE FROM tasks":                                           true,
		"WITH moved AS (DELETE FROM a RETURNING *) INSERT INTO b SELECT * FROM moved": true,
		"WITH t AS (SELECT update_count, deleted_at FROM tasks) SELECT * FROM t":      false,
		"with t as (select 1) update tasks set done = true":                           true,
		"": true,
	}
	for sqlStr, want := range tests {
		assert.Equal(t, want, isWriteStatement(sqlStr), sqlStr)
	}
}

func TestRememberShared(t *testing.T) {
	config.InitDefaultConfig()
	engine, err := CreateTestEngine()
	require.NoError(t, err)
	require.NoError(t, engine.Sync(&dumpRestoreTest{}))

	newSession := func(t *testing.T) *xorm.Session {
		s := NewSession()
		t.Cleanup(func() { _ = s.Close() })
		return s
	}

	t.Run("fills and serves across sessions", func(t *testing.T) {
		key := "remember-shared-fill"
		InvalidateShared(key)

		calls := 0
		fn := func() (int, error) {
			calls++
			return 42, nil
		}

		v, err := RememberShared(newSession(t), key, time.Minute, fn)
		require.NoError(t, err)
		assert.Equal(t, 42, v)
		assert.Equal(t, 1, calls)

		v, err = RememberShared(newSession(t), key, time.Minute, fn)
		require.NoError(t, err)
		assert.Equal(t, 42, v)
		assert.Equal(t, 1, calls)
	})

	t.Run("an invalidation during the query drops the fill", func(t *testing.T) {
		key := "remember-shared-race"
		InvalidateShared(key)

		calls := 0
		fn := func() (int, error) {
			calls++
			InvalidateShared(key)
			return 1, nil
		}

		v, err := RememberShared(newSession(t), key, time.Minute, fn)
		require.NoError(t, err)
		assert.Equal(t, 1, v)

		_, exists, err := keyvalue.Get(sharedCachePrefix + key)
		require.NoError(t, err)
		assert.False(t, exists, "the fill must not resurrect the invalidated entry")

		_, err = RememberShared(newSession(t), key, time.Minute, fn)
		require.NoError(t, err)
		assert.Equal(t, 2, calls)
	})

	t.Run("an invalidation after the session started drops the fill", func(t *testing.T) {
		key := "remember-shared-snapshot"
		InvalidateShared(key)

		calls := 0
		fn := func() (int, error) {
			calls++
			return 1, nil
		}

		s := newSession(t)
		InvalidateShared(key)

		v, err := RememberShared(s, key, time.Minute, fn)
		require.NoError(t, err)
		assert.Equal(t, 1, v)

		_, exists, err := keyvalue.Get(sharedCachePrefix + key)
		require.NoError(t, err)
		assert.False(t, exists, "a session older than the invalidation must not fill")

		_, err = RememberShared(newSession(t), key, time.Minute, fn)
		require.NoError(t, err)
		assert.Equal(t, 2, calls)
	})

	t.Run("a failing fill caches nothing", func(t *testing.T) {
		key := "remember-shared-error"
		InvalidateShared(key)

		calls := 0
		fn := func() (int, error) {
			calls++
			return 0, errors.New("boom")
		}

		s := newSession(t)
		_, err := RememberShared(s, key, time.Minute, fn)
		require.ErrorContains(t, err, "boom")

		_, exists, err := keyvalue.Get(sharedCachePrefix + key)
		require.NoError(t, err)
		assert.False(t, exists)

		_, err = RememberShared(s, key, time.Minute, fn)
		require.ErrorContains(t, err, "boom")
		assert.Equal(t, 2, calls, "the failed fill must not be memoized on the session")
	})

	t.Run("a session that has written neither serves nor fills", func(t *testing.T) {
		key := "remember-shared-written"
		InvalidateShared(key)

		calls := 0
		fn := func() (int, error) {
			calls++
			return calls, nil
		}

		_, err := RememberShared(newSession(t), key, time.Minute, fn)
		require.NoError(t, err)
		require.Equal(t, 1, calls)

		written := newSession(t)
		_, err = written.Insert(&dumpRestoreTest{Title: "dirty", Created: time.Now()})
		require.NoError(t, err)

		v, err := RememberShared(written, key, time.Minute, fn)
		require.NoError(t, err)
		assert.Equal(t, 2, v)
		assert.Equal(t, 2, calls)

		var shared int
		exists, err := keyvalue.GetWithValue(sharedCachePrefix+key, &shared)
		require.NoError(t, err)
		require.True(t, exists)
		assert.Equal(t, 1, shared, "a written session must not overwrite the shared entry")
	})
}
