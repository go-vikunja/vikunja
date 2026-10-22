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
	"strconv"
	"testing"

	"code.vikunja.io/api/pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

type memoTestRow struct {
	ID    int64  `xorm:"autoincr not null unique pk"`
	Title string `xorm:"varchar(250)"`
}

func TestRemember(t *testing.T) {
	config.InitDefaultConfig()
	engine, err := CreateTestEngine()
	require.NoError(t, err)
	// DDL through the engine bypasses the session write hook.
	require.NoError(t, engine.Sync(&memoTestRow{}))
	t.Cleanup(func() { _ = engine.DropTables(&memoTestRow{}) })

	s := NewSession()
	defer s.Close()

	calls := 0
	fetch := func() (int, error) {
		calls++
		return 42, nil
	}

	first, err := Remember(s, "remember-test", fetch)
	require.NoError(t, err)
	second, err := Remember(s, "remember-test", fetch)
	require.NoError(t, err)
	assert.Equal(t, 42, first)
	assert.Equal(t, 42, second)
	assert.Equal(t, 1, calls)

	_, err = s.Insert(&memoTestRow{Title: "x"})
	require.NoError(t, err)

	third, err := Remember(s, "remember-test", fetch)
	require.NoError(t, err)
	assert.Equal(t, 42, third)
	assert.Equal(t, 2, calls)
}

func TestRememberEach(t *testing.T) {
	config.InitDefaultConfig()
	_, err := CreateTestEngine()
	require.NoError(t, err)

	s := NewSession()
	defer s.Close()

	var fetched [][]int64
	key := func(id int64) string { return "remember-each-" + strconv.FormatInt(id, 10) }
	fetch := func(missing []int64) (map[int64]string, error) {
		fetched = append(fetched, append([]int64(nil), missing...))
		values := map[int64]string{}
		for _, id := range missing {
			if id == 3 {
				continue
			}
			values[id] = "v" + strconv.FormatInt(id, 10)
		}
		return values, nil
	}

	first, err := RememberEach(s, []int64{1, 2}, key, fetch)
	require.NoError(t, err)
	assert.Equal(t, map[int64]string{1: "v1", 2: "v2"}, first)

	second, err := RememberEach(s, []int64{2, 3}, key, fetch)
	require.NoError(t, err)
	assert.Equal(t, map[int64]string{2: "v2"}, second)

	third, err := RememberEach(s, []int64{3}, key, fetch)
	require.NoError(t, err)
	assert.Empty(t, third)

	deduped, err := RememberEach(s, []int64{4, 4, 5}, key, fetch)
	require.NoError(t, err)
	assert.Equal(t, map[int64]string{4: "v4", 5: "v5"}, deduped)

	assert.Equal(t, [][]int64{{1, 2}, {3}, {3}, {4, 5}}, fetched)
}

func TestRememberFetchError(t *testing.T) {
	config.InitDefaultConfig()
	_, err := CreateTestEngine()
	require.NoError(t, err)

	s := NewSession()
	defer s.Close()

	calls := 0
	fetchErr := errors.New("fetch failed")
	fetch := func() (int, error) {
		calls++
		return 42, fetchErr
	}

	first, err := Remember(s, "remember-error", fetch)
	require.ErrorIs(t, err, fetchErr)
	assert.Equal(t, 0, first)

	_, err = Remember(s, "remember-error", fetch)
	require.ErrorIs(t, err, fetchErr)
	assert.Equal(t, 2, calls, "a failed fetch stores nothing")
}

func TestRememberWithoutMemo(t *testing.T) {
	config.InitDefaultConfig()
	engine, err := CreateTestEngine()
	require.NoError(t, err)

	// Only NewSession attaches a memo, so a session straight from the engine has none.
	s := engine.NewSession()
	defer s.Close()

	calls := 0
	fetch := func() (int, error) {
		calls++
		return 42, nil
	}

	_, err = Remember(s, "remember-no-memo", fetch)
	require.NoError(t, err)
	_, err = Remember(s, "remember-no-memo", fetch)
	require.NoError(t, err)
	assert.Equal(t, 2, calls)

	var fetched [][]int64
	key := func(id int64) string { return "remember-each-no-memo-" + strconv.FormatInt(id, 10) }
	values, err := RememberEach(s, []int64{4, 4, 5}, key, func(missing []int64) (map[int64]string, error) {
		fetched = append(fetched, append([]int64(nil), missing...))
		loaded := map[int64]string{}
		for _, id := range missing {
			loaded[id] = "v" + strconv.FormatInt(id, 10)
		}
		return loaded, nil
	})
	require.NoError(t, err)
	assert.Equal(t, map[int64]string{4: "v4", 5: "v5"}, values)
	assert.Equal(t, [][]int64{{4, 5}}, fetched)
}
