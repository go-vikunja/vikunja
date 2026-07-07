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

package models

import (
	"time"

	"code.vikunja.io/api/pkg/cron"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"

	"xorm.io/builder"
)

// taskDeleteRetention is how long soft-deleted tasks are kept before permanent
// removal. Hard-coded like the user deletion grace period.
const taskDeleteRetention = 30 * 24 * time.Hour

// RegisterTaskCleanupCron registers the cron job that permanently removes
// tasks which were soft-deleted more than taskDeleteRetention ago.
func RegisterTaskCleanupCron() {
	err := cron.Schedule("0 * * * *", func() {
		deleteExpiredTasks(time.Now())
	})
	if err != nil {
		log.Errorf("Could not register task cleanup cron: %s", err.Error())
	}
}

func deleteExpiredTasks(now time.Time) {
	s := db.NewSession()
	tasks := []*Task{}
	err := s.Unscoped().
		Where(builder.And(
			builder.NotNull{"deleted_at"},
			builder.Lt{"deleted_at": now.Add(-taskDeleteRetention)},
		)).
		Find(&tasks)
	s.Close()
	if err != nil {
		log.Errorf("Could not get tasks scheduled for permanent deletion: %s", err)
		return
	}

	if len(tasks) == 0 {
		return
	}

	log.Debugf("Found %d tasks scheduled for permanent deletion", len(tasks))

	for _, task := range tasks {
		func() {
			ts := db.NewSession()
			defer ts.Close()

			err = hardDeleteTask(ts, task)
			if err != nil {
				_ = ts.Rollback()
				log.Errorf("Could not permanently delete task %d: %s", task.ID, err)
				return
			}

			log.Debugf("Permanently deleted task %d", task.ID)

			err = ts.Commit()
			if err != nil {
				log.Errorf("Could not commit transaction: %s", err)
			}
		}()
	}
}
