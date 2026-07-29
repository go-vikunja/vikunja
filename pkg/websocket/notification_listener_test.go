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
	"encoding/json"
	"sync"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/notifications"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

// Only this test needs a database, so the engine is built lazily instead of in
// TestMain where it would slow down every other test in the package.
var notificationListenerDBOnce sync.Once

// Fixtures: project 3 is shared with user1 directly and via team 1.
func TestNotificationListener(t *testing.T) {
	t.Run("pushes a notification about a project the user can still read", func(t *testing.T) {
		s := setupNotificationListenerTest(t)
		InitHub()
		conn := notificationConn(1)
		GetHub().Register(conn)

		dbn := storeNotification(t, s, 1, taskCommentNotificationForProject(3))
		handleNotificationCreated(t, s, dbn.ID, 1)

		require.Len(t, conn.send, 1)
		msg := <-conn.send
		assert.Equal(t, "notification.created", msg.Event)
		pushed, ok := msg.Data.(*notifications.DatabaseNotification)
		require.True(t, ok, "payload must be the notification itself")
		assert.Equal(t, dbn.ID, pushed.ID)
	})

	t.Run("does not push once project access is revoked", func(t *testing.T) {
		s := setupNotificationListenerTest(t)
		InitHub()
		conn := notificationConn(1)
		GetHub().Register(conn)

		dbn := storeNotification(t, s, 1, taskCommentNotificationForProject(3))
		unshareProject(t, s, 3, 1)
		handleNotificationCreated(t, s, dbn.ID, 1)

		assert.Empty(t, conn.send)
	})
}

func setupNotificationListenerTest(t *testing.T) *xorm.Session {
	t.Helper()
	notificationListenerDBOnce.Do(func() {
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

func notificationConn(userID int64) *Connection {
	return &Connection{
		userID:        userID,
		subscriptions: map[string]bool{"notification.created": true},
		send:          make(chan OutgoingMessage, 16),
	}
}

// Commits first: the listener opens its own session, so anything still inside
// the test transaction is invisible to it.
func handleNotificationCreated(t *testing.T, s *xorm.Session, notificationID, userID int64) {
	t.Helper()
	require.NoError(t, s.Commit())
	events.TestListener(t,
		&notifications.NotificationCreatedEvent{NotificationID: notificationID, UserID: userID},
		&NotificationListener{},
	)
}

func taskCommentNotificationForProject(projectID int64) *models.TaskCommentNotification {
	return &models.TaskCommentNotification{
		Doer:    &user.User{ID: 3},
		Task:    &models.Task{ID: 32, ProjectID: projectID},
		Comment: &models.TaskComment{ID: 1, Comment: "a secret comment"},
		Project: &models.Project{ID: projectID, Title: "a secret project"},
	}
}

func storeNotification(t *testing.T, s *xorm.Session, notifiableID int64, n notifications.Notification) *notifications.DatabaseNotification {
	t.Helper()
	content, err := json.Marshal(n)
	require.NoError(t, err)
	dbn := &notifications.DatabaseNotification{
		NotifiableID: notifiableID,
		Notification: json.RawMessage(content),
		Name:         n.Name(),
		ProjectID:    notifications.ProjectIDOf(n),
	}
	_, err = s.Insert(dbn)
	require.NoError(t, err)
	return dbn
}

func unshareProject(t *testing.T, s *xorm.Session, projectID, userID int64) {
	t.Helper()
	_, err := s.Where("project_id = ? AND user_id = ?", projectID, userID).Delete(&models.ProjectUser{})
	require.NoError(t, err)
	_, err = s.Where("project_id = ?", projectID).Delete(&models.TeamProject{})
	require.NoError(t, err)
}
