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
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
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
	t.Run("special characters in credentials", func(t *testing.T) {
		for _, host := range []string{"db:5432", "/var/run/postgresql"} {
			connStr := getPostgreSQLConnectionString(host, "vi:k@user", "aa@-Dx!:%40/?#+ &", "vikunja", "", "disable", "", "", "", "")
			cfg, err := pgconn.ParseConfig(connStr)
			require.NoError(t, err)
			assert.Equal(t, "vi:k@user", cfg.User)
			assert.Equal(t, "aa@-Dx!:%40/?#+ &", cfg.Password)
			assert.Equal(t, "vikunja", cfg.Database)
		}
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

func TestSQLiteWriteReservation(t *testing.T) {
	e, err := xorm.NewEngine("sqlite3", filepath.Join(t.TempDir(), "writers.db")+"?_journal_mode=WAL&_busy_timeout=1000")
	require.NoError(t, err)
	defer e.Close()
	_, err = e.Exec("CREATE TABLE writes (id INTEGER PRIMARY KEY AUTOINCREMENT, value INTEGER)")
	require.NoError(t, err)
	_, err = e.Exec("INSERT INTO writes (value) VALUES (0)")
	require.NoError(t, err)
	foreground := e.NewSession()
	defer foreground.Close()
	require.NoError(t, foreground.Begin())
	require.NoError(t, reserveWrite(foreground))
	_, err = foreground.QueryString("SELECT value FROM writes")
	require.NoError(t, err)
	done := make(chan error, 1)
	go func() {
		_, err := e.Exec("UPDATE writes SET value = value + 1")
		done <- err
	}()
	select {
	case err := <-done:
		require.NoError(t, err)
		t.Fatal("background writer committed during the foreground transaction")
	case <-time.After(50 * time.Millisecond):
	}
	_, err = foreground.Exec("UPDATE writes SET value = value + 1")
	require.NoError(t, err)
	require.NoError(t, foreground.Commit())
	require.NoError(t, <-done)
	rows, err := e.QueryString("SELECT value FROM writes")
	require.NoError(t, err)
	require.Equal(t, "2", rows[0]["value"])
}
