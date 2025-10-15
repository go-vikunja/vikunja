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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjectViewService_GetByIDAndProject(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	pvs := NewProjectViewService(testEngine)

	t.Run("Success", func(t *testing.T) {
		view, err := pvs.GetByIDAndProject(s, 1, 1)
		require.NoError(t, err)
		assert.NotNil(t, view)
		assert.Equal(t, int64(1), view.ID)
		assert.Equal(t, int64(1), view.ProjectID)
	})

	t.Run("NotFound_WrongProject", func(t *testing.T) {
		view, err := pvs.GetByIDAndProject(s, 1, 999)
		assert.Error(t, err)
		assert.True(t, models.IsErrProjectViewDoesNotExist(err))
		assert.Nil(t, view)
	})

	t.Run("NotFound_WrongView", func(t *testing.T) {
		view, err := pvs.GetByIDAndProject(s, 9999, 1)
		assert.Error(t, err)
		assert.True(t, models.IsErrProjectViewDoesNotExist(err))
		assert.Nil(t, view)
	})
}

func TestProjectViewService_GetByID(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	pvs := NewProjectViewService(testEngine)

	t.Run("Success", func(t *testing.T) {
		view, err := pvs.GetByID(s, 1)
		require.NoError(t, err)
		assert.NotNil(t, view)
		assert.Equal(t, int64(1), view.ID)
	})

	t.Run("NotFound", func(t *testing.T) {
		view, err := pvs.GetByID(s, 9999)
		assert.Error(t, err)
		assert.True(t, models.IsErrProjectViewDoesNotExist(err))
		assert.Nil(t, view)
	})
}
