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
	"encoding/json"
	"sync"
	"time"

	"code.vikunja.io/api/pkg/log"

	"github.com/coder/websocket"
)

const (
	writeTimeout = 10 * time.Second
	pingInterval = 30 * time.Second
	sendBufSize  = 64
)

// Connection wraps a single WebSocket connection.
type Connection struct {
	ws  *websocket.Conn
	hub *Hub

	mu            sync.RWMutex
	userID        int64
	authenticated bool
	subscriptions map[string]bool

	send chan OutgoingMessage
}

// NewConnection creates a new unauthenticated Connection.
func NewConnection(ws *websocket.Conn, hub *Hub) *Connection {
	return &Connection{
		ws:            ws,
		hub:           hub,
		authenticated: false,
		subscriptions: make(map[string]bool),
		send:          make(chan OutgoingMessage, sendBufSize),
	}
}

// Subscribe adds a topic subscription.
func (c *Connection) Subscribe(topic string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.subscriptions[topic] = true
}

// Unsubscribe removes a topic subscription.
func (c *Connection) Unsubscribe(topic string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.subscriptions, topic)
}

// IsSubscribed checks if the connection is subscribed to a topic.
func (c *Connection) IsSubscribed(topic string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.subscriptions[topic]
}

// IsAuthenticated returns whether the connection is authenticated.
func (c *Connection) IsAuthenticated() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.authenticated
}

// UserID returns the authenticated user's ID.
func (c *Connection) UserID() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.userID
}

// ReadLoop reads messages from the WebSocket and handles auth/subscribe/unsubscribe.
func (c *Connection) ReadLoop(ctx context.Context, cancel context.CancelFunc) {
	defer func() {
		cancel()
		if c.IsAuthenticated() {
			c.hub.Unregister(c)
		}
		c.ws.Close(websocket.StatusNormalClosure, "")
	}()

	for {
		_, data, err := c.ws.Read(ctx)
		if err != nil {
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
				log.Debugf("WebSocket: connection closed normally for user %d", c.UserID())
			} else {
				log.Debugf("WebSocket: read error for user %d: %v", c.UserID(), err)
			}
			return
		}

		var msg IncomingMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Warningf("WebSocket: invalid message: %v", err)
			continue
		}

		if !c.handleMessage(ctx, msg) {
			return // close connection
		}
	}
}

// handleMessage processes an incoming message. Returns false if connection should be closed.
func (c *Connection) handleMessage(ctx context.Context, msg IncomingMessage) bool {
	switch msg.Action {
	case ActionAuth:
		return c.handleAuth(ctx, msg.Token)
	case ActionSubscribe:
		if !c.IsAuthenticated() {
			c.sendError("auth_required", "")
			return true
		}
		if !isValidTopic(msg.Topic) {
			c.sendError("invalid_topic", msg.Topic)
			return true
		}
		c.Subscribe(msg.Topic)
		log.Debugf("WebSocket: user %d subscribed to %s", c.UserID(), msg.Topic)
	case ActionUnsubscribe:
		if !c.IsAuthenticated() {
			c.sendError("auth_required", "")
			return true
		}
		c.Unsubscribe(msg.Topic)
		log.Debugf("WebSocket: user %d unsubscribed from %s", c.UserID(), msg.Topic)
	default:
		log.Warningf("WebSocket: unknown action %q", msg.Action)
	}
	return true
}

func (c *Connection) handleAuth(ctx context.Context, token string) bool {
	if c.IsAuthenticated() {
		c.sendError("already_authenticated", "")
		return true
	}

	userID, err := ValidateToken(token)
	if err != nil {
		log.Debugf("WebSocket: auth failed: %v", err)
		// Write the error directly to the websocket since ReadLoop will close the
		// connection immediately after we return false, before WriteLoop can drain the channel.
		c.writeMessageDirect(ctx, OutgoingMessage{Error: "invalid_token"})
		return false
	}

	c.mu.Lock()
	c.userID = userID
	c.authenticated = true
	c.mu.Unlock()

	c.hub.Register(c)

	// Send auth success
	select {
	case c.send <- OutgoingMessage{Action: ActionAuthSuccess, Success: true}:
	default:
		log.Warningf("WebSocket: send buffer full for user %d", userID)
	}

	log.Debugf("WebSocket: user %d authenticated", userID)
	return true
}

// writeMessageDirect writes a message directly to the websocket, bypassing the send channel.
// Use this when the message must be sent before the connection is closed.
func (c *Connection) writeMessageDirect(ctx context.Context, msg OutgoingMessage) {
	writeCtx, cancel := context.WithTimeout(ctx, writeTimeout)
	defer cancel()
	data, err := json.Marshal(msg)
	if err != nil {
		log.Errorf("WebSocket: marshal error: %v", err)
		return
	}
	if err := c.ws.Write(writeCtx, websocket.MessageText, data); err != nil {
		log.Debugf("WebSocket: direct write error: %v", err)
	}
}

func (c *Connection) sendError(errMsg, topic string) {
	select {
	case c.send <- OutgoingMessage{Error: errMsg, Topic: topic}:
	default:
		log.Warningf("WebSocket: send buffer full, dropping error")
	}
}

// WriteLoop drains the send channel and writes messages to the WebSocket.
// It also sends periodic pings.
func (c *Connection) WriteLoop(ctx context.Context, cancel context.CancelFunc) {
	defer cancel()
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				return
			}
			writeCtx, cancel := context.WithTimeout(ctx, writeTimeout)
			data, err := json.Marshal(msg)
			if err != nil {
				cancel()
				log.Errorf("WebSocket: marshal error: %v", err)
				continue
			}
			err = c.ws.Write(writeCtx, websocket.MessageText, data)
			cancel()
			if err != nil {
				log.Debugf("WebSocket: write error for user %d: %v", c.UserID(), err)
				return
			}
		case <-ticker.C:
			pingCtx, cancel := context.WithTimeout(ctx, writeTimeout)
			err := c.ws.Ping(pingCtx)
			cancel()
			if err != nil {
				log.Debugf("WebSocket: ping error for user %d: %v", c.UserID(), err)
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

// validTopics is the set of topics clients are allowed to subscribe to.
var validTopics = map[string]bool{
	"notifications": true,
}

func isValidTopic(topic string) bool {
	return validTopics[topic]
}
