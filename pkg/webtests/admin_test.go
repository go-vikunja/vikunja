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
	"net/http/httptest"
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/license"
	"code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/user"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func promoteToAdmin(t *testing.T, userID int64) *user.User {
	s := db.NewSession()
	defer s.Close()

	u := &user.User{ID: userID}
	has, err := s.Get(u)
	require.NoError(t, err)
	require.True(t, has)

	u.IsAdmin = true
	_, err = s.ID(u.ID).Cols("is_admin").Update(u)
	require.NoError(t, err)
	require.NoError(t, s.Commit())
	return u
}

func adminReq(t *testing.T, e *echo.Echo, method, path string, u *user.User, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	if u != nil {
		tok, err := auth.NewUserJWTAuthtoken(u, "test-session-id")
		require.NoError(t, err)
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+tok)
	}
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)
	return res
}

func TestAdmin_GateUnlicensed(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	license.ResetForTests()

	admin := promoteToAdmin(t, 1)

	res := adminReq(t, e, http.MethodGet, "/api/v1/admin/overview", admin, "")
	assert.Equal(t, http.StatusNotFound, res.Code)
}

func TestAdmin_GateNonAdmin(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	license.SetForTests([]license.Feature{license.FeatureAdminPanel})
	defer license.ResetForTests()

	s := db.NewSession()
	defer s.Close()
	u, err := user.GetUserByID(s, 1)
	require.NoError(t, err)

	res := adminReq(t, e, http.MethodGet, "/api/v1/admin/overview", u, "")
	assert.Equal(t, http.StatusNotFound, res.Code)
}

func TestAdmin_GateUnauthenticated(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	license.SetForTests([]license.Feature{license.FeatureAdminPanel})
	defer license.ResetForTests()

	// echojwt rejects with 401 before the license/admin gates see the request.
	res := adminReq(t, e, http.MethodGet, "/api/v1/admin/overview", nil, "")
	assert.Equal(t, http.StatusUnauthorized, res.Code)
}

func TestAdmin_Overview(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	license.SetForTests([]license.Feature{license.FeatureAdminPanel})
	defer license.ResetForTests()

	admin := promoteToAdmin(t, 1)
	res := adminReq(t, e, http.MethodGet, "/api/v1/admin/overview", admin, "")
	assert.Equal(t, http.StatusOK, res.Code)
	body := res.Body.String()
	assert.Contains(t, body, `"users"`)
	assert.Contains(t, body, `"projects"`)
	assert.Contains(t, body, `"tasks"`)
	assert.Contains(t, body, `"shares"`)
	assert.Contains(t, body, `"license"`)
	assert.Contains(t, body, `"licensed":true`)
	assert.Contains(t, body, `"features"`)
	assert.Contains(t, body, `"expires_at"`)
	assert.Contains(t, body, `"instance_id"`)
}

func TestAdmin_ListUsers(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	license.SetForTests([]license.Feature{license.FeatureAdminPanel})
	defer license.ResetForTests()

	admin := promoteToAdmin(t, 1)

	t.Run("returns users including hidden is_admin and status fields", func(t *testing.T) {
		res := adminReq(t, e, http.MethodGet, "/api/v1/admin/users", admin, "")
		assert.Equal(t, http.StatusOK, res.Code)
		body := res.Body.String()
		assert.Contains(t, body, `"is_admin"`)
		assert.Contains(t, body, `"status"`)
		assert.Contains(t, body, `"username":"user1"`)
	})

	t.Run("search filters by username", func(t *testing.T) {
		res := adminReq(t, e, http.MethodGet, "/api/v1/admin/users?s=user2", admin, "")
		assert.Equal(t, http.StatusOK, res.Code)
		body := res.Body.String()
		assert.Contains(t, body, `"username":"user2"`)
		// user15 should not be present when searching exactly "user2".
		assert.NotContains(t, body, `"username":"user15"`)
	})
}

