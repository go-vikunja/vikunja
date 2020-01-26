// Copyright2018-2020 Vikunja and contriubtors. All rights reserved.
//
// This file is part of Vikunja.
//
// Vikunja is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Vikunja is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Vikunja.  If not, see <https://www.gnu.org/licenses/>.

package db

import (
	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"
	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"
	"os"
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
	engine.ShowSQL(os.Getenv("UNIT_TESTS_VERBOSE") == "1")
	engine.SetLogger(xorm.NewSimpleLogger(log.GetLogWriter("database")))
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
