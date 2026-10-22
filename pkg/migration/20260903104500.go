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

	"code.vikunja.io/api/pkg/db"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/builder"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

type projects20260903104500 struct {
	ParentProjectID *int64 `xorm:"bigint INDEX null"`
}

func (projects20260903104500) TableName() string {
	return "projects"
}

func rootProjectsParentToNull20260903104500(tx *xorm.Engine) error {
	_, err := tx.
		Where(builder.Eq{"parent_project_id": 0}).
		Cols("parent_project_id").
		Nullable("parent_project_id").
		Update(&projects20260903104500{})
	if err != nil {
		return fmt.Errorf("could not set the parent of top-level projects to null: %w", err)
	}

	// MySQL has no partial indexes; it keeps using IDX_projects_parent_project_id.
	if db.Type() == schemas.MYSQL {
		return nil
	}

	// The recursive access CTE joins parent_project_id against a project id, which implies
	// NOT NULL, so the planner can walk this index over the real children only instead of
	// the full one where the root rows dominate.
	_, err = tx.Exec("CREATE INDEX IF NOT EXISTS IDX_projects_parent_project_id_children ON projects (parent_project_id) WHERE parent_project_id IS NOT NULL")
	if err != nil {
		return fmt.Errorf("could not create the partial index on projects.parent_project_id: %w", err)
	}

	return nil
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260903104500",
		Description: "Store the parent of top-level projects as null and index only real children",
		Migrate:     rootProjectsParentToNull20260903104500,
		Rollback: func(_ *xorm.Engine) error {
			return nil
		},
	})
}
