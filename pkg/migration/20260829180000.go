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

type subscriptions20260829180000 struct {
	Muted bool `xorm:"not null default false"`
}

func (subscriptions20260829180000) TableName() string {
	return "subscriptions"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260829180000",
		Description: "Add muted column to subscriptions so users can opt out of an inherited subscription",
		Migrate: func(tx *xorm.Engine) error {
			return partialSync(tx, subscriptions20260829180000{})
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
