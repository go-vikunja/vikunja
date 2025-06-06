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

type projects20230903143017 struct {
	ID              int64 `xorm:"bigint autoincr not null unique pk" json:"id" param:"project"`
	DefaultBucketID int64 `xorm:"bigint INDEX null" json:"default_bucket_id"`
	DoneBucketID    int64 `xorm:"bigint INDEX null" json:"done_bucket_id"`
}

func (projects20230903143017) TableName() string {
	return "projects"
}

type bucket20230903143017 struct {
	ID           int64 `xorm:"bigint autoincr not null unique pk" json:"id" param:"bucket"`
	IsDoneBucket bool  `xorm:"BOOL" json:"is_done_bucket"`
	ProjectID    int64 `xorm:"bigint not null" json:"project_id" param:"project"`
}

func (bucket20230903143017) TableName() string {
	return "buckets"
}

const dropIsDoneBucketColSqlite20230903143017 = `
create table buckets_dg_tmp
(
    id            INTEGER  not null
        primary key autoincrement,
    title         TEXT     not null,
    project_id    INTEGER  not null,
    "limit"       INTEGER default 0,
    position      REAL,
    created       DATETIME not null,
    updated       DATETIME not null,
    created_by_id INTEGER  not null
);

insert into buckets_dg_tmp(id, title, project_id, "limit", position, created, updated, created_by_id)
select id,
       title,
       project_id,
       "limit",
       position,
       created,
       updated,
       created_by_id
from buckets;

drop table buckets;

alter table buckets_dg_tmp
    rename to buckets;

create unique index UQE_buckets_id
    on buckets (id);
`

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20230903143017",
		Description: "Move done bucket state to project + add default bucket setting",
		Migrate: func(tx *xorm.Engine) (err error) {
			err = tx.Sync2(projects20230903143017{})
			if err != nil {
				return
			}

			doneBuckets := []*bucket20230903143017{}
			err = tx.Where("is_done_bucket = true").
				Find(&doneBuckets)
			if err != nil {
				return
			}

			for _, bucket := range doneBuckets {
				_, err = tx.Where("id = ?", bucket.ProjectID).
					Cols("done_bucket_id").
					Update(&projects20230903143017{
						DoneBucketID: bucket.ID,
					})
				if err != nil {
					return
				}
			}

			if tx.Dialect().URI().DBType == schemas.SQLITE {
				_, err = tx.Exec(dropIsDoneBucketColSqlite20230903143017)
				return err
			}

			return dropTableColum(tx, "buckets", "is_done_bucket")
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
