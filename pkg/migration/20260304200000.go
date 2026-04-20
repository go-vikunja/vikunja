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

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260304200000",
		Description: "Add bucket sort options to project views",
		Migrate: func(tx *xorm.Engine) error {
			columns := map[string]string{
				"bucket_sort_by":    "VARCHAR(50) NULL DEFAULT NULL",
				"bucket_sort_order": "VARCHAR(4) NULL DEFAULT NULL",
			}
			for col, colType := range columns {
				exists, err := columnExists(tx, "project_views", col)
				if err != nil {
					return err
				}
				if exists {
					continue
				}

				if _, err = tx.Exec("ALTER TABLE project_views ADD COLUMN " + col + " " + colType); err != nil {
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
