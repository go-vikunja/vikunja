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
	"fmt"
	"net/http"
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/license"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/routes"
	"code.vikunja.io/api/pkg/user"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/require"
)

func TestHumaInviteLinkAdmin(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	license.SetForTests([]license.Feature{license.FeatureAdminPanel, license.FeatureUserInvites})
	defer license.ResetForTests()
	admin := promoteToAdmin(t, 1)
	created := adminReq(t, e, http.MethodPost, "/api/v2/admin/invite-links", admin, `{"name":"welcome","team_ids":[1,8],"skip_email_confirm":true}`)
	require.Equal(t, http.StatusCreated, created.Code, created.Body.String())
	link := &models.UserInviteLink{}
	require.NoError(t, json.Unmarshal(created.Body.Bytes(), link))
	require.Len(t, link.ClearTextToken, 64)
	listed := adminReq(t, e, http.MethodGet, "/api/v2/admin/invite-links", admin, "")
	require.Equal(t, http.StatusOK, listed.Code, listed.Body.String())
	require.Contains(t, listed.Body.String(), "welcome")
	require.NotContains(t, listed.Body.String(), link.ClearTextToken)
	require.NotContains(t, listed.Body.String(), "token_hash")
	teams := adminReq(t, e, http.MethodGet, "/api/v2/admin/teams?q=testteam8", admin, "")
	require.Equal(t, http.StatusOK, teams.Code, teams.Body.String())
	require.Contains(t, teams.Body.String(), "testteam8")
	teams = adminReq(t, e, http.MethodGet, "/api/v2/admin/teams", admin, "")
	require.NotContains(t, teams.Body.String(), "testteam14")
	for _, tc := range []struct {
		name   string
		body   string
		status int
	}{
		{"unknown team", `{"name":"bad","team_ids":[999]}`, http.StatusNotFound},
		{"external team", `{"name":"external","team_ids":[14]}`, http.StatusBadRequest},
		{"missing name", `{}`, http.StatusUnprocessableEntity},
		{"empty name", `{"name":""}`, http.StatusUnprocessableEntity},
		{"blank name", `{"name":"  "}`, http.StatusUnprocessableEntity},
		{"long name", fmt.Sprintf(`{"name":%q}`, strings.Repeat("a", 251)), http.StatusUnprocessableEntity},
		{"zero uses", `{"name":"zero","max_uses":0}`, http.StatusUnprocessableEntity},
		{"negative uses", `{"name":"negative","max_uses":-1}`, http.StatusUnprocessableEntity},
		{"past expiry", `{"name":"expired","expires_at":"2000-01-01T00:00:00Z"}`, http.StatusUnprocessableEntity},
	} {
		t.Run(tc.name, func(t *testing.T) {
			res := adminReq(t, e, http.MethodPost, "/api/v2/admin/invite-links", admin, tc.body)
			require.Equal(t, tc.status, res.Code, res.Body.String())
		})
	}
	deleted := adminReq(t, e, http.MethodDelete, fmt.Sprintf("/api/v2/admin/invite-links/%d", link.ID), admin, "")
	require.Equal(t, http.StatusNoContent, deleted.Code, deleted.Body.String())
}

func TestHumaInviteLinkAdminDenied(t *testing.T) {
	for _, mode := range []string{"non-admin", "no invites", "no admin panel"} {
		t.Run(mode, func(t *testing.T) {
			e, err := setupTestEnv()
			require.NoError(t, err)
			license.SetForTests([]license.Feature{license.FeatureAdminPanel, license.FeatureUserInvites})
			defer license.ResetForTests()
			admin := promoteToAdmin(t, 1)
			switch mode {
			case "non-admin":
				admin = &user.User{ID: 2, Username: "user2"}
			case "no invites":
				license.SetForTests([]license.Feature{license.FeatureAdminPanel})
			case "no admin panel":
				license.SetForTests([]license.Feature{license.FeatureUserInvites})
			}
			for _, req := range []struct{ method, path, body string }{
				{http.MethodGet, "/api/v2/admin/invite-links", ""},
				{http.MethodPost, "/api/v2/admin/invite-links", `{"name":"denied"}`},
				{http.MethodDelete, "/api/v2/admin/invite-links/1", ""},
				{http.MethodGet, "/api/v2/admin/teams", ""},
			} {
				res := adminReq(t, e, req.method, req.path, admin, req.body)
				require.Equal(t, http.StatusNotFound, res.Code, res.Body.String())
			}
		})
	}
}

