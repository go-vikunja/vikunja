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
	"sync"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/user"
)

// runningMigrations holds the cancel func of every migration executing in this
// process, keyed by its status id. The queue is in-process, so a migration that
// is still running is always in here.
var runningMigrations = struct {
	sync.Mutex
	cancel map[int64]context.CancelFunc
}{cancel: map[int64]context.CancelFunc{}}

// StartRun makes a migration cancellable for as long as it runs. The returned
// func must be called once it ends.
func StartRun(ctx context.Context, statusID int64) (context.Context, func()) {
	ctx, cancel := context.WithCancel(ctx)

	runningMigrations.Lock()
	runningMigrations.cancel[statusID] = cancel
	runningMigrations.Unlock()

	return ctx, func() {
		runningMigrations.Lock()
		delete(runningMigrations.cancel, statusID)
		runningMigrations.Unlock()
		cancel()
	}
}

// Cancel stops the user's running migration and releases their claim so they can
// start a new one right away.
//
// The job is asked to stop rather than waited for: it aborts at its next project
// or query, and because the whole import is one transaction it rolls back
// everything it had written.
func Cancel(u *user.User) error {
	s := db.NewSession()
	defer s.Close()

	status := &Status{}
	has, err := s.Where("active_user_id = ?", u.ID).Desc("id").Get(status)
	if err != nil {
		return err
	}
	if !has {
		return &ErrNoMigrationRunning{}
	}
	if err := s.Commit(); err != nil {
		return err
	}

	runningMigrations.Lock()
	cancel, isRunning := runningMigrations.cancel[status.ID]
	runningMigrations.Unlock()

	if isRunning {
		cancel()
	} else {
		// A claim that outlived the process that took it: nothing left to stop,
		// only to release.
		log.Infof("[Migration] Migration %d of user %d is not running here, only releasing its claim", status.ID, u.ID)
	}

	return FinishMigration(status)
}
