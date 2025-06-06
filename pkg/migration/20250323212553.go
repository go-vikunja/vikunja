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
	"xorm.io/xorm/schemas"
)

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20250323212553",
		Description: "",
		Migrate: func(tx *xorm.Engine) (err error) {
			oldViews := []*projectViews20241118123644{}

			// PostgreSQL needs explicit casting for LIKE operations on JSON fields
			if tx.Dialect().URI().DBType == schemas.POSTGRES {
				err = tx.Where("filter::text not like '{%' AND filter is not null AND filter::text != ''").Find(&oldViews)
			} else {
				err = tx.Where("filter not like '{%' AND filter is not null AND filter != ''").Find(&oldViews)
			}

			if err != nil {
				return
			}

			for _, view := range oldViews {
				newView := &projectViews20241118123644New{
					ID: view.ID,
					Filter: &taskCollection20241118123644{
						Filter: view.Filter,
					},
				}

				_, err = tx.Where("id = ?", view.ID).
					Cols("filter").
					Update(newView)
				if err != nil {
					return
				}
			}

			return
		},

		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
