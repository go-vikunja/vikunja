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
)

// TestProjectPermissions_OwnerInMultipleTeamsWithDifferentPermissions tests the scenario where:
// - User A creates a project (becomes owner)
// - User A is part of two teams Y and Z
// - Project is shared with team Y with admin permissions
// - Project is shared with team Z with read-only permissions
// - User A should still have owner/admin permissions (highest permission should apply)
func TestProjectPermissions_OwnerInMultipleTeamsWithDifferentPermissions(t *testing.T) {
	// Setup: Load fixtures and create a database session
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	// User A (we'll use user 1 from fixtures - needs Username for GetFromAuth)
	userA := &user.User{
		ID:       1,
		Username: "user1",
	}

	// Create a new project owned by User A
	project := &Project{
		Title:       "Test Project for Multiple Teams",
		Description: "Testing permissions when owner is in multiple teams",
	}
	err := project.Create(s, userA)
	require.NoError(t, err)
	err = s.Commit()
	require.NoError(t, err)

	// Verify User A has admin permissions as the owner
	t.Run("owner has admin permissions before sharing", func(t *testing.T) {
		s = db.NewSession()
		defer s.Close()

		canRead, maxPerm, err := project.CanRead(s, userA)
		require.NoError(t, err)
		assert.True(t, canRead, "Owner should be able to read their project")
		assert.Equal(t, int(PermissionAdmin), maxPerm, "Owner should have admin permission")

		canWrite, err := project.CanWrite(s, userA)
		require.NoError(t, err)
		assert.True(t, canWrite, "Owner should be able to write to their project")

		isAdmin, err := project.IsAdmin(s, userA)
		require.NoError(t, err)
		assert.True(t, isAdmin, "Owner should be admin of their project")
	})

	prepareSharingY := func() {
		s = db.NewSession()
		defer s.Close()

		// Create Team Y
		teamY := &Team{
			Name:        "Team Y",
			Description: "Team with admin permissions",
		}
		err = teamY.Create(s, userA)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		// Add User A to Team Y
		teamMemberY := &TeamMember{
			TeamID:   teamY.ID,
			Username: "user1",
		}
		err = teamMemberY.Create(s, userA)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		// Share project with Team Y (admin permissions)
		teamProjectY := &TeamProject{
			TeamID:     teamY.ID,
			ProjectID:  project.ID,
			Permission: PermissionAdmin,
		}
		err = teamProjectY.Create(s, userA)
		require.NoError(t, err)
		require.NoError(t, s.Commit())
	}

	// Verify User A still has admin permissions after sharing with Team Y
	t.Run("owner has admin permissions after sharing with team Y (admin)", func(t *testing.T) {
		s = db.NewSession()
		defer s.Close()

		prepareSharingY()

		canRead, maxPerm, err := project.CanRead(s, userA)
		require.NoError(t, err)
		assert.True(t, canRead, "Owner should be able to read their project after sharing with Team Y")
		assert.Equal(t, int(PermissionAdmin), maxPerm, "Owner should still have admin permission after sharing with Team Y")

		canWrite, err := project.CanWrite(s, userA)
		require.NoError(t, err)
		assert.True(t, canWrite, "Owner should be able to write to their project after sharing with Team Y")

		isAdmin, err := project.IsAdmin(s, userA)
		require.NoError(t, err)
		assert.True(t, isAdmin, "Owner should still be admin after sharing with Team Y")
	})

	prepareTeamSharedMultiple := func() {
		s = db.NewSession()
		defer s.Close()

		// Create Team Z
		teamZ := &Team{
			Name:        "Team Z",
			Description: "Team with read-only permissions",
		}
		err = teamZ.Create(s, userA)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		// Add User A to Team Z
		teamMemberZ := &TeamMember{
			TeamID:   teamZ.ID,
			Username: "user1",
		}
		err = teamMemberZ.Create(s, userA)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		// Share project with Team Z (read-only permissions)
		teamProjectZ := &TeamProject{
			TeamID:     teamZ.ID,
			ProjectID:  project.ID,
			Permission: PermissionRead,
		}
		err = teamProjectZ.Create(s, userA)
		require.NoError(t, err)
		require.NoError(t, s.Commit())
	}

	// Verify User A STILL has admin permissions after sharing with Team Z
	// user should retain highest permission (owner/admin), not be downgraded to read-only
	t.Run("owner has admin permissions after sharing with team Z (read-only)", func(t *testing.T) {
		s = db.NewSession()
		defer s.Close()

		prepareTeamSharedMultiple()

		canRead, maxPerm, err := project.CanRead(s, userA)
		require.NoError(t, err)
		assert.True(t, canRead, "Owner should be able to read their project after sharing with Team Z")
		assert.Equal(t, int(PermissionAdmin), maxPerm, "Owner should STILL have admin permission (not downgraded to read-only)")

		canWrite, err := project.CanWrite(s, userA)
		require.NoError(t, err)
		assert.True(t, canWrite, "Owner should STILL be able to write to their project (not downgraded to read-only)")

		isAdmin, err := project.IsAdmin(s, userA)
		require.NoError(t, err)
		assert.True(t, isAdmin, "Owner should STILL be admin (not downgraded to read-only)")
	})
}

