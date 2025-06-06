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
	"time"

	"code.vikunja.io/api/pkg/config"
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

type migrationStatus20231108231513 struct {
	ID         int64     `xorm:"bigint autoincr not null unique pk" json:"id"`
	Created    time.Time `xorm:"created not null 'created'" json:"time"`
	StartedAt  time.Time `xorm:"null" json:"started_at"`
	FinishedAt time.Time `xorm:"null" json:"finished_at"`
}

func (migrationStatus20231108231513) TableName() string {
	return "migration_status"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20231108231513",
		Description: "",
		Migrate: func(tx *xorm.Engine) error {

			err := tx.Sync2(migrationStatus20231108231513{})
			if err != nil {
				return err
			}

			all := []*migrationStatus20231108231513{}
			err = tx.Find(&all)
			if err != nil {
				return err
			}

			for _, status := range all {
				status.StartedAt = status.Created
				status.FinishedAt = status.Created
				_, err = tx.Where("id = ?", status.ID).Update(status)
				if err != nil {
					return err
				}
			}

			if config.DatabaseType.GetString() == "sqlite" {
				_, err = tx.Exec(`create table migration_status_dg_tmp
				(
					id            INTEGER not null
				primary key autoincrement,
					user_id       INTEGER not null,
					migrator_name TEXT,
					started_at DATETIME null,
					finished_at DATETIME null
				);

				insert into migration_status_dg_tmp(id, user_id, migrator_name, started_at, finished_at)
				select id, user_id, migrator_name, started_at, finished_at
					from migration_status;

					drop table migration_status;

					alter table migration_status_dg_tmp
					rename to migration_status;

					create unique index UQE_migration_status_id
					on migration_status (id);
`)
				return err
			}

			err = dropTableColum(tx, "migration_status", "created")
			return err
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
