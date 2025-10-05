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
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"

	"src.techknowlogick.com/xormigrate"
)

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20250903072809",
		Description: "Rename overdue_tasks_reminders_time column to today_tasks_reminders_time",
		Migrate: func(tx *xorm.Engine) error {
			switch tx.Dialect().URI().DBType {
			case schemas.SQLITE:
				_, err := tx.Exec("ALTER TABLE `users` RENAME COLUMN `overdue_tasks_reminders_time` TO `today_tasks_reminders_time`")
				return err
			case schemas.MYSQL:
				_, err := tx.Exec("ALTER TABLE `users` CHANGE `overdue_tasks_reminders_time` `today_tasks_reminders_time` VARCHAR(5) NOT NULL DEFAULT '09:00'")
				return err
			default: // postgres
				_, err := tx.Exec("ALTER TABLE \"users\" RENAME COLUMN \"overdue_tasks_reminders_time\" TO \"today_tasks_reminders_time\"")
				return err
			}
		},
		Rollback: func(tx *xorm.Engine) error { return nil },
	})
}
