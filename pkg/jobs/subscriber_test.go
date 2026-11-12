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
