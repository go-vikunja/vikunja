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

// RegisterSoftDeletedProjectPurgeCron registers the cron job that permanently
// deletes projects whose soft-delete retention period has expired.
func RegisterSoftDeletedProjectPurgeCron() {
	err := cron.Schedule("0 * * * *", purgeSoftDeletedProjects)
	if err != nil {
		log.Errorf("Could not register soft-deleted project purge cron: %s", err.Error())
	}
}

func purgeSoftDeletedProjects() {
	cutoff := time.Now().Add(-SoftDeleteRetentionDays * 24 * time.Hour)

	s := db.NewSession()
	defer s.Close()

	var projects []*Project
	err := s.Unscoped().
		Where("deleted_at IS NOT NULL AND deleted_at < ?", cutoff).
		Find(&projects)
	if err != nil {
		log.Errorf("Could not get soft-deleted projects for purge: %s", err)
		return
	}

	if len(projects) == 0 {
		return
	}

	log.Debugf("Found %d soft-deleted projects past retention period for purge", len(projects))

	for _, p := range projects {
		func() {
			ps := db.NewSession()
			defer ps.Close()

			// Use a system user auth for the permanent delete
			doer := &user.User{ID: p.OwnerID}

			err = p.PermanentDelete(ps, doer)
			if err != nil {
				_ = ps.Rollback()
				log.Errorf("Could not permanently delete project %d: %s", p.ID, err)
				return
			}

			err = ps.Commit()
			if err != nil {
				_ = ps.Rollback()
				log.Errorf("Could not commit permanent deletion of project %d: %s", p.ID, err)
				return
			}

			log.Debugf("Permanently deleted project %d", p.ID)
		}()
	}
}
