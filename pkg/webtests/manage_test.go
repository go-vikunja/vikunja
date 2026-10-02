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
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The people area is gated by the admin flag only. Every test runs WITHOUT a license on purpose.

func fixtureUser(t *testing.T, id int64) *user.User {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	u := &user.User{ID: id}
	has, err := s.Get(u)
	require.NoError(t, err)
	require.True(t, has)
	return u
}

func TestManageGate(t *testing.T) {
	t.Run("an admin gets in without a license", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.ResetForTests()

		admin := promoteToAdmin(t, 1)
		res := adminReq(t, e, http.MethodGet, "/api/v2/manage/users", admin, "")
		require.Equal(t, http.StatusOK, res.Code, res.Body.String())
	})
	t.Run("the licensed admin area stays closed without a license", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.ResetForTests()

		admin := promoteToAdmin(t, 1)
		res := adminReq(t, e, http.MethodGet, "/api/v2/admin/users", admin, "")
		assert.Equal(t, http.StatusNotFound, res.Code)
	})
	t.Run("a non-admin gets 404", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.ResetForTests()

		res := adminReq(t, e, http.MethodGet, "/api/v2/manage/users", fixtureUser(t, 2), "")
		assert.Equal(t, http.StatusNotFound, res.Code)
	})
	t.Run("an unauthenticated caller gets 401", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		res := adminReq(t, e, http.MethodGet, "/api/v2/manage/users", nil, "")
		assert.Equal(t, http.StatusUnauthorized, res.Code)
	})
	t.Run("every manage route is closed to non-admins, not just the list", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.ResetForTests()
		u := fixtureUser(t, 2)

		for _, tc := range []struct{ method, path, body string }{
			{http.MethodGet, "/api/v2/manage/users/1", ""},
			{http.MethodPut, "/api/v2/manage/users/1", `{"name":"x","email":"x@example.com"}`},
			{http.MethodPatch, "/api/v2/manage/users/1/admin", `{"is_admin":true}`},
			{http.MethodPatch, "/api/v2/manage/users/1/status", `{"status":2}`},
			{http.MethodPatch, "/api/v2/manage/users/1/password", `{"new_password":"abcdefgh12"}`},
			{http.MethodDelete, "/api/v2/manage/users/1", ""},
			{http.MethodGet, "/api/v2/manage/users/1/projects", ""},
			{http.MethodPost, "/api/v2/manage/users/1/transfer-projects", `{"new_owner_id":2}`},
			{http.MethodGet, "/api/v2/manage/user-import", ""},
			{http.MethodPost, "/api/v2/manage/user-import/preview", ""},
			{http.MethodPost, "/api/v2/manage/user-import/run", ""},
		} {
			res := adminReq(t, e, tc.method, tc.path, u, tc.body)
			assert.Equal(t, http.StatusNotFound, res.Code, "%s %s", tc.method, tc.path)
		}
	})
}

func TestManageUsersList(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	license.ResetForTests()
	admin := promoteToAdmin(t, 1)

	type row struct {
		ID              int64  `json:"id"`
		Username        string `json:"username"`
		Email           string `json:"email"`
		Source          string `json:"source"`
		ProfileEditable bool   `json:"profile_editable"`
	}
	type envelope struct {
		Items []row `json:"items"`
		Total int64 `json:"total"`
	}
	list := func(t *testing.T, query string) envelope {
		res := adminReq(t, e, http.MethodGet, "/api/v2/manage/users"+query, admin, "")
		require.Equal(t, http.StatusOK, res.Code, res.Body.String())
		var env envelope
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &env))
		return env
	}

	t.Run("lists users with email and source, bots excluded", func(t *testing.T) {
		env := list(t, "?per_page=100")
		require.NotEmpty(t, env.Items)
		var found bool
		for _, r := range env.Items {
			assert.NotContains(t, r.Username, "bot-owner", "bots are not part of the list")
			if r.Username == "user1" {
				found = true
				assert.Equal(t, "user1@example.com", r.Email)
				assert.Equal(t, models.UserSourceLocal, r.Source)
				assert.True(t, r.ProfileEditable)
			}
		}
		assert.True(t, found)
	})
	t.Run("an openid user is not editable", func(t *testing.T) {
		env := list(t, "?q=user14")
		require.Len(t, env.Items, 1)
		assert.Equal(t, models.UserSourceOther, env.Items[0].Source)
		assert.False(t, env.Items[0].ProfileEditable)
	})
	t.Run("source filter", func(t *testing.T) {
		env := list(t, "?source=other&per_page=100")
		require.NotEmpty(t, env.Items)
		for _, r := range env.Items {
			assert.Equal(t, models.UserSourceOther, r.Source)
		}
		local := list(t, "?source=local&per_page=100")
		for _, r := range local.Items {
			assert.Equal(t, models.UserSourceLocal, r.Source)
		}
	})
	t.Run("status filter", func(t *testing.T) {
		env := list(t, "?status=2&per_page=100")
		require.NotEmpty(t, env.Items, "fixture user17 is disabled")
		for _, r := range env.Items {
			assert.NotEqual(t, "user1", r.Username)
		}
	})
	t.Run("paging", func(t *testing.T) {
		env := list(t, "?per_page=2&page=1")
		assert.Len(t, env.Items, 2)
		assert.Greater(t, env.Total, int64(2))
	})
}

