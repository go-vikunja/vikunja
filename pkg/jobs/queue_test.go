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
	"errors"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

func TestQueueTransactionAndOwnership(t *testing.T) {
	e := queueTestEngine(t)
	q := NewQueue(e, "first")
	now := time.Now().UTC().Truncate(time.Second)
	req := Request{Queue: t.Name(), Kind: "test", Payload: []byte(`{}`), UniqueKey: "one", RunAt: now}
	s := e.NewSession()
	defer s.Close()
	require.NoError(t, s.Begin())
	require.NoError(t, Enqueue(s, req))
	require.NoError(t, s.Rollback())
	got, err := q.Claim(t.Context(), t.Name(), 1, now, time.Minute)
	require.NoError(t, err)
	require.Empty(t, got)
	s = e.NewSession()
	defer s.Close()
	require.NoError(t, s.Begin())
	require.NoError(t, Enqueue(s, req))
	require.NoError(t, Enqueue(s, req))
	require.NoError(t, s.Commit())
	got, err = q.Claim(t.Context(), t.Name(), 2, now, time.Minute)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, 1, got[0].Attempt)
	stale := got[0]
	count, err := q.Rescue(t.Context(), now.Add(61*time.Second), 100)
	require.NoError(t, err)
	assert.EqualValues(t, 1, count)
	q2 := NewQueue(e, "second")
	got, err = q2.Claim(t.Context(), t.Name(), 1, now.Add(61*time.Second), time.Minute)
	require.NoError(t, err)
	require.Len(t, got, 1)
	s = e.NewSession()
	defer s.Close()
	require.NoError(t, s.Begin())
	require.ErrorIs(t, Complete(s, stale, now), ErrStaleClaim)
	require.NoError(t, s.Rollback())
	require.NoError(t, q2.Fail(t.Context(), got[0], now, errors.New("permanent"), 2))
	row := Job{}
	_, err = e.ID(got[0].ID).Get(&row)
	require.NoError(t, err)
	assert.Equal(t, StateFailed, row.State)
	assert.Equal(t, "permanent", row.LastError)
	count, err = q.Cleanup(t.Context(), now.Add(72*time.Hour), 48*time.Hour, 100)
	require.NoError(t, err)
	assert.Zero(t, count, "parked jobs must remain")
}

func TestQueueRenewalCompletionAndRetention(t *testing.T) {
	e := queueTestEngine(t)
	q := NewQueue(e, "renewal")
	now := time.Now().UTC().Truncate(time.Second)
	req := Request{Queue: t.Name(), Kind: "test", Payload: []byte(`{}`), UniqueKey: "sent", RunAt: now}
	s := e.NewSession()
	defer s.Close()
	require.NoError(t, Enqueue(s, req))
	batch, err := q.Claim(t.Context(), t.Name(), 1, now, time.Minute)
	require.NoError(t, err)
	require.Len(t, batch, 1)
	count, err := q.Renew(t.Context(), batch, now.Add(30*time.Second), time.Minute)
	require.NoError(t, err)
	assert.EqualValues(t, 1, count)
	count, err = q.Rescue(t.Context(), now.Add(70*time.Second), 100)
	require.NoError(t, err)
	assert.Zero(t, count)
	require.NoError(t, s.Begin())
	require.NoError(t, Complete(s, batch[0], now.Add(70*time.Second)))
	require.NoError(t, s.Rollback())
	stored := Job{}
	_, err = e.ID(batch[0].ID).Get(&stored)
	require.NoError(t, err)
	assert.Equal(t, StateRunning, stored.State, "completion rolls back with effects")
	s = e.NewSession()
	defer s.Close()
	require.NoError(t, Complete(s, batch[0], now.Add(70*time.Second)))
	require.NoError(t, Enqueue(s, req))
	count, err = e.Where("queue = ?", t.Name()).Count(new(Job))
	require.NoError(t, err)
	assert.EqualValues(t, 1, count)
	count, err = q.Renew(t.Context(), batch, now.Add(75*time.Second), time.Minute)
	require.NoError(t, err)
	assert.Zero(t, count)
	count, err = q.Cleanup(t.Context(), now.Add(47*time.Hour), 48*time.Hour, 100)
	require.NoError(t, err)
	assert.Zero(t, count)
	count, err = q.Cleanup(t.Context(), now.Add(49*time.Hour), 48*time.Hour, 100)
	require.NoError(t, err)
	assert.EqualValues(t, 1, count)
}

func TestQueueUsesUTCForLocalClock(t *testing.T) {
	e := queueTestEngine(t)
	q := NewQueue(e, "local-clock")
	now := time.Now().In(time.FixedZone("east", 3600)).Truncate(time.Second)
	s := e.NewSession()
	defer s.Close()
	require.NoError(t, Enqueue(s, Request{Queue: t.Name(), Kind: "test", Payload: []byte(`{}`), UniqueKey: "local", RunAt: now}))
	batch, err := q.Claim(t.Context(), t.Name(), 1, now, time.Minute)
	require.NoError(t, err)
	require.Len(t, batch, 1)
	require.NoError(t, Complete(s, batch[0], now))
	require.NoError(t, Enqueue(s, Request{Queue: t.Name(), Kind: "test", Payload: []byte(`{}`), UniqueKey: "later", RunAt: now.Add(time.Minute)}))
	batch, err = q.Claim(t.Context(), t.Name(), 1, now, time.Minute)
	require.NoError(t, err)
	require.Empty(t, batch, "future work must stay scheduled")
}

func queueTestEngine(t *testing.T) *xorm.Engine {
	t.Helper()
	config.InitDefaultConfig()
	e, err := db.CreateTestEngine()
	require.NoError(t, err)
	require.NoError(t, e.Sync(new(Job)))
	clearQueue := func() { _, err := e.Where("queue = ?", t.Name()).Delete(new(Job)); require.NoError(t, err) }
	clearQueue()
	t.Cleanup(clearQueue)
	return e
}
