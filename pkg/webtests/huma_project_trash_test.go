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

// Fixture projects 45 and 46 (child of 45) are in user 1's trash.
func TestHumaProjectTrash(t *testing.T) {
	user1 := func(t *testing.T) string { return humaTokenFor(t, &testuser1) }
	stranger := func(t *testing.T) string { return humaTokenFor(t, &testuser13) }

	t.Run("list", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		rec := humaRequest(t, e, http.MethodGet, "/api/v2/projects/deleted", "", user1(t), "")
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		var list struct {
			Items []struct {
				ID                 int64 `json:"id"`
				TrashedSubProjects int64 `json:"trashed_sub_projects"`
			} `json:"items"`
		}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &list))
		require.Len(t, list.Items, 1)
		assert.Equal(t, int64(45), list.Items[0].ID)
		assert.Equal(t, int64(1), list.Items[0].TrashedSubProjects)

		token, err := auth.NewLinkShareJWTAuthtoken(&models.LinkSharing{ID: 3, Hash: "test3", ProjectID: 3, Permission: models.PermissionAdmin})
		require.NoError(t, err)
		rec = humaRequest(t, e, http.MethodGet, "/api/v2/projects/deleted", "", token, "")
		assert.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())
	})
	t.Run("delete moves to the trash and reading tells", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		rec := humaRequest(t, e, http.MethodDelete, "/api/v2/projects/1", "", user1(t), "")
		require.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())

		rec = humaRequest(t, e, http.MethodGet, "/api/v2/projects/1", "", user1(t), "")
		require.Equal(t, http.StatusNotFound, rec.Code, rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"code":3018`)
		assert.Contains(t, rec.Body.String(), `"deleted_at"`)

		rec = humaRequest(t, e, http.MethodGet, "/api/v2/projects/1", "", stranger(t), "")
		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.NotContains(t, rec.Body.String(), `"code":3018`)
	})
	t.Run("restore", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		rec := humaRequest(t, e, http.MethodPost, "/api/v2/projects/45/restore", "", stranger(t), "")
		assert.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())

		rec = humaRequest(t, e, http.MethodPost, "/api/v2/projects/46/restore", "", user1(t), "")
		assert.Equal(t, http.StatusPreconditionFailed, rec.Code, rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"code":3019`)

		rec = humaRequest(t, e, http.MethodPost, "/api/v2/projects/45/restore", "", user1(t), "")
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

		rec = humaRequest(t, e, http.MethodGet, "/api/v2/projects/46", "", user1(t), "")
		assert.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	})
	t.Run("delete permanently", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		rec := humaRequest(t, e, http.MethodDelete, "/api/v2/projects/45/permanent", "", stranger(t), "")
		assert.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())

		rec = humaRequest(t, e, http.MethodDelete, "/api/v2/projects/1/permanent", "", user1(t), "")
		assert.Equal(t, http.StatusNotFound, rec.Code, "live project: %s", rec.Body.String())

		rec = humaRequest(t, e, http.MethodDelete, "/api/v2/projects/45/permanent", "", user1(t), "")
		require.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())

		rec = humaRequest(t, e, http.MethodGet, "/api/v2/projects/45", "", user1(t), "")
		assert.NotContains(t, rec.Body.String(), `"code":3018`)
	})
	t.Run("empty trash", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		rec := humaRequest(t, e, http.MethodDelete, "/api/v2/projects/deleted", "", user1(t), "")
		require.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())

		rec = humaRequest(t, e, http.MethodGet, "/api/v2/projects/deleted", "", user1(t), "")
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"items":[]`)
	})
	t.Run("count", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		rec := humaRequest(t, e, http.MethodGet, "/api/v2/projects/1/count", "", user1(t), "")
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"time_entries":3`)
		assert.Contains(t, rec.Body.String(), `"time_entry_authors":1`)

		rec = humaRequest(t, e, http.MethodGet, "/api/v2/projects/1/count", "", stranger(t), "")
		assert.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())
	})
}
