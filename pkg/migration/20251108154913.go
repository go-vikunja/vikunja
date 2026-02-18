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

	"code.vikunja.io/api/pkg/db"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20251108154913",
		Description: "Add index on task_comments.task_id for better query performance",
		Migrate: func(tx *xorm.Engine) error {
			var query string

			switch db.Type() {
			case schemas.POSTGRES, schemas.SQLITE:
				query = "CREATE INDEX IF NOT EXISTS IDX_task_comments_task_id ON task_comments (task_id)"
			case schemas.MYSQL:
				query = "CREATE INDEX IDX_task_comments_task_id ON task_comments (task_id)"
			}

			_, err := tx.Exec(query)
			// For MySQL, ignore duplicate key name error (Error 1061)
			if err != nil && (!strings.Contains(err.Error(), "Error 1061") || !strings.Contains(err.Error(), "Duplicate key name")) {
				return err
			}
			return nil
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