func TestHumaInviteLinkPublic(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	license.SetForTests([]license.Feature{license.FeatureUserInvites})
	defer license.ResetForTests()
	res := adminReq(t, e, http.MethodPost, "/api/v2/invite-links/check", nil, `{"token":"unlimited"}`)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())
	require.Contains(t, res.Body.String(), "testteam1")
	require.NotContains(t, res.Body.String(), "token_hash")
	var deadBody string
	for _, token := range []string{"expired", "exhausted", "unknown"} {
		res = adminReq(t, e, http.MethodPost, "/api/v2/invite-links/check", nil, fmt.Sprintf(`{"token":%q}`, token))
		require.Equal(t, http.StatusNotFound, res.Code)
		if deadBody == "" {
			deadBody = res.Body.String()
		} else {
			require.Equal(t, deadBody, res.Body.String())
		}
		res = adminReq(t, e, http.MethodPost, "/api/v2/register", nil, fmt.Sprintf(`{"invite_token":%q,"username":"invalid-invite","email":"invalid@example.com","password":"12345678"}`, token))
		require.Equal(t, http.StatusNotFound, res.Code, res.Body.String())
	}
	license.ResetForTests()
	res = adminReq(t, e, http.MethodPost, "/api/v2/invite-links/check", nil, `{"token":"unlimited"}`)
	require.Equal(t, http.StatusNotFound, res.Code)
	res = adminReq(t, e, http.MethodPost, "/api/v2/register", nil, `{"invite_token":"unlimited","username":"invalid-invite","email":"invalid@example.com","password":"12345678"}`)
	require.Equal(t, http.StatusNotFound, res.Code)
}

func TestHumaInviteLinkRegistrationDisabled(t *testing.T) {
	_, err := setupTestEnv()
	require.NoError(t, err)
	old := config.ServiceEnableRegistration.GetBool()
	config.ServiceEnableRegistration.Set(false)
	defer config.ServiceEnableRegistration.Set(old)
	e := echo.New()
	routes.RegisterRoutes(e)
	license.SetForTests([]license.Feature{license.FeatureUserInvites})
	defer license.ResetForTests()
	body := `{"username":"invite-web","email":"invite-web@example.com","password":"12345678"}`
	regular := adminReq(t, e, http.MethodPost, "/api/v2/register", nil, body)
	require.Equal(t, http.StatusNotFound, regular.Code, regular.Body.String())
	res := adminReq(t, e, http.MethodPost, "/api/v2/register", nil, `{"invite_token":"unlimited","username":"invite-web","email":"invite-web@example.com","password":"12345678"}`)
	require.Equal(t, http.StatusCreated, res.Code, res.Body.String())
	u := &user.User{}
	require.NoError(t, json.Unmarshal(res.Body.Bytes(), u))
	db.AssertExists(t, "team_members", map[string]interface{}{"team_id": 1, "user_id": u.ID}, false)
	login := adminReq(t, e, http.MethodPost, "/api/v2/login", nil, `{"username":"invite-web","password":"12345678"}`)
	require.Equal(t, http.StatusOK, login.Code, login.Body.String())
}

func TestHumaInviteLinkTokenScopes(t *testing.T) {
	_, err := setupTestEnv()
	require.NoError(t, err)
	defer license.ResetForTests()
	for _, enabled := range []bool{false, true} {
		features := []license.Feature{license.FeatureAdminPanel}
		if enabled {
			features = append(features, license.FeatureUserInvites)
		}
		license.SetForTests(features)
		found, existingAdmin := false, false
		for _, group := range models.GetAPITokenRoutes() {
			for _, route := range group {
				if route.Path == "/api/v2/admin/invite-links" {
					found = true
				}
				if route.Path == "/api/v2/admin/teams" {
					require.True(t, enabled)
				}
				if route.Path == "/api/v1/admin/users" || route.Path == "/api/v2/admin/users" {
					existingAdmin = true
				}
			}
		}
		require.Equal(t, enabled, found)
		require.True(t, existingAdmin)
	}
}

func TestHumaInviteLinkRejectsEmptyToken(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	license.SetForTests([]license.Feature{license.FeatureUserInvites})
	defer license.ResetForTests()
	for _, request := range []struct{ path, body string }{
		{"/api/v2/invite-links/check", `{}`},
		{"/api/v2/register", `{"invite_token":"","username":"missing-invite","email":"missing-invite@example.com","password":"12345678"}`},
	} {
		res := adminReq(t, e, http.MethodPost, request.path, nil, request.body)
		require.Equal(t, http.StatusNotFound, res.Code, res.Body.String())
	}
	db.AssertMissing(t, "users", map[string]interface{}{"username": "missing-invite"})
}
