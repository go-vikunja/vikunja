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
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/user"

	ws "github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/stretchr/testify/require"
)

type loadNotification struct {
	UserID int64 `json:"user_id"`
	SentAt int64 `json:"sent_at"`
}

func TestWebsocketClusterLoad(t *testing.T) {
	rawCount := os.Getenv("VIKUNJA_LOAD_CONNECTIONS")
	if rawCount == "" {
		t.Skip("set VIKUNJA_LOAD_CONNECTIONS and VIKUNJA_TEST_REDIS_ADDR for the opt-in socket load test")
	}
	count, err := strconv.Atoi(rawCount)
	require.NoError(t, err)
	require.GreaterOrEqual(t, count, 10)
	require.NotEmpty(t, os.Getenv("VIKUNJA_TEST_REDIS_ADDR"), "the opt-in load test requires a real Redis address")
	var hold time.Duration
	if value := os.Getenv("VIKUNJA_LOAD_HOLD"); value != "" {
		hold, err = time.ParseDuration(value)
		require.NoError(t, err, "VIKUNJA_LOAD_HOLD must be a duration")
		require.GreaterOrEqual(t, hold, time.Duration(0))
		require.Less(t, hold, 5*time.Minute, "hold must leave room for ramp and delivery within the test deadline")
	}
	client := redisTestClient(t)

	secret := config.ServiceSecret.GetString()
	config.ServiceSecret.Set("websocket-load-test-secret")
	t.Cleanup(func() { config.ServiceSecret.Set(secret) })
	log.ConfigureStandardLogger(false, "off", "", "ERROR", "text")
	t.Cleanup(func() {
		log.ConfigureStandardLogger(config.LogEnabled.GetBool(), config.LogStandard.GetString(), config.LogPath.GetString(), config.LogLevel.GetString(), config.LogFormat.GetString())
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	channel := fmt.Sprintf("vikunja-load:%d", time.Now().UnixNano())
	hubs := make([]*Hub, 10)
	servers := make([]*httptest.Server, len(hubs))
	sockets := make([]*ws.Conn, count)
	var serverLoops, readers sync.WaitGroup
	var stopping atomic.Bool
	var cleanupOnce sync.Once
	cleanup := func() {
		cleanupOnce.Do(func() {
			stopping.Store(true)
			cancel()
			for _, server := range servers {
				if server != nil {
					server.Close()
				}
			}
			for _, hub := range hubs {
				if hub != nil {
					hub.Shutdown()
				}
			}
			for _, socket := range sockets {
				if socket != nil {
					_ = socket.CloseNow()
				}
			}
			drained := make(chan struct{})
			go func() {
				serverLoops.Wait()
				readers.Wait()
				close(drained)
			}()
			select {
			case <-drained:
			case <-time.After(30 * time.Second):
				t.Error("socket loops did not drain within 30 seconds")
			}
		})
	}
	t.Cleanup(cleanup)

	runtime.GC()
	var baseline runtime.MemStats
	runtime.ReadMemStats(&baseline)
	baselineGoroutines := runtime.NumGoroutine()
	for index := range hubs {
		hub := NewHub()
		hubs[index] = hub
		require.NoError(t, hub.StartRedis(ctx, client, channel))
		// UpgradeHandler uses a global hub; the adapter only supplies a distinct hub per server.
		servers[index] = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			serverLoops.Add(1)
			defer serverLoops.Done()
			socket, acceptErr := ws.Accept(w, r, nil)
			if acceptErr != nil {
				return
			}
			conn := NewConnection(socket, hub)
			if !hub.track(conn) {
				conn.closeNow()
				return
			}
			connectionCtx, connectionCancel := context.WithCancel(ctx)
			serverLoops.Add(2)
			go func() {
				defer serverLoops.Done()
				conn.ReadLoop(connectionCtx, connectionCancel)
			}()
			go func() {
				defer serverLoops.Done()
				conn.WriteLoop(connectionCtx, connectionCancel)
			}()
		}))
	}

	failed := make(chan error, 1)
	var failOnce sync.Once
	fail := func(err error) {
		failOnce.Do(func() {
			failed <- err
			cancel()
		})
	}
	var connected, delivered atomic.Int64
	latencies := make([]time.Duration, count)
	rampStarted := time.Now()
	runLoadWorkers(ctx, count, 128, func(index int) {
		userID := int64(index + 1)
		token, tokenErr := auth.NewUserJWTAuthtoken(&user.User{
			ID:       userID,
			Username: fmt.Sprintf("load-%d", userID),
		}, "load-test")
		if tokenErr != nil {
			fail(tokenErr)
			return
		}
		dialCtx, dialCancel := context.WithTimeout(ctx, 15*time.Second)
		defer dialCancel()
		url := "ws" + strings.TrimPrefix(servers[index%len(servers)].URL, "http")
		socket, response, dialErr := ws.Dial(dialCtx, url, nil)
		if dialErr != nil {
			if response != nil {
				_ = response.Body.Close()
			}
			fail(fmt.Errorf("dial user %d: %w", userID, dialErr))
			return
		}
		sockets[index] = socket
		if err := wsjson.Write(dialCtx, socket, IncomingMessage{
			Action: ActionAuth,
			Token:  token,
		}); err != nil {
			fail(fmt.Errorf("authenticate user %d: %w", userID, err))
			return
		}
		var authenticated OutgoingMessage
		if err := wsjson.Read(dialCtx, socket, &authenticated); err != nil {
			fail(fmt.Errorf("read authentication for user %d: %w", userID, err))
			return
		}
		if authenticated.Action != ActionAuthSuccess || !authenticated.Success {
			fail(fmt.Errorf("user %d was not authenticated", userID))
			return
		}
		if err := wsjson.Write(dialCtx, socket, IncomingMessage{
			Action: ActionSubscribe,
			Event:  "notification.created",
		}); err != nil {
			fail(fmt.Errorf("subscribe user %d: %w", userID, err))
			return
		}
		readers.Add(1)
		go func() {
			defer readers.Done()
			received := false
			for {
				var message struct {
					Event string           `json:"event"`
					Data  loadNotification `json:"data"`
				}
				if err := wsjson.Read(ctx, socket, &message); err != nil {
					if !stopping.Load() && ctx.Err() == nil {
						fail(fmt.Errorf("user %d disconnected before shutdown: %w", userID, err))
					}
					return
				}
				if message.Event != "notification.created" || message.Data.UserID != userID || received {
					fail(fmt.Errorf("unexpected, unauthorized or duplicate notification for user %d: %+v", userID, message))
					return
				}
				received = true
				latencies[index] = time.Duration(time.Now().UnixNano() - message.Data.SentAt)
				delivered.Add(1)
			}
		}()
		if total := connected.Add(1); total%5000 == 0 {
			t.Logf("authenticated %d/%d sockets", total, count)
		}
	})
	select {
	case err := <-failed:
		t.Fatal(err)
	default:
	}
	require.NoError(t, ctx.Err())
	require.EqualValues(t, count, connected.Load())
	require.Eventually(t, func() bool {
		subscribed := 0
		for _, hub := range hubs {
			hub.mu.RLock()
			for _, connections := range hub.connections {
				for _, conn := range connections {
					if conn.IsSubscribed("notification.created") {
						subscribed++
					}
				}
			}
			hub.mu.RUnlock()
		}
		return subscribed == count
	}, 10*time.Second, 10*time.Millisecond)
	rampDuration := time.Since(rampStarted)
	if hold > 0 {
		t.Logf("holding %d authenticated sockets for %s", count, hold)
		timer := time.NewTimer(hold)
		defer timer.Stop()
		select {
		case <-timer.C:
		case err := <-failed:
			t.Fatal(err)
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		}
	}
	runtime.GC()
	var loaded runtime.MemStats
	runtime.ReadMemStats(&loaded)
	loadedGoroutines := runtime.NumGoroutine()

	deliveryStarted := time.Now()
	runLoadWorkers(ctx, count, 8, func(index int) {
		hubs[(index+1)%len(hubs)].PublishForUser(int64(index+1), "notification.created", loadNotification{
			UserID: int64(index + 1),
			SentAt: time.Now().UnixNano(),
		})
	})
	require.Eventually(t, func() bool {
		return delivered.Load() == int64(count) || ctx.Err() != nil
	}, 45*time.Second, 10*time.Millisecond)
	select {
	case err := <-failed:
		t.Fatal(err)
	default:
	}
	require.NoError(t, ctx.Err())
	require.EqualValues(t, count, delivered.Load())
	deliveryDuration := time.Since(deliveryStarted)

	drainStarted := time.Now()
	cleanup()
	drainDuration := time.Since(drainStarted)
	for _, hub := range hubs {
		hub.mu.RLock()
		remaining := len(hub.connections) + len(hub.all)
		hub.mu.RUnlock()
		require.Zero(t, remaining, "shutdown must remove authenticated and unauthenticated sockets")
	}
	slices.Sort(latencies)
	runtime.GC()
	var drained runtime.MemStats
	runtime.ReadMemStats(&drained)
	result, err := json.Marshal(map[string]any{
		"connections":         count,
		"servers":             len(hubs),
		"delivered":           delivered.Load(),
		"ramp_ms":             rampDuration.Milliseconds(),
		"hold_ms":             hold.Milliseconds(),
		"delivery_ms":         deliveryDuration.Milliseconds(),
		"drain_ms":            drainDuration.Milliseconds(),
		"latency_p50_ms":      latencies[(count-1)*50/100].Milliseconds(),
		"latency_p95_ms":      latencies[(count-1)*95/100].Milliseconds(),
		"latency_p99_ms":      latencies[(count-1)*99/100].Milliseconds(),
		"baseline_heap_bytes": baseline.HeapAlloc,
		"loaded_heap_bytes":   loaded.HeapAlloc,
		"drained_heap_bytes":  drained.HeapAlloc,
		"loaded_go_sys_bytes": loaded.Sys,
		"baseline_goroutines": baselineGoroutines,
		"loaded_goroutines":   loadedGoroutines,
		"drained_goroutines":  runtime.NumGoroutine(),
		"scope":               "one process: ten server hubs plus all socket clients; excludes database, HTTP CRUD and Kubernetes network",
	})
	require.NoError(t, err)
	t.Logf("socket load result: %s", result)
}

func runLoadWorkers(ctx context.Context, count, concurrency int, run func(int)) {
	var next atomic.Int64
	var workers sync.WaitGroup
	for range min(count, concurrency) {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for ctx.Err() == nil {
				index := int(next.Add(1) - 1)
				if index >= count {
					return
				}
				run(index)
			}
		}()
	}
	workers.Wait()
}
