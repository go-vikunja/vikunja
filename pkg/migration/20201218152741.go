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

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

func changeColumnToBigint(x *xorm.Session, table, column string, nullable, defaultValue bool) (err error) {
	switch config.DatabaseType.GetString() {
	case "sqlite":
		// Sqlite only has one "INTEGER" type which is at the same time int and int64
	case "mysql":
		var notnull = " NOT NULL"
		if nullable {
			notnull = ""
		}

		var def = ""
		if defaultValue {
			def = " DEFAULT 0"
		}

		_, err := x.Exec("ALTER TABLE " + table + " MODIFY `" + column + "` BIGINT" + def + notnull)
		if err != nil {
			return err
		}
	case "postgres":
		_, err := x.Exec("ALTER TABLE " + table + " ALTER COLUMN `" + column + "` TYPE BIGINT using `" + column + "`::bigint")
		if err != nil {
			return err
		}
	default:
		log.Fatal("Unknown db.")
	}
	return
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20201218152741",
		Description: "Make sure all int64 fields are bigint in the db",
		Migrate: func(tx *xorm.Engine) error {
			// table is the key, columns are the contents
			columns := map[string][]string{
				"totp": {
					"id",
					"user_id",
				},
				"users": {
					"id",
				},
				"files": {
					"id",
					"size", // TODO: This should be a uint64
					"created_by_id",
				},
				"namespaces": {
					"id",
					"owner_id",
				},
				"team_list": {
					"id",
					"team_id",
					"list_id",
					"default:right",
				},
				"saved_filters": {
					"owner_id",
				},
				"label_task": {
					"id",
					"task_id",
					"label_id",
				},
				"task_relations": {
					"id",
					"task_id",
					"other_task_id",
					"created_by_id",
				},
				"team_namespaces": {
					"id",
					"team_id",
					"namespace_id",
					"default:right",
				},
				"users_namespace": {
					"id",
					"user_id",
					"namespace_id",
					"default:right",
				},
				"teams": {
					"id",
					"created_by_id",
				},
				"team_members": {
					"id",
					"team_id",
					"user_id",
				},
				"task_assignees": {
					"id",
					"task_id",
					"user_id",
				},
				"users_list": {
					"id",
					"user_id",
					"list_id",
					"default:right",
				},
				"tasks": {
					"id",
					"created_by_id",
					"list_id",
					"nullable:repeat_after",
					"nullable:priority",
					"default:index",
					"nullable:bucket_id",
				},
				"task_reminders": {
					"id",
					"task_id",
				},
				"task_attachments": {
					"id",
					"task_id",
					"file_id",
					"created_by_id",
				},
				"list": {
					"id",
					"owner_id",
					"namespace_id",
				},
				"labels": {
					"id",
					"created_by_id",
				},
				"buckets": {
					"id",
					"list_id",
					"created_by_id",
				},
				"link_sharing": {
					"id",
					"list_id",
					"default:right",
					"sharing_type",
					"shared_by_id",
				},
				"migration_status": {
					"id",
					"user_id",
				},
			}

			s := tx.NewSession()
			for table, cols := range columns {
				for _, col := range cols {
					var nullable = false
					if strings.HasPrefix(col, "nullable:") {
						col = strings.ReplaceAll(col, "nullable:", "")
						nullable = true
					}

					var defaultValue = false
					if strings.HasPrefix(col, "default:") {
						col = strings.ReplaceAll(col, "default:", "")
						defaultValue = true
					}

					log.Debugf("Migrating %s.%s to bigint", table, col)
					err := changeColumnToBigint(s, table, col, nullable, defaultValue)
					if err != nil {
						_ = s.Rollback()
						return err
					}
				}
			}
			return s.Commit()
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
