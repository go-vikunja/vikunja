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

	"code.vikunja.io/api/pkg/license"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests pin down Task 15 from the admin-panel plan: site admins use
// existing per-project share endpoints via the Can* bypass — there are no
// dedicated admin share endpoints. Project 2 is owned by user 3; user 1 has
// no read/write/admin access to it without the site-admin flag.

func TestAdminBypass_CanListProjectShares(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	defer license.ResetForTests()
	license.SetForTests([]license.Feature{license.FeatureAdminPanel})

	admin := promoteToAdmin(t, 1)
	res := adminReq(t, e, http.MethodGet, "/api/v1/projects/2/shares", admin, "")
	assert.Equal(t, http.StatusOK, res.Code)
	// Fixture link share id=2 belongs to project 2.
	assert.Contains(t, res.Body.String(), `"hash":"test2"`)
}

func TestAdminBypass_CanDeleteLinkShare(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	defer license.ResetForTests()
	license.SetForTests([]license.Feature{license.FeatureAdminPanel})

	admin := promoteToAdmin(t, 1)
	// Link share id=2 on project 2 (owned by user 3).
	res := adminReq(t, e, http.MethodDelete, "/api/v1/projects/2/shares/2", admin, "")
	assert.Equal(t, http.StatusOK, res.Code)
}

func TestAdminBypass_CanDeleteTeamShare(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	defer license.ResetForTests()
	license.SetForTests([]license.Feature{license.FeatureAdminPanel})

	admin := promoteToAdmin(t, 1)
	// team_projects id=1: team 1 shared on project 3. User 1 has only read
	// access to project 3 (users_projects id=1 permission=0), so removing a
	// team share would be forbidden without the site-admin bypass.
	res := adminReq(t, e, http.MethodDelete, "/api/v1/projects/3/teams/1", admin, "")
	assert.Equal(t, http.StatusOK, res.Code)
}

func TestAdminBypass_CanDeleteUserShare(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	defer license.ResetForTests()
	license.SetForTests([]license.Feature{license.FeatureAdminPanel})

	admin := promoteToAdmin(t, 1)
	// users_projects id=2: user2 shared on project 3 with read permission.
	// The delete endpoint keys by username, not numeric ID.
	res := adminReq(t, e, http.MethodDelete, "/api/v1/projects/3/users/user2", admin, "")
	assert.Equal(t, http.StatusOK, res.Code)
}

// Negative: without the admin flag, the same user hits a permission error,
// proving the bypass is what makes the positive tests above pass.
func TestAdminBypass_NonAdminCannotDeleteLinkShare(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	defer license.ResetForTests()

	nonAdmin := promoteToAdmin(t, 1)
	nonAdmin.IsAdmin = false // issue token without the admin claim

	res := adminReq(t, e, http.MethodDelete, "/api/v1/projects/2/shares/2", nonAdmin, "")
	assert.NotEqual(t, http.StatusOK, res.Code, "non-admin must not be able to delete a share on a project they don't own")
}
