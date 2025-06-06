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
	"code.vikunja.io/api/pkg/config"
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

type users20220112211537 struct {
	Timezone string `xorm:"varchar(255) null" json:"-"`
}

func (users20220112211537) TableName() string {
	return "users"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20220112211537",
		Description: "Add time zone setting for users",
		Migrate: func(tx *xorm.Engine) error {
			err := tx.Sync2(users20220112211537{})
			if err != nil {
				return err
			}

			_, err = tx.Update(&users20220112211537{Timezone: config.GetTimeZone().String()})
			return err
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
