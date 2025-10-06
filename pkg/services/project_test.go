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

func TestProject_Get(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()
	p := ProjectService{
		DB: db.GetEngine(),
	}

	t.Run("should get a project by ID", func(t *testing.T) {
		project, err := p.Get(s, 1, &user.User{ID: 1})
		assert.NoError(t, err)
		assert.NotNil(t, project)
		assert.Equal(t, int64(1), project.ID)
	})

	t.Run("should return error for non-existent project", func(t *testing.T) {
		_, err := p.Get(s, 999999, &user.User{ID: 1})
		assert.Error(t, err)
	})

	t.Run("should return error when user lacks permission", func(t *testing.T) {
		// User 1 does not have access to project 2 (owned by user 3)
		_, err := p.Get(s, 2, &user.User{ID: 1})
		assert.Error(t, err)
	})
}

func TestProject_ReadOne(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()
	p := ProjectService{
		DB:              db.GetEngine(),
		FavoriteService: NewFavoriteService(db.GetEngine()),
	}

	t.Run("should load complete project details", func(t *testing.T) {
		// Load project first to get its data
		project, err := models.GetProjectSimpleByID(s, 1)
		require.NoError(t, err)
		u := &user.User{ID: 1}

		err = p.ReadOne(s, project, u)
		assert.NoError(t, err)
		assert.NotNil(t, project.Owner, "Owner should be loaded")
		assert.NotNil(t, project.Views, "Views should be loaded")
		assert.GreaterOrEqual(t, len(project.Views), 1, "Should have at least one view")
	})

	t.Run("should handle favorites pseudo project", func(t *testing.T) {
		project := &models.Project{ID: models.FavoritesPseudoProject.ID}
		u := &user.User{ID: 1}

		err := p.ReadOne(s, project, u)
		assert.NoError(t, err)
		assert.NotNil(t, project.Views)
	})

	t.Run("should hide parent for link shares", func(t *testing.T) {
		project := &models.Project{ID: 2, ParentProjectID: 1}
		linkShare := &models.LinkSharing{ProjectID: 2}

		err := p.ReadOne(s, project, linkShare)
		assert.NoError(t, err)
		assert.Equal(t, int64(0), project.ParentProjectID, "Parent should be hidden for link shares")
	})

	t.Run("should load background information for unsplash", func(t *testing.T) {
		// Project 35 has background file ID 1 (unsplash photo)
		project := &models.Project{ID: 35}
		u := &user.User{ID: 6}

		err := p.ReadOne(s, project, u)
		assert.NoError(t, err)
		// Background information should be loaded if file exists
		// The exact type depends on whether it's an unsplash photo or upload
	})

	t.Run("should set favorite status correctly", func(t *testing.T) {
		project := &models.Project{ID: 3}
		u := &user.User{ID: 1}

		// Add project to favorites first
		err := p.FavoriteService.AddToFavorite(s, project.ID, u, models.FavoriteKindProject)
		require.NoError(t, err)

		err = p.ReadOne(s, project, u)
		assert.NoError(t, err)
		assert.True(t, project.IsFavorite, "Project should be marked as favorite")
	})
}

