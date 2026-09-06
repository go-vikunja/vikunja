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

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// user16 carries the restricted fixture rows (pkg/db/fixtures/user_entitlements.yml).
var testuser16 = user.User{ID: 16, Username: "user16", Issuer: "local"}

func TestHumaAdminEntitlements(t *testing.T) {
	const path = "/api/v2/admin/users/1/entitlements"

	t.Run("non-admin gets 404", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.SetForTests([]license.Feature{license.FeatureAdminPanel})
		defer license.ResetForTests()

		res := adminReq(t, e, http.MethodGet, path, &testuser2, "")
		assert.Equal(t, http.StatusNotFound, res.Code)
		res = adminReq(t, e, http.MethodPut, path, &testuser2, `{"entitlements":{}}`)
		assert.Equal(t, http.StatusNotFound, res.Code)
	})

	t.Run("admin reads raw rows", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.SetForTests([]license.Feature{license.FeatureAdminPanel})
		defer license.ResetForTests()
		admin := promoteToAdmin(t, 1)

		var out struct {
			Entitlements map[string]int64 `json:"entitlements"`
		}
		res := adminReq(t, e, http.MethodGet, "/api/v2/admin/users/16/entitlements", admin, "")
		require.Equal(t, http.StatusOK, res.Code, res.Body.String())
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &out))
		assert.Equal(t, map[string]int64{"time_tracking": 0, "team_creation": 0, "max_projects": 1, "max_storage_bytes": 1024}, out.Entitlements)

		res = adminReq(t, e, http.MethodGet, path, admin, "")
		require.Equal(t, http.StatusOK, res.Code, res.Body.String())
		out.Entitlements = nil
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &out))
		assert.Empty(t, out.Entitlements)
	})

	t.Run("put replaces the set", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.SetForTests([]license.Feature{license.FeatureAdminPanel})
		defer license.ResetForTests()
		admin := promoteToAdmin(t, 1)

		res := adminReq(t, e, http.MethodPut, "/api/v2/admin/users/16/entitlements", admin, `{"entitlements":{"time_tracking":1,"max_projects":50}}`)
		require.Equal(t, http.StatusOK, res.Code, res.Body.String())

		db.AssertExists(t, "user_entitlements", map[string]interface{}{"user_id": 16, "feature": "time_tracking", "value": 1}, false)
		db.AssertExists(t, "user_entitlements", map[string]interface{}{"user_id": 16, "feature": "max_projects", "value": 50}, false)
		db.AssertMissing(t, "user_entitlements", map[string]interface{}{"user_id": 16, "feature": "team_creation"})
		db.AssertMissing(t, "user_entitlements", map[string]interface{}{"user_id": 16, "feature": "max_storage_bytes"})

		res = adminReq(t, e, http.MethodPut, "/api/v2/admin/users/16/entitlements", admin, `{"entitlements":{}}`)
		require.Equal(t, http.StatusOK, res.Code, res.Body.String())
		db.AssertMissing(t, "user_entitlements", map[string]interface{}{"user_id": 16})
	})

	t.Run("unknown feature is 400", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.SetForTests([]license.Feature{license.FeatureAdminPanel})
		defer license.ResetForTests()
		admin := promoteToAdmin(t, 1)

		res := adminReq(t, e, http.MethodPut, path, admin, `{"entitlements":{"max_unicorns":3}}`)
		assert.Equal(t, http.StatusBadRequest, res.Code, res.Body.String())
		res = adminReq(t, e, http.MethodPut, path, admin, `{"entitlements":{"admin_panel":0}}`)
		assert.Equal(t, http.StatusBadRequest, res.Code, res.Body.String())
	})

	t.Run("unknown user is 404", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.SetForTests([]license.Feature{license.FeatureAdminPanel})
		defer license.ResetForTests()
		admin := promoteToAdmin(t, 1)

		res := adminReq(t, e, http.MethodPut, "/api/v2/admin/users/9999/entitlements", admin, `{"entitlements":{}}`)
		assert.Equal(t, http.StatusNotFound, res.Code, res.Body.String())
	})
}

func TestHumaUser_Entitlements(t *testing.T) {
	type userResp struct {
		Entitlements map[string]int64 `json:"entitlements"`
		Usage        map[string]int64 `json:"usage"`
	}
	get := func(t *testing.T, e *echo.Echo, u *user.User) userResp {
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/user", "", humaTokenFor(t, u), "")
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		var out userResp
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
		return out
	}

	t.Run("unrestricted user on a free instance", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		out := get(t, e, &testuser1)
		assert.Equal(t, map[string]int64{"admin_panel": 0, "audit_logs": 0, "time_tracking": 0, "team_creation": 1}, out.Entitlements)
		assert.Equal(t, map[string]int64{"max_projects": 5, "max_storage_bytes": 200}, out.Usage)
	})

	t.Run("restricted user on a licensed instance", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.SetForTests([]license.Feature{license.FeatureTimeTracking})
		defer license.ResetForTests()

		out := get(t, e, &testuser16)
		assert.Equal(t, map[string]int64{
			"admin_panel": 0, "audit_logs": 0, "time_tracking": 0, "team_creation": 0,
			"max_projects": 1, "max_storage_bytes": 1024,
		}, out.Entitlements)
		assert.Equal(t, map[string]int64{"max_projects": 1, "max_storage_bytes": 0}, out.Usage)

		out = get(t, e, &testuser1)
		assert.Equal(t, int64(1), out.Entitlements["time_tracking"])
	})
}

func TestHumaEntitlement_Gates(t *testing.T) {
	t.Run("time entries 403 for a user whose plan lacks the feature", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.SetForTests([]license.Feature{license.FeatureTimeTracking})
		defer license.ResetForTests()

		rec := humaRequest(t, e, http.MethodGet, "/api/v2/time-entries", "", humaTokenFor(t, &testuser16), "")
		assert.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"code":20002`)

		rec = humaRequest(t, e, http.MethodPost, "/api/v2/time-entries", `{"project_id":37,"start_time":"2024-01-01T10:00:00Z"}`, humaTokenFor(t, &testuser16), "")
		assert.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())
	})

	t.Run("team creation 403", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		rec := humaRequest(t, e, http.MethodPost, "/api/v2/teams", `{"name":"nope"}`, humaTokenFor(t, &testuser16), "")
		assert.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"code":20002`)
	})

	t.Run("project limit 403 with the limit error code", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		rec := humaRequest(t, e, http.MethodPost, "/api/v2/projects", `{"title":"nope"}`, humaTokenFor(t, &testuser16), "")
		assert.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"code":20001`)
		db.AssertMissing(t, "projects", map[string]interface{}{"title": "nope"})
	})

	t.Run("unrestricted user is unaffected", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		rec := humaRequest(t, e, http.MethodPost, "/api/v2/teams", `{"name":"fine"}`, humaTokenFor(t, &testuser1), "")
		assert.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	})
}
