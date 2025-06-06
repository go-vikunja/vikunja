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

type buckets20240315093418 struct {
	ID            int64 `xorm:"bigint autoincr not null"`
	ProjectID     int64 `xorm:"bigint not null"`
	ProjectViewID int64 `xorm:"bigint not null default 0"`
}

func (buckets20240315093418) TableName() string {
	return "buckets"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20240315093418",
		Description: "Relate buckets to views instead of projects",
		Migrate: func(tx *xorm.Engine) (err error) {
			err = tx.Sync2(buckets20240315093418{})
			if err != nil {
				return
			}

			buckets := []*buckets20240315093418{}
			err = tx.Find(&buckets)
			if err != nil {
				return err
			}

			views := []*projectView20240313230538{}
			err = tx.Find(&views)
			if err != nil {
				return err
			}

			viewMap := make(map[int64][]*projectView20240313230538)
			for _, view := range views {
				if _, has := viewMap[view.ProjectID]; !has {
					viewMap[view.ProjectID] = []*projectView20240313230538{}
				}

				viewMap[view.ProjectID] = append(viewMap[view.ProjectID], view)
			}

			for _, bucket := range buckets {
				for _, view := range viewMap[bucket.ProjectID] {
					if view.ViewKind == 3 { // Kanban view

						bucket.ProjectViewID = view.ID

						_, err = tx.
							Where("id = ?", bucket.ID).
							Cols("project_view_id").
							Update(bucket)
						if err != nil {
							return err
						}
					}
				}
			}

			if config.DatabaseType.GetString() == "sqlite" {
				_, err = tx.Exec(`
create table buckets_dg_tmp
(
    id            INTEGER  not null
        primary key autoincrement,
    title         TEXT     not null,
    "limit"       INTEGER default 0,
    position      REAL,
    created       DATETIME not null,
    updated       DATETIME not null,
    created_by_id INTEGER  not null,
    project_view_id INTEGER  not null default 0
);

insert into buckets_dg_tmp(id, title, "limit", position, created, updated, created_by_id, project_view_id)
select id, title, "limit", position, created, updated, created_by_id, project_view_id
from buckets;

drop index if exists UQE_buckets_id;

drop table buckets;

alter table buckets_dg_tmp
    rename to buckets;

create unique index if not exists UQE_buckets_id
    on buckets (id);
`)
				return err
			}

			return dropTableColum(tx, "buckets", "project_id")
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
