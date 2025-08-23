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

package services

import (
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
)

// InitTests handles the actual bootstrapping of the test env
func InitTests() {
	x, err := db.CreateTestEngine()
	if err != nil {
		log.Fatal(err)
	}

	tables := append(models.GetTables(), user.GetTables()...)
	tables = append(tables, &files.File{})
	err = x.Sync(tables...)
	if err != nil {
		log.Fatal(err)
	}

	err = db.InitTestFixtures()
	if err != nil {
		log.Fatal(err)
	}
}