func TestManageUsersUpdate(t *testing.T) {
	t.Run("changes only name, email and language of a local user", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.ResetForTests()
		admin := promoteToAdmin(t, 1)
		before := fixtureUser(t, 2)

		res := adminReq(t, e, http.MethodPut, "/api/v2/manage/users/2", admin,
			`{"name":"Renamed By Admin","email":"renamed2@example.com"}`)
		require.Equal(t, http.StatusOK, res.Code, res.Body.String())

		after := fixtureUser(t, 2)
		assert.Equal(t, "Renamed By Admin", after.Name)
		assert.Equal(t, "renamed2@example.com", after.Email)
		// Everything else is untouched.
		assert.Equal(t, before.Username, after.Username)
		assert.Equal(t, before.Status, after.Status)
		assert.Equal(t, before.Timezone, after.Timezone)
		assert.Equal(t, before.WeekStart, after.WeekStart)
		assert.Equal(t, before.DefaultProjectID, after.DefaultProjectID)
		assert.Equal(t, before.DiscoverableByName, after.DiscoverableByName)
		assert.Equal(t, before.DiscoverableByEmail, after.DiscoverableByEmail)
		assert.Equal(t, before.EmailRemindersEnabled, after.EmailRemindersEnabled)
		assert.Equal(t, before.Language, after.Language)
		assert.Equal(t, before.IsAdmin, after.IsAdmin)
	})
	t.Run("a user managed by a third-party provider is refused with 403", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.ResetForTests()
		admin := promoteToAdmin(t, 1)
		before := fixtureUser(t, 14)

		res := adminReq(t, e, http.MethodPut, "/api/v2/manage/users/14", admin,
			`{"name":"Hacked","email":"hacked@example.com"}`)

		require.Equal(t, http.StatusForbidden, res.Code, res.Body.String())
		assert.Contains(t, res.Body.String(), "1042")
		after := fixtureUser(t, 14)
		assert.Equal(t, before.Name, after.Name)
		assert.Equal(t, before.Email, after.Email)
	})
	t.Run("a duplicate email is refused", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.ResetForTests()
		admin := promoteToAdmin(t, 1)

		res := adminReq(t, e, http.MethodPut, "/api/v2/manage/users/2", admin,
			`{"name":"x","email":"user3@example.com"}`)

		assert.GreaterOrEqual(t, res.Code, 400, res.Body.String())
		assert.Equal(t, "user2@example.com", fixtureUser(t, 2).Email)
	})
	t.Run("a malformed email and an unknown language are refused", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.ResetForTests()
		admin := promoteToAdmin(t, 1)

		bad := adminReq(t, e, http.MethodPut, "/api/v2/manage/users/2", admin, `{"name":"x","email":"not an email"}`)
		assert.GreaterOrEqual(t, bad.Code, 400, bad.Body.String())

		lang := adminReq(t, e, http.MethodPut, "/api/v2/manage/users/2", admin, `{"name":"x","email":"user2@example.com","language":"xx-nope"}`)
		assert.GreaterOrEqual(t, lang.Code, 400, lang.Body.String())
	})
	t.Run("the username is never changed, even if sent", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.ResetForTests()
		admin := promoteToAdmin(t, 1)

		res := adminReq(t, e, http.MethodPut, "/api/v2/manage/users/2", admin,
			`{"name":"x","email":"user2@example.com","username":"renamed"}`)

		// Unknown fields are either ignored or rejected, but the username must not change.
		_ = res
		assert.Equal(t, "user2", fixtureUser(t, 2).Username)
	})
}

