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

package migration

import (
	"fmt"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

type usersImportColumns20260930103000 struct {
	ImportID   string `xorm:"varchar(64) null index"`
	JobTitle   string `xorm:"varchar(250) null"`
	Department string `xorm:"varchar(250) null"`
}

func (usersImportColumns20260930103000) TableName() string {
	return "users"
}

func addUsersImportColumns20260930103000(tx *xorm.Engine) error {
	if err := partialSync(tx, usersImportColumns20260930103000{}); err != nil {
		return fmt.Errorf("could not add the user import columns to users: %w", err)
	}
	return nil
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260930103000",
		Description: "Add import_id, job_title and department to users for the scheduled user list import",
		Migrate:     addUsersImportColumns20260930103000,
		Rollback: func(_ *xorm.Engine) error {
			return nil
		},
	})
}
