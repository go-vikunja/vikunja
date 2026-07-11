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
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

// Old task struct with legacy repeat fields for reading
type taskOld20260711120000 struct {
	ID          int64  `xorm:"bigint autoincr not null unique pk"`
	RepeatAfter int64  `xorm:"bigint INDEX null"`
	RepeatMode  int    `xorm:"not null default 0"`
	Repeats     string `xorm:"varchar(500) null"`
}

func (taskOld20260711120000) TableName() string {
	return "tasks"
}

// New task struct with RRULE fields only
type taskNew20260711120000 struct {
	ID                     int64  `xorm:"bigint autoincr not null unique pk"`
	Repeats                string `xorm:"varchar(500) null"`
	RepeatsFromCurrentDate bool   `xorm:"null default false"`
}

func (taskNew20260711120000) TableName() string {
	return "tasks"
}

// taskRepeatBackup20260711120000 preserves the legacy repeat columns before they
// are dropped, so a mis-conversion can be recovered. It can be dropped by a later
// migration once the RRULE conversion has been verified in production.
type taskRepeatBackup20260711120000 struct {
	ID          int64 `xorm:"bigint not null pk"`
	RepeatAfter int64 `xorm:"bigint null"`
	RepeatMode  int   `xorm:"not null default 0"`
}

func (taskRepeatBackup20260711120000) TableName() string {
	return "task_repeat_legacy_backup"
}

// convertLegacyRepeatToRRule converts legacy repeat_after/repeat_mode to an RRULE string.
func convertLegacyRepeatToRRule(repeatAfter int64, repeatMode int) string {
	const (
		TaskRepeatModeDefault         = 0
		TaskRepeatModeMonth           = 1
		TaskRepeatModeFromCurrentDate = 2
	)

	switch repeatMode {
	case TaskRepeatModeMonth:
		return "FREQ=MONTHLY;INTERVAL=1"
	case TaskRepeatModeDefault, TaskRepeatModeFromCurrentDate:
		if repeatAfter <= 0 {
			return ""
		}
		return secondsToRRule(repeatAfter)
	}
	return ""
}