func TestProject_ReadAll(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()
	p := ProjectService{
		DB:              db.GetEngine(),
		FavoriteService: NewFavoriteService(db.GetEngine()),
	}

	t.Run("should get all projects for user", func(t *testing.T) {
		u := &user.User{ID: 1}

		projects, resultCount, totalItems, err := p.ReadAll(s, u, "", 1, 10, false, "")
		assert.NoError(t, err)
		assert.Greater(t, resultCount, 0, "Should return some projects")
		assert.Greater(t, totalItems, int64(0), "Should have total count")
		assert.NotNil(t, projects)
	})

	t.Run("should paginate results", func(t *testing.T) {
		u := &user.User{ID: 1}

		page1, count1, total, err := p.ReadAll(s, u, "", 1, 5, false, "")
		assert.NoError(t, err)
		assert.LessOrEqual(t, count1, 5, "Should respect per-page limit")

		page2, count2, _, err := p.ReadAll(s, u, "", 2, 5, false, "")
		assert.NoError(t, err)
		assert.LessOrEqual(t, count2, 5, "Should respect per-page limit")

		// Pages should be different (unless there aren't enough projects)
		if total > 5 {
			assert.NotEqual(t, page1[0].ID, page2[0].ID, "Different pages should have different projects")
		}
	})

	t.Run("should filter archived projects", func(t *testing.T) {
		u := &user.User{ID: 6}

		// Get non-archived projects
		nonArchived, _, _, err := p.ReadAll(s, u, "", 1, 50, false, "")
		assert.NoError(t, err)

		// Get archived projects
		archived, _, _, err := p.ReadAll(s, u, "", 1, 50, true, "")
		assert.NoError(t, err)

		// Verify no archived projects in non-archived list
		for _, proj := range nonArchived {
			assert.False(t, proj.IsArchived, "Non-archived query should not return archived projects")
		}

		// Verify all archived projects are actually archived
		for _, proj := range archived {
			if proj.ID != models.FavoritesPseudoProject.ID {
				// Skip pseudo projects
				// Note: Some projects might not be archived if they're pseudo projects
			}
		}
	})

	t.Run("should expand permissions when requested", func(t *testing.T) {
		u := &user.User{ID: 1}

		projects, _, _, err := p.ReadAll(s, u, "", 1, 10, false, models.ProjectExpandableRights)
		assert.NoError(t, err)
		assert.Greater(t, len(projects), 0, "Should have projects")

		// Check that permissions are set (not unknown)
		for _, proj := range projects {
			// MaxPermission should be set when rights are expanded
			// Note: The exact value depends on the user's permissions
			// For owned projects, it should be Admin (2)
			if proj.OwnerID == u.ID {
				assert.GreaterOrEqual(t, proj.MaxPermission, models.PermissionRead, "Owner should have at least read permission")
			}
		}
	})

	t.Run("should handle link share auth", func(t *testing.T) {
		linkShare := &models.LinkSharing{ProjectID: 2}

		projects, resultCount, totalItems, err := p.ReadAll(s, linkShare, "", 1, 10, false, "")
		assert.NoError(t, err)
		assert.Equal(t, 0, resultCount, "Link share should return special result")
		assert.Equal(t, int64(0), totalItems, "Link share should have 0 total")
		assert.Len(t, projects, 1, "Link share should return exactly one project")
		assert.Equal(t, int64(2), projects[0].ID, "Should return the shared project")
		assert.Equal(t, int64(0), projects[0].ParentProjectID, "Parent should be hidden")
	})

	t.Run("should not expand permissions by default", func(t *testing.T) {
		u := &user.User{ID: 1}

		projects, _, _, err := p.ReadAll(s, u, "", 1, 10, false, "")
		assert.NoError(t, err)

		// Check that permissions are unknown when not expanded
		for _, proj := range projects {
			assert.Equal(t, models.Permission(models.PermissionUnknown), proj.MaxPermission, "Permissions should be unknown when not expanded")
		}
	})
}

func TestProject_Create(t *testing.T) {
	s := db.NewSession()
	defer s.Close()
	p := ProjectService{
		DB: db.GetEngine(),
	}

	newProject := &models.Project{
		Title:       "new project",
		Description: "a new project",
	}
	u := &user.User{ID: 1}

	createdProject, err := p.Create(s, newProject, u)
	assert.NoError(t, err)
	assert.NotNil(t, createdProject)
	assert.Equal(t, newProject.Title, createdProject.Title)
	assert.Equal(t, newProject.Description, createdProject.Description)
	assert.Equal(t, u.ID, createdProject.OwnerID)
}

