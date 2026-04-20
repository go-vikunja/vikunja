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
