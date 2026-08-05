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

	"github.com/stretchr/testify/require"
)

// TestTaskAssignee_DoerHasDisplayName guards against the regression in #2720: the doer attached to
// notification events was built straight from the JWT (id + username only), so notifications and
// emails rendered the auto-generated username instead of the user's display Name. The dispatch sites
// now resolve the full user from the database, so the doer must carry the display Name even when the
// acting auth object only has id + username (as GetUserFromClaims produces).
func TestTaskAssignee_DoerHasDisplayName(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	// Mimics the partial user GetUserFromClaims builds from a JWT: id + username, no Name.
	// user12 has the display name "Name with spaces" in the fixtures and owns project 23.
	doer := &user.User{ID: 12, Username: "user12"}
	require.Equal(t, "user12", doer.GetName(), "the auth doer must start without a display name")

	task := &Task{Title: "assign me", ProjectID: 23}
	require.NoError(t, task.Create(s, doer))

	events.ClearDispatchedEvents()

	ta := &TaskAssginee{TaskID: task.ID, UserID: 12}
	require.NoError(t, ta.Create(s, doer))
	require.NoError(t, s.Commit())

	events.DispatchPending(context.Background(), s)

	dispatched := events.GetDispatchedEvents((&TaskAssigneeCreatedEvent{}).Name())
	require.Len(t, dispatched, 1)
	ev := dispatched[0].(*TaskAssigneeCreatedEvent)
	require.NotNil(t, ev.Doer)
	require.Equal(t, "Name with spaces", ev.Doer.GetName(),
		"notification doer must carry the display Name, not the username")
}

// TestDoerFromAuth_DisabledUser ensures resolving the event doer keeps working when acting on behalf
// of a disabled account (e.g. user deletion deletes that user's tasks). The full user is still
// returned with its display name, the disabled status error is swallowed.
func TestDoerFromAuth_DisabledUser(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	// user17 is disabled in the fixtures.
	_, err := user.GetUserByID(s, 17)
	require.Error(t, err, "fixture user17 is expected to be disabled")
	require.True(t, user.IsErrAccountDisabled(err))

	doer := doerFromAuth(s, &user.User{ID: 17, Username: "user17"})
	require.NotNil(t, doer)
	require.Equal(t, int64(17), doer.ID)
}