// TestProjectPermissions_OwnerInMultipleTeamsAdminThenWrite tests the variant where:
// - User A creates a project
// - Project is shared with team Y with admin permissions
// - Project is shared with team Z with write permissions
// - User A should still have admin permissions (not downgraded to write)
func TestProjectPermissions_OwnerInMultipleTeamsAdminThenWrite(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	userA := &user.User{
		ID:       1,
		Username: "user1",
	}

	// Create a new project owned by User A
	project := &Project{
		Title:       "Test Project Admin Then Write",
		Description: "Testing permissions when owner is in teams with admin then write",
	}
	err := project.Create(s, userA)
	require.NoError(t, err)
	err = s.Commit()
	require.NoError(t, err)

	// Create Team Y with admin permissions
	teamY := &Team{
		Name: "Team Y Admin",
	}
	err = teamY.Create(s, userA)
	require.NoError(t, err)
	err = s.Commit()
	require.NoError(t, err)

	teamMemberY := &TeamMember{
		TeamID:   teamY.ID,
		Username: "user1",
	}
	err = teamMemberY.Create(s, userA)
	require.NoError(t, err)
	err = s.Commit()
	require.NoError(t, err)

	teamProjectY := &TeamProject{
		TeamID:     teamY.ID,
		ProjectID:  project.ID,
		Permission: PermissionAdmin,
	}
	err = teamProjectY.Create(s, userA)
	require.NoError(t, err)
	err = s.Commit()
	require.NoError(t, err)

	// Create Team Z with write permissions
	teamZ := &Team{
		Name: "Team Z Write",
	}
	err = teamZ.Create(s, userA)
	require.NoError(t, err)
	err = s.Commit()
	require.NoError(t, err)

	teamMemberZ := &TeamMember{
		TeamID:   teamZ.ID,
		Username: "user1",
	}
	err = teamMemberZ.Create(s, userA)
	require.NoError(t, err)
	err = s.Commit()
	require.NoError(t, err)

	teamProjectZ := &TeamProject{
		TeamID:     teamZ.ID,
		ProjectID:  project.ID,
		Permission: PermissionWrite,
	}
	err = teamProjectZ.Create(s, userA)
	require.NoError(t, err)
	err = s.Commit()
	require.NoError(t, err)

	// Test: User A should still have admin permissions (not downgraded to write)
	t.Run("owner has admin permissions after sharing with admin team then write team", func(t *testing.T) {
		s = db.NewSession()
		defer s.Close()

		canRead, maxPerm, err := project.CanRead(s, userA)
		require.NoError(t, err)
		assert.True(t, canRead)
		assert.Equal(t, int(PermissionAdmin), maxPerm, "Owner should have admin permission, not downgraded to write")

		isAdmin, err := project.IsAdmin(s, userA)
		require.NoError(t, err)
		assert.True(t, isAdmin, "Owner should still be admin, not downgraded to write")
	})
}

