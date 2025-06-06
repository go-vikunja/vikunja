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

type notifications20210207192805 struct {
	ID           int64       `xorm:"bigint autoincr not null unique pk" json:"id"`
	NotifiableID int64       `xorm:"bigint not null" json:"-"`
	Notification interface{} `xorm:"json not null" json:"notification"`
	Created      time.Time   `xorm:"created not null" json:"created"`
}

func (notifications20210207192805) TableName() string {
	return "notifications"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20210207192805",
		Description: "Add notifications table",
		Migrate: func(tx *xorm.Engine) error {
			return tx.Sync2(notifications20210207192805{})
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
