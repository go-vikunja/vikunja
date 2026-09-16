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
	"slices"
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/license"
	"code.vikunja.io/api/pkg/user"

	"github.com/labstack/echo/v5"
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

		// Ported from v1 TestAdmin_ListProjects (admin_test.go:222-226): the
		// response body must carry project fields and a hydrated owner.
		body := res.Body.String()
		assert.Contains(t, body, `"id":`)
		assert.Contains(t, body, `"title":`)
		// Owner is xorm:"-" and must be hydrated explicitly (project 1 is owned by user1).
		assert.Contains(t, body, `"username":"user1"`)
		assert.NotContains(t, body, `"owner":null`)
	})

	type listedProject struct {
		ID    int64 `json:"id"`
		Owner struct {
			Username string `json:"username"`
		} `json:"owner"`
	}
	list := func(t *testing.T, e *echo.Echo, admin *user.User, query string) []listedProject {
		res := adminReq(t, e, http.MethodGet, "/api/v2/admin/projects?per_page=1000&"+query, admin, "")
		require.Equal(t, http.StatusOK, res.Code, res.Body.String())
		var envelope struct {
			Items []listedProject `json:"items"`
		}
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &envelope))
		return envelope.Items
	}
	listIDs := func(t *testing.T, e *echo.Echo, admin *user.User, query string) []int64 {
		items := list(t, e, admin, query)
		ids := make([]int64, 0, len(items))
		for _, item := range items {
			ids = append(ids, item.ID)
		}
		return ids
	}

	t.Run("filters by title search", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.SetForTests([]license.Feature{license.FeatureAdminPanel})
		defer license.ResetForTests()

		ids := listIDs(t, e, promoteToAdmin(t, 1), "q=Project+37")
		assert.Equal(t, []int64{37}, ids)
	})

	t.Run("filters by owner", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.SetForTests([]license.Feature{license.FeatureAdminPanel})
		defer license.ResetForTests()

		ids := listIDs(t, e, promoteToAdmin(t, 1), "owner_id=16")
		assert.Equal(t, []int64{37}, ids)
	})

	t.Run("excludes inboxes", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.SetForTests([]license.Feature{license.FeatureAdminPanel})
		defer license.ResetForTests()

		admin := promoteToAdmin(t, 1)
		assert.Subset(t, listIDs(t, e, admin, ""), []int64{4, 37})

		// Projects 4 and 37 are default projects of users 2/3 and 16.
		ids := listIDs(t, e, admin, "exclude_inboxes=true")
		assert.NotContains(t, ids, int64(4))
		assert.NotContains(t, ids, int64(37))
		assert.Contains(t, ids, int64(6))
	})

	t.Run("sorts by column", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.SetForTests([]license.Feature{license.FeatureAdminPanel})
		defer license.ResetForTests()

		admin := promoteToAdmin(t, 1)

		ids := listIDs(t, e, admin, "sort_by=id&order_by=asc")
		require.NotEmpty(t, ids)
		assert.True(t, slices.IsSorted(ids), "expected ascending ids, got %v", ids)

		items := list(t, e, admin, "sort_by=owner&order_by=desc")
		require.NotEmpty(t, items)
		usernames := make([]string, 0, len(items))
		for _, item := range items {
			usernames = append(usernames, item.Owner.Username)
		}
		assert.True(t, slices.IsSortedFunc(usernames, func(a, b string) int { return strings.Compare(b, a) }), "expected owners descending, got %v", usernames)
	})

	t.Run("rejects an unknown sort field", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.SetForTests([]license.Feature{license.FeatureAdminPanel})
		defer license.ResetForTests()

		res := adminReq(t, e, http.MethodGet, "/api/v2/admin/projects?sort_by=password", promoteToAdmin(t, 1), "")
		assert.Equal(t, http.StatusUnprocessableEntity, res.Code, res.Body.String())
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
