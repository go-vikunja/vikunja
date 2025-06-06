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

type user20200801183357 struct {
	AvatarProvider string `xorm:"varchar(255) null" json:"-"`
	AvatarFileID   int64  `xorn:"null" json:"-"`
}

func (s user20200801183357) TableName() string {
	return "users"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20200801183357",
		Description: "Add avatar provider setting to user",
		Migrate: func(tx *xorm.Engine) error {
			err := tx.Sync2(user20200801183357{})
			if err != nil {
				return err
			}

			_, err = tx.Cols("avatar_provider").Update(&user20200801183357{AvatarProvider: "initials"})
			return err
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
