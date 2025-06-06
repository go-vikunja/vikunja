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

type users20210713213622 struct {
	ID       int64 `xorm:"bigint autoincr not null" json:"id"`
	IsActive bool  `xorm:"null" json:"-"`
	Status   int   `xorm:"default 0" json:"-"`
}

func (users20210713213622) TableName() string {
	return "users"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20210713213622",
		Description: "Add users status instead of is_active",
		Migrate: func(tx *xorm.Engine) error {
			err := tx.Sync2(users20210713213622{})
			if err != nil {
				return err
			}

			users := []*users20210713213622{}
			err = tx.Find(&users)
			if err != nil {
				return err
			}

			for _, user := range users {
				if user.IsActive {
					continue
				}

				user.Status = 1 // 1 is "email confirmation required" - as that's the only way is_active was used before we'll use that
				_, err := tx.
					Where("id = ?", user.ID).
					Cols("status").
					Update(user)
				if err != nil {
					return err
				}
			}

			return dropTableColum(tx, "users", "is_active")
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
