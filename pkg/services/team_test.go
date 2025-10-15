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

package services

import (
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTeamService_Create(t *testing.T) {
	service := NewTeamService(testEngine)

	t.Run("Normal creation", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		doer := &user.User{ID: 1, Username: "user1"}
		team := &models.Team{
			Name:        "Test Team",
			Description: "Test Description",
		}

		created, err := service.Create(s, team, doer, true)
		require.NoError(t, err)
		assert.NotZero(t, created.ID)
		assert.Equal(t, "Test Team", created.Name)
		assert.Equal(t, "Test Description", created.Description)
		assert.Equal(t, doer.ID, created.CreatedByID)
		assert.NotNil(t, created.CreatedBy)
		assert.Len(t, created.Members, 1)
		assert.Equal(t, doer.ID, created.Members[0].ID)
		assert.True(t, created.Members[0].Admin)
	})

	t.Run("Creator as non-admin", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		doer := &user.User{ID: 1, Username: "user1"}
		team := &models.Team{
			Name: "Test Team Non-Admin",
		}

		created, err := service.Create(s, team, doer, false)
		require.NoError(t, err)
		assert.NotZero(t, created.ID)
		assert.Len(t, created.Members, 1)
		assert.False(t, created.Members[0].Admin)
	})

	t.Run("Empty name", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		doer := &user.User{ID: 1, Username: "user1"}
		team := &models.Team{}

		_, err := service.Create(s, team, doer, true)
		require.Error(t, err)
		assert.True(t, models.IsErrTeamNameCannotBeEmpty(err))
	})

	t.Run("Public team", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		doer := &user.User{ID: 1, Username: "user1"}
		team := &models.Team{
			Name:     "Public Test Team",
			IsPublic: true,
		}

		created, err := service.Create(s, team, doer, true)
		require.NoError(t, err)
		assert.True(t, created.IsPublic)
	})
}

func TestTeamService_GetByID(t *testing.T) {
	service := NewTeamService(testEngine)

	t.Run("Valid team", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		team, err := service.GetByID(s, 1)
		require.NoError(t, err)
		assert.Equal(t, int64(1), team.ID)
		assert.Equal(t, "testteam1", team.Name)
		assert.NotNil(t, team.CreatedBy)
		assert.NotEmpty(t, team.Members)
	})

	t.Run("Invalid ID", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := service.GetByID(s, -1)
		require.Error(t, err)
		assert.True(t, models.IsErrTeamDoesNotExist(err))
	})

	t.Run("Non-existent team", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := service.GetByID(s, 99999)
		require.Error(t, err)
		assert.True(t, models.IsErrTeamDoesNotExist(err))
	})
}

func TestTeamService_GetAll(t *testing.T) {
	service := NewTeamService(testEngine)

	t.Run("Get all teams for user", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		doer := &user.User{ID: 1}
		teams, count, total, err := service.GetAll(s, doer, "", 1, 50, false)
		require.NoError(t, err)
		assert.Greater(t, len(teams), 0)
		assert.Equal(t, len(teams), count)
		assert.Greater(t, total, int64(0))
	})

	t.Run("Search by name", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		doer := &user.User{ID: 1}
		teams, count, _, err := service.GetAll(s, doer, "testteam1", 1, 50, false)
		require.NoError(t, err)
		assert.Greater(t, count, 0)
		for _, team := range teams {
			assert.Contains(t, team.Name, "testteam1")
		}
	})

	t.Run("Pagination", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		doer := &user.User{ID: 1}
		teams, count, total, err := service.GetAll(s, doer, "", 1, 2, false)
		require.NoError(t, err)
		assert.LessOrEqual(t, count, 2)
		assert.Greater(t, total, int64(0))
		_ = teams
	})

	t.Run("Link share forbidden", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		linkShare := &models.LinkSharing{ID: 1}
		_, _, _, err := service.GetAll(s, linkShare, "", 1, 50, false)
		require.Error(t, err)
		assert.True(t, models.IsErrGenericForbidden(err))
	})
}

func TestTeamService_Update(t *testing.T) {
	service := NewTeamService(testEngine)

	t.Run("Update name and description", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		team := &models.Team{
			ID:          1,
			Name:        "Updated Team Name",
			Description: "Updated Description",
		}

		updated, err := service.Update(s, team)
		require.NoError(t, err)
		assert.Equal(t, "Updated Team Name", updated.Name)
		assert.Equal(t, "Updated Description", updated.Description)
	})

	t.Run("Update public status", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		team := &models.Team{
			ID:       1,
			Name:     "testteam1",
			IsPublic: true,
		}

		updated, err := service.Update(s, team)
		require.NoError(t, err)
		assert.True(t, updated.IsPublic)
	})

	t.Run("Empty name", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		team := &models.Team{
			ID:   1,
			Name: "",
		}

		_, err := service.Update(s, team)
		require.Error(t, err)
		assert.True(t, models.IsErrTeamNameCannotBeEmpty(err))
	})

	t.Run("Non-existent team", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		team := &models.Team{
			ID:   99999,
			Name: "Updated Name",
		}

		_, err := service.Update(s, team)
		require.Error(t, err)
		assert.True(t, models.IsErrTeamDoesNotExist(err))
	})
}

