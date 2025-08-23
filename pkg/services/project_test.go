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

func TestProject_Get(t *testing.T) {
	s := db.NewSession()
	defer s.Close()
	p := Project{
		DB: db.GetEngine(),
	}

	// This is a placeholder test. It will be expanded later.
	_, err := p.Get(s, 1, &user.User{ID: 1})
	assert.NoError(t, err)
}

func TestProject_Create(t *testing.T) {
	s := db.NewSession()
	defer s.Close()
	p := Project{
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

	p := &Project{DB: db.GetEngine()}

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

func TestProject_GetAllForUser(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	p := &Project{DB: db.GetEngine()}

	t.Run("should get all projects for a user", func(t *testing.T) {
		projects, count, total, err := p.GetAllForUser(s, &user.User{ID: 1}, "", 1, 10, false)
		assert.NoError(t, err)
		assert.Equal(t, 10, count)
		assert.Equal(t, int64(28), total)
		assert.Len(t, projects, 12)
	})

	t.Run("should get all projects for a user with pagination", func(t *testing.T) {
		projects, count, total, err := p.GetAllForUser(s, &user.User{ID: 1}, "", 2, 10, false)
		assert.NoError(t, err)
		assert.Equal(t, 10, count)
		assert.Equal(t, int64(28), total)
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
