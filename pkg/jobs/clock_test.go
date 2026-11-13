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
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

func postgresClockEngine(t *testing.T) *xorm.Engine {
	t.Helper()
	e := queueTestEngine(t)
	if e.Dialect().URI().DBType != schemas.POSTGRES {
		t.Skip("PostgreSQL database clock")
	}
	previous := e.DB().Stats().MaxOpenConnections
	e.SetMaxOpenConns(1)
	t.Cleanup(func() { _, err := e.Exec("RESET TIME ZONE"); require.NoError(t, err); e.SetMaxOpenConns(previous) })
	_, err := e.Exec("SET TIME ZONE 'Pacific/Auckland'")
	require.NoError(t, err)
	return e
}

func databaseNow(t *testing.T, e *xorm.Engine) time.Time {
	t.Helper()
	var row struct{ Now time.Time }
	_, err := e.SQL("SELECT clock_timestamp() AT TIME ZONE 'UTC' AS now").Get(&row)
	require.NoError(t, err)
	return row.Now
}

func TestQueuePostgresClockIgnoresCallerSkew(t *testing.T) {
	for _, skew := range []time.Duration{-3 * time.Minute, 3 * time.Minute} {
		t.Run(skew.String(), func(t *testing.T) {
			e := postgresClockEngine(t)
			now := databaseNow(t, e)
			caller := now.Add(skew)
			q := NewQueue(e, "database-clock")
			s := e.NewSession()
			defer s.Close()
			require.NoError(t, Enqueue(s, Request{Queue: t.Name(), Kind: "future", Payload: []byte(`{}`), RunAt: now.Add(time.Minute)}))
			future, err := q.Claim(t.Context(), t.Name(), 1, caller, time.Minute)
			require.NoError(t, err)
			require.Empty(t, future, "fast caller cannot claim future work")
			require.NoError(t, Enqueue(s, Request{Queue: t.Name(), Kind: "due", Payload: []byte(`{}`), RunAt: now.Add(-time.Minute)}))
			batch, err := q.Claim(t.Context(), t.Name(), 1, caller, time.Minute)
			require.NoError(t, err)
			require.Len(t, batch, 1, "slow caller cannot postpone due work")
			assert.WithinDuration(t, now, batch[0].ClaimedAt, 2*time.Second)
			assert.WithinDuration(t, now.Add(time.Minute), batch[0].ExpiresAt, 2*time.Second)
			count, err := q.Rescue(t.Context(), now.Add(3*time.Minute), 100)
			require.NoError(t, err)
			require.Zero(t, count, "fast caller cannot rescue a live claim")
			count, err = q.Renew(t.Context(), batch, caller, 2*time.Minute)
			require.NoError(t, err)
			require.EqualValues(t, 1, count)
			var renewed Job
			_, err = s.ID(batch[0].ID).Get(&renewed)
			require.NoError(t, err)
			assert.WithinDuration(t, databaseNow(t, e).Add(2*time.Minute), renewed.ExpiresAt, 2*time.Second)
			require.NoError(t, Complete(s, batch[0], now.Add(3*time.Minute)), "fast caller cannot reject a live completion")
		})
	}
}

func TestQueuePostgresClockRetriesRescueAndPark(t *testing.T) {
	e := postgresClockEngine(t)
	now := databaseNow(t, e)
	q := NewQueue(e, "database-clock")
	s := e.NewSession()
	defer s.Close()
	job := Job{Queue: t.Name(), Kind: "retry", Payload: "{}", State: StateRunning, Attempt: 1, MaxAttempts: 5, ClaimOwner: "database-clock", RunAt: now, ExpiresAt: now.Add(time.Minute)}
	_, err := s.Insert(&job)
	require.NoError(t, err)
	require.NoError(t, q.Fail(t.Context(), job, now.Add(3*time.Minute), errors.New("offline"), 5))
	var stored Job
	stored = Job{}
	_, err = s.ID(job.ID).Get(&stored)
	require.NoError(t, err)
	assert.WithinDuration(t, now.Add(time.Second), stored.RunAt, 2*time.Second)
	job.State = StateRunning
	job.Attempt = 5
	job.ExpiresAt = now.Add(-time.Second)
	_, err = s.ID(job.ID).Cols("state", "attempt", "expires_at").Update(&job)
	require.NoError(t, err)
	count, err := q.Rescue(t.Context(), now.Add(-3*time.Minute), 100)
	require.NoError(t, err)
	require.EqualValues(t, 1, count, "slow caller cannot conceal expiration")
	stored = Job{}
	_, err = s.ID(job.ID).Get(&stored)
	require.NoError(t, err)
	assert.Equal(t, StateFailed, stored.State)
	assert.WithinDuration(t, now, stored.FinishedAt, 2*time.Second)
	job.State = StateRunning
	_, err = s.ID(job.ID).Cols("state").Update(&job)
	require.NoError(t, err)
	require.NoError(t, q.Fail(t.Context(), job, now.Add(3*time.Minute), errors.New("offline"), 5))
	stored = Job{}
	_, err = s.ID(job.ID).Get(&stored)
	require.NoError(t, err)
	assert.WithinDuration(t, now, stored.FinishedAt, 2*time.Second)
}

