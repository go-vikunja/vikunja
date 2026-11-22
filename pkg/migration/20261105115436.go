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
	"code.vikunja.io/api/pkg/log"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/builder"
	"xorm.io/xorm"
)

// Of projects sharing a background file, the oldest keeps it.
var sharedBackgroundCond20261105115436 = builder.And(
	builder.Gt{"background_file_id": 0},
	builder.Or(
		builder.In("background_file_id", builder.Select("file_id").From("task_attachments")),
		builder.In("background_file_id", builder.Select("avatar_file_id").From("users")),
		builder.In("background_file_id", builder.Select("export_file_id").From("users")),
		builder.In("background_file_id", builder.Select("upload_file_id").From("migration_status")),
		builder.Expr("EXISTS (SELECT 1 FROM projects other WHERE other.background_file_id = projects.background_file_id AND other.id < projects.id)"),
	),
)

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20261105115436",
		Description: "Clear project backgrounds that point at files used elsewhere",
		Migrate: func(tx *xorm.Engine) error {
			ids := []int64{}
			if err := tx.Table("projects").Cols("id").Where(sharedBackgroundCond20261105115436).Find(&ids); err != nil {
				return err
			}
			if len(ids) == 0 {
				return nil
			}

			// Usually none; any hit means the background was set to someone else's file.
			log.Warningf("Clearing the background of %d projects whose background file is used elsewhere: %v", len(ids), ids)
			for start := 0; start < len(ids); start += 500 {
				chunk := ids[start:min(start+500, len(ids))]
				_, err := tx.Table("projects").In("id", chunk).Update(map[string]any{
					"background_file_id":   0,
					"background_blur_hash": "",
				})
				if err != nil {
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