func TestAdmin_PatchAdmin(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	license.SetForTests([]license.Feature{license.FeatureAdminPanel})
	defer license.ResetForTests()

	admin := promoteToAdmin(t, 1)

	t.Run("promote a non-admin user", func(t *testing.T) {
		res := adminReq(t, e, http.MethodPatch, "/api/v1/admin/users/2/admin", admin, `{"is_admin":true}`)
		assert.Equal(t, http.StatusOK, res.Code)

		s := db.NewSession()
		defer s.Close()
		u, err := user.GetUserByID(s, 2)
		require.NoError(t, err)
		assert.True(t, u.IsAdmin)
	})

	t.Run("demote when another admin exists is allowed", func(t *testing.T) {
		res := adminReq(t, e, http.MethodPatch, "/api/v1/admin/users/2/admin", admin, `{"is_admin":false}`)
		assert.Equal(t, http.StatusOK, res.Code)

		s := db.NewSession()
		defer s.Close()
		u, err := user.GetUserByID(s, 2)
		require.NoError(t, err)
		assert.False(t, u.IsAdmin)
	})

	t.Run("last-admin guard refuses demotion", func(t *testing.T) {
		res := adminReq(t, e, http.MethodPatch, "/api/v1/admin/users/1/admin", admin, `{"is_admin":false}`)
		assert.Equal(t, http.StatusBadRequest, res.Code)

		s := db.NewSession()
		defer s.Close()
		u, err := user.GetUserByID(s, 1)
		require.NoError(t, err)
		assert.True(t, u.IsAdmin, "last admin must remain admin after refused demotion")
	})

	t.Run("unknown user returns 404", func(t *testing.T) {
		res := adminReq(t, e, http.MethodPatch, "/api/v1/admin/users/9999999/admin", admin, `{"is_admin":true}`)
		assert.Equal(t, http.StatusNotFound, res.Code)
	})

	t.Run("empty body is rejected rather than demoting", func(t *testing.T) {
		// Promote user 2 first so we can detect an accidental silent demotion.
		res := adminReq(t, e, http.MethodPatch, "/api/v1/admin/users/2/admin", admin, `{"is_admin":true}`)
		require.Equal(t, http.StatusOK, res.Code)

		res = adminReq(t, e, http.MethodPatch, "/api/v1/admin/users/2/admin", admin, `{}`)
		assert.Equal(t, http.StatusBadRequest, res.Code)

		s := db.NewSession()
		defer s.Close()
		u, err := user.GetUserByID(s, 2)
		require.NoError(t, err)
		assert.True(t, u.IsAdmin, "empty body must not silently demote")
	})
}

func TestAdmin_ListProjects(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	license.SetForTests([]license.Feature{license.FeatureAdminPanel})
	defer license.ResetForTests()

	admin := promoteToAdmin(t, 1)
	res := adminReq(t, e, http.MethodGet, "/api/v1/admin/projects", admin, "")
	assert.Equal(t, http.StatusOK, res.Code)
	body := res.Body.String()
	assert.Contains(t, body, `"id":`)
	assert.Contains(t, body, `"title":`)
	// Owner is xorm:"-" and must be hydrated explicitly.
	assert.Contains(t, body, `"username":"user1"`)
	assert.NotContains(t, body, `"owner":null`)
}

