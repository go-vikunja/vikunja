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

// Mirrors models.TimeEntry. No partial unique index for the single-active-timer
// rule — MySQL has no filtered indexes; it's enforced in the model instead.
type TimeEntry20260607132257 struct {
	ID        int64      `xorm:"bigint autoincr not null unique pk"`
	UserID    int64      `xorm:"bigint not null INDEX"`
	TaskID    int64      `xorm:"bigint null INDEX"`
	ProjectID int64      `xorm:"bigint null INDEX"`
	StartTime time.Time  `xorm:"not null INDEX"`
	EndTime   *time.Time `xorm:"null"`
	Comment   string     `xorm:"text null"`
	Created   time.Time  `xorm:"created not null"`
	Updated   time.Time  `xorm:"updated not null"`
}

func (TimeEntry20260607132257) TableName() string {
	return "time_entries"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260607132257",
		Description: "Add time_entries table",
		Migrate: func(tx *xorm.Engine) error {
			return tx.Sync(TimeEntry20260607132257{})
		},
		Rollback: func(tx *xorm.Engine) error {
			return tx.DropTables(TimeEntry20260607132257{})
		},
	})
}
