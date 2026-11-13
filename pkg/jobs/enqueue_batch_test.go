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
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm/contexts"
)

type batchInsertCounter struct {
	enabled atomic.Bool
	inserts atomic.Int64
}

func (h *batchInsertCounter) BeforeProcess(c *contexts.ContextHook) (context.Context, error) {
	if h.enabled.Load() && strings.HasPrefix(c.SQL, "INSERT INTO jobs") {
		h.inserts.Add(1)
	}
	return c.Ctx, nil
}
func (*batchInsertCounter) AfterProcess(c *contexts.ContextHook) error { return c.Err }

func TestEnqueueBatchBoundsStatementsAndPreservesKeys(t *testing.T) {
	e := queueTestEngine(t)
	counter := &batchInsertCounter{}
	e.AddHook(counter)
	counter.enabled.Store(true)
	t.Cleanup(func() { counter.enabled.Store(false) })
	key := config.Key("jobs.queues." + t.Name() + ".maxattempts")
	old := key.GetInt()
	key.Set(9)
	t.Cleanup(func() { key.Set(old) })
	now := time.Now().UTC().Truncate(time.Second)
	requests := make([]Request, 1000)
	for i := range requests {
		requests[i] = Request{Queue: t.Name(), Kind: "batch", Payload: []byte(`{}`), UniqueKey: fmt.Sprint(i), RunAt: now}
	}
	s := e.NewSession()
	defer s.Close()
	defer DiscardPending(s)
	require.NoError(t, s.Begin())
	require.NoError(t, EnqueueBatch(s, requests))
	require.LessOrEqual(t, counter.inserts.Load(), int64(10), "enqueue must use bounded multi-row inserts")
	require.NoError(t, Enqueue(s, requests[0]))
	require.NoError(t, EnqueueBatch(s, requests[:2]))
	require.NoError(t, s.Commit())
	WakePending(s)
	rows := []Job{}
	require.NoError(t, e.Where("queue = ?", t.Name()).Find(&rows))
	require.Len(t, rows, 1000)
	for _, row := range rows {
		assert.Equal(t, 9, row.MaxAttempts)
		assert.True(t, now.Equal(row.RunAt), "run_at preserves the requested instant")
		assert.Equal(t, StateAvailable, row.State)
	}
}

func TestEnqueueBatchRollbackAndWake(t *testing.T) {
	e := queueTestEngine(t)
	wake, unwatch := watchQueue(t.Name())
	defer unwatch()
	requests := []Request{{Queue: t.Name(), Kind: "batch", Payload: []byte(`{}`), UniqueKey: "key"}, {Queue: t.Name(), Kind: "batch", Payload: []byte(`{}`)}}
	for _, commit := range []bool{false, true} {
		s := e.NewSession()
		require.NoError(t, s.Begin())
		require.NoError(t, EnqueueBatch(s, requests))
		select {
		case <-wake:
			t.Fatal("uncommitted batch woke consumers")
		default:
		}
		if commit {
			require.NoError(t, s.Commit())
			WakePending(s)
		} else {
			require.NoError(t, s.Rollback())
			DiscardPending(s)
			WakePending(s)
		}
		require.NoError(t, s.Close())
		count, err := e.Where("queue = ?", t.Name()).Count(new(Job))
		require.NoError(t, err)
		if commit {
			require.EqualValues(t, 2, count)
			select {
			case <-wake:
			default:
				t.Fatal("committed batch did not wake consumers")
			}
		} else {
			require.Zero(t, count)
			select {
			case <-wake:
				t.Fatal("rollback woke consumers")
			default:
			}
		}
	}
}

func TestEnqueueBatchValidatesBeforeInserting(t *testing.T) {
	e := queueTestEngine(t)
	s := e.NewSession()
	defer s.Close()
	require.Error(t, EnqueueBatch(s, []Request{{Queue: t.Name(), Kind: "batch", Payload: []byte(`{}`)}, {Queue: t.Name(), Kind: "batch", Payload: []byte(`invalid`)}}))
	count, err := e.Where("queue = ?", t.Name()).Count(new(Job))
	require.NoError(t, err)
	require.Zero(t, count)
}
