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
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/builder"
	"xorm.io/xorm"
	"xorm.io/xorm/names"
	"xorm.io/xorm/schemas"
)

const (
	testLockKey  = int64(0x56494b554e4a41) // "VIKUNJA"
	testLockName = "vikunja-tests"
	// Longer than the 45m go test timeout, and no reliance on how each server treats a negative timeout.
	testLockTimeoutSeconds = 3 * 60 * 60
)

// Holds the lock's connection for the life of the process: returning it to the pool would release the lock.
var testLockConn *sql.Conn

// CreateTestEngine creates an instance of the db engine which lives in memory
func CreateTestEngine() (engine *xorm.Engine, err error) {

	if x != nil {
		return x, nil
	}

	if os.Getenv("VIKUNJA_TESTS_USE_CONFIG") == "1" {
		config.InitConfig()
		engine, err = CreateDBEngine()
		if err != nil {
			return nil, err
		}
		if err = lockSharedTestDatabase(engine); err != nil {
			return nil, err
		}
	} else {
		engine, err = xorm.NewEngine("sqlite3", "file::memory:?cache=shared")
		if err != nil {
			return nil, err
		}
	}

	engine.SetMapper(names.GonicMapper{})
	logger := log.NewXormLogger(config.LogEnabled.GetBool(), config.LogDatabase.GetString(), "DEBUG", config.LogFormat.GetString())
	logger.ShowSQL(os.Getenv("TESTS_VERBOSE") == "1")
	engine.SetLogger(logger)
	engine.SetTZLocation(config.GetTimeZone())
	engine.AddHook(writeInvalidationHook{})
	engine.AddHook(writeCountingHook{})
	x = engine
	return
}

// lockSharedTestDatabase makes test packages that share one server database run one at a time,
// which lets go test run them in parallel and still cache results. The lock dies with the process.
func lockSharedTestDatabase(engine *xorm.Engine) error {
	var query string
	switch engine.Dialect().URI().DBType {
	case schemas.POSTGRES:
		query = fmt.Sprintf("SELECT pg_advisory_lock(%d)", testLockKey)
	case schemas.MYSQL:
		query = fmt.Sprintf("SELECT GET_LOCK('%s', %d)", testLockName, testLockTimeoutSeconds)
	default:
		return nil // sqlite files are per package directory
	}

	ctx := context.Background()
	conn, err := engine.DB().DB.Conn(ctx)
	if err != nil {
		return fmt.Errorf("could not take test database lock: %w", err)
	}

	if engine.Dialect().URI().DBType == schemas.MYSQL {
		// NULL means error and 0 means timeout.
		var acquired sql.NullInt64
		if err := conn.QueryRowContext(ctx, query).Scan(&acquired); err != nil {
			return fmt.Errorf("could not take test database lock: %w", err)
		}
		if !acquired.Valid || acquired.Int64 != 1 {
			return fmt.Errorf("could not take test database lock: GET_LOCK returned %v", acquired)
		}
	} else if _, err := conn.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("could not take test database lock: %w", err)
	}

	testLockConn = conn
	return nil
}

// InitTestFixtures populates the db with all fixtures from the fixtures folder
func InitTestFixtures(tablenames ...string) (err error) {
	// Create all fixtures
	config.InitDefaultConfig()

	// Sync fixtures
	err = InitFixtures(tablenames...)
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

// AssertExists checks and asserts the existence of certain entries in the db
func AssertExists(t *testing.T, table string, values map[string]interface{}, custom bool) {
	var exists bool
	var err error
	v := make(map[string]interface{})
	// Postgres sometimes needs to build raw sql. Because it won't always need to do this and this isn't fun, it's a flag.
	if custom {
		//#nosec
		sql := "SELECT * FROM " + table + " WHERE "
		for col, val := range values {
			sql += col + "=" + fmt.Sprintf("%v", val) + " AND "
		}
		sql = sql[:len(sql)-5]
		exists, err = x.SQL(sql).Get(&v)
	} else {
		exists, err = x.Table(table).Where(condFromValues(values)).Get(&v)
	}
	require.NoError(t, err, "Failed to assert entries exist in db")
	if !exists {

		all := []map[string]interface{}{}
		err = x.Table(table).Find(&all)
		require.NoErrorf(t, err, "Failed to assert entries exist in db, error was: %s", err)
		pretty, err := json.MarshalIndent(all, "", "    ")
		require.NoErrorf(t, err, "Failed to assert entries exist in db, error was: %s", err)

		t.Errorf("Entries %v do not exist in table %s\n\nFound entries instead: %v", values, table, string(pretty))
	}
}

// A nil value asserts the column is NULL — builder.Eq would render col = NULL, which never matches.
func condFromValues(values map[string]interface{}) builder.Cond {
	cond := builder.NewCond()
	for col, val := range values {
		// builder does not quote identifiers, so reserved words like "limit" need it here.
		quoted := x.Quote(col)
		if val == nil {
			cond = cond.And(builder.IsNull{quoted})
			continue
		}
		cond = cond.And(builder.Eq{quoted: val})
	}
	return cond
}

// AssertMissing checks and asserts the nonexistence of certain entries in the db
func AssertMissing(t *testing.T, table string, values map[string]interface{}) {
	all := []map[string]interface{}{}
	err := x.Table(table).Where(condFromValues(values)).Find(&all)
	require.NoErrorf(t, err, "Failed to assert entries don't exist in db, error was: %s", err)

	if len(all) > 0 {
		pretty, err := json.MarshalIndent(all, "", "    ")
		require.NoErrorf(t, err, "Failed to assert entries do not exist in db, error was: %s", err)

		t.Errorf("Entries %v exist in table %s:\n\n%v", values, table, string(pretty))
	}
}

// AssertCount checks if a number of entries exists in the database
func AssertCount(t *testing.T, table string, where builder.Cond, count int64) {
	dbCount, err := x.Table(table).Where(where).Count()
	require.NoErrorf(t, err, "Failed to assert count in db, error was: %s", err)
	assert.Equalf(t, count, dbCount, "Found %d entries instead of expected %d in table %s", dbCount, count, table)
}
