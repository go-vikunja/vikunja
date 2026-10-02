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

func TestUserSource(t *testing.T) {
	cases := map[string]string{
		"":                       UserSourceLocal,
		user.IssuerLocal:         UserSourceLocal,
		user.IssuerImport:        UserSourceImport,
		user.IssuerLDAP:          UserSourceLDAP,
		"https://login.microsoftonline.com/t/v2.0": UserSourceEntra,
		"https://accounts.google.com":              UserSourceOther,
	}
	for issuer, want := range cases {
		assert.Equal(t, want, UserSource(&user.User{Issuer: issuer}), issuer)
	}
}

func TestUpdateUserGeneralSettings_NonLocalKeepsName(t *testing.T) {
	t.Run("a non-local user cannot change the name but saves the rest", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		u, err := user.GetUserByID(s, 14) // issuer https://some.service.com
		require.NoError(t, err)
		require.False(t, u.IsLocalUser())
		storedName := u.Name

		settings := NewUserGeneralSettings(u)
		settings.Name = "Typed By The User"
		settings.Language = "de"
		settings.Timezone = "Europe/Berlin"
		require.NoError(t, UpdateUserGeneralSettings(s, u, settings))
		require.NoError(t, s.Commit())

		db.AssertExists(t, "users", map[string]interface{}{"id": 14, "name": storedName, "language": "de", "timezone": "Europe/Berlin"}, false)
	})
	t.Run("a local user can change the name", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		u, err := user.GetUserByID(s, 2)
		require.NoError(t, err)

		settings := NewUserGeneralSettings(u)
		settings.Name = "Typed By The User"
		require.NoError(t, UpdateUserGeneralSettings(s, u, settings))
		require.NoError(t, s.Commit())

		db.AssertExists(t, "users", map[string]interface{}{"id": 2, "name": "Typed By The User"}, false)
	})
	t.Run("job title and department cannot be written through settings", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		_, err := s.ID(2).Cols("job_title", "department").Update(&user.User{JobTitle: "Engineer", Department: "IT"})
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		s.Close()

		s = db.NewSession()
		defer s.Close()
		u, err := user.GetUserByID(s, 2)
		require.NoError(t, err)
		settings := NewUserGeneralSettings(u)
		assert.Equal(t, "Engineer", settings.JobTitle, "visible in the settings")
		assert.Equal(t, "IT", settings.Department)

		settings.JobTitle = "CEO"
		settings.Department = "Board"
		settings.Language = "de"
		require.NoError(t, UpdateUserGeneralSettings(s, u, settings))
		require.NoError(t, s.Commit())

		db.AssertExists(t, "users", map[string]interface{}{"id": 2, "job_title": "Engineer", "department": "IT", "language": "de"}, false)
	})
}

