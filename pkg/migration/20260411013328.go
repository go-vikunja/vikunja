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

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260411013328",
		Description: "add unique constraint on (project_id, index) for tasks",
		Migrate: func(tx *xorm.Engine) error {
			// `index` is MySQL-reserved; Postgres rejects backticks, MySQL
			// rejects double quotes — pick the quote char at runtime.
			idx := `"index"`
			if tx.Dialect().URI().DBType == schemas.MYSQL {
				idx = "`index`"
			}

			// Heal pre-existing duplicates before adding the constraint —
			// setNewTaskIndex's Go-level guard isn't race-safe. Wrapped in a
			// session transaction so a mid-way failure can't leave rows
			// partially re-indexed without the constraint in place.
			s := tx.NewSession()
			defer s.Close()

			if err := s.Begin(); err != nil {
				return err
			}

			type dupRow struct {
				ProjectID int64 `xorm:"project_id"`
				Index     int64 `xorm:"'index'"`
			}
			var dupes []dupRow
			err := s.SQL(`
				SELECT project_id, ` + idx + ` FROM tasks
				GROUP BY project_id, ` + idx + `
				HAVING COUNT(*) > 1
			`).Find(&dupes)
			if err != nil {
				_ = s.Rollback()
				return fmt.Errorf("failed to scan duplicate task indexes: %w", err)
			}

			for _, d := range dupes {
				// Keep rows[0] at its current index, push the rest past max(index).
				type taskRow struct {
					ID int64
				}
				var rows []taskRow
				err := s.SQL(
					"SELECT id FROM tasks WHERE project_id = ? AND "+idx+" = ? ORDER BY id ASC",
					d.ProjectID, d.Index,
				).Find(&rows)
				if err != nil {
					_ = s.Rollback()
					return err
				}
				if len(rows) < 2 {
					continue
				}

				var maxIdx struct {
					M int64 `xorm:"m"`
				}
				_, err = s.SQL(
					"SELECT COALESCE(MAX("+idx+"), 0) AS m FROM tasks WHERE project_id = ?",
					d.ProjectID,
				).Get(&maxIdx)
				if err != nil {
					_ = s.Rollback()
					return err
				}

				for i := 1; i < len(rows); i++ {
					maxIdx.M++
					_, err = s.Exec(
						"UPDATE tasks SET "+idx+" = ? WHERE id = ?",
						maxIdx.M, rows[i].ID,
					)
					if err != nil {
						_ = s.Rollback()
						return err
					}
				}
			}

			if err := s.Commit(); err != nil {
				return err
			}

			// MySQL lacks IF NOT EXISTS on CREATE INDEX.
			var query string
			switch tx.Dialect().URI().DBType {
			case schemas.MYSQL:
				query = "CREATE UNIQUE INDEX UQE_tasks_project_index ON tasks (project_id, " + idx + ")"
			default:
				query = "CREATE UNIQUE INDEX IF NOT EXISTS UQE_tasks_project_index ON tasks (project_id, " + idx + ")"
			}
			_, err = tx.Exec(query)
			return err
		},
		Rollback: func(_ *xorm.Engine) error {
			return nil
		},
	})
}
