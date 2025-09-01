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
	"code.vikunja.io/api/pkg/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjectDuplicateService_NewProjectDuplicateService(t *testing.T) {
	service := NewProjectDuplicateService(db.GetEngine())
	
	assert.NotNil(t, service)
	assert.NotNil(t, service.DB)
	assert.NotNil(t, service.ProjectService)
	assert.NotNil(t, service.TaskService)
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
		
		// Get an existing project from fixtures (project 1 exists in test fixtures)
		sourceProjectID := int64(1)
		parentProjectID := int64(0) // No parent
		
		duplicatedProject, err := service.Duplicate(s, sourceProjectID, parentProjectID, user1)
		
		// For now, we expect this to work with the basic structure
		// The actual duplication logic will be implemented in subsequent steps
		require.NoError(t, err)
		assert.NotNil(t, duplicatedProject)
		assert.NotEqual(t, sourceProjectID, duplicatedProject.ID)
		assert.Contains(t, duplicatedProject.Title, "duplicate")
		assert.Equal(t, user1.ID, duplicatedProject.OwnerID)
	})

	t.Run("permission denied - no read access to source", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()
		
		service := NewProjectDuplicateService(db.GetEngine())
		user2 := &user.User{ID: 2}
		
		// Try to duplicate a project the user doesn't have access to
		sourceProjectID := int64(3) // Project 3 should not be accessible to user 2
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
		sourceProjectID := int64(1)  // User 2 can read this
		parentProjectID := int64(3)  // But can't write to this parent
		
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
	t.Run("empty project", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()
		
		service := NewProjectDuplicateService(db.GetEngine())
		user1 := &user.User{ID: 1}
		
		// Test with a project that has no tasks
		sourceProjectID := int64(1)
		targetProjectID := int64(2)
		
		taskIDMap, err := service.duplicateTasksAndRelatedData(s, sourceProjectID, targetProjectID, user1)
		
		assert.NoError(t, err)
		assert.NotNil(t, taskIDMap)
		// For now, this returns an empty map as we haven't implemented the logic yet
		assert.Equal(t, 0, len(taskIDMap))
	})
}

func TestProjectDuplicateService_duplicateProjectViews(t *testing.T) {
	t.Run("basic views duplication", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()
		
		service := NewProjectDuplicateService(db.GetEngine())
		user1 := &user.User{ID: 1}
		taskIDMap := make(map[int64]int64)
		
		sourceProjectID := int64(1)
		targetProjectID := int64(2)
		
		err := service.duplicateProjectViews(s, sourceProjectID, targetProjectID, user1, taskIDMap)
		
		// For now, this should not error as it's a stub
		assert.NoError(t, err)
	})
}

func TestProjectDuplicateService_duplicateProjectMetadata(t *testing.T) {
	t.Run("basic metadata duplication", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()
		
		service := NewProjectDuplicateService(db.GetEngine())
		user1 := &user.User{ID: 1}
		
		sourceProjectID := int64(1)
		targetProjectID := int64(2)
		
		err := service.duplicateProjectMetadata(s, sourceProjectID, targetProjectID, user1)
		
		// For now, this should not error as it's a stub
		assert.NoError(t, err)
	})
}
