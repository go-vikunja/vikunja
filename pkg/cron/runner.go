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

package cron

import (
	"context"
	"errors"
	"sync"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/log"
	"xorm.io/builder"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

const CatchUpLimit = 24 * time.Hour

type State struct {
	Job          string    `xorm:"varchar(100) pk 'job'"`
	LastRunUntil time.Time `xorm:"datetime not null"`
}

func (State) TableName() string { return "cron_state" }

func init() { db.RegisterTables([]interface{}{new(State)}) }

// Window includes To and excludes From.
type Window struct{ From, To time.Time }
type Job func(context.Context, *xorm.Session, Window) error

type Runner struct {
	engine  *xorm.Engine
	running sync.Map
}

func NewRunner(engine *xorm.Engine) *Runner { return &Runner{engine: engine} }

func (r *Runner) Register(name string, now time.Time) error {
	s := r.engine.NewSession()
	defer s.Close()
	// Concurrent startup must preserve the existing watermark.
	query := "INSERT INTO cron_state (job, last_run_until) VALUES (?, ?) ON CONFLICT (job) DO NOTHING"
	if r.engine.Dialect().URI().DBType == schemas.POSTGRES {
		_, err := s.Exec("INSERT INTO cron_state (job, last_run_until) VALUES (?, clock_timestamp() AT TIME ZONE 'UTC') ON CONFLICT (job) DO NOTHING", name)
		return err
	}
	if r.engine.Dialect().URI().DBType == schemas.MYSQL {
		query = "INSERT INTO cron_state (job, last_run_until) VALUES (?, ?) ON DUPLICATE KEY UPDATE job = job"
	}
	_, err := s.Exec(query, name, now.UTC().Format("2006-01-02 15:04:05"))
	return err
}

func (r *Runner) Tick(ctx context.Context, name string, to time.Time, job Job) (bool, error) {
	lock, _ := r.running.LoadOrStore(name, &sync.Mutex{})
	mu := lock.(*sync.Mutex)
	if !mu.TryLock() {
		return false, nil
	}
	defer mu.Unlock()
	to = to.UTC().Truncate(time.Minute)
	s := r.engine.NewSession()
	defer s.Close()
	defer events.CleanupPending(s)
	db.SetSessionContext(ctx, s)
	state := State{}
	postgres := r.engine.Dialect().URI().DBType == schemas.POSTGRES
	if !postgres {
		found, err := r.engine.Where(builder.Eq{"job": name}).Get(&state)
		if err != nil {
			return false, err
		}
		if !found {
			return false, errors.New("cron job is not registered")
		}
	}
	if err := s.Begin(); err != nil {
		return false, err
	}
	if postgres {
		var locked struct {
			LastRunUntil time.Time
			DatabaseNow  time.Time
		}
		found, err := s.SQL("SELECT last_run_until, clock_timestamp() AT TIME ZONE 'UTC' AS database_now FROM cron_state WHERE job = ? FOR UPDATE SKIP LOCKED", name).Get(&locked)
		if err != nil || !found {
			return false, err
		}
		state.LastRunUntil = locked.LastRunUntil
		to = locked.DatabaseNow.UTC().Truncate(time.Minute)
		// State-driven callbacks may work through other sessions for minutes.
		if _, err = s.Exec("SELECT set_config('idle_in_transaction_session_timeout', '0', true) WHERE current_setting('server_version_num')::integer >= 90600"); err != nil {
			return false, err
		}
	}
	if !to.After(state.LastRunUntil) {
		return false, nil
	}
	window := Window{From: state.LastRunUntil, To: to}
	if to.Sub(window.From) > CatchUpLimit {
		log.Warningf("[Cron] %s skipped work before %s (catch-up cap)", name, to.Add(-CatchUpLimit))
		window.From = to.Add(-CatchUpLimit)
	}
	if err := job(ctx, s, window); err != nil {
		return false, err
	}
	if _, err := s.Where(builder.Eq{"job": name}).Cols("last_run_until").Update(&State{LastRunUntil: to}); err != nil {
		return false, err
	}
	if err := s.Commit(); err != nil {
		return false, err
	}
	events.DispatchPending(ctx, s)
	return true, nil
}
