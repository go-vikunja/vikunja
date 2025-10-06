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
)

// testUserAuth implements web.Auth for testing
type testUserAuth struct {
	id int64
}

func (a *testUserAuth) GetID() int64 { return a.id }

func TestProjectTeamService_Create(t *testing.T) {
	db.LoadAndAssertFixtures(t)

	service := NewProjectTeamService(db.GetEngine())

	t.Run("create normally", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		tp := &models.TeamProject{
			TeamID:     1,
			ProjectID:  1,
			Permission: models.PermissionAdmin,
		}
		doer := &testUserAuth{id: 1}

		err := service.Create(s, tp, doer)
		assert.NoError(t, err)
		assert.NotEqual(t, int64(0), tp.ID, "ID should be set")

		// Verify it was created
		s.Commit()
		db.AssertExists(t, "team_projects", map[string]interface{}{
			"team_id":    1,
			"project_id": 1,
			"permission": models.PermissionAdmin,
		}, false)
	})

	t.Run("create for duplicate", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		tp := &models.TeamProject{
			TeamID:     1,
			ProjectID:  3, // team1 already has access to project 3
			Permission: models.PermissionAdmin,
		}
		doer := &testUserAuth{id: 1}

		err := service.Create(s, tp, doer)
		assert.Error(t, err)
		assert.True(t, models.IsErrTeamAlreadyHasAccess(err))
	})

	t.Run("create with invalid permission", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		tp := &models.TeamProject{
			TeamID:     1,
			ProjectID:  1,
			Permission: 500, // Invalid permission
		}
		doer := &testUserAuth{id: 1}

		err := service.Create(s, tp, doer)
		assert.Error(t, err)
		assert.True(t, models.IsErrInvalidPermission(err))
	})

	t.Run("create with nonexistent team", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		tp := &models.TeamProject{
			TeamID:    9999,
			ProjectID: 1,
		}
		doer := &testUserAuth{id: 1}

		err := service.Create(s, tp, doer)
		assert.Error(t, err)
		assert.True(t, models.IsErrTeamDoesNotExist(err))
	})

	t.Run("create with nonexistent project", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		tp := &models.TeamProject{
			TeamID:    1,
			ProjectID: 9999,
		}
		doer := &testUserAuth{id: 1}

		err := service.Create(s, tp, doer)
		assert.Error(t, err)
		assert.True(t, models.IsErrProjectDoesNotExist(err))
	})
}

func TestProjectTeamService_Delete(t *testing.T) {
	db.LoadAndAssertFixtures(t)

	service := NewProjectTeamService(db.GetEngine())

	t.Run("delete normally", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		tp := &models.TeamProject{
			TeamID:    1,
			ProjectID: 3, // team1 has access to project 3
		}

		err := service.Delete(s, tp)
		assert.NoError(t, err)

		s.Commit()
		db.AssertMissing(t, "team_projects", map[string]interface{}{
			"team_id":    1,
			"project_id": 3,
		})
	})

	t.Run("delete nonexistent team", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		tp := &models.TeamProject{
			TeamID:    9999,
			ProjectID: 3,
		}

		err := service.Delete(s, tp)
		assert.Error(t, err)
		assert.True(t, models.IsErrTeamDoesNotExist(err))
	})

	t.Run("delete team without access", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		tp := &models.TeamProject{
			TeamID:    1,
			ProjectID: 1, // team1 doesn't have access to project 1
		}

		err := service.Delete(s, tp)
		assert.Error(t, err)
		assert.True(t, models.IsErrTeamDoesNotHaveAccessToProject(err))
	})
}

