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

const sqliteRename20221113170740 = `
-- buckets
create table buckets_dg_tmp
(
    id             INTEGER  not null
        primary key autoincrement,
    title          TEXT     not null,
    project_id     INTEGER  not null,
    "limit"        INTEGER default 0,
    is_done_bucket INTEGER,
    position       REAL,
    created        DATETIME not null,
    updated        DATETIME not null,
    created_by_id  INTEGER  not null
);

insert into buckets_dg_tmp(id, title, project_id, "limit", is_done_bucket, position, created, updated, created_by_id)
select id,
       title,
       list_id,
       "limit",
       is_done_bucket,
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

-- link shares
create table link_shares_dg_tmp
(
    id           INTEGER           not null
        primary key autoincrement,
    hash         TEXT              not null,
    name         TEXT,
    project_id   INTEGER           not null,
    "right"      INTEGER default 0 not null,
    sharing_type INTEGER default 0 not null,
    password     TEXT,
    shared_by_id INTEGER           not null,
    created      DATETIME          not null,
    updated      DATETIME          not null
);

insert into link_shares_dg_tmp(id, hash, name, project_id, "right", sharing_type, password, shared_by_id, created,
                               updated)
select id,
       hash,
       name,
       list_id,
       "right",
       sharing_type,
       password,
       shared_by_id,
       created,
       updated
from link_shares;

drop table link_shares;

alter table link_shares_dg_tmp
    rename to link_shares;

create index IDX_link_shares_right
    on link_shares ("right");

create index IDX_link_shares_shared_by_id
    on link_shares (shared_by_id);

create index IDX_link_shares_sharing_type
    on link_shares (sharing_type);

create unique index UQE_link_shares_hash
    on link_shares (hash);

create unique index UQE_link_shares_id
    on link_shares (id);

-- tasks
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
    bucket_id                 INTEGER,
    position                  REAL,
    kanban_position           REAL,
    created_by_id             INTEGER           not null
);

insert into tasks_dg_tmp(id, title, description, done, done_at, due_date, project_id, repeat_after, repeat_mode,
                         priority, start_date, end_date, hex_color, percent_done, "index", uid,
                         cover_image_attachment_id, created, updated, bucket_id, position, kanban_position,
                         created_by_id)
select id,
       title,
       description,
       done,
       done_at,
       due_date,
       list_id,
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
       bucket_id,
       position,
       kanban_position,
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

create index IDX_tasks_list_id
    on tasks (project_id);

create index IDX_tasks_repeat_after
    on tasks (repeat_after);

create index IDX_tasks_start_date
    on tasks (start_date);

create unique index UQE_tasks_id
    on tasks (id);

--- team_lists
create table team_lists_dg_tmp
(
    id         INTEGER           not null
        primary key autoincrement,
    team_id    INTEGER           not null,
    project_id INTEGER           not null,
    "right"    INTEGER default 0 not null,
    created    DATETIME          not null,
    updated    DATETIME          not null
);

insert into team_lists_dg_tmp(id, team_id, project_id, "right", created, updated)
select id, team_id, list_id, "right", created, updated
from team_lists;

drop table team_lists;

alter table team_lists_dg_tmp
    rename to team_lists;

create index IDX_team_lists_list_id
    on team_lists (project_id);

create index IDX_team_lists_right
    on team_lists ("right");

create index IDX_team_lists_team_id
    on team_lists (team_id);

create unique index UQE_team_lists_id
    on team_lists (id);

--- users
create table users_dg_tmp
(
    id                              INTEGER                 not null
        primary key autoincrement,
    name                            TEXT,
    username                        TEXT                    not null,
    password                        TEXT,
    email                           TEXT,
    status                          INTEGER default 0,
    avatar_provider                 TEXT,
    avatar_file_id                  INTEGER,
    issuer                          TEXT,
    subject                         TEXT,
    email_reminders_enabled         INTEGER default 1,
    discoverable_by_name            INTEGER default 0,
    discoverable_by_email           INTEGER default 0,
    overdue_tasks_reminders_enabled INTEGER default 1,
    overdue_tasks_reminders_time    TEXT    default '09:00' not null,
    default_project_id              INTEGER,
    week_start                      INTEGER,
    language                        TEXT,
    timezone                        TEXT,
    deletion_scheduled_at           DATETIME,
    deletion_last_reminder_sent     DATETIME,
    export_file_id                  INTEGER,
    created                         DATETIME                not null,
    updated                         DATETIME                not null
);

insert into users_dg_tmp(id, name, username, password, email, status, avatar_provider, avatar_file_id, issuer, subject,
                         email_reminders_enabled, discoverable_by_name, discoverable_by_email,
                         overdue_tasks_reminders_enabled, overdue_tasks_reminders_time, default_project_id, week_start,
                         language, timezone, deletion_scheduled_at, deletion_last_reminder_sent, export_file_id,
                         created, updated)
select id,
       name,
       username,
       password,
       email,
       status,
       avatar_provider,
       avatar_file_id,
       issuer,
       subject,
       email_reminders_enabled,
       discoverable_by_name,
       discoverable_by_email,
       overdue_tasks_reminders_enabled,
       overdue_tasks_reminders_time,
       default_list_id,
       week_start,
       language,
       timezone,
       deletion_scheduled_at,
       deletion_last_reminder_sent,
       export_file_id,
       created,
       updated
from users;

drop table users;

alter table users_dg_tmp
    rename to users;

create index IDX_users_default_list_id
    on users (default_project_id);

create index IDX_users_discoverable_by_email
    on users (discoverable_by_email);

create index IDX_users_discoverable_by_name
    on users (discoverable_by_name);

create index IDX_users_overdue_tasks_reminders_enabled
    on users (overdue_tasks_reminders_enabled);

create unique index UQE_users_id
    on users (id);

create unique index UQE_users_username
    on users (username);

--- users_list
create table users_lists_dg_tmp
(
    id         INTEGER           not null
        primary key autoincrement,
    user_id    INTEGER           not null,
    project_id INTEGER           not null,
    "right"    INTEGER default 0 not null,
    created    DATETIME          not null,
    updated    DATETIME          not null
);

insert into users_lists_dg_tmp(id, user_id, project_id, "right", created, updated)
select id, user_id, list_id, "right", created, updated
from users_lists;

drop table users_lists;

alter table users_lists_dg_tmp
    rename to users_lists;

create index IDX_users_lists_list_id
    on users_lists (project_id);

create index IDX_users_lists_right
    on users_lists ("right");

create index IDX_users_lists_user_id
    on users_lists (user_id);

create unique index UQE_users_lists_id
    on users_lists (id);
`