// secondsToRRule converts seconds to an appropriate RRULE string.
func secondsToRRule(seconds int64) string {
	const (
		minute = int64(time.Minute / time.Second)
		hour   = int64(time.Hour / time.Second)
		day    = int64((24 * time.Hour) / time.Second)
		week   = int64((7 * 24 * time.Hour) / time.Second)
	)

	if seconds%week == 0 {
		return fmt.Sprintf("FREQ=WEEKLY;INTERVAL=%d", seconds/week)
	}
	if seconds%day == 0 {
		return fmt.Sprintf("FREQ=DAILY;INTERVAL=%d", seconds/day)
	}
	if seconds%hour == 0 {
		return fmt.Sprintf("FREQ=HOURLY;INTERVAL=%d", seconds/hour)
	}
	if seconds%minute == 0 {
		return fmt.Sprintf("FREQ=MINUTELY;INTERVAL=%d", seconds/minute)
	}
	return fmt.Sprintf("FREQ=SECONDLY;INTERVAL=%d", seconds)
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260711120000",
		Description: "Replace legacy repeat fields with RRULE",
		Migrate: func(tx *xorm.Engine) error {
			const TaskRepeatModeFromCurrentDate = 2

			// Step 1: Add new RRULE columns
			err := tx.Sync2(taskNew20260711120000{})
			if err != nil {
				return err
			}

			// Step 2: Back up the legacy repeat columns, then convert them to RRULE.
			// Page by id so we never load the whole table into memory, and copy each
			// row's legacy values into a backup table before they are dropped below,
			// so a mis-conversion can be recovered.
			if err := tx.Sync2(taskRepeatBackup20260711120000{}); err != nil {
				return err
			}

			const batchSize = 500
			var (
				lastID                                   int64
				total, converted, skipped, unconvertible int
			)
			for {
				var tasks []taskOld20260711120000
				err = tx.Where("(repeat_after > 0 OR repeat_mode > 0) AND id > ?", lastID).
					OrderBy("id ASC").
					Limit(batchSize).
					Find(&tasks)
				if err != nil {
					return err
				}
				if len(tasks) == 0 {
					break
				}

				// Back up this page's legacy values before any mutation or drop.
				ids := make([]int64, 0, len(tasks))
				backup := make([]taskRepeatBackup20260711120000, 0, len(tasks))
				for _, task := range tasks {
					ids = append(ids, task.ID)
					backup = append(backup, taskRepeatBackup20260711120000{
						ID:          task.ID,
						RepeatAfter: task.RepeatAfter,
						RepeatMode:  task.RepeatMode,
					})
				}
				// Migrations run without a wrapping transaction; a mid-run failure can leave
				// this page's backup rows behind, so clear them before re-inserting to keep
				// a retry from hitting a PK conflict.
				if _, err := tx.In("id", ids).Delete(&taskRepeatBackup20260711120000{}); err != nil {
					return err
				}
				if _, err := tx.Insert(&backup); err != nil {
					return err
				}

				for _, task := range tasks {
					lastID = task.ID
					total++

					// Defensive: don't overwrite a row that already has an RRULE.
					if task.Repeats != "" {
						skipped++
						continue
					}

					rr := convertLegacyRepeatToRRule(task.RepeatAfter, task.RepeatMode)
					if rr == "" {
						// Unexpected mode / non-convertible value. Leave repeats empty;
						// the original is preserved in the backup table for recovery.
						unconvertible++
						log.Warningf("RRULE migration: task %d has legacy repeat (after=%d, mode=%d) that did not convert; left empty (original preserved in task_repeat_legacy_backup)", task.ID, task.RepeatAfter, task.RepeatMode)
						continue
					}

					repeatsFromCurrentDate := task.RepeatMode == TaskRepeatModeFromCurrentDate

					if _, err := tx.ID(task.ID).
						Cols("repeats", "repeats_from_current_date").
						Update(&taskNew20260711120000{
							Repeats:                rr,
							RepeatsFromCurrentDate: repeatsFromCurrentDate,
						}); err != nil {
						return err
					}
					converted++
				}
			}

			log.Infof("RRULE migration: %d legacy-repeat task(s) processed — %d converted, %d already had an RRULE, %d unconvertible (originals backed up to task_repeat_legacy_backup)", total, converted, skipped, unconvertible)

			// Step 3: Drop legacy columns (database-specific)
			if config.DatabaseType.GetString() == "sqlite" {
				// SQLite requires a table rebuild to drop columns. The column list must
				// track the current tasks schema (pkg/models Task) minus repeat_after/
				// repeat_mode; anything omitted here is silently lost. DROP the temp table
				// first so a retry after a mid-run failure starts from a clean slate.
				_, err = tx.Exec(`
drop table if exists tasks_dg_tmp;

create table tasks_dg_tmp
(
    id                         INTEGER           not null
        primary key autoincrement,
    title                      TEXT              not null,
    description                TEXT,
    done                       INTEGER,
    done_at                    DATETIME,
    due_date                   DATETIME,
    project_id                 INTEGER           not null,
    repeats                    TEXT,
    repeats_from_current_date  INTEGER default 0,
    priority                   INTEGER,
    start_date                 DATETIME,
    end_date                   DATETIME,
    hex_color                  TEXT,
    percent_done               REAL,
    "index"                    INTEGER default 0 not null,
    uid                        TEXT,
    cover_image_attachment_id  INTEGER default 0,
    created                    DATETIME          not null,
    updated                    DATETIME          not null,
    deleted_at                 DATETIME,
    created_by_id              INTEGER           not null
);

insert into tasks_dg_tmp(id, title, description, done, done_at, due_date, project_id, repeats, repeats_from_current_date,
                         priority, start_date, end_date, hex_color, percent_done, "index", uid,
                         cover_image_attachment_id, created, updated, deleted_at, created_by_id)
select id, title, description, done, done_at, due_date, project_id, repeats, repeats_from_current_date,
       priority, start_date, end_date, hex_color, percent_done, "index", uid,
       cover_image_attachment_id, created, updated, deleted_at, created_by_id
from tasks;

drop table tasks;

alter table tasks_dg_tmp rename to tasks;

create index IDX_tasks_deleted_at on tasks (deleted_at);
create index IDX_tasks_done on tasks (done);
create index IDX_tasks_done_at on tasks (done_at);
create index IDX_tasks_due_date on tasks (due_date);
create index IDX_tasks_end_date on tasks (end_date);
create index IDX_tasks_project_id on tasks (project_id);
create index IDX_tasks_start_date on tasks (start_date);
create unique index UQE_tasks_id on tasks (id);
create unique index UQE_tasks_project_index on tasks (project_id, "index");
`)
				return err
			}

			// MySQL and PostgreSQL can drop columns directly
			if err := dropTableColum(tx, "tasks", "repeat_after"); err != nil {
				return err
			}
			if err := dropTableColum(tx, "tasks", "repeat_mode"); err != nil {
				return err
			}

			return nil
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
