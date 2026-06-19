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
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/builder"
)

func TestListUsers(t *testing.T) {
	user1 := &user.User{ID: 1}
	user10 := &user.User{ID: 10}

	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		all, err := user.ListUsers(s, "user1", user1, nil)
		require.NoError(t, err)
		assert.NotEmpty(t, all)
		assert.Equal(t, "user1", all[0].Username)
	})
	t.Run("case insensitive", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		all, err := user.ListUsers(s, "uSEr1", user1, nil)
		require.NoError(t, err)
		assert.NotEmpty(t, all)
		assert.Equal(t, "user1", all[0].Username)
	})
	t.Run("all users", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		all, err := user.ListAllUsers(s)
		require.NoError(t, err)
		assert.Len(t, all, 24)
	})
	t.Run("no search term", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		all, err := user.ListUsers(s, "", user1, nil)
		require.NoError(t, err)
		assert.Empty(t, all)
	})
	t.Run("not discoverable by email", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		all, err := user.ListUsers(s, "user1@example.com", user1, nil)
		require.NoError(t, err)
		assert.Empty(t, all)
		db.AssertExists(t, "users", map[string]interface{}{
			"email":                 "user1@example.com",
			"discoverable_by_email": false,
		}, false)
	})
	t.Run("not discoverable by name", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		all, err := user.ListUsers(s, "one else", user1, nil)
		require.NoError(t, err)
		assert.Empty(t, all)
		db.AssertExists(t, "users", map[string]interface{}{
			"name":                 "Some one else",
			"discoverable_by_name": false,
		}, false)
	})
	t.Run("discoverable by email", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		all, err := user.ListUsers(s, "user7@example.com", user1, nil)
		require.NoError(t, err)
		assert.Len(t, all, 1)
		assert.Equal(t, int64(7), all[0].ID)
		db.AssertExists(t, "users", map[string]interface{}{
			"email":                 "user7@example.com",
			"discoverable_by_email": true,
		}, false)
	})
	t.Run("discoverable by partial name", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		all, err := user.ListUsers(s, "with space", user1, nil)
		require.NoError(t, err)
		assert.Len(t, all, 1)
		assert.Equal(t, int64(12), all[0].ID)
		db.AssertExists(t, "users", map[string]interface{}{
			"name":                 "Name with spaces",
			"discoverable_by_name": true,
		}, false)
	})
	t.Run("discoverable by email with extra condition", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		all, err := user.ListUsers(s, "user7@example.com", user1, &user.ProjectUserOpts{AdditionalCond: builder.In("id", 7)})
		require.NoError(t, err)
		assert.Len(t, all, 1)
		assert.Equal(t, int64(7), all[0].ID)
		db.AssertExists(t, "users", map[string]interface{}{
			"email":                 "user7@example.com",
			"discoverable_by_email": true,
		}, false)
	})
	t.Run("discoverable by exact username", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		all, err := user.ListUsers(s, "user7", user1, nil)
		require.NoError(t, err)
		assert.Len(t, all, 1)
		assert.Equal(t, int64(7), all[0].ID)
		db.AssertExists(t, "users", map[string]interface{}{
			"username": "user7",
		}, false)
	})
	t.Run("not discoverable by partial username", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		all, err := user.ListUsers(s, "user", user1, nil)
		require.NoError(t, err)
		assert.Empty(t, all)
		db.AssertExists(t, "users", map[string]interface{}{
			"username": "user7",
		}, false)
	})
	t.Run("discoverable by partial username, email and name when matching fuzzily", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		all, err := user.ListUsers(s, "user", user1, &user.ProjectUserOpts{
			MatchFuzzily: true,
		})
		require.NoError(t, err)
		// 22 non-bot users have "user" in their username; the two bot
		// fixtures are filtered out because they don't belong to user1
		// and their usernames/names don't contain "user".
		assert.Len(t, all, 22)
	})

	// External team discoverability bypass tests
	// User 10 and user 11 share external team 14 (has external_id).
	// User 11 has discoverable_by_name=false and discoverable_by_email=false.
	t.Run("external team member discoverable by name", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		all, err := user.ListUsers(s, "Some one else", user10, nil)
		require.NoError(t, err)
		assert.Len(t, all, 1)
		assert.Equal(t, int64(11), all[0].ID)
	})
	t.Run("external team member discoverable by email", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		all, err := user.ListUsers(s, "user11@example.com", user10, nil)
		require.NoError(t, err)
		assert.Len(t, all, 1)
		assert.Equal(t, int64(11), all[0].ID)
	})
	t.Run("non-external-team user cannot discover by name", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// User 1 does NOT share an external team with user 11.
		all, err := user.ListUsers(s, "Some one else", user1, nil)
		require.NoError(t, err)
		assert.Empty(t, all)
	})
	t.Run("non-external-team user cannot discover by email", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// User 1 does NOT share an external team with user 11.
		all, err := user.ListUsers(s, "user11@example.com", user1, nil)
		require.NoError(t, err)
		assert.Empty(t, all)
	})
	t.Run("regular team does not bypass discoverability", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// User 1 and user 2 share team 1 (regular team, no external_id).
		// User 2 has discoverable_by_email=false.
		all, err := user.ListUsers(s, "user2@example.com", user1, nil)
		require.NoError(t, err)
		assert.Empty(t, all)
	})
	t.Run("external team member email masked when searching by name", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		all, err := user.ListUsers(s, "Some one else", user10, nil)
		require.NoError(t, err)
		require.Len(t, all, 1)
		// Email should be masked because the search was by name, not email
		assert.Empty(t, all[0].Email)
	})

	// Bot visibility in user search:
	// - A user's own bots are filtered by the search string (matched against
	//   username and name), but bypass the discoverable_by_name flag that
	//   hides regular users unless they opt in.
	// - When no search string is provided and ReturnAllIfNoSearchProvided is
	//   set, own bots are returned alongside regular users.
	// - Other users' bots must never leak into the results, even on an exact
	//   username match.
	t.Run("own bot NOT returned when query does not match it", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		botOwnerA := &user.User{ID: 21}
		// Query string deliberately does not match the bot's username or name.
		all, err := user.ListUsers(s, "user7", botOwnerA, nil)
		require.NoError(t, err)

		for _, u := range all {
			assert.NotEqual(t, int64(23), u.ID, "owner A's bot must not appear when the query does not match it")
		}
	})
	t.Run("other user's bot not returned by exact username match", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		botOwnerA := &user.User{ID: 21}
		// Searching for owner B's bot by its exact username must not return it.
		all, err := user.ListUsers(s, "bot-owner-b-assistant", botOwnerA, nil)
		require.NoError(t, err)

		for _, u := range all {
			assert.NotEqual(t, int64(24), u.ID, "owner B's bot must not leak into owner A's results")
		}
	})
	t.Run("neither own nor other bot returned when query matches only other owner's bot", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		botOwnerB := &user.User{ID: 22}
		all, err := user.ListUsers(s, "bot-owner-a-assistant", botOwnerB, nil)
		require.NoError(t, err)

		// Owner A's bot must not leak. Owner B's bot must not appear either since
		// the query does not match its username or name.
		for _, u := range all {
			assert.NotEqual(t, int64(23), u.ID, "owner A's bot must not leak to owner B")
			assert.NotEqual(t, int64(24), u.ID, "owner B's bot must not appear when the query does not match it")
		}
	})
	t.Run("own bot returned by username match", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		botOwnerA := &user.User{ID: 21}
		all, err := user.ListUsers(s, "bot-owner-a-assistant", botOwnerA, nil)
		require.NoError(t, err)

		var foundBot bool
		for _, u := range all {
			if u.ID == 23 {
				foundBot = true
				assert.Equal(t, int64(21), u.BotOwnerID)
			}
		}
		assert.True(t, foundBot, "owner A's bot (id=23) should appear when searching by exact username")
	})
	t.Run("own bot returned by name match without discoverable_by_name", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		botOwnerA := &user.User{ID: 21}
		// The bot fixture does not set discoverable_by_name=true, but own bots
		// should still match by name for their owner.
		all, err := user.ListUsers(s, "Owner A Assistant", botOwnerA, nil)
		require.NoError(t, err)

		var foundBot bool
		for _, u := range all {
			if u.ID == 23 {
				foundBot = true
				assert.Equal(t, int64(21), u.BotOwnerID)
			}
		}
		assert.True(t, foundBot, "owner A's bot (id=23) should appear when searching by name even without discoverable_by_name")
	})
	t.Run("own bot returned when no search but ReturnAllIfNoSearchProvided", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		botOwnerA := &user.User{ID: 21}
		all, err := user.ListUsers(s, "", botOwnerA, &user.ProjectUserOpts{ReturnAllIfNoSearchProvided: true})
		require.NoError(t, err)

		var foundBot bool
		for _, u := range all {
			if u.ID == 23 {
				foundBot = true
				assert.Equal(t, int64(21), u.BotOwnerID)
			}
		}
		assert.True(t, foundBot, "owner A's bot (id=23) should appear in results when no search is provided and ReturnAllIfNoSearchProvided is true")
	})
}
