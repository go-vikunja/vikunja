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
)

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
	x = engine
	return
}

// InitTestFixtures populates the db with all fixtures from the fixtures folder
func InitTestFixtures(tablenames ...string) (err error) {
	// Create all fixtures
	config.InitDefaultConfig()
	// We need to set the root path even if we're not using the config, otherwise fixtures are not loaded correctly
	config.ServiceRootpath.Set(os.Getenv("VIKUNJA_SERVICE_ROOTPATH"))

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
		exists, err = x.Table(table).Where(values).Get(&v)
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

// AssertMissing checks and asserts the nonexistence of certain entries in the db
func AssertMissing(t *testing.T, table string, values map[string]interface{}) {
	all := []map[string]interface{}{}
	err := x.Table(table).Where(values).Find(&all)
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
