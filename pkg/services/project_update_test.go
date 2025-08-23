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

func TestProject_Update(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	p := &Project{DB: db.GetEngine()}

	t.Run("should update a project", func(t *testing.T) {
		projectToUpdate, err := models.GetProjectSimpleByID(s, 1)
		assert.NoError(t, err)

		projectToUpdate.Title = "updated title"
		updatedProject, err := p.Update(s, projectToUpdate, &user.User{ID: 1})
		assert.NoError(t, err)
		assert.NotNil(t, updatedProject)
		assert.Equal(t, "updated title", updatedProject.Title)
	})

	t.Run("should not update a project without permission", func(t *testing.T) {
		projectToUpdate, err := models.GetProjectSimpleByID(s, 1)
		assert.NoError(t, err)

		projectToUpdate.Title = "updated title"
		_, err = p.Update(s, projectToUpdate, &user.User{ID: 2})
		assert.Error(t, err)
	})

	t.Run("should not update a project with an invalid title", func(t *testing.T) {
		projectToUpdate, err := models.GetProjectSimpleByID(s, 1)
		assert.NoError(t, err)

		projectToUpdate.Title = ""
		_, err = p.Update(s, projectToUpdate, &user.User{ID: 1})
		assert.Error(t, err)
	})

	t.Run("should not archive a default project", func(t *testing.T) {
		projectToUpdate, err := models.GetProjectSimpleByID(s, 4)
		assert.NoError(t, err)

		projectToUpdate.IsArchived = true
		_, err = p.Update(s, projectToUpdate, &user.User{ID: 3})
		assert.Error(t, err)
		assert.True(t, models.IsErrCannotArchiveDefaultProject(err))
	})

	t.Run("should not move a project to a parent without permissions", func(t *testing.T) {
		projectToUpdate, err := models.GetProjectSimpleByID(s, 1)
		assert.NoError(t, err)

		projectToUpdate.ParentProjectID = 2
		_, err = p.Update(s, projectToUpdate, &user.User{ID: 1})
		assert.Error(t, err)
	})

	t.Run("archiving a project should archive its descendants", func(t *testing.T) {
		// Project 27 is parent of 12
		projectToUpdate, err := models.GetProjectSimpleByID(s, 27)
		assert.NoError(t, err)

		projectToUpdate.IsArchived = true
		_, err = p.Update(s, projectToUpdate, &user.User{ID: 6})
		assert.NoError(t, err)

		childProject, err := models.GetProjectSimpleByID(s, 12)
		assert.NoError(t, err)
		assert.True(t, childProject.IsArchived)
	})
}