// TestProjectPermissions_NonOwnerInMultipleTeams tests the scenario where:
// - User B (not the owner) is part of two teams
// - Team Y has admin permissions on a project
// - Team Z has read-only permissions on the same project
// - User B should have admin permissions (highest of the two)
func TestProjectPermissions_NonOwnerInMultipleTeams(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	// User 1 is the owner
	owner := &user.User{
		ID:       1,
		Username: "user1",
	}
	// User 2 will be in multiple teams
	userB := &user.User{
		ID:       2,
		Username: "user2",
	}

	// Create a new project owned by User 1
	project := &Project{
		Title:       "Test Project Non-Owner Multiple Teams",
		Description: "Testing permissions for non-owner in multiple teams",
	}
	err := project.Create(s, owner)
	require.NoError(t, err)
	err = s.Commit()
	require.NoError(t, err)

	// Create Team Y with admin permissions
	teamY := &Team{
		Name: "Team Y Admin for Non-Owner",
	}
	err = teamY.Create(s, owner)
	require.NoError(t, err)
	err = s.Commit()
	require.NoError(t, err)

	// Add User B to Team Y
	teamMemberY := &TeamMember{
		TeamID:   teamY.ID,
		Username: "user2",
	}
	err = teamMemberY.Create(s, owner)
	require.NoError(t, err)
	err = s.Commit()
	require.NoError(t, err)

	// Share project with Team Y (admin)
	teamProjectY := &TeamProject{
		TeamID:     teamY.ID,
		ProjectID:  project.ID,
		Permission: PermissionAdmin,
	}
	err = teamProjectY.Create(s, owner)
	require.NoError(t, err)
	err = s.Commit()
	require.NoError(t, err)

	// Verify User B has admin permissions through Team Y
	t.Run("non-owner has admin through team Y", func(t *testing.T) {
		s = db.NewSession()
		defer s.Close()

		canRead, maxPerm, err := project.CanRead(s, userB)
		require.NoError(t, err)
		assert.True(t, canRead)
		assert.Equal(t, int(PermissionAdmin), maxPerm)

		isAdmin, err := project.IsAdmin(s, userB)
		require.NoError(t, err)
		assert.True(t, isAdmin)
	})

	// Create Team Z with read-only permissions
	teamZ := &Team{
		Name: "Team Z Read for Non-Owner",
	}
	err = teamZ.Create(s, owner)
	require.NoError(t, err)
	err = s.Commit()
	require.NoError(t, err)

	// Add User B to Team Z
	teamMemberZ := &TeamMember{
		TeamID:   teamZ.ID,
		Username: "user2",
	}
	err = teamMemberZ.Create(s, owner)
	require.NoError(t, err)
	err = s.Commit()
	require.NoError(t, err)

	// Share project with Team Z (read-only)
	teamProjectZ := &TeamProject{
		TeamID:     teamZ.ID,
		ProjectID:  project.ID,
		Permission: PermissionRead,
	}
	err = teamProjectZ.Create(s, owner)
	require.NoError(t, err)
	err = s.Commit()
	require.NoError(t, err)

	// Test: User B should still have admin permissions (highest of Team Y and Team Z)
	t.Run("non-owner retains admin after adding to read-only team Z", func(t *testing.T) {
		s = db.NewSession()
		defer s.Close()

		canRead, maxPerm, err := project.CanRead(s, userB)
		require.NoError(t, err)
		assert.True(t, canRead)
		assert.Equal(t, int(PermissionAdmin), maxPerm, "Non-owner should have admin (highest permission), not downgraded to read")

		isAdmin, err := project.IsAdmin(s, userB)
		require.NoError(t, err)
		assert.True(t, isAdmin, "Non-owner should still have admin permission, not downgraded to read")
	})
}

