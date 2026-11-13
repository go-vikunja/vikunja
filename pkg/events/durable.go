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
	"encoding/json"
	"fmt"

	"code.vikunja.io/api/pkg/jobs"
	"xorm.io/xorm"
)

type durableListener struct {
	queue string
	kind  string
}

var durableListeners = map[string][]durableListener{}

func RegisterDurableListener(topic, queue string, listener Listener) {
	kind := topic + "." + listener.Name()
	for _, registered := range durableListeners[topic] {
		if registered.kind == kind {
			return
		}
	}
	durableListeners[topic] = append(durableListeners[topic], durableListener{queue: queue, kind: kind})
	jobs.Register(kind, queue, listener.Handle)
}

func enqueueDurable(s *xorm.Session, event Event) error {
	targets := durableListeners[event.Name()]
	if len(targets) == 0 || isUnderTest {
		return nil
	}
	if s == nil {
		return fmt.Errorf("durable event %s requires a transaction", event.Name())
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	for _, target := range targets {
		if err := jobs.Enqueue(s, jobs.Request{Queue: target.queue, Kind: target.kind, Payload: payload}); err != nil {
			return err
		}
	}
	return nil
}
