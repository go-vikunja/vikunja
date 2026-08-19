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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Projects 44, 45 and 46 are the soft-deleted fixtures owned by user1; 45 is
// the parent of 46 so a restore of 45 must bring 46 back with it.
func TestHumaProjectRestore(t *testing.T) {
	deletedProjectIDs := func(t *testing.T, body []byte) []int64 {
		t.Helper()
		var projects []struct {
			ID        int64   `json:"id"`
			DeletedAt *string `json:"deleted_at"`
		}
		require.NoError(t, json.Unmarshal(body, &projects))
		ids := make([]int64, 0, len(projects))
		for _, p := range projects {
			assert.NotNil(t, p.DeletedAt, "project %d in the bin must carry deleted_at", p.ID)
			ids = append(ids, p.ID)
		}
		return ids
	}

	t.Run("ReadAll", func(t *testing.T) {
		t.Run("owner sees their bin", func(t *testing.T) {
			e, err := setupTestEnv()
			require.NoError(t, err)

			rec := humaRequest(t, e, http.MethodGet, "/api/v2/projects/deleted", "", humaTokenFor(t, &testuser1), "")
			require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
			assert.ElementsMatch(t, []int64{44, 45, 46}, deletedProjectIDs(t, rec.Body.Bytes()))
		})
		t.Run("other user sees an empty bin", func(t *testing.T) {
			e, err := setupTestEnv()
			require.NoError(t, err)

			rec := humaRequest(t, e, http.MethodGet, "/api/v2/projects/deleted", "", humaTokenFor(t, &testuser2), "")
			require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
			assert.Empty(t, deletedProjectIDs(t, rec.Body.Bytes()))
		})
		t.Run("unauthenticated", func(t *testing.T) {
			e, err := setupTestEnv()
			require.NoError(t, err)

			rec := humaRequest(t, e, http.MethodGet, "/api/v2/projects/deleted", "", "", "")
			assert.Equal(t, http.StatusUnauthorized, rec.Code, "body: %s", rec.Body.String())
		})
	})

	t.Run("Restore", func(t *testing.T) {
		t.Run("owner", func(t *testing.T) {
			e, err := setupTestEnv()
			require.NoError(t, err)
			token := humaTokenFor(t, &testuser1)

			rec := humaRequest(t, e, http.MethodPost, "/api/v2/projects/44/restore", "", token, "")
			require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
			assert.Contains(t, rec.Body.String(), `"deleted_at":null`)

			// The restored project is readable through the regular route again.
			rec = humaRequest(t, e, http.MethodGet, "/api/v2/projects/44", "", token, "")
			assert.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		})
		t.Run("parent restores its children", func(t *testing.T) {
			e, err := setupTestEnv()
			require.NoError(t, err)
			token := humaTokenFor(t, &testuser1)

			rec := humaRequest(t, e, http.MethodPost, "/api/v2/projects/45/restore", "", token, "")
			require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())

			rec = humaRequest(t, e, http.MethodGet, "/api/v2/projects/46", "", token, "")
			assert.Equal(t, http.StatusOK, rec.Code, "child 46 must be restored with its parent; body: %s", rec.Body.String())
		})
		t.Run("without access", func(t *testing.T) {
			e, err := setupTestEnv()
			require.NoError(t, err)

			rec := humaRequest(t, e, http.MethodPost, "/api/v2/projects/44/restore", "", humaTokenFor(t, &testuser2), "")
			assert.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
		})
		t.Run("nonexistent", func(t *testing.T) {
			e, err := setupTestEnv()
			require.NoError(t, err)

			rec := humaRequest(t, e, http.MethodPost, "/api/v2/projects/9999/restore", "", humaTokenFor(t, &testuser1), "")
			assert.Equal(t, http.StatusNotFound, rec.Code, "body: %s", rec.Body.String())
		})
		t.Run("project that is not deleted", func(t *testing.T) {
			e, err := setupTestEnv()
			require.NoError(t, err)

			rec := humaRequest(t, e, http.MethodPost, "/api/v2/projects/1/restore", "", humaTokenFor(t, &testuser1), "")
			assert.Equal(t, http.StatusNotFound, rec.Code, "body: %s", rec.Body.String())
		})
		t.Run("unauthenticated", func(t *testing.T) {
			e, err := setupTestEnv()
			require.NoError(t, err)

			rec := humaRequest(t, e, http.MethodPost, "/api/v2/projects/44/restore", "", "", "")
			assert.Equal(t, http.StatusUnauthorized, rec.Code, "body: %s", rec.Body.String())
		})
	})
}
