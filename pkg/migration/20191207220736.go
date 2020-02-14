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

package migration

import (
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

type task20191207220736 struct {
	ID     int64 `xorm:"int(11) autoincr not null unique pk" json:"id" param:"listtask"`
	Index  int64 `xorm:"int(11) not null default 0" json:"index"`
	ListID int64 `xorm:"int(11) INDEX not null" json:"listID" param:"list"`
}

func (task20191207220736) TableName() string {
	return "tasks"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20191207220736",
		Description: "Add task index to tasks",
		Migrate: func(tx *xorm.Engine) error {
			err := tx.Sync2(task20191207220736{})
			if err != nil {
				return err
			}

			// Get all tasks, ordered by list and id
			tasks := []*task20191207220736{}
			err = tx.
				OrderBy("list_id asc, id asc").
				Find(&tasks)
			if err != nil {
				return err
			}

			var currentIndex int64 = 1
			for i, task := range tasks {
				// Reset the current counter if we're encountering a new list
				// We can do this because the list is sorted by list id
				if i > 0 && tasks[i-1].ListID != task.ListID {
					currentIndex = 1
				}

				task.Index = currentIndex
				_, err = tx.Where("id = ?", task.ID).Update(task)
				if err != nil {
					return err
				}

				currentIndex++
			}

			return nil
		},
		Rollback: func(tx *xorm.Engine) error {
			return dropTableColum(tx, "tasks", "index")
		},
	})
}
