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

type taskAttachment20191010131430 struct {
	ID     int64 `xorm:"int(11) autoincr not null unique pk" json:"id" param:"attachment"`
	TaskID int64 `xorm:"int(11) not null" json:"task_id" param:"task"`
	FileID int64 `xorm:"int(11) not null" json:"-"`

	CreatedByID int64 `xorm:"int(11) not null" json:"-"`

	Created int64 `xorm:"created"`
}

func (taskAttachment20191010131430) TableName() string {
	return "task_attachments"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20191010131430",
		Description: "Added task attachments table",
		Migrate: func(tx *xorm.Engine) error {
			return tx.Sync2(taskAttachment20191010131430{})
		},
		Rollback: func(tx *xorm.Engine) error {
			return tx.DropTables(taskAttachment20191010131430{})
		},
	})
}
