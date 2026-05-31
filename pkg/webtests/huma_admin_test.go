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

// TestHumaAdminProjects exercises the v2 admin gate via GET /api/v2/admin/projects.
// It mirrors v1's TestAdmin_ListProjects but additionally asserts that the same
// two-stage gate (license feature + instance admin) v1 uses on /admin carries
// through to the Huma-backed /api/v2/admin group, returning 404 (not 403) on
// failure. The RFC 9457 error body is asserted once globally in
// TestHuma_ErrorShapeIsRFC9457, so here we only assert the status codes.
func TestHumaAdminProjects(t *testing.T) {
	t.Run("non-admin user gets 404", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.SetForTests([]license.Feature{license.FeatureAdminPanel})
		defer license.ResetForTests()

		s := db.NewSession()
		u, err := user.GetUserByID(s, 1)
		require.NoError(t, err)
		require.False(t, u.IsAdmin, "fixture precondition: user1 is not an admin")
		s.Close()

		res := adminReq(t, e, http.MethodGet, "/api/v2/admin/projects", u, "")
		assert.Equal(t, http.StatusNotFound, res.Code)
	})

	t.Run("admin without the feature gets 404", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.ResetForTests()

		admin := promoteToAdmin(t, 1)

		res := adminReq(t, e, http.MethodGet, "/api/v2/admin/projects", admin, "")
		assert.Equal(t, http.StatusNotFound, res.Code)
	})

	t.Run("admin with the feature sees every project", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.SetForTests([]license.Feature{license.FeatureAdminPanel})
		defer license.ResetForTests()

		admin := promoteToAdmin(t, 1)

		res := adminReq(t, e, http.MethodGet, "/api/v2/admin/projects", admin, "")
		require.Equal(t, http.StatusOK, res.Code, res.Body.String())
		body := res.Body.String()
		// v2 wraps lists in the Paginated envelope.
		assert.Contains(t, body, `"items":`)
		assert.Contains(t, body, `"total":`)
		// Project 6 is owned by user6, not shared with user1 — the admin list
		// surfaces it regardless of ownership.
		assert.Contains(t, body, `"id":6`)
		// Project 22 is archived; the admin list includes archived projects.
		assert.Contains(t, body, `"id":22`)
	})

	t.Run("unauthenticated caller gets 401", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.SetForTests([]license.Feature{license.FeatureAdminPanel})
		defer license.ResetForTests()

		// The token middleware rejects with 401 before the gate runs, matching v1.
		res := adminReq(t, e, http.MethodGet, "/api/v2/admin/projects", nil, "")
		assert.Equal(t, http.StatusUnauthorized, res.Code)
	})
}
