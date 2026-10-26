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

// tasks20261023190000 is the one column this migration adds: the due date's intended local time
// of day, in seconds after local midnight, which a completion remembers when a spring-forward gap
// shifted the due date (NULL when it did not).
type tasks20261023190000 struct {
	IntendedDueTime *int `xorm:"int null 'intended_due_time'"`
}

func (tasks20261023190000) TableName() string {
	return "tasks"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20261023190000",
		Description: "Add intended_due_time column to tasks",
		Migrate: func(tx *xorm.Engine) error {
			return partialSync(tx, tasks20261023190000{})
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
