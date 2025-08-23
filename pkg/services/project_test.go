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
