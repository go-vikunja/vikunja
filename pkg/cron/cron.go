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
	"time"

	"code.vikunja.io/api/pkg/db"
	"xorm.io/xorm"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"

	"github.com/robfig/cron/v3"
)

var c *cron.Cron

// Init starts the cron
func Init() {
	logger := schedulerLogger{}
	c = cron.New(cron.WithChain(cron.SkipIfStillRunning(logger), cron.Recover(logger)))
	c.Start()
}

// ScheduleLocal registers process maintenance independently of the shared scheduler.
func ScheduleLocal(schedule string, f func()) error {
	_, err := c.AddFunc(schedule, f)
	return err
}

// Stop waits for running jobs after stopping new invocations.
func Stop() {
	_ = StopWithContext(context.Background())
}

// StopWithContext bounds the drain; running callbacks cannot be cancelled.
func StopWithContext(ctx context.Context) error {
	if c == nil {
		return nil
	}
	stopped := c.Stop()
	select {
	case <-stopped.Done():
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

type schedulerLogger struct{}

func (schedulerLogger) Info(message string, values ...interface{}) {
	log.Infof("[Cron] %s %v", message, values)
}

func (schedulerLogger) Error(err error, message string, values ...interface{}) {
	log.Errorf("[Cron] %s: %v %v", message, err, values)
}

// Register schedules shared work with a transactional watermark.
func Register(name, spec string, job Job) error {
	if _, err := cron.ParseStandard(spec); err != nil {
		return err
	}
	if !config.ServiceEnableScheduler.GetBool() {
		return nil
	}
	engine, err := db.CreateDBEngine()
	if err != nil {
		return err
	}
	runner := NewRunner(engine)
	if err := runner.Register(name, time.Now()); err != nil {
		return err
	}
	if err := runner.registerLag(name); err != nil {
		return err
	}
	return ScheduleLocal(spec, func() {
		if _, err := runner.Tick(context.Background(), name, time.Now(), job); err != nil {
			log.Errorf("[Cron] %s: %v", name, err)
		}
	})
}

// RegisterState coordinates callbacks which manage their own item transactions.
func RegisterState(name, spec string, job func()) error {
	return Register(name, spec, func(context.Context, *xorm.Session, Window) error {
		job()
		return nil
	})
}
