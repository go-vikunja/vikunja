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
		ID:          "20260829165019",
		Description: "Add index on tasks (done, due_date) so the overview's due-date sort can walk an index instead of sorting every open task",
		Migrate: func(tx *xorm.Engine) error {
			// Name must match what xorm derives from the index(done_due_date) tag on
			// the Task struct, otherwise a later sync creates a second copy.
			query := "CREATE INDEX IF NOT EXISTS IDX_tasks_done_due_date ON tasks (done, due_date)"
			if db.Type() == schemas.MYSQL {
				// MySQL lacks IF NOT EXISTS on CREATE INDEX.
				query = "CREATE INDEX IDX_tasks_done_due_date ON tasks (done, due_date)"
			}

			_, err := tx.Exec(query)
			if err != nil && !strings.Contains(err.Error(), "Duplicate key name") {
				return fmt.Errorf("could not create index on tasks: %w", err)
			}
			return nil
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
