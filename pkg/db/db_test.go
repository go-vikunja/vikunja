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
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

func TestGetPostgreSQLConnectionString(t *testing.T) {
	t.Run("with schema", func(t *testing.T) {
		connStr := getPostgreSQLConnectionString("localhost:5432", "vikunja", "secret", "vikunja", "vikunja", "disable", "", "", "", "")
		assert.Equal(t, "postgres://vikunja:secret@localhost:5432/vikunja?sslmode=disable&sslcert=&sslkey=&sslrootcert=&search_path=%22vikunja%22%2Cpublic", connStr)
	})
	t.Run("without schema", func(t *testing.T) {
		connStr := getPostgreSQLConnectionString("localhost:5432", "vikunja", "secret", "vikunja", "", "disable", "", "", "", "")
		assert.Equal(t, "postgres://vikunja:secret@localhost:5432/vikunja?sslmode=disable&sslcert=&sslkey=&sslrootcert=", connStr)
	})
	t.Run("schema needing quoting", func(t *testing.T) {
		connStr := getPostgreSQLConnectionString("localhost:5432", "vikunja", "secret", "vikunja", "MySchema", "disable", "", "", "", "")
		assert.Equal(t, "postgres://vikunja:secret@localhost:5432/vikunja?sslmode=disable&sslcert=&sslkey=&sslrootcert=&search_path=%22MySchema%22%2Cpublic", connStr)
	})
	t.Run("query exec mode", func(t *testing.T) {
		connStr := getPostgreSQLConnectionString("localhost:5432", "vikunja", "secret", "vikunja", "", "disable", "", "", "", "exec")
		assert.Equal(t, "postgres://vikunja:secret@localhost:5432/vikunja?sslmode=disable&sslcert=&sslkey=&sslrootcert=&default_query_exec_mode=exec", connStr)
	})
	t.Run("unix socket", func(t *testing.T) {
		connStr := getPostgreSQLConnectionString("/var/run/postgresql", "vikunja", "secret", "vikunja", "public", "disable", "", "", "", "")
		assert.Equal(t, "postgres://vikunja:secret@:5432/vikunja?sslmode=disable&sslcert=&sslkey=&sslrootcert=&host=/var/run/postgresql&search_path=%22public%22", connStr)
	})
}

func TestSanitizePostgresConnectionError(t *testing.T) {
	err := errors.New(`parse "postgres://vikunja:secret@invalid host/vikunja": invalid IP-literal`)

	sanitized := sanitizePostgresConnectionError(err, "vikunja", "secret")

	assert.NotContains(t, sanitized.Error(), "vikunja:secret")
	assert.NotContains(t, sanitized.Error(), "secret")
	assert.Contains(t, sanitized.Error(), "postgres://<redacted>@")
	assert.Contains(t, sanitized.Error(), "invalid IP-literal")
}

func TestSanitizePostgresConnectionErrorWithShortPassword(t *testing.T) {
	err := errors.New(`parse "postgres://postgres:x@[object Object]:5432/railway": invalid IP-literal`)

	sanitized := sanitizePostgresConnectionError(err, "postgres", "x")

	assert.NotContains(t, sanitized.Error(), "postgres:x@")
	assert.Contains(t, sanitized.Error(), "postgres://<redacted>@")
	assert.Contains(t, sanitized.Error(), "invalid IP-literal")
}

func TestGetSqliteConnectionString(t *testing.T) {
	assert.Equal(t,
		"/data/vikunja.db?_busy_timeout=5000&_journal_mode=WAL",
		getSqliteConnectionString("/data/vikunja.db"),
	)
}

type lockTestRow struct {
	ID       int64   `xorm:"bigint pk"`
	Position float64 `xorm:"double not null"`
}

func (lockTestRow) TableName() string {
	return "lock_test_rows"
}

// Reproduces API-OSS-31: every write request reads before it writes (permission
// checks, then the update), and SQLite fails that promotion with "database is
// locked" as soon as a second request commits in between.
//
// Skipped because it still fails: _txlock=immediate would fix it but deadlocks
// on nested sessions, see getSqliteConnectionString. Unskip once read and write
// sessions are separated.
func TestSqliteConcurrentReadThenWrite(t *testing.T) {
	t.Skip("known failure, see getSqliteConnectionString")

	engine, err := xorm.NewEngine("sqlite3", getSqliteConnectionString(filepath.Join(t.TempDir(), "vikunja.db")))
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = engine.Close()
	})

	require.NoError(t, engine.Sync(&lockTestRow{}))
	_, err = engine.Insert(&lockTestRow{ID: 1, Position: 1})
	require.NoError(t, err)

	const writers = 4
	const updatesPerWriter = 25

	errs := make(chan error, writers*updatesPerWriter)
	var wg sync.WaitGroup
	for range writers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range updatesPerWriter {
				if err := incrementLockTestRow(engine); err != nil {
					errs <- err
				}
			}
		}()
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		require.NoError(t, err)
	}

	row := &lockTestRow{}
	_, err = engine.ID(1).Get(row)
	require.NoError(t, err)
	assert.InDelta(t, float64(1+writers*updatesPerWriter), row.Position, 0.001)
}

func incrementLockTestRow(engine *xorm.Engine) error {
	s := engine.NewSession()
	defer s.Close()

	if err := s.Begin(); err != nil {
		return err
	}

	row := &lockTestRow{}
	if _, err := s.ID(1).Get(row); err != nil {
		return err
	}

	row.Position++
	if _, err := s.ID(1).Cols("position").Update(row); err != nil {
		return err
	}

	return s.Commit()
}
