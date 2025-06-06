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
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

type namespace20200515195546 struct {
	ID    int64  `xorm:"int(11) autoincr not null unique pk"`
	Title string `xorm:"varchar(250) null"`
	Name  string `xorm:"varchar(250) not null"`
}

func (t *namespace20200515195546) TableName() string {
	return "namespaces"
}

// We can't do the migration if the title column is not null but has a default value of null,
// so we initialize it as null and change it after migrating.
type namespace20200515195547 struct {
	Title string `xorm:"varchar(250) not null"`
}

func (t *namespace20200515195547) TableName() string {
	return "namespaces"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20200515195546",
		Description: "Change namespace name to title",
		Migrate: func(tx *xorm.Engine) error {
			err := tx.Sync2(namespace20200515195546{})
			if err != nil {
				return err
			}

			namespaces := []*namespace20200515195546{}
			err = tx.Find(&namespaces)
			if err != nil {
				return err
			}

			for _, n := range namespaces {
				n.Title = n.Name
				_, err = tx.Where("id = ?", n.ID).Update(n)
				if err != nil {
					return err
				}
			}

			err = tx.Sync2(namespace20200515195547{})
			if err != nil {
				return err
			}

			// sqlite is not able to drop columns. To have inserting new namespaces still work, we drop the column manually.
			if config.DatabaseType.GetString() == "sqlite" {
				_, err = tx.Exec(`
create table namespaces_dg_tmp
(
	id INTEGER not null
		primary key autoincrement,
	description TEXT,
	owner_id INTEGER not null,
	created INTEGER not null,
	updated INTEGER not null,
	is_archived INTEGER default 0 not null,
	hex_color TEXT,
	title TEXT
);

insert into namespaces_dg_tmp(id, description, owner_id, created, updated, is_archived, hex_color, title) select id, description, owner_id, created, updated, is_archived, hex_color, title from namespaces;

drop table namespaces;

alter table namespaces_dg_tmp rename to namespaces;

create unique index UQE_namespaces_id
	on namespaces (id);
`)
				return err
			}

			return dropTableColum(tx, "namespaces", "name")
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
