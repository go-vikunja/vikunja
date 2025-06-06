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
	"math"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

type task20200420215928 struct {
	ID       int64   `xorm:"int(11) autoincr not null unique pk" json:"id" param:"listtask"`
	Position float64 `xorm:"double null" json:"position"`
}

func (s task20200420215928) TableName() string {
	return "tasks"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20200420215928",
		Description: "Add position property to task",
		Migrate: func(tx *xorm.Engine) error {
			err := tx.Sync2(task20200420215928{})
			if err != nil {
				return err
			}

			// Create a position according to their id -> gives a starting position
			tasks := []*task20200420215928{}
			err = tx.Find(&tasks)
			if err != nil {
				return err
			}

			for _, task := range tasks {
				task.Position = float64(task.ID) * math.Pow(2, 16)
				_, err = tx.Where("id = ?", task.ID).Update(task)
				if err != nil {
					return err
				}
			}

			return nil
		},
		Rollback: func(tx *xorm.Engine) error {
			return tx.DropTables(task20200420215928{})
		},
	})
}
