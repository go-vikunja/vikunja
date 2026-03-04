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

package webtests

import (
	"net/http"
	"testing"

	apiv1 "code.vikunja.io/api/pkg/routes/api/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserProject(t *testing.T) {
	t.Run("Normal test", func(t *testing.T) {
		rec, err := newTestRequestWithUser(t, http.MethodPost, apiv1.UserList, &testuser1, "", nil, nil)
		require.NoError(t, err)
		assert.Equal(t, "null\n", rec.Body.String())
	})
	t.Run("Search for user3", func(t *testing.T) {
		rec, err := newTestRequestWithUser(t, http.MethodPost, apiv1.UserList, &testuser1, "", map[string][]string{"s": {"user3"}}, nil)
		require.NoError(t, err)
		assert.Contains(t, rec.Body.String(), `user3`)
		assert.NotContains(t, rec.Body.String(), `user1`)
		assert.NotContains(t, rec.Body.String(), `user2`)
		assert.NotContains(t, rec.Body.String(), `user4`)
		assert.NotContains(t, rec.Body.String(), `user5`)
	})
	t.Run("external team member discoverable by name", func(t *testing.T) {
		// User 10 searches for "Some one else" (user 11's name).
		// User 11 has discoverable_by_name=false, but they share external team 14.
		// Should find user 11.
		rec, err := newTestRequestWithUser(t, http.MethodPost, apiv1.UserList, &testuser10, "", map[string][]string{"s": {"Some one else"}}, nil)
		require.NoError(t, err)
		assert.Contains(t, rec.Body.String(), `user11`)
	})
	t.Run("external team member discoverable by email", func(t *testing.T) {
		// User 10 searches for user 11's email.
		// User 11 has discoverable_by_email=false, but they share external team 14.
		// Should find user 11.
		rec, err := newTestRequestWithUser(t, http.MethodPost, apiv1.UserList, &testuser10, "", map[string][]string{"s": {"user11@example.com"}}, nil)
		require.NoError(t, err)
		assert.Contains(t, rec.Body.String(), `user11`)
	})
	t.Run("non-external-team user cannot discover by name", func(t *testing.T) {
		// User 1 searches for "Some one else" (user 11's name).
		// User 1 does NOT share an external team with user 11.
		// User 11 has discoverable_by_name=false.
		// Should NOT find user 11.
		rec, err := newTestRequestWithUser(t, http.MethodPost, apiv1.UserList, &testuser1, "", map[string][]string{"s": {"Some one else"}}, nil)
		require.NoError(t, err)
		assert.NotContains(t, rec.Body.String(), `user11`)
	})
	t.Run("non-external-team user cannot discover by email", func(t *testing.T) {
		// User 1 searches for user 11's email.
		// User 1 does NOT share an external team with user 11.
		// User 11 has discoverable_by_email=false.
		// Should NOT find user 11.
		rec, err := newTestRequestWithUser(t, http.MethodPost, apiv1.UserList, &testuser1, "", map[string][]string{"s": {"user11@example.com"}}, nil)
		require.NoError(t, err)
		assert.NotContains(t, rec.Body.String(), `user11`)
	})
	t.Run("regular team does not bypass discoverability", func(t *testing.T) {
		// User 1 and user 2 share team 1 (a regular team, no external_id).
		// User 2 has discoverable_by_name=false and discoverable_by_email=false.
		// Searching by email should NOT find user 2 (regular team doesn't bypass).
		rec, err := newTestRequestWithUser(t, http.MethodPost, apiv1.UserList, &testuser1, "", map[string][]string{"s": {"user2@example.com"}}, nil)
		require.NoError(t, err)
		assert.NotContains(t, rec.Body.String(), `user2`)
	})
}
