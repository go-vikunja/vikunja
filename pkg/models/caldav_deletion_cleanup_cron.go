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
)

// RegisterCaldavDeletionCleanupCron registers a daily cron job that removes
// task_caldav_deletions records older than 30 days. CalDAV sync tokens are
// only useful within a reasonable window; clients that haven't synced in over
// a month will receive a valid-sync-token error and fall back to a full resync.
func RegisterCaldavDeletionCleanupCron() {
	err := cron.Schedule("0 0 * * *", func() {
		s := db.NewSession()
		defer s.Close()

		cutoff := time.Now().Add(-30 * 24 * time.Hour)
		if err := CleanupOldCaldavDeletions(s, cutoff); err != nil {
			_ = s.Rollback()
			log.Errorf("[CalDAV deletion cleanup] Could not clean up old CalDAV deletions: %s", err)
			return
		}
		if err := s.Commit(); err != nil {
			log.Errorf("[CalDAV deletion cleanup] Could not commit CalDAV deletion cleanup: %s", err)
		}
	})
	if err != nil {
		log.Fatalf("Could not register CalDAV deletion cleanup cron: %s", err)
	}
}
