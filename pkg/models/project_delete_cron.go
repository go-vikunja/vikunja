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

// ProjectDeleteRetention is how long soft-deleted projects are kept before
// permanent removal. Same grace period as tasks.
const ProjectDeleteRetention = 30 * 24 * time.Hour

// RegisterProjectCleanupCron registers the cron job that permanently removes
// projects which were soft-deleted more than ProjectDeleteRetention ago.
func RegisterProjectCleanupCron() {
	err := cron.Schedule("0 * * * *", func() {
		deleteExpiredProjects(time.Now())
	})
	if err != nil {
		log.Errorf("Could not register project cleanup cron: %s", err.Error())
	}
}

func deleteExpiredProjects(now time.Time) {
	s := db.NewSession()
	projects := []*Project{}
	err := s.Unscoped().
		Where(builder.And(
			builder.NotNull{"deleted_at"},
			builder.Lt{"deleted_at": now.Add(-ProjectDeleteRetention)},
		)).
		Find(&projects)
	s.Close()
	if err != nil {
		log.Errorf("Could not get projects scheduled for permanent deletion: %s", err)
		return
	}

	if len(projects) == 0 {
		return
	}

	log.Debugf("Found %d projects scheduled for permanent deletion", len(projects))

	for _, project := range projects {
		func() {
			ps := db.NewSession()
			defer ps.Close()

			err = project.PermanentDelete(ps)
			if err != nil {
				_ = ps.Rollback()
				log.Errorf("Could not permanently delete project %d: %s", project.ID, err)
				return
			}

			log.Debugf("Permanently deleted project %d", project.ID)

			err = ps.Commit()
			if err != nil {
				log.Errorf("Could not commit transaction: %s", err)
			}
		}()
	}
}
