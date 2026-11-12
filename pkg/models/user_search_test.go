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
	"fmt"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

func withConfig(t *testing.T, key config.Key, value bool) {
	t.Helper()
	prev := key.GetBool()
	key.Set(value)
	t.Cleanup(func() { key.Set(prev) })
}

func searchUserIDs(users []*user.User) []int64 {
	ids := make([]int64, 0, len(users))
	for _, u := range users {
		ids = append(ids, u.ID)
	}
	return ids
}

func insertSearchUser(t *testing.T, s *xorm.Session, u *user.User) *user.User {
	t.Helper()
	_, err := s.Insert(u)
	require.NoError(t, err)
	return u
}

func TestSearchUsers_FlagOff(t *testing.T) {
	user1 := &user.User{ID: 1}
	withConfig(t, config.ServiceEnableGlobalUserSearch, false)

	t.Run("partial username does not match", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		users, total, err := user.SearchUsers(s, "user", user1, 1)
		require.NoError(t, err)
		assert.Empty(t, users)
		assert.Zero(t, total)
	})
	t.Run("partial name needs discoverable_by_name", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		users, _, err := user.SearchUsers(s, "one else", user1, 1)
		require.NoError(t, err)
		assert.Empty(t, users)

		users, total, err := user.SearchUsers(s, "with space", user1, 1)
		require.NoError(t, err)
		assert.Equal(t, []int64{12}, searchUserIDs(users))
		assert.EqualValues(t, 1, total)
	})
	t.Run("email needs discoverable_by_email and is not returned", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		users, _, err := user.SearchUsers(s, "user1@example.com", user1, 1)
		require.NoError(t, err)
		assert.Empty(t, users)

		users, _, err = user.SearchUsers(s, "user7@example.com", user1, 1)
		require.NoError(t, err)
		require.Equal(t, []int64{7}, searchUserIDs(users))
		assert.Empty(t, users[0].Email)
	})
	t.Run("exact username, email hidden", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		users, _, err := user.SearchUsers(s, "user7", user1, 1)
		require.NoError(t, err)
		require.Equal(t, []int64{7}, searchUserIDs(users))
		assert.Empty(t, users[0].Email)
	})
	t.Run("results are capped", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		for i := range 30 {
			insertSearchUser(t, s, &user.User{
				Username:           fmt.Sprintf("capoff-%02d", i),
				Name:               fmt.Sprintf("Capoff Person %02d", i),
				Email:              fmt.Sprintf("capoff-%02d@example.com", i),
				DiscoverableByName: true,
			})
		}

		users, total, err := user.SearchUsers(s, "Capoff Person", user1, 1)
		require.NoError(t, err)
		assert.Len(t, users, user.SearchUsersPageSize)
		assert.EqualValues(t, 30, total)
		assert.Equal(t, "capoff-00", users[0].Username)

		users, _, err = user.SearchUsers(s, "Capoff Person", user1, 2)
		require.NoError(t, err)
		assert.Len(t, users, 5)
	})
}