func TestAdmin_PatchStatus(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	license.SetForTests([]license.Feature{license.FeatureAdminPanel})
	defer license.ResetForTests()

	admin := promoteToAdmin(t, 1)

	res := adminReq(t, e, http.MethodPatch, "/api/v1/admin/users/2/status", admin, `{"status":2}`)
	assert.Equal(t, http.StatusOK, res.Code)

	// GetUserByID refuses disabled accounts, so assert against the raw row.
	s := db.NewSession()
	defer s.Close()
	var row struct {
		Status int `xorm:"status"`
	}
	_, err = s.Table("users").Where("id = ?", 2).Get(&row)
	require.NoError(t, err)
	assert.Equal(t, 2, row.Status)

	t.Run("last-admin guard refuses self-disable", func(t *testing.T) {
		res := adminReq(t, e, http.MethodPatch, "/api/v1/admin/users/1/status", admin, `{"status":2}`)
		assert.Equal(t, http.StatusBadRequest, res.Code)

		var row struct {
			Status int `xorm:"status"`
		}
		_, err := s.Table("users").Where("id = ?", 1).Get(&row)
		require.NoError(t, err)
		assert.Equal(t, int(user.StatusActive), row.Status, "last admin must stay active after refused disable")
	})

	t.Run("last-admin guard refuses self-lock", func(t *testing.T) {
		res := adminReq(t, e, http.MethodPatch, "/api/v1/admin/users/1/status", admin, `{"status":3}`)
		assert.Equal(t, http.StatusBadRequest, res.Code)
	})

	t.Run("last-admin guard refuses email-confirmation status", func(t *testing.T) {
		res := adminReq(t, e, http.MethodPatch, "/api/v1/admin/users/1/status", admin, `{"status":1}`)
		assert.Equal(t, http.StatusBadRequest, res.Code)

		var row struct {
			Status int `xorm:"status"`
		}
		_, err := s.Table("users").Where("id = ?", 1).Get(&row)
		require.NoError(t, err)
		assert.Equal(t, int(user.StatusActive), row.Status, "last admin must stay active when email-confirmation status would be set")
	})

	t.Run("rejects invalid status value", func(t *testing.T) {
		res := adminReq(t, e, http.MethodPatch, "/api/v1/admin/users/2/status", admin, `{"status":99}`)
		assert.Equal(t, http.StatusBadRequest, res.Code)
		assert.Contains(t, res.Body.String(), "invalid status")
	})

	t.Run("empty body is rejected rather than reactivating", func(t *testing.T) {
		// User 2 was disabled earlier in this test; empty body must leave that intact.
		res := adminReq(t, e, http.MethodPatch, "/api/v1/admin/users/2/status", admin, `{}`)
		assert.Equal(t, http.StatusBadRequest, res.Code)

		var row struct {
			Status int `xorm:"status"`
		}
		_, err := s.Table("users").Where("id = ?", 2).Get(&row)
		require.NoError(t, err)
		assert.Equal(t, int(user.StatusDisabled), row.Status, "empty body must not silently reactivate")
	})
}

// Non-active admins must not count toward the last-admin invariant.
func TestAdmin_GuardLastAdmin_IgnoresNonActive(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	license.SetForTests([]license.Feature{license.FeatureAdminPanel})
	defer license.ResetForTests()

	admin := promoteToAdmin(t, 1)

	// A disabled admin cannot log in and must not satisfy the last-admin check.
	s := db.NewSession()
	u17 := &user.User{ID: 17}
	has, err := s.Get(u17)
	require.NoError(t, err)
	require.True(t, has)
	require.Equal(t, user.StatusDisabled, u17.Status, "fixture precondition: user17 is disabled")
	u17.IsAdmin = true
	_, err = s.ID(u17.ID).Cols("is_admin").Update(u17)
	require.NoError(t, err)
	require.NoError(t, s.Commit())
	s.Close()

	res := adminReq(t, e, http.MethodPatch, "/api/v1/admin/users/1/admin", admin, `{"is_admin":false}`)
	assert.Equal(t, http.StatusBadRequest, res.Code)

	s = db.NewSession()
	defer s.Close()
	u, err := user.GetUserByID(s, 1)
	require.NoError(t, err)
	assert.True(t, u.IsAdmin, "active admin must not be demoted when the only other admin is disabled")
}

