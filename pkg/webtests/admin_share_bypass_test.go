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

package webtests

import (
	"net/http"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/license"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Instance admins manage shares through the existing per-project endpoints via
// the Can* bypass; there are no dedicated admin share endpoints. Project 2
// is owned by user 3; user 1 has no share on it.

func TestAdminBypass_CanListProjectShares(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	defer license.ResetForTests()
	license.SetForTests([]license.Feature{license.FeatureAdminPanel})

	admin := promoteToAdmin(t, 1)
	res := adminReq(t, e, http.MethodGet, "/api/v1/projects/2/shares", admin, "")
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Contains(t, res.Body.String(), `"hash":"test2"`)
}

func TestAdminBypass_CanDeleteLinkShare(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	defer license.ResetForTests()
	license.SetForTests([]license.Feature{license.FeatureAdminPanel})

	admin := promoteToAdmin(t, 1)
	res := adminReq(t, e, http.MethodDelete, "/api/v1/projects/2/shares/2", admin, "")
	assert.Equal(t, http.StatusOK, res.Code)
}

func TestAdminBypass_CanDeleteTeamShare(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	defer license.ResetForTests()
	license.SetForTests([]license.Feature{license.FeatureAdminPanel})

	admin := promoteToAdmin(t, 1)
	// User 1 has only read access to project 3; removing a team share would
	// be forbidden without the instance-admin bypass.
	res := adminReq(t, e, http.MethodDelete, "/api/v1/projects/3/teams/1", admin, "")
	assert.Equal(t, http.StatusOK, res.Code)
}

func TestAdminBypass_CanDeleteUserShare(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	defer license.ResetForTests()
	license.SetForTests([]license.Feature{license.FeatureAdminPanel})

	admin := promoteToAdmin(t, 1)
	// The delete endpoint keys by username, not numeric ID.
	res := adminReq(t, e, http.MethodDelete, "/api/v1/projects/3/users/user2", admin, "")
	assert.Equal(t, http.StatusOK, res.Code)
}

// Regression: the instance-admin short-circuit in Project.CanRead used to swallow
// GetProjectSimpleByID errors, surfacing user-not-found (1005) instead of
// project-not-found (3001) for missing project IDs.
func TestAdminBypass_NonexistentProjectReturns404(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	defer license.ResetForTests()
	license.SetForTests([]license.Feature{license.FeatureAdminPanel})

	admin := promoteToAdmin(t, 1)
	res := adminReq(t, e, http.MethodGet, "/api/v1/projects/99999", admin, "")
	assert.Equal(t, http.StatusNotFound, res.Code)
	body := res.Body.String()
	assert.Contains(t, body, `"code":3001`, "must surface ErrCodeProjectDoesNotExist, not user-not-found")
	assert.NotContains(t, body, `"code":1005`, "must not surface ErrUserDoesNotExist when the project is missing")
}

// The bypass reads is_admin from the DB, so the test must demote in the DB
// rather than flipping the local struct field.
func TestAdminBypass_NonAdminCannotDeleteLinkShare(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	defer license.ResetForTests()

	s := db.NewSession()
	u, err := user.GetUserByID(s, 1)
	require.NoError(t, err)
	require.False(t, u.IsAdmin, "fixture precondition: user1 is not an instance admin")
	s.Close()

	res := adminReq(t, e, http.MethodDelete, "/api/v1/projects/2/shares/2", u, "")
	assert.NotEqual(t, http.StatusOK, res.Code, "non-admin must not be able to delete a share on a project they don't own")
}
