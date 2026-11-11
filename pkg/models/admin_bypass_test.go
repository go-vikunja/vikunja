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
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/license"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

func TestAdminBypass_Project(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	license.SetForTests([]license.Feature{license.FeatureAdminPanel})
	defer license.ResetForTests()
	s := db.NewSession()
	defer s.Close()

	_, err := s.ID(int64(2)).Cols("is_admin").Update(&user.User{IsAdmin: true})
	require.NoError(t, err)

	admin := &user.User{ID: 2, IsAdmin: true}
	p := &Project{ID: 1}

	t.Run("CanRead", func(t *testing.T) {
		can, _, err := p.CanRead(s, admin)
		require.NoError(t, err)
		assert.True(t, can, "admin must be able to read any project")
	})

	t.Run("CanWrite", func(t *testing.T) {
		can, err := p.CanWrite(s, admin)
		require.NoError(t, err)
		assert.True(t, can)
	})

	t.Run("CanUpdate", func(t *testing.T) {
		can, err := p.CanUpdate(s, admin)
		require.NoError(t, err)
		assert.True(t, can)
	})

	t.Run("CanDelete", func(t *testing.T) {
		can, err := p.CanDelete(s, admin)
		require.NoError(t, err)
		assert.True(t, can)
	})
}

// Without the admin-panel license, flipping is_admin must not recover the paid bypass.
func TestAdminBypass_Project_LicenseInactive(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	license.ResetForTests()
	s := db.NewSession()
	defer s.Close()

	_, err := s.ID(int64(2)).Cols("is_admin").Update(&user.User{IsAdmin: true})
	require.NoError(t, err)

	admin := &user.User{ID: 2, IsAdmin: true}
	p := &Project{ID: 1}

	t.Run("CanRead", func(t *testing.T) {
		can, _, err := p.CanRead(s, admin)
		require.NoError(t, err)
		assert.False(t, can, "unlicensed admin must not read another user's project")
	})

	t.Run("CanWrite", func(t *testing.T) {
		can, err := p.CanWrite(s, admin)
		require.NoError(t, err)
		assert.False(t, can)
	})

	t.Run("CanDelete", func(t *testing.T) {
		can, err := p.CanDelete(s, admin)
		require.NoError(t, err)
		assert.False(t, can)
	})
}

// A stale JWT admin claim must not grant the bypass after DB demotion.
func TestAdminBypass_StaleJWT_Demoted(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	license.SetForTests([]license.Feature{license.FeatureAdminPanel})
	defer license.ResetForTests()
	s := db.NewSession()
	defer s.Close()

	stale := &user.User{ID: 3, IsAdmin: true}
	p := &Project{ID: 1}

	t.Run("CanRead", func(t *testing.T) {
		can, _, err := p.CanRead(s, stale)
		require.NoError(t, err)
		assert.False(t, can, "stale admin claim must not grant project read without DB confirmation")
	})

	t.Run("CanWrite", func(t *testing.T) {
		can, err := p.CanWrite(s, stale)
		require.NoError(t, err)
		assert.False(t, can)
	})

	t.Run("CanDelete", func(t *testing.T) {
		can, err := p.CanDelete(s, stale)
		require.NoError(t, err)
		assert.False(t, can)
	})
}

func TestAdminBypass_StaleJWT_DeletedUser(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	license.SetForTests([]license.Feature{license.FeatureAdminPanel})
	defer license.ResetForTests()
	s := db.NewSession()
	defer s.Close()

	stale := &user.User{ID: 99999, IsAdmin: true}
	p := &Project{ID: 1}

	t.Run("CanRead", func(t *testing.T) {
		can, _, err := p.CanRead(s, stale)
		require.NoError(t, err)
		assert.False(t, can, "deleted admin must not grant bypass")
	})
}

func setUpInstanceAdmin(t *testing.T) *user.User {
	license.SetForTests([]license.Feature{license.FeatureAdminPanel})
	t.Cleanup(license.ResetForTests)

	s := db.NewAutocommitSession()
	defer s.Close()
	_, err := s.ID(int64(2)).Cols("is_admin").Update(&user.User{IsAdmin: true})
	require.NoError(t, err)

	return &user.User{
		ID:      2,
		IsAdmin: true,
	}
}