func TestAdmin_DeleteUser(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	license.SetForTests([]license.Feature{license.FeatureAdminPanel})
	defer license.ResetForTests()

	admin := promoteToAdmin(t, 1)

	t.Run("mode=now deletes a regular user immediately", func(t *testing.T) {
		res := adminReq(t, e, http.MethodDelete, "/api/v1/admin/users/15?mode=now", admin, "")
		assert.Equal(t, http.StatusNoContent, res.Code)

		s := db.NewSession()
		defer s.Close()
		_, err := user.GetUserByID(s, 15)
		assert.Error(t, err, "deleted user must no longer be fetchable")
	})

	t.Run("mode=scheduled triggers RequestDeletion without removing the user", func(t *testing.T) {
		res := adminReq(t, e, http.MethodDelete, "/api/v1/admin/users/16?mode=scheduled", admin, "")
		assert.Equal(t, http.StatusNoContent, res.Code)

		s := db.NewSession()
		defer s.Close()

		// Scheduled deletion only records a token; the row is not removed.
		u := &user.User{ID: 16}
		has, err := s.Get(u)
		require.NoError(t, err)
		assert.True(t, has, "scheduled deletion must not remove the user row")
	})

	t.Run("default (no mode) is scheduled", func(t *testing.T) {
		res := adminReq(t, e, http.MethodDelete, "/api/v1/admin/users/2", admin, "")
		assert.Equal(t, http.StatusNoContent, res.Code)

		s := db.NewSession()
		defer s.Close()
		u := &user.User{ID: 2}
		has, err := s.Get(u)
		require.NoError(t, err)
		assert.True(t, has, "default mode must not remove the user row")
	})

	t.Run("rejects invalid mode", func(t *testing.T) {
		res := adminReq(t, e, http.MethodDelete, "/api/v1/admin/users/3?mode=bogus", admin, "")
		assert.Equal(t, http.StatusBadRequest, res.Code)
	})

	t.Run("mode=now last-admin guard refuses self-delete", func(t *testing.T) {
		res := adminReq(t, e, http.MethodDelete, "/api/v1/admin/users/1?mode=now", admin, "")
		assert.Equal(t, http.StatusBadRequest, res.Code)
	})

	t.Run("unknown user returns 404", func(t *testing.T) {
		res := adminReq(t, e, http.MethodDelete, "/api/v1/admin/users/9999999?mode=now", admin, "")
		assert.Equal(t, http.StatusNotFound, res.Code)
	})
}

