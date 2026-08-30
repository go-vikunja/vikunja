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

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

type teamVisibilityEnv struct {
	s          *xorm.Session
	owner      *user.User
	attacker   *user.User
	proj       *Project
	team       *Team
	publicTeam *Team
}

func setupTeamVisibilityEnv(t *testing.T) *teamVisibilityEnv {
	t.Helper()
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()

	owner := &user.User{ID: 1}
	attacker := &user.User{ID: 2}

	proj := &Project{Title: "team-visibility project"}
	require.NoError(t, proj.Create(s, attacker))

	team := &Team{Name: "team-visibility private team", CreatedByID: owner.ID}
	_, err := s.Insert(team)
	require.NoError(t, err)
	_, err = s.Insert(&TeamMember{TeamID: team.ID, UserID: owner.ID})
	require.NoError(t, err)

	publicTeam := &Team{Name: "team-visibility public team", CreatedByID: owner.ID, IsPublic: true}
	_, err = s.Insert(publicTeam)
	require.NoError(t, err)
	_, err = s.Insert(&TeamMember{TeamID: publicTeam.ID, UserID: owner.ID})
	require.NoError(t, err)
	require.NoError(t, s.Commit())

	return &teamVisibilityEnv{
		s:          s,
		owner:      owner,
		attacker:   attacker,
		proj:       proj,
		team:       team,
		publicTeam: publicTeam,
	}
}

// Guards GHSA-39p5-2wrr-xh29 against attaching unreadable teams.
func TestTeamProjectCanCreateRequiresTeamRead(t *testing.T) {
	t.Run("missing team returns not found", func(t *testing.T) {
		env := setupTeamVisibilityEnv(t)
		defer env.s.Close()

		tl := &TeamProject{TeamID: 999999, ProjectID: env.proj.ID, Permission: PermissionRead}
		can, err := tl.CanCreate(env.s, env.attacker)
		require.Error(t, err)
		require.False(t, can)
		assert.True(t, IsErrTeamDoesNotExist(err))
	})

	t.Run("foreign private team denied", func(t *testing.T) {
		env := setupTeamVisibilityEnv(t)
		defer env.s.Close()

		tl := &TeamProject{TeamID: env.team.ID, ProjectID: env.proj.ID, Permission: PermissionRead}
		can, err := tl.CanCreate(env.s, env.attacker)
		require.NoError(t, err)
		require.False(t, can, "attaching a team the caller is not a member of must be denied")
	})

	t.Run("own team allowed", func(t *testing.T) {
		env := setupTeamVisibilityEnv(t)
		defer env.s.Close()

		ownTeam := &Team{Name: "team-visibility attacker team", CreatedByID: env.attacker.ID}
		_, err := env.s.Insert(ownTeam)
		require.NoError(t, err)
		_, err = env.s.Insert(&TeamMember{TeamID: ownTeam.ID, UserID: env.attacker.ID})
		require.NoError(t, err)
		require.NoError(t, env.s.Commit())

		tl := &TeamProject{TeamID: ownTeam.ID, ProjectID: env.proj.ID, Permission: PermissionRead}
		can, err := tl.CanCreate(env.s, env.attacker)
		require.NoError(t, err)
		require.True(t, can)
	})

	t.Run("public team denied when public teams are disabled", func(t *testing.T) {
		env := setupTeamVisibilityEnv(t)
		defer env.s.Close()

		config.ServiceEnablePublicTeams.Set("false")
		t.Cleanup(func() { config.ServiceEnablePublicTeams.Set("false") })

		tl := &TeamProject{TeamID: env.publicTeam.ID, ProjectID: env.proj.ID, Permission: PermissionRead}
		can, err := tl.CanCreate(env.s, env.attacker)
		require.NoError(t, err)
		require.False(t, can)
	})

	t.Run("public team allowed when enabled", func(t *testing.T) {
		env := setupTeamVisibilityEnv(t)
		defer env.s.Close()

		config.ServiceEnablePublicTeams.Set("true")
		t.Cleanup(func() { config.ServiceEnablePublicTeams.Set("false") })

		tl := &TeamProject{TeamID: env.publicTeam.ID, ProjectID: env.proj.ID, Permission: PermissionRead}
		can, err := tl.CanCreate(env.s, env.attacker)
		require.NoError(t, err)
		require.True(t, can, "a public team may be attached when the instance enables public teams")
	})
}

func TestTeamProjectReadAllScrubsForeignTeams(t *testing.T) {
	env := setupTeamVisibilityEnv(t)
	defer env.s.Close()

	// Simulate a pre-fix foreign-team relation.
	_, err := env.s.Insert(&TeamProject{
		TeamID:     env.team.ID,
		ProjectID:  env.proj.ID,
		Permission: PermissionRead,
	})
	require.NoError(t, err)
	require.NoError(t, env.s.Commit())

	tl := &TeamProject{ProjectID: env.proj.ID}
	result, _, _, err := tl.ReadAll(env.s, env.attacker, "", 1, -1)
	require.NoError(t, err)

	teams, ok := result.([]*TeamWithPermission)
	require.True(t, ok)
	require.Len(t, teams, 1)
	assert.Equal(t, env.team.ID, teams[0].ID, "the attach row itself must be listed")
	assert.Equal(t, PermissionRead, teams[0].Permission, "the project permission must be listed")
	assert.Empty(t, teams[0].Name, "the foreign team name must be scrubbed")
	assert.Empty(t, teams[0].Description, "the foreign team description must be scrubbed")
	assert.Nil(t, teams[0].Members, "the foreign team members must be scrubbed")
}

