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

	// Use a real fixture user (user 2 owns nothing involving project 1) and
	// promote them in the DB — the bypass now reads is_admin from the DB.
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

// TestAdminBypass_StaleJWT_Demoted covers the core of bug_003: an auth
// object carrying IsAdmin=true (as if from a freshly-minted JWT) must not
// grant the bypass once the underlying user has been demoted in the DB.
func TestAdminBypass_StaleJWT_Demoted(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	// user 3 is not an admin in fixtures. Construct an auth object that claims
	// to be admin (as a stale JWT would) but does not match the DB.
	stale := &user.User{ID: 3, IsAdmin: true}

	// Project 1 is owned by user 1 — user 3 has no share on it.
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

// TestAdminBypass_StaleJWT_DeletedUser covers the case where a stale admin
// JWT belongs to a user that has been removed from the DB entirely.
func TestAdminBypass_StaleJWT_DeletedUser(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	// No user with this ID exists in fixtures.
	stale := &user.User{ID: 99999, IsAdmin: true}
	p := &Project{ID: 1}

	can, _, err := p.CanRead(s, stale)
	require.NoError(t, err)
	assert.False(t, can, "deleted admin must not grant bypass")
}
