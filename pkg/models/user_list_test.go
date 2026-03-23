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
		assert.Len(t, all, 18)
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
		assert.Len(t, all, 18)
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
}