func TestQueuePostgresClockEnqueueDefaults(t *testing.T) {
	e := postgresClockEngine(t)
	before := databaseNow(t, e).Truncate(time.Second)
	s := e.NewSession()
	defer s.Close()
	for _, key := range []string{"", "keyed"} {
		require.NoError(t, Enqueue(s, Request{Queue: t.Name(), Kind: "single", Payload: []byte(`{}`), UniqueKey: key}))
	}
	require.NoError(t, EnqueueBatch(s, []Request{{Queue: t.Name(), Kind: "batch", Payload: []byte(`{}`)}, {Queue: t.Name(), Kind: "batch", Payload: []byte(`{}`), UniqueKey: "keyed"}}))
	var rows []Job
	require.NoError(t, s.Where("queue = ?", t.Name()).Find(&rows))
	require.Len(t, rows, 4)
	after := databaseNow(t, e)
	for _, row := range rows {
		assert.False(t, row.RunAt.Before(before))
		assert.False(t, row.RunAt.After(after))
	}
}

func TestQueuePostgresClockKeepsIndexedClaimPlans(t *testing.T) {
	e := postgresClockEngine(t)
	_, err := e.Exec("CREATE TEMP TABLE jobs (LIKE public.jobs INCLUDING ALL)")
	require.NoError(t, err)
	t.Cleanup(func() { _, err := e.Exec("DROP TABLE pg_temp.jobs"); require.NoError(t, err) })
	_, err = e.Exec("CREATE INDEX clock_claim_order ON jobs (queue,state,run_at,id)")
	require.NoError(t, err)
	_, err = e.Exec(`INSERT INTO pg_temp.jobs (queue,kind,payload,state,run_at)
 SELECT 'clock-index','test','{}','available',(clock_timestamp() AT TIME ZONE 'UTC') + INTERVAL '1 hour' FROM generate_series(1,2000)`)
	require.NoError(t, err)
	_, err = e.Exec(`INSERT INTO pg_temp.jobs (queue,kind,payload,state,run_at)
 SELECT 'clock-index','test','{}','available',(clock_timestamp() AT TIME ZONE 'UTC') - INTERVAL '1 minute' FROM generate_series(1,2)`)
	require.NoError(t, err)
	_, err = e.Exec("ANALYZE pg_temp.jobs")
	require.NoError(t, err)
	query := postgresClaimSQL
	for i := 1; strings.Contains(query, "?"); i++ {
		query = strings.Replace(query, "?", fmt.Sprintf("$%d", i), 1)
	}
	_, err = e.Exec("PREPARE clock_claim(text,integer,text,double precision) AS " + query)
	require.NoError(t, err)
	t.Cleanup(func() { _, err := e.Exec("DEALLOCATE clock_claim"); require.NoError(t, err) })
	var version struct{ Version int }
	_, err = e.SQL("SELECT current_setting('server_version_num')::integer AS version").Get(&version)
	require.NoError(t, err)
	modes := []string{"auto"}
	if version.Version >= 120000 {
		modes = []string{"force_custom_plan", "force_generic_plan"}
		t.Cleanup(func() { _, err := e.Exec("RESET plan_cache_mode"); require.NoError(t, err) })
	}
	for _, mode := range modes {
		if version.Version >= 120000 {
			_, err = e.Exec("SET plan_cache_mode = " + mode)
			require.NoError(t, err)
		}
		plan, err := e.QueryString("EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON) EXECUTE clock_claim('clock-index',2,'worker',60)")
		require.NoError(t, err)
		text := plan[0]["QUERY PLAN"]
		assert.Contains(t, text, `"Index Name": "clock_claim_order"`, mode)
		assert.Contains(t, text, `run_at <=`, mode)
		assert.Contains(t, text, `InitPlan`, mode)
		t.Logf("%s claim plan: %s", mode, text)
	}
}

func TestQueuePostgresClockRejectsExpiredSlowCaller(t *testing.T) {
	e := postgresClockEngine(t)
	now := databaseNow(t, e)
	s := e.NewSession()
	defer s.Close()
	job := Job{Queue: t.Name(), Kind: "test", Payload: "{}", State: StateRunning, Attempt: 1, MaxAttempts: 5, ClaimOwner: "expired", RunAt: now, ExpiresAt: now.Add(-time.Second)}
	_, err := s.Insert(&job)
	require.NoError(t, err)
	count, err := NewQueue(e, "expired").Renew(t.Context(), []Job{job}, now.Add(-3*time.Minute), time.Minute)
	require.NoError(t, err)
	require.Zero(t, count)
	require.ErrorIs(t, Complete(s, job, now.Add(-3*time.Minute)), ErrStaleClaim)
}
