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
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	ws "github.com/coder/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func redisTestClient(t *testing.T) *redis.Client {
	t.Helper()
	addr := os.Getenv("VIKUNJA_TEST_REDIS_ADDR")
	if addr == "" {
		t.Skip("set VIKUNJA_TEST_REDIS_ADDR to run Redis integration tests")
	}
	client := redis.NewClient(&redis.Options{Addr: addr})
	t.Cleanup(func() { _ = client.Close() })
	require.NoError(t, client.Ping(context.Background()).Err())
	return client
}

func subscribedConnection(h *Hub, userID int64) *Connection {
	c := &Connection{
		userID:        userID,
		subscriptions: map[string]bool{"notification.created": true},
		send:          make(chan OutgoingMessage, 64),
	}
	h.Register(c)
	return c
}

func TestRedisHubFanout(t *testing.T) {
	client := redisTestClient(t)
	channel := fmt.Sprintf("vikunja-test:%d", time.Now().UnixNano())
	a, b, isolated := NewHub(), NewHub(), NewHub()
	for _, h := range []*Hub{a, b, isolated} {
		t.Cleanup(h.Shutdown)
	}
	require.NoError(t, a.StartRedis(context.Background(), client, channel))
	require.NoError(t, b.StartRedis(context.Background(), client, channel))
	require.NoError(t, isolated.StartRedis(context.Background(), client, channel+":other"))
	local := subscribedConnection(a, 10)
	remote := subscribedConnection(b, 10)
	otherUser := subscribedConnection(b, 20)
	otherInstall := subscribedConnection(isolated, 10)

	a.PublishForUser(10, "notification.created", map[string]int{"id": 42})
	select {
	case msg := <-local.send:
		assert.Equal(t, "notification.created", msg.Event)
	case <-time.After(time.Second):
		t.Fatal("local message missing")
	}
	select {
	case msg := <-remote.send:
		assert.Equal(t, "notification.created", msg.Event)
	case <-time.After(time.Second):
		t.Fatal("remote message missing")
	}
	assert.Eventually(t, func() bool { return len(local.send) == 0 }, time.Second, time.Millisecond)
	assert.Never(t, func() bool { return len(local.send)+len(otherUser.send)+len(otherInstall.send) > 0 }, 50*time.Millisecond, time.Millisecond)
}

func TestHubShutdownRejectsConnections(t *testing.T) {
	h := NewHub()
	assert.True(t, h.Available())
	h.Shutdown()
	assert.False(t, h.Available())
	c := subscribedConnection(h, 10)
	h.PublishForUser(10, "notification.created", nil)
	assert.Empty(t, c.send)
	assert.Empty(t, h.connections)
}

func TestHubShutdownClosesUnauthenticatedSocket(t *testing.T) {
	h := NewHub()
	t.Cleanup(h.Shutdown)
	tracked := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		socket, err := ws.Accept(w, r, nil)
		if err != nil {
			return
		}
		conn := NewConnection(socket, h)
		h.track(conn)
		close(tracked)
	}))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	socket, _, err := ws.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http"), nil) //nolint:bodyclose // Dial owns the handshake body.
	require.NoError(t, err)
	defer func() { _ = socket.CloseNow() }()
	<-tracked
	h.Shutdown()
	_, _, err = socket.Read(ctx)
	require.Error(t, err)
	assert.NoError(t, ctx.Err(), "socket should close before the read deadline")
}

func TestRedisHubUnavailableBroker(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr:        "127.0.0.1:1",
		MaxRetries:  -1,
		DialTimeout: 50 * time.Millisecond,
	})
	defer client.Close()
	h := NewHub()
	defer h.Shutdown()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	require.Error(t, h.StartRedis(ctx, client, "test"))
	assert.False(t, h.Available())
}

func TestRedisHubReconnectsAfterSubscriptionLoss(t *testing.T) {
	client := redisTestClient(t)
	h := NewHub()
	defer h.Shutdown()
	channel := fmt.Sprintf("vikunja-test:%d", time.Now().UnixNano())
	require.NoError(t, h.StartRedis(context.Background(), client, channel))
	c := subscribedConnection(h, 10)
	// The integration Redis is dedicated to tests.
	require.NoError(t, client.ClientKillByFilter(context.Background(), "TYPE", "pubsub").Err())
	require.Eventually(t, func() bool {
		h.mu.RLock()
		defer h.mu.RUnlock()
		return len(h.connections) == 0
	}, 2*time.Second, time.Millisecond)
	assert.Empty(t, c.send)
	require.Eventually(t, h.Available, 2*time.Second, time.Millisecond)
	c = subscribedConnection(h, 10)
	h.PublishForUser(10, "notification.created", 42)
	assert.Len(t, c.send, 1)
}

func BenchmarkHubTwentyThousandConnections(b *testing.B) {
	h := NewHub()
	for id := int64(1); id <= 20000; id++ {
		subscribedConnection(h, id)
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		id := int64(n%20000 + 1)
		h.PublishForUser(id, "notification.created", n)
		<-h.connections[id][0].send
	}
}
