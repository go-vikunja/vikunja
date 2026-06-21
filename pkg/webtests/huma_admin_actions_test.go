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

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/license"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Gate behaviour (404 on non-admin/unlicensed, 401 unauthenticated) is shared by
// every /api/v2/admin route; covered once here against the overview endpoint.
func TestHumaAdminOverview(t *testing.T) {
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

		res := adminReq(t, e, http.MethodGet, "/api/v2/admin/overview", u, "")
		assert.Equal(t, http.StatusNotFound, res.Code)
	})

	t.Run("admin without the feature gets 404", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.SetForTests([]license.Feature{})
		defer license.ResetForTests()

		admin := promoteToAdmin(t, 1)

		res := adminReq(t, e, http.MethodGet, "/api/v2/admin/overview", admin, "")
		assert.Equal(t, http.StatusNotFound, res.Code)
	})

	t.Run("unauthenticated caller gets 401", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.SetForTests([]license.Feature{license.FeatureAdminPanel})
		defer license.ResetForTests()

		res := adminReq(t, e, http.MethodGet, "/api/v2/admin/overview", nil, "")
		assert.Equal(t, http.StatusUnauthorized, res.Code)
	})

	t.Run("admin with the feature sees the overview", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.SetForTests([]license.Feature{license.FeatureAdminPanel})
		defer license.ResetForTests()

		admin := promoteToAdmin(t, 1)
		res := adminReq(t, e, http.MethodGet, "/api/v2/admin/overview", admin, "")
		require.Equal(t, http.StatusOK, res.Code, res.Body.String())
		body := res.Body.String()
		assert.Contains(t, body, `"users"`)
		assert.Contains(t, body, `"projects"`)
		assert.Contains(t, body, `"tasks"`)
		assert.Contains(t, body, `"shares"`)
		assert.Contains(t, body, `"license"`)
		assert.Contains(t, body, `"licensed":true`)
		assert.Contains(t, body, `"instance_id"`)
	})
}

func TestHumaAdminCreateUser(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	license.SetForTests([]license.Feature{license.FeatureAdminPanel})
	defer license.ResetForTests()

	// Admin endpoint must bypass the public-registration toggle.
	prev := config.ServiceEnableRegistration.GetBool()
	config.ServiceEnableRegistration.Set(false)
	defer config.ServiceEnableRegistration.Set(prev)

	admin := promoteToAdmin(t, 1)

	t.Run("creates a plain user and returns 201", func(t *testing.T) {
		body := `{"username":"v2adm-create-1","password":"averyl0ngpassword","email":"v2adm-create-1@example.com"}`
		res := adminReq(t, e, http.MethodPost, "/api/v2/admin/users", admin, body)
		assert.Equal(t, http.StatusCreated, res.Code, res.Body.String())
		assert.Contains(t, res.Body.String(), `"username":"v2adm-create-1"`)
	})

	t.Run("creates an is_admin user", func(t *testing.T) {
		body := `{"username":"v2adm-create-2","password":"averyl0ngpassword","email":"v2adm-create-2@example.com","is_admin":true}`
		res := adminReq(t, e, http.MethodPost, "/api/v2/admin/users", admin, body)
		require.Equal(t, http.StatusCreated, res.Code, res.Body.String())

		s := db.NewSession()
		defer s.Close()
		u, err := user.GetUserByUsername(s, "v2adm-create-2")
		require.NoError(t, err)
		assert.True(t, u.IsAdmin, "new user should have been promoted")
	})

	t.Run("skip_email_confirm forces Status=Active", func(t *testing.T) {
		body := `{"username":"v2adm-create-3","password":"averyl0ngpassword","email":"v2adm-create-3@example.com","skip_email_confirm":true}`
		res := adminReq(t, e, http.MethodPost, "/api/v2/admin/users", admin, body)
		require.Equal(t, http.StatusCreated, res.Code, res.Body.String())

		s := db.NewSession()
		defer s.Close()
		u, err := user.GetUserByUsername(s, "v2adm-create-3")
		require.NoError(t, err)
		assert.Equal(t, user.StatusActive, u.Status)
	})

	t.Run("persists the name field", func(t *testing.T) {
		body := `{"username":"v2adm-create-4","password":"averyl0ngpassword","email":"v2adm-create-4@example.com","name":"Adm Create"}`
		res := adminReq(t, e, http.MethodPost, "/api/v2/admin/users", admin, body)
		require.Equal(t, http.StatusCreated, res.Code, res.Body.String())

		s := db.NewSession()
		defer s.Close()
		u, err := user.GetUserByUsername(s, "v2adm-create-4")
		require.NoError(t, err)
		assert.Equal(t, "Adm Create", u.Name)
	})

	t.Run("rejects an invalid body with 422", func(t *testing.T) {
		// Password below the 8-char minimum fails govalidator before the create.
		body := `{"username":"v2adm-invalid","password":"short","email":"v2adm-invalid@example.com"}`
		res := adminReq(t, e, http.MethodPost, "/api/v2/admin/users", admin, body)
		assert.Equal(t, http.StatusUnprocessableEntity, res.Code, res.Body.String())
	})

	t.Run("non-admin caller gets 404", func(t *testing.T) {
		s := db.NewSession()
		u2, err := user.GetUserByID(s, 2)
		require.NoError(t, err)
		require.False(t, u2.IsAdmin, "fixture precondition: user2 is not an admin")
		s.Close()

		body := `{"username":"v2nonadmin","password":"averyl0ngpassword","email":"v2nonadmin@example.com"}`
		res := adminReq(t, e, http.MethodPost, "/api/v2/admin/users", u2, body)
		assert.Equal(t, http.StatusNotFound, res.Code)
	})
}

