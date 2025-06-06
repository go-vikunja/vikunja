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

type notifications20210221111953 struct {
	Name string `xorm:"varchar(250) index null" json:"name"`
}

func (notifications20210221111953) TableName() string {
	return "notifications"
}

type notifications20210221111954 struct {
	Name string `xorm:"varchar(250) index not null" json:"name"`
}

func (notifications20210221111954) TableName() string {
	return "notifications"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20210221111953",
		Description: "Add name property to database notifications",
		Migrate: func(tx *xorm.Engine) error {
			err := tx.Sync2(notifications20210221111953{})
			if err != nil {
				return err
			}

			_, err = tx.
				Cols("name").
				Update(&notifications20210221111953{})
			if err != nil {
				return err
			}

			return tx.Sync2(notifications20210221111954{})
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
