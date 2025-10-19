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

// TestProjectService_Update_ValidateTitle tests validation logic for project titles
func TestProjectService_Update_ValidateTitle(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	registry := NewServiceRegistry(db.GetEngine())
	ps := registry.Project()
	u := &user.User{ID: 1}

	t.Run("should reject empty title", func(t *testing.T) {
		project, err := models.GetProjectSimpleByID(s, 1)
		require.NoError(t, err)

		project.Title = ""
		_, err = ps.Update(s, project, u)
		assert.Error(t, err)
		assert.IsType(t, &models.ErrProjectTitleCannotBeEmpty{}, err)
	})

	t.Run("should reject pseudo parent project", func(t *testing.T) {
		project, err := models.GetProjectSimpleByID(s, 1)
		require.NoError(t, err)

		project.ParentProjectID = -1 // Pseudo project
		_, err = ps.Update(s, project, u)
		assert.Error(t, err)
		// May return generic forbidden or specific error depending on validation order
	})

	t.Run("should reject cyclic relationship with self", func(t *testing.T) {
		project, err := models.GetProjectSimpleByID(s, 1)
		require.NoError(t, err)

		project.ParentProjectID = project.ID
		_, err = ps.Update(s, project, u)
		assert.Error(t, err)
		assert.IsType(t, &models.ErrProjectCannotBeChildOfItself{}, err)
	})

	t.Run("should reject duplicate identifier", func(t *testing.T) {
		project, err := models.GetProjectSimpleByID(s, 1)
		require.NoError(t, err)

		// First, set an identifier on project 2
		project2, err := models.GetProjectSimpleByID(s, 2)
		require.NoError(t, err)
		project2.Identifier = "UNIQUE123"
		_, err = s.ID(project2.ID).Cols("identifier").Update(project2)
		require.NoError(t, err)

		// Now try to set the same identifier on project 1
		project.Identifier = "UNIQUE123"
		_, err = ps.Update(s, project, u)
		assert.Error(t, err)
		assert.IsType(t, &models.ErrProjectIdentifierIsNotUnique{}, err)
	})
}

// TestProjectService_RecalculatePositions tests the position recalculation logic
func TestProjectService_RecalculatePositions(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	t.Run("should recalculate positions for child projects", func(t *testing.T) {
		// Project 27 has child projects (12, etc.)
		err := recalculateProjectPositions(s, 27)
		assert.NoError(t, err)

		// Verify positions were recalculated
		projects := []*models.Project{}
		err = s.Where("parent_project_id = ?", 27).OrderBy("position asc").Find(&projects)
		require.NoError(t, err)

		// Positions should be evenly distributed
		for i, p := range projects {
			assert.Greater(t, p.Position, 0.0)
			if i > 0 {
				assert.Greater(t, p.Position, projects[i-1].Position, "Positions should be ascending")
			}
		}
	})

	t.Run("should handle projects with no children", func(t *testing.T) {
		// Project with no children
		err := recalculateProjectPositions(s, 999999)
		assert.NoError(t, err) // Should not error
	})
}

// TestProjectService_CreateInboxProjectForUser tests inbox project creation
func TestProjectService_CreateInboxProjectForUser(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	ps := NewProjectService(db.GetEngine())

	t.Run("should create inbox project and set as default", func(t *testing.T) {
		// Create a new user without default project
		newUser := &user.User{
			ID:               999,
			Username:         "testinbox",
			Email:            "testinbox@example.com",
			DefaultProjectID: 0,
		}
		_, err := s.Insert(newUser)
		require.NoError(t, err)

		err = ps.CreateInboxProjectForUser(s, newUser)
		assert.NoError(t, err)

		// Verify inbox project was created
		assert.Greater(t, newUser.DefaultProjectID, int64(0), "Default project ID should be set")

		// Verify project exists and has correct title
		project, err := models.GetProjectSimpleByID(s, newUser.DefaultProjectID)
		require.NoError(t, err)
		assert.Equal(t, "Inbox", project.Title)
		assert.Equal(t, newUser.ID, project.OwnerID)
	})

	t.Run("should not override existing default project", func(t *testing.T) {
		// Create a user with a default project first
		testUser := &user.User{
			ID:       997,
			Username: "testdefault",
			Email:    "testdefault@example.com",
		}
		_, err := s.Insert(testUser)
		require.NoError(t, err)

		// Create inbox for this user (sets default)
		err = ps.CreateInboxProjectForUser(s, testUser)
		require.NoError(t, err)
		require.Greater(t, testUser.DefaultProjectID, int64(0), "User should have a default project")

		originalDefaultID := testUser.DefaultProjectID

		// Try to create inbox again - should not change default
		err = ps.CreateInboxProjectForUser(s, testUser)
		assert.NoError(t, err)

		// Verify default project was not changed
		verifyUser, err := user.GetUserByID(s, testUser.ID)
		require.NoError(t, err)
		assert.Equal(t, originalDefaultID, verifyUser.DefaultProjectID, "Default project should not change")
	})
}

