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

type projects20240315104205 struct {
	ID              int64 `xorm:"bigint autoincr not null unique pk" json:"id" param:"project"`
	DefaultBucketID int64 `xorm:"bigint INDEX null" json:"default_bucket_id"`
	DoneBucketID    int64 `xorm:"bigint INDEX null" json:"done_bucket_id"`
}

func (projects20240315104205) TableName() string {
	return "projects"
}

type projectView20240315104205 struct {
	ID              int64 `xorm:"bigint autoincr not null unique pk" json:"id" param:"project"`
	ViewKind        int   `xorm:"not null" json:"view_kind"`
	DefaultBucketID int64 `xorm:"bigint INDEX null" json:"default_bucket_id"`
	DoneBucketID    int64 `xorm:"bigint INDEX null" json:"done_bucket_id"`
	ProjectID       int64 `xorm:"not null index" json:"project_id" param:"project"`
}

func (projectView20240315104205) TableName() string {
	return "project_views"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20240315104205",
		Description: "Move done and default bucket id to views",
		Migrate: func(tx *xorm.Engine) (err error) {
			err = tx.Sync(projectView20240315104205{})
			if err != nil {
				return
			}

			projects := []*projects20240315104205{}
			err = tx.Find(&projects)
			if err != nil {
				return
			}

			views := []*projectView20240315104205{}
			err = tx.Find(&views)
			if err != nil {
				return err
			}

			viewMap := make(map[int64][]*projectView20240315104205)
			for _, view := range views {
				if _, has := viewMap[view.ProjectID]; !has {
					viewMap[view.ProjectID] = []*projectView20240315104205{}
				}

				viewMap[view.ProjectID] = append(viewMap[view.ProjectID], view)
			}

			for _, project := range projects {
				for _, view := range viewMap[project.ID] {
					if view.ViewKind == 3 { // Kanban view
						view.DefaultBucketID = project.DefaultBucketID
						view.DoneBucketID = project.DoneBucketID
						_, err = tx.
							Where("id = ?", view.ID).
							Cols("default_bucket_id", "done_bucket_id").
							Update(view)
						if err != nil {
							return
						}
					}
				}
			}

			if config.DatabaseType.GetString() == "sqlite" {
				_, err = tx.Exec(`
create table projects_dg_tmp
(
    id                   INTEGER           not null
        primary key autoincrement,
    title                TEXT              not null,
    description          TEXT,
    identifier           TEXT,
    hex_color            TEXT,
    owner_id             INTEGER           not null,
    parent_project_id    INTEGER,
    is_archived          INTEGER default 0 not null,
    background_file_id   INTEGER,
    background_blur_hash TEXT,
    position             REAL,
    created              DATETIME          not null,
    updated              DATETIME          not null
);

insert into projects_dg_tmp(id, title, description, identifier, hex_color, owner_id, parent_project_id, is_archived,
                            background_file_id, background_blur_hash, position, created, updated)
select id,
       title,
       description,
       identifier,
       hex_color,
       owner_id,
       parent_project_id,
       is_archived,
       background_file_id,
       background_blur_hash,
       position,
       created,
       updated
from projects;

drop table projects;

alter table projects_dg_tmp
    rename to projects;

create index IDX_projects_owner_id
    on projects (owner_id);

create index IDX_projects_parent_project_id
    on projects (parent_project_id);

create unique index UQE_projects_id
    on projects (id);
`)
				return err
			}

			err = dropTableColum(tx, "projects", "done_bucket_id")
			if err != nil {
				return
			}
			return dropTableColum(tx, "projects", "default_bucket_id")
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
