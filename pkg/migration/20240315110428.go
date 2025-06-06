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

type task20240315110428 struct {
	ID        int64 `xorm:"bigint autoincr not null unique pk" json:"id" param:"projecttask"`
	BucketID  int64 `xorm:"bigint not null"`
	ProjectID int64 `xorm:"bigint INDEX not null" json:"project_id" param:"project"`
}

func (task20240315110428) TableName() string {
	return "tasks"
}

type taskBuckets20240315110428 struct {
	BucketID      int64 `xorm:"bigint not null index"`
	TaskID        int64 `xorm:"bigint not null index"`
	ProjectViewID int64 `xorm:"bigint not null index"`
}

func (taskBuckets20240315110428) TableName() string {
	return "task_buckets"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20240315110428",
		Description: "",
		Migrate: func(tx *xorm.Engine) (err error) {
			err = tx.Sync2(taskBuckets20240315110428{})
			if err != nil {
				return
			}

			tasks := []*task20240315110428{}
			err = tx.Find(&tasks)
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

			for _, task := range tasks {
				for _, view := range viewMap[task.ProjectID] {
					if view.ViewKind == 3 { // Kanban view

						pos := taskBuckets20240315110428{
							TaskID:        task.ID,
							BucketID:      task.BucketID,
							ProjectViewID: view.ID,
						}

						_, err = tx.Insert(pos)
						if err != nil {
							return err
						}
					}
				}
			}

			if config.DatabaseType.GetString() == "sqlite" {
				_, err = tx.Exec(`
create table tasks_dg_tmp
(
    id                        INTEGER           not null
        primary key autoincrement,
    title                     TEXT              not null,
    description               TEXT,
    done                      INTEGER,
    done_at                   DATETIME,
    due_date                  DATETIME,
    project_id                INTEGER           not null,
    repeat_after              INTEGER,
    repeat_mode               INTEGER default 0 not null,
    priority                  INTEGER,
    start_date                DATETIME,
    end_date                  DATETIME,
    hex_color                 TEXT,
    percent_done              REAL,
    "index"                   INTEGER default 0 not null,
    uid                       TEXT,
    cover_image_attachment_id INTEGER default 0,
    created                   DATETIME          not null,
    updated                   DATETIME          not null,
    created_by_id             INTEGER           not null
);

insert into tasks_dg_tmp(id, title, description, done, done_at, due_date, project_id, repeat_after, repeat_mode,
                         priority, start_date, end_date, hex_color, percent_done, "index", uid,
                         cover_image_attachment_id, created, updated, created_by_id)
select id,
       title,
       description,
       done,
       done_at,
       due_date,
       project_id,
       repeat_after,
       repeat_mode,
       priority,
       start_date,
       end_date,
       hex_color,
       percent_done,
       "index",
       uid,
       cover_image_attachment_id,
       created,
       updated,
       created_by_id
from tasks;

drop table tasks;

alter table tasks_dg_tmp
    rename to tasks;

create index IDX_tasks_done
    on tasks (done);

create index IDX_tasks_done_at
    on tasks (done_at);

create index IDX_tasks_due_date
    on tasks (due_date);

create index IDX_tasks_end_date
    on tasks (end_date);

create index IDX_tasks_project_id
    on tasks (project_id);

create index IDX_tasks_repeat_after
    on tasks (repeat_after);

create index IDX_tasks_start_date
    on tasks (start_date);

create unique index UQE_tasks_id
    on tasks (id);

`)
				return err
			}

			return dropTableColum(tx, "tasks", "bucket_id")
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
