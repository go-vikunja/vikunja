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

	"code.vikunja.io/api/pkg/license"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type proFeatureStateBody struct {
	Feature           string `json:"feature"`
	Licensed          bool   `json:"licensed"`
	PerUserToggleable bool   `json:"per_user_toggleable"`
	DefaultEnabled    bool   `json:"default_enabled"`
	DefaultSource     string `json:"default_source"`
}

type userProFeatureStateBody struct {
	Feature   string `json:"feature"`
	Override  *bool  `json:"override"`
	Effective bool   `json:"effective"`
}

func findProFeature(t *testing.T, states []proFeatureStateBody, feature string) proFeatureStateBody {
	t.Helper()
	for _, st := range states {
		if st.Feature == feature {
			return st
		}
	}
	t.Fatalf("feature %s not in response: %v", feature, states)
	return proFeatureStateBody{}
}

func TestHumaAdminProFeatures(t *testing.T) {
	t.Run("non-admin user gets 404", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.SetForTests([]license.Feature{license.FeatureAdminPanel})
		defer license.ResetForTests()

		res := adminReq(t, e, http.MethodGet, "/api/v2/admin/pro-features", &testuser1, "")
		assert.Equal(t, http.StatusNotFound, res.Code)
	})

	t.Run("list reports license state, toggleability and defaults", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.SetForTests([]license.Feature{license.FeatureAdminPanel, license.FeatureTimeTracking})
		defer license.ResetForTests()

		admin := promoteToAdmin(t, 1)

		res := adminReq(t, e, http.MethodGet, "/api/v2/admin/pro-features", admin, "")
		require.Equal(t, http.StatusOK, res.Code, res.Body.String())

		var states []proFeatureStateBody
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &states))

		tt := findProFeature(t, states, "time_tracking")
		assert.True(t, tt.Licensed)
		assert.True(t, tt.PerUserToggleable)
		assert.True(t, tt.DefaultEnabled)
		assert.Equal(t, "code", tt.DefaultSource)

		ap := findProFeature(t, states, "admin_panel")
		assert.True(t, ap.Licensed)
		assert.False(t, ap.PerUserToggleable)
	})

	t.Run("setting and resetting the instance default", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.SetForTests([]license.Feature{license.FeatureAdminPanel, license.FeatureTimeTracking})
		defer license.ResetForTests()

		admin := promoteToAdmin(t, 1)

		res := adminReq(t, e, http.MethodPut, "/api/v2/admin/pro-features/time_tracking", admin, `{"default_enabled": false}`)
		require.Equal(t, http.StatusOK, res.Code, res.Body.String())

		var states []proFeatureStateBody
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &states))
		tt := findProFeature(t, states, "time_tracking")
		assert.False(t, tt.DefaultEnabled)
		assert.Equal(t, "instance", tt.DefaultSource)

		res = adminReq(t, e, http.MethodDelete, "/api/v2/admin/pro-features/time_tracking", admin, "")
		require.Equal(t, http.StatusNoContent, res.Code, res.Body.String())

		res = adminReq(t, e, http.MethodGet, "/api/v2/admin/pro-features", admin, "")
		require.Equal(t, http.StatusOK, res.Code)
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &states))
		tt = findProFeature(t, states, "time_tracking")
		assert.True(t, tt.DefaultEnabled)
		assert.Equal(t, "code", tt.DefaultSource)
	})

	t.Run("an instance-wide feature rejects a default", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.SetForTests([]license.Feature{license.FeatureAdminPanel})
		defer license.ResetForTests()

		admin := promoteToAdmin(t, 1)

		res := adminReq(t, e, http.MethodPut, "/api/v2/admin/pro-features/admin_panel", admin, `{"default_enabled": false}`)
		assert.Equal(t, http.StatusUnprocessableEntity, res.Code, res.Body.String())
	})

	t.Run("an unknown feature is a 404", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.SetForTests([]license.Feature{license.FeatureAdminPanel})
		defer license.ResetForTests()

		admin := promoteToAdmin(t, 1)

		res := adminReq(t, e, http.MethodPut, "/api/v2/admin/pro-features/nope", admin, `{"default_enabled": false}`)
		assert.Equal(t, http.StatusNotFound, res.Code, res.Body.String())
	})
}

