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

package jobs

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm/contexts"
)

func TestMaintenanceDrainsMultipleBatches(t *testing.T) {
	e := queueTestEngine(t)
	now := time.Now().UTC().Truncate(time.Second)
	rows := make([]Job, 0, 1004)
	for range 1001 {
		rows = append(rows, Job{Queue: t.Name(), Kind: "expired", Version: 1, Payload: "{}", State: StateRunning, Attempt: 1, MaxAttempts: 5, RunAt: now.Add(-time.Minute), ClaimOwner: "old", ExpiresAt: now.Add(-time.Second)})
	}
	for offset := 0; offset < len(rows); offset += 50 {
		batch := rows[offset:min(offset+50, len(rows))]
		_, err := e.Insert(&batch)
		require.NoError(t, err)
	}
	protected := []Job{
		{Queue: t.Name(), Kind: "parked", Version: 1, Payload: "{}", State: StateFailed, RunAt: now.Add(-72 * time.Hour), FinishedAt: now.Add(-72 * time.Hour)},
		{Queue: t.Name(), Kind: "live", Version: 1, Payload: "{}", State: StateRunning, RunAt: now, ClaimOwner: "live", ExpiresAt: now.Add(time.Hour)},
		{Queue: t.Name(), Kind: "exhausted", Version: 1, Payload: "{}", State: StateRunning, Attempt: 5, MaxAttempts: 5, RunAt: now, ClaimOwner: "old", ExpiresAt: now.Add(-time.Second)},
	}
	_, err := e.Insert(&protected)
	require.NoError(t, err)
	require.NoError(t, Maintain(t.Context()))
	for state, want := range map[string]int64{StateAvailable: 1001, StateRunning: 1, StateFailed: 2} {
		count, err := e.Where("queue = ? AND state = ?", t.Name(), state).Count(new(Job))
		require.NoError(t, err)
		assert.Equal(t, want, count, state)
	}
}

func TestMaintenanceHonorsCancellation(t *testing.T) {
	queueTestEngine(t)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	require.ErrorIs(t, Maintain(ctx), context.Canceled)
}

type maintenancePause struct{}

type maintenancePauseHook struct{}

func (maintenancePauseHook) BeforeProcess(c *contexts.ContextHook) (context.Context, error) {
	if pause, ok := c.Ctx.Value(maintenancePause{}).(bool); ok && pause && strings.HasPrefix(c.SQL, "SELECT") && strings.Contains(c.SQL, "expires_at") {
		<-c.Ctx.Done()
		return c.Ctx, c.Ctx.Err()
	}
	return c.Ctx, nil
}

func (maintenancePauseHook) AfterProcess(c *contexts.ContextHook) error { return c.Err }

func TestMaintenanceCancellationDuringRescue(t *testing.T) {
	e := queueTestEngine(t)
	e.AddHook(maintenancePauseHook{})
	ctx, cancel := context.WithTimeout(context.WithValue(t.Context(), maintenancePause{}, true), 20*time.Millisecond)
	defer cancel()
	require.ErrorIs(t, Maintain(ctx), context.DeadlineExceeded)
}

func TestMaintenanceDrainStopsAtBudget(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	calls := 0
	require.NoError(t, drainMaintenance(ctx, 20*time.Millisecond, func(context.Context) (int64, error) { calls++; return 1, nil }))
	require.Greater(t, calls, 1)
}

func TestMaintenanceDrainPreservesErrors(t *testing.T) {
	expected := errors.New("database failed")
	calls := 0
	err := drainMaintenance(t.Context(), time.Second, func(context.Context) (int64, error) { calls++; return 0, expected })
	require.ErrorIs(t, err, expected)
	require.Equal(t, 1, calls)
}
