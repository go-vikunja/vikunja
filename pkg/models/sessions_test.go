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

package models

import (
	"context"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// GHSA-4hv6-xc92-j86g
func TestDeleteSessionByID(t *testing.T) {
	const sid = "550e8400-e29b-41d4-a716-446655440003" // user 2

	t.Run("revokes the session once committed", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		events.ClearDispatchedEvents()
		s := db.NewSession()
		defer s.Close()

		session, err := DeleteSessionByID(s, sid)
		require.NoError(t, err)
		assert.Equal(t, int64(2), session.UserID)
		assert.Zero(t, events.CountDispatchedEvents((&SessionsRevokedEvent{}).Name()))

		require.NoError(t, s.Commit())
		events.DispatchPending(context.Background(), s)

		db.AssertMissing(t, "sessions", map[string]interface{}{"id": sid})
		assert.Equal(t, &SessionsRevokedEvent{UserID: 2, SessionID: sid}, singleDispatchedEvent[*SessionsRevokedEvent](t))
	})

	t.Run("unknown session", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		events.ClearDispatchedEvents()
		s := db.NewSession()
		defer s.Close()

		_, err := DeleteSessionByID(s, "00000000-0000-0000-0000-000000000000")
		require.True(t, IsErrSessionNotFound(err), "unexpected error: %v", err)
		events.DispatchPending(context.Background(), s)
		assert.Zero(t, events.CountDispatchedEvents((&SessionsRevokedEvent{}).Name()))
	})
}

// GHSA-4hv6-xc92-j86g
func TestSessionDelete(t *testing.T) {
	const sid = "550e8400-e29b-41d4-a716-446655440003" // user 2

	t.Run("revokes the session once committed", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		events.ClearDispatchedEvents()
		s := db.NewSession()
		defer s.Close()

		require.NoError(t, (&Session{ID: sid}).Delete(s, &user.User{ID: 2}))
		assert.Zero(t, events.CountDispatchedEvents((&SessionsRevokedEvent{}).Name()))

		require.NoError(t, s.Commit())
		events.DispatchPending(context.Background(), s)

		db.AssertMissing(t, "sessions", map[string]interface{}{"id": sid})
		assert.Equal(t, &SessionsRevokedEvent{UserID: 2, SessionID: sid}, singleDispatchedEvent[*SessionsRevokedEvent](t))
	})

	t.Run("another user's session", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		events.ClearDispatchedEvents()
		s := db.NewSession()
		defer s.Close()

		require.NoError(t, (&Session{ID: sid}).Delete(s, &user.User{ID: 1}))
		require.NoError(t, s.Commit())
		events.DispatchPending(context.Background(), s)

		db.AssertExists(t, "sessions", map[string]interface{}{"id": sid}, false)
		assert.Zero(t, events.CountDispatchedEvents((&SessionsRevokedEvent{}).Name()))
	})
}

// GHSA-4hv6-xc92-j86g
func TestDeleteAllUserSessions(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	events.ClearDispatchedEvents()
	s := db.NewSession()
	defer s.Close()

	require.NoError(t, DeleteAllUserSessions(s, 2))
	assert.Zero(t, events.CountDispatchedEvents((&SessionsRevokedEvent{}).Name()))

	require.NoError(t, s.Commit())
	events.DispatchPending(context.Background(), s)

	db.AssertMissing(t, "sessions", map[string]interface{}{"user_id": 2})
	assert.Equal(t, &SessionsRevokedEvent{UserID: 2}, singleDispatchedEvent[*SessionsRevokedEvent](t))
}
