package websocket

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
	assert.Len(t, h.connections[1], 0)
}

func TestHubPublishToSubscribedConnection(t *testing.T) {
	h := NewHub()
	conn := &Connection{
		userID:        1,
		subscriptions: make(map[string]bool),
		send:          make(chan OutgoingMessage, 16),
	}
	h.Register(conn)
	conn.subscriptions["notifications"] = true

	h.PublishForUser(1, "notifications", "notification.created", map[string]string{"id": "1"})

	msg := <-conn.send
	assert.Equal(t, "notification.created", msg.Event)
	assert.Equal(t, "notifications", msg.Topic)
}

func TestHubPublishSkipsUnsubscribedConnection(t *testing.T) {
	h := NewHub()
	conn := &Connection{
		userID:        1,
		subscriptions: make(map[string]bool),
		send:          make(chan OutgoingMessage, 16),
	}
	h.Register(conn)
	// Not subscribed to "notifications"

	h.PublishForUser(1, "notifications", "notification.created", map[string]string{"id": "1"})

	assert.Len(t, conn.send, 0)
}

func TestHubPublishSkipsOtherUsers(t *testing.T) {
	h := NewHub()
	conn := &Connection{
		userID:        2,
		subscriptions: make(map[string]bool),
		send:          make(chan OutgoingMessage, 16),
	}
	h.Register(conn)
	conn.subscriptions["notifications"] = true

	h.PublishForUser(1, "notifications", "notification.created", map[string]string{"id": "1"})

	assert.Len(t, conn.send, 0)
}