func TestTeamService_Delete(t *testing.T) {
	service := NewTeamService(testEngine)

	t.Run("Delete team successfully", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		doer := &user.User{ID: 1}
		err := service.Delete(s, 1, doer)
		require.NoError(t, err)

		// Verify team is deleted
		_, err = service.GetByID(s, 1)
		require.Error(t, err)
		assert.True(t, models.IsErrTeamDoesNotExist(err))
	})

	t.Run("Delete removes team members", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		doer := &user.User{ID: 1}

		// Verify members exist before delete
		var countBefore int64
		countBefore, _ = s.Where("team_id = ?", 2).Count(&models.TeamMember{})
		assert.Greater(t, countBefore, int64(0))

		err := service.Delete(s, 2, doer)
		require.NoError(t, err)

		// Verify members are deleted
		var countAfter int64
		countAfter, _ = s.Where("team_id = ?", 2).Count(&models.TeamMember{})
		assert.Equal(t, int64(0), countAfter)
	})

	t.Run("Delete removes project associations", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		doer := &user.User{ID: 1}

		// Verify associations exist before delete
		var countBefore int64
		countBefore, _ = s.Where("team_id = ?", 1).Count(&models.TeamProject{})
		_ = countBefore // May or may not have associations

		err := service.Delete(s, 1, doer)
		require.NoError(t, err)

		// Verify associations are deleted
		var countAfter int64
		countAfter, _ = s.Where("team_id = ?", 1).Count(&models.TeamProject{})
		assert.Equal(t, int64(0), countAfter)
	})

	t.Run("Delete non-existent team", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		doer := &user.User{ID: 1}
		err := service.Delete(s, 99999, doer)
		require.Error(t, err)
		assert.True(t, models.IsErrTeamDoesNotExist(err))
	})
}

func TestTeamService_CanRead(t *testing.T) {
	service := NewTeamService(testEngine)

	t.Run("Member can read", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		doer := &user.User{ID: 1}
		can, maxPerm, err := service.CanRead(s, 1, doer)
		require.NoError(t, err)
		assert.True(t, can)
		assert.Greater(t, maxPerm, 0)
	})

	t.Run("Admin has admin permissions", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		doer := &user.User{ID: 1}
		can, maxPerm, err := service.CanRead(s, 1, doer)
		require.NoError(t, err)
		assert.True(t, can)
		assert.Equal(t, int(models.PermissionAdmin), maxPerm)
	})

	t.Run("Non-member cannot read", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		doer := &user.User{ID: 4}
		can, _, err := service.CanRead(s, 1, doer)
		require.NoError(t, err)
		assert.False(t, can)
	})
}

func TestTeamService_IsAdmin(t *testing.T) {
	service := NewTeamService(testEngine)

	t.Run("Admin user", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		doer := &user.User{ID: 1}
		isAdmin, err := service.IsAdmin(s, 1, doer)
		require.NoError(t, err)
		assert.True(t, isAdmin)
	})

	t.Run("Non-admin member", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// User 2 is a member of team 1 but not admin (based on fixtures)
		doer := &user.User{ID: 2}
		isAdmin, err := service.IsAdmin(s, 1, doer)
		require.NoError(t, err)
		assert.False(t, isAdmin)
	})

	t.Run("Non-member", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		doer := &user.User{ID: 4}
		isAdmin, err := service.IsAdmin(s, 1, doer)
		require.NoError(t, err)
		assert.False(t, isAdmin)
	})

	t.Run("Link share cannot be admin", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		linkShare := &models.LinkSharing{ID: 1}
		isAdmin, err := service.IsAdmin(s, 1, linkShare)
		require.NoError(t, err)
		assert.False(t, isAdmin)
	})
}

