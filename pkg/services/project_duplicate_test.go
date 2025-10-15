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
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjectDuplicateService_NewProjectDuplicateService(t *testing.T) {
	service := NewProjectDuplicateService(db.GetEngine())

	assert.NotNil(t, service)
	assert.NotNil(t, service.DB)
	assert.NotNil(t, service.Registry)
}

func TestProjectDuplicateService_Duplicate(t *testing.T) {
	// Initialize test fixtures
	files.InitTestFileFixtures(t)
	db.LoadAndAssertFixtures(t)

	t.Run("basic duplication", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewProjectDuplicateService(db.GetEngine())
		user1 := &user.User{ID: 1}

		// Get an existing project from fixtures (project 1 exists in test fixtures and is owned by user 1)
		sourceProjectID := int64(1)
		parentProjectID := int64(0) // No parent

		duplicatedProject, err := service.Duplicate(s, sourceProjectID, parentProjectID, user1)

		// For now, we expect this to work with the basic structure
		// The actual duplication logic will be implemented in subsequent steps
		require.NoError(t, err)
		assert.NotNil(t, duplicatedProject)
		assert.NotEqual(t, sourceProjectID, duplicatedProject.ID)
		// Check the title contains "duplicate" as per the service logic
		assert.Contains(t, duplicatedProject.Title, "duplicate")
		assert.Equal(t, user1.ID, duplicatedProject.OwnerID)
	})

	t.Run("permission denied - no read access to source", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewProjectDuplicateService(db.GetEngine())
		user2 := &user.User{ID: 2}

		// Try to duplicate a project the user doesn't have access to
		sourceProjectID := int64(1) // Project 1 is owned by user 1, user 2 has no access
		parentProjectID := int64(0)

		duplicatedProject, err := service.Duplicate(s, sourceProjectID, parentProjectID, user2)

		assert.Error(t, err)
		assert.Nil(t, duplicatedProject)
		assert.Equal(t, ErrAccessDenied, err)
	})

	t.Run("permission denied - no write access to parent", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewProjectDuplicateService(db.GetEngine())
		user2 := &user.User{ID: 2}

		// Try to duplicate into a parent project the user doesn't have write access to
		sourceProjectID := int64(3) // User 2 can read this (permission level 0)
		parentProjectID := int64(3) // But can't write to this same project (permission level 0 = read-only)

		duplicatedProject, err := service.Duplicate(s, sourceProjectID, parentProjectID, user2)

		assert.Error(t, err)
		assert.Nil(t, duplicatedProject)
		assert.Equal(t, ErrAccessDenied, err)
	})

	t.Run("nonexistent source project", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewProjectDuplicateService(db.GetEngine())
		user1 := &user.User{ID: 1}

		sourceProjectID := int64(99999) // Non-existent project
		parentProjectID := int64(0)

		duplicatedProject, err := service.Duplicate(s, sourceProjectID, parentProjectID, user1)

		assert.Error(t, err)
		assert.Nil(t, duplicatedProject)
	})
}

func TestProjectDuplicateService_duplicateTasksAndRelatedData(t *testing.T) {
	// Initialize test fixtures
	files.InitTestFileFixtures(t)
	db.LoadAndAssertFixtures(t)

	t.Run("duplicate tasks with related data", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewProjectDuplicateService(db.GetEngine())
		user1 := &user.User{ID: 1}

		// Test with project 1 which has tasks
		sourceProjectID := int64(1)
		targetProjectID := int64(2) // This should be a valid target project

		taskIDMap, err := service.duplicateTasksAndRelatedData(s, sourceProjectID, targetProjectID, user1)

		assert.NoError(t, err)
		assert.NotNil(t, taskIDMap)
		// Project 1 has multiple tasks, so the map should not be empty
		assert.Greater(t, len(taskIDMap), 0)
	})
}

func TestProjectDuplicateService_duplicateProjectViews(t *testing.T) {
	// Initialize test fixtures
	files.InitTestFileFixtures(t)
	db.LoadAndAssertFixtures(t)

	t.Run("basic views duplication", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewProjectDuplicateService(db.GetEngine())
		user1 := &user.User{ID: 1}
		taskIDMap := make(map[int64]int64)

		// Use projects that User 1 has access to
		sourceProjectID := int64(1) // User 1 owns this project
		targetProjectID := int64(1) // Use the same project for simplicity in unit testing

		err := service.duplicateProjectViews(s, sourceProjectID, targetProjectID, user1, taskIDMap)

		// This should work since User 1 has access to both projects
		assert.NoError(t, err)
	})
}

func TestProjectDuplicateService_duplicateProjectMetadata(t *testing.T) {
	t.Run("basic metadata duplication", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewProjectDuplicateService(db.GetEngine())
		user1 := &user.User{ID: 1}

		sourceProjectID := int64(1) // User 1 owns this project
		targetProjectID := int64(1) // Use the same project for simplicity in unit testing

		err := service.duplicateProjectMetadata(s, sourceProjectID, targetProjectID, user1)

		// This should work since User 1 has access to both projects
		assert.NoError(t, err)
	})
}

func TestProjectDuplicateService_CanCreate(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	service := NewProjectDuplicateService(testEngine)

	t.Run("owner with read access can duplicate", func(t *testing.T) {
		u := &user.User{ID: 1} // Owner of project 1
		can, err := service.CanCreate(s, 1, 0, u)
		assert.NoError(t, err)
		assert.True(t, can)
	})

	t.Run("user with read access can duplicate to parent they have write access to", func(t *testing.T) {
		u := &user.User{ID: 6} // Has write access to project 7 and read to project 6
		can, err := service.CanCreate(s, 6, 7, u)
		assert.NoError(t, err)
		assert.True(t, can)
	})

	t.Run("user without read access cannot duplicate", func(t *testing.T) {
		u := &user.User{ID: 13} // No access to project 1
		can, err := service.CanCreate(s, 1, 0, u)
		assert.NoError(t, err)
		assert.False(t, can)
	})

	t.Run("user cannot duplicate to parent they don't have write access to", func(t *testing.T) {
		u := &user.User{ID: 3} // Has read-only access to project 1
		can, err := service.CanCreate(s, 1, 2, u)
		assert.NoError(t, err)
		assert.False(t, can)
	})

	t.Run("link share cannot duplicate", func(t *testing.T) {
		ls := &models.LinkSharing{ID: 1}
		can, err := service.CanCreate(s, 1, 0, ls)
		assert.NoError(t, err)
		assert.False(t, can)
	})
}
