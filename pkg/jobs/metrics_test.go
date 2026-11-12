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

	"code.vikunja.io/api/pkg/metrics"
	"github.com/stretchr/testify/require"
)

func queueCounter(t *testing.T, name, queue, kind string) float64 {
	t.Helper()
	families, err := metrics.GetRegistry().Gather()
	require.NoError(t, err)
	for _, family := range families {
		if family.GetName() != "vikunja_jobs_"+name+"_total" {
			continue
		}
		for _, metric := range family.GetMetric() {
			labels := map[string]string{}
			for _, label := range metric.GetLabel() {
				labels[label.GetName()] = label.GetValue()
			}
			if labels["queue"] == queue && labels["kind"] == kind {
				return metric.GetCounter().GetValue()
			}
		}
	}
	return 0
}

func TestQueueCountersFollowCommitAndDeduplication(t *testing.T) {
	e := queueTestEngine(t)
	now := time.Now().UTC().Truncate(time.Second)
	request := Request{Queue: t.Name(), Kind: "test", Payload: []byte(`{}`), UniqueKey: "one", RunAt: now}
	for _, commit := range []bool{false, true} {
		s := e.NewSession()
		require.NoError(t, s.Begin())
		require.NoError(t, Enqueue(s, request))
		require.NoError(t, Enqueue(s, request))
		second := request
		second.UniqueKey = "two"
		third := request
		third.UniqueKey = ""
		require.NoError(t, EnqueueBatch(s, []Request{request, second, third}))
		require.Zero(t, queueCounter(t, "enqueued", t.Name(), "test"))
		if commit {
			require.NoError(t, s.Commit())
			WakePending(s)
			require.InDelta(t, 3, queueCounter(t, "enqueued", t.Name(), "test"), 0.0001)
		} else {
			require.NoError(t, s.Rollback())
			DiscardPending(s)
			WakePending(s)
			require.Zero(t, queueCounter(t, "enqueued", t.Name(), "test"))
		}
		require.NoError(t, s.Close())
	}
	batch, err := NewQueue(e, "metrics").Claim(t.Context(), t.Name(), 1, now, time.Minute)
	require.NoError(t, err)
	require.Len(t, batch, 1)
	for _, commit := range []bool{false, true} {
		s := e.NewSession()
		require.NoError(t, s.Begin())
		require.NoError(t, Complete(s, batch[0], now))
		require.Zero(t, queueCounter(t, "completed", t.Name(), "test"))
		if commit {
			require.NoError(t, s.Commit())
			WakePending(s)
		} else {
			require.NoError(t, s.Rollback())
			DiscardPending(s)
		}
		require.NoError(t, s.Close())
	}
	require.InDelta(t, 1, queueCounter(t, "completed", t.Name(), "test"), 0.0001)
}

func TestQueueCountersCountOnlyFencedTransitions(t *testing.T) {
	e := queueTestEngine(t)
	now := time.Now().UTC().Truncate(time.Second)
	s := e.NewSession()
	defer s.Close()
	q := NewQueue(e, "metrics")
	for _, kind := range []string{"failed", "rescued"} {
		job := Job{Queue: t.Name(), Kind: kind, Payload: "{}", State: StateRunning, Attempt: 5, MaxAttempts: 5, ClaimOwner: "metrics", RunAt: now, ExpiresAt: now.Add(-time.Second)}
		_, err := s.Insert(&job)
		require.NoError(t, err)
		if kind == "failed" {
			require.NoError(t, q.Fail(t.Context(), job, now, errors.New("offline"), 5))
			require.NoError(t, q.Fail(t.Context(), job, now, errors.New("stale"), 5))
			require.InDelta(t, 1, queueCounter(t, "failed_attempts", t.Name(), kind), 0.0001)
		} else {
			count, err := q.Rescue(t.Context(), now, 10)
			require.NoError(t, err)
			require.EqualValues(t, 1, count)
			count, err = q.Rescue(t.Context(), now, 10)
			require.NoError(t, err)
			require.Zero(t, count)
			require.InDelta(t, 1, queueCounter(t, "rescued", t.Name(), kind), 0.0001)
		}
		require.InDelta(t, 1, queueCounter(t, "parked", t.Name(), kind), 0.0001)
	}
}

func TestQueueCountersDiscardRolledBackSessionReuse(t *testing.T) {
	e := queueTestEngine(t)
	s := e.NewSession()
	defer s.Close()
	request := Request{Queue: t.Name(), Kind: "test", Payload: []byte(`{}`)}
	require.NoError(t, s.Begin())
	require.NoError(t, Enqueue(s, request))
	request.UniqueKey = "rolled-back"
	require.NoError(t, Enqueue(s, request))
	require.NoError(t, s.Rollback())
	require.NoError(t, s.Begin())
	request.UniqueKey = "committed"
	require.NoError(t, Enqueue(s, request))
	require.NoError(t, s.Commit())
	WakePending(s)
	require.InDelta(t, 1, queueCounter(t, "enqueued", t.Name(), "test"), 0.0001)
	require.NoError(t, s.Begin())
	request.UniqueKey = "another-rollback"
	require.NoError(t, Enqueue(s, request))
	require.NoError(t, s.Rollback())
	require.NoError(t, s.Begin())
	require.NoError(t, s.Commit())
	WakePending(s)
	require.InDelta(t, 1, queueCounter(t, "enqueued", t.Name(), "test"), 0.0001)
}

func TestQueueUnkeyedCounterFollowsPlainCommit(t *testing.T) {
	e := queueTestEngine(t)
	s := e.NewSession()
	defer s.Close()
	require.NoError(t, s.Begin())
	require.NoError(t, Enqueue(s, Request{Queue: t.Name(), Kind: "test", Payload: []byte(`{}`)}))
	require.Zero(t, queueCounter(t, "enqueued", t.Name(), "test"))
	require.NoError(t, s.Commit())
	require.InDelta(t, 1, queueCounter(t, "enqueued", t.Name(), "test"), 0.0001)
}