func TestTeamService_AddMember(t *testing.T) {
	service := NewTeamService(testEngine)

	t.Run("Add member successfully", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		doer := &user.User{ID: 1, Username: "user1"}
		member, err := service.AddMember(s, 1, "user3", false, doer)
		require.NoError(t, err)
		assert.Equal(t, int64(1), member.TeamID)
		assert.Equal(t, "user3", member.Username)
		assert.False(t, member.Admin)
	})

	t.Run("Add member as admin", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		doer := &user.User{ID: 1, Username: "user1"}
		member, err := service.AddMember(s, 1, "user4", true, doer)
		require.NoError(t, err)
		assert.True(t, member.Admin)
	})

	t.Run("Duplicate member", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		doer := &user.User{ID: 1, Username: "user1"}
		// user1 is already a member of team 1
		_, err := service.AddMember(s, 1, "user1", false, doer)
		require.Error(t, err)
		assert.True(t, models.IsErrUserIsMemberOfTeam(err))
	})

	t.Run("Non-existent user", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		doer := &user.User{ID: 1, Username: "user1"}
		_, err := service.AddMember(s, 1, "nonexistent", false, doer)
		require.Error(t, err)
	})

	t.Run("Non-existent team", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		doer := &user.User{ID: 1, Username: "user1"}
		_, err := service.AddMember(s, 99999, "user2", false, doer)
		require.Error(t, err)
		assert.True(t, models.IsErrTeamDoesNotExist(err))
	})
}

func TestTeamService_RemoveMember(t *testing.T) {
	service := NewTeamService(testEngine)

	t.Run("Remove member successfully", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Add a member first
		doer := &user.User{ID: 1, Username: "user1"}
		_, _ = service.AddMember(s, 1, "user3", false, doer)

		// Now remove them
		err := service.RemoveMember(s, 1, "user3")
		require.NoError(t, err)

		// Verify they're removed
		exists, _ := service.MembershipExists(s, 1, 3)
		assert.False(t, exists)
	})

	t.Run("Cannot remove last member", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Create a new team with only one member
		doer := &user.User{ID: 1, Username: "user1"}
		team, _ := service.Create(s, &models.Team{Name: "Single Member Team"}, doer, true)

		// Try to remove the only member
		err := service.RemoveMember(s, team.ID, "user1")
		require.Error(t, err)
		assert.True(t, models.IsErrCannotDeleteLastTeamMember(err))
	})

	t.Run("Non-existent user", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		err := service.RemoveMember(s, 1, "nonexistent")
		require.Error(t, err)
	})
}

func TestTeamService_UpdateMemberAdmin(t *testing.T) {
	service := NewTeamService(testEngine)

	t.Run("Toggle admin status", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Add a non-admin member
		doer := &user.User{ID: 1, Username: "user1"}
		_, _ = service.AddMember(s, 1, "user3", false, doer)

		// Toggle to admin
		isAdmin, err := service.UpdateMemberAdmin(s, 1, "user3")
		require.NoError(t, err)
		assert.True(t, isAdmin)

		// Toggle back to non-admin
		isAdmin, err = service.UpdateMemberAdmin(s, 1, "user3")
		require.NoError(t, err)
		assert.False(t, isAdmin)
	})

	t.Run("Non-existent user", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := service.UpdateMemberAdmin(s, 1, "nonexistent")
		require.Error(t, err)
	})
}

func TestTeamService_GetTeamsByIDs(t *testing.T) {
	service := NewTeamService(testEngine)

	t.Run("Get multiple teams", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		teams, err := service.GetTeamsByIDs(s, []int64{1, 2})
		require.NoError(t, err)
		assert.Len(t, teams, 2)
	})

	t.Run("Empty list", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		teams, err := service.GetTeamsByIDs(s, []int64{})
		require.NoError(t, err)
		assert.Len(t, teams, 0)
	})

	t.Run("Partial matches", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		teams, err := service.GetTeamsByIDs(s, []int64{1, 99999})
		require.NoError(t, err)
		assert.Len(t, teams, 1)
		assert.Equal(t, int64(1), teams[0].ID)
	})
}

func TestTeamService_HasPermission(t *testing.T) {
	service := NewTeamService(testEngine)

	t.Run("Admin has write permission", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		doer := &user.User{ID: 1}
		has, err := service.HasPermission(s, 1, doer, models.PermissionWrite)
		require.NoError(t, err)
		assert.True(t, has)
	})

	t.Run("Admin has admin permission", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		doer := &user.User{ID: 1}
		has, err := service.HasPermission(s, 1, doer, models.PermissionAdmin)
		require.NoError(t, err)
		assert.True(t, has)
	})

	t.Run("Member has read permission", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		doer := &user.User{ID: 2}
		has, err := service.HasPermission(s, 1, doer, models.PermissionRead)
		require.NoError(t, err)
		assert.True(t, has)
	})

	t.Run("Non-member does not have permission", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		doer := &user.User{ID: 4}
		has, err := service.HasPermission(s, 1, doer, models.PermissionRead)
		require.NoError(t, err)
		assert.False(t, has)
	})

	t.Run("Invalid permission", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		doer := &user.User{ID: 1}
		_, err := service.HasPermission(s, 1, doer, models.Permission(99))
		require.Error(t, err)
		assert.True(t, models.IsErrInvalidPermission(err))
	})

	t.Run("Nil user", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		has, err := service.HasPermission(s, 1, nil, models.PermissionRead)
		require.NoError(t, err)
		assert.False(t, has)
	})
}

