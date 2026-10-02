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
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/user"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Fixture sessions: 1 and 2 belong to user 1, 3 to user 2.
const (
	sessionUser1A = "550e8400-e29b-41d4-a716-446655440001"
	sessionUser1B = "550e8400-e29b-41d4-a716-446655440002"
	sessionUser2  = "550e8400-e29b-41d4-a716-446655440003"
)

// Uses the global hub because the listeners do.
func startSocketServer(t *testing.T) string {
	t.Helper()
	InitHub()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		conn := NewConnection(ws, GetHub())
		ctx, cancel := context.WithCancel(r.Context())
		go conn.WriteLoop(ctx, cancel)
		conn.ReadLoop(ctx, cancel)
	}))
	t.Cleanup(srv.Close)
	return "ws" + strings.TrimPrefix(srv.URL, "http")
}

func userToken(t *testing.T, userID int64, sid string) string {
	t.Helper()
	token, err := auth.NewUserJWTAuthtoken(&user.User{ID: userID, Username: "user"}, sid)
	require.NoError(t, err)
	return token
}

func stubCheckUserTokenSession(t *testing.T, f func(*auth.UserTokenClaims) error) {
	t.Helper()
	orig := checkUserTokenSession
	checkUserTokenSession = f
	t.Cleanup(func() { checkUserTokenSession = orig })
}

func dial(t *testing.T, url string) *websocket.Conn {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c, resp, err := websocket.Dial(ctx, url, nil)
	require.NoError(t, err)
	if resp.Body != nil {
		_ = resp.Body.Close()
	}
	t.Cleanup(func() { _ = c.CloseNow() })
	return c
}

func dialAndAuth(t *testing.T, url, token string) (*websocket.Conn, OutgoingMessage) {
	t.Helper()
	c := dial(t, url)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	require.NoError(t, wsjson.Write(ctx, c, IncomingMessage{Action: ActionAuth, Token: token}))
	var reply OutgoingMessage
	require.NoError(t, wsjson.Read(ctx, c, &reply))
	return c, reply
}

func assertClosed(t *testing.T, c *websocket.Conn) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, _, err := c.Read(ctx)
	require.Error(t, err)
	assert.Equal(t, websocket.StatusPolicyViolation, websocket.CloseStatus(err), "unexpected close: %v", err)
}

func assertUnregistered(t *testing.T, userID int64) {
	t.Helper()
	require.Eventually(t, func() bool {
		GetHub().mu.RLock()
		defer GetHub().mu.RUnlock()
		return len(GetHub().connections[userID]) == 0
	}, 2*time.Second, 10*time.Millisecond)
}

func assertOpen(t *testing.T, c *websocket.Conn, userID int64) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	require.NoError(t, wsjson.Write(ctx, c, IncomingMessage{Action: ActionSubscribe, Event: "timer.created"}))
	// The server never acks a subscribe; wait until the hub sees it.
	require.Eventually(t, func() bool {
		GetHub().mu.RLock()
		defer GetHub().mu.RUnlock()
		for _, conn := range GetHub().connections[userID] {
			if conn.IsSubscribed("timer.created") {
				return true
			}
		}
		return false
	}, 2*time.Second, 10*time.Millisecond)
	GetHub().PublishForUser(userID, "timer.created", "ping")

	var msg OutgoingMessage
	require.NoError(t, wsjson.Read(ctx, c, &msg))
	assert.Equal(t, "timer.created", msg.Event)
}