func TestManagePasswordAndForcedChange(t *testing.T) {
	t.Run("admin set-password flags the user and the flagged token is limited", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.ResetForTests()
		admin := promoteToAdmin(t, 1)

		res := adminReq(t, e, http.MethodPatch, "/api/v2/manage/users/2/password", admin, `{"new_password":"a-brand-new-pass"}`)
		require.Equal(t, http.StatusOK, res.Code, res.Body.String())

		flagged := fixtureUser(t, 2)
		require.True(t, flagged.MustChangePassword, "require_change defaults to true")

		// The flagged token may read itself ...
		self := adminReq(t, e, http.MethodGet, "/api/v2/user", flagged, "")
		assert.Equal(t, http.StatusOK, self.Code, self.Body.String())
		assert.Contains(t, self.Body.String(), `"must_change_password":true`)

		// ... but nothing else.
		for _, path := range []string{"/api/v2/projects", "/api/v2/tasks", "/api/v2/labels", "/api/v2/user/settings/general"} {
			blocked := adminReq(t, e, http.MethodGet, path, flagged, "")
			assert.Equal(t, http.StatusForbidden, blocked.Code, path)
			assert.Contains(t, blocked.Body.String(), "1041", path)
		}
		v1 := adminReq(t, e, http.MethodGet, "/api/v1/projects", flagged, "")
		assert.Equal(t, http.StatusForbidden, v1.Code)
		v1User := adminReq(t, e, http.MethodGet, "/api/v1/user", flagged, "")
		assert.Equal(t, http.StatusOK, v1User.Code)
	})
	t.Run("require_change false leaves the flag off", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.ResetForTests()
		admin := promoteToAdmin(t, 1)

		res := adminReq(t, e, http.MethodPatch, "/api/v2/manage/users/3/password", admin, `{"new_password":"a-brand-new-pass","require_change":false}`)
		require.Equal(t, http.StatusOK, res.Code, res.Body.String())

		assert.False(t, fixtureUser(t, 3).MustChangePassword)
	})
	t.Run("changing the password clears the flag", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.ResetForTests()
		admin := promoteToAdmin(t, 1)

		set := adminReq(t, e, http.MethodPatch, "/api/v2/manage/users/2/password", admin, `{"new_password":"a-brand-new-pass"}`)
		require.Equal(t, http.StatusOK, set.Code, set.Body.String())
		flagged := fixtureUser(t, 2)
		require.True(t, flagged.MustChangePassword)

		// The v2 password endpoint is reachable while flagged.
		res := adminReq(t, e, http.MethodPost, "/api/v2/user/password", flagged,
			`{"old_password":"a-brand-new-pass","new_password":"my-own-new-pass"}`)
		require.Equal(t, http.StatusOK, res.Code, res.Body.String())

		assert.False(t, fixtureUser(t, 2).MustChangePassword)
	})
	t.Run("the new password must differ from the old one", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.ResetForTests()
		admin := promoteToAdmin(t, 1)
		set := adminReq(t, e, http.MethodPatch, "/api/v2/manage/users/2/password", admin, `{"new_password":"a-brand-new-pass"}`)
		require.Equal(t, http.StatusOK, set.Code, set.Body.String())
		flagged := fixtureUser(t, 2)

		res := adminReq(t, e, http.MethodPost, "/api/v2/user/password", flagged,
			`{"old_password":"a-brand-new-pass","new_password":"a-brand-new-pass"}`)

		assert.GreaterOrEqual(t, res.Code, 400, res.Body.String())
		assert.True(t, fixtureUser(t, 2).MustChangePassword, "the forced change is not ended by re-using the password")
	})
	t.Run("a third-party user cannot be given a password", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.ResetForTests()
		admin := promoteToAdmin(t, 1)

		res := adminReq(t, e, http.MethodPatch, "/api/v2/manage/users/14/password", admin, `{"new_password":"a-brand-new-pass"}`)

		assert.Equal(t, http.StatusPreconditionFailed, res.Code, res.Body.String())
	})
	t.Run("an unflagged user is not limited", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		res := adminReq(t, e, http.MethodGet, "/api/v2/projects", fixtureUser(t, 2), "")
		assert.Equal(t, http.StatusOK, res.Code, res.Body.String())
	})
}

