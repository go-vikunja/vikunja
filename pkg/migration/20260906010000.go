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

const parentNotNullIndex20260906010000 = "IDX_projects_parent_project_id_not_null"

// MySQL has no partial indexes; it keeps using the full IDX_projects_parent_project_id.
func createParentNotNullIndex20260906010000(tx *xorm.Engine) error {
	if tx.Dialect().URI().DBType == schemas.MYSQL {
		return nil
	}
	quote := tx.Dialect().Quoter().Quote
	_, err := tx.Exec("CREATE INDEX IF NOT EXISTS " + quote(parentNotNullIndex20260906010000) +
		" ON " + quote("projects") + " (" + quote("parent_project_id") + ")" +
		" WHERE " + quote("parent_project_id") + " IS NOT NULL")
	if err != nil {
		return fmt.Errorf("could not create index %s on projects: %w", parentNotNullIndex20260906010000, err)
	}
	return nil
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID: "20260906010000",
		Description: "Partial index on projects.parent_project_id for child projects only, so the recursive " +
			"access resolution scans the few children instead of hashing every project",
		Migrate: createParentNotNullIndex20260906010000,
		Rollback: func(tx *xorm.Engine) error {
			if tx.Dialect().URI().DBType == schemas.MYSQL {
				return nil
			}
			_, err := tx.Exec("DROP INDEX IF EXISTS " + tx.Dialect().Quoter().Quote(parentNotNullIndex20260906010000))
			return err
		},
	})
}
