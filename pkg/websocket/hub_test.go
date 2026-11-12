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

package websocket

import (
	"testing"

	"code.vikunja.io/api/pkg/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHubRegisterUnregister(t *testing.T) {
	h := NewHub()
	conn := &Connection{
		userID:        1,
		subscriptions: make(map[string]bool),
		send:          make(chan OutgoingMessage, 16),
	}
	h.Register(conn)
	assert.Len(t, h.connections[1], 1)

	h.Unregister(conn)
	assert.Empty(t, h.connections[1])
	_, exists := h.connections[1]
	assert.False(t, exists, "map entry should be deleted when last connection is removed")
}

func TestHubConnectionMetrics(t *testing.T) {
	InitHub()
	subscribedConnection(GetHub(), 1)
	subscribedConnection(GetHub(), 2)
	families, err := metrics.GetRegistry().Gather()
	require.NoError(t, err)
	var count float64
	for _, family := range families {
		if family.GetName() == "vikunja_websocket_connections" {
			count = family.GetMetric()[0].GetGauge().GetValue()
		}
	}
	assert.InDelta(t, 2, count, 0.001)
	GetHub().Shutdown()
}

func TestHubPublishToSubscribedConnection(t *testing.T) {
	h := NewHub()
	conn := &Connection{
		userID:        1,
		subscriptions: make(map[string]bool),
		send:          make(chan OutgoingMessage, 16),
	}
	h.Register(conn)
	conn.subscriptions["notification.created"] = true

	h.PublishForUser(1, "notification.created", map[string]string{"id": "1"})

	msg := <-conn.send
	assert.Equal(t, "notification.created", msg.Event)
}

func TestHubPublishSkipsUnsubscribedConnection(t *testing.T) {
	h := NewHub()
	conn := &Connection{
		userID:        1,
		subscriptions: make(map[string]bool),
		send:          make(chan OutgoingMessage, 16),
	}
	h.Register(conn)
	// Not subscribed to "notification.created"

	h.PublishForUser(1, "notification.created", map[string]string{"id": "1"})

	assert.Empty(t, conn.send)
}

func TestHubPublishSkipsOtherUsers(t *testing.T) {
	h := NewHub()
	conn := &Connection{
		userID:        2,
		subscriptions: make(map[string]bool),
		send:          make(chan OutgoingMessage, 16),
	}
	h.Register(conn)
	conn.subscriptions["notification.created"] = true

	h.PublishForUser(1, "notification.created", map[string]string{"id": "1"})

	assert.Empty(t, conn.send)
}
