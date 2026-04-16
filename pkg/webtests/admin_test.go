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
	assert.Contains(t, body, `"shares"`)
	assert.Contains(t, body, `"version"`)
	assert.Contains(t, body, `"enabled_pro_features"`)
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

func TestAdmin_CreateUser(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	license.SetForTests([]license.Feature{license.FeatureAdminPanel})
	defer license.ResetForTests()

	admin := promoteToAdmin(t, 1)

	t.Run("creates regular user with password", func(t *testing.T) {
		body := `{"username":"newuser1","email":"newuser1@example.com","password":"averyl0ngpassword","name":"New User"}`
		res := adminReq(t, e, http.MethodPost, "/api/v1/admin/users", admin, body)
		assert.Equal(t, http.StatusOK, res.Code)
		assert.Contains(t, res.Body.String(), `"username":"newuser1"`)

		s := db.NewSession()
		defer s.Close()
		u, err := user.GetUserByUsername(s, "newuser1")
		require.NoError(t, err)
		assert.False(t, u.IsAdmin)
		// Mailer off in tests + password provided → status forced Active.
		assert.Equal(t, user.StatusActive, u.Status)

		// An Inbox project must have been created so the user has somewhere to land.
		has, err := s.Table("projects").Where("owner_id = ? AND title = ?", u.ID, "Inbox").Exist()
		require.NoError(t, err)
		assert.True(t, has, "inbox project must exist for new user")
	})

	t.Run("creates admin user when is_admin true", func(t *testing.T) {
		body := `{"username":"newadmin","email":"newadmin@example.com","password":"averyl0ngpassword","is_admin":true}`
		res := adminReq(t, e, http.MethodPost, "/api/v1/admin/users", admin, body)
		assert.Equal(t, http.StatusOK, res.Code)
		assert.Contains(t, res.Body.String(), `"is_admin":true`)

		s := db.NewSession()
		defer s.Close()
		u, err := user.GetUserByUsername(s, "newadmin")
		require.NoError(t, err)
		assert.True(t, u.IsAdmin)
	})

	t.Run("rejects duplicate username", func(t *testing.T) {
		// user1 already exists in fixtures.
		body := `{"username":"user1","email":"dupe@example.com","password":"averyl0ngpassword"}`
		res := adminReq(t, e, http.MethodPost, "/api/v1/admin/users", admin, body)
		assert.Equal(t, http.StatusBadRequest, res.Code)
	})

	t.Run("rejects duplicate email", func(t *testing.T) {
		// user1@example.com already exists in fixtures.
		body := `{"username":"fresh-username","email":"user1@example.com","password":"averyl0ngpassword"}`
		res := adminReq(t, e, http.MethodPost, "/api/v1/admin/users", admin, body)
		assert.Equal(t, http.StatusBadRequest, res.Code)
	})

	t.Run("rejects missing password when mailer disabled", func(t *testing.T) {
		body := `{"username":"nopassword","email":"nopassword@example.com"}`
		res := adminReq(t, e, http.MethodPost, "/api/v1/admin/users", admin, body)
		assert.Equal(t, http.StatusBadRequest, res.Code)
		assert.Contains(t, res.Body.String(), "password is required when mailer is disabled")
	})

	t.Run("rejects invalid email", func(t *testing.T) {
		body := `{"username":"bademail","email":"not-an-email","password":"averyl0ngpassword"}`
		res := adminReq(t, e, http.MethodPost, "/api/v1/admin/users", admin, body)
		assert.Equal(t, http.StatusBadRequest, res.Code)
	})

	t.Run("rejects username with spaces", func(t *testing.T) {
		body := `{"username":"has space","email":"space@example.com","password":"averyl0ngpassword"}`
		res := adminReq(t, e, http.MethodPost, "/api/v1/admin/users", admin, body)
		assert.Equal(t, http.StatusBadRequest, res.Code)
	})

	t.Run("unauthorized non-admin caller gets 404", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()
		u, err := user.GetUserByID(s, 2)
		require.NoError(t, err)
		// user2 is not an admin in this test's state.
		assert.False(t, u.IsAdmin)

		body := `{"username":"sneaky","email":"sneaky@example.com","password":"averyl0ngpassword"}`
		res := adminReq(t, e, http.MethodPost, "/api/v1/admin/users", u, body)
		assert.Equal(t, http.StatusNotFound, res.Code)
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
		assert.Equal(t, http.StatusBadRequest, res.Code)
	})

	t.Run("nonexistent project returns 404", func(t *testing.T) {
		res := adminReq(t, e, http.MethodPatch, "/api/v1/admin/projects/99999/owner", admin, `{"owner_id":1}`)
		assert.Equal(t, http.StatusNotFound, res.Code)
	})
}
