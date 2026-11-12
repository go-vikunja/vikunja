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

func TestRunnerPostgresClockIgnoresCallerSkew(t *testing.T) {
	config.InitDefaultConfig()
	log.InitLogger()
	e, err := db.CreateTestEngine()
	require.NoError(t, err)
	if e.Dialect().URI().DBType != schemas.POSTGRES {
		t.Skip("PostgreSQL database clock")
	}
	require.NoError(t, e.Sync(new(State)))
	previous := e.DB().Stats().MaxOpenConnections
	e.SetMaxOpenConns(1)
	t.Cleanup(func() { _, err := e.Exec("RESET TIME ZONE"); require.NoError(t, err); e.SetMaxOpenConns(previous) })
	_, err = e.Exec("SET TIME ZONE 'Pacific/Auckland'")
	require.NoError(t, err)
	dbNow := func() time.Time {
		var row struct{ Now time.Time }
		_, err := e.SQL("SELECT clock_timestamp() AT TIME ZONE 'UTC' AS now").Get(&row)
		require.NoError(t, err)
		return row.Now
	}
	for _, skew := range []time.Duration{-3 * time.Minute, 3 * time.Minute} {
		t.Run(skew.String(), func(t *testing.T) {
			name := t.Name()
			_, err := e.Where("job = ?", name).Delete(new(State))
			require.NoError(t, err)
			t.Cleanup(func() { _, err := e.Where("job = ?", name).Delete(new(State)); require.NoError(t, err) })
			r := NewRunner(e)
			before := dbNow()
			require.NoError(t, r.Register(name, before.Add(skew)))
			var initial State
			_, err = e.Where("job = ?", name).Get(&initial)
			require.NoError(t, err)
			assert.WithinDuration(t, before, initial.LastRunUntil, 2*time.Second)
			require.NoError(t, r.Register(name, before.Add(time.Hour)))
			var preserved State
			preserved = State{}
			_, err = e.Where("job = ?", name).Get(&preserved)
			require.NoError(t, err)
			assert.True(t, initial.LastRunUntil.Equal(preserved.LastRunUntil))
			start := before.Truncate(time.Minute).Add(-2 * time.Minute)
			_, err = e.Where("job = ?", name).Cols("last_run_until").Update(&State{LastRunUntil: start})
			require.NoError(t, err)
			var window Window
			ran, err := r.Tick(t.Context(), name, before.Add(skew), func(_ context.Context, _ *xorm.Session, w Window) error { window = w; return nil })
			require.NoError(t, err)
			require.True(t, ran)
			assert.True(t, window.From.Equal(start))
			assert.False(t, window.To.Before(before.Truncate(time.Minute)))
			assert.False(t, window.To.After(dbNow().Truncate(time.Minute)))
			ahead := dbNow().Add(time.Hour).Truncate(time.Minute)
			_, err = e.Where("job = ?", name).Cols("last_run_until").Update(&State{LastRunUntil: ahead})
			require.NoError(t, err)
			ran, err = r.Tick(t.Context(), name, ahead.Add(time.Hour), func(context.Context, *xorm.Session, Window) error {
				t.Error("future cursor must not run or rewind")
				return nil
			})
			require.NoError(t, err)
			require.False(t, ran)
			preserved = State{}
			_, err = e.Where("job = ?", name).Get(&preserved)
			require.NoError(t, err)
			assert.True(t, ahead.Equal(preserved.LastRunUntil))
		})
	}
}
