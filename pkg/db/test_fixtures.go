// Copyright 2018-2020 Vikunja and contriubtors. All rights reserved.
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
	"github.com/stretchr/testify/assert"
	"gopkg.in/testfixtures.v2"
	"path/filepath"
	"testing"
)

var fixtures *testfixtures.Context

// InitFixtures initialize test fixtures for a test database
func InitFixtures(tablenames ...string) (err error) {

	var helper testfixtures.Helper = &testfixtures.SQLite{}
	if config.DatabaseType.GetString() == "mysql" {
		helper = &testfixtures.MySQL{}
	}
	dir := filepath.Join(config.ServiceRootpath.GetString(), "pkg", "db", "fixtures")

	testfixtures.SkipDatabaseNameCheck(true)

	// If fixture table names are specified, load them
	// Otherwise, load all fixtures
	if len(tablenames) > 0 {
		for i, name := range tablenames {
			tablenames[i] = filepath.Join(dir, name+".yml")
		}
		fixtures, err = testfixtures.NewFiles(x.DB().DB, helper, tablenames...)
	} else {
		fixtures, err = testfixtures.NewFolder(x.DB().DB, helper, dir)
	}
	return err
}

// LoadFixtures load fixtures for a test database
func LoadFixtures() error {
	return fixtures.Load()
}

// LoadAndAssertFixtures loads all fixtures defined before and asserts they are correctly loaded
func LoadAndAssertFixtures(t *testing.T) {
	err := LoadFixtures()
	assert.NoError(t, err)
}
