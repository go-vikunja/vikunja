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

// promoteToAdmin flips is_admin on the given user in the DB so a freshly-issued
// JWT carries the claim. Webtests use this to simulate "CLI set-admin was run".
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

	// 404 — feature not enabled. Gate must look like the route doesn't exist.
	res := adminReq(t, e, http.MethodGet, "/api/v1/admin/ping", admin, "")
	assert.Equal(t, http.StatusNotFound, res.Code)
}

func TestAdmin_GateNonAdmin(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	license.SetForTests([]license.Feature{license.FeatureAdminPanel})
	defer license.ResetForTests()

	// user1 is not admin in fixtures.
	s := db.NewSession()
	defer s.Close()
	u, err := user.GetUserByID(s, 1)
	require.NoError(t, err)

	res := adminReq(t, e, http.MethodGet, "/api/v1/admin/ping", u, "")
	assert.Equal(t, http.StatusNotFound, res.Code)
}

func TestAdmin_GateUnauthenticated(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	license.SetForTests([]license.Feature{license.FeatureAdminPanel})
	defer license.ResetForTests()

	// No token at all — the echojwt middleware rejects with 401 before the
	// license/admin gates ever see the request.
	res := adminReq(t, e, http.MethodGet, "/api/v1/admin/ping", nil, "")
	assert.Equal(t, http.StatusUnauthorized, res.Code)
}

