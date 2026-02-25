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
	"code.vikunja.io/api/pkg/user"
)

// RegisterAutoTaskCron registers a cron job that runs every minute to check
// all active auto-task templates across all users and create task instances
// that are due. This ensures auto-tasks are created reliably without requiring
// a frontend trigger (e.g. the user visiting the auto-task page).
func RegisterAutoTaskCron() {
	err := cron.Schedule("* * * * *", func() {
		s := db.NewSession()
		defer s.Close()

		if err := s.Begin(); err != nil {
			log.Errorf("[Auto-Task Cron] Could not start session: %s", err)
			return
		}

		now := time.Now()

		// Find all distinct owner IDs that have active templates due now
		type ownerInfo struct {
			OwnerID int64 `xorm:"owner_id"`
		}
		owners := make([]*ownerInfo, 0)
		err := s.SQL(
			"SELECT DISTINCT owner_id FROM auto_task_templates WHERE active = ? AND next_due_at <= ?",
			true, now,
		).Find(&owners)
		if err != nil {
			log.Errorf("[Auto-Task Cron] Could not query due templates: %s", err)
			_ = s.Rollback()
			return
		}

		if len(owners) == 0 {
			_ = s.Rollback()
			return
		}

		log.Debugf("[Auto-Task Cron] Found %d users with due auto-task templates", len(owners))

		totalCreated := 0
		for _, o := range owners {
			u := &user.User{ID: o.OwnerID}
			has, err := s.ID(o.OwnerID).Get(u)
			if err != nil || !has {
				log.Errorf("[Auto-Task Cron] Could not load user %d: %v", o.OwnerID, err)
				continue
			}

			created, err := CheckAndCreateAutoTasks(s, u)
			if err != nil {
				log.Errorf("[Auto-Task Cron] Error checking auto-tasks for user %d: %s", o.OwnerID, err)
				continue
			}

			if len(created) > 0 {
				log.Debugf("[Auto-Task Cron] Created %d tasks for user %d (%s)", len(created), u.ID, u.Username)
				totalCreated += len(created)
			}
		}

		if err := s.Commit(); err != nil {
			log.Errorf("[Auto-Task Cron] Could not commit: %s", err)
			return
		}

		if totalCreated > 0 {
			log.Infof("[Auto-Task Cron] Created %d auto-task instances across %d users", totalCreated, len(owners))
		}
	})
	if err != nil {
		log.Fatalf("Could not register auto-task cron: %s", err)
	}
}