func TestAdmin_CreateUser(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	license.SetForTests([]license.Feature{license.FeatureAdminPanel})
	defer license.ResetForTests()

	// Admin endpoint must bypass the public-registration toggle.
	prev := config.ServiceEnableRegistration.GetBool()
	config.ServiceEnableRegistration.Set(false)
	defer config.ServiceEnableRegistration.Set(prev)

	admin := promoteToAdmin(t, 1)

	t.Run("creates a plain user", func(t *testing.T) {
		body := `{"username":"adm-create-1","password":"averyl0ngpassword","email":"adm-create-1@example.com"}`
		res := adminReq(t, e, http.MethodPost, "/api/v1/admin/users", admin, body)
		assert.Equal(t, http.StatusOK, res.Code, res.Body.String())
		assert.Contains(t, res.Body.String(), `"username":"adm-create-1"`)
	})

	t.Run("creates an is_admin user", func(t *testing.T) {
		body := `{"username":"adm-create-2","password":"averyl0ngpassword","email":"adm-create-2@example.com","is_admin":true}`
		res := adminReq(t, e, http.MethodPost, "/api/v1/admin/users", admin, body)
		assert.Equal(t, http.StatusOK, res.Code, res.Body.String())

		s := db.NewSession()
		defer s.Close()
		u, err := user.GetUserByUsername(s, "adm-create-2")
		require.NoError(t, err)
		assert.True(t, u.IsAdmin, "new user should have been promoted")
	})

	t.Run("skip_email_confirm forces Status=Active", func(t *testing.T) {
		body := `{"username":"adm-create-3","password":"averyl0ngpassword","email":"adm-create-3@example.com","skip_email_confirm":true}`
		res := adminReq(t, e, http.MethodPost, "/api/v1/admin/users", admin, body)
		assert.Equal(t, http.StatusOK, res.Code, res.Body.String())

		s := db.NewSession()
		defer s.Close()
		u, err := user.GetUserByUsername(s, "adm-create-3")
		require.NoError(t, err)
		assert.Equal(t, user.StatusActive, u.Status)
	})

	t.Run("persists the name field", func(t *testing.T) {
		body := `{"username":"adm-create-4","password":"averyl0ngpassword","email":"adm-create-4@example.com","name":"Adm Create"}`
		res := adminReq(t, e, http.MethodPost, "/api/v1/admin/users", admin, body)
		assert.Equal(t, http.StatusOK, res.Code, res.Body.String())

		s := db.NewSession()
		defer s.Close()
		u, err := user.GetUserByUsername(s, "adm-create-4")
		require.NoError(t, err)
		assert.Equal(t, "Adm Create", u.Name)
	})

	t.Run("non-admin caller gets 404", func(t *testing.T) {
		s := db.NewSession()
		u2, err := user.GetUserByID(s, 2)
		require.NoError(t, err)
		require.False(t, u2.IsAdmin, "fixture precondition: user2 is not an admin")
		s.Close()

		body := `{"username":"nonadmin-create","password":"averyl0ngpassword","email":"nonadmin-create@example.com"}`
		res := adminReq(t, e, http.MethodPost, "/api/v1/admin/users", u2, body)
		assert.Equal(t, http.StatusNotFound, res.Code)
	})

	t.Run("unauthenticated caller gets 401", func(t *testing.T) {
		body := `{"username":"anon-create","password":"averyl0ngpassword","email":"anon-create@example.com"}`
		res := adminReq(t, e, http.MethodPost, "/api/v1/admin/users", nil, body)
		assert.Equal(t, http.StatusUnauthorized, res.Code)
	})
}

// Without the admin-panel license the endpoint must 404 so unlicensed instances cannot mint admins.
func TestAdmin_CreateUser_LicenseInactive(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	license.ResetForTests()

	admin := promoteToAdmin(t, 1)

	body := `{"username":"unlicensed-create","password":"averyl0ngpassword","email":"unlicensed-create@example.com"}`
	res := adminReq(t, e, http.MethodPost, "/api/v1/admin/users", admin, body)
	assert.Equal(t, http.StatusNotFound, res.Code)
}

func TestAdmin_ReassignProjectOwner(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	license.SetForTests([]license.Feature{license.FeatureAdminPanel})
	defer license.ResetForTests()

	admin := promoteToAdmin(t, 1)

	t.Run("updates owner_id", func(t *testing.T) {
		res := adminReq(t, e, http.MethodPatch, "/api/v1/admin/projects/2/owner", admin, `{"owner_id":2}`)
		assert.Equal(t, http.StatusOK, res.Code)

		s := db.NewSession()
		defer s.Close()
		var row struct {
			OwnerID int64 `xorm:"owner_id"`
		}
		_, err := s.Table("projects").Where("id = ?", 2).Get(&row)
		require.NoError(t, err)
		assert.Equal(t, int64(2), row.OwnerID)
	})

	t.Run("rejects nonexistent owner", func(t *testing.T) {
		res := adminReq(t, e, http.MethodPatch, "/api/v1/admin/projects/2/owner", admin, `{"owner_id":99999}`)
		assert.Equal(t, http.StatusNotFound, res.Code)
	})

	t.Run("nonexistent project returns 404", func(t *testing.T) {
		res := adminReq(t, e, http.MethodPatch, "/api/v1/admin/projects/99999/owner", admin, `{"owner_id":1}`)
		assert.Equal(t, http.StatusNotFound, res.Code)
	})

	t.Run("rejects disabled user as new owner", func(t *testing.T) {
		res := adminReq(t, e, http.MethodPatch, "/api/v1/admin/projects/2/owner", admin, `{"owner_id":17}`)
		assert.Equal(t, http.StatusPreconditionFailed, res.Code)
	})

	t.Run("rejects locked user as new owner", func(t *testing.T) {
		res := adminReq(t, e, http.MethodPatch, "/api/v1/admin/projects/2/owner", admin, `{"owner_id":18}`)
		assert.Equal(t, http.StatusPreconditionFailed, res.Code)
	})

	t.Run("rejects deletion-scheduled user as new owner", func(t *testing.T) {
		// DeleteUser cascades to their projects, so such a reassignment would be destroyed on the delayed delete.
		res := adminReq(t, e, http.MethodPatch, "/api/v1/admin/projects/2/owner", admin, `{"owner_id":20}`)
		assert.Equal(t, http.StatusBadRequest, res.Code)
	})
}