func TestUpdateUserProfileAsAdmin(t *testing.T) {
	doer := &user.User{ID: 1}

	run := func(t *testing.T, id int64, update ManageProfileUpdate) (*user.User, error) {
		t.Helper()
		s := db.NewSession()
		defer s.Close()
		u, err := UpdateUserProfileAsAdmin(s, doer, id, update)
		if err != nil {
			_ = s.Rollback()
			return nil, err
		}
		require.NoError(t, s.Commit())
		return u, nil
	}

	t.Run("writes only name, email and language, and audits the fields", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		events.ClearDispatchedEvents()
		s := db.NewSession()
		before, err := user.GetUserByID(s, 2)
		require.NoError(t, err)
		s.Close()

		_, err = run(t, 2, ManageProfileUpdate{Name: "New Name", Email: "new2@example.com", Language: "de"})
		require.NoError(t, err)

		db.AssertExists(t, "users", map[string]interface{}{
			"id": 2, "name": "New Name", "email": "new2@example.com", "language": "de",
			"timezone": before.Timezone, "week_start": before.WeekStart, "status": int(before.Status),
		}, false)
		evts := events.GetDispatchedEvents((&ManageUserProfileUpdatedEvent{}).Name())
		require.Len(t, evts, 1)
		assert.ElementsMatch(t, []string{"name", "email", "language"}, evts[0].(*ManageUserProfileUpdatedEvent).Fields)
	})
	t.Run("an unchanged profile writes and audits nothing", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		events.ClearDispatchedEvents()
		s := db.NewSession()
		current, err := GetManagedUser(s, 2)
		require.NoError(t, err)
		s.Close()

		_, err = run(t, 2, ManageProfileUpdate{Name: current.Name, Email: current.Email})
		require.NoError(t, err)

		assert.Zero(t, events.CountDispatchedEvents((&ManageUserProfileUpdatedEvent{}).Name()))
	})
	t.Run("an omitted language keeps the current one", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		_, err := s.ID(2).Cols("language").Update(&user.User{Language: "fr"})
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		s.Close()

		_, err = run(t, 2, ManageProfileUpdate{Name: "X", Email: "user2@example.com"})
		require.NoError(t, err)

		db.AssertExists(t, "users", map[string]interface{}{"id": 2, "language": "fr"}, false)
	})
	t.Run("an admin set email replaces an unconfirmed pending one", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		_, err := s.ID(2).Cols("pending_email").Update(&user.User{PendingEmail: "pending@example.com"})
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		s.Close()

		_, err = run(t, 2, ManageProfileUpdate{Name: "X", Email: "admin-set@example.com"})
		require.NoError(t, err)

		db.AssertExists(t, "users", map[string]interface{}{"id": 2, "email": "admin-set@example.com", "pending_email": ""}, false)
	})
	t.Run("refusals", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		_, err := run(t, 14, ManageProfileUpdate{Name: "X", Email: "x@example.com"})
		require.Error(t, err)
		assert.True(t, user.IsErrProfileManagedExternally(err), "third-party user: %v", err)

		_, err = run(t, 23, ManageProfileUpdate{Name: "X", Email: "x@example.com"})
		require.Error(t, err, "bot")

		_, err = run(t, 2, ManageProfileUpdate{Name: "X", Email: "user3@example.com"})
		require.Error(t, err, "duplicate email")

		_, err = run(t, 2, ManageProfileUpdate{Name: "X", Email: ""})
		require.Error(t, err, "empty email")

		_, err = run(t, 2, ManageProfileUpdate{Name: "X", Email: "Jane <j@example.com>"})
		require.Error(t, err, "display-name form")

		_, err = run(t, 2, ManageProfileUpdate{Name: "X", Email: "x@example.com", Language: "xx-nope"})
		require.Error(t, err, "unknown language")

		_, err = run(t, 99999, ManageProfileUpdate{Name: "X", Email: "x@example.com"})
		require.Error(t, err, "unknown user")

		db.AssertExists(t, "users", map[string]interface{}{"id": 2, "email": "user2@example.com"}, false)
	})
}

