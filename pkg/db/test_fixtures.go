// Copyright 2019 Vikunja and contriubtors. All rights reserved.
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
	"gopkg.in/testfixtures.v2"
)

var fixtures *testfixtures.Context

// InitFixtures initialize test fixtures for a test database
func InitFixtures(helper testfixtures.Helper, dir string) (err error) {
	testfixtures.SkipDatabaseNameCheck(true)
	fixtures, err = testfixtures.NewFolder(x.DB().DB, helper, dir)
	return err
}

// LoadFixtures load fixtures for a test database
func LoadFixtures() error {
	return fixtures.Load()
}
