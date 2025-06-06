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

type reactions20240311173251 struct {
	ID         int64     `xorm:"autoincr not null unique pk" json:"id" param:"reaction"`
	UserID     int64     `xorm:"bigint not null INDEX" json:"-"`
	EntityID   int64     `xorm:"bigint not null INDEX" json:"entity_id"`
	EntityKind int       `xorm:"bigint not null INDEX" json:"entity_kind"`
	Value      string    `xorm:"varchar(20) not null INDEX" json:"value"`
	Created    time.Time `xorm:"created not null" json:"created"`
}

func (reactions20240311173251) TableName() string {
	return "reactions"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20240311173251",
		Description: "Create reactions table",
		Migrate: func(tx *xorm.Engine) error {
			return tx.Sync2(reactions20240311173251{})
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
