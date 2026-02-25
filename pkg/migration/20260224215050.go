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
		ID:          "20260224215050",
		Description: "Add user_id to webhooks table and make project_id nullable",
		Migrate: func(tx *xorm.Engine) error {
			exists, err := columnExists(tx, "webhooks", "user_id")
			if err != nil {
				return err
			}
			if !exists {
				if _, err = tx.Exec("ALTER TABLE webhooks ADD COLUMN user_id bigint NULL"); err != nil {
					return err
				}
			}

			if _, err = tx.Exec("CREATE INDEX IF NOT EXISTS IDX_webhooks_user_id ON webhooks (user_id)"); err != nil {
				return err
			}

			return nil
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
