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

package models

import (
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/license"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjectCanWritePseudoProject(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	u := &user.User{ID: 1}

	t.Run("favorites", func(t *testing.T) {
		can, err := (&Project{ID: FavoritesPseudoProjectID}).CanWrite(s, u)
		require.NoError(t, err)
		assert.False(t, can)
	})

	// Even owning the filter must not bypass the write-denied rule for its pseudo project.
	t.Run("saved filter owned by the user", func(t *testing.T) {
		can, err := (&Project{ID: getProjectIDFromSavedFilterID(1)}).CanWrite(s, u)
		require.NoError(t, err)
		assert.False(t, can)
	})

	t.Run("id 0", func(t *testing.T) {
		can, err := (&Project{ID: 0}).CanWrite(s, u)
		require.NoError(t, err)
		assert.False(t, can)
	})

	t.Run("owned project", func(t *testing.T) {
		can, err := (&Project{ID: 1}).CanWrite(s, u)
		require.NoError(t, err)
		assert.True(t, can)
	})

	t.Run("project without access", func(t *testing.T) {
		can, err := (&Project{ID: 20}).CanWrite(s, u)
		require.NoError(t, err)
		assert.False(t, can)
	})
}

func TestProjectCanWritePseudoProjectAsInstanceAdmin(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	license.SetForTests([]license.Feature{license.FeatureAdminPanel})
	defer license.ResetForTests()
	s := db.NewSession()
	defer s.Close()

	_, err := s.ID(int64(2)).Cols("is_admin").Update(&user.User{IsAdmin: true})
	require.NoError(t, err)

	admin := &user.User{ID: 2, IsAdmin: true}

	t.Run("favorites", func(t *testing.T) {
		can, err := (&Project{ID: FavoritesPseudoProjectID}).CanWrite(s, admin)
		require.NoError(t, err)
		assert.False(t, can)
	})

	t.Run("saved filter", func(t *testing.T) {
		can, err := (&Project{ID: getProjectIDFromSavedFilterID(1)}).CanWrite(s, admin)
		require.NoError(t, err)
		assert.False(t, can)
	})
}

func TestIsPseudoProjectID(t *testing.T) {
	assert.True(t, IsPseudoProjectID(FavoritesPseudoProjectID))
	assert.True(t, IsPseudoProjectID(getProjectIDFromSavedFilterID(1)))
	assert.False(t, IsPseudoProjectID(0))
	assert.False(t, IsPseudoProjectID(1))
}

func TestProjectIsAdminPseudoProject(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	u := &user.User{ID: 1}

	t.Run("favorites", func(t *testing.T) {
		can, err := (&Project{ID: FavoritesPseudoProjectID}).IsAdmin(s, u)
		require.NoError(t, err)
		assert.False(t, can)
	})

	t.Run("saved filter owned by the user", func(t *testing.T) {
		can, err := (&Project{ID: getProjectIDFromSavedFilterID(1)}).IsAdmin(s, u)
		require.NoError(t, err)
		assert.False(t, can)
	})

	t.Run("saved filter owned by someone else", func(t *testing.T) {
		can, err := (&Project{ID: getProjectIDFromSavedFilterID(1)}).IsAdmin(s, &user.User{ID: 2})
		require.NoError(t, err)
		assert.False(t, can)
	})

	// 0 is not a pseudo id, it means "no project" and keeps erroring out.
	t.Run("id 0", func(t *testing.T) {
		can, err := (&Project{ID: 0}).IsAdmin(s, u)
		require.ErrorIs(t, err, ErrProjectDoesNotExist{ID: 0})
		assert.False(t, can)
	})

	t.Run("owned project", func(t *testing.T) {
		can, err := (&Project{ID: 1}).IsAdmin(s, u)
		require.NoError(t, err)
		assert.True(t, can)
	})

	t.Run("project without access", func(t *testing.T) {
		can, err := (&Project{ID: 20}).IsAdmin(s, u)
		require.NoError(t, err)
		assert.False(t, can)
	})
}

