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

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

// Old task struct with legacy repeat fields for reading
type taskOld20251229100000 struct {
	ID          int64  `xorm:"bigint autoincr not null unique pk"`
	RepeatAfter int64  `xorm:"bigint INDEX null"`
	RepeatMode  int    `xorm:"not null default 0"`
	Repeats     string `xorm:"varchar(500) null"`
}

func (taskOld20251229100000) TableName() string {
	return "tasks"
}

// New task struct with RRULE fields only
type taskNew20251229100000 struct {
	ID                     int64  `xorm:"bigint autoincr not null unique pk"`
	Repeats                string `xorm:"varchar(500) null"`
	RepeatsFromCurrentDate bool   `xorm:"null default false"`
}

func (taskNew20251229100000) TableName() string {
	return "tasks"
}

// convertLegacyRepeatToRRule converts legacy repeat_after/repeat_mode to an RRULE string.
func convertLegacyRepeatToRRule(repeatAfter int64, repeatMode int) string {
	const (
		TaskRepeatModeDefault         = 0
		TaskRepeatModeMonth           = 1
		TaskRepeatModeFromCurrentDate = 2
		TaskRepeatModeYear            = 3
	)

	switch repeatMode {
	case TaskRepeatModeMonth:
		return "FREQ=MONTHLY;INTERVAL=1"
	case TaskRepeatModeYear:
		return "FREQ=YEARLY;INTERVAL=1"
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
		minute = 60
		hour   = 60 * minute
		day    = 24 * hour
		week   = 7 * day
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
		ID:          "20251229100000",
		Description: "Replace legacy repeat fields with RRULE",
		Migrate: func(tx *xorm.Engine) error {
			const TaskRepeatModeFromCurrentDate = 2

			// Step 1: Add new RRULE columns
			err := tx.Sync2(taskNew20251229100000{})
			if err != nil {
				return err
			}

			// Step 2: Migrate existing legacy repeat data to RRULE format
			var tasks []taskOld20251229100000
			err = tx.Where("repeat_after > 0 OR repeat_mode > 0").Find(&tasks)
			if err != nil {
				return err
			}

			log.Infof("Migrating %d tasks with legacy repeat settings to RRULE", len(tasks))

			for _, task := range tasks {
				// Skip if already has RRULE (shouldn't happen, but defensive)
				if task.Repeats != "" {
					continue
				}

				rrule := convertLegacyRepeatToRRule(task.RepeatAfter, task.RepeatMode)
				if rrule == "" {
					continue
				}

				repeatsFromCurrentDate := task.RepeatMode == TaskRepeatModeFromCurrentDate

				_, err := tx.ID(task.ID).
					Cols("repeats", "repeats_from_current_date").
					Update(&taskNew20251229100000{
						Repeats:                rrule,
						RepeatsFromCurrentDate: repeatsFromCurrentDate,
					})
				if err != nil {
					return err
				}
			}

			// Step 3: Drop legacy columns (database-specific)
			if config.DatabaseType.GetString() == "sqlite" {
				// SQLite requires table rebuild to drop columns
				_, err = tx.Exec(`
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
    created_by_id              INTEGER           not null
);

insert into tasks_dg_tmp(id, title, description, done, done_at, due_date, project_id, repeats, repeats_from_current_date,
                         priority, start_date, end_date, hex_color, percent_done, "index", uid,
                         cover_image_attachment_id, created, updated, created_by_id)
select id, title, description, done, done_at, due_date, project_id, repeats, repeats_from_current_date,
       priority, start_date, end_date, hex_color, percent_done, "index", uid,
       cover_image_attachment_id, created, updated, created_by_id
from tasks;

drop table tasks;

alter table tasks_dg_tmp rename to tasks;

create index IDX_tasks_done on tasks (done);
create index IDX_tasks_done_at on tasks (done_at);
create index IDX_tasks_due_date on tasks (due_date);
create index IDX_tasks_end_date on tasks (end_date);
create index IDX_tasks_project_id on tasks (project_id);
create index IDX_tasks_start_date on tasks (start_date);
create unique index UQE_tasks_id on tasks (id);
`)
				return err
			}

			// MySQL and PostgreSQL can drop columns directly
			if err := dropTableColum(tx, "tasks", "repeat_after"); err != nil {
				log.Warningf("Could not drop repeat_after column: %v", err)
			}
			if err := dropTableColum(tx, "tasks", "repeat_mode"); err != nil {
				log.Warningf("Could not drop repeat_mode column: %v", err)
			}

			return nil
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
