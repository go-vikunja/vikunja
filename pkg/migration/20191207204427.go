// Copyright 2018-2020 Vikunja and contriubtors. All rights reserved.
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

package migration

import (
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

type list20191207204427 struct {
	Identifier string `xorm:"varchar(10) null" json:"identifier"`
}

func (list20191207204427) TableName() string {
	return "list"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20191207204427",
		Description: "Add task identifier to list",
		Migrate: func(tx *xorm.Engine) error {
			return tx.Sync2(list20191207204427{})
		},
		Rollback: func(tx *xorm.Engine) error {
			return dropTableColum(tx, "list", "indentifier")
		},
	})
}
