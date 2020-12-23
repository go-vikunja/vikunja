// Copyright2018-2020 Vikunja and contriubtors. All rights reserved.
//
// This file is part of Vikunja.
//
// Vikunja is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Vikunja is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with Vikunja.  If not, see <https://www.gnu.org/licenses/>.

package db

import (
	"fmt"
	"os"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"
	"github.com/stretchr/testify/assert"
	"xorm.io/core"
	"xorm.io/xorm"
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

	engine.SetMapper(core.GonicMapper{})
	logger := log.NewXormLogger("DEBUG")
	logger.ShowSQL(os.Getenv("UNIT_TESTS_VERBOSE") == "1")
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
	assert.NoError(t, err, fmt.Sprintf("Failed to assert entries exist in db, error was: %s", err))
	assert.True(t, exists, fmt.Sprintf("Entries %v do not exist in table %s", values, table))
}

// AssertMissing checks and asserts the nonexiste nce of certain entries in the db
func AssertMissing(t *testing.T, table string, values map[string]interface{}) {
	v := make(map[string]interface{})
	exists, err := x.Table(table).Where(values).Exist(&v)
	assert.NoError(t, err, fmt.Sprintf("Failed to assert entries don't exist in db, error was: %s", err))
	assert.False(t, exists, fmt.Sprintf("Entries %v exist in table %s", values, table))
}
