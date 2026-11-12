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
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

func TestRunnerWindowsAndRollback(t *testing.T) {
	config.InitDefaultConfig()
	log.InitLogger()
	e, err := db.CreateTestEngine()
	require.NoError(t, err)
	require.NoError(t, e.Sync(new(State))) //nolint:forbidigo // Test-only schema.
	_, err = e.Where("job = ?", t.Name()).Delete(new(State))
	require.NoError(t, err)
	r := NewRunner(e)
	start := time.Date(2026, 11, 12, 8, 0, 0, 0, time.UTC)
	if e.Dialect().URI().DBType == schemas.POSTGRES {
		start = time.Now().UTC().Truncate(time.Minute).Add(-time.Minute)
	}
	require.NoError(t, r.Register(t.Name(), start))
	require.NoError(t, r.Register(t.Name(), start.Add(time.Hour)))
	if e.Dialect().URI().DBType == schemas.POSTGRES {
		_, err = e.Where("job = ?", t.Name()).Cols("last_run_until").Update(&State{LastRunUntil: start})
		require.NoError(t, err)
	}
	calls := 0
	failure := errors.New("failed run")
	job := func(_ context.Context, _ *xorm.Session, w Window) error {
		calls++
		assert.True(t, w.From.Equal(start))
		assert.True(t, w.To.Equal(start.Add(time.Minute)))
		if calls == 1 {
			return failure
		}
		return nil
	}
	_, err = r.Tick(t.Context(), t.Name(), start.Add(time.Minute), job)
	require.ErrorIs(t, err, failure)
	ran, err := r.Tick(t.Context(), t.Name(), start.Add(time.Minute), job)
	require.NoError(t, err)
	assert.True(t, ran)
	ran, err = r.Tick(t.Context(), t.Name(), start.Add(time.Minute), job)
	require.NoError(t, err)
	assert.False(t, ran)
	ran, err = r.Tick(t.Context(), t.Name(), start, job)
	require.NoError(t, err)
	assert.False(t, ran)
	assert.Equal(t, 2, calls)
}

func TestRunnerCatchupAndOverlap(t *testing.T) {
	config.InitDefaultConfig()
	log.InitLogger()
	e, err := db.CreateTestEngine()
	require.NoError(t, err)
	require.NoError(t, e.Sync(new(State))) //nolint:forbidigo // Test-only schema.
	_, err = e.Where("job = ?", t.Name()).Delete(new(State))
	require.NoError(t, err)
	r := NewRunner(e)
	start := time.Date(2026, 11, 12, 8, 0, 0, 0, time.UTC)
	require.NoError(t, r.Register(t.Name(), start))
	if e.Dialect().URI().DBType == schemas.POSTGRES {
		start = time.Now().UTC().Add(-48 * time.Hour)
		_, err = e.Where("job = ?", t.Name()).Cols("last_run_until").Update(&State{LastRunUntil: start})
		require.NoError(t, err)
	}
	started := make(chan struct{})
	release := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		_, err := r.Tick(t.Context(), t.Name(), start.Add(48*time.Hour), func(_ context.Context, _ *xorm.Session, w Window) error {
			assert.Equal(t, CatchUpLimit, w.To.Sub(w.From))
			close(started)
			<-release
			return nil
		})
		done <- err
	}()
	<-started
	ran, err := r.Tick(t.Context(), t.Name(), start.Add(49*time.Hour), func(context.Context, *xorm.Session, Window) error {
		t.Error("overlapping run")
		return nil
	})
	require.NoError(t, err)
	assert.False(t, ran)
	close(release)
	require.NoError(t, <-done)
}

func TestRunnerPostgresExcludesOtherReplicaAcrossTicks(t *testing.T) {
	config.InitDefaultConfig()
	log.InitLogger()
	e, err := db.CreateTestEngine()
	require.NoError(t, err)
	if e.Dialect().URI().DBType != schemas.POSTGRES {
		t.Skip("PostgreSQL row-lock test")
	}
	require.NoError(t, e.Sync(new(State)))
	_, err = e.Where("job = ?", t.Name()).Delete(new(State))
	require.NoError(t, err)
	first, second := NewRunner(e), NewRunner(e)
	now := time.Now().UTC().Truncate(time.Minute).Add(-time.Minute)
	require.NoError(t, first.Register(t.Name(), now))
	_, err = e.Where("job = ?", t.Name()).Cols("last_run_until").Update(&State{LastRunUntil: now})
	require.NoError(t, err)
	started, release := make(chan struct{}), make(chan struct{})
	done := make(chan error, 1)
	go func() {
		_, err := first.Tick(t.Context(), t.Name(), now.Add(time.Minute), func(context.Context, *xorm.Session, Window) error {
			close(started)
			<-release
			return errors.New("holder rolled back")
		})
		done <- err
	}()
	<-started
	for _, tick := range []time.Time{now.Add(time.Minute), now.Add(2 * time.Minute)} {
		ran, err := second.Tick(t.Context(), t.Name(), tick, func(context.Context, *xorm.Session, Window) error { t.Error("overlapping holder"); return nil })
		require.NoError(t, err)
		assert.False(t, ran)
	}
	close(release)
	require.Error(t, <-done)
	ran, err := second.Tick(t.Context(), t.Name(), now.Add(2*time.Minute), func(_ context.Context, _ *xorm.Session, w Window) error {
		assert.True(t, w.From.Equal(now))
		return nil
	})
	require.NoError(t, err)
	assert.True(t, ran)
}
