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
)

// Hub maintains the set of active connections and delivers messages to them.
type Hub struct {
	mu          sync.RWMutex
	connections map[int64][]*Connection // userID -> connections
	all         map[*Connection]struct{}
	closed      bool
	unavailable bool
	fanout      *redisFanout
	resyncs     uint64
}

// NewHub creates a new Hub.
func NewHub() *Hub {
	return &Hub{
		connections: make(map[int64][]*Connection),
		all:         make(map[*Connection]struct{}),
	}
}

// Register adds a connection to the hub.
func (h *Hub) Register(conn *Connection) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed || h.unavailable {
		conn.closeNow()
		return
	}
	h.all[conn] = struct{}{}
	h.connections[conn.userID] = append(h.connections[conn.userID], conn)
	log.Debugf("WebSocket: registered connection for user %d (total: %d)", conn.userID, len(h.connections[conn.userID]))
}

// Unregister removes a connection from the hub.
func (h *Hub) Unregister(conn *Connection) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.all, conn)
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
	h.publishLocal(userID, event, data)
	h.mu.RLock()
	fanout := h.fanout
	closed := h.closed
	h.mu.RUnlock()
	if fanout != nil && !closed {
		if err := fanout.publish(userID, event, data); err != nil {
			log.Errorf("WebSocket: Redis publish failed: %v", err)
			h.setUnavailable()
		}
	}
}

func (h *Hub) publishLocal(userID int64, event string, data any) {
	h.mu.RLock()

	conns := h.connections[userID]
	msg := OutgoingMessage{
		Event: event,
		Data:  data,
	}

	var slow []*Connection
	for _, conn := range conns {
		if !conn.IsSubscribed(event) {
			continue
		}
		select {
		case conn.send <- msg:
		default:
			slow = append(slow, conn)
		}
	}
	h.mu.RUnlock()
	if len(slow) > 0 {
		h.mu.Lock()
		h.resyncs += uint64(len(slow))
		h.mu.Unlock()
	}
	for _, conn := range slow {
		conn.closeNow()
	}
}

func (h *Hub) Available() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return !h.closed && !h.unavailable
}

func (h *Hub) track(conn *Connection) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed || h.unavailable {
		return false
	}
	h.all[conn] = struct{}{}
	return true
}

func (h *Hub) setUnavailable() {
	h.mu.Lock()
	h.unavailable = true
	connections := h.all
	h.resyncs += uint64(len(connections))
	h.all = make(map[*Connection]struct{})
	h.connections = make(map[int64][]*Connection)
	h.mu.Unlock()
	for conn := range connections {
		conn.closeNow()
	}
}

func (h *Hub) Shutdown() {
	h.mu.Lock()
	h.closed = true
	fanout := h.fanout
	h.mu.Unlock()
	h.setUnavailable()
	if fanout != nil {
		fanout.stop()
	}
}
