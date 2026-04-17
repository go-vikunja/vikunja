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
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdminBypass_Project(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	// User 2 owns nothing involving project 1.
	_, err := s.ID(int64(2)).Cols("is_admin").Update(&user.User{IsAdmin: true})
	require.NoError(t, err)

	admin := &user.User{ID: 2, IsAdmin: true}
	p := &Project{ID: 1}

	can, _, err := p.CanRead(s, admin)
	require.NoError(t, err)
	assert.True(t, can, "admin must be able to read any project")

	can, err = p.CanWrite(s, admin)
	require.NoError(t, err)
	assert.True(t, can)

	can, err = p.CanUpdate(s, admin)
	require.NoError(t, err)
	assert.True(t, can)

	can, err = p.CanDelete(s, admin)
	require.NoError(t, err)
	assert.True(t, can)
}

// A stale admin-claim auth (as if from a JWT minted before demotion) must
// not grant the bypass once the DB row is demoted.
func TestAdminBypass_StaleJWT_Demoted(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	// User 3 is not admin in fixtures; project 1 is owned by user 1.
	stale := &user.User{ID: 3, IsAdmin: true}
	p := &Project{ID: 1}

	can, _, err := p.CanRead(s, stale)
	require.NoError(t, err)
	assert.False(t, can, "stale admin claim must not grant project read without DB confirmation")

	can, err = p.CanWrite(s, stale)
	require.NoError(t, err)
	assert.False(t, can)

	can, err = p.CanDelete(s, stale)
	require.NoError(t, err)
	assert.False(t, can)
}

func TestAdminBypass_StaleJWT_DeletedUser(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	stale := &user.User{ID: 99999, IsAdmin: true}
	p := &Project{ID: 1}

	can, _, err := p.CanRead(s, stale)
	require.NoError(t, err)
	assert.False(t, can, "deleted admin must not grant bypass")
}
