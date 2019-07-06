//  Vikunja is a todo-list application to facilitate your life.
//  Copyright 2018 Vikunja and contributors. All rights reserved.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU General Public License as published by
//  the Free Software Foundation, either version 3 of the License, or
//  (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU General Public License for more details.
//
//  You should have received a copy of the GNU General Public License
//  along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"code.vikunja.io/api/pkg/config"
	_ "code.vikunja.io/api/pkg/config" // To trigger its init() which initializes the config
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/mail"
	"fmt"
	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"
	"gopkg.in/testfixtures.v2"
	"os"
	"path/filepath"
	"testing"
)

// MainTest creates the test engine
func MainTest(m *testing.M, pathToRoot string) {
	SetupTests(pathToRoot)
	os.Exit(m.Run())
}

// SetupTests takes care of seting up the db, fixtures etc.
// This is an extra function to be able to call the fixtures setup from the integration tests.
func SetupTests(pathToRoot string) {
	var err error
	fixturesDir := filepath.Join(pathToRoot, "pkg", "models", "fixtures")
	if err = createTestEngine(fixturesDir); err != nil {
		log.Log.Fatalf("Error creating test engine: %v\n", err)
	}

	// Start the pseudo mail queue
	mail.StartMailDaemon()

	// Create test database
	if err = LoadFixtures(); err != nil {
		log.Log.Fatalf("Error preparing test database: %v", err.Error())
	}
}

func createTestEngine(fixturesDir string) error {
	var err error
	var fixturesHelper testfixtures.Helper = &testfixtures.SQLite{}
	// If set, use the config we provided instead of normal
	if os.Getenv("VIKUNJA_TESTS_USE_CONFIG") == "1" {
		config.InitConfig()
		err = SetEngine()
		if err != nil {
			return err
		}

		err = x.Sync2(GetTables()...)
		if err != nil {
			return err
		}

		if config.DatabaseType.GetString() == "mysql" {
			fixturesHelper = &testfixtures.MySQL{}
		}
	} else {
		x, err = xorm.NewEngine("sqlite3", "file::memory:?cache=shared")
		if err != nil {
			return err
		}

		x.SetMapper(core.GonicMapper{})

		// Sync dat shit
		if err := x.Sync2(GetTables()...); err != nil {
			return fmt.Errorf("sync database struct error: %v", err)
		}

		// Show SQL-Queries if necessary
		if os.Getenv("UNIT_TESTS_VERBOSE") == "1" {
			x.ShowSQL(true)
		}
	}

	return InitFixtures(fixturesHelper, fixturesDir)
}
