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
	"testing"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/stretchr/testify/require"
)

func TestSubscriberDeadlineStopsRenewal(t *testing.T) {
	e := queueTestEngine(t)
	s := e.NewSession()
	defer s.Close()
	require.NoError(t, Enqueue(s, Request{Queue: t.Name(), Kind: "test", Payload: []byte(`{}`)}))
	sub := NewSubscriber(NewQueue(e, "deadline"), map[string]Options{t.Name(): {Concurrency: 1, Lease: time.Minute, Timeout: 30 * time.Millisecond}})
	defer sub.Close()
	out, err := sub.Subscribe(t.Context(), t.Name())
	require.NoError(t, err)
	select {
	case msg := <-out:
		select {
		case <-msg.Context().Done():
			require.ErrorIs(t, msg.Context().Err(), context.DeadlineExceeded)
		case <-time.After(time.Second):
			t.Fatal("handler was not cancelled")
		}
		msg.Nack()
	case <-time.After(time.Second):
		t.Fatal("job was not delivered")
	}
}

func TestSubscriberRouterContextMutation(t *testing.T) {
	e := queueTestEngine(t)
	previousMax := e.DB().Stats().MaxOpenConnections
	e.SetMaxOpenConns(1)
	t.Cleanup(func() { e.SetMaxOpenConns(previousMax) })
	session := e.NewSession()
	defer session.Close()
	for range 100 {
		require.NoError(t, Enqueue(session, Request{Queue: t.Name(), Kind: "test", Payload: []byte(`{}`)}))
	}
	sub := NewSubscriber(NewQueue(e, "router"), map[string]Options{t.Name(): {Concurrency: 2, Lease: time.Minute, Timeout: time.Minute}})
	router, err := message.NewRouter(message.RouterConfig{}, watermill.NopLogger{})
	require.NoError(t, err)
	router.AddConsumerHandler(t.Name(), t.Name(), sub, func(*message.Message) error { return nil })
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	finished := make(chan error, 1)
	go func() { finished <- router.Run(ctx) }()
	<-router.Running()
	t.Cleanup(func() {
		require.NoError(t, router.Close())
		require.NoError(t, <-finished)
	})
	require.Eventually(t, func() bool {
		count, err := e.Where("queue = ? AND state = ?", t.Name(), StateCompleted).Count(new(Job))
		require.NoError(t, err)
		return count == 100
	}, 10*time.Second, 10*time.Millisecond)
}
