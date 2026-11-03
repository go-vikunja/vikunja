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

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Team 1 is shared with projects 3, 21, 28, 32, 33 and 34, with admin only on 34.
// user1 additionally owns 21.
func TestHumaTeamProjects(t *testing.T) {
	listIDs := func(t *testing.T, body []byte) []int64 {
		var res struct {
			Items []struct {
				ID int64 `json:"id"`
			} `json:"items"`
		}
		require.NoError(t, json.Unmarshal(body, &res))
		ids := []int64{}
		for _, p := range res.Items {
			ids = append(ids, p.ID)
		}
		return ids
	}

	t.Run("Only projects the caller administers", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/teams/1/projects", "", humaTokenFor(t, &testuser1), "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Equal(t, []int64{21, 34}, listIDs(t, rec.Body.Bytes()))
	})
	t.Run("Admin through the team share only", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/teams/1/projects", "", humaTokenFor(t, &testuser2), "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Equal(t, []int64{34}, listIDs(t, rec.Body.Bytes()))
	})
	t.Run("Forbidden for non-members", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		// user6 owns projects 28/32/33/34 but is not in team 1.
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/teams/1/projects", "", humaTokenFor(t, &testuser6), "")
		assert.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
	})
	t.Run("Forbidden for link shares", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token, err := auth.NewLinkShareJWTAuthtoken(&models.LinkSharing{ID: 1, Hash: "test", ProjectID: 1, Permission: models.PermissionRead})
		require.NoError(t, err)
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/teams/1/projects", "", token, "")
		assert.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
	})
	t.Run("Unauthenticated", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/teams/1/projects", "", "", "")
		assert.Equal(t, http.StatusUnauthorized, rec.Code, "body: %s", rec.Body.String())
	})
	t.Run("Nonexistent team", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/teams/9999/projects", "", humaTokenFor(t, &testuser1), "")
		assert.Equal(t, http.StatusNotFound, rec.Code, "body: %s", rec.Body.String())
	})
}
