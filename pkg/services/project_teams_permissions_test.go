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

func TestProjectTeamService_CanCreate(t *testing.T) {
	t.Run("project admin can create team relation", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		pts := NewProjectTeamService(db.GetEngine())
		u := &user.User{ID: 1}

		canCreate, err := pts.CanCreate(s, 1, u)
		require.NoError(t, err)
		assert.True(t, canCreate)
	})

	t.Run("non-admin cannot create team relation", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		pts := NewProjectTeamService(db.GetEngine())
		u := &user.User{ID: 2}

		canCreate, err := pts.CanCreate(s, 1, u)
		require.NoError(t, err)
		assert.False(t, canCreate)
	})

	t.Run("link share cannot create team relation", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		pts := NewProjectTeamService(db.GetEngine())
		linkShare := &models.LinkSharing{ID: 1, ProjectID: 1, Permission: models.PermissionAdmin}

		canCreate, err := pts.CanCreate(s, 1, linkShare)
		require.NoError(t, err)
		assert.False(t, canCreate)
	})
}

func TestProjectTeamService_CanUpdate(t *testing.T) {
	t.Run("project admin can update team relation", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		pts := NewProjectTeamService(db.GetEngine())
		u := &user.User{ID: 1}

		canUpdate, err := pts.CanUpdate(s, 1, u)
		require.NoError(t, err)
		assert.True(t, canUpdate)
	})

	t.Run("non-admin cannot update team relation", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		pts := NewProjectTeamService(db.GetEngine())
		u := &user.User{ID: 2}

		canUpdate, err := pts.CanUpdate(s, 1, u)
		require.NoError(t, err)
		assert.False(t, canUpdate)
	})
}

func TestProjectTeamService_CanDelete(t *testing.T) {
	t.Run("project admin can delete team relation", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		pts := NewProjectTeamService(db.GetEngine())
		u := &user.User{ID: 1}

		canDelete, err := pts.CanDelete(s, 1, u)
		require.NoError(t, err)
		assert.True(t, canDelete)
	})

	t.Run("non-admin cannot delete team relation", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		pts := NewProjectTeamService(db.GetEngine())
		u := &user.User{ID: 2}

		canDelete, err := pts.CanDelete(s, 1, u)
		require.NoError(t, err)
		assert.False(t, canDelete)
	})
}

func TestProjectTeamService_CanRead(t *testing.T) {
	t.Run("user with read access can read team relations", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		pts := NewProjectTeamService(db.GetEngine())
		u := &user.User{ID: 1}

		canRead, err := pts.CanRead(s, 1, u)
		require.NoError(t, err)
		assert.True(t, canRead)
	})

	t.Run("user without access cannot read team relations", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		pts := NewProjectTeamService(db.GetEngine())
		u := &user.User{ID: 13} // User with no access to project 1

		canRead, err := pts.CanRead(s, 1, u)
		require.NoError(t, err)
		assert.False(t, canRead)
	})
}
