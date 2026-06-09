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
	"testing"

	"code.vikunja.io/api/pkg/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTaskPositionV2 covers PUT /tasks/{task}/position. It drives the Echo+Huma
// stack directly (humaRequest/humaTokenFor) because webHandlerTestV2's buildURL
// only models base[/{id}] paths, not action sub-paths.
func TestTaskPositionV2(t *testing.T) {
	t.Run("updates the position of a writable task", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, &testuser1)

		// Task 1 lives in project 1, which testuser1 owns; view 1 belongs to project 1.
		rec := humaRequest(t, e, http.MethodPut, "/api/v2/tasks/1/position", `{"project_view_id":1,"position":256}`, token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())

		var resp models.TaskPosition
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, int64(1), resp.TaskID, "task id is taken from the URL")
		assert.Equal(t, int64(1), resp.ProjectViewID)
		assert.InDelta(t, 256.0, resp.Position, 0)
	})

	t.Run("path task id wins over the body", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, &testuser1)

		// Body names task 2, URL names task 1; the URL must win.
		rec := humaRequest(t, e, http.MethodPut, "/api/v2/tasks/1/position", `{"task_id":2,"project_view_id":1,"position":300}`, token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())

		var resp models.TaskPosition
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, int64(1), resp.TaskID)
	})

	t.Run("nonexistent task", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, &testuser1)

		rec := humaRequest(t, e, http.MethodPut, "/api/v2/tasks/99999/position", `{"project_view_id":1,"position":1}`, token, "")
		require.Equal(t, http.StatusNotFound, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"code":%d`, models.ErrCodeTaskDoesNotExist), "body must surface ErrCodeTaskDoesNotExist; body: %s", rec.Body.String())
	})

	t.Run("no access to the task is forbidden", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		// testuser15 cannot access task 1 (project 1, owned by testuser1).
		token := humaTokenFor(t, &testuser15)

		rec := humaRequest(t, e, http.MethodPut, "/api/v2/tasks/1/position", `{"project_view_id":1,"position":1}`, token, "")
		require.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
	})

	t.Run("read but no write on the task is forbidden", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		// Task 32 lives in project 3, on which testuser1 has read-only access.
		token := humaTokenFor(t, &testuser1)

		rec := humaRequest(t, e, http.MethodPut, "/api/v2/tasks/32/position", `{"project_view_id":1,"position":1}`, token, "")
		require.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
	})
}
