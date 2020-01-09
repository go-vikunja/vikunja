// Vikunja is a todo-list application to facilitate your life.
// Copyright 2018-2020 Vikunja and contributors. All rights reserved.
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

package migration

import (
	"github.com/go-xorm/xorm"
	"src.techknowlogick.com/xormigrate"
)

// Used for rollback
type tasksReminderDateMigration20190324205606 struct {
	ReminderUnix int64 `xorm:"int(11) INDEX"`
}

func (tasksReminderDateMigration20190324205606) TableName() string {
	return "tasks"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20190324205606",
		Description: "Remove reminders_unix from tasks",
		Migrate: func(tx *xorm.Engine) error {
			return dropTableColum(tx, "tasks", "reminders_unix")
		},
		Rollback: func(tx *xorm.Engine) error {
			return tx.Sync2(tasksReminderDateMigration20190324205606{})
		},
	})
}
