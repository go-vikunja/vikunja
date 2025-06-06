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
		ID:          "20241028131622",
		Description: "Add potentially missing indexes",
		Migrate: func(tx *xorm.Engine) error {

			var queries []string

			switch db.Type() {
			case schemas.POSTGRES:
				queries = []string{
					"CREATE INDEX IF NOT EXISTS IDX_projects_owner_id ON projects (owner_id)",
					"CREATE INDEX IF NOT EXISTS IDX_projects_parent_project_id ON projects (parent_project_id)",
				}
			case schemas.MYSQL:
				queries = []string{
					"CREATE INDEX IDX_projects_owner_id ON projects (owner_id)",
					"CREATE INDEX IDX_projects_parent_project_id ON projects (parent_project_id)",
				}
			case schemas.SQLITE:
				queries = []string{
					"CREATE INDEX IF NOT EXISTS IDX_projects_owner_id ON projects (owner_id)",
					"CREATE INDEX IF NOT EXISTS IDX_projects_parent_project_id ON projects (parent_project_id)",
				}
			}

			for _, query := range queries {
				_, err := tx.Exec(query)
				if err != nil && (!strings.Contains(err.Error(), "Error 1061") || !strings.Contains(err.Error(), "Duplicate key name")) {
					return err
				}
			}

			return nil
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
