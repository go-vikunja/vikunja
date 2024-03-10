// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package openid

import (
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetOrCreateUser(t *testing.T) {
	t.Run("new user", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		cl := &claims{
			Email:             "test@example.com",
			PreferredUsername: "someUserWhoDoesNotExistYet",
		}
		u, err := getOrCreateUser(s, cl, "https://some.issuer", "12345")
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)

		db.AssertExists(t, "users", map[string]interface{}{
			"id":       u.ID,
			"email":    cl.Email,
			"username": "someUserWhoDoesNotExistYet",
		}, false)
	})
	t.Run("new user, no username provided", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		cl := &claims{
			Email:             "test@example.com",
			PreferredUsername: "",
		}
		u, err := getOrCreateUser(s, cl, "https://some.issuer", "12345")
		require.NoError(t, err)
		assert.NotEmpty(t, u.Username)
		err = s.Commit()
		require.NoError(t, err)

		db.AssertExists(t, "users", map[string]interface{}{
			"id":    u.ID,
			"email": cl.Email,
		}, false)
	})
	t.Run("new user, no email address", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		cl := &claims{
			Email: "",
		}
		_, err := getOrCreateUser(s, cl, "https://some.issuer", "12345")
		require.Error(t, err)
	})
	t.Run("existing user, different email address", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		cl := &claims{
			Email: "other-email-address@some.service.com",
		}
		u, err := getOrCreateUser(s, cl, "https://some.service.com", "12345")
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)

		db.AssertExists(t, "users", map[string]interface{}{
			"id":    u.ID,
			"email": cl.Email,
		}, false)
	})
	t.Run("existing user, non existing team", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		team := "new sso team"
		oidcID := "47404"
		cl := &claims{
			Email: "other-email-address@some.service.com",
			VikunjaGroups: []map[string]interface{}{
				{"name": team, "oidcID": oidcID},
			},
		}

		u, err := getOrCreateUser(s, cl, "https://some.service.com", "12345")
		require.NoError(t, err)
		teamData, errs := getTeamDataFromToken(cl.VikunjaGroups, nil)
		for _, err := range errs {
			require.NoError(t, err)
		}
		require.NoError(t, err)
		oidcTeams, err := AssignOrCreateUserToTeams(s, u, teamData, "https://some.issuer")
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)

		db.AssertExists(t, "users", map[string]interface{}{
			"id":    u.ID,
			"email": cl.Email,
		}, false)
		db.AssertExists(t, "teams", map[string]interface{}{
			"id":   oidcTeams,
			"name": team + " (OIDC)",
		}, false)
	})

	t.Run("existing user, assign to existing team", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		team := "testteam14"
		oidcID := "14"
		cl := &claims{
			Email: "other-email-address@some.service.com",
			VikunjaGroups: []map[string]interface{}{
				{"name": team, "oidcID": oidcID},
			},
		}

		u := &user.User{ID: 10}
		teamData, errs := getTeamDataFromToken(cl.VikunjaGroups, nil)
		for _, err := range errs {
			require.NoError(t, err)
		}
		oidcTeams, err := AssignOrCreateUserToTeams(s, u, teamData, "https://some.issuer")
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)

		db.AssertExists(t, "team_members", map[string]interface{}{
			"team_id": oidcTeams,
			"user_id": u.ID,
		}, false)
	})
	t.Run("existing user, remove from existing team", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		cl := &claims{
			Email:         "other-email-address@some.service.com",
			VikunjaGroups: []map[string]interface{}{},
		}

		u := &user.User{ID: 10}
		teamData, errs := getTeamDataFromToken(cl.VikunjaGroups, nil)
		if len(errs) > 0 {
			for _, err := range errs {
				require.NoError(t, err)
			}
		}
		oldOidcTeams, err := models.FindAllOidcTeamIDsForUser(s, u.ID)
		require.NoError(t, err)
		oidcTeams, err := AssignOrCreateUserToTeams(s, u, teamData, "https://some.issuer")
		require.NoError(t, err)
		teamIDsToLeave := utils.NotIn(oldOidcTeams, oidcTeams)
		require.NoError(t, err)
		err = RemoveUserFromTeamsByIDs(s, u, teamIDsToLeave)
		require.NoError(t, err)
		errs = RemoveEmptySSOTeams(s, teamIDsToLeave)
		for _, err = range errs {
			require.NoError(t, err)
		}
		errs = RemoveEmptySSOTeams(s, teamIDsToLeave)
		for _, err = range errs {
			require.NoError(t, err)
		}
		err = s.Commit()
		require.NoError(t, err)

		db.AssertMissing(t, "team_members", map[string]interface{}{
			"team_id": oidcTeams,
			"user_id": u.ID,
		})
	})
	t.Run("existing user, remove from existing team and delete team", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		cl := &claims{
			Email:         "other-email-address@some.service.com",
			VikunjaGroups: []map[string]interface{}{},
		}

		u := &user.User{ID: 10}
		teamData, errs := getTeamDataFromToken(cl.VikunjaGroups, nil)
		if len(errs) > 0 {
			for _, err := range errs {
				require.NoError(t, err)
			}
		}
		oldOidcTeams, err := models.FindAllOidcTeamIDsForUser(s, u.ID)
		require.NoError(t, err)
		oidcTeams, err := AssignOrCreateUserToTeams(s, u, teamData, "https://some.issuer")
		require.NoError(t, err)
		teamIDsToLeave := utils.NotIn(oldOidcTeams, oidcTeams)
		require.NoError(t, err)
		err = RemoveUserFromTeamsByIDs(s, u, teamIDsToLeave)
		require.NoError(t, err)
		errs = RemoveEmptySSOTeams(s, teamIDsToLeave)
		for _, err := range errs {
			require.NoError(t, err)
		}
		err = s.Commit()
		require.NoError(t, err)
		db.AssertMissing(t, "teams", map[string]interface{}{
			"id": oidcTeams,
		})
	})
}