func TestProjectCanDeletePseudoProject(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	u := &user.User{ID: 1}

	t.Run("favorites", func(t *testing.T) {
		can, err := (&Project{ID: FavoritesPseudoProjectID}).CanDelete(s, u)
		require.NoError(t, err)
		assert.False(t, can)
	})

	t.Run("saved filter owned by the user", func(t *testing.T) {
		can, err := (&Project{ID: getProjectIDFromSavedFilterID(1)}).CanDelete(s, u)
		require.NoError(t, err)
		assert.False(t, can)
	})

	t.Run("saved filter owned by someone else", func(t *testing.T) {
		can, err := (&Project{ID: getProjectIDFromSavedFilterID(1)}).CanDelete(s, &user.User{ID: 2})
		require.NoError(t, err)
		assert.False(t, can)
	})

	t.Run("id 0", func(t *testing.T) {
		can, err := (&Project{ID: 0}).CanDelete(s, u)
		require.ErrorIs(t, err, ErrProjectDoesNotExist{ID: 0})
		assert.False(t, can)
	})

	t.Run("owned project", func(t *testing.T) {
		can, err := (&Project{ID: 1}).CanDelete(s, u)
		require.NoError(t, err)
		assert.True(t, can)
	})

	t.Run("project without access", func(t *testing.T) {
		can, err := (&Project{ID: 20}).CanDelete(s, u)
		require.NoError(t, err)
		assert.False(t, can)
	})
}

func TestProjectCanUpdatePseudoProject(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	u := &user.User{ID: 1}

	t.Run("favorites", func(t *testing.T) {
		can, err := (&Project{ID: FavoritesPseudoProjectID}).CanUpdate(s, u)
		require.NoError(t, err)
		assert.False(t, can)
	})

	t.Run("saved filter owned by the user", func(t *testing.T) {
		can, err := (&Project{ID: getProjectIDFromSavedFilterID(1)}).CanUpdate(s, u)
		require.NoError(t, err)
		assert.True(t, can)
	})

	t.Run("saved filter owned by someone else", func(t *testing.T) {
		can, err := (&Project{ID: getProjectIDFromSavedFilterID(1)}).CanUpdate(s, &user.User{ID: 2})
		require.NoError(t, err)
		assert.False(t, can)
	})

	t.Run("id 0", func(t *testing.T) {
		can, err := (&Project{ID: 0}).CanUpdate(s, u)
		require.ErrorIs(t, err, ErrProjectDoesNotExist{ID: 0})
		assert.False(t, can)
	})

	t.Run("owned project", func(t *testing.T) {
		can, err := (&Project{ID: 1}).CanUpdate(s, u)
		require.NoError(t, err)
		assert.True(t, can)
	})

	t.Run("project without access", func(t *testing.T) {
		can, err := (&Project{ID: 20}).CanUpdate(s, u)
		require.NoError(t, err)
		assert.False(t, can)
	})
}

func TestProjectCanReadPseudoProject(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	u := &user.User{ID: 1}

	t.Run("favorites", func(t *testing.T) {
		can, maxPermission, err := (&Project{ID: FavoritesPseudoProjectID}).CanRead(s, u)
		require.NoError(t, err)
		assert.True(t, can)
		assert.Equal(t, int(PermissionRead), maxPermission)
	})

	t.Run("saved filter owned by the user", func(t *testing.T) {
		can, maxPermission, err := (&Project{ID: getProjectIDFromSavedFilterID(1)}).CanRead(s, u)
		require.NoError(t, err)
		assert.True(t, can)
		assert.Equal(t, int(PermissionAdmin), maxPermission)
	})

	t.Run("saved filter owned by someone else", func(t *testing.T) {
		can, _, err := (&Project{ID: getProjectIDFromSavedFilterID(1)}).CanRead(s, &user.User{ID: 2})
		require.NoError(t, err)
		assert.False(t, can)
	})

	t.Run("id 0", func(t *testing.T) {
		can, _, err := (&Project{ID: 0}).CanRead(s, u)
		require.ErrorIs(t, err, ErrProjectDoesNotExist{ID: 0})
		assert.False(t, can)
	})

	t.Run("owned project", func(t *testing.T) {
		can, _, err := (&Project{ID: 1}).CanRead(s, u)
		require.NoError(t, err)
		assert.True(t, can)
	})

	t.Run("project without access", func(t *testing.T) {
		can, _, err := (&Project{ID: 20}).CanRead(s, u)
		require.NoError(t, err)
		assert.False(t, can)
	})
}