func TestHumaAdminPatchAdmin(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	license.SetForTests([]license.Feature{license.FeatureAdminPanel})
	defer license.ResetForTests()

	admin := promoteToAdmin(t, 1)

	t.Run("promote a non-admin user", func(t *testing.T) {
		res := adminReq(t, e, http.MethodPatch, "/api/v2/admin/users/2/admin", admin, `{"is_admin":true}`)
		assert.Equal(t, http.StatusOK, res.Code, res.Body.String())

		s := db.NewSession()
		defer s.Close()
		u, err := user.GetUserByID(s, 2)
		require.NoError(t, err)
		assert.True(t, u.IsAdmin)
	})

	t.Run("demote when another admin exists is allowed", func(t *testing.T) {
		res := adminReq(t, e, http.MethodPatch, "/api/v2/admin/users/2/admin", admin, `{"is_admin":false}`)
		assert.Equal(t, http.StatusOK, res.Code)

		s := db.NewSession()
		defer s.Close()
		u, err := user.GetUserByID(s, 2)
		require.NoError(t, err)
		assert.False(t, u.IsAdmin)
	})

	t.Run("last-admin guard refuses demotion with 400", func(t *testing.T) {
		res := adminReq(t, e, http.MethodPatch, "/api/v2/admin/users/1/admin", admin, `{"is_admin":false}`)
		assert.Equal(t, http.StatusBadRequest, res.Code)

		s := db.NewSession()
		defer s.Close()
		u, err := user.GetUserByID(s, 1)
		require.NoError(t, err)
		assert.True(t, u.IsAdmin, "last admin must remain admin after refused demotion")
	})

	t.Run("unknown user returns 404", func(t *testing.T) {
		res := adminReq(t, e, http.MethodPatch, "/api/v2/admin/users/9999999/admin", admin, `{"is_admin":true}`)
		assert.Equal(t, http.StatusNotFound, res.Code)
	})

	t.Run("omitted is_admin is rejected rather than demoting", func(t *testing.T) {
		res := adminReq(t, e, http.MethodPatch, "/api/v2/admin/users/2/admin", admin, `{"is_admin":true}`)
		require.Equal(t, http.StatusOK, res.Code)

		res = adminReq(t, e, http.MethodPatch, "/api/v2/admin/users/2/admin", admin, `{}`)
		assert.Equal(t, http.StatusBadRequest, res.Code)

		s := db.NewSession()
		defer s.Close()
		u, err := user.GetUserByID(s, 2)
		require.NoError(t, err)
		assert.True(t, u.IsAdmin, "omitted is_admin must not silently demote")
	})
}

func TestHumaAdminPatchStatus(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	license.SetForTests([]license.Feature{license.FeatureAdminPanel})
	defer license.ResetForTests()

	admin := promoteToAdmin(t, 1)

	res := adminReq(t, e, http.MethodPatch, "/api/v2/admin/users/2/status", admin, `{"status":2}`)
	assert.Equal(t, http.StatusOK, res.Code, res.Body.String())

	// GetUserByID refuses disabled accounts, so assert against the raw row.
	s := db.NewSession()
	defer s.Close()
	var row struct {
		Status int `xorm:"status"`
	}
	_, err = s.Table("users").Where("id = ?", 2).Get(&row)
	require.NoError(t, err)
	assert.Equal(t, 2, row.Status)

	t.Run("last-admin guard refuses self-disable with 400", func(t *testing.T) {
		res := adminReq(t, e, http.MethodPatch, "/api/v2/admin/users/1/status", admin, `{"status":2}`)
		assert.Equal(t, http.StatusBadRequest, res.Code)

		var row struct {
			Status int `xorm:"status"`
		}
		_, err := s.Table("users").Where("id = ?", 1).Get(&row)
		require.NoError(t, err)
		assert.Equal(t, int(user.StatusActive), row.Status, "last admin must stay active after refused disable")
	})

	t.Run("rejects invalid status value with 400", func(t *testing.T) {
		res := adminReq(t, e, http.MethodPatch, "/api/v2/admin/users/2/status", admin, `{"status":99}`)
		assert.Equal(t, http.StatusBadRequest, res.Code)
		assert.Contains(t, res.Body.String(), "invalid status")
	})

	t.Run("omitted status is rejected rather than reactivating", func(t *testing.T) {
		// User 2 was disabled above; an empty body must leave that intact.
		res := adminReq(t, e, http.MethodPatch, "/api/v2/admin/users/2/status", admin, `{}`)
		assert.Equal(t, http.StatusBadRequest, res.Code)

		var row struct {
			Status int `xorm:"status"`
		}
		_, err := s.Table("users").Where("id = ?", 2).Get(&row)
		require.NoError(t, err)
		assert.Equal(t, int(user.StatusDisabled), row.Status, "omitted status must not silently reactivate")
	})
}

