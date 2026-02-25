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

type Session20260224113347 struct {
	ID            string    `xorm:"varchar(36) not null unique pk"`
	UserID        int64     `xorm:"bigint not null index"`
	TokenHash     string    `xorm:"varchar(64) not null unique index"`
	DeviceInfo    string    `xorm:"text"`
	IPAddress     string    `xorm:"varchar(100)"`
	IsLongSession bool      `xorm:"not null default false"`
	LastActive    time.Time `xorm:"not null"`
	Created       time.Time `xorm:"created not null"`
}

func (Session20260224113347) TableName() string {
	return "sessions"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260224113347",
		Description: "Add sessions table",
		Migrate: func(tx *xorm.Engine) error {
			return tx.Sync(Session20260224113347{})
		},
		Rollback: func(tx *xorm.Engine) error {
			return tx.DropTables(Session20260224113347{})
		},
	})
}
