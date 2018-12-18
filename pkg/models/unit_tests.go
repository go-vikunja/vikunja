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
	_ "code.vikunja.io/api/pkg/config" // To trigger its init() which initializes the config
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/mail"
	"fmt"
	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"
	"github.com/spf13/viper"
	"gopkg.in/testfixtures.v2"
	"os"
	"path/filepath"
	"testing"
)

// IsTesting is set to true when we're running tests.
// We don't have a good solution to test email sending yet, so we disable email sending when testing
var IsTesting bool

// MainTest creates the test engine
func MainTest(m *testing.M, pathToRoot string) {
	var err error
	fixturesDir := filepath.Join(pathToRoot, "pkg", "models", "fixtures")
	if err = createTestEngine(fixturesDir); err != nil {
		log.Log.Fatalf("Error creating test engine: %v\n", err)
	}

	IsTesting = true

	// Start the pseudo mail queue
	mail.StartMailDaemon()

	// Create test database
	if err = PrepareTestDatabase(); err != nil {
		log.Log.Fatalf("Error preparing test database: %v", err.Error())
	}

	os.Exit(m.Run())
}

func createTestEngine(fixturesDir string) error {
	var err error
	// If set, use the config we provided instead of normal
	if os.Getenv("VIKUNJA_TESTS_USE_CONFIG") == "1" {
		err = SetEngine()
		if err != nil {
			return err
		}
	} else {
		x, err = xorm.NewEngine("sqlite3", "file::memory:?cache=shared")
		if err != nil {
			return err
		}

		x.SetMapper(core.GonicMapper{})

		// Sync dat shit
		if err := x.StoreEngine("InnoDB").Sync2(tables...); err != nil {
			return fmt.Errorf("sync database struct error: %v", err)
		}

		// Show SQL-Queries if necessary
		if os.Getenv("UNIT_TESTS_VERBOSE") == "1" {
			x.ShowSQL(true)
		}
	}

	var fixturesHelper testfixtures.Helper
	if viper.GetString("database.type") == "mysql" {
		fixturesHelper = &testfixtures.MySQL{}
	} else {
		fixturesHelper = &testfixtures.SQLite{}
	}

	return InitFixtures(fixturesHelper, fixturesDir)
}

// PrepareTestDatabase load test fixtures into test database
func PrepareTestDatabase() error {
	return LoadFixtures()
}