func TestTransferProjectsAsAdmin(t *testing.T) {
	doer := &user.User{ID: 1}

	transfer := func(t *testing.T, from, to int64, ids []int64) (int, error) {
		t.Helper()
		s := db.NewSession()
		defer s.Close()
		n, err := TransferProjectsAsAdmin(s, doer, from, to, ids)
		if err != nil {
			_ = s.Rollback()
			events.CleanupPending(s)
			return 0, err
		}
		require.NoError(t, s.Commit())
		events.DispatchPending(context.Background(), s)
		return n, nil
	}
	owned := func(t *testing.T, id int64) []*TransferableProject {
		t.Helper()
		s := db.NewSession()
		defer s.Close()
		list, err := ListOwnedProjects(s, id)
		require.NoError(t, err)
		return list
	}

	t.Run("hands over every owned project and audits each", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		events.ClearDispatchedEvents()
		before := owned(t, 3)
		require.NotEmpty(t, before)
		destBefore := len(owned(t, 2))

		n, err := transfer(t, 3, 2, nil)

		require.NoError(t, err)
		assert.Equal(t, len(before), n)
		assert.Empty(t, owned(t, 3))
		assert.Len(t, owned(t, 2), destBefore+len(before))
		assert.Equal(t, len(before), events.CountDispatchedEvents((&AdminProjectOwnerChangedEvent{}).Name()))
	})
	t.Run("only the selected projects move", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		before := owned(t, 3)
		require.GreaterOrEqual(t, len(before), 2, "fixture user3 owns several projects")

		n, err := transfer(t, 3, 2, []int64{before[0].ID})

		require.NoError(t, err)
		assert.Equal(t, 1, n)
		assert.Len(t, owned(t, 3), len(before)-1)
	})
	t.Run("a duplicate id in the selection is counted once", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		before := owned(t, 3)

		n, err := transfer(t, 3, 2, []int64{before[0].ID, before[0].ID})

		require.NoError(t, err)
		assert.Equal(t, 1, n)
	})
	t.Run("an empty selection moves nothing", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		before := owned(t, 3)

		n, err := transfer(t, 3, 2, []int64{})

		require.NoError(t, err)
		assert.Zero(t, n)
		assert.Len(t, owned(t, 3), len(before))
	})
	t.Run("one project the source does not own cancels the whole transfer", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		before := owned(t, 3)
		foreign := owned(t, 1)
		require.NotEmpty(t, foreign)

		_, err := transfer(t, 3, 2, []int64{before[0].ID, foreign[0].ID})

		require.Error(t, err)
		assert.Len(t, owned(t, 3), len(before), "the valid project did not move either")
	})
	t.Run("refused new owners", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		before := len(owned(t, 3))

		for name, to := range map[string]int64{
			"same user":      3,
			"disabled":      17,
			"locked":        18,
			"bot":           23,
			"does not exist": 99999,
		} {
			_, err := transfer(t, 3, to, nil)
			require.Error(t, err, name)
		}
		assert.Len(t, owned(t, 3), before)
	})
	t.Run("an unknown source user is an error", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		_, err := transfer(t, 99999, 2, nil)
		require.Error(t, err)
	})
}

func TestListManagedUsers(t *testing.T) {
	list := func(t *testing.T, f ManageUserFilter, page, perPage int) ([]*user.User, int64) {
		t.Helper()
		s := db.NewSession()
		defer s.Close()
		users, total, err := ListManagedUsers(s, &user.User{ID: 1}, f, page, perPage)
		require.NoError(t, err)
		return users, total
	}
	usernames := func(users []*user.User) []string {
		out := make([]string, 0, len(users))
		for _, u := range users {
			out = append(out, u.Username)
		}
		return out
	}

	t.Run("bots are never listed", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		users, _ := list(t, ManageUserFilter{Status: -1}, 1, 100)
		for _, u := range users {
			assert.Zero(t, u.BotOwnerID)
		}
	})
	t.Run("search matches username, name and email", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		byUsername, _ := list(t, ManageUserFilter{Search: "user2", Status: -1}, 1, 100)
		assert.Contains(t, usernames(byUsername), "user2")
		byEmail, _ := list(t, ManageUserFilter{Search: "user3@example", Status: -1}, 1, 100)
		assert.Equal(t, []string{"user3"}, usernames(byEmail))
	})
	t.Run("source and status filters", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		other, _ := list(t, ManageUserFilter{Source: UserSourceOther, Status: -1}, 1, 100)
		assert.ElementsMatch(t, []string{"user14", "user_openid_avatar"}, usernames(other))

		disabled, _ := list(t, ManageUserFilter{Status: int(user.StatusDisabled)}, 1, 100)
		assert.Contains(t, usernames(disabled), "user17")

		imported, total := list(t, ManageUserFilter{Source: UserSourceImport, Status: -1}, 1, 100)
		assert.Empty(t, imported)
		assert.Zero(t, total)
	})
	t.Run("an unknown source filter is ignored, not an error", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		_, total := list(t, ManageUserFilter{Source: "nonsense", Status: -1}, 1, 100)
		assert.Positive(t, total)
	})
	t.Run("paging and a total that ignores paging", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		first, total := list(t, ManageUserFilter{Status: -1}, 1, 3)
		second, _ := list(t, ManageUserFilter{Status: -1}, 2, 3)
		assert.Len(t, first, 3)
		assert.Greater(t, total, int64(3))
		assert.NotEqual(t, usernames(first), usernames(second))
	})
}
