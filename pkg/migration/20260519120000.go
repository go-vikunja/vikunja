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

	"code.vikunja.io/api/pkg/log"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260519120000",
		Description: "uppercase existing project identifiers",
		Migrate: func(tx *xorm.Engine) error {
			s := tx.NewSession()
			defer s.Close()

			if err := s.Begin(); err != nil {
				return err
			}

			// Postgres/SQLite default to case-sensitive comparisons, so
			// projects like "foo" and "FOO" may coexist today. Uppercasing
			// them blindly would create duplicate identifiers and break the
			// invariant that task identifiers built from them are unique.
			// Detect each colliding group, keep the oldest project's
			// identifier and clear the rest so the operator can re-assign
			// them after the migration runs.
			type collidingGroup struct {
				UpperIdentifier string `xorm:"upper_identifier"`
			}
			var groups []collidingGroup
			err := s.SQL(`
				SELECT UPPER(identifier) AS upper_identifier FROM projects
				WHERE identifier IS NOT NULL AND identifier <> ''
				GROUP BY UPPER(identifier)
				HAVING COUNT(*) > 1
			`).Find(&groups)
			if err != nil {
				_ = s.Rollback()
				return fmt.Errorf("failed to scan for colliding project identifiers: %w", err)
			}

			for _, g := range groups {
				type projectRow struct {
					ID         int64
					Identifier string
				}
				var rows []projectRow
				err := s.SQL(
					"SELECT id, identifier FROM projects WHERE UPPER(identifier) = ? ORDER BY id ASC",
					g.UpperIdentifier,
				).Find(&rows)
				if err != nil {
					_ = s.Rollback()
					return err
				}
				if len(rows) < 2 {
					continue
				}

				kept := rows[0]
				for i := 1; i < len(rows); i++ {
					log.Warningf(
						"Project identifier collision during uppercase migration: clearing identifier %q on project %d (kept %q on project %d). Re-assign a unique identifier after the migration.",
						rows[i].Identifier, rows[i].ID, kept.Identifier, kept.ID,
					)
					if _, err := s.Exec(
						"UPDATE projects SET identifier = ? WHERE id = ?",
						"", rows[i].ID,
					); err != nil {
						_ = s.Rollback()
						return err
					}
				}
			}

			// UPPER() is supported by MySQL, PostgreSQL and SQLite.
			if _, err := s.Exec("UPDATE projects SET identifier = UPPER(identifier) WHERE identifier IS NOT NULL AND identifier <> UPPER(identifier)"); err != nil {
				_ = s.Rollback()
				return err
			}

			return s.Commit()
		},
		Rollback: func(_ *xorm.Engine) error {
			return nil
		},
	})
}
