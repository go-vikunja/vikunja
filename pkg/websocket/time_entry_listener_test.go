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
	"testing"

	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/license"
	"code.vikunja.io/api/pkg/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func timerConn(userID int64) *Connection {
	return &Connection{
		userID:        userID,
		subscriptions: map[string]bool{"timer.created": true, "timer.updated": true, "timer.deleted": true},
		send:          make(chan OutgoingMessage, 16),
	}
}

func TestTimeEntryListener(t *testing.T) {
	t.Run("a create pushes timer.created with the entry to its owner", func(t *testing.T) {
		InitHub()
		conn := timerConn(1)
		GetHub().Register(conn)
		license.SetForTests([]license.Feature{license.FeatureTimeTracking})
		defer license.ResetForTests()

		ev := &models.TimeEntryCreatedEvent{TimeEntry: &models.TimeEntry{ID: 4, UserID: 1}}
		events.TestListener(t, ev, &TimeEntryListener{wsEvent: "timer.created"})

		require.Len(t, conn.send, 1)
		msg := <-conn.send
		assert.Equal(t, "timer.created", msg.Event)
		te, ok := msg.Data.(*models.TimeEntry)
		require.True(t, ok, "payload must be the time entry itself")
		assert.Equal(t, int64(4), te.ID)
	})

	t.Run("an update pushes timer.updated", func(t *testing.T) {
		InitHub()
		conn := timerConn(1)
		GetHub().Register(conn)
		license.SetForTests([]license.Feature{license.FeatureTimeTracking})
		defer license.ResetForTests()

		ev := &models.TimeEntryUpdatedEvent{TimeEntry: &models.TimeEntry{ID: 4, UserID: 1}}
		events.TestListener(t, ev, &TimeEntryListener{wsEvent: "timer.updated"})

		require.Len(t, conn.send, 1)
		assert.Equal(t, "timer.updated", (<-conn.send).Event)
	})

	t.Run("a delete pushes timer.deleted so other tabs drop it", func(t *testing.T) {
		InitHub()
		conn := timerConn(1)
		GetHub().Register(conn)
		license.SetForTests([]license.Feature{license.FeatureTimeTracking})
		defer license.ResetForTests()

		ev := &models.TimeEntryDeletedEvent{TimeEntry: &models.TimeEntry{ID: 4, UserID: 1}}
		events.TestListener(t, ev, &TimeEntryListener{wsEvent: "timer.deleted"})

		require.Len(t, conn.send, 1)
		msg := <-conn.send
		assert.Equal(t, "timer.deleted", msg.Event)
		te, ok := msg.Data.(*models.TimeEntry)
		require.True(t, ok)
		assert.Equal(t, int64(4), te.ID)
	})

	t.Run("does not push when the feature is disabled", func(t *testing.T) {
		InitHub()
		conn := timerConn(1)
		GetHub().Register(conn)
		license.ResetForTests() // free mode

		ev := &models.TimeEntryUpdatedEvent{TimeEntry: &models.TimeEntry{ID: 4, UserID: 1}}
		events.TestListener(t, ev, &TimeEntryListener{wsEvent: "timer.updated"})

		assert.Empty(t, conn.send)
	})

	t.Run("only pushes to the entry owner", func(t *testing.T) {
		InitHub()
		other := timerConn(2)
		GetHub().Register(other)
		license.SetForTests([]license.Feature{license.FeatureTimeTracking})
		defer license.ResetForTests()

		ev := &models.TimeEntryUpdatedEvent{TimeEntry: &models.TimeEntry{ID: 4, UserID: 1}}
		events.TestListener(t, ev, &TimeEntryListener{wsEvent: "timer.updated"})

		assert.Empty(t, other.send, "a different user must not receive the timer update")
	})
}
