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

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20250624092830",
		Description: "add unique index for task buckets",
		Migrate: func(tx *xorm.Engine) error {

			s := tx.NewSession()
			defer s.Close()

			err := s.Begin()
			if err != nil {
				return err
			}

			// First remove all duplicate entries
			duplicateTaskBuckets := []taskBucket20240406125227{}
			err = s.
				Select("task_id, project_view_id").
				GroupBy("task_id, project_view_id").
				Having("count(*) > 1").
				Find(&duplicateTaskBuckets)
			if err != nil {
				_ = s.Rollback()
				return err
			}

			newTaskBuckets := []taskBucket20240406125227{}
			for _, bucket := range duplicateTaskBuckets {
				newBucket := taskBucket20240406125227{}
				_, err = s.Where("task_id = ? AND project_view_id = ?", bucket.TaskID, bucket.ProjectViewID).
					Get(&newBucket)
				if err != nil {
					_ = s.Rollback()
					return err
				}

				newTaskBuckets = append(newTaskBuckets, newBucket)
			}

			for _, bucket := range duplicateTaskBuckets {
				_, err = s.Where("task_id = ? AND project_view_id = ?", bucket.TaskID, bucket.ProjectViewID).
					Delete(&taskBucket20240406125227{})
				if err != nil {
					_ = s.Rollback()
					return err
				}
			}

			for _, bucket := range newTaskBuckets {
				_, err = s.Insert(&bucket)
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
				query = "CREATE UNIQUE INDEX UQE_task_buckets_task_project_view ON task_buckets (task_id, project_view_id)"
			default:
				query = "CREATE UNIQUE INDEX IF NOT EXISTS UQE_task_buckets_task_project_view ON task_buckets (task_id, project_view_id)"
			}
			_, err = tx.Exec(query)
			return err
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
