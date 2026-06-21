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
	"fmt"
	"net/http"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTaskRelationV2 covers POST /tasks/{task}/relations and
// DELETE /tasks/{task}/relations/{relationKind}/{otherTask}. It drives the
// Echo+Huma stack directly (humaRequest/humaTokenFor) because the action
// sub-paths aren't modelled by webHandlerTestV2's buildURL. Coverage mirrors
// the v1 model matrix in pkg/models/task_relation_test.go.
func TestTaskRelationV2(t *testing.T) {
	t.Run("Create", func(t *testing.T) {
		t.Run("creates forward and inverse rows", func(t *testing.T) {
			e, err := setupTestEnv()
			require.NoError(t, err)
			token := humaTokenFor(t, &testuser1)

			rec := humaRequest(t, e, http.MethodPost, "/api/v2/tasks/1/relations",
				`{"other_task_id":2,"relation_kind":"subtask"}`, token, "")
			require.Equal(t, http.StatusCreated, rec.Code, "body: %s", rec.Body.String())
			assert.Contains(t, rec.Body.String(), `"relation_kind":"subtask"`)
			assert.Contains(t, rec.Body.String(), `"task_id":1`)
			assert.Contains(t, rec.Body.String(), `"other_task_id":2`)

			// Create must store both directions: the forward subtask and the
			// automatically derived inverse parenttask.
			db.AssertExists(t, "task_relations", map[string]interface{}{
				"task_id":       1,
				"other_task_id": 2,
				"relation_kind": models.RelationKindSubtask,
				"created_by_id": 1,
			}, false)
			db.AssertExists(t, "task_relations", map[string]interface{}{
				"task_id":       2,
				"other_task_id": 1,
				"relation_kind": models.RelationKindParenttask,
				"created_by_id": 1,
			}, false)
		})

		t.Run("path task id wins over body", func(t *testing.T) {
			e, err := setupTestEnv()
			require.NoError(t, err)
			token := humaTokenFor(t, &testuser1)

			// task_id in the body is ignored; the row is created for the path task.
			rec := humaRequest(t, e, http.MethodPost, "/api/v2/tasks/1/relations",
				`{"task_id":999,"other_task_id":2,"relation_kind":"related"}`, token, "")
			require.Equal(t, http.StatusCreated, rec.Code, "body: %s", rec.Body.String())
			db.AssertExists(t, "task_relations", map[string]interface{}{
				"task_id":       1,
				"other_task_id": 2,
				"relation_kind": models.RelationKindRelated,
			}, false)
		})

		t.Run("cycle is rejected", func(t *testing.T) {
			e, err := setupTestEnv()
			require.NoError(t, err)
			token := humaTokenFor(t, &testuser1)

			// task 29 is already a subtask of task 1 (fixture); making task 1 a
			// subtask of task 29 would close the loop.
			rec := humaRequest(t, e, http.MethodPost, "/api/v2/tasks/29/relations",
				`{"other_task_id":1,"relation_kind":"subtask"}`, token, "")
			require.Equal(t, http.StatusConflict, rec.Code, "body: %s", rec.Body.String())
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"code":%d`, models.ErrCodeTaskRelationCycle))
		})

		t.Run("same task is rejected", func(t *testing.T) {
			e, err := setupTestEnv()
			require.NoError(t, err)
			token := humaTokenFor(t, &testuser1)

			rec := humaRequest(t, e, http.MethodPost, "/api/v2/tasks/1/relations",
				`{"other_task_id":1,"relation_kind":"related"}`, token, "")
			require.Equal(t, http.StatusBadRequest, rec.Code, "body: %s", rec.Body.String())
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"code":%d`, models.ErrCodeRelationTasksCannotBeTheSame))
		})

		t.Run("invalid relation kind in body is rejected by the enum", func(t *testing.T) {
			e, err := setupTestEnv()
			require.NoError(t, err)
			token := humaTokenFor(t, &testuser1)

			// relation_kind carries an enum constraint, so Huma rejects an unknown
			// kind with 422 before the handler runs (consistent with the delete path).
			rec := humaRequest(t, e, http.MethodPost, "/api/v2/tasks/1/relations",
				`{"other_task_id":2,"relation_kind":"bogus"}`, token, "")
			require.Equal(t, http.StatusUnprocessableEntity, rec.Code, "body: %s", rec.Body.String())
		})

		t.Run("nonexistent base task", func(t *testing.T) {
			e, err := setupTestEnv()
			require.NoError(t, err)
			token := humaTokenFor(t, &testuser1)

			rec := humaRequest(t, e, http.MethodPost, "/api/v2/tasks/999999/relations",
				`{"other_task_id":1,"relation_kind":"subtask"}`, token, "")
			require.Equal(t, http.StatusNotFound, rec.Code, "body: %s", rec.Body.String())
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"code":%d`, models.ErrCodeTaskDoesNotExist))
		})

		t.Run("forbidden - no write on base task", func(t *testing.T) {
			e, err := setupTestEnv()
			require.NoError(t, err)
			token := humaTokenFor(t, &testuser1)

			// task 15 is read-only for user1, so CanCreate (needs write on base) denies.
			rec := humaRequest(t, e, http.MethodPost, "/api/v2/tasks/15/relations",
				`{"other_task_id":1,"relation_kind":"subtask"}`, token, "")
			require.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
		})
	})

	t.Run("Delete", func(t *testing.T) {
		t.Run("removes forward and inverse rows", func(t *testing.T) {
			e, err := setupTestEnv()
			require.NoError(t, err)
			token := humaTokenFor(t, &testuser1)

			// Fixture relation 1: task 1 -subtask-> task 29, with the inverse
			// parenttask row (task 29 -> task 1).
			rec := humaRequest(t, e, http.MethodDelete, "/api/v2/tasks/1/relations/subtask/29", "", token, "")
			require.Equal(t, http.StatusNoContent, rec.Code, "body: %s", rec.Body.String())
			assert.Empty(t, rec.Body.String())

			db.AssertMissing(t, "task_relations", map[string]interface{}{
				"task_id":       1,
				"other_task_id": 29,
				"relation_kind": models.RelationKindSubtask,
			})
			db.AssertMissing(t, "task_relations", map[string]interface{}{
				"task_id":       29,
				"other_task_id": 1,
				"relation_kind": models.RelationKindParenttask,
			})
		})

		t.Run("invalid relation kind in path is rejected by the enum", func(t *testing.T) {
			e, err := setupTestEnv()
			require.NoError(t, err)
			token := humaTokenFor(t, &testuser1)

			// The path param carries an enum constraint, so Huma rejects an unknown
			// kind with 422 before the handler runs.
			rec := humaRequest(t, e, http.MethodDelete, "/api/v2/tasks/1/relations/bogus/29", "", token, "")
			require.Equal(t, http.StatusUnprocessableEntity, rec.Code, "body: %s", rec.Body.String())
		})

		t.Run("nonexistent relation", func(t *testing.T) {
			e, err := setupTestEnv()
			require.NoError(t, err)
			token := humaTokenFor(t, &testuser1)

			rec := humaRequest(t, e, http.MethodDelete, "/api/v2/tasks/1/relations/subtask/2", "", token, "")
			require.Equal(t, http.StatusNotFound, rec.Code, "body: %s", rec.Body.String())
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"code":%d`, models.ErrCodeRelationDoesNotExist))
		})

		t.Run("forbidden - no write on base task", func(t *testing.T) {
			e, err := setupTestEnv()
			require.NoError(t, err)
			token := humaTokenFor(t, &testuser1)

			// Fixture relation 7: task 41 -subtask-> task 43, owned by user15 in
			// project 36, which user1 cannot access — CanDelete denies.
			rec := humaRequest(t, e, http.MethodDelete, "/api/v2/tasks/41/relations/subtask/43", "", token, "")
			require.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
		})
	})
}
