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
	"code.vikunja.io/api/pkg/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjectViewService_CanRead(t *testing.T) {
	t.Run("user with read access can read project views", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		pvs := NewProjectViewService(db.GetEngine())
		u := &user.User{ID: 1}

		canRead, maxRight, err := pvs.CanRead(s, 1, u)
		require.NoError(t, err)
		assert.True(t, canRead)
		assert.Greater(t, maxRight, 0)
	})

	t.Run("user without access cannot read project views", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		pvs := NewProjectViewService(db.GetEngine())
		u := &user.User{ID: 13} // User with no access to project 1

		canRead, _, err := pvs.CanRead(s, 1, u)
		require.NoError(t, err)
		assert.False(t, canRead)
	})
}

func TestProjectViewService_CanCreate(t *testing.T) {
	t.Run("project admin can create views", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		pvs := NewProjectViewService(db.GetEngine())
		u := &user.User{ID: 1}

		canCreate, err := pvs.CanCreate(s, 1, u)
		require.NoError(t, err)
		assert.True(t, canCreate)
	})

	t.Run("non-admin cannot create views", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		pvs := NewProjectViewService(db.GetEngine())
		u := &user.User{ID: 2}

		canCreate, err := pvs.CanCreate(s, 1, u)
		require.NoError(t, err)
		assert.False(t, canCreate)
	})
}

func TestProjectViewService_CanUpdate(t *testing.T) {
	t.Run("project admin can update views", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		pvs := NewProjectViewService(db.GetEngine())
		u := &user.User{ID: 1}

		canUpdate, err := pvs.CanUpdate(s, 1, u)
		require.NoError(t, err)
		assert.True(t, canUpdate)
	})

	t.Run("non-admin cannot update views", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		pvs := NewProjectViewService(db.GetEngine())
		u := &user.User{ID: 2}

		canUpdate, err := pvs.CanUpdate(s, 1, u)
		require.NoError(t, err)
		assert.False(t, canUpdate)
	})
}

func TestProjectViewService_CanDelete(t *testing.T) {
	t.Run("project admin can delete views", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		pvs := NewProjectViewService(db.GetEngine())
		u := &user.User{ID: 1}

		canDelete, err := pvs.CanDelete(s, 1, u)
		require.NoError(t, err)
		assert.True(t, canDelete)
	})

	t.Run("non-admin cannot delete views", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		pvs := NewProjectViewService(db.GetEngine())
		u := &user.User{ID: 2}

		canDelete, err := pvs.CanDelete(s, 1, u)
		require.NoError(t, err)
		assert.False(t, canDelete)
	})
}
