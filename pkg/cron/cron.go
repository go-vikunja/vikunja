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

// Schedule registers shared work on instances with the scheduler enabled.
func Schedule(schedule string, f func()) error {
	if !config.ServiceEnableScheduler.GetBool() {
		_, err := cron.ParseStandard(schedule)
		return err
	}
	return ScheduleLocal(schedule, f)
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