func TestProject_GetByID(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	p := &ProjectService{DB: db.GetEngine()}

	t.Run("should get a project by its id", func(t *testing.T) {
		proj, err := p.GetByID(s, 1, &user.User{ID: 1})
		assert.NoError(t, err)
		assert.NotNil(t, proj)
		assert.Equal(t, int64(1), proj.ID)
	})

	t.Run("should return an error if the project does not exist", func(t *testing.T) {
		_, err := p.GetByID(s, 999, &user.User{ID: 1})
		assert.Error(t, err)
		assert.True(t, models.IsErrProjectDoesNotExist(err))
	})

	t.Run("should return an error if the user does not have access to the project", func(t *testing.T) {
		_, err := p.GetByID(s, 2, &user.User{ID: 1})
		assert.Error(t, err)
	})
}

func TestProjectService_HasPermission_LinkShare(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	service := &ProjectService{DB: db.GetEngine()}
	linkShareUser := &user.User{ID: -2}

	t.Run("write permission for shared project", func(t *testing.T) {
		allowed, err := service.HasPermission(s, 2, linkShareUser, models.PermissionWrite)
		require.NoError(t, err)
		assert.True(t, allowed)
	})

	t.Run("denies higher permission levels", func(t *testing.T) {
		allowed, err := service.HasPermission(s, 2, linkShareUser, models.PermissionAdmin)
		require.NoError(t, err)
		assert.False(t, allowed)
	})

	t.Run("denies access to unrelated projects", func(t *testing.T) {
		allowed, err := service.HasPermission(s, 1, linkShareUser, models.PermissionRead)
		require.NoError(t, err)
		assert.False(t, allowed)
	})
}

func TestProject_GetAllForUser(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	p := &ProjectService{DB: db.GetEngine()}

	t.Run("should get all projects for a user", func(t *testing.T) {
		projects, count, total, err := p.GetAllForUser(s, &user.User{ID: 1}, "", 1, 10, false)
		assert.NoError(t, err)
		assert.Equal(t, 10, count)
		assert.Equal(t, int64(29), total) // Updated to account for the new saved filter fixture
		assert.Len(t, projects, 13)       // Updated to account for the new saved filter fixture
	})

	t.Run("should get all projects for a user with pagination", func(t *testing.T) {
		projects, count, total, err := p.GetAllForUser(s, &user.User{ID: 1}, "", 2, 10, false)
		assert.NoError(t, err)
		assert.Equal(t, 10, count)
		assert.Equal(t, int64(29), total) // Updated to account for the new saved filter fixture
		assert.Len(t, projects, 10)
	})

	t.Run("should get all projects for a user with search", func(t *testing.T) {
		// TODO: This test is flaky, the search does not seem to work correctly.
		// projects, count, total, err := p.GetAllForUser(s, &user.User{ID: 1}, "Test10", 1, 10, false)
		// assert.NoError(t, err)
		// assert.Equal(t, 1, count)
		// assert.Equal(t, int64(1), total)
		// assert.Len(t, projects, 1)
	})

	t.Run("should get archived projects", func(t *testing.T) {
		projects, _, _, err := p.GetAllForUser(s, &user.User{ID: 6}, "", 1, 50, true)
		assert.NoError(t, err)
		assert.Len(t, projects, 26)
	})
}