func TestProjectTeamService_GetAll(t *testing.T) {
	db.LoadAndAssertFixtures(t)

	service := NewProjectTeamService(db.GetEngine())

	t.Run("get all teams for project", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		teams, count, total, err := service.GetAll(s, 3, &testUserAuth{id: 1}, "", 0, 50)
		assert.NoError(t, err)
		assert.Equal(t, 1, count, "Should have 1 team")
		assert.Equal(t, int64(1), total, "Should have total count of 1")
		assert.NotNil(t, teams)
		if len(teams) > 0 {
			assert.Equal(t, int64(1), teams[0].ID, "Should be team 1")
		}
	})

	t.Run("pagination", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		// Project 19 has 3 teams (teams 8, 9, 10)
		// Page 1 (or 0): start=0, limit=2 -> items 0-1
		teams, count, total, err := service.GetAll(s, 19, &user.User{ID: 1}, "", 1, 2)
		assert.NoError(t, err)
		assert.Equal(t, 2, count, "Should have 2 teams (page 1)")
		assert.Equal(t, int64(3), total, "Should have total count of 3")

		// Page 2: start=2, limit=2 -> item 2 (only 1 item left)
		teams2, count2, total2, err := service.GetAll(s, 19, &user.User{ID: 1}, "", 2, 2)
		assert.NoError(t, err)
		assert.Equal(t, 1, count2, "Should have 1 team (page 2)")
		assert.Equal(t, int64(3), total2, "Should have total count of 3")

		// Verify different teams on different pages
		if len(teams) > 0 && len(teams2) > 0 {
			assert.NotEqual(t, teams[0].ID, teams2[0].ID, "Different pages should have different teams")
		}
	})

	t.Run("search by team name", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		// Search for team9 on project 19
		teams, count, total, err := service.GetAll(s, 19, &user.User{ID: 1}, "TEAM9", 0, 50)
		assert.NoError(t, err)
		assert.Equal(t, 1, count, "Should find 1 team")
		assert.Equal(t, int64(1), total, "Should have total count of 1")
		if len(teams) > 0 {
			assert.Equal(t, int64(9), teams[0].ID, "Should be team 9")
		}
	})

	t.Run("no permission to read project", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		// User 2 doesn't have access to project 1
		_, _, _, err := service.GetAll(s, 1, &user.User{ID: 2}, "", 0, 50)
		assert.Error(t, err)
		assert.True(t, models.IsErrNeedToHaveProjectReadAccess(err))
	})
}

func TestProjectTeamService_Update(t *testing.T) {
	db.LoadAndAssertFixtures(t)

	service := NewProjectTeamService(db.GetEngine())

	t.Run("update normally", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		tp := &models.TeamProject{
			TeamID:     1,
			ProjectID:  3,
			Permission: models.PermissionAdmin,
		}

		err := service.Update(s, tp)
		assert.NoError(t, err)

		s.Commit()
		db.AssertExists(t, "team_projects", map[string]interface{}{
			"team_id":    1,
			"project_id": 3,
			"permission": models.PermissionAdmin,
		}, false)
	})

	t.Run("update to write", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		tp := &models.TeamProject{
			TeamID:     1,
			ProjectID:  3,
			Permission: models.PermissionWrite,
		}

		err := service.Update(s, tp)
		assert.NoError(t, err)

		s.Commit()
		db.AssertExists(t, "team_projects", map[string]interface{}{
			"team_id":    1,
			"project_id": 3,
			"permission": models.PermissionWrite,
		}, false)
	})

	t.Run("update to read", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		tp := &models.TeamProject{
			TeamID:     1,
			ProjectID:  3,
			Permission: models.PermissionRead,
		}

		err := service.Update(s, tp)
		assert.NoError(t, err)

		s.Commit()
		db.AssertExists(t, "team_projects", map[string]interface{}{
			"team_id":    1,
			"project_id": 3,
			"permission": models.PermissionRead,
		}, false)
	})

	t.Run("update with invalid permission", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		tp := &models.TeamProject{
			TeamID:     1,
			ProjectID:  3,
			Permission: 500, // Invalid permission
		}

		err := service.Update(s, tp)
		assert.Error(t, err)
		assert.True(t, models.IsErrInvalidPermission(err))
	})
}

func TestProjectTeamService_HasAccess(t *testing.T) {
	db.LoadAndAssertFixtures(t)

	service := NewProjectTeamService(db.GetEngine())

	t.Run("team with access", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		hasAccess, err := service.HasAccess(s, 3, 1) // team1 has access to project 3
		assert.NoError(t, err)
		assert.True(t, hasAccess)
	})

	t.Run("team without access", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		hasAccess, err := service.HasAccess(s, 1, 1) // team1 doesn't have access to project 1
		assert.NoError(t, err)
		assert.False(t, hasAccess)
	})
}

func TestProjectTeamService_GetPermission(t *testing.T) {
	db.LoadAndAssertFixtures(t)

	service := NewProjectTeamService(db.GetEngine())

	t.Run("get existing permission", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		permission, err := service.GetPermission(s, 3, 1) // team1 has read access to project 3
		assert.NoError(t, err)
		assert.Equal(t, models.PermissionRead, permission)
	})

	t.Run("get permission for team without access", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		permission, err := service.GetPermission(s, 1, 1) // team1 doesn't have access to project 1
		assert.NoError(t, err)
		assert.Equal(t, models.PermissionRead, permission) // Returns default PermissionRead
	})

	t.Run("get write permission", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		permission, err := service.GetPermission(s, 7, 3) // team3 has write access to project 7
		assert.NoError(t, err)
		assert.Equal(t, models.PermissionWrite, permission)
	})

	t.Run("get admin permission", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		permission, err := service.GetPermission(s, 8, 4) // team4 has admin access to project 8
		assert.NoError(t, err)
		assert.Equal(t, models.PermissionAdmin, permission)
	})
}
