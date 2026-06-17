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
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

type taskPosition20260617153629 struct {
	TaskID        int64   `xorm:"bigint not null index"`
	ProjectViewID int64   `xorm:"bigint not null index"`
	Position      float64 `xorm:"double not null"`
}

func (taskPosition20260617153629) TableName() string {
	return "task_positions"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260617153629",
		Description: "deduplicate task positions and add a unique index on task_id + project_view_id",
		Migrate: func(tx *xorm.Engine) error {

			s := tx.NewSession()
			defer s.Close()

			err := s.Begin()
			if err != nil {
				return err
			}

			// First remove all duplicate entries. A task may only ever have a
			// single position per view; rapid task creation could race and
			// insert more than one row before this constraint existed.
			duplicates := []taskPosition20260617153629{}
			err = s.
				Select("task_id, project_view_id").
				GroupBy("task_id, project_view_id").
				Having("count(*) > 1").
				Find(&duplicates)
			if err != nil {
				_ = s.Rollback()
				return err
			}

			// Keep the lowest position of each group so the result is
			// deterministic across databases.
			kept := []taskPosition20260617153629{}
			for _, dup := range duplicates {
				row := taskPosition20260617153629{}
				has, err := s.
					Where("task_id = ? AND project_view_id = ?", dup.TaskID, dup.ProjectViewID).
					OrderBy("position ASC").
					Get(&row)
				if err != nil {
					_ = s.Rollback()
					return err
				}
				if !has {
					continue
				}
				kept = append(kept, row)
			}

			for _, dup := range duplicates {
				_, err = s.
					Where("task_id = ? AND project_view_id = ?", dup.TaskID, dup.ProjectViewID).
					Delete(&taskPosition20260617153629{})
				if err != nil {
					_ = s.Rollback()
					return err
				}
			}

			for _, position := range kept {
				_, err = s.Insert(&position)
				if err != nil {
					_ = s.Rollback()
					return err
				}
			}

			err = s.Commit()
			if err != nil {
				return err
			}

			// Then create the unique index
			var query string
			switch tx.Dialect().URI().DBType {
			case schemas.MYSQL:
				query = "CREATE UNIQUE INDEX UQE_task_positions_task_project_view ON task_positions (task_id, project_view_id)"
			default:
				query = "CREATE UNIQUE INDEX IF NOT EXISTS UQE_task_positions_task_project_view ON task_positions (task_id, project_view_id)"
			}
			_, err = tx.Exec(query)
			return err
		},
		Rollback: func(_ *xorm.Engine) error {
			return nil
		},
	})
}
