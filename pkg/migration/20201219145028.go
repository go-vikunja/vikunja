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
	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

func addAutoIncrementToIDColumns(x *xorm.Session, table string) (err error) {
	_, err = x.Exec("ALTER TABLE " + table + " MODIFY `id` BIGINT AUTO_INCREMENT NOT NULL")
	return
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20201219145028",
		Description: "Re-add auto increment columns",
		Migrate: func(tx *xorm.Engine) error {
			if config.DatabaseType.GetString() != "mysql" {
				log.Debugf("Not migrating since mysql is the only thing that broke in 20201218152741")
				return nil
			}

			columns := []string{
				"totp",
				"users",
				"files",
				"namespaces",
				"team_list",
				"label_task",
				"task_relations",
				"team_namespaces",
				"users_namespace",
				"teams",
				"team_members",
				"task_assignees",
				"users_list",
				"tasks",
				"task_reminders",
				"task_attachments",
				"list",
				"labels",
				"buckets",
				"link_sharing",
				"migration_status",
			}

			s := tx.NewSession()
			for _, table := range columns {
				if err := addAutoIncrementToIDColumns(s, table); err != nil {
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
