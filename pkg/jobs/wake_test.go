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
	"fmt"
	"testing"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

func TestSubscriberWakesTheQueueWithNewWork(t *testing.T) {
	e := queueTestEngine(t)
	previousMax := e.DB().Stats().MaxOpenConnections
	e.SetMaxOpenConns(1)
	t.Cleanup(func() {
		e.SetMaxOpenConns(previousMax)
		_, err := e.Where("queue LIKE ?", t.Name()+"-%").Delete(new(Job))
		require.NoError(t, err)
	})
	opts := make(map[string]Options)
	for i := range 8 {
		opts[fmt.Sprintf("%s-%d", t.Name(), i)] = Options{Concurrency: 1, Lease: time.Minute, Timeout: time.Minute}
	}
	sub := NewSubscriber(NewQueue(e, "wake"), opts)
	defer sub.Close()
	outputs := make(map[string]<-chan *message.Message)
	for queue := range opts {
		out, err := sub.Subscribe(t.Context(), queue)
		require.NoError(t, err)
		outputs[queue] = out
	}
	time.Sleep(7 * time.Second)
	for queue, out := range outputs {
		s := e.NewSession()
		require.NoError(t, s.Begin())
		require.NoError(t, Enqueue(s, Request{Queue: queue, Kind: "test", Payload: []byte(`{}`)}))
		require.NoError(t, s.Commit())
		require.NoError(t, s.Close())
		select {
		case msg := <-out:
			msg.Ack()
		case <-time.After(300 * time.Millisecond):
			t.Fatalf("new work did not wake queue %s", queue)
		}
	}
}

func TestQueueWithNoFreeSlotDoesNotQuery(t *testing.T) {
	e, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	require.NoError(t, e.Close())
	jobs, err := NewQueue(e, "full").Claim(t.Context(), t.Name(), 0, time.Now(), time.Minute)
	require.NoError(t, err)
	require.Empty(t, jobs)
}

func TestQueueWakeTargetsEveryLocalSubscriber(t *testing.T) {
	first, stopFirst := watchQueue(t.Name())
	defer stopFirst()
	second, stopSecond := watchQueue(t.Name())
	defer stopSecond()
	other, stopOther := watchQueue(t.Name() + "-other")
	defer stopOther()
	Wake(t.Name())
	for _, listener := range []<-chan struct{}{first, second} {
		select {
		case <-listener:
		default:
			t.Fatal("matching queue was not woken")
		}
	}
	select {
	case <-other:
		t.Fatal("unrelated queue was woken")
	default:
	}
}

func TestKeyedEnqueueWakesOnlyAfterCommit(t *testing.T) {
	e := queueTestEngine(t)
	wakeups, unwatch := watchQueue(t.Name())
	defer unwatch()
	request := Request{Queue: t.Name(), Kind: "test", Payload: []byte(`{}`), UniqueKey: "key"}
	s := e.NewSession()
	defer s.Close()
	defer DiscardPending(s)
	require.NoError(t, s.Begin())
	require.NoError(t, Enqueue(s, request))
	select {
	case <-wakeups:
		t.Fatal("uncommitted insert woke subscriber")
	default:
	}
	require.NoError(t, s.Rollback())
	DiscardPending(s)
	WakePending(s)
	select {
	case <-wakeups:
		t.Fatal("rolled-back insert woke subscriber")
	default:
	}
	require.NoError(t, s.Close())
	s = e.NewSession()
	defer s.Close()
	defer DiscardPending(s)
	require.NoError(t, s.Begin())
	require.NoError(t, Enqueue(s, request))
	require.NoError(t, s.Commit())
	WakePending(s)
	select {
	case <-wakeups:
	default:
		t.Fatal("committed insert did not wake subscriber")
	}
	WakePending(s)
	select {
	case <-wakeups:
		t.Fatal("commit wake was repeated")
	default:
	}
	_, err := e.Where("queue = ?", t.Name()).Delete(new(Job))
	require.NoError(t, err)
}