func TestProjectPseudoProjectPermissionsAsInstanceAdmin(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	license.SetForTests([]license.Feature{license.FeatureAdminPanel})
	defer license.ResetForTests()
	s := db.NewSession()
	defer s.Close()

	_, err := s.ID(int64(2)).Cols("is_admin").Update(&user.User{IsAdmin: true})
	require.NoError(t, err)

	admin := &user.User{ID: 2, IsAdmin: true}
	filterProjectID := getProjectIDFromSavedFilterID(1)

	t.Run("IsAdmin", func(t *testing.T) {
		can, err := (&Project{ID: FavoritesPseudoProjectID}).IsAdmin(s, admin)
		require.NoError(t, err)
		assert.False(t, can)

		can, err = (&Project{ID: filterProjectID}).IsAdmin(s, admin)
		require.NoError(t, err)
		assert.False(t, can)

		can, err = (&Project{ID: 1}).IsAdmin(s, admin)
		require.NoError(t, err)
		assert.True(t, can)
	})

	t.Run("CanDelete", func(t *testing.T) {
		can, err := (&Project{ID: FavoritesPseudoProjectID}).CanDelete(s, admin)
		require.NoError(t, err)
		assert.False(t, can)

		can, err = (&Project{ID: filterProjectID}).CanDelete(s, admin)
		require.NoError(t, err)
		assert.False(t, can)

		can, err = (&Project{ID: 1}).CanDelete(s, admin)
		require.NoError(t, err)
		assert.True(t, can)
	})

	// The admin bypass must not skip the filter owner check.
	t.Run("CanUpdate", func(t *testing.T) {
		can, err := (&Project{ID: FavoritesPseudoProjectID}).CanUpdate(s, admin)
		require.NoError(t, err)
		assert.False(t, can)

		can, err = (&Project{ID: filterProjectID}).CanUpdate(s, admin)
		require.NoError(t, err)
		assert.False(t, can)

		can, err = (&Project{ID: 1}).CanUpdate(s, admin)
		require.NoError(t, err)
		assert.True(t, can)
	})

	t.Run("CanRead", func(t *testing.T) {
		can, maxPermission, err := (&Project{ID: FavoritesPseudoProjectID}).CanRead(s, admin)
		require.NoError(t, err)
		assert.True(t, can)
		assert.Equal(t, int(PermissionRead), maxPermission)

		can, _, err = (&Project{ID: filterProjectID}).CanRead(s, admin)
		require.NoError(t, err)
		assert.False(t, can)

		can, _, err = (&Project{ID: 1}).CanRead(s, admin)
		require.NoError(t, err)
		assert.True(t, can)
	})
}

func TestProjectReadAllReportsFavoritesPermission(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	u := &user.User{ID: 1, Username: "user1"}

	all, _, _, err := (&Project{Expand: ProjectExpandableRights}).ReadAll(s, u, "", 1, 500)
	require.NoError(t, err)

	seen := false
	for _, project := range all.([]*Project) {
		if project.ID != FavoritesPseudoProjectID {
			continue
		}

		seen = true
		require.NotNil(t, project.MaxPermission, "favorites pseudo project must report a permission")
		assert.Equal(t, PermissionRead, *project.MaxPermission)
	}
	assert.True(t, seen, "favorites pseudo project should show up in the project list")
}
