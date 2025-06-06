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
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

type users20220616145228 struct {
	OverdueTasksRemindersTime string `xorm:"varchar(5) not null default '09:00'" json:"-"`
}

func (users20220616145228) TableName() string {
	return "users"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20220616145228",
		Description: "Add overdue task summary time field to users",
		Migrate: func(tx *xorm.Engine) error {
			return tx.Sync2(users20220616145228{})
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
