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

type task20200515172220 struct {
	ID    int64  `xorm:"int(11) autoincr not null unique pk"`
	Title string `xorm:"varchar(250) null"`
	Text  string `xorm:"varchar(250) not null"`
}

func (t *task20200515172220) TableName() string {
	return "tasks"
}

// We can't do the migration if the title column is not null but has a default value of null,
// so we initialize it as null and change it after migrating.
type task20200515172221 struct {
	Title string `xorm:"varchar(250) not null"`
}

func (t *task20200515172221) TableName() string {
	return "tasks"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20200515172220",
		Description: "Change task text to title",
		Migrate: func(tx *xorm.Engine) error {
			err := tx.Sync2(task20200515172220{})
			if err != nil {
				return err
			}

			tasks := []*task20200515172220{}
			err = tx.Find(&tasks)
			if err != nil {
				return err
			}

			for _, task := range tasks {
				task.Title = task.Text
				_, err = tx.Where("id = ?", task.ID).Update(task)
				if err != nil {
					return err
				}
			}

			err = tx.Sync2(task20200515172221{})
			if err != nil {
				return err
			}

			// sqlite is not able to drop columns. To have inserting new tasks still work, we drop the column manually.
			if config.DatabaseType.GetString() == "sqlite" {
				_, err = tx.Exec(`
create table tasks_dg_tmp
(
	id INTEGER not null
		primary key autoincrement,
	description TEXT,
	done INTEGER,
	done_at_unix INTEGER,
	due_date_unix INTEGER,
	created_by_id INTEGER not null,
	list_id INTEGER not null,
	repeat_after INTEGER,
	priority INTEGER,
	start_date_unix INTEGER,
	end_date_unix INTEGER,
	hex_color TEXT,
	percent_done REAL,
	"index" INTEGER default 0 not null,
	uid TEXT,
	created INTEGER not null,
	updated INTEGER not null,
	bucket_id INTEGER,
	position REAL,
	title TEXT
);

insert into tasks_dg_tmp(id, description, done, done_at_unix, due_date_unix, created_by_id, list_id, repeat_after, priority, start_date_unix, end_date_unix, hex_color, percent_done, "index", uid, created, updated, bucket_id, position, title) select id, description, done, done_at_unix, due_date_unix, created_by_id, list_id, repeat_after, priority, start_date_unix, end_date_unix, hex_color, percent_done, "index", uid, created, updated, bucket_id, position, title from tasks;

drop table tasks;

alter table tasks_dg_tmp rename to tasks;

create unique index tasks_id_uindex
	on tasks (id);
`)
				return err
			}

			return dropTableColum(tx, "tasks", "text")
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
