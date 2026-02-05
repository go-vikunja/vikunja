package websocket

import (
	"sync"

	"code.vikunja.io/api/pkg/log"
)

// Hub maintains the set of active connections and delivers messages to them.
type Hub struct {
	mu          sync.RWMutex
	connections map[int64][]*Connection // userID -> connections
}

// NewHub creates a new Hub.
func NewHub() *Hub {
	return &Hub{
		connections: make(map[int64][]*Connection),
	}
}

// Register adds a connection to the hub.
func (h *Hub) Register(conn *Connection) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.connections[conn.userID] = append(h.connections[conn.userID], conn)
	log.Debugf("WebSocket: registered connection for user %d (total: %d)", conn.userID, len(h.connections[conn.userID]))
}

// Unregister removes a connection from the hub.
func (h *Hub) Unregister(conn *Connection) {
	h.mu.Lock()
	defer h.mu.Unlock()
	conns := h.connections[conn.userID]
	for i, c := range conns {
		if c == conn {
			h.connections[conn.userID] = append(conns[:i], conns[i+1:]...)
			break
		}
	}
	log.Debugf("WebSocket: unregistered connection for user %d (remaining: %d)", conn.userID, len(h.connections[conn.userID]))
}

// PublishForUser sends an event to all connections of a specific user that are subscribed to the given topic.
func (h *Hub) PublishForUser(userID int64, topic, event string, data any) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	conns := h.connections[userID]
	msg := OutgoingMessage{
		Event: event,
		Topic: topic,
		Data:  data,
	}

	for _, conn := range conns {
		if !conn.IsSubscribed(topic) {
			continue
		}
		select {
		case conn.send <- msg:
		default:
			log.Warningf("WebSocket: send buffer full for user %d, dropping message", userID)
		}
	}
}
