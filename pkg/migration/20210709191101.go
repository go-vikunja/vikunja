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

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20210709191101",
		Description: "Make the task title type TEXT instead of varchar(250)",
		Migrate: func(tx *xorm.Engine) error {
			switch config.DatabaseType.GetString() {
			case "sqlite":
				// Sqlite only has a "TEXT" type so we don't need to modify it
			case "mysql":
				_, err := tx.Exec("alter table tasks modify title text not null")
				if err != nil {
					return err
				}
			case "postgres":
				_, err := tx.Exec("alter table tasks alter column title type text using title::text")
				if err != nil {
					return err
				}
			default:
				log.Fatal("Unknown db.")
			}
			return nil
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
