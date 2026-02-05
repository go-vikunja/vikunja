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

	"github.com/stretchr/testify/assert"
)

func TestConnectionSubscribeUnsubscribe(t *testing.T) {
	conn := &Connection{
		userID:        1,
		authenticated: true,
		subscriptions: make(map[string]bool),
		send:          make(chan OutgoingMessage, 16),
	}

	conn.Subscribe("notifications")
	assert.True(t, conn.IsSubscribed("notifications"))

	conn.Unsubscribe("notifications")
	assert.False(t, conn.IsSubscribed("notifications"))
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

func TestConnectionRejectsActionsBeforeAuth(t *testing.T) {
	conn := &Connection{
		userID:        0, // not authenticated
		authenticated: false,
		subscriptions: make(map[string]bool),
		send:          make(chan OutgoingMessage, 16),
	}

	// Try to subscribe before auth - should be rejected
	conn.handleMessage(IncomingMessage{Action: ActionSubscribe, Topic: "notifications"})

	// Should have sent an error
	msg := <-conn.send
	assert.Equal(t, "auth_required", msg.Error)
	assert.False(t, conn.IsSubscribed("notifications"))
}