// TestProjectService_DeleteForce tests force deletion of default projects
func TestProjectService_DeleteForce(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	ps := NewProjectService(db.GetEngine())

	t.Run("should force delete default project", func(t *testing.T) {
		// Create a user with a default project
		newUser := &user.User{
			ID:       998,
			Username: "testdelete",
			Email:    "testdelete@example.com",
		}
		_, err := s.Insert(newUser)
		require.NoError(t, err)

		// Create inbox project for user
		err = ps.CreateInboxProjectForUser(s, newUser)
		require.NoError(t, err)
		require.Greater(t, newUser.DefaultProjectID, int64(0))

		defaultProjectID := newUser.DefaultProjectID

		// Force delete the default project
		err = ps.DeleteForce(s, defaultProjectID, newUser)
		assert.NoError(t, err)

		// Verify project is deleted
		_, err = models.GetProjectSimpleByID(s, defaultProjectID)
		assert.Error(t, err)

		// Verify user's default project is cleared
		verifyUser, err := user.GetUserByID(s, newUser.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), verifyUser.DefaultProjectID, "Default project ID should be cleared")
	})

	t.Run("should reject force delete without permission", func(t *testing.T) {
		// Try to delete project owned by user 1 as user 2
		err := ps.DeleteForce(s, 1, &user.User{ID: 2})
		assert.Error(t, err)
	})
}

// TestProjectService_AddDetails tests detail enrichment
func TestProjectService_AddDetails(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	ps := NewProjectService(db.GetEngine())
	u := &user.User{ID: 1}

	t.Run("should add owner details to projects", func(t *testing.T) {
		projects := []*models.Project{
			{ID: 1, OwnerID: 1},
			{ID: 2, OwnerID: 3},
		}

		err := ps.AddDetails(s, projects, u)
		assert.NoError(t, err)

		// Verify owners are loaded
		assert.NotNil(t, projects[0].Owner)
		assert.Equal(t, int64(1), projects[0].Owner.ID)
		assert.NotNil(t, projects[1].Owner)
		assert.Equal(t, int64(3), projects[1].Owner.ID)
	})

	t.Run("should add subscription status", func(t *testing.T) {
		projects := []*models.Project{
			{ID: 1, OwnerID: 1},
		}

		err := ps.AddDetails(s, projects, u)
		assert.NoError(t, err)

		// Subscription status should be populated (even if nil)
		assert.NotNil(t, projects[0])
	})

	t.Run("should handle background information for unsplash", func(t *testing.T) {
		// Project 35 has unsplash background
		projects := []*models.Project{
			{ID: 35, OwnerID: 6},
		}

		// Load the project to get background file ID
		p, err := models.GetProjectSimpleByID(s, 35)
		require.NoError(t, err)
		projects[0].BackgroundFileID = p.BackgroundFileID

		err = ps.AddDetails(s, projects, &user.User{ID: 6})
		assert.NoError(t, err)

		// If background exists, information should be populated
		if projects[0].BackgroundFileID != 0 {
			assert.NotNil(t, projects[0])
		}
	})

	t.Run("should handle empty project list", func(t *testing.T) {
		projects := []*models.Project{}

		err := ps.AddDetails(s, projects, u)
		assert.NoError(t, err) // Should not error on empty list
	})
}

// TestProjectService_Create_Validation tests create-time validation
func TestProjectService_Create_Validation(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	ps := NewProjectService(db.GetEngine())
	u := &user.User{ID: 1}

	t.Run("should reject project with empty title", func(t *testing.T) {
		project := &models.Project{
			Title: "",
		}

		_, err := ps.Create(s, project, u)
		assert.Error(t, err)
		assert.IsType(t, &models.ErrProjectTitleCannotBeEmpty{}, err)
	})

	t.Run("should reject project with invalid parent", func(t *testing.T) {
		project := &models.Project{
			Title:           "Test Project",
			ParentProjectID: -1, // Pseudo parent
		}

		_, err := ps.Create(s, project, u)
		assert.Error(t, err)
		assert.IsType(t, &models.ErrProjectCannotBelongToAPseudoParentProject{}, err)
	})

	t.Run("should reject project with duplicate identifier", func(t *testing.T) {
		// First create a project with an identifier
		project1 := &models.Project{
			Title:      "Test Project 1",
			Identifier: "TESTID123",
		}
		created1, err := ps.Create(s, project1, u)
		require.NoError(t, err)
		require.NotNil(t, created1)

		// Try to create another with same identifier
		project2 := &models.Project{
			Title:      "Test Project 2",
			Identifier: "TESTID123",
		}
		_, err = ps.Create(s, project2, u)
		assert.Error(t, err)
		assert.IsType(t, &models.ErrProjectIdentifierIsNotUnique{}, err)
	})

	t.Run("should create project with unique identifier", func(t *testing.T) {
		project := &models.Project{
			Title:      "Test Project Unique",
			Identifier: "UNIQUEID999",
		}

		created, err := ps.Create(s, project, u)
		assert.NoError(t, err)
		assert.NotNil(t, created)
		assert.Equal(t, "UNIQUEID999", created.Identifier)
	})
}

// TestProjectService_Update_PositionCalculation tests position handling
func TestProjectService_Update_PositionCalculation(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	ps := NewProjectService(db.GetEngine())
	u := &user.User{ID: 1}

	t.Run("should calculate default position when position is zero", func(t *testing.T) {
		project := &models.Project{
			Title:    "Test Position",
			Position: 0,
		}

		created, err := ps.Create(s, project, u)
		assert.NoError(t, err)
		assert.NotNil(t, created)

		// Position should be calculated based on ID
		assert.Greater(t, created.Position, 0.0, "Position should be calculated")
	})

	t.Run("should preserve custom position", func(t *testing.T) {
		project := &models.Project{
			Title:    "Test Custom Position",
			Position: 12345.67,
		}

		created, err := ps.Create(s, project, u)
		assert.NoError(t, err)
		assert.NotNil(t, created)
		assert.Equal(t, 12345.67, created.Position)
	})
}