func TestHumaAdminUserProFeatures(t *testing.T) {
	t.Run("set, list and clear an override", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.SetForTests([]license.Feature{license.FeatureAdminPanel, license.FeatureTimeTracking})
		defer license.ResetForTests()

		admin := promoteToAdmin(t, 1)

		// Revoke for user 2.
		res := adminReq(t, e, http.MethodPut, "/api/v2/admin/users/2/pro-features/time_tracking", admin, `{"enabled": false}`)
		require.Equal(t, http.StatusOK, res.Code, res.Body.String())

		var states []userProFeatureStateBody
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &states))
		require.Len(t, states, 1)
		assert.Equal(t, "time_tracking", states[0].Feature)
		require.NotNil(t, states[0].Override)
		assert.False(t, *states[0].Override)
		assert.False(t, states[0].Effective)

		res = adminReq(t, e, http.MethodGet, "/api/v2/admin/users/2/pro-features", admin, "")
		require.Equal(t, http.StatusOK, res.Code, res.Body.String())
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &states))
		require.Len(t, states, 1)
		require.NotNil(t, states[0].Override)
		assert.False(t, *states[0].Override)

		// Clearing the override falls back to the (code) default: enabled.
		res = adminReq(t, e, http.MethodDelete, "/api/v2/admin/users/2/pro-features/time_tracking", admin, "")
		require.Equal(t, http.StatusOK, res.Code, res.Body.String())
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &states))
		require.Len(t, states, 1)
		assert.Nil(t, states[0].Override)
		assert.True(t, states[0].Effective)
	})

	t.Run("an override wins over the instance default", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.SetForTests([]license.Feature{license.FeatureAdminPanel, license.FeatureTimeTracking})
		defer license.ResetForTests()

		admin := promoteToAdmin(t, 1)

		res := adminReq(t, e, http.MethodPut, "/api/v2/admin/pro-features/time_tracking", admin, `{"default_enabled": false}`)
		require.Equal(t, http.StatusOK, res.Code, res.Body.String())

		res = adminReq(t, e, http.MethodPut, "/api/v2/admin/users/2/pro-features/time_tracking", admin, `{"enabled": true}`)
		require.Equal(t, http.StatusOK, res.Code, res.Body.String())

		var states []userProFeatureStateBody
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &states))
		require.Len(t, states, 1)
		assert.True(t, states[0].Effective)
	})

	t.Run("a missing user is a 404", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.SetForTests([]license.Feature{license.FeatureAdminPanel, license.FeatureTimeTracking})
		defer license.ResetForTests()

		admin := promoteToAdmin(t, 1)

		res := adminReq(t, e, http.MethodGet, "/api/v2/admin/users/9999/pro-features", admin, "")
		assert.Equal(t, http.StatusNotFound, res.Code, res.Body.String())
	})

	t.Run("non-admin user gets 404", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.SetForTests([]license.Feature{license.FeatureAdminPanel, license.FeatureTimeTracking})
		defer license.ResetForTests()

		res := adminReq(t, e, http.MethodPut, "/api/v2/admin/users/2/pro-features/time_tracking", &testuser1, `{"enabled": false}`)
		assert.Equal(t, http.StatusNotFound, res.Code)
	})
}

// The per-user gate: a revoked user gets 404s on every time-tracking route
// while other users keep access.
func TestHumaTimeEntry_PerUserGate(t *testing.T) {
	t.Run("revoked user is gated, others are not", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.SetForTests([]license.Feature{license.FeatureAdminPanel, license.FeatureTimeTracking})
		defer license.ResetForTests()

		admin := promoteToAdmin(t, 6)

		res := adminReq(t, e, http.MethodPut, "/api/v2/admin/users/1/pro-features/time_tracking", admin, `{"enabled": false}`)
		require.Equal(t, http.StatusOK, res.Code, res.Body.String())

		rec := humaRequest(t, e, http.MethodGet, "/api/v2/time-entries", "", humaTokenFor(t, &testuser1), "")
		assert.Equal(t, http.StatusNotFound, rec.Code, "revoked user must get a 404")

		rec = humaRequest(t, e, http.MethodGet, "/api/v2/time-entries", "", humaTokenFor(t, &testuser2), "")
		assert.Equal(t, http.StatusOK, rec.Code, "other users keep access: %s", rec.Body.String())
	})

	t.Run("disabled instance default gates everyone but granted users", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.SetForTests([]license.Feature{license.FeatureAdminPanel, license.FeatureTimeTracking})
		defer license.ResetForTests()

		admin := promoteToAdmin(t, 6)

		res := adminReq(t, e, http.MethodPut, "/api/v2/admin/pro-features/time_tracking", admin, `{"default_enabled": false}`)
		require.Equal(t, http.StatusOK, res.Code, res.Body.String())
		res = adminReq(t, e, http.MethodPut, "/api/v2/admin/users/1/pro-features/time_tracking", admin, `{"enabled": true}`)
		require.Equal(t, http.StatusOK, res.Code, res.Body.String())

		rec := humaRequest(t, e, http.MethodGet, "/api/v2/time-entries", "", humaTokenFor(t, &testuser1), "")
		assert.Equal(t, http.StatusOK, rec.Code, "granted user keeps access: %s", rec.Body.String())

		rec = humaRequest(t, e, http.MethodGet, "/api/v2/time-entries", "", humaTokenFor(t, &testuser2), "")
		assert.Equal(t, http.StatusNotFound, rec.Code, "non-granted user must get a 404")
	})
}
