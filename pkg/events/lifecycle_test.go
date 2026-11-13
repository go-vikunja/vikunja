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

package events

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type lifecycleEvent string

func (e lifecycleEvent) Name() string { return string(e) }

type lifecycleListener struct {
	handle func(*message.Message) error
}

func (l lifecycleListener) Name() string                      { return "lifecycle" }
func (l lifecycleListener) Handle(msg *message.Message) error { return l.handle(msg) }

func startLifecycleTest(t *testing.T, concurrency int, handlers map[string][]Listener) {
	t.Helper()
	config.InitDefaultConfig()
	config.EventsMaxConcurrentHandlers.Set(concurrency)
	log.InitLogger()
	previous := listeners
	listeners = handlers
	Unfake()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(func() {
		cancel()
		shutdownCtx, stop := context.WithTimeout(context.Background(), time.Second)
		defer stop()
		_ = Shutdown(shutdownCtx)
		listeners = previous
		Fake()
		config.EventsMaxConcurrentHandlers.Set(16)
	})
	ready, err := InitEventsForTesting(ctx)
	require.NoError(t, err)
	select {
	case <-ready:
	case <-time.After(time.Second):
		t.Fatal("event router did not become ready")
	}
	assert.True(t, Available())
}

func TestEventLifecycleDrainsQueuedAndChildEvents(t *testing.T) {
	entered := make(chan struct{})
	release := make(chan struct{})
	var children atomic.Int32
	startLifecycleTest(t, 16, map[string][]Listener{
		"parent": {lifecycleListener{handle: func(_ *message.Message) error {
			close(entered)
			<-release
			return Dispatch(lifecycleEvent("child"))
		}}},
		"child": {lifecycleListener{handle: func(_ *message.Message) error {
			children.Add(1)
			return nil
		}}},
	})
	require.NoError(t, Dispatch(lifecycleEvent("parent")))
	<-entered
	stopped := make(chan error, 1)
	go func() { stopped <- Shutdown(context.Background()) }()
	select {
	case <-stopped:
		t.Fatal("shutdown returned while handler was running")
	case <-time.After(20 * time.Millisecond):
	}
	close(release)
	select {
	case err := <-stopped:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("shutdown did not drain child event")
	}
	assert.EqualValues(t, 1, children.Load())
	assert.False(t, Available())
	assert.Error(t, Dispatch(lifecycleEvent("parent")))
}

func TestEventLifecycleLimitsHandlerConcurrency(t *testing.T) {
	entered := make(chan struct{}, 20)
	release := make(chan struct{})
	startLifecycleTest(t, 2, map[string][]Listener{
		"limited": {lifecycleListener{handle: func(_ *message.Message) error {
			entered <- struct{}{}
			<-release
			return nil
		}}},
	})
	for i := 0; i < 20; i++ {
		require.NoError(t, Dispatch(lifecycleEvent("limited")))
	}
	<-entered
	<-entered
	select {
	case <-entered:
		t.Error("more handlers running than configured")
	case <-time.After(20 * time.Millisecond):
	}
	close(release)
	require.NoError(t, Shutdown(context.Background()))
}
