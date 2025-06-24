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
		ID:          "20250624092830",
		Description: "add unique index for task buckets",
		Migrate: func(tx *xorm.Engine) error {
			var query string
			switch tx.Dialect().URI().DBType {
			case schemas.MYSQL:
				query = "CREATE UNIQUE INDEX UQE_task_buckets_task_project_view ON task_buckets (task_id, project_view_id)"
			default:
				query = "CREATE UNIQUE INDEX IF NOT EXISTS UQE_task_buckets_task_project_view ON task_buckets (task_id, project_view_id)"
			}
			_, err := tx.Exec(query)
			return err
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
