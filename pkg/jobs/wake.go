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
	"runtime"
	"sync"
	"weak"

	"xorm.io/xorm"
	"xorm.io/xorm/core"
)

var queueWakes = struct {
	sync.Mutex
	listeners map[string]map[chan struct{}]struct{}
}{listeners: make(map[string]map[chan struct{}]struct{})}

func watchQueue(queue string) (<-chan struct{}, func()) {
	ch := make(chan struct{}, 1)
	queueWakes.Lock()
	if queueWakes.listeners[queue] == nil {
		queueWakes.listeners[queue] = make(map[chan struct{}]struct{})
	}
	queueWakes.listeners[queue][ch] = struct{}{}
	queueWakes.Unlock()
	return ch, func() {
		queueWakes.Lock()
		defer queueWakes.Unlock()
		delete(queueWakes.listeners[queue], ch)
		if len(queueWakes.listeners[queue]) == 0 {
			delete(queueWakes.listeners, queue)
		}
	}
}

// Wake coalesces local hints; polling still discovers other replicas' commits.
func Wake(queues ...string) {
	queueWakes.Lock()
	defer queueWakes.Unlock()
	notify := func(listeners map[chan struct{}]struct{}) {
		for ch := range listeners {
			select {
			case ch <- struct{}{}:
			default:
			}
		}
	}
	if len(queues) == 0 {
		for _, listeners := range queueWakes.listeners {
			notify(listeners)
		}
		return
	}
	for _, queue := range queues {
		notify(queueWakes.listeners[queue])
	}
}

type pendingWake struct {
	sync.Mutex
	queues map[string]struct{}
	counts map[metricLabel]int64
	tx     *core.Tx
}

var pendingWakes sync.Map

func pendingFor(s *xorm.Session) *pendingWake {
	key := weak.Make(s)
	item, loaded := pendingWakes.LoadOrStore(key, &pendingWake{queues: make(map[string]struct{}), counts: make(map[metricLabel]int64), tx: s.Tx()})
	if !loaded {
		runtime.AddCleanup(s, func(key weak.Pointer[xorm.Session]) { pendingWakes.Delete(key) }, key)
	}
	pending := item.(*pendingWake)
	pending.Lock()
	if pending.tx != s.Tx() {
		pending.tx = s.Tx()
		clear(pending.queues)
		clear(pending.counts)
	}
	pending.Unlock()
	return pending
}

func wakeAfterCommit(s *xorm.Session, queue string) {
	if !s.IsInTx() {
		Wake(queue)
		return
	}
	pending := pendingFor(s)
	pending.Lock()
	pending.queues[queue] = struct{}{}
	pending.Unlock()
}

// WakePending follows a successful commit of keyed inserts, which bypass XORM hooks.
func WakePending(s *xorm.Session) {
	if s == nil {
		return
	}
	value, ok := pendingWakes.LoadAndDelete(weak.Make(s))
	if !ok {
		return
	}
	pending := value.(*pendingWake)
	pending.Lock()
	defer pending.Unlock()
	if pending.tx != s.Tx() {
		return
	}
	for label, count := range pending.counts {
		jobCounters[label.event].WithLabelValues(label.queue, label.kind).Add(float64(count))
	}
	for queue := range pending.queues {
		Wake(queue)
	}
}

func DiscardPending(s *xorm.Session) {
	if s != nil {
		pendingWakes.Delete(weak.Make(s))
	}
}
