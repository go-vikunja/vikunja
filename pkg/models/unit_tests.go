// Vikunja is a todo-list application to facilitate your life.
// Copyright 2018-2019 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"code.vikunja.io/api/pkg/config"
	_ "code.vikunja.io/api/pkg/config" // To trigger its init() which initializes the config
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/mail"
	"fmt"
	"github.com/go-xorm/xorm"
	"github.com/stretchr/testify/assert"
	"gopkg.in/testfixtures.v2"
	"os"
	"path/filepath"
	"testing"
)

// SetupTests takes care of seting up the db, fixtures etc.
// This is an extra function to be able to call the fixtures setup from the integration tests.
func SetupTests(pathToRoot string) {
	var err error
	fixturesDir := filepath.Join(pathToRoot, "pkg", "models", "fixtures")
	if err = createTestEngine(fixturesDir); err != nil {
		log.Fatalf("Error creating test engine: %v\n", err)
	}

	// Start the pseudo mail queue
	mail.StartMailDaemon()

	// Create test database
	if err = db.LoadFixtures(); err != nil {
		log.Fatalf("Error preparing test database: %v", err.Error())
	}
}

func createTestEngine(fixturesDir string) error {
	var err error
	var fixturesHelper testfixtures.Helper = &testfixtures.SQLite{}
	// If set, use the config we provided instead of normal
	if os.Getenv("VIKUNJA_TESTS_USE_CONFIG") == "1" {
		x, err = db.CreateTestEngine()
		if err != nil {
			return fmt.Errorf("error getting test engine: %v", err)
		}

		err = initSchema(x)
		if err != nil {
			return err
		}

		if config.DatabaseType.GetString() == "mysql" {
			fixturesHelper = &testfixtures.MySQL{}
		}
	} else {
		x, err = db.CreateTestEngine()
		if err != nil {
			return fmt.Errorf("error getting test engine: %v", err)
		}

		// Sync dat shit
		err = initSchema(x)
		if err != nil {
			return fmt.Errorf("sync database struct error: %v", err)
		}
	}

	return db.InitFixtures(fixturesHelper, fixturesDir)
}

func initSchema(tx *xorm.Engine) error {
	return tx.Sync2(GetTables()...)
}

func initFixtures(t *testing.T) {
	// Init db fixtures
	err := db.LoadFixtures()
	assert.NoError(t, err)
}
