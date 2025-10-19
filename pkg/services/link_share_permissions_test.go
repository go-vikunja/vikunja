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

func TestLinkShareService_CanRead(t *testing.T) {
	t.Run("CanReadProject_CanReadLinkShare", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 1} // Owner of project 1
		lss := NewLinkShareService(s.Engine())
		share := &models.LinkSharing{
			ID:        1,
			Hash:      "test",
			ProjectID: 1,
		}

		canRead, maxRight, err := lss.CanRead(s, share, u)

		require.NoError(t, err)
		assert.True(t, canRead)
		assert.GreaterOrEqual(t, maxRight, int(models.PermissionRead))
	})

	t.Run("LinkShare_CannotRead", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		linkShare := &models.LinkSharing{ID: 1, ProjectID: 1}
		lss := NewLinkShareService(s.Engine())
		share := &models.LinkSharing{
			ID:        2,
			Hash:      "test2",
			ProjectID: 1,
		}

		canRead, maxRight, err := lss.CanRead(s, share, linkShare)

		require.NoError(t, err)
		assert.False(t, canRead)
		assert.Equal(t, 0, maxRight)
	})

	t.Run("NoAccessToProject_CannotRead", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 13} // No access to project 1
		lss := NewLinkShareService(s.Engine())
		share := &models.LinkSharing{
			ID:        1,
			Hash:      "test",
			ProjectID: 1,
		}

		canRead, maxRight, err := lss.CanRead(s, share, u)

		require.NoError(t, err)
		assert.False(t, canRead)
		assert.Equal(t, 0, maxRight)
	})
}

func TestLinkShareService_CanCreate(t *testing.T) {
	t.Run("CanWriteToProject_CanCreate", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 1} // Owner of project 1
		lss := NewLinkShareService(s.Engine())
		share := &models.LinkSharing{
			ProjectID:  1,
			Permission: models.PermissionRead,
		}

		canCreate, err := lss.CanCreate(s, share, u)

		require.NoError(t, err)
		assert.True(t, canCreate)
	})

	t.Run("AdminShare_RequiresAdminPermission", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 1} // Owner/admin of project 1
		lss := NewLinkShareService(s.Engine())
		share := &models.LinkSharing{
			ProjectID:  1,
			Permission: models.PermissionAdmin,
		}

		canCreate, err := lss.CanCreate(s, share, u)

		require.NoError(t, err)
		assert.True(t, canCreate)
	})

	t.Run("LinkShare_CannotCreate", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		linkShare := &models.LinkSharing{ID: 1, ProjectID: 1}
		lss := NewLinkShareService(s.Engine())
		share := &models.LinkSharing{
			ProjectID:  1,
			Permission: models.PermissionRead,
		}

		canCreate, err := lss.CanCreate(s, share, linkShare)

		require.NoError(t, err)
		assert.False(t, canCreate)
	})

	t.Run("NoWriteAccess_CannotCreate", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 13} // No access to project 1
		lss := NewLinkShareService(s.Engine())
		share := &models.LinkSharing{
			ProjectID:  1,
			Permission: models.PermissionRead,
		}

		canCreate, err := lss.CanCreate(s, share, u)

		require.NoError(t, err)
		assert.False(t, canCreate)
	})
}

func TestLinkShareService_CanUpdate(t *testing.T) {
	t.Run("CanWriteToProject_CanUpdate", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 1} // Owner of project 1
		lss := NewLinkShareService(s.Engine())
		share := &models.LinkSharing{
			ID:         1,
			ProjectID:  1,
			Permission: models.PermissionRead,
		}

		canUpdate, err := lss.CanUpdate(s, share, u)

		require.NoError(t, err)
		assert.True(t, canUpdate)
	})

	t.Run("LinkShare_CannotUpdate", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		linkShare := &models.LinkSharing{ID: 1, ProjectID: 1}
		lss := NewLinkShareService(s.Engine())
		share := &models.LinkSharing{
			ID:         2,
			ProjectID:  1,
			Permission: models.PermissionRead,
		}

		canUpdate, err := lss.CanUpdate(s, share, linkShare)

		require.NoError(t, err)
		assert.False(t, canUpdate)
	})
}

func TestLinkShareService_CanDelete(t *testing.T) {
	t.Run("CanWriteToProject_CanDelete", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 1} // Owner of project 1
		lss := NewLinkShareService(s.Engine())
		share := &models.LinkSharing{
			ID:        1,
			ProjectID: 1,
		}

		canDelete, err := lss.CanDelete(s, share, u)

		require.NoError(t, err)
		assert.True(t, canDelete)
	})

	t.Run("LinkShare_CannotDelete", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		linkShare := &models.LinkSharing{ID: 1, ProjectID: 1}
		lss := NewLinkShareService(s.Engine())
		share := &models.LinkSharing{
			ID:        2,
			ProjectID: 1,
		}

		canDelete, err := lss.CanDelete(s, share, linkShare)

		require.NoError(t, err)
		assert.False(t, canDelete)
	})
}
