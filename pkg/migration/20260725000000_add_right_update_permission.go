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
)

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260725000000",
		Description: "Add RightUpdate permission tier and shift Write (1 -> 2) and Admin (2 -> 3) permission values",
		Migrate: func(tx *xorm.Engine) error {
			tables := []string{"users_projects", "team_projects", "link_sharings"}
			for _, table := range tables {
				if _, err := tx.Exec(fmt.Sprintf("UPDATE %s SET permission = 3 WHERE permission = 2", table)); err != nil {
					return fmt.Errorf("could not shift admin permission in %s: %w", table, err)
				}
				if _, err := tx.Exec(fmt.Sprintf("UPDATE %s SET permission = 2 WHERE permission = 1", table)); err != nil {
					return fmt.Errorf("could not shift write permission in %s: %w", table, err)
				}
			}
			return nil
		},
		Rollback: func(tx *xorm.Engine) error {
			tables := []string{"users_projects", "team_projects", "link_sharings"}
			for _, table := range tables {
				if _, err := tx.Exec(fmt.Sprintf("UPDATE %s SET permission = 0 WHERE permission = 1", table)); err != nil {
					return fmt.Errorf("could not rollback update permission in %s: %w", table, err)
				}
				if _, err := tx.Exec(fmt.Sprintf("UPDATE %s SET permission = 1 WHERE permission = 2", table)); err != nil {
					return fmt.Errorf("could not rollback write permission in %s: %w", table, err)
				}
				if _, err := tx.Exec(fmt.Sprintf("UPDATE %s SET permission = 2 WHERE permission = 3", table)); err != nil {
					return fmt.Errorf("could not rollback admin permission in %s: %w", table, err)
				}
			}
			return nil
		},
	})
}