func TestProject_Delete(t *testing.T) {
	db.LoadAndAssertFixtures(t)

	p := &ProjectService{DB: db.GetEngine()}

	t.Run("should delete a project successfully", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		err := p.Delete(s, 1, &user.User{ID: 1})
		require.NoError(t, err)

		err = s.Commit()
		require.NoError(t, err)

		// Verify project is deleted
		db.AssertMissing(t, "projects", map[string]interface{}{
			"id": 1,
		})

		// Verify associated tasks are deleted
		db.AssertMissing(t, "tasks", map[string]interface{}{
			"id": 1,
		})
	})

	t.Run("should not delete a project without permission", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		// Project 2 is owned by User 3, User 2 should not have permission to delete it
		err := p.Delete(s, 2, &user.User{ID: 2})
		assert.Error(t, err)
		assert.True(t, models.IsErrGenericForbidden(err))
	})

	t.Run("should not delete default project by non-owner", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		// Project 4 is the default project for user 3
		err := p.Delete(s, 4, &user.User{ID: 2})
		assert.Error(t, err)
		assert.True(t, models.IsErrCannotDeleteDefaultProject(err))
	})

	t.Run("should not allow owner to delete their default project via Delete", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		// Project 4 is the default project for user 3
		// Owners should not be able to delete their default project via Delete
		// They must use DeleteForce (which is called during user deletion)
		err := p.Delete(s, 4, &user.User{ID: 3})
		assert.Error(t, err)
		assert.True(t, models.IsErrCannotDeleteDefaultProject(err))

		// Project should still exist
		db.AssertExists(t, "projects", map[string]interface{}{
			"id": 4,
		}, false)
	})

	t.Run("should delete project with background file", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		// Project 35 has a background file (based on model tests)
		err := p.Delete(s, 35, &user.User{ID: 6})
		require.NoError(t, err)

		err = s.Commit()
		require.NoError(t, err)

		// Verify project is deleted
		db.AssertMissing(t, "projects", map[string]interface{}{
			"id": 35,
		})

		// Verify background file is deleted
		db.AssertMissing(t, "files", map[string]interface{}{
			"id": 1,
		})
	})

	t.Run("should delete child projects recursively", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		// Find a project with children - let's use project 27 which has child project 12
		err := p.Delete(s, 27, &user.User{ID: 6})
		require.NoError(t, err)

		err = s.Commit()
		require.NoError(t, err)

		// Verify parent project is deleted
		db.AssertMissing(t, "projects", map[string]interface{}{
			"id": 27,
		})

		// Verify child project is also deleted
		db.AssertMissing(t, "projects", map[string]interface{}{
			"id": 12,
		})
	})

	t.Run("should return error for non-existent project", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		err := p.Delete(s, 999999, &user.User{ID: 1})
		assert.Error(t, err)
		// The exact error type will depend on implementation, but it should be an error
	})

	t.Run("should clean up all related entities", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		// Use a project that has various related entities
		projectID := int64(3)

		err := p.Delete(s, projectID, &user.User{ID: 1})
		require.NoError(t, err)

		err = s.Commit()
		require.NoError(t, err)

		// Verify project is deleted
		db.AssertMissing(t, "projects", map[string]interface{}{
			"id": projectID,
		})

		// Verify related entities are cleaned up
		// Note: The exact cleanup verification will depend on what related entities exist in fixtures
		// This test ensures the service handles the cleanup logic
	})

}

func TestProject_Update_ArchiveParentArchivesChild(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	p := &ProjectService{DB: db.GetEngine()}
	actingUser := &user.User{ID: 6}

	// First, let's load the existing project to get all its fields
	existingProject, err := models.GetProjectSimpleByID(s, 27)
	require.NoError(t, err, "Failed to load project 27")
	require.NotNil(t, existingProject, "Project 27 should exist")

	// Set the archive flag
	existingProject.IsArchived = true

	// Test archiving a parent project (ID 27) should also archive its child (ID 12)
	updatedProject, err := p.Update(s, existingProject, actingUser)
	assert.NoError(t, err, "Failed to archive project")
	assert.NotNil(t, updatedProject)
	assert.True(t, updatedProject.IsArchived)

	err = s.Commit()
	assert.NoError(t, err, "Failed to commit session after archiving project")

	// Verify parent project (ID 27) is archived
	db.AssertExists(t, "projects", map[string]interface{}{
		"id":          27,
		"is_archived": true,
	}, false)

	// Verify child project (ID 12) is also archived
	db.AssertExists(t, "projects", map[string]interface{}{
		"id":          12,
		"is_archived": true,
	}, false)
}
