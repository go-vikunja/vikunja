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
	"net/http/httptest"
	"testing"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMCPInfo(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	shareToken, err := auth.NewLinkShareJWTAuthtoken(&models.LinkSharing{ID: 1, Hash: "test", ProjectID: 1, Permission: models.PermissionRead, SharingType: models.SharingTypeWithoutPassword, SharedByID: 1})
	require.NoError(t, err)
	for _, tc := range []struct {
		name, token string
		status      int
	}{
		{"session", humaTokenFor(t, &user.User{ID: 1}), http.StatusOK},
		{"anonymous", "", http.StatusUnauthorized},
		{"api token", mcpFullToken, http.StatusForbidden},
		{"link share", shareToken, http.StatusForbidden},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v2/mcp/info", nil)
			if tc.token != "" {
				req.Header.Set("Authorization", "Bearer "+tc.token)
			}
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			require.Equal(t, tc.status, rec.Code, rec.Body.String())
			if tc.status != http.StatusOK {
				return
			}
			var info struct {
				Endpoint string                     `json:"endpoint"`
				Routes   map[string]map[string]any  `json:"routes"`
				Presets  map[string]json.RawMessage `json:"presets"`
			}
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &info))
			assert.Contains(t, info.Endpoint, "/api/v2/mcp")
			for _, group := range []string{"mcp", "tasks", "projects", "labels", "tasks_assignees", "tasks_comments"} {
				assert.NotEmpty(t, info.Routes[group], group)
			}
			assert.Contains(t, info.Routes["mcp"], "access")
			assert.Contains(t, info.Routes["other"], "users")
			for _, group := range []string{"tokens", "users", "webhooks", "projects_shares", "caldav", "mcp_info"} {
				assert.NotContains(t, info.Routes, group)
			}
			var typed map[string][]string
			require.NoError(t, json.Unmarshal(info.Presets["typed"], &typed))
			crud := []string{"create", "delete", "read_all", "read_one", "update"}
			assert.Len(t, typed, 6)
			for _, group := range []string{"tasks", "projects", "labels", "tasks_comments"} {
				assert.ElementsMatch(t, crud, typed[group], group)
			}
			assert.ElementsMatch(t, []string{"create", "delete", "read_all"}, typed["tasks_assignees"])
			assert.Contains(t, typed["other"], "users")
			for _, permission := range typed["other"] {
				token := &models.APIToken{APIPermissions: models.APIPermissions{"other": {permission}}}
				assert.True(t, token.CanUseRoute("/api/v2/users", http.MethodGet), permission)
			}
		})
	}
}