func TestTeamProjectReadAllSearchesOnlyReadableTeams(t *testing.T) {
	env := setupTeamVisibilityEnv(t)
	defer env.s.Close()

	ownTeam := &Team{Name: "team-visibility readable needle", CreatedByID: env.attacker.ID}
	_, err := env.s.Insert(ownTeam)
	require.NoError(t, err)
	_, err = env.s.Insert(&TeamMember{TeamID: ownTeam.ID, UserID: env.attacker.ID})
	require.NoError(t, err)
	_, err = env.s.Insert(&TeamProject{
		TeamID:     env.team.ID,
		ProjectID:  env.proj.ID,
		Permission: PermissionRead,
	})
	require.NoError(t, err)
	_, err = env.s.Insert(&TeamProject{
		TeamID:     ownTeam.ID,
		ProjectID:  env.proj.ID,
		Permission: PermissionRead,
	})
	require.NoError(t, err)
	_, err = env.s.Insert(&TeamProject{
		TeamID:     env.publicTeam.ID,
		ProjectID:  env.proj.ID,
		Permission: PermissionRead,
	})
	require.NoError(t, err)
	require.NoError(t, env.s.Commit())

	tl := &TeamProject{ProjectID: env.proj.ID}
	assertSearch := func(search string, want *Team) {
		result, resultCount, totalItems, err := tl.ReadAll(env.s, env.attacker, search, 1, -1)
		require.NoError(t, err)
		teams, ok := result.([]*TeamWithPermission)
		require.True(t, ok)
		if want == nil {
			require.Empty(t, teams)
			assert.Zero(t, resultCount)
			assert.Zero(t, totalItems)
			return
		}
		require.Len(t, teams, 1)
		assert.Equal(t, want.ID, teams[0].ID)
		assert.Equal(t, want.Name, teams[0].Name)
		assert.Equal(t, 1, resultCount)
		assert.Equal(t, int64(1), totalItems)
	}

	assertSearch("private team", nil)
	assertSearch("readable needle", ownTeam)
	assertSearch("public team", nil)

	config.ServiceEnablePublicTeams.Set("true")
	t.Cleanup(func() { config.ServiceEnablePublicTeams.Set("false") })
	assertSearch("public team", env.publicTeam)
}

func TestListUsersFromProjectHidesUnreadableTeamMembers(t *testing.T) {
	env := setupTeamVisibilityEnv(t)
	defer env.s.Close()

	_, err := env.s.Insert(&TeamProject{
		TeamID:     env.team.ID,
		ProjectID:  env.proj.ID,
		Permission: PermissionRead,
	})
	require.NoError(t, err)

	ownTeam := &Team{Name: "team-visibility attacker team", CreatedByID: env.attacker.ID}
	_, err = env.s.Insert(ownTeam)
	require.NoError(t, err)
	_, err = env.s.Insert(&TeamMember{TeamID: ownTeam.ID, UserID: env.attacker.ID})
	require.NoError(t, err)
	_, err = env.s.Insert(&TeamMember{TeamID: env.team.ID, UserID: 3})
	require.NoError(t, err)

	reader := &user.User{ID: 7}
	_, err = env.s.Insert(&ProjectUser{UserID: reader.ID, ProjectID: env.proj.ID, Permission: PermissionRead})
	require.NoError(t, err)
	require.NoError(t, env.s.Commit())

	usernameSet := func(users []*user.User) map[string]bool {
		names := map[string]bool{}
		for _, u := range users {
			names[u.Username] = true
		}
		return names
	}

	users, err := ListUsersFromProject(env.s, env.proj, reader, "")
	require.NoError(t, err)
	names := usernameSet(users)
	assert.NotContains(t, names, "user1", "members of a team the caller cannot read must not be listed")
	assert.NotContains(t, names, "user3", "members of a team the caller cannot read must not be listed")
	assert.Contains(t, names, "user7", "the caller themself must be listed")

	users, err = ListUsersFromProject(env.s, env.proj, env.attacker, "")
	require.NoError(t, err)
	names = usernameSet(users)
	assert.Contains(t, names, "user1", "a project admin sees members of all attached teams")
	assert.Contains(t, names, "user3", "a project admin sees members of all attached teams")
}

func TestTeamProjectAdminStillManagesForeignAttachedTeams(t *testing.T) {
	env := setupTeamVisibilityEnv(t)
	defer env.s.Close()

	_, err := env.s.Insert(&TeamProject{
		TeamID:     env.team.ID,
		ProjectID:  env.proj.ID,
		Permission: PermissionRead,
	})
	require.NoError(t, err)
	require.NoError(t, env.s.Commit())

	tl := &TeamProject{TeamID: env.team.ID, ProjectID: env.proj.ID, Permission: PermissionWrite}
	can, err := tl.CanUpdate(env.s, env.attacker)
	require.NoError(t, err)
	require.True(t, can, "the project admin may update the permission of a foreign attached team")

	can, err = tl.CanDelete(env.s, env.attacker)
	require.NoError(t, err)
	require.True(t, can, "the project admin may remove a foreign attached team from their project")
}