func setUpTrashedProjectAdmin(t *testing.T) *user.User {
	db.LoadAndAssertFixtures(t)
	admin := setUpInstanceAdmin(t)

	s := db.NewAutocommitSession()
	defer s.Close()
	// Project 1 has views and buckets, unlike the trashed fixtures.
	_, err := s.Table("projects").Where("id = ?", 1).Update(map[string]any{"deleted_at": time.Now()})
	require.NoError(t, err)

	return admin
}

// Sub-resource checks a regular project admin passes, built per project so trashed and live run the same set.
func adminProjectWriteChecks(projectID, viewID, bucketID int64) map[string]func(*xorm.Session, web.Auth) (bool, error) {
	view := &ProjectView{
		ID:        viewID,
		ProjectID: projectID,
	}
	projectUser := &ProjectUser{
		ProjectID: projectID,
		Username:  "user1",
	}
	teamProject := &TeamProject{
		ProjectID: projectID,
		TeamID:    1,
	}
	bucket := &Bucket{
		ID:            bucketID,
		ProjectViewID: viewID,
		ProjectID:     projectID,
	}
	return map[string]func(*xorm.Session, web.Auth) (bool, error){
		"ProjectView.CanCreate": view.CanCreate,
		"ProjectView.CanUpdate": view.CanUpdate,
		"ProjectView.CanDelete": view.CanDelete,
		"ProjectUser.CanCreate": projectUser.CanCreate,
		"ProjectUser.CanUpdate": projectUser.CanUpdate,
		"ProjectUser.CanDelete": projectUser.CanDelete,
		"TeamProject.CanCreate": teamProject.CanCreate,
		"TeamProject.CanUpdate": teamProject.CanUpdate,
		"TeamProject.CanDelete": teamProject.CanDelete,
		"Bucket.CanUpdate":      bucket.CanUpdate,
	}
}

func TestAdminBypass_TrashedProject(t *testing.T) {
	admin := setUpTrashedProjectAdmin(t)
	s := db.NewSession()
	defer s.Close()
	notFound := ErrProjectDoesNotExist{ID: 45}

	t.Run("Project.CanWrite", func(t *testing.T) {
		can, err := (&Project{ID: 45}).CanWrite(s, admin)
		require.ErrorIs(t, err, notFound)
		assert.False(t, can)
	})
	t.Run("Project.CanUpdate", func(t *testing.T) {
		can, err := (&Project{ID: 45}).CanUpdate(s, admin)
		require.ErrorIs(t, err, notFound)
		assert.False(t, can)
	})
	t.Run("Project.IsAdmin", func(t *testing.T) {
		can, err := (&Project{ID: 45}).IsAdmin(s, admin)
		require.ErrorIs(t, err, notFound)
		assert.False(t, can)
	})
	t.Run("ProjectView.CanRead", func(t *testing.T) {
		pv := &ProjectView{
			ID:        1,
			ProjectID: 1,
		}
		can, _, err := pv.CanRead(s, admin)
		var inTrash ErrProjectIsInTrash
		require.ErrorAs(t, err, &inTrash)
		assert.Equal(t, int64(1), inTrash.ID)
		assert.False(t, can)
	})
	for name, check := range adminProjectWriteChecks(1, 4, 1) {
		t.Run(name, func(t *testing.T) {
			can, err := check(s, admin)
			require.ErrorIs(t, err, ErrProjectDoesNotExist{ID: 1})
			assert.False(t, can)
		})
	}

	t.Run("live project", func(t *testing.T) {
		t.Run("Project.CanWrite", func(t *testing.T) {
			can, err := (&Project{ID: 2}).CanWrite(s, admin)
			require.NoError(t, err)
			assert.True(t, can)
		})
		t.Run("ProjectView.CanRead", func(t *testing.T) {
			pv := &ProjectView{
				ID:        5,
				ProjectID: 2,
			}
			can, maxPermission, err := pv.CanRead(s, admin)
			require.NoError(t, err)
			assert.True(t, can)
			assert.Equal(t, int(PermissionAdmin), maxPermission)
		})
		for name, check := range adminProjectWriteChecks(2, 8, 4) {
			t.Run(name, func(t *testing.T) {
				can, err := check(s, admin)
				require.NoError(t, err)
				assert.True(t, can)
			})
		}
	})
}