func TestAdmin_PingSucceedsWhenLicensedAndAdmin(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	license.SetForTests([]license.Feature{license.FeatureAdminPanel})
	defer license.ResetForTests()

	admin := promoteToAdmin(t, 1)
	res := adminReq(t, e, http.MethodGet, "/api/v1/admin/ping", admin, "")
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Contains(t, res.Body.String(), `"status":"ok"`)
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
	assert.Contains(t, body, `"version"`)
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
		// user1 and user2 are both admin at this point.
		res := adminReq(t, e, http.MethodPatch, "/api/v1/admin/users/2/admin", admin, `{"is_admin":false}`)
		assert.Equal(t, http.StatusOK, res.Code)

		s := db.NewSession()
		defer s.Close()
		u, err := user.GetUserByID(s, 2)
		require.NoError(t, err)
		assert.False(t, u.IsAdmin)
	})

	t.Run("last-admin guard refuses demotion", func(t *testing.T) {
		// Only user1 is admin now.
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
}

func TestAdmin_ListProjects(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	license.SetForTests([]license.Feature{license.FeatureAdminPanel})
	defer license.ResetForTests()

	admin := promoteToAdmin(t, 1)
	res := adminReq(t, e, http.MethodGet, "/api/v1/admin/projects", admin, "")
	assert.Equal(t, http.StatusOK, res.Code)
	// Fixture projects exist across many users — admin sees them all regardless of ownership.
	body := res.Body.String()
	assert.Contains(t, body, `"id":`)
	assert.Contains(t, body, `"title":`)
}

func TestAdmin_PatchStatus(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	license.SetForTests([]license.Feature{license.FeatureAdminPanel})
	defer license.ResetForTests()

	admin := promoteToAdmin(t, 1)

	// user.Status: 0=Active, 2=Disabled
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
		// user1 is still the only admin — disabling them would lock the instance
		// out once the JWT expires.
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
}

// TestAdmin_GuardLastAdmin_IgnoresNonActive verifies that non-active admins
// (disabled, locked, or scheduled for deletion) are not counted as reachable
// admins — demoting the only active admin must fail even if another admin
// row exists with status != StatusActive.
func TestAdmin_GuardLastAdmin_IgnoresNonActive(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	license.SetForTests([]license.Feature{license.FeatureAdminPanel})
	defer license.ResetForTests()

	admin := promoteToAdmin(t, 1)

	// Promote user17 (status=2, disabled per fixtures) to admin. They cannot
	// log in, so they must not count toward the last-admin invariant.
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

	// Demoting user1 would leave only the disabled user17 as admin — refuse.
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

	t.Run("deletes a regular user", func(t *testing.T) {
		res := adminReq(t, e, http.MethodDelete, "/api/v1/admin/users/15", admin, "")
		assert.Equal(t, http.StatusNoContent, res.Code)

		s := db.NewSession()
		defer s.Close()
		_, err := user.GetUserByID(s, 15)
		assert.Error(t, err, "deleted user must no longer be fetchable")
	})

	t.Run("last-admin guard refuses self-delete", func(t *testing.T) {
		// user 1 is currently the only admin (no one else was promoted in this test).
		res := adminReq(t, e, http.MethodDelete, "/api/v1/admin/users/1", admin, "")
		assert.Equal(t, http.StatusBadRequest, res.Code)
	})

	t.Run("unknown user returns 404", func(t *testing.T) {
		res := adminReq(t, e, http.MethodDelete, "/api/v1/admin/users/9999999", admin, "")
		assert.Equal(t, http.StatusNotFound, res.Code)
	})
}

// TestAdmin_RegisterBypass covers the admin bypass on the public /register
// endpoint: a site admin's bearer token lets them create users even when
// ServiceEnableRegistration is false, and carries the admin-only is_admin /
// skip_email_confirm fields. Non-admin callers keep the original behavior.
func TestAdmin_RegisterBypass(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	license.SetForTests([]license.Feature{license.FeatureAdminPanel})
	defer license.ResetForTests()

	// Turn off public registration to prove the bypass works. The default is true,
	// so remember to restore it for any subsequent test run.
	prev := config.ServiceEnableRegistration.GetBool()
	config.ServiceEnableRegistration.Set(false)
	defer config.ServiceEnableRegistration.Set(prev)

	admin := promoteToAdmin(t, 1)

	s := db.NewSession()
	u2, err := user.GetUserByID(s, 2)
	require.NoError(t, err)
	require.False(t, u2.IsAdmin, "user2 must be a non-admin for these tests")
	s.Close()

	t.Run("admin bearer bypasses ServiceEnableRegistration=false", func(t *testing.T) {
		body := `{"username":"byp-admin-1","password":"averyl0ngpassword","email":"byp-admin-1@example.com"}`
		res := adminReq(t, e, http.MethodPost, "/api/v1/register", admin, body)
		assert.Equal(t, http.StatusOK, res.Code, res.Body.String())
		assert.Contains(t, res.Body.String(), `"username":"byp-admin-1"`)
	})

	t.Run("admin bearer creates is_admin user", func(t *testing.T) {
		body := `{"username":"byp-admin-2","password":"averyl0ngpassword","email":"byp-admin-2@example.com","is_admin":true}`
		res := adminReq(t, e, http.MethodPost, "/api/v1/register", admin, body)
		assert.Equal(t, http.StatusOK, res.Code, res.Body.String())

		s := db.NewSession()
		defer s.Close()
		u, err := user.GetUserByUsername(s, "byp-admin-2")
		require.NoError(t, err)
		assert.True(t, u.IsAdmin, "new user should have been promoted")
	})

	t.Run("admin bearer skip_email_confirm forces Status=Active", func(t *testing.T) {
		body := `{"username":"byp-admin-3","password":"averyl0ngpassword","email":"byp-admin-3@example.com","skip_email_confirm":true}`
		res := adminReq(t, e, http.MethodPost, "/api/v1/register", admin, body)
		assert.Equal(t, http.StatusOK, res.Code, res.Body.String())

		s := db.NewSession()
		defer s.Close()
		u, err := user.GetUserByUsername(s, "byp-admin-3")
		require.NoError(t, err)
		assert.Equal(t, user.StatusActive, u.Status)
	})

	t.Run("no bearer with registration disabled returns 404", func(t *testing.T) {
		body := `{"username":"nobearer","password":"averyl0ngpassword","email":"nobearer@example.com"}`
		res := adminReq(t, e, http.MethodPost, "/api/v1/register", nil, body)
		assert.Equal(t, http.StatusNotFound, res.Code)
	})

	t.Run("non-admin bearer with registration disabled returns 404", func(t *testing.T) {
		body := `{"username":"nonadminbearer","password":"averyl0ngpassword","email":"nonadminbearer@example.com"}`
		res := adminReq(t, e, http.MethodPost, "/api/v1/register", u2, body)
		assert.Equal(t, http.StatusNotFound, res.Code)
	})

	t.Run("non-admin bearer cannot set is_admin", func(t *testing.T) {
		// Re-enable registration so the non-admin request reaches the handler body;
		// the flag should still be silently ignored.
		config.ServiceEnableRegistration.Set(true)
		defer config.ServiceEnableRegistration.Set(false)

		body := `{"username":"sneaky-admin","password":"averyl0ngpassword","email":"sneaky-admin@example.com","is_admin":true}`
		res := adminReq(t, e, http.MethodPost, "/api/v1/register", u2, body)
		assert.Equal(t, http.StatusOK, res.Code, res.Body.String())

		s := db.NewSession()
		defer s.Close()
		u, err := user.GetUserByUsername(s, "sneaky-admin")
		require.NoError(t, err)
		assert.False(t, u.IsAdmin, "non-admin caller must not be able to promote")
	})
}

func TestAdmin_ReassignProjectOwner(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	license.SetForTests([]license.Feature{license.FeatureAdminPanel})
	defer license.ResetForTests()

	admin := promoteToAdmin(t, 1)

	t.Run("updates owner_id", func(t *testing.T) {
		// Project 2 is owned by user 1 in fixtures. Reassign to user 2.
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
		// user17 has status=2 (disabled) per fixtures. ErrAccountDisabled -> 412.
		res := adminReq(t, e, http.MethodPatch, "/api/v1/admin/projects/2/owner", admin, `{"owner_id":17}`)
		assert.Equal(t, http.StatusPreconditionFailed, res.Code)
	})

	t.Run("rejects locked user as new owner", func(t *testing.T) {
		// user18 has status=3 (locked) per fixtures. ErrAccountLocked -> 412.
		res := adminReq(t, e, http.MethodPatch, "/api/v1/admin/projects/2/owner", admin, `{"owner_id":18}`)
		assert.Equal(t, http.StatusPreconditionFailed, res.Code)
	})

	t.Run("rejects deletion-scheduled user as new owner", func(t *testing.T) {
		// user20 has deletion_scheduled_at set. Handing a project to a user whose
		// row is about to be deleted would cascade-delete the project.
		res := adminReq(t, e, http.MethodPatch, "/api/v1/admin/projects/2/owner", admin, `{"owner_id":20}`)
		assert.Equal(t, http.StatusBadRequest, res.Code)
	})
}

// TestAdmin_StaleAdminJWT_Gate proves the gate does not rely on the JWT's
// is_admin claim alone. A user who was admin at token-mint time but has since
// been demoted in the DB must be rejected the next time they hit /admin/*.
func TestAdmin_StaleAdminJWT_Gate(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	license.SetForTests([]license.Feature{license.FeatureAdminPanel})
	defer license.ResetForTests()

	// Mint a token with is_admin=true while the DB says the user is admin...
	admin := promoteToAdmin(t, 1)

	// ...then flip is_admin=false in the DB behind the token's back.
	s := db.NewSession()
	_, err = s.ID(int64(1)).Cols("is_admin").Update(&user.User{IsAdmin: false})
	require.NoError(t, err)
	require.NoError(t, s.Commit())
	s.Close()

	// The stale JWT still claims admin, but the gate re-checks the DB and denies.
	res := adminReq(t, e, http.MethodGet, "/api/v1/admin/ping", admin, "")
	assert.Equal(t, http.StatusNotFound, res.Code, "demoted admin with stale JWT must be rejected")
}

// TestAdmin_StaleAdminJWT_DeletedUser covers the case where the admin was
// deleted (not just demoted) — the gate must treat "user does not exist" the
// same as "not admin".
func TestAdmin_StaleAdminJWT_DeletedUser(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	license.SetForTests([]license.Feature{license.FeatureAdminPanel})
	defer license.ResetForTests()

	admin := promoteToAdmin(t, 1)

	// Delete the admin row directly. The JWT is still valid and carries is_admin=true.
	s := db.NewSession()
	_, err = s.ID(int64(1)).Delete(&user.User{})
	require.NoError(t, err)
	require.NoError(t, s.Commit())
	s.Close()

	res := adminReq(t, e, http.MethodGet, "/api/v1/admin/ping", admin, "")
	assert.Equal(t, http.StatusNotFound, res.Code, "deleted admin with stale JWT must be rejected")
}

// TestAdmin_StaleAdminJWT_PermissionBypass proves the model-level permission
// bypass (used by project/team/view CRUD) also re-checks the DB. A demoted
// admin's stale JWT must not let them access a project they do not otherwise
// have permissions on.
func TestAdmin_StaleAdminJWT_PermissionBypass(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	defer license.ResetForTests()

	// Project 2 is owned by user 3; user 1 has no share on it. Without the
	// admin bypass, reading it must fail.
	admin := promoteToAdmin(t, 1)

	// Sanity: while still admin, user 1 can read project 2.
	res := adminReq(t, e, http.MethodGet, "/api/v1/projects/2", admin, "")
	require.Equal(t, http.StatusOK, res.Code, "fresh admin must be able to read a project they do not own")

	// Demote in DB, keep the JWT.
	s := db.NewSession()
	_, err = s.ID(int64(1)).Cols("is_admin").Update(&user.User{IsAdmin: false})
	require.NoError(t, err)
	require.NoError(t, s.Commit())
	s.Close()

	// Same token, same request — must now fail because the bypass re-reads the DB.
	res = adminReq(t, e, http.MethodGet, "/api/v1/projects/2", admin, "")
	assert.NotEqual(t, http.StatusOK, res.Code, "demoted admin must lose project bypass after DB update")
}

// TestAdmin_StaleAdminJWT_Register covers the admin bypass on the public
// /register endpoint: after demotion, a stale admin JWT can no longer bypass
// ServiceEnableRegistration=false or set admin-only fields.
func TestAdmin_StaleAdminJWT_Register(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	license.SetForTests([]license.Feature{license.FeatureAdminPanel})
	defer license.ResetForTests()

	prev := config.ServiceEnableRegistration.GetBool()
	config.ServiceEnableRegistration.Set(false)
	defer config.ServiceEnableRegistration.Set(prev)

	admin := promoteToAdmin(t, 1)

	// Demote in DB — JWT still carries is_admin=true.
	s := db.NewSession()
	_, err = s.ID(int64(1)).Cols("is_admin").Update(&user.User{IsAdmin: false})
	require.NoError(t, err)
	require.NoError(t, s.Commit())
	s.Close()

	body := `{"username":"stale-admin","password":"averyl0ngpassword","email":"stale-admin@example.com","is_admin":true}`
	res := adminReq(t, e, http.MethodPost, "/api/v1/register", admin, body)
	assert.Equal(t, http.StatusNotFound, res.Code, "demoted admin must not bypass registration toggle")
}
