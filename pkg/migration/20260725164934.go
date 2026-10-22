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
	"fmt"
	"strings"

	"code.vikunja.io/api/pkg/db"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260725164934",
		Description: "Add index on task_positions (project_view_id, position) so duplicate-position lookups during task creation don't scan the whole view",
		Migrate: func(tx *xorm.Engine) error {
			query := "CREATE INDEX IF NOT EXISTS IDX_task_positions_view_position ON task_positions (project_view_id, position)"
			if db.Type() == schemas.MYSQL {
				// MySQL lacks IF NOT EXISTS on CREATE INDEX.
				query = "CREATE INDEX IDX_task_positions_view_position ON task_positions (project_view_id, position)"
			}

			_, err := tx.Exec(query)
			if err != nil && !strings.Contains(err.Error(), "Duplicate key name") {
				return fmt.Errorf("could not create index on task_positions: %w", err)
			}
			return nil
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
