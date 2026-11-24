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

package handler

import (
	"context"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/web"

	"github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

// readContextObject is a mock CObject used to test read pipelines (DoReadOne / DoReadAll).
// By injecting a custom read closure, tests can observe context propagation and session state
// while a database session is active.
type readContextObject struct {
	CObject
	read func(*xorm.Session) error
}

func (o *readContextObject) CanRead(*xorm.Session, web.Auth) (bool, int, error) {
	return true, 1, nil
}

func (o *readContextObject) ReadOne(s *xorm.Session, _ web.Auth) error {
	return o.read(s)
}

func (o *readContextObject) ReadAll(s *xorm.Session, _ web.Auth, _ string, _, _ int) (any, int, int64, error) {
	return nil, 0, 0, o.read(s)
}

// runReadOperation dispatches to DoReadOne or DoReadAll based on param readOne.
func runReadOperation(ctx context.Context, obj CObject, readOne bool) error {
	if readOne {
		_, err := DoReadOne(ctx, obj, nil)
		return err
	}
	_, _, _, err := DoReadAll(ctx, obj, nil, "", 1, 10)
	return err
}

// TestDoReadContextCancellation verifies that DoReadOne and DoReadAll propagate the incoming
// context.Context to the underlying database session via db.SetSessionContext:
// 1. Canceling the request context mid-query immediately aborts the SQL execution with context.Canceled.
// 2. After the query is interrupted, the connection is cleanly returned to the pool for subsequent queries.
func TestDoReadContextCancellation(t *testing.T) {
	cases := []struct {
		name    string
		readOne bool
	}{
		{name: "DoReadAll", readOne: false},
		{name: "DoReadOne", readOne: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			config.InitDefaultConfig()
			engine, err := db.CreateTestEngine()
			require.NoError(t, err)

			if db.Type() != schemas.SQLITE {
				t.Skip("query cancellation probe relies on a custom SQLite function")
			}

			// Restrict the pool to a single connection so the registered SQLite function
			// and subsequent queries operate on the exact same physical connection.
			engine.SetMaxOpenConns(1)
			t.Cleanup(func() { engine.SetMaxOpenConns(0) })

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Register a custom SQLite function that cancels ctx while the query is actively running.
			conn, err := engine.DB().Conn(context.Background())
			require.NoError(t, err)
			err = conn.Raw(func(driverConn any) error {
				return driverConn.(*sqlite3.SQLiteConn).RegisterFunc("cancel_read_request", func() int {
					cancel()
					return 1
				}, false)
			})
			require.NoError(t, conn.Close())
			require.NoError(t, err)

			// Run a recursive CTE that invokes cancel_read_request() on the 100th row to abort mid-flight.
			obj := &readContextObject{
				read: func(s *xorm.Session) error {
					_, err := s.QueryString(`WITH RECURSIVE numbers(n) AS (
						SELECT 1 UNION ALL SELECT n + 1 FROM numbers WHERE n < 1000000
					) SELECT sum(CASE WHEN n = 100 THEN cancel_read_request() ELSE n END) FROM numbers`)
					return err
				},
			}

			err = runReadOperation(ctx, obj, tc.readOne)
			require.ErrorIs(t, err, context.Canceled, "underlying query should abort immediately when context is canceled")

			// Ensure the connection pool remains healthy and usable after the interrupted query rolls back.
			_, err = engine.QueryString("SELECT 1")
			require.NoError(t, err, "database connection should remain usable in the pool after query cancellation")
		})
	}
}

// TestReadSessionCacheRegression ensures that attaching a context via db.SetSessionContext(ctx, s)
// preserves the session's identity so that session-scoped db.Remember caching continues to hit
// on repeated reads and properly invalidates after a write within the same session.
func TestReadSessionCacheRegression(t *testing.T) {
	config.InitDefaultConfig()
	engine, err := db.CreateTestEngine()
	require.NoError(t, err)

	_, err = engine.Exec("CREATE TABLE read_context_cache (value INTEGER)")
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err := engine.Exec("DROP TABLE read_context_cache")
		require.NoError(t, err)
	})

	_, err = engine.Exec("INSERT INTO read_context_cache VALUES (1)")
	require.NoError(t, err)

	cases := []struct {
		name    string
		readOne bool
	}{
		{name: "DoReadAll", readOne: false},
		{name: "DoReadOne", readOne: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			obj := &readContextObject{
				read: func(s *xorm.Session) error {
					calls := 0
					fetch := func() (string, error) {
						calls++
						rows, err := s.QueryString("SELECT value FROM read_context_cache")
						if err != nil {
							return "", err
						}
						return rows[0]["value"], nil
					}

					// 1. Initial read should miss the cache and execute the underlying query.
					first, err := db.Remember(s, "read-context-value", fetch)
					require.NoError(t, err)

					// 2. Subsequent read with the same key in the same session should hit the cache.
					again, err := db.Remember(s, "read-context-value", fetch)
					require.NoError(t, err)
					require.Equal(t, first, again)
					require.Equal(t, 1, calls, "repeated read within the same session should hit db.Remember cache")

					// 3. Modifying data in the same session should invalidate the cache and trigger a fresh query.
					_, err = s.Exec("UPDATE read_context_cache SET value = value + 1")
					require.NoError(t, err)

					after, err := db.Remember(s, "read-context-value", fetch)
					require.NoError(t, err)
					require.NotEqual(t, first, after)
					require.Equal(t, 2, calls, "write operation should invalidate session cache and re-fetch")

					return nil
				},
			}

			if tc.readOne {
				permission, err := DoReadOne(context.Background(), obj, nil)
				require.NoError(t, err)
				require.Equal(t, 1, permission)
			} else {
				_, _, _, err := DoReadAll(context.Background(), obj, nil, "", 1, 10)
				require.NoError(t, err)
			}
		})
	}
}
