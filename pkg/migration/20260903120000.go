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

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

// True for the "index already exists" / "no such index" errors of all three
// dialects. xorm's CreateIndexSQL emits no IF NOT EXISTS and MySQL supports it on
// neither statement, so re-running has to be absorbed here rather than in the SQL.
func isIndexPresenceError20260903120000(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "already exists") || // postgres, sqlite
		strings.Contains(msg, "Duplicate key name") || // mysql create
		strings.Contains(msg, "check that") // mysql drop (1091)
}

func swapTasksDueDateIndex20260903120000(tx *xorm.Engine) error {
	// Names must match what xorm derives from the index(project_done_due_date) tag on
	// the Task struct, otherwise a later sync creates a second copy. Postgres folds
	// unquoted identifiers to lowercase, so every name here goes through the quoter.
	quote := tx.Dialect().Quoter().Quote

	newIndex := schemas.NewIndex("project_done_due_date", schemas.IndexType)
	newIndex.AddColumn("project_id", "done", "due_date")

	var drops []string
	switch tx.Dialect().URI().DBType {
	case schemas.MYSQL:
		// Index names are case-insensitive here, so one spelling covers every install.
		drops = []string{"DROP INDEX " + quote("IDX_tasks_done_due_date") + " ON " + quote("tasks")}
	case schemas.POSTGRES:
		// v2.6.0's migration created the index with an unquoted name, which postgres
		// stored lowercased; installs that reached the index through a xorm sync
		// instead have the mixed-case spelling. Either one has to go.
		drops = []string{
			"DROP INDEX IF EXISTS " + quote("IDX_tasks_done_due_date"),
			"DROP INDEX IF EXISTS " + quote("idx_tasks_done_due_date"),
		}
	default:
		drops = []string{"DROP INDEX IF EXISTS " + quote("IDX_tasks_done_due_date")}
	}

	if _, err := tx.Exec(tx.Dialect().CreateIndexSQL("tasks", newIndex)); err != nil && !isIndexPresenceError20260903120000(err) {
		return fmt.Errorf("could not create index on tasks: %w", err)
	}

	for _, drop := range drops {
		if _, err := tx.Exec(drop); err != nil && !isIndexPresenceError20260903120000(err) {
			return fmt.Errorf("could not drop index IDX_tasks_done_due_date on tasks: %w", err)
		}
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
