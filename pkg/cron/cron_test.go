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
	"sync/atomic"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"

	"github.com/robfig/cron/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupScheduler(t *testing.T, enabled bool) {
	t.Helper()
	log.InitLogger()
	config.InitDefaultConfig()
	config.ServiceEnableScheduler.Set(enabled)
	Init()
	t.Cleanup(func() {
		Stop()
		config.ResetForTests()
	})
}

func TestCronDisabledDoesNotRegisterSharedJobs(t *testing.T) {
	setupScheduler(t, false)
	require.NoError(t, Schedule("* * * * *", func() { t.Error("shared job ran") }))
	assert.Empty(t, c.Entries())
}

func TestCronDisabledKeepsLocalMaintenance(t *testing.T) {
	setupScheduler(t, false)
	runs := 0
	require.NoError(t, ScheduleLocal("0 0 1 1 *", func() { runs++ }))
	entries := c.Entries()
	require.Len(t, entries, 1)
	entries[0].WrappedJob.Run()
	assert.Equal(t, 1, runs)
}

func TestCronDisabledRejectsInvalidSchedule(t *testing.T) {
	setupScheduler(t, false)
	require.Error(t, Schedule("invalid", func() {}))
}

func TestCronSkipsOverlappingInvocations(t *testing.T) {
	setupScheduler(t, true)
	started := make(chan struct{}, 2)
	release := make(chan struct{})
	var calls atomic.Int64
	require.NoError(t, Schedule("0 0 1 1 *", func() {
		calls.Add(1)
		started <- struct{}{}
		<-release
	}))
	job := c.Entries()[0].WrappedJob
	firstDone := make(chan struct{})
	go func() {
		job.Run()
		close(firstDone)
	}()
	<-started
	secondDone := make(chan struct{})
	go func() {
		job.Run()
		close(secondDone)
	}()
	select {
	case <-secondDone:
	case <-time.After(time.Second):
		t.Error("overlapping invocation did not return promptly")
	}
	assert.EqualValues(t, 1, calls.Load())
	close(release)
	<-firstDone
	<-secondDone
	job.Run()
	assert.EqualValues(t, 2, calls.Load(), "the next invocation runs after its predecessor finishes")
}

func TestCronRecoversPanicsBeforeAllowingNextRun(t *testing.T) {
	setupScheduler(t, true)
	calls := 0
	require.NoError(t, Schedule("0 0 1 1 *", func() {
		calls++
		if calls == 1 {
			panic("failed job")
		}
	}))
	job := c.Entries()[0].WrappedJob
	assert.NotPanics(t, job.Run)
	assert.NotPanics(t, job.Run)
	assert.Equal(t, 2, calls, "a recovered panic must not leave the job locked")
}

func TestCronDifferentJobsRunConcurrently(t *testing.T) {
	setupScheduler(t, true)
	started := make(chan struct{}, 2)
	release := make(chan struct{})
	for range 2 {
		require.NoError(t, Schedule("0 0 1 1 *", func() {
			started <- struct{}{}
			<-release
		}))
	}
	finished := make(chan struct{}, 2)
	for _, entry := range c.Entries() {
		go func() {
			entry.WrappedJob.Run()
			finished <- struct{}{}
		}()
	}
	for range 2 {
		select {
		case <-started:
		case <-time.After(time.Second):
			t.Error("independent job was blocked")
		}
	}
	close(release)
	for range 2 {
		<-finished
	}
}

type onceSchedule struct{ when time.Time }

func (s onceSchedule) Next(now time.Time) time.Time {
	if now.Before(s.when) {
		return s.when
	}
	return time.Time{}
}

func TestCronStopWaitsForRunningJob(t *testing.T) {
	setupScheduler(t, true)
	started := make(chan struct{})
	release := make(chan struct{})
	c.Schedule(onceSchedule{when: time.Now().Add(20 * time.Millisecond)}, cron.FuncJob(func() {
		close(started)
		<-release
	}))
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("job did not start")
	}
	stopped := make(chan struct{})
	go func() {
		Stop()
		close(stopped)
	}()
	select {
	case <-stopped:
		t.Error("Stop returned while a job was running")
	case <-time.After(25 * time.Millisecond):
	}
	close(release)
	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("Stop did not return after the job finished")
	}
}

func TestCronStopWithContextBoundsRunningJob(t *testing.T) {
	setupScheduler(t, true)
	started := make(chan struct{})
	release := make(chan struct{})
	defer close(release)
	c.Schedule(onceSchedule{when: time.Now().Add(20 * time.Millisecond)}, cron.FuncJob(func() {
		close(started)
		<-release
	}))
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("job did not start")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	assert.ErrorIs(t, StopWithContext(ctx), context.DeadlineExceeded)
}
