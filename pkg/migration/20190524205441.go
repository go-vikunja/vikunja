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

type taskReminder20190524205441 struct {
	ID           int64 `xorm:"int(11) autoincr not null unique pk"`
	TaskID       int64 `xorm:"int(11) not null INDEX"`
	ReminderUnix int64 `xorm:"int(11) not null INDEX"`
	Created      int64 `xorm:"created not null"`
}

// TableName returns a pretty table name
func (taskReminder20190524205441) TableName() string {
	return "task_reminders"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20190524205441",
		Description: "Add extra table for reminders",
		Migrate: func(tx *xorm.Engine) error {
			err := tx.Sync2(taskReminder20190524205441{})
			if err != nil {
				return err
			}

			// get all current reminders and put them into the new table
			var allTasks []*listTask20190511202210
			err = tx.Find(&allTasks)
			if err != nil {
				return err
			}

			reminders := []*taskReminder20190524205441{}
			for _, t := range allTasks {
				for _, reminder := range t.RemindersUnix {
					reminders = append(reminders, &taskReminder20190524205441{TaskID: t.ID, ReminderUnix: reminder})
				}
			}
			_, err = tx.Insert(reminders)
			if err != nil {
				return err
			}

			return dropTableColum(tx, "tasks", "reminders_unix")
		},
		Rollback: func(tx *xorm.Engine) error {
			return tx.DropTables(taskReminder20190524205441{})
		},
	})
}
