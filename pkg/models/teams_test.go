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
	"reflect"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTeam_Create(t *testing.T) {
	doer := &user.User{
		ID:       1,
		Username: "user1",
	}
	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		team := &Team{
			Name:        "Testteam293",
			Description: "Lorem Ispum",
		}
		err := team.Create(s, doer)
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)
		db.AssertExists(t, "teams", map[string]interface{}{
			"id":          team.ID,
			"name":        "Testteam293",
			"description": "Lorem Ispum",
			"is_public":   false,
		}, false)
	})
	t.Run("empty name", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		team := &Team{}
		err := team.Create(s, doer)
		require.Error(t, err)
		assert.True(t, IsErrTeamNameCannotBeEmpty(err))
	})
	t.Run("public", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		team := &Team{
			Name:        "Testteam293_Public",
			Description: "Lorem Ispum",
			IsPublic:    true,
		}
		err := team.Create(s, doer)
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)
		db.AssertExists(t, "teams", map[string]interface{}{
			"id":          team.ID,
			"name":        "Testteam293_Public",
			"description": "Lorem Ispum",
			"is_public":   true,
		}, false)
	})
}

func TestTeam_ReadOne(t *testing.T) {
	u := &user.User{ID: 1}

	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		team := &Team{ID: 1}
		err := team.ReadOne(s, u)
		require.NoError(t, err)
		assert.Equal(t, "testteam1", team.Name)
		assert.Equal(t, "Lorem Ipsum", team.Description)
		assert.Equal(t, int64(1), team.CreatedBy.ID)
		assert.Equal(t, int64(1), team.CreatedByID)
	})
	t.Run("invalid id", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		team := &Team{ID: -1}
		err := team.ReadOne(s, u)
		require.Error(t, err)
		assert.True(t, IsErrTeamDoesNotExist(err))
	})
	t.Run("nonexisting", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		team := &Team{ID: 99999}
		err := team.ReadOne(s, u)
		require.Error(t, err)
		assert.True(t, IsErrTeamDoesNotExist(err))
	})
}

func TestTeam_ReadAll(t *testing.T) {
	doer := &user.User{ID: 1}
	t.Run("normal", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		team := &Team{}
		teams, _, _, err := team.ReadAll(s, doer, "", 1, 50)
		require.NoError(t, err)
		assert.Equal(t, reflect.Slice, reflect.TypeOf(teams).Kind())
		ts := reflect.ValueOf(teams)
		assert.Equal(t, 5, ts.Len())
	})
	t.Run("search", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		team := &Team{}
		teams, _, _, err := team.ReadAll(s, doer, "READ_only_on_project6", 1, 50)
		require.NoError(t, err)
		assert.Equal(t, reflect.Slice, reflect.TypeOf(teams).Kind())
		ts := teams.([]*Team)
		assert.Len(t, ts, 1)
		assert.Equal(t, int64(2), ts[0].ID)
	})
	t.Run("public discovery disabled", func(t *testing.T) {

		s := db.NewSession()
		defer s.Close()

		team := &Team{}

		// Default setting is having ServiceEnablePublicTeams disabled
		// In this default case, fetching teams with or without public flag should return the same result

		// Fetch without public flag
		teams, _, _, err := team.ReadAll(s, doer, "", 1, 50)
		require.NoError(t, err)
		assert.Equal(t, reflect.Slice, reflect.TypeOf(teams).Kind())
		ts := teams.([]*Team)
		assert.Len(t, ts, 5)

		// Fetch with public flag
		team.IncludePublic = true
		teams, _, _, err = team.ReadAll(s, doer, "", 1, 50)
		require.NoError(t, err)
		assert.Equal(t, reflect.Slice, reflect.TypeOf(teams).Kind())
		ts = teams.([]*Team)
		assert.Len(t, ts, 5)
	})

	t.Run("public discovery enabled", func(t *testing.T) {

		s := db.NewSession()
		defer s.Close()

		team := &Team{}

		// Enable ServiceEnablePublicTeams feature
		config.ServiceEnablePublicTeams.Set(true)

		// Fetch without public flag should be the same as before
		team.IncludePublic = false
		teams, _, _, err := team.ReadAll(s, doer, "", 1, 50)
		require.NoError(t, err)
		assert.Equal(t, reflect.Slice, reflect.TypeOf(teams).Kind())
		ts := teams.([]*Team)
		assert.Len(t, ts, 5)

		// Fetch with public flag should return more teams
		team.IncludePublic = true
		teams, _, _, err = team.ReadAll(s, doer, "", 1, 50)
		require.NoError(t, err)
		assert.Equal(t, reflect.Slice, reflect.TypeOf(teams).Kind())
		ts = teams.([]*Team)
		assert.Len(t, ts, 7)
	})
}

func TestTeam_Update(t *testing.T) {
	u := &user.User{ID: 1}

	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		team := &Team{
			ID:   1,
			Name: "SomethingNew",
		}
		err := team.Update(s, u)
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)
		db.AssertExists(t, "teams", map[string]interface{}{
			"id":   team.ID,
			"name": "SomethingNew",
		}, false)
	})
	t.Run("empty name", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		team := &Team{
			ID:   1,
			Name: "",
		}
		err := team.Update(s, u)
		require.Error(t, err)
		assert.True(t, IsErrTeamNameCannotBeEmpty(err))
	})
	t.Run("nonexisting", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		team := &Team{
			ID:   9999,
			Name: "SomethingNew",
		}
		err := team.Update(s, u)
		require.Error(t, err)
		assert.True(t, IsErrTeamDoesNotExist(err))
	})
}

func TestTeam_Delete(t *testing.T) {
	u := &user.User{ID: 1}

	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		team := &Team{
			ID: 1,
		}
		err := team.Delete(s, u)
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)
		db.AssertMissing(t, "teams", map[string]interface{}{
			"id": 1,
		})
	})
}

func TestIsErrInvalidPermission(t *testing.T) {
	require.NoError(t, PermissionAdmin.isValid())
	require.NoError(t, PermissionRead.isValid())
	require.NoError(t, PermissionWrite.isValid())

	// Check invalid
	var tr Permission = 938
	err := tr.isValid()
	require.Error(t, err)
	assert.True(t, IsErrInvalidPermission(err))
}
