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
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
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

// Two simultaneous claims can collide on a write lock before either row is visible;
// retrying keeps that from surfacing as a raw driver error.
const claimAttempts = 5

// ClaimMigration inserts a status row holding the account's unique migration claim.
func ClaimMigration(m MigratorName, u *user.User) (status *Status, err error) {
	for attempt := 1; ; attempt++ {
		status, err = claimMigration(m, u)

		var running *ErrMigrationAlreadyRunning
		if err == nil || attempt == claimAttempts || errors.As(err, &running) {
			return status, err
		}

		time.Sleep(time.Duration(attempt) * 10 * time.Millisecond)
	}
}

func claimMigration(m MigratorName, u *user.User) (status *Status, err error) {
	s := db.NewSession()
	defer s.Close()

	if err = releaseStaleClaims(s, u.ID); err != nil {
		return nil, claimConflictOr(u.ID, err)
	}

	// Reading before inserting keeps a losing request read-only so it cannot fight the running import.
	running, has, err := liveClaim(s, u.ID)
	if err != nil {
		_ = s.Rollback()
		return nil, claimConflictOr(u.ID, err)
	}
	if has {
		return nil, &ErrMigrationAlreadyRunning{StartedAt: running.StartedAt, MigratorName: running.MigratorName}
	}

	status = &Status{
		UserID:       u.ID,
		MigratorName: m.Name(),
		StartedAt:    time.Now(),
		ActiveUserID: &u.ID,
	}
	if _, err = s.Insert(status); err != nil {
		_ = s.Rollback()
		if conflict := claimConflict(u.ID); conflict != nil {
			return nil, conflict
		}
		if db.IsUniqueConstraintError(err, "active_user_id") {
			return nil, &ErrMigrationAlreadyRunning{}
		}
		return nil, err
	}

	return status, s.Commit()
}

// liveClaim reports the unfinished status row blocking new migrations for userID. Legacy rows
// predate active_user_id and have it NULL, but still count as live.
func liveClaim(s *xorm.Session, userID int64) (status *Status, has bool, err error) {
	status = &Status{}
	has, err = s.
		Where("finished_at IS NULL AND user_id = ?", userID).
		Desc("id").
		Get(status)
	if err != nil {
		return nil, false, err
	}
	return status, has, nil
}

// claimConflict reports the live claim of userID, if any. Re-reading is the only check that
// holds on every driver: each spells the unique violation differently, and contention can
// fail the insert with an unrelated error.
func claimConflict(userID int64) *ErrMigrationAlreadyRunning {
	s := db.NewAutocommitSession()
	defer s.Close()

	running, has, err := liveClaim(s, userID)
	if err != nil || !has {
		return nil
	}

	return &ErrMigrationAlreadyRunning{StartedAt: running.StartedAt, MigratorName: running.MigratorName}
}

func claimConflictOr(userID int64, err error) error {
	if conflict := claimConflict(userID); conflict != nil {
		return conflict
	}
	return err
}

// releaseStaleClaims unblocks migrations abandoned by a dead instance.
func releaseStaleClaims(s *xorm.Session, userID int64) error {
	timeout := config.MigrationClaimTimeout.GetDuration()
	if timeout <= 0 {
		return nil
	}

	// Probing first avoids taking a write lock on every claim attempt.
	staleBefore := time.Now().Add(-timeout)
	has, err := s.
		Where("active_user_id = ? AND started_at < ?", userID, staleBefore).
		Exist(&Status{})
	if err != nil {
		_ = s.Rollback()
		return err
	}
	if !has {
		return nil
	}

	_, err = s.
		Where("active_user_id = ? AND started_at < ?", userID, staleBefore).
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

// runningMigrations holds the cancel func of every migration executing in this
// process, keyed by its status id. The queue is in-process, so a migration that
// is still running here is always in here.
var runningMigrations = struct {
	sync.Mutex
	runs map[int64]migrationRun
}{runs: map[int64]migrationRun{}}

var migrationRunToken atomic.Int64

type migrationRun struct {
	cancel context.CancelFunc
	token  int64
}

// StartRun makes a migration cancellable for as long as it runs. The returned
// func must be called once it ends.
func StartRun(statusID int64) (context.Context, func()) {
	ctx, cancel := context.WithCancel(context.Background())
	run := migrationRun{cancel: cancel, token: migrationRunToken.Add(1)}

	runningMigrations.Lock()
	runningMigrations.runs[statusID] = run
	runningMigrations.Unlock()

	return ctx, func() {
		runningMigrations.Lock()
		// A redelivered event can run the same status twice, and only this run may unregister itself.
		if current, has := runningMigrations.runs[statusID]; has && current.token == run.token {
			delete(runningMigrations.runs, statusID)
		}
		runningMigrations.Unlock()
		cancel()
	}
}

// Cancel asks the user's running migration to stop. The job aborts at its next
// cancellation point, rolling back only the transaction it is in at that moment,
// and releases the claim itself as it exits.
func Cancel(u *user.User) error {
	s := db.NewAutocommitSession()
	defer s.Close()

	status, has, err := liveClaim(s, u.ID)
	if err != nil {
		return fmt.Errorf("could not look up the running migration of user %d: %w", u.ID, err)
	}
	if !has {
		return &ErrNoMigrationRunning{}
	}

	runningMigrations.Lock()
	run, isRunning := runningMigrations.runs[status.ID]
	runningMigrations.Unlock()

	if isRunning {
		run.cancel()
		return nil
	}

	// A claim held by another instance must not be released here: its job keeps running and
	// a second import would duplicate the data. migration.claimtimeout covers a dead instance.
	if status.ActiveUserID != nil {
		return &ErrMigrationNotCancellableHere{}
	}

	log.Infof("[Migration] Migration %d of user %d holds no claim and is not running here, releasing it", status.ID, u.ID)
	return FinishMigration(status)
}
