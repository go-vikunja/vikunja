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
	"time"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

// taskCaldavDeletion20260427120000 tracks tasks that have been deleted so that
// the CalDAV sync-collection REPORT handler can include them as 404 responses
// for clients that already have a sync token referencing the time before deletion.
type taskCaldavDeletion20260427120000 struct {
	ID        int64     `xorm:"bigint autoincr not null pk"`
	UID       string    `xorm:"varchar(250) not null index"`
	ProjectID int64     `xorm:"bigint not null index"`
	DeletedAt time.Time `xorm:"timestampz not null index"`
}

func (taskCaldavDeletion20260427120000) TableName() string {
	return "task_caldav_deletions"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260427120000",
		Description: "add task_caldav_deletions table for CalDAV sync-collection support",
		Migrate: func(tx *xorm.Engine) error {
			return tx.Sync(taskCaldavDeletion20260427120000{})
		},
		Rollback: func(tx *xorm.Engine) error {
			return tx.DropTables(taskCaldavDeletion20260427120000{})
		},
	})
}
