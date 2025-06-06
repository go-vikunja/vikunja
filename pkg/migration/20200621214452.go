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
	"fmt"
	"strings"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20200621214452",
		Description: "Make all dates to iso time",
		Migrate: func(tx *xorm.Engine) error {

			// Big query for sqlite goes here
			// SQLite is not capable of modifying columns directly like mysql or postgres, so we need to add
			// all the sql we need manually here.
			if tx.Dialect().URI().DBType == schemas.SQLITE {
				sql := `
--- Buckets
create table buckets_dg_tmp
(
	id INTEGER not null
		primary key autoincrement,
	title TEXT not null,
	list_id INTEGER not null,
	created datetime not null,
	updated datetime not null,
	created_by_id INTEGER not null
);

insert into buckets_dg_tmp(id, title, list_id, created, updated, created_by_id) select id, title, list_id, created, updated, created_by_id from buckets;

drop table buckets;

alter table buckets_dg_tmp rename to buckets;

create unique index UQE_buckets_id
	on buckets (id);

--- files

create table files_dg_tmp
(
	id INTEGER not null
		primary key autoincrement,
	name TEXT not null,
	mime TEXT,
	size INTEGER not null,
	created datetime not null,
	created_by_id INTEGER not null
);

insert into files_dg_tmp(id, name, mime, size, created, created_by_id) select id, name, mime, size, created_unix, created_by_id from files;

drop table files;

alter table files_dg_tmp rename to files;

create unique index UQE_files_id
	on files (id);

--- label_task
create table label_task_dg_tmp
(
	id INTEGER not null
		primary key autoincrement,
	task_id INTEGER not null,
	label_id INTEGER not null,
	created datetime not null
);

insert into label_task_dg_tmp(id, task_id, label_id, created) select id, task_id, label_id, created from label_task;

drop table label_task;

alter table label_task_dg_tmp rename to label_task;

create index IDX_label_task_label_id
	on label_task (label_id);

create index IDX_label_task_task_id
	on label_task (task_id);

create unique index UQE_label_task_id
	on label_task (id);

--- labels
create table labels_dg_tmp
(
	id INTEGER not null
		primary key autoincrement,
	title TEXT not null,
	description TEXT,
	hex_color TEXT,
	created_by_id INTEGER not null,
	created datetime not null,
	updated datetime not null
);

insert into labels_dg_tmp(id, title, description, hex_color, created_by_id, created, updated) select id, title, description, hex_color, created_by_id, created, updated from labels;

drop table labels;

alter table labels_dg_tmp rename to labels;

create unique index UQE_labels_id
	on labels (id);

--- link_sharing
create table link_sharing_dg_tmp
(
	id INTEGER not null
		primary key autoincrement,
	hash TEXT not null,
	list_id INTEGER not null,
	"right" INTEGER default 0 not null,
	sharing_type INTEGER default 0 not null,
	shared_by_id INTEGER not null,
	created datetime not null,
	updated datetime not null
);

insert into link_sharing_dg_tmp(id, hash, list_id, "right", sharing_type, shared_by_id, created, updated) select id, hash, list_id, "right", sharing_type, shared_by_id, created, updated from link_sharing;

drop table link_sharing;

alter table link_sharing_dg_tmp rename to link_sharing;

create index IDX_link_sharing_right
	on link_sharing ("right");

create index IDX_link_sharing_shared_by_id
	on link_sharing (shared_by_id);

create index IDX_link_sharing_sharing_type
	on link_sharing (sharing_type);

create unique index UQE_link_sharing_hash
	on link_sharing (hash);

create unique index UQE_link_sharing_id
	on link_sharing (id);

--- list
create table list_dg_tmp
(
	id INTEGER not null
		primary key autoincrement,
	title TEXT not null,
	description TEXT,
	identifier TEXT,
	hex_color TEXT,
	owner_id INTEGER not null,
	namespace_id INTEGER not null,
	is_archived INTEGER default 0 not null,
	background_file_id INTEGER,
	created datetime not null,
	updated datetime not null
);

insert into list_dg_tmp(id, title, description, identifier, hex_color, owner_id, namespace_id, is_archived, background_file_id, created, updated) select id, title, description, identifier, hex_color, owner_id, namespace_id, is_archived, background_file_id, created, updated from list;

drop table list;

alter table list_dg_tmp rename to list;

create index IDX_list_namespace_id
	on list (namespace_id);

create index IDX_list_owner_id
	on list (owner_id);

create unique index UQE_list_id
	on list (id);

--- migration_status
create table migration_status_dg_tmp
(
	id INTEGER not null
		primary key autoincrement,
	user_id INTEGER not null,
	migrator_name TEXT,
	created datetime not null
);

insert into migration_status_dg_tmp(id, user_id, migrator_name, created) select id, user_id, migrator_name, created_unix from migration_status;

drop table migration_status;

alter table migration_status_dg_tmp rename to migration_status;

create unique index UQE_migration_status_id
	on migration_status (id);

--- namespaces
create table namespaces_dg_tmp
(
	id INTEGER not null
		primary key autoincrement,
	title TEXT not null,
	description TEXT,
	owner_id INTEGER not null,
	hex_color TEXT,
	is_archived INTEGER default 0 not null,
	created datetime not null,
	updated datetime not null
);

insert into namespaces_dg_tmp(id, title, description, owner_id, hex_color, is_archived, created, updated) select id, title, description, owner_id, hex_color, is_archived, created, updated from namespaces;

drop table namespaces;

alter table namespaces_dg_tmp rename to namespaces;

create index IDX_namespaces_owner_id
	on namespaces (owner_id);

create unique index UQE_namespaces_id
	on namespaces (id);

--- task_assignees
create table task_assignees_dg_tmp
(
	id INTEGER not null
		primary key autoincrement,
	task_id INTEGER not null,
	user_id INTEGER not null,
	created datetime not null
);

insert into task_assignees_dg_tmp(id, task_id, user_id, created) select id, task_id, user_id, created from task_assignees;

drop table task_assignees;

alter table task_assignees_dg_tmp rename to task_assignees;

create index IDX_task_assignees_task_id
	on task_assignees (task_id);

create index IDX_task_assignees_user_id
	on task_assignees (user_id);

create unique index UQE_task_assignees_id
	on task_assignees (id);

--- task_attachments
create table task_attachments_dg_tmp
(
	id INTEGER not null
		primary key autoincrement,
	task_id INTEGER not null,
	file_id INTEGER not null,
	created_by_id INTEGER not null,
	created datetime not null
);

insert into task_attachments_dg_tmp(id, task_id, file_id, created_by_id, created) select id, task_id, file_id, created_by_id, created from task_attachments;

drop table task_attachments;

alter table task_attachments_dg_tmp rename to task_attachments;

create unique index UQE_task_attachments_id
	on task_attachments (id);

--- task_comments
create table task_comments_dg_tmp
(
	id INTEGER not null
		primary key autoincrement,
	comment TEXT not null,
	author_id INTEGER not null,
	task_id INTEGER not null,
	created datetime not null,
	updated datetime not null
);

insert into task_comments_dg_tmp(id, comment, author_id, task_id, created, updated) select id, comment, author_id, task_id, created, updated from task_comments;

drop table task_comments;

alter table task_comments_dg_tmp rename to task_comments;

create unique index UQE_task_comments_id
	on task_comments (id);

--- task_relations
create table task_relations_dg_tmp
(
	id INTEGER not null
		primary key autoincrement,
	task_id INTEGER not null,
	other_task_id INTEGER not null,
	relation_kind TEXT not null,
	created_by_id INTEGER not null,
	created datetime not null
);

insert into task_relations_dg_tmp(id, task_id, other_task_id, relation_kind, created_by_id, created) select id, task_id, other_task_id, relation_kind, created_by_id, created from task_relations;

drop table task_relations;

alter table task_relations_dg_tmp rename to task_relations;

create unique index UQE_task_relations_id
	on task_relations (id);

--- task_reminders
create table task_reminders_dg_tmp
(
	id INTEGER not null
		primary key autoincrement,
	task_id INTEGER not null,
	reminder datetime not null,
	created datetime not null
);

insert into task_reminders_dg_tmp(id, task_id, reminder, created) select id, task_id, reminder_unix, created from task_reminders;

drop table task_reminders;

alter table task_reminders_dg_tmp rename to task_reminders;

create index IDX_task_reminders_reminder_unix
	on task_reminders (reminder);

create index IDX_task_reminders_task_id
	on task_reminders (task_id);

create unique index UQE_task_reminders_id
	on task_reminders (id);

--- tasks
create table tasks_dg_tmp
(
	id INTEGER not null
		primary key autoincrement,
	title TEXT not null,
	description TEXT,
	done INTEGER,
	done_at datetime,
	due_date datetime,
	created_by_id INTEGER not null,
	list_id INTEGER not null,
	repeat_after INTEGER,
	repeat_from_current_date INTEGER,
	priority INTEGER,
	start_date datetime,
	end_date datetime,
	hex_color TEXT,
	percent_done REAL,
	"index" INTEGER default 0 not null,
	uid TEXT,
	created datetime not null,
	updated datetime not null,
	bucket_id INTEGER,
	position REAL
);

insert into tasks_dg_tmp(id, title, description, done, done_at, due_date, created_by_id, list_id, repeat_after, repeat_from_current_date, priority, start_date, end_date, hex_color, percent_done, "index", uid, created, updated, bucket_id, position) select id, title, description, done, done_at_unix, due_date_unix, created_by_id, list_id, repeat_after, repeat_from_current_date, priority, start_date_unix, end_date_unix, hex_color, percent_done, "index", uid, created, updated, bucket_id, position from tasks;

drop table tasks;

alter table tasks_dg_tmp rename to tasks;

create index IDX_tasks_done
	on tasks (done);

create index IDX_tasks_done_at_unix
	on tasks (done_at);

create index IDX_tasks_due_date_unix
	on tasks (due_date);

create index IDX_tasks_end_date_unix
	on tasks (end_date);

create index IDX_tasks_list_id
	on tasks (list_id);

create index IDX_tasks_repeat_after
	on tasks (repeat_after);

create index IDX_tasks_start_date_unix
	on tasks (start_date);

create unique index UQE_tasks_id
	on tasks (id);

--- team_list
create table team_list_dg_tmp
(
	id INTEGER not null
		primary key autoincrement,
	team_id INTEGER not null,
	list_id INTEGER not null,
	"right" INTEGER default 0 not null,
	created datetime not null,
	updated datetime not null
);

insert into team_list_dg_tmp(id, team_id, list_id, "right", created, updated) select id, team_id, list_id, "right", created, updated from team_list;

drop table team_list;

alter table team_list_dg_tmp rename to team_list;

create index IDX_team_list_list_id
	on team_list (list_id);

create index IDX_team_list_right
	on team_list ("right");

create index IDX_team_list_team_id
	on team_list (team_id);

create unique index UQE_team_list_id
	on team_list (id);

--- team_members
create table team_members_dg_tmp
(
	id INTEGER not null
		primary key autoincrement,
	team_id INTEGER not null,
	user_id INTEGER not null,
	admin INTEGER,
	created datetime not null
);

insert into team_members_dg_tmp(id, team_id, user_id, admin, created) select id, team_id, user_id, admin, created from team_members;

drop table team_members;

alter table team_members_dg_tmp rename to team_members;

create index IDX_team_members_team_id
	on team_members (team_id);

create index IDX_team_members_user_id
	on team_members (user_id);

create unique index UQE_team_members_id
	on team_members (id);

--- team_namespaces
create table team_namespaces_dg_tmp
(
	id INTEGER not null
		primary key autoincrement,
	team_id INTEGER not null,
	namespace_id INTEGER not null,
	"right" INTEGER default 0 not null,
	created datetime not null,
	updated datetime not null
);

insert into team_namespaces_dg_tmp(id, team_id, namespace_id, "right", created, updated) select id, team_id, namespace_id, "right", created, updated from team_namespaces;

drop table team_namespaces;

alter table team_namespaces_dg_tmp rename to team_namespaces;

create index IDX_team_namespaces_namespace_id
	on team_namespaces (namespace_id);

create index IDX_team_namespaces_right
	on team_namespaces ("right");

create index IDX_team_namespaces_team_id
	on team_namespaces (team_id);

create unique index UQE_team_namespaces_id
	on team_namespaces (id);

--- teams
create table teams_dg_tmp
(
	id INTEGER not null
		primary key autoincrement,
	name TEXT not null,
	description TEXT,
	created_by_id INTEGER not null,
	created datetime not null,
	updated datetime not null
);

insert into teams_dg_tmp(id, name, description, created_by_id, created, updated) select id, name, description, created_by_id, created, updated from teams;

drop table teams;

alter table teams_dg_tmp rename to teams;

create index IDX_teams_created_by_id
	on teams (created_by_id);

create unique index UQE_teams_id
	on teams (id);

--- users
create table users_dg_tmp
(
	id INTEGER not null
		primary key autoincrement,
	username TEXT not null,
	password TEXT not null,
	email TEXT,
	is_active INTEGER,
	password_reset_token TEXT,
	email_confirm_token TEXT,
	created datetime not null,
	updated datetime not null
);

insert into users_dg_tmp(id, username, password, email, is_active, password_reset_token, email_confirm_token, created, updated) select id, username, password, email, is_active, password_reset_token, email_confirm_token, created, updated from users;

drop table users;

alter table users_dg_tmp rename to users;

create unique index UQE_users_id
	on users (id);

create unique index UQE_users_username
	on users (username);

--- users_list
create table users_list_dg_tmp
(
	id INTEGER not null
		primary key autoincrement,
	user_id INTEGER not null,
	list_id INTEGER not null,
	"right" INTEGER default 0 not null,
	created datetime not null,
	updated datetime not null
);

insert into users_list_dg_tmp(id, user_id, list_id, "right", created, updated) select id, user_id, list_id, "right", created, updated from users_list;

drop table users_list;

alter table users_list_dg_tmp rename to users_list;

create index IDX_users_list_list_id
	on users_list (list_id);

create index IDX_users_list_right
	on users_list ("right");

create index IDX_users_list_user_id
	on users_list (user_id);

create unique index UQE_users_list_id
	on users_list (id);

--- users_namespace
create table users_namespace_dg_tmp
(
	id INTEGER not null
		primary key autoincrement,
	user_id INTEGER not null,
	namespace_id INTEGER not null,
	"right" INTEGER default 0 not null,
	created datetime not null,
	updated datetime not null
);

insert into users_namespace_dg_tmp(id, user_id, namespace_id, "right", created, updated) select id, user_id, namespace_id, "right", created, updated from users_namespace;

drop table users_namespace;

alter table users_namespace_dg_tmp rename to users_namespace;

create index IDX_users_namespace_namespace_id
	on users_namespace (namespace_id);

create index IDX_users_namespace_right
	on users_namespace ("right");

create index IDX_users_namespace_user_id
	on users_namespace (user_id);

create unique index UQE_users_namespace_id
	on users_namespace (id);
`
				sess := tx.NewSession()
				if err := sess.Begin(); err != nil {
					return err
				}
				_, err := sess.Exec(sql)
				if err != nil {
					_ = sess.Rollback()
					return err
				}
				if err := sess.Commit(); err != nil {
					return err
				}
			}

			convertTime := func(table, column string) error {

				var sql []string
				colOld := "`" + column + "`"
				colTmp := "`" + column + `_ts` + "`"
				// If the column namme ends with "_unix", we want to directly remove that since the timestamp
				// isn't a unix one anymore.
				var colFinal = colOld
				if strings.HasSuffix(column, "_unix") {
					colFinal = "`" + column[:len(column)-5] + "`"
				}

				switch tx.Dialect().URI().DBType {
				case schemas.POSTGRES:
					sql = []string{
						"ALTER TABLE " + table + " DROP COLUMN IF EXISTS " + colTmp + ";",
						"ALTER TABLE " + table + " ADD COLUMN " + colTmp + " TIMESTAMP WITHOUT TIME ZONE NULL;",
					}
					if colFinal != colOld {
						sql = append(sql, "ALTER TABLE "+table+" ADD COLUMN "+colFinal+" TIMESTAMP WITHOUT TIME ZONE NULL;")
					}
					sql = append(sql,
						// #nosec
						"UPDATE "+table+" SET "+colTmp+" = (CASE WHEN "+colOld+" = 0 THEN NULL ELSE TIMESTAMP 'epoch' + "+colOld+" * INTERVAL '1 second' END);",
						"ALTER TABLE "+table+" ALTER COLUMN "+colFinal+" TYPE TIMESTAMP USING "+colTmp+";",
						"ALTER TABLE "+table+" DROP COLUMN "+colTmp+";",
					)
					if colFinal != colOld {
						sql = append(sql, "ALTER TABLE "+table+" DROP COLUMN "+colOld+";")
					}
				case schemas.MYSQL:
					sql = []string{
						// mysql does not support the IF EXISTS part of the following statement. To not break
						// compatibility with mysql over mariadb, we're not using it.
						// The statement is probably useless anyway since its only purpose is to clean up old tables
						// which may be leftovers from a previously failed migration. However, since the whole thing
						// is wrapped in sessions, this is extremely unlikely to happen anyway.
						// "ALTER TABLE " + table + " DROP COLUMN IF EXISTS " + colTmp + ";",
						"ALTER TABLE " + table + " ADD COLUMN " + colTmp + " DATETIME NULL;",
						// #nosec
						"UPDATE " + table + " SET " + colTmp + " = IF(" + colOld + " = 0, NULL, FROM_UNIXTIME(" + colOld + "));",
						"ALTER TABLE " + table + " DROP COLUMN " + colOld + ";",
						"ALTER TABLE " + table + " CHANGE " + colTmp + " " + colFinal + " DATETIME NULL;",
					}
				case schemas.SQLITE:
					// welp
					// All created and updated columns are set to not null
					// But some of the test data is 0 so we can't use our update script on it.
					if column != "updated" && column != "created" {
						sql = []string{
							// #nosec
							"UPDATE " + table + " SET " + colFinal + " = CASE WHEN " + colFinal + " > 0 THEN DATETIME(" + colFinal + ", 'unixepoch', 'localtime') ELSE NULL END",
						}
					} else {
						sql = []string{
							// #nosec
							"UPDATE " + table + " SET " + colFinal + " = DATETIME(" + colFinal + ", 'unixepoch', 'localtime')",
						}
					}
				default:
					return fmt.Errorf("unsupported dbms: %s", tx.Dialect().URI().DBType)
				}

				sess := tx.NewSession()
				if err := sess.Begin(); err != nil {
					return fmt.Errorf("unable to open session: %w", err)
				}
				for _, s := range sql {
					_, err := sess.Exec(s)
					if err != nil {
						_ = sess.Rollback()
						return fmt.Errorf("error executing update data for table %s, column %s: %w", table, column, err)
					}
				}
				if err := sess.Commit(); err != nil {
					return fmt.Errorf("error committing data change: %w", err)
				}
				return nil
			}

			for table, columns := range map[string][]string{
				"buckets": {
					"created",
					"updated",
				},
				"files": {
					"created_unix",
				},
				"label_task": {
					"created",
				},
				"labels": {
					"created",
					"updated",
				},
				"link_sharing": {
					"created",
					"updated",
				},
				"list": {
					"created",
					"updated",
				},
				"migration_status": {
					"created_unix",
				},
				"namespaces": {
					"created",
					"updated",
				},
				"task_assignees": {
					"created",
				},
				"task_attachments": {
					"created",
				},
				"task_comments": {
					"created",
					"updated",
				},
				"task_relations": {
					"created",
				},
				"task_reminders": {
					"created",
					"reminder_unix",
				},
				"tasks": {
					"done_at_unix",
					"due_date_unix",
					"start_date_unix",
					"end_date_unix",
					"created",
					"updated",
				},
				"team_list": {
					"created",
					"updated",
				},
				"team_members": {
					"created",
				},
				"team_namespaces": {
					"created",
					"updated",
				},
				"teams": {
					"created",
					"updated",
				},
				"users": {
					"created",
					"updated",
				},
				"users_list": {
					"created",
					"updated",
				},
				"users_namespace": {
					"created",
					"updated",
				},
			} {
				for _, column := range columns {
					if err := convertTime(table, column); err != nil {
						return err
					}
				}
			}

			return nil
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
