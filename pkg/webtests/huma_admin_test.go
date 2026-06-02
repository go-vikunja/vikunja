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
	"encoding/json"
	"net/http"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/license"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The error body shape is covered by TestHuma_ErrorShapeIsRFC9457; this test
// only asserts gate status codes (404 on failure, matching v1).
func TestHumaAdminProjects(t *testing.T) {
	t.Run("non-admin user gets 404", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.SetForTests([]license.Feature{license.FeatureAdminPanel})
		defer license.ResetForTests()

		s := db.NewSession()
		defer s.Close()
		u, err := user.GetUserByID(s, 1)
		require.NoError(t, err)
		require.False(t, u.IsAdmin, "fixture precondition: user1 is not an admin")

		res := adminReq(t, e, http.MethodGet, "/api/v2/admin/projects", u, "")
		assert.Equal(t, http.StatusNotFound, res.Code)
	})

	t.Run("admin without the feature gets 404", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		// Empty feature set = licensed instance without the admin feature.
		license.SetForTests([]license.Feature{})
		defer license.ResetForTests()

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

		var envelope struct {
			Items []struct {
				ID int64 `json:"id"`
			} `json:"items"`
			Total int64 `json:"total"`
		}
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &envelope))

		ids := make(map[int64]bool, len(envelope.Items))
		for _, item := range envelope.Items {
			ids[item.ID] = true
		}
		// Project 6 (owned by user6, not shared with user1) proves the list ignores ownership.
		assert.True(t, ids[6], "expected project 6 in the admin list, got items %v", ids)
		// Project 22 is archived, proving the list includes archived projects.
		assert.True(t, ids[22], "expected archived project 22 in the admin list, got items %v", ids)
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
