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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/builder"
)

func TestQueueRetryScheduleUsesConfiguration(t *testing.T) {
	e := queueTestEngine(t)
	prefix := "jobs.queues." + t.Name() + "."
	config.Key(prefix + "maxattempts").Set(7)
	config.Key(prefix + "retryinitialdelay").Set("45s")
	config.Key(prefix + "retrymaxdelay").Set("2m")
	now := time.Now().UTC().Truncate(time.Second)
	session := e.NewSession()
	defer session.Close()
	for _, key := range []string{"", "keyed"} {
		require.NoError(t, Enqueue(session, Request{
			Queue:     t.Name(),
			Kind:      "test",
			Payload:   []byte(`{}`),
			UniqueKey: key,
			RunAt:     now,
		}))
	}
	queue := NewQueue(e, "retry")
	for _, delay := range []time.Duration{45 * time.Second, 90 * time.Second, 2 * time.Minute} {
		batch, err := queue.Claim(t.Context(), t.Name(), 2, now, time.Minute)
		require.NoError(t, err)
		require.Len(t, batch, 2)
		for _, job := range batch {
			assert.Equal(t, 7, job.MaxAttempts)
			require.NoError(t, queue.Fail(t.Context(), job, now, errors.New("offline"), job.MaxAttempts))
			var stored Job
			found, err := session.ID(job.ID).Get(&stored)
			require.NoError(t, err)
			require.True(t, found)
			assert.Equal(t, StateAvailable, stored.State)
			assert.False(t, stored.RunAt.Before(now.Add(delay/2)), "retry before jitter lower bound")
			assert.False(t, stored.RunAt.After(now.Add(delay)), "retry above capped backoff")
		}
		now = now.Add(delay)
	}
}

func TestQueueRetryDefaults(t *testing.T) {
	config.InitDefaultConfig()
	for _, queue := range []string{"mail", "bulk-mail", "webhook-delivery"} {
		prefix := "jobs.queues." + queue + "."
		assert.Equal(t, 12, config.Key(prefix+"maxattempts").GetInt())
		assert.Equal(t, 30*time.Second, config.Key(prefix+"retryinitialdelay").GetDuration())
		assert.Equal(t, time.Hour, config.Key(prefix+"retrymaxdelay").GetDuration())
	}
	assert.Equal(t, 16, queueOptions("webhook-delivery").Concurrency)
	assert.Equal(t, 2, queueOptions("webhooks").Concurrency)
}

func TestQueueParkedTimeRecordedOnFailureAndRescue(t *testing.T) {
	e := queueTestEngine(t)
	now := time.Now().UTC().Truncate(time.Second)
	session := e.NewSession()
	defer session.Close()
	job := Job{
		Queue:       t.Name(),
		Kind:        "test",
		Payload:     "{}",
		State:       StateRunning,
		Attempt:     5,
		MaxAttempts: 5,
		ClaimOwner:  "worker",
		RunAt:       now,
		ExpiresAt:   now.Add(-time.Second),
	}
	_, err := session.Insert(&job)
	require.NoError(t, err)
	queue := NewQueue(e, "worker")
	require.NoError(t, queue.Fail(t.Context(), job, now, errors.New("offline"), job.MaxAttempts))
	var failed Job
	_, err = session.ID(job.ID).Get(&failed)
	require.NoError(t, err)
	assert.Equal(t, StateFailed, failed.State)
	assert.True(t, failed.FinishedAt.Equal(now))
	_, err = session.ID(job.ID).Cols("state", "finished_at").Update(&Job{State: StateRunning})
	require.NoError(t, err)
	count, err := queue.Rescue(t.Context(), now, 1)
	require.NoError(t, err)
	require.EqualValues(t, 1, count)
	var rescued Job
	_, err = session.Where(builder.Eq{"id": job.ID}).Get(&rescued)
	require.NoError(t, err)
	assert.True(t, rescued.FinishedAt.Equal(now))
}

func TestQueueRetryJitterSpreadsWithinBounds(t *testing.T) {
	policy := RetryPolicy{Attempts: 12, InitialDelay: 30 * time.Second, MaxDelay: time.Hour}
	for _, test := range []struct {
		attempt int
		base    time.Duration
	}{
		{1, 30 * time.Second}, {2, time.Minute}, {8, time.Hour}, {10000, time.Hour},
	} {
		delays := make(map[time.Duration]bool)
		for range 100 {
			delay := policy.Delay(test.attempt)
			assert.GreaterOrEqual(t, delay, test.base/2)
			assert.LessOrEqual(t, delay, test.base)
			delays[delay] = true
		}
		assert.Greater(t, len(delays), 1, "retries at the same attempt must spread their persisted deadlines")
	}
}
