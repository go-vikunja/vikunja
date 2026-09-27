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
	"sync"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

// Only the database-backed tests need an engine; everything else in this file
// runs against in-memory structures.
var taskListenerDBOnce sync.Once

func setupTaskListenerTest(t *testing.T) *xorm.Session {
	t.Helper()
	taskListenerDBOnce.Do(func() {
		// The fixture loader cleans every registered table, so files and users
		// have to be set up before models.
		files.InitTests()
		user.InitTests()
		models.SetupTests()
		events.Fake()
	})
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func taskConn(userID int64, events ...string) *Connection {
	subs := make(map[string]bool)
	for _, e := range events {
		subs[e] = true
	}
	return &Connection{
		userID:        userID,
		subscriptions: subs,
		send:          make(chan OutgoingMessage, 16),
	}
}

//
// Auth
//

func TestAuthenticateWithAPIToken(t *testing.T) {
	setupTaskListenerTest(t)
	conn := &Connection{subscriptions: map[string]bool{}}

	// Fixture token 1 belongs to user1, expires 2099.
	userID, err := conn.authenticate("tk_2eef46f40ebab3304919ab2e7e39993f75f29d2e")
	require.NoError(t, err)
	assert.Equal(t, int64(1), userID)
}

func TestAuthenticateWithExpiredAPIToken(t *testing.T) {
	setupTaskListenerTest(t)
	conn := &Connection{subscriptions: map[string]bool{}}

	// Fixture token 2 expired 2023-01-01.
	_, err := conn.authenticate("tk_a5e6f92ddbad68f49ee2c63e52174db0235008c8")
	require.Error(t, err)
}

func TestAuthenticateWithUnknownAPIToken(t *testing.T) {
	setupTaskListenerTest(t)
	conn := &Connection{subscriptions: map[string]bool{}}

	_, err := conn.authenticate("tk_0000000000000000000000000000000000000000")
	require.Error(t, err)
}

func TestAuthenticateWithDisabledOwnerAPIToken(t *testing.T) {
	setupTaskListenerTest(t)
	conn := &Connection{subscriptions: map[string]bool{}}

	// Fixture token 4 belongs to the disabled user 17.
	_, err := conn.authenticate("tk_disabled_user_test_token_000000001234abcd")
	require.Error(t, err)
}

func TestAuthenticateWithJWT(t *testing.T) {
	setupTaskListenerTest(t)
	conn := &Connection{subscriptions: map[string]bool{}}

	token, err := auth.NewUserJWTAuthtoken(&user.User{ID: 1, Username: "user1"}, "test-session")
	require.NoError(t, err)

	userID, err := conn.authenticate(token)
	require.NoError(t, err)
	assert.Equal(t, int64(1), userID)
}

func TestAuthenticateWithGarbageToken(t *testing.T) {
	conn := &Connection{subscriptions: map[string]bool{}}

	_, err := conn.authenticate("not-a-token")
	require.Error(t, err)
}

//
// Task event fan-out
//

func TestTaskEventListenerFanOut(t *testing.T) {
	t.Run("pushes task events to subscribed users with project access", func(t *testing.T) {
		InitHub()
		// Project 3 is shared with user1 (directly) and user2 (via team 1).
		conn1 := taskConn(1, "task.created")
		conn2 := taskConn(2, "task.created")
		GetHub().Register(conn1)
		GetHub().Register(conn2)

		s := setupTaskListenerTest(t)
		require.NoError(t, s.Commit()) // the listener opens its own session
		events.TestListener(t,
			&models.TaskCreatedEvent{Task: &models.Task{ID: 32, ProjectID: 3}},
			&TaskEventListener{wsEvent: "task.created"},
		)

		require.Len(t, conn1.send, 1, "user1 (direct share) must receive the event")
		msg := <-conn1.send
		assert.Equal(t, "task.created", msg.Event)
		require.Len(t, conn2.send, 1, "user2 (team share) must receive the event")
	})

	t.Run("does not push to users without project access", func(t *testing.T) {
		InitHub()
		// user7 has no share on project 3.
		conn := taskConn(7, "task.created")
		GetHub().Register(conn)

		s := setupTaskListenerTest(t)
		require.NoError(t, s.Commit())
		events.TestListener(t,
			&models.TaskCreatedEvent{Task: &models.Task{ID: 32, ProjectID: 3}},
			&TaskEventListener{wsEvent: "task.created"},
		)

		assert.Empty(t, conn.send)
	})

	t.Run("does not push to unsubscribed connections", func(t *testing.T) {
		InitHub()
		conn := taskConn(1, "notification.created") // subscribed, but not to task events
		GetHub().Register(conn)

		s := setupTaskListenerTest(t)
		require.NoError(t, s.Commit())
		events.TestListener(t,
			&models.TaskCreatedEvent{Task: &models.Task{ID: 32, ProjectID: 3}},
			&TaskEventListener{wsEvent: "task.created"},
		)

		assert.Empty(t, conn.send)
	})

	t.Run("does not push once project access is revoked", func(t *testing.T) {
		InitHub()
		conn := taskConn(1, "task.updated")
		GetHub().Register(conn)

		s := setupTaskListenerTest(t)
		unshareProject(t, s, 3, 1)
		require.NoError(t, s.Commit())
		events.TestListener(t,
			&models.TaskUpdatedEvent{Task: &models.Task{ID: 32, ProjectID: 3}},
			&TaskEventListener{wsEvent: "task.updated"},
		)

		assert.Empty(t, conn.send)
	})
}

//
// Comment event fan-out + mentions
//

func TestTaskCommentEventListener(t *testing.T) {
	t.Run("carries mention metadata for mentioned users", func(t *testing.T) {
		InitHub()
		conn := taskConn(1, "task.comment.created")
		GetHub().Register(conn)

		s := setupTaskListenerTest(t)
		require.NoError(t, s.Commit())
		events.TestListener(t,
			&models.TaskCommentCreatedEvent{
				Task:    &models.Task{ID: 32, ProjectID: 3},
				Comment: &models.TaskComment{ID: 1, TaskID: 32, Comment: `a mention: <mention-user data-id="user2">user2</mention-user>`},
			},
			&TaskCommentEventListener{wsEvent: "task.comment.created"},
		)

		require.Len(t, conn.send, 1)
		msg := <-conn.send
		assert.Equal(t, "task.comment.created", msg.Event)
		payload, ok := msg.Data.(*CommentEventPayload)
		require.True(t, ok, "payload must be the comment event payload")
		require.Len(t, payload.Mentions, 1)
		assert.Equal(t, MentionedUser{ID: 2, Username: "user2"}, payload.Mentions[0])
	})

	t.Run("empty mentions for comments without mentions", func(t *testing.T) {
		InitHub()
		conn := taskConn(1, "task.comment.edited")
		GetHub().Register(conn)

		s := setupTaskListenerTest(t)
		require.NoError(t, s.Commit())
		events.TestListener(t,
			&models.TaskCommentUpdatedEvent{
				Task:    &models.Task{ID: 32, ProjectID: 3},
				Comment: &models.TaskComment{ID: 1, TaskID: 32, Comment: "plain comment"},
			},
			&TaskCommentEventListener{wsEvent: "task.comment.edited"},
		)

		require.Len(t, conn.send, 1)
		msg := <-conn.send
		payload, ok := msg.Data.(*CommentEventPayload)
		require.True(t, ok)
		assert.Empty(t, payload.Mentions)
	})

	t.Run("scopes delivery by project access", func(t *testing.T) {
		InitHub()
		conn := taskConn(7, "task.comment.deleted")
		GetHub().Register(conn)

		s := setupTaskListenerTest(t)
		require.NoError(t, s.Commit())
		events.TestListener(t,
			&models.TaskCommentDeletedEvent{
				Task:    &models.Task{ID: 32, ProjectID: 3},
				Comment: &models.TaskComment{ID: 1, TaskID: 32},
			},
			&TaskCommentEventListener{wsEvent: "task.comment.deleted"},
		)

		assert.Empty(t, conn.send)
	})
}

//
// Subscription surface
//

func TestConnectionAllowsTaskAndCommentEvents(t *testing.T) {
	conn := &Connection{
		userID:        1,
		authenticated: true,
		subscriptions: make(map[string]bool),
		send:          make(chan OutgoingMessage, 16),
	}

	for _, event := range []string{
		"task.created", "task.updated", "task.deleted",
		"task.comment.created", "task.comment.edited", "task.comment.deleted",
	} {
		conn.handleMessage(context.Background(), IncomingMessage{Action: ActionSubscribe, Event: event})
		assert.True(t, conn.IsSubscribed(event), "client must be able to subscribe to %s", event)
	}
}
