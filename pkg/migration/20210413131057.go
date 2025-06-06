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

type tasks20210413131057 struct {
	RepeatFromCurrentDate bool `xorm:"null" json:"repeat_from_current_date"`
	RepeatMode            int  `xorm:"not null default 0" json:"repeat_mode"`
}

func (tasks20210413131057) TableName() string {
	return "tasks"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20210413131057",
		Description: "Add repeat mode column to tasks",
		Migrate: func(tx *xorm.Engine) error {
			err := tx.Sync2(tasks20210413131057{})
			if err != nil {
				return err
			}

			_, err = tx.
				Where("repeat_from_current_date = ?", true).
				Update(&tasks20210413131057{RepeatMode: 2})
			return err
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
