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
		ID:          "20250317174522",
		Description: "",
		Migrate: func(tx *xorm.Engine) (err error) {
			if tx.Dialect().URI().DBType == schemas.SQLITE {
				_, err = tx.Exec(`create table teams_dg_tmp
(
    id            INTEGER           not null
        primary key autoincrement,
    name          TEXT              not null,
    description   TEXT,
    created_by_id INTEGER           not null,
    external_id   TEXT,
    issuer        TEXT,
    created       DATETIME,
    updated       DATETIME,
    is_public     INTEGER default 0 not null
);

insert into teams_dg_tmp(id, name, description, created_by_id, external_id, issuer, created, updated, is_public)
select id,
       name,
       description,
       created_by_id,
       oidc_id,
       issuer,
       created,
       updated,
       is_public
from teams;

drop table teams;

alter table teams_dg_tmp
    rename to teams;

create index IDX_teams_created_by_id
    on teams (created_by_id);

create unique index UQE_teams_id
    on teams (id);
`)
				return
			}

			_, err = tx.Exec("ALTER TABLE `teams` RENAME COLUMN `oidc_id` TO `external_id`")
			return
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
