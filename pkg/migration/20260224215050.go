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
	"strings"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
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

			var indexQuery string
			switch db.Type() {
			case schemas.POSTGRES, schemas.SQLITE:
				indexQuery = "CREATE INDEX IF NOT EXISTS IDX_webhooks_user_id ON webhooks (user_id)"
			case schemas.MYSQL:
				indexQuery = "CREATE INDEX IDX_webhooks_user_id ON webhooks (user_id)"
			}
			if _, err = tx.Exec(indexQuery); err != nil {
				// For MySQL, ignore duplicate key name error (Error 1061)
				if !strings.Contains(err.Error(), "Error 1061") && !strings.Contains(err.Error(), "Duplicate key name") {
					return err
				}
			}

			// Make project_id nullable so user-level webhooks can have NULL project_id.
			// SQLite does not support ALTER COLUMN, but it already allows NULL in bigint columns.
			switch config.DatabaseType.GetString() {
			case "mysql":
				_, err = tx.Exec("ALTER TABLE webhooks MODIFY COLUMN project_id bigint NULL")
			case "postgres":
				_, err = tx.Exec("ALTER TABLE webhooks ALTER COLUMN project_id DROP NOT NULL")
			}
			if err != nil {
				return err
			}

			return nil
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
