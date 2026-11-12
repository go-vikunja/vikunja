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
	"errors"
	"sync"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
)

type eventRuntime struct {
	mu      sync.Mutex
	pending map[string]int
	drained chan struct{}
	closed  bool
	router  *message.Router
	channel *gochannel.GoChannel
	cancel  context.CancelFunc
	done    chan struct{}
	runErr  error
}

var runtimeState struct {
	sync.RWMutex
	current *eventRuntime
}

func currentRuntime() *eventRuntime {
	runtimeState.RLock()
	defer runtimeState.RUnlock()
	return runtimeState.current
}

func Available() bool {
	r := currentRuntime()
	if r == nil {
		return false
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return !r.closed && r.router.IsRunning()
}

func setRuntime(r *eventRuntime) {
	runtimeState.Lock()
	defer runtimeState.Unlock()
	runtimeState.current = r
}

func newEventRuntime(router *message.Router, channel *gochannel.GoChannel, cancel context.CancelFunc) *eventRuntime {
	drained := make(chan struct{})
	close(drained)
	return &eventRuntime{
		pending: make(map[string]int),
		drained: drained,
		router:  router,
		channel: channel,
		cancel:  cancel,
		done:    make(chan struct{}),
	}
}

func (r *eventRuntime) publish(topic string, msg *message.Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed || !r.router.IsRunning() {
		return errors.New("event system is not running")
	}
	count := len(listeners[topic])
	if count > 0 {
		if len(r.pending) == 0 {
			r.drained = make(chan struct{})
		}
		r.pending[msg.UUID] = count
	}
	if err := r.channel.Publish(topic, msg); err != nil {
		delete(r.pending, msg.UUID)
		if count > 0 && len(r.pending) == 0 {
			close(r.drained)
		}
		return err
	}
	return nil
}

func (r *eventRuntime) finish(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	count, exists := r.pending[id]
	if !exists {
		return
	}
	if count > 1 {
		r.pending[id] = count - 1
		return
	}
	delete(r.pending, id)
	if len(r.pending) == 0 {
		close(r.drained)
	}
}

func (r *eventRuntime) wait(ctx context.Context, closeWhenDrained bool) error {
	for {
		r.mu.Lock()
		if len(r.pending) == 0 {
			if closeWhenDrained {
				r.closed = true
			}
			r.mu.Unlock()
			return nil
		}
		drained := r.drained
		r.mu.Unlock()
		select {
		case <-drained:
		case <-ctx.Done():
			return ctx.Err()
		case <-r.done:
			return errors.New("event router stopped before draining pending events")
		}
	}
}

func (r *eventRuntime) waitForPendingHandlers() {
	for {
		r.mu.Lock()
		pending := len(r.pending)
		drained := r.drained
		r.mu.Unlock()
		if pending == 0 {
			return
		}
		select {
		case <-drained:
		case <-r.done:
			return
		}
	}
}

// Shutdown drains accepted work, including events dispatched by running handlers.
func Shutdown(ctx context.Context) error {
	r := currentRuntime()
	if r == nil {
		return nil
	}
	if err := r.wait(ctx, true); err != nil {
		r.mu.Lock()
		r.closed = true
		r.mu.Unlock()
		r.cancel()
		return err
	}
	r.cancel()
	select {
	case <-r.done:
		return r.runErr
	case <-ctx.Done():
		return ctx.Err()
	}
}
