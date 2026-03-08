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
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConnectionSubscribeUnsubscribe(t *testing.T) {
	conn := &Connection{
		userID:        1,
		authenticated: true,
		subscriptions: make(map[string]bool),
		send:          make(chan OutgoingMessage, 16),
	}

	conn.Subscribe("notification.created")
	assert.True(t, conn.IsSubscribed("notification.created"))

	conn.Unsubscribe("notification.created")
	assert.False(t, conn.IsSubscribed("notification.created"))
}

func TestConnectionIsSubscribedReturnsFalseForUnknownTopic(t *testing.T) {
	conn := &Connection{
		userID:        1,
		authenticated: true,
		subscriptions: make(map[string]bool),
		send:          make(chan OutgoingMessage, 16),
	}

	assert.False(t, conn.IsSubscribed("something"))
}

func TestConnectionAcceptsEventNameTopic(t *testing.T) {
	hub := NewHub()
	conn := &Connection{
		hub:           hub,
		userID:        1,
		authenticated: true,
		subscriptions: make(map[string]bool),
		send:          make(chan OutgoingMessage, 16),
	}
	hub.Register(conn)

	conn.handleMessage(context.Background(), IncomingMessage{Action: ActionSubscribe, Topic: "notification.created"})

	assert.True(t, conn.IsSubscribed("notification.created"))
}

func TestConnectionRejectsOldTopicName(t *testing.T) {
	conn := &Connection{
		userID:        1,
		authenticated: true,
		subscriptions: make(map[string]bool),
		send:          make(chan OutgoingMessage, 16),
	}

	conn.handleMessage(context.Background(), IncomingMessage{Action: ActionSubscribe, Topic: "notifications"})

	msg := <-conn.send
	assert.Equal(t, "invalid_topic", msg.Error)
	assert.False(t, conn.IsSubscribed("notifications"))
}

func TestConnectionRejectsActionsBeforeAuth(t *testing.T) {
	conn := &Connection{
		userID:        0, // not authenticated
		authenticated: false,
		subscriptions: make(map[string]bool),
		send:          make(chan OutgoingMessage, 16),
	}

	// Try to subscribe before auth - should be rejected
	conn.handleMessage(context.Background(), IncomingMessage{Action: ActionSubscribe, Topic: "notification.created"})

	// Should have sent an error
	msg := <-conn.send
	assert.Equal(t, "auth_required", msg.Error)
	assert.False(t, conn.IsSubscribed("notification.created"))
}
