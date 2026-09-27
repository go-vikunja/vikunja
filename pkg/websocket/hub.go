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
	"sync"

	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"xorm.io/xorm"
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
	remaining := len(h.connections[conn.userID])
	if remaining == 0 {
		delete(h.connections, conn.userID)
	}
	log.Debugf("WebSocket: unregistered connection for user %d (remaining: %d)", conn.userID, remaining)
}

// PublishForUser sends an event to all connections of a specific user that are subscribed to the given event.
func (h *Hub) PublishForUser(userID int64, event string, data any) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	conns := h.connections[userID]
	msg := OutgoingMessage{
		Event: event,
		Data:  data,
	}

	for _, conn := range conns {
		if !conn.IsSubscribed(event) {
			continue
		}
		select {
		case conn.send <- msg:
		default:
			log.Warningf("WebSocket: send buffer full for user %d, dropping message", userID)
		}
	}
}

// PublishTaskEvent sends a task event to every subscribed connection whose
// user can read the task's project. Delivery is scoped per connection because
// subscribers of a task event are not known in advance: any authenticated user
// may subscribe, and project access differs per user.
func (h *Hub) PublishTaskEvent(s *xorm.Session, event string, task *models.Task) {
	h.mu.RLock()
	conns := h.allConnections()
	h.mu.RUnlock()

	msg := OutgoingMessage{Event: event, Data: task}
	for _, conn := range conns {
		if !conn.IsSubscribed(event) {
			continue
		}
		if !h.canReadTask(s, conn.UserID(), task) {
			continue
		}
		conn.trySend(msg)
	}
}

// PublishCommentEvent sends a task comment event to every subscribed
// connection whose user can read the comment's task. See PublishTaskEvent.
func (h *Hub) PublishCommentEvent(s *xorm.Session, event string, task *models.Task, payload *CommentEventPayload) {
	h.mu.RLock()
	conns := h.allConnections()
	h.mu.RUnlock()

	msg := OutgoingMessage{Event: event, Data: payload}
	for _, conn := range conns {
		if !conn.IsSubscribed(event) {
			continue
		}
		if !h.canReadTask(s, conn.UserID(), task) {
			continue
		}
		conn.trySend(msg)
	}
}

// allConnections returns every registered connection. Callers must hold h.mu.
func (h *Hub) allConnections() []*Connection {
	conns := make([]*Connection, 0, len(h.connections))
	for _, userConns := range h.connections {
		conns = append(conns, userConns...)
	}
	return conns
}

func (h *Hub) canReadTask(s *xorm.Session, userID int64, task *models.Task) bool {
	// A fresh minimal task keeps the check honest: resolving from the event
	// payload would trust the task's own ProjectID, which every producer
	// sets correctly, but a copy costs little and guards against listeners
	// wired to events whose payload is user-controlled.
	canRead, _, err := (&models.Task{ID: task.ID}).CanRead(s, &user.User{ID: userID})
	if err != nil {
		log.Errorf("WebSocket: access check for task %d failed: %v", task.ID, err)
		return false
	}
	return canRead
}

// trySend enqueues msg without ever blocking: a slow consumer must not hold
// up fan-out to other connections.
func (c *Connection) trySend(msg OutgoingMessage) {
	select {
	case c.send <- msg:
	default:
		log.Warningf("WebSocket: send buffer full for user %d, dropping message", c.UserID())
	}
}