func TestManageUsersCreateAndDelete(t *testing.T) {
	t.Run("creates a local user who has to change the password", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.ResetForTests()
		admin := promoteToAdmin(t, 1)

		res := adminReq(t, e, http.MethodPost, "/api/v2/manage/users", admin,
			`{"username":"newcomer","password":"temporary-pass","email":"newcomer@example.com","name":"New Comer","skip_email_confirm":true}`)
		require.Equal(t, http.StatusCreated, res.Code, res.Body.String())

		var created struct {
			ID                 int64 `json:"id"`
			MustChangePassword bool  `json:"must_change_password"`
			ProfileEditable    bool  `json:"profile_editable"`
		}
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &created))
		assert.True(t, created.MustChangePassword)
		assert.True(t, created.ProfileEditable)
		assert.True(t, fixtureUser(t, created.ID).MustChangePassword)
	})
	t.Run("deleting yourself is refused", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.ResetForTests()
		admin := promoteToAdmin(t, 1)

		res := adminReq(t, e, http.MethodDelete, "/api/v2/manage/users/1?mode=now", admin, "")

		assert.GreaterOrEqual(t, res.Code, 400, res.Body.String())
		_ = fixtureUser(t, 1) // still exists
	})
	t.Run("the last admin cannot be demoted", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.ResetForTests()
		admin := promoteToAdmin(t, 1)

		res := adminReq(t, e, http.MethodPatch, "/api/v2/manage/users/1/admin", admin, `{"is_admin":false}`)

		assert.Equal(t, http.StatusBadRequest, res.Code, res.Body.String())
		assert.True(t, fixtureUser(t, 1).IsAdmin)
	})
}

func TestManageProjectTransfer(t *testing.T) {
	type owned struct {
		Items []struct {
			ID    int64  `json:"id"`
			Title string `json:"title"`
		} `json:"items"`
	}

	t.Run("lists and hands over every project, in one go", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.ResetForTests()
		admin := promoteToAdmin(t, 1)

		res := adminReq(t, e, http.MethodGet, "/api/v2/manage/users/3/projects", admin, "")
		require.Equal(t, http.StatusOK, res.Code, res.Body.String())
		var before owned
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &before))
		require.NotEmpty(t, before.Items, "fixture user3 owns projects")

		move := adminReq(t, e, http.MethodPost, "/api/v2/manage/users/3/transfer-projects", admin, `{"new_owner_id":2}`)
		require.Equal(t, http.StatusOK, move.Code, move.Body.String())
		var moved struct {
			Transferred int `json:"transferred"`
		}
		require.NoError(t, json.Unmarshal(move.Body.Bytes(), &moved))
		assert.Equal(t, len(before.Items), moved.Transferred)

		after := adminReq(t, e, http.MethodGet, "/api/v2/manage/users/3/projects", admin, "")
		var left owned
		require.NoError(t, json.Unmarshal(after.Body.Bytes(), &left))
		assert.Empty(t, left.Items)
	})
	t.Run("a project the source does not own is refused and nothing moves", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.ResetForTests()
		admin := promoteToAdmin(t, 1)

		res := adminReq(t, e, http.MethodGet, "/api/v2/manage/users/3/projects", admin, "")
		var ownedBefore owned
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &ownedBefore))
		require.NotEmpty(t, ownedBefore.Items)

		// Project 1 belongs to user1, not user2.
		bad := adminReq(t, e, http.MethodPost, "/api/v2/manage/users/3/transfer-projects", admin, `{"new_owner_id":2,"project_ids":[1]}`)
		assert.Equal(t, http.StatusBadRequest, bad.Code, bad.Body.String())

		res = adminReq(t, e, http.MethodGet, "/api/v2/manage/users/3/projects", admin, "")
		var ownedAfter owned
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &ownedAfter))
		assert.Len(t, ownedAfter.Items, len(ownedBefore.Items))
	})
	t.Run("same owner, disabled owner and bot owner are refused", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.ResetForTests()
		admin := promoteToAdmin(t, 1)

		for name, newOwner := range map[string]string{
			"same user":     `{"new_owner_id":3}`,
			"disabled user": `{"new_owner_id":17}`,
			"bot":           `{"new_owner_id":23}`,
			"unknown user":  `{"new_owner_id":99999}`,
		} {
			res := adminReq(t, e, http.MethodPost, "/api/v2/manage/users/3/transfer-projects", admin, newOwner)
			assert.GreaterOrEqual(t, res.Code, 400, "%s: %s", name, res.Body.String())
		}
	})
}

func TestManageUserImportEndpoints(t *testing.T) {
	t.Run("status, preview and run answer for an admin and say what is wrong without a file", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.ResetForTests()
		admin := promoteToAdmin(t, 1)

		status := adminReq(t, e, http.MethodGet, "/api/v2/manage/user-import", admin, "")
		require.Equal(t, http.StatusOK, status.Code, status.Body.String())
		var st struct {
			FileExists bool `json:"file_exists"`
			Running    bool `json:"running"`
		}
		require.NoError(t, json.Unmarshal(status.Body.Bytes(), &st))
		assert.False(t, st.Running)

		if !st.FileExists {
			preview := adminReq(t, e, http.MethodPost, "/api/v2/manage/user-import/preview", admin, "")
			assert.Equal(t, http.StatusUnprocessableEntity, preview.Code, "a missing file is reported, not a 500: %s", preview.Body.String())
		}
	})
}