// TestProjectPermissions_OrderOfSharingDoesNotMatter tests that the order in which
// teams are granted access doesn't affect the final permission level
func TestProjectPermissions_OrderOfSharingDoesNotMatter(t *testing.T) {
	db.LoadAndAssertFixtures(t)

	owner := &user.User{
		ID:       1,
		Username: "user1",
	}
	userB := &user.User{
		ID:       2,
		Username: "user2",
	}

	// Test 1: Share read-only first, then admin
	t.Run("share read-only first then admin", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		project := &Project{
			Title: "Test Order: Read then Admin",
		}
		err := project.Create(s, owner)
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)

		// Create and share with read-only team first
		teamRead := &Team{Name: "Read First Team"}
		err = teamRead.Create(s, owner)
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)

		teamMemberRead := &TeamMember{TeamID: teamRead.ID, Username: "user2"}
		err = teamMemberRead.Create(s, owner)
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)

		teamProjectRead := &TeamProject{
			TeamID:     teamRead.ID,
			ProjectID:  project.ID,
			Permission: PermissionRead,
		}
		err = teamProjectRead.Create(s, owner)
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)

		// Now share with admin team
		teamAdmin := &Team{Name: "Admin Second Team"}
		err = teamAdmin.Create(s, owner)
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)

		teamMemberAdmin := &TeamMember{TeamID: teamAdmin.ID, Username: "user2"}
		err = teamMemberAdmin.Create(s, owner)
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)

		teamProjectAdmin := &TeamProject{
			TeamID:     teamAdmin.ID,
			ProjectID:  project.ID,
			Permission: PermissionAdmin,
		}
		err = teamProjectAdmin.Create(s, owner)
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)

		// Verify user has admin permission
		s = db.NewSession()
		defer s.Close()

		canRead, maxPerm, err := project.CanRead(s, userB)
		require.NoError(t, err)
		assert.True(t, canRead)
		assert.Equal(t, int(PermissionAdmin), maxPerm, "User should have admin even though read-only was shared first")

		isAdmin, err := project.IsAdmin(s, userB)
		require.NoError(t, err)
		assert.True(t, isAdmin)
	})

	// Test 2: Share admin first, then read-only (the reported bug scenario)
	t.Run("share admin first then read-only - REPORTED BUG", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		project := &Project{
			Title: "Test Order: Admin then Read",
		}
		err := project.Create(s, owner)
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)

		// Create and share with admin team first
		teamAdmin := &Team{Name: "Admin First Team"}
		err = teamAdmin.Create(s, owner)
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)

		teamMemberAdmin := &TeamMember{TeamID: teamAdmin.ID, Username: "user2"}
		err = teamMemberAdmin.Create(s, owner)
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)

		teamProjectAdmin := &TeamProject{
			TeamID:     teamAdmin.ID,
			ProjectID:  project.ID,
			Permission: PermissionAdmin,
		}
		err = teamProjectAdmin.Create(s, owner)
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)

		// Now share with read-only team
		teamRead := &Team{Name: "Read Second Team"}
		err = teamRead.Create(s, owner)
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)

		teamMemberRead := &TeamMember{TeamID: teamRead.ID, Username: "user2"}
		err = teamMemberRead.Create(s, owner)
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)

		teamProjectRead := &TeamProject{
			TeamID:     teamRead.ID,
			ProjectID:  project.ID,
			Permission: PermissionRead,
		}
		err = teamProjectRead.Create(s, owner)
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)

		// Verify user STILL has admin permission (not downgraded to read)
		s = db.NewSession()
		defer s.Close()

		canRead, maxPerm, err := project.CanRead(s, userB)
		require.NoError(t, err)
		assert.True(t, canRead)
		assert.Equal(t, int(PermissionAdmin), maxPerm, "User should STILL have admin even though read-only was shared last")

		isAdmin, err := project.IsAdmin(s, userB)
		require.NoError(t, err)
		assert.True(t, isAdmin, "User should STILL be admin even though read-only was shared last")
	})
}
