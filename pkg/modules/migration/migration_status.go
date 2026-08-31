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
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"

	"xorm.io/xorm"
)

// Status represents this migration status
type Status struct {
	ID           int64     `xorm:"bigint autoincr not null unique pk" json:"id" readOnly:"true" doc:"The unique, numeric id of this migration status."`
	UserID       int64     `xorm:"bigint not null" json:"-"`
	MigratorName string    `xorm:"varchar(255)" json:"migrator_name" readOnly:"true" doc:"The name of the migrator this status belongs to, e.g. \"todoist\"."`
	StartedAt    time.Time `xorm:"not null" json:"started_at" readOnly:"true" doc:"When the last migration started. Zero value if the user never migrated from this service."`
	FinishedAt   time.Time `xorm:"null" json:"finished_at" readOnly:"true" doc:"When the last migration finished. Zero value while a migration is still running or was never run."`
	// ActiveUserID's unique index serializes migrations per account; finished rows use NULL.
	ActiveUserID *int64 `xorm:"bigint null unique" json:"-"`
}

// TableName holds the table name for the migration status table
func (s *Status) TableName() string {
	return "migration_status"
}

// ClaimMigration inserts a status row holding the account's unique migration claim.
func ClaimMigration(m MigratorName, u *user.User) (status *Status, err error) {
	s := db.NewSession()
	defer s.Close()

	if err = releaseStaleClaims(s, u.ID); err != nil {
		return nil, err
	}

	// Refuse if a legacy row (created before claims existed) is still open.
	legacy := &Status{}
	has, err := s.
		Where("finished_at IS NULL AND active_user_id IS NULL AND user_id = ?", u.ID).
		Desc("id").
		Get(legacy)
	if err != nil {
		return nil, err
	}
	if has {
		return nil, &ErrMigrationAlreadyRunning{StartedAt: legacy.StartedAt, MigratorName: legacy.MigratorName}
	}

	status = &Status{
		UserID:       u.ID,
		MigratorName: m.Name(),
		StartedAt:    time.Now(),
		ActiveUserID: &u.ID,
	}
	if _, err = s.Insert(status); err != nil {
		if db.IsUniqueConstraintError(err, "active_user_id") {
			_ = s.Rollback()
			e := &ErrMigrationAlreadyRunning{}
			runningSession := db.NewSession()
			defer runningSession.Close()
			running := &Status{}
			if has, gerr := runningSession.Where("active_user_id = ?", u.ID).Desc("id").Get(running); gerr == nil && has {
				e.StartedAt = running.StartedAt
				e.MigratorName = running.MigratorName
			}
			return nil, e
		}
		_ = s.Rollback()
		return nil, err
	}

	return status, s.Commit()
}

// releaseStaleClaims unblocks migrations abandoned by a dead instance.
func releaseStaleClaims(s *xorm.Session, userID int64) error {
	timeout := config.MigrationClaimTimeout.GetDuration()
	if timeout <= 0 {
		return nil
	}

	_, err := s.
		Where("active_user_id = ? AND started_at < ?", userID, time.Now().Add(-timeout)).
		Cols("finished_at", "active_user_id").
		Update(&Status{FinishedAt: time.Now()})
	if err != nil {
		_ = s.Rollback()
		return err
	}
	return nil
}

// FinishMigration records completion and releases the user's claim.
func FinishMigration(status *Status) (err error) {
	s := db.NewSession()
	defer s.Close()

	status.FinishedAt = time.Now()
	status.ActiveUserID = nil

	// Cols is required: a plain Update skips nil pointers, so the claim would never be released.
	_, err = s.Where("id = ?", status.ID).Cols("finished_at", "active_user_id").Update(status)
	if err != nil {
		_ = s.Rollback()
		return
	}

	return s.Commit()
}

// GetMigrationStatus returns the migration status for a migration and a user
func GetMigrationStatus(m MigratorName, u *user.User) (status *Status, err error) {
	s := db.NewSession()
	defer s.Close()

	status = &Status{}
	_, err = s.
		Where("user_id = ? and migrator_name = ?", u.ID, m.Name()).
		Desc("id").
		Get(status)
	return
}

// GetMigrationStatusByID returns the migration status with the given id.
func GetMigrationStatusByID(id int64) (status *Status, err error) {
	s := db.NewSession()
	defer s.Close()

	status = &Status{}
	has, err := s.ID(id).Get(status)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, fmt.Errorf("migration status %d not found", id)
	}
	return status, nil
}