func TestTeamService_Get(t *testing.T) {
	service := NewTeamService(testEngine)

	t.Run("valid team", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		team, err := service.Get(s, 1)
		assert.NoError(t, err)
		assert.NotNil(t, team)
		assert.Equal(t, int64(1), team.ID)
	})

	t.Run("nonexistent team", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		team, err := service.Get(s, 999)
		assert.Error(t, err)
		assert.True(t, models.IsErrTeamDoesNotExist(err))
		assert.Nil(t, team)
	})
}

func TestTeamService_GetByIDSimple(t *testing.T) {
	service := NewTeamService(testEngine)

	t.Run("valid team", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		team, err := service.GetByIDSimple(s, 1)
		assert.NoError(t, err)
		assert.NotNil(t, team)
		assert.Equal(t, int64(1), team.ID)
	})

	t.Run("nonexistent team", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		team, err := service.GetByIDSimple(s, 999)
		assert.Error(t, err)
		assert.True(t, models.IsErrTeamDoesNotExist(err))
		assert.Nil(t, team)
	})
}

func TestTeamService_CanWrite(t *testing.T) {
	service := NewTeamService(testEngine)
	u := &user.User{ID: 1}

	t.Run("can write", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		canWrite, err := service.CanWrite(s, 1, u)
		assert.NoError(t, err)
		assert.True(t, canWrite)
	})

	t.Run("cannot write", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 999}
		canWrite, err := service.CanWrite(s, 1, u)
		assert.NoError(t, err)
		assert.False(t, canWrite)
	})
}

func TestTeamService_UpdateMemberPermission(t *testing.T) {
	service := NewTeamService(testEngine)

	t.Run("promote to admin", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// User 2 is a member of team 1 but not admin
		err := service.UpdateMemberPermission(s, 1, 2, true)
		assert.NoError(t, err)

		// Verify admin status changed
		tm := &models.TeamMember{}
		_, err = s.Where("team_id = ? AND user_id = ?", 1, 2).Get(tm)
		assert.NoError(t, err)
		assert.True(t, tm.Admin)
	})

	t.Run("demote from admin", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// User 1 is admin of team 1
		err := service.UpdateMemberPermission(s, 1, 1, false)
		assert.NoError(t, err)

		// Verify admin status changed
		tm := &models.TeamMember{}
		_, err = s.Where("team_id = ? AND user_id = ?", 1, 1).Get(tm)
		assert.NoError(t, err)
		assert.False(t, tm.Admin)
	})

	t.Run("nonexistent member", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// User 999 is not a member
		err := service.UpdateMemberPermission(s, 1, 999, true)
		assert.NoError(t, err) // No error, just doesn't update anything
	})
}

func TestTeamService_GetMembers(t *testing.T) {
	service := NewTeamService(testEngine)

	t.Run("all members", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		members, resultCount, total, err := service.GetMembers(s, 1, "", 1, 50)
		assert.NoError(t, err)
		assert.Greater(t, resultCount, 0)
		assert.Equal(t, int64(resultCount), total)
		assert.Len(t, members, resultCount)
	})

	t.Run("with search", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		members, resultCount, total, err := service.GetMembers(s, 1, "user", 1, 50)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, resultCount, 0)
		assert.Equal(t, int64(resultCount), total)
		assert.Len(t, members, resultCount)
	})

	t.Run("with pagination", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		members, resultCount, total, err := service.GetMembers(s, 1, "", 1, 1)
		assert.NoError(t, err)
		assert.LessOrEqual(t, resultCount, 1)
		assert.GreaterOrEqual(t, total, int64(1))
		assert.Len(t, members, resultCount)
	})

	t.Run("nonexistent team", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		members, resultCount, total, err := service.GetMembers(s, 999, "", 1, 50)
		assert.Error(t, err)
		assert.True(t, models.IsErrTeamDoesNotExist(err))
		assert.Equal(t, 0, resultCount)
		assert.Equal(t, int64(0), total)
		assert.Nil(t, members)
	})
}

func TestTeamService_IsMember(t *testing.T) {
	service := NewTeamService(testEngine)

	t.Run("is member", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		isMember, err := service.IsMember(s, 1, 1)
		assert.NoError(t, err)
		assert.True(t, isMember)
	})

	t.Run("is not member", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		isMember, err := service.IsMember(s, 1, 999)
		assert.NoError(t, err)
		assert.False(t, isMember)
	})

	t.Run("nonexistent team", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		isMember, err := service.IsMember(s, 999, 1)
		assert.NoError(t, err)
		assert.False(t, isMember)
	})
}
