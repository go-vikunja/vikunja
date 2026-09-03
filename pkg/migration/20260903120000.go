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

func swapTasksDueDateIndex20260903120000(tx *xorm.Engine) error {
	// Name must match what xorm derives from the index(project_done_due_date) tag on
	// the Task struct, otherwise a later sync creates a second copy.
	create := "CREATE INDEX IF NOT EXISTS IDX_tasks_project_done_due_date ON tasks (project_id, done, due_date)"
	drop := "DROP INDEX IF EXISTS IDX_tasks_done_due_date"
	if db.Type() == schemas.MYSQL {
		// MySQL supports IF [NOT] EXISTS on neither statement, so both errors are tolerated below.
		create = "CREATE INDEX IDX_tasks_project_done_due_date ON tasks (project_id, done, due_date)"
		drop = "DROP INDEX IDX_tasks_done_due_date ON tasks"
	}

	if _, err := tx.Exec(create); err != nil && !strings.Contains(err.Error(), "Duplicate key name") {
		return fmt.Errorf("could not create index on tasks: %w", err)
	}

	// Every MySQL/MariaDB variant of the "no such index" error (1091) says "check that".
	if _, err := tx.Exec(drop); err != nil && !strings.Contains(err.Error(), "check that") {
		return fmt.Errorf("could not drop index IDX_tasks_done_due_date on tasks: %w", err)
	}

	return nil
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID: "20260903120000",
		Description: "Replace the tasks index on (done, due_date) with (project_id, done, due_date) so the " +
			"overview's due-date sort can seek to the user's projects instead of filtering every open task",
		Migrate: swapTasksDueDateIndex20260903120000,
		Rollback: func(_ *xorm.Engine) error {
			return nil
		},
	})
}
