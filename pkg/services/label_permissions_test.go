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

func TestLabelService_CanRead(t *testing.T) {
	t.Run("Owner_CanRead", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 1} // Owner of label 1
		ls := NewLabelService(s.Engine())

		canRead, maxRight, err := ls.CanRead(s, 1, u)

		require.NoError(t, err)
		assert.True(t, canRead)
		assert.Equal(t, int(models.PermissionWrite), maxRight) // Owner has write permission
	})

	t.Run("HasAccessViaTask_CanRead", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 1} // Has tasks with label 4
		ls := NewLabelService(s.Engine())

		canRead, maxRight, err := ls.CanRead(s, 4, u)

		require.NoError(t, err)
		assert.True(t, canRead)
		assert.GreaterOrEqual(t, maxRight, int(models.PermissionRead))
	})

	t.Run("NoAccess_CannotRead", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 13} // No access to label 3
		ls := NewLabelService(s.Engine())

		canRead, maxRight, err := ls.CanRead(s, 3, u)

		require.NoError(t, err)
		assert.False(t, canRead)
		assert.Equal(t, 0, maxRight)
	})

	t.Run("NonExistentLabel", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 1}
		ls := NewLabelService(s.Engine())

		canRead, maxRight, err := ls.CanRead(s, 9999, u)

		assert.Error(t, err)
		assert.True(t, models.IsErrLabelDoesNotExist(err))
		assert.False(t, canRead)
		assert.Equal(t, 0, maxRight)
	})
}

func TestLabelService_CanUpdate(t *testing.T) {
	t.Run("Owner_CanUpdate", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 1} // Owner of label 1
		ls := NewLabelService(s.Engine())

		canUpdate, err := ls.CanUpdate(s, 1, u)

		require.NoError(t, err)
		assert.True(t, canUpdate)
	})

	t.Run("NonOwner_CannotUpdate", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 2} // Not owner of label 1
		ls := NewLabelService(s.Engine())

		canUpdate, err := ls.CanUpdate(s, 1, u)

		require.NoError(t, err)
		assert.False(t, canUpdate)
	})

	t.Run("LinkShare_CannotUpdate", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		linkShare := &models.LinkSharing{ID: 1, ProjectID: 1}
		ls := NewLabelService(s.Engine())

		canUpdate, err := ls.CanUpdate(s, 1, linkShare)

		require.NoError(t, err)
		assert.False(t, canUpdate)
	})
}

func TestLabelService_CanDelete(t *testing.T) {
	t.Run("Owner_CanDelete", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 1} // Owner of label 1
		ls := NewLabelService(s.Engine())

		canDelete, err := ls.CanDelete(s, 1, u)

		require.NoError(t, err)
		assert.True(t, canDelete)
	})

	t.Run("NonOwner_CannotDelete", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 2} // Not owner of label 1
		ls := NewLabelService(s.Engine())

		canDelete, err := ls.CanDelete(s, 1, u)

		require.NoError(t, err)
		assert.False(t, canDelete)
	})
}

func TestLabelService_CanCreate(t *testing.T) {
	t.Run("AuthenticatedUser_CanCreate", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 1}
		ls := NewLabelService(s.Engine())
		label := &models.Label{Title: "New Label"}

		canCreate, err := ls.CanCreate(s, label, u)

		require.NoError(t, err)
		assert.True(t, canCreate)
	})

	t.Run("LinkShare_CannotCreate", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		linkShare := &models.LinkSharing{ID: 1, ProjectID: 1}
		ls := NewLabelService(s.Engine())
		label := &models.Label{Title: "New Label"}

		canCreate, err := ls.CanCreate(s, label, linkShare)

		require.NoError(t, err)
		assert.False(t, canCreate)
	})
}

func TestLabelService_CanCreateLabelTask(t *testing.T) {
	t.Run("HasAccessToLabelAndTask_CanCreate", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 1} // Has access to both label 1 and task 1
		ls := NewLabelService(s.Engine())

		canCreate, err := ls.CanCreateLabelTask(s, 1, 1, u)

		require.NoError(t, err)
		assert.True(t, canCreate)
	})

	t.Run("NoAccessToLabel_CannotCreate", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 13} // No access to label 3
		ls := NewLabelService(s.Engine())

		canCreate, err := ls.CanCreateLabelTask(s, 3, 1, u)

		require.NoError(t, err)
		assert.False(t, canCreate)
	})

	t.Run("NoAccessToTask_CannotCreate", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 13} // No access to task 1
		ls := NewLabelService(s.Engine())

		canCreate, err := ls.CanCreateLabelTask(s, 1, 1, u)

		require.NoError(t, err)
		assert.False(t, canCreate)
	})
}

func TestLabelService_CanDeleteLabelTask(t *testing.T) {
	t.Run("CanUpdateTask_CanDelete", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 1} // Can update task 1, and label 4 is associated with task 1
		ls := NewLabelService(s.Engine())

		canDelete, err := ls.CanDeleteLabelTask(s, 4, 1, u)

		require.NoError(t, err)
		assert.True(t, canDelete)
	})

	t.Run("CannotUpdateTask_CannotDelete", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 13} // Cannot update task 1
		ls := NewLabelService(s.Engine())

		canDelete, err := ls.CanDeleteLabelTask(s, 4, 1, u)

		require.NoError(t, err)
		assert.False(t, canDelete)
	})
}