// GHSA-4hv6-xc92-j86g
func TestConnectionAuth(t *testing.T) {
	t.Run("accepts a live session", func(t *testing.T) {
		setupDBTest(t)
		url := startSocketServer(t)

		_, reply := dialAndAuth(t, url, userToken(t, 1, sessionUser1A))
		assert.Equal(t, ActionAuthSuccess, reply.Action)
	})
	t.Run("rejects an unknown session", func(t *testing.T) {
		setupDBTest(t)
		url := startSocketServer(t)

		_, reply := dialAndAuth(t, url, userToken(t, 1, "00000000-0000-0000-0000-000000000000"))
		assert.Equal(t, "invalid_token", reply.Error)
		assertUnregistered(t, 1)
	})
	t.Run("rejects another user's session", func(t *testing.T) {
		setupDBTest(t)
		url := startSocketServer(t)

		_, reply := dialAndAuth(t, url, userToken(t, 1, sessionUser2))
		assert.Equal(t, "invalid_token", reply.Error)
		assertUnregistered(t, 1)
	})
	t.Run("rejects a disabled user", func(t *testing.T) {
		s := setupDBTest(t)
		const sid = "550e8400-e29b-41d4-a716-446655440017"
		_, err := s.Insert(&models.Session{ID: sid, UserID: 17, TokenHash: "disabled", LastActive: time.Now()})
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		url := startSocketServer(t)

		_, reply := dialAndAuth(t, url, userToken(t, 17, sid))
		assert.Equal(t, "invalid_token", reply.Error)
		assertUnregistered(t, 17)
	})
	t.Run("rejects a deleted user", func(t *testing.T) {
		s := setupDBTest(t)
		const sid = "550e8400-e29b-41d4-a716-446655449999"
		_, err := s.Insert(&models.Session{
			ID:         sid,
			UserID:     9999,
			TokenHash:  "deleted",
			LastActive: time.Now(),
		})
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		url := startSocketServer(t)

		_, reply := dialAndAuth(t, url, userToken(t, 9999, sid))
		assert.Equal(t, "invalid_token", reply.Error)
		assertUnregistered(t, 9999)
	})
	t.Run("rejects a malformed token", func(t *testing.T) {
		setupDBTest(t)
		url := startSocketServer(t)

		_, reply := dialAndAuth(t, url, "not-a-jwt")
		assert.Equal(t, "invalid_token", reply.Error)
	})
	t.Run("a transient failure closes without invalid_token so the client retries", func(t *testing.T) {
		setupDBTest(t)
		stubCheckUserTokenSession(t, func(*auth.UserTokenClaims) error {
			return errors.New("database unavailable")
		})
		url := startSocketServer(t)

		c := dial(t, url)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		require.NoError(t, wsjson.Write(ctx, c, IncomingMessage{Action: ActionAuth, Token: userToken(t, 1, sessionUser1A)}))
		var reply OutgoingMessage
		err := wsjson.Read(ctx, c, &reply)
		require.Error(t, err, "got reply %+v", reply)
		assert.Equal(t, websocket.StatusTryAgainLater, websocket.CloseStatus(err), "unexpected close: %v", err)
	})
	t.Run("closes the socket when the token expires", func(t *testing.T) {
		setupDBTest(t)
		ttl := config.ServiceJWTTTLShort.GetInt64()
		config.ServiceJWTTTLShort.Set(1)
		t.Cleanup(func() { config.ServiceJWTTTLShort.Set(ttl) })
		url := startSocketServer(t)

		c, reply := dialAndAuth(t, url, userToken(t, 1, sessionUser1A))
		require.Equal(t, ActionAuthSuccess, reply.Action)
		assertClosed(t, c)
	})
}

// GHSA-4hv6-xc92-j86g
func TestConnectionRevocation(t *testing.T) {
	t.Run("deleting a session closes only its socket", func(t *testing.T) {
		setupDBTest(t)
		url := startSocketServer(t)
		revoked, _ := dialAndAuth(t, url, userToken(t, 1, sessionUser1A))
		other, _ := dialAndAuth(t, url, userToken(t, 1, sessionUser1B))

		events.TestListener(t, &models.SessionsRevokedEvent{UserID: 1, SessionID: sessionUser1A}, &SessionsRevokedListener{})

		assertClosed(t, revoked)
		assertOpen(t, other, 1)
	})
	t.Run("revoking all sessions closes every socket of the user", func(t *testing.T) {
		setupDBTest(t)
		url := startSocketServer(t)
		a, _ := dialAndAuth(t, url, userToken(t, 1, sessionUser1A))
		b, _ := dialAndAuth(t, url, userToken(t, 1, sessionUser1B))
		otherUser, _ := dialAndAuth(t, url, userToken(t, 2, sessionUser2))

		events.TestListener(t, &models.SessionsRevokedEvent{UserID: 1}, &SessionsRevokedListener{})

		assertClosed(t, a)
		assertClosed(t, b)
		assertOpen(t, otherUser, 2)
	})
}
