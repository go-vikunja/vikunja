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
