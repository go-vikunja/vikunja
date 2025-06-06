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
	"math"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

type lists20210727204942 struct {
	ID       int64   `xorm:"bigint autoincr not null" json:"id" param:"list"`
	Position float64 `xorm:"double null" json:"position"`
}

func (lists20210727204942) TableName() string {
	return "lists"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20210727204942",
		Description: "Add list position parameter",
		Migrate: func(tx *xorm.Engine) error {
			err := tx.Sync2(lists20210727204942{})
			if err != nil {
				return err
			}

			lists := []*lists20210727204942{}
			err = tx.Find(&lists)
			if err != nil {
				return err
			}

			for _, list := range lists {
				list.Position = float64(list.ID) * math.Pow(2, 16)

				_, err = tx.
					Where("id = ?", list.ID).
					Update(list)
				if err != nil {
					return err
				}
			}

			return nil
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
