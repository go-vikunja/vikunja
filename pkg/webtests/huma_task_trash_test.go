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

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type trashListResponse struct {
	Items []struct {
		ID        int64  `json:"id"`
		DeletedAt string `json:"deleted_at"`
	} `json:"items"`
	Total int64 `json:"total"`
}

func trashedIDs(t *testing.T, body []byte) []int64 {
	t.Helper()
	var list trashListResponse
	require.NoError(t, json.Unmarshal(body, &list))
	ids := []int64{}
	for _, item := range list.Items {
		ids = append(ids, item.ID)
	}
	return ids
}

func assertTaskInTrash(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	require.Equal(t, http.StatusNotFound, rec.Code, rec.Body.String())
	var body struct {
		Code    int `json:"code"`
		Details struct {
			ProjectID int64  `json:"project_id"`
			DeletedAt string `json:"deleted_at"`
		} `json:"details"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, models.ErrCodeTaskIsInTrash, body.Code, rec.Body.String())
	assert.Equal(t, int64(1), body.Details.ProjectID, rec.Body.String())
	assert.NotEmpty(t, body.Details.DeletedAt, rec.Body.String())
}

// Task 51 in project 1 (owned by user 1) is soft-deleted in the fixtures.
func TestHumaTaskTrash(t *testing.T) {
	user1 := func(t *testing.T) string { return humaTokenFor(t, &testuser1) }
	stranger := func(t *testing.T) string { return humaTokenFor(t, &testuser13) }
	writeShare := &models.LinkSharing{
		ID:         2,
		Hash:       "test2",
		ProjectID:  2,
		Permission: models.PermissionWrite,
	}

	t.Run("list", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		rec := humaRequest(t, e, http.MethodGet, "/api/v2/tasks/deleted", "", user1(t), "")
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		assert.Equal(t, []int64{51}, trashedIDs(t, rec.Body.Bytes()))
		assert.Contains(t, rec.Body.String(), `"deleted_at"`)

		rec = humaRequest(t, e, http.MethodGet, "/api/v2/tasks/deleted", "", stranger(t), "")
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		assert.Empty(t, trashedIDs(t, rec.Body.Bytes()))
	})
	t.Run("link shares have no trash", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token, err := auth.NewLinkShareJWTAuthtoken(writeShare)
		require.NoError(t, err)

		rec := humaRequest(t, e, http.MethodGet, "/api/v2/tasks/deleted", "", token, "")
		assert.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())
		rec = humaRequest(t, e, http.MethodDelete, "/api/v2/tasks/deleted", "", token, "")
		assert.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())
	})
	t.Run("reading a trashed task", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		rec := humaRequest(t, e, http.MethodGet, "/api/v2/tasks/51", "", user1(t), "")
		assertTaskInTrash(t, rec)

		rec = humaRequest(t, e, http.MethodGet, "/api/v2/tasks/51", "", stranger(t), "")
		require.Equal(t, http.StatusNotFound, rec.Code, rec.Body.String())
		assert.NotContains(t, rec.Body.String(), "details")
		assert.Contains(t, rec.Body.String(), `"code":4002`)
	})
	t.Run("writing a trashed task", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		writes := []struct {
			method string
			path   string
			body   string
		}{
			{
				method: http.MethodPut,
				path:   "/api/v2/tasks/51",
				body:   `{"title":"edited"}`,
			},
			{
				method: http.MethodDelete,
				path:   "/api/v2/tasks/51",
			},
			{
				method: http.MethodPost,
				path:   "/api/v1/tasks/51",
				body:   `{"title":"edited"}`,
			},
			{
				method: http.MethodDelete,
				path:   "/api/v1/tasks/51",
			},
		}
		for _, w := range writes {
			t.Run(w.method+" "+w.path, func(t *testing.T) {
				rec := humaRequest(t, e, w.method, w.path, w.body, user1(t), "")
				assertTaskInTrash(t, rec)

				rec = humaRequest(t, e, w.method, w.path, w.body, stranger(t), "")
				require.Equal(t, http.StatusNotFound, rec.Code, rec.Body.String())
				assert.NotContains(t, rec.Body.String(), "details")
				assert.Contains(t, rec.Body.String(), `"code":4002`)
			})
		}
		db.AssertExists(t, "tasks", map[string]interface{}{
			"id":    51,
			"title": "task #51 soft deleted",
		}, false)
	})
	t.Run("restore", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		rec := humaRequest(t, e, http.MethodPost, "/api/v2/tasks/51/restore", "", stranger(t), "")
		assert.Equal(t, http.StatusNotFound, rec.Code, rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"code":4002`)

		user3 := &user.User{
			ID:       3,
			Username: "user3",
			Issuer:   "local",
		}
		rec = humaRequest(t, e, http.MethodDelete, "/api/v2/tasks/32", "", humaTokenFor(t, user3), "")
		require.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())
		rec = humaRequest(t, e, http.MethodPost, "/api/v2/tasks/32/restore", "", user1(t), "")
		assert.Equal(t, http.StatusForbidden, rec.Code, "reader: %s", rec.Body.String())

		rec = humaRequest(t, e, http.MethodPost, "/api/v2/tasks/51/restore", "", user1(t), "")
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		var restored map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &restored))
		assert.InDelta(t, 51, restored["id"], 0)
		assert.NotEmpty(t, restored["identifier"], rec.Body.String())
		createdBy, ok := restored["created_by"].(map[string]any)
		require.True(t, ok, rec.Body.String())
		assert.NotZero(t, createdBy["id"])
		assert.Nil(t, restored["deleted_at"], rec.Body.String())

		rec = humaRequest(t, e, http.MethodGet, "/api/v2/tasks/51", "", user1(t), "")
		assert.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

		rec = humaRequest(t, e, http.MethodPost, "/api/v2/tasks/51/restore", "", user1(t), "")
		assert.Equal(t, http.StatusNotFound, rec.Code, "live task is not in the trash: %s", rec.Body.String())
	})
	t.Run("write link share undoes its own delete", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token, err := auth.NewLinkShareJWTAuthtoken(writeShare)
		require.NoError(t, err)

		rec := humaRequest(t, e, http.MethodDelete, "/api/v2/tasks/13", "", token, "")
		require.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())
		rec = humaRequest(t, e, http.MethodPost, "/api/v2/tasks/13/restore", "", token, "")
		assert.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	})
	t.Run("delete permanently", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		rec := humaRequest(t, e, http.MethodDelete, "/api/v2/tasks/51/permanent", "", stranger(t), "")
		assert.Equal(t, http.StatusNotFound, rec.Code, rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"code":4002`)

		rec = humaRequest(t, e, http.MethodDelete, "/api/v2/tasks/1/permanent", "", user1(t), "")
		assert.Equal(t, http.StatusNotFound, rec.Code, "live task: %s", rec.Body.String())

		rec = humaRequest(t, e, http.MethodDelete, "/api/v2/tasks/51/permanent", "", user1(t), "")
		require.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())

		rec = humaRequest(t, e, http.MethodGet, "/api/v2/tasks/51", "", user1(t), "")
		assert.Equal(t, http.StatusNotFound, rec.Code, rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"code":4002`)
	})
	t.Run("purging needs admin", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		// User 1 can write to project 10 (task 19) and is admin of project 11 (task 20)
		for _, path := range []string{
			"/api/v2/tasks/19",
			"/api/v2/tasks/20",
		} {
			rec := humaRequest(t, e, http.MethodDelete, path, "", user1(t), "")
			require.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())
		}

		rec := humaRequest(t, e, http.MethodDelete, "/api/v2/tasks/19/permanent", "", user1(t), "")
		assert.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())

		rec = humaRequest(t, e, http.MethodDelete, "/api/v2/tasks/deleted", "", user1(t), "")
		require.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())
		db.AssertMissing(t, "tasks", map[string]interface{}{"id": 20})

		rec = humaRequest(t, e, http.MethodGet, "/api/v2/tasks/deleted", "", user1(t), "")
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		assert.Equal(t, []int64{19}, trashedIDs(t, rec.Body.Bytes()))

		rec = humaRequest(t, e, http.MethodPost, "/api/v2/tasks/19/restore", "", user1(t), "")
		assert.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	})
	t.Run("stranger emptying their trash leaves others' trash alone", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		rec := humaRequest(t, e, http.MethodDelete, "/api/v2/tasks/deleted", "", stranger(t), "")
		require.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())

		db.AssertExists(t, "tasks", map[string]interface{}{"id": 51}, false)
		rec = humaRequest(t, e, http.MethodGet, "/api/v2/tasks/deleted", "", user1(t), "")
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		assert.Equal(t, []int64{51}, trashedIDs(t, rec.Body.Bytes()))
	})
	t.Run("empty trash", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		rec := humaRequest(t, e, http.MethodDelete, "/api/v2/tasks/1", "", user1(t), "")
		require.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())

		rec = humaRequest(t, e, http.MethodDelete, "/api/v2/tasks/deleted", "", user1(t), "")
		require.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())

		rec = humaRequest(t, e, http.MethodGet, "/api/v2/tasks/deleted", "", user1(t), "")
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		assert.Empty(t, trashedIDs(t, rec.Body.Bytes()))
	})
}