// A demoted admin with a stale JWT claim must lose access immediately.
func TestAdmin_StaleAdminJWT_Gate(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	license.SetForTests([]license.Feature{license.FeatureAdminPanel})
	defer license.ResetForTests()

	admin := promoteToAdmin(t, 1)

	s := db.NewSession()
	_, err = s.ID(int64(1)).Cols("is_admin").Update(&user.User{IsAdmin: false})
	require.NoError(t, err)
	require.NoError(t, s.Commit())
	s.Close()

	res := adminReq(t, e, http.MethodGet, "/api/v1/admin/ping", admin, "")
	assert.Equal(t, http.StatusNotFound, res.Code, "demoted admin with stale JWT must be rejected")
}

func TestAdmin_StaleAdminJWT_DeletedUser(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	license.SetForTests([]license.Feature{license.FeatureAdminPanel})
	defer license.ResetForTests()

	admin := promoteToAdmin(t, 1)

	s := db.NewSession()
	_, err = s.ID(int64(1)).Delete(&user.User{})
	require.NoError(t, err)
	require.NoError(t, s.Commit())
	s.Close()

	res := adminReq(t, e, http.MethodGet, "/api/v1/admin/ping", admin, "")
	assert.Equal(t, http.StatusNotFound, res.Code, "deleted admin with stale JWT must be rejected")
}

// The model-level permission bypass must also re-check the DB, not just the JWT.
func TestAdmin_StaleAdminJWT_PermissionBypass(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	license.SetForTests([]license.Feature{license.FeatureAdminPanel})
	defer license.ResetForTests()

	admin := promoteToAdmin(t, 1)

	res := adminReq(t, e, http.MethodGet, "/api/v1/projects/2", admin, "")
	require.Equal(t, http.StatusOK, res.Code, "fresh admin must be able to read a project they do not own")

	s := db.NewSession()
	_, err = s.ID(int64(1)).Cols("is_admin").Update(&user.User{IsAdmin: false})
	require.NoError(t, err)
	require.NoError(t, s.Commit())
	s.Close()

	res = adminReq(t, e, http.MethodGet, "/api/v1/projects/2", admin, "")
	assert.NotEqual(t, http.StatusOK, res.Code, "demoted admin must lose project bypass after DB update")
}

func TestAdmin_StaleAdminJWT_CreateUser(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	license.SetForTests([]license.Feature{license.FeatureAdminPanel})
	defer license.ResetForTests()

	admin := promoteToAdmin(t, 1)

	s := db.NewSession()
	_, err = s.ID(int64(1)).Cols("is_admin").Update(&user.User{IsAdmin: false})
	require.NoError(t, err)
	require.NoError(t, s.Commit())
	s.Close()

	body := `{"username":"stale-admin","password":"averyl0ngpassword","email":"stale-admin@example.com","is_admin":true}`
	res := adminReq(t, e, http.MethodPost, "/api/v1/admin/users", admin, body)
	assert.Equal(t, http.StatusNotFound, res.Code, "demoted admin with stale JWT must be rejected by the admin gate")
}