type colToRename struct {
	table   string
	oldName string
	newName string
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20221113170740",
		Description: "Rename lists to projects",
		Migrate: func(tx *xorm.Engine) error {
			// SQLite does not support renaming columns. Instead, we'll need to run manual sql.
			if tx.Dialect().URI().DBType == schemas.SQLITE {
				_, err := tx.Exec(sqliteRename20221113170740)
				if err != nil {
					return err
				}
			} else {

				colsToRename := []*colToRename{
					{
						table:   "buckets",
						oldName: "list_id",
						newName: "project_id",
					},
					{
						table:   "link_shares",
						oldName: "list_id",
						newName: "project_id",
					},
					{
						table:   "tasks",
						oldName: "list_id",
						newName: "project_id",
					},
					{
						table:   "team_lists",
						oldName: "list_id",
						newName: "project_id",
					},
					{
						table:   "users",
						oldName: "default_list_id",
						newName: "default_project_id",
					},
					{
						table:   "users_lists",
						oldName: "list_id",
						newName: "project_id",
					},
				}

				for _, col := range colsToRename {
					if tx.Dialect().URI().DBType == schemas.POSTGRES || tx.Dialect().URI().DBType == schemas.MYSQL {
						_, err := tx.Exec("ALTER TABLE `" + col.table + "` RENAME COLUMN `" + col.oldName + "` TO `" + col.newName + "`")
						if err != nil {
							return err
						}
					}
				}
			}

			err := renameTable(tx, "lists", "projects")
			if err != nil {
				return err
			}

			err = renameTable(tx, "team_lists", "team_projects")
			if err != nil {
				return err
			}

			err = renameTable(tx, "users_lists", "users_projects")
			if err != nil {
				return err
			}

			return nil
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