func TestHumaAdminDeleteUser(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	license.SetForTests([]license.Feature{license.FeatureAdminPanel})
	defer license.ResetForTests()

	admin := promoteToAdmin(t, 1)

	t.Run("mode=now deletes a regular user immediately with 204", func(t *testing.T) {
		res := adminReq(t, e, http.MethodDelete, "/api/v2/admin/users/15?mode=now", admin, "")
		assert.Equal(t, http.StatusNoContent, res.Code)

		s := db.NewSession()
		defer s.Close()
		_, err := user.GetUserByID(s, 15)
		assert.Error(t, err, "deleted user must no longer be fetchable")
	})

	t.Run("mode=scheduled keeps the user row", func(t *testing.T) {
		res := adminReq(t, e, http.MethodDelete, "/api/v2/admin/users/16?mode=scheduled", admin, "")
		assert.Equal(t, http.StatusNoContent, res.Code)

		s := db.NewSession()
		defer s.Close()
		u := &user.User{ID: 16}
		has, err := s.Get(u)
		require.NoError(t, err)
		assert.True(t, has, "scheduled deletion must not remove the user row")
	})

	t.Run("default (no mode) is scheduled", func(t *testing.T) {
		res := adminReq(t, e, http.MethodDelete, "/api/v2/admin/users/2", admin, "")
		assert.Equal(t, http.StatusNoContent, res.Code)

		s := db.NewSession()
		defer s.Close()
		u := &user.User{ID: 2}
		has, err := s.Get(u)
		require.NoError(t, err)
		assert.True(t, has, "default mode must not remove the user row")
	})

	t.Run("rejects invalid mode with 400", func(t *testing.T) {
		res := adminReq(t, e, http.MethodDelete, "/api/v2/admin/users/3?mode=bogus", admin, "")
		assert.Equal(t, http.StatusBadRequest, res.Code)
	})

	t.Run("mode=now last-admin guard refuses self-delete with 400", func(t *testing.T) {
		res := adminReq(t, e, http.MethodDelete, "/api/v2/admin/users/1?mode=now", admin, "")
		assert.Equal(t, http.StatusBadRequest, res.Code)
	})

	t.Run("unknown user returns 404", func(t *testing.T) {
		res := adminReq(t, e, http.MethodDelete, "/api/v2/admin/users/9999999?mode=now", admin, "")
		assert.Equal(t, http.StatusNotFound, res.Code)
	})
}

func TestHumaAdminReassignProjectOwner(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	license.SetForTests([]license.Feature{license.FeatureAdminPanel})
	defer license.ResetForTests()

	admin := promoteToAdmin(t, 1)

	t.Run("updates owner_id", func(t *testing.T) {
		res := adminReq(t, e, http.MethodPatch, "/api/v2/admin/projects/2/owner", admin, `{"owner_id":2}`)
		assert.Equal(t, http.StatusOK, res.Code, res.Body.String())

		s := db.NewSession()
		defer s.Close()
		var row struct {
			OwnerID int64 `xorm:"owner_id"`
		}
		_, err := s.Table("projects").Where("id = ?", 2).Get(&row)
		require.NoError(t, err)
		assert.Equal(t, int64(2), row.OwnerID)
	})

	t.Run("rejects nonexistent owner with 404", func(t *testing.T) {
		res := adminReq(t, e, http.MethodPatch, "/api/v2/admin/projects/2/owner", admin, `{"owner_id":99999}`)
		assert.Equal(t, http.StatusNotFound, res.Code)
	})

	t.Run("nonexistent project returns 404", func(t *testing.T) {
		res := adminReq(t, e, http.MethodPatch, "/api/v2/admin/projects/99999/owner", admin, `{"owner_id":1}`)
		assert.Equal(t, http.StatusNotFound, res.Code)
	})

	t.Run("rejects disabled user as new owner with 412", func(t *testing.T) {
		res := adminReq(t, e, http.MethodPatch, "/api/v2/admin/projects/2/owner", admin, `{"owner_id":17}`)
		assert.Equal(t, http.StatusPreconditionFailed, res.Code)
	})

	t.Run("rejects locked user as new owner with 412", func(t *testing.T) {
		res := adminReq(t, e, http.MethodPatch, "/api/v2/admin/projects/2/owner", admin, `{"owner_id":18}`)
		assert.Equal(t, http.StatusPreconditionFailed, res.Code)
	})

	t.Run("rejects deletion-scheduled user as new owner with 400", func(t *testing.T) {
		res := adminReq(t, e, http.MethodPatch, "/api/v2/admin/projects/2/owner", admin, `{"owner_id":20}`)
		assert.Equal(t, http.StatusBadRequest, res.Code)
	})
}