func TestSearchUsers_FlagOn(t *testing.T) {
	user1 := &user.User{ID: 1}
	withConfig(t, config.ServiceEnableGlobalUserSearch, true)

	t.Run("partial name ignores discoverable_by_name and returns email", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		users, total, err := user.SearchUsers(s, "one else", user1, 1)
		require.NoError(t, err)
		require.Equal(t, []int64{11}, searchUserIDs(users))
		assert.EqualValues(t, 1, total)
		assert.Equal(t, "user11@example.com", users[0].Email)
	})
	t.Run("partial username", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		users, _, err := user.SearchUsers(s, "SER1", user1, 1)
		require.NoError(t, err)
		ids := searchUserIDs(users)
		assert.Contains(t, ids, int64(1))
		assert.Contains(t, ids, int64(18))
		for _, u := range users {
			assert.NotEmpty(t, u.Email)
		}
	})
	t.Run("partial email ignores discoverable_by_email", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		users, _, err := user.SearchUsers(s, "some.service", user1, 1)
		require.NoError(t, err)
		require.Equal(t, []int64{14}, searchUserIDs(users))
		assert.Equal(t, "user15@some.service.com", users[0].Email)
	})
	t.Run("fewer than 2 characters return nothing", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		for _, search := range []string{"u", " u ", "u,1", "ü"} {
			users, total, err := user.SearchUsers(s, search, user1, 1)
			require.NoError(t, err)
			assert.Empty(t, users, "search %q", search)
			assert.Zero(t, total, "search %q", search)
		}
	})
	t.Run("ranking, cap and total", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Substring matches sort alphabetically before the ranked ones, so
		// only the ranking can put those first.
		for i := range 30 {
			insertSearchUser(t, s, &user.User{
				Username: fmt.Sprintf("a-rankq-%02d", i),
				Email:    fmt.Sprintf("a-rankq-%02d@example.com", i),
			})
		}
		usernamePrefix := insertSearchUser(t, s, &user.User{
			Username: "rankq-prefix",
			Email:    "rankq-prefix@example.com",
		})
		namePrefix := insertSearchUser(t, s, &user.User{
			Username: "b-name",
			Name:     "Rankq Name",
			Email:    "b-name@example.com",
		})
		exact := insertSearchUser(t, s, &user.User{
			Username: "rankq",
			Email:    "z-exact@example.com",
		})

		users, total, err := user.SearchUsers(s, "RANKQ", user1, 1)
		require.NoError(t, err)
		assert.EqualValues(t, 33, total)
		require.Len(t, users, user.SearchUsersPageSize)
		assert.Equal(t, exact.ID, users[0].ID)
		assert.Equal(t, namePrefix.ID, users[1].ID)
		assert.Equal(t, usernamePrefix.ID, users[2].ID)
		assert.Equal(t, "a-rankq-00", users[3].Username)
		assert.Equal(t, "a-rankq-01", users[4].Username)

		users, total, err = user.SearchUsers(s, "RANKQ", user1, 2)
		require.NoError(t, err)
		assert.EqualValues(t, 33, total)
		assert.Len(t, users, 8)
	})
	t.Run("other users' bots stay hidden", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		users, _, err := user.SearchUsers(s, "assistant", &user.User{ID: 21}, 1)
		require.NoError(t, err)
		assert.Equal(t, []int64{23}, searchUserIDs(users))

		users, _, err = user.SearchUsers(s, "owner", user1, 1)
		require.NoError(t, err)
		ids := searchUserIDs(users)
		assert.ElementsMatch(t, []int64{21, 22}, ids)
	})
	t.Run("team-only search still restricts results", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		withConfig(t, config.ServiceEnableOpenIDTeamUserOnlySearch, true)

		// user1 shares no team with user11, user10 does (team 14).
		users, _, err := user.SearchUsers(s, "one else", user1, 1)
		require.NoError(t, err)
		assert.Empty(t, users)

		users, _, err = user.SearchUsers(s, "one else", &user.User{ID: 10}, 1)
		require.NoError(t, err)
		assert.Equal(t, []int64{11}, searchUserIDs(users))

		// The flag-off exact-email bypass of the team restriction does not apply.
		users, _, err = user.SearchUsers(s, "user7@example.com", user1, 1)
		require.NoError(t, err)
		assert.Empty(t, users)
	})
}

func TestSearchUsers_ProjectSearchUnaffectedByGlobalFlag(t *testing.T) {
	user1 := &user.User{ID: 1}

	search := func(t *testing.T, enabled bool, query string) []*user.User {
		withConfig(t, config.ServiceEnableGlobalUserSearch, enabled)
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		users, err := ListUsersFromProject(s, &Project{ID: 3}, user1, query)
		require.NoError(t, err)
		return users
	}

	for _, query := range []string{"", "u", "user"} {
		t.Run(fmt.Sprintf("query %q", query), func(t *testing.T) {
			off := search(t, false, query)
			on := search(t, true, query)
			assert.NotEmpty(t, off)
			assert.Equal(t, searchUserIDs(off), searchUserIDs(on))
			for i := range off {
				assert.Equal(t, off[i].Email, on[i].Email)
			}
		})
	}
}
