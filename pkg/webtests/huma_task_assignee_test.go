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
	"net/url"
	"testing"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHumaTaskAssignee re-proves the v1 assignee contract on /api/v2. Labels
// has a dedicated v1 webtest; assignees never did — the v1 coverage lived in
// the model and in archived_test.go — so this ports the full create/list/delete
// matrix 1:1 so the v2 HTTP surface independently proves it once v1's routes go.
//
// create/delete both require WRITE access to the task's project
// (canDoTaskAssingee → project.CanUpdate); list requires READ. The share-permission
// tasks 15–26 map to the same projects the comment test's matrix uses, so the
// read-only-forbidden / write+admin-allowed split is identical to comment CREATE.
//
// Fixture topology (pkg/db/fixtures/task_assignees.yml, tasks.yml, projects.yml):
//   - task 30 (project 1, owned by user1): assignees user1 (#1) and user2 (#2).
//   - task 1 (project 1, owned by user1): no assignees; only user1 has project access.
//   - tasks 15–26: shared to user1 via every team/user/parent share kind.
//   - task 34 (project 20, user13): user1 has no access at all.
func TestHumaTaskAssignee(t *testing.T) {
	// task 30 belongs to project 1, owned by user1, and already has assignees.
	onTask30 := webHandlerTestV2{
		user:     &testuser1,
		basePath: "/api/v2/tasks/30/assignees",
		idParam:  "user",
		t:        t,
	}
	require.NoError(t, onTask30.ensureEnv())
	// onTaskAs reuses the one Echo instance (and its single fixture load) for a
	// different task. v2 does not reload fixtures per request, so the subtests
	// are ordered to avoid clobbering each other's rows.
	onTaskAs := func(taskID string, u *user.User) *webHandlerTestV2 {
		return &webHandlerTestV2{
			user:     u,
			basePath: "/api/v2/tasks/" + taskID + "/assignees",
			idParam:  "user",
			t:        t,
			e:        onTask30.e,
		}
	}
	// task 1 also belongs to project 1; used for clean creates (no fixture assignees).
	onTask1 := onTaskAs("1", &testuser1)
	// user6 has no access to project 1, so it can neither read nor write task 1.
	asUser6 := onTaskAs("1", &testuser6)

	t.Run("ReadAll", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := onTask30.testReadAllWithUser(nil, nil)
			require.NoError(t, err)
			ids := assigneeIDsFromReadAll(t, rec.Body.Bytes())
			// task 30's assignees are exactly user1 and user2.
			assert.ElementsMatch(t, []int64{1, 2}, ids,
				"ReadAll must return exactly {1,2}; body: %s", rec.Body.String())
			assert.Contains(t, rec.Body.String(), `"username":"user1"`)
			assert.Contains(t, rec.Body.String(), `"username":"user2"`)
		})
		t.Run("Empty", func(t *testing.T) {
			// task 1 has no assignees; the list envelope is still well-formed.
			rec, err := onTask1.testReadAllWithUser(nil, nil)
			require.NoError(t, err)
			ids := assigneeIDsFromReadAll(t, rec.Body.Bytes())
			assert.Empty(t, ids)
		})
		t.Run("Search filter", func(t *testing.T) {
			// ReadAll's search is an ILIKE on username; case-insensitive.
			rec, err := onTask30.testReadAllWithUser(url.Values{"q": []string{"USER2"}}, nil)
			require.NoError(t, err)
			ids := assigneeIDsFromReadAll(t, rec.Body.Bytes())
			assert.Equal(t, []int64{2}, ids, "search must narrow to user2; body: %s", rec.Body.String())
		})
		t.Run("Nonexisting task", func(t *testing.T) {
			_, err := onTaskAs("99999", &testuser1).testReadAllWithUser(nil, nil)
			require.Error(t, err)
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
			assertHandlerErrorCode(t, err, models.ErrCodeProjectDoesNotExist)
		})
		t.Run("Forbidden", func(t *testing.T) {
			// user6 cannot read task 1.
			_, err := asUser6.testReadAllWithUser(nil, nil)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
	})

	t.Run("Create", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			// Assign user1 to task 1: user1 has access to project 1 and may write it.
			rec, err := onTask1.testCreateWithUser(nil, nil, `{"user_id":1}`)
			require.NoError(t, err)
			assert.Equal(t, http.StatusCreated, rec.Code)
			assert.Contains(t, rec.Body.String(), `"user_id":1`)
			// created is server-set and serialized in snake_case.
			assert.Contains(t, rec.Body.String(), `"created":`)
		})
		t.Run("Assignee without project access", func(t *testing.T) {
			// user2 has no access to project 1, so it cannot be assigned to task 1.
			_, err := onTask1.testCreateWithUser(nil, nil, `{"user_id":2}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
			assertHandlerErrorCode(t, err, models.ErrCodeUserDoesNotHaveAccessToProject)
		})
		t.Run("Already assigned", func(t *testing.T) {
			// task 30 already has user1 assigned (fixture #1).
			_, err := onTask30.testCreateWithUser(nil, nil, `{"user_id":1}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusBadRequest, getHTTPErrorCode(err))
			assertHandlerErrorCode(t, err, models.ErrCodeUserAlreadyAssigned)
		})
		t.Run("Nonexisting user", func(t *testing.T) {
			_, err := onTask1.testCreateWithUser(nil, nil, `{"user_id":9999}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
			assertHandlerErrorCode(t, err, user.ErrCodeUserDoesNotExist)
		})
		t.Run("Nonexisting task", func(t *testing.T) {
			// The write check resolves the project from the task, so a missing
			// task surfaces project-does-not-exist as a 404.
			_, err := onTaskAs("99999", &testuser1).testCreateWithUser(nil, nil, `{"user_id":1}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
			assertHandlerErrorCode(t, err, models.ErrCodeProjectDoesNotExist)
		})
		t.Run("Forbidden no access", func(t *testing.T) {
			// user6 has no write access to task 1.
			_, err := asUser6.testCreateWithUser(nil, nil, `{"user_id":6}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})

		// Permission matrix: CREATE requires write access to the task, so
		// read-only shares are forbidden while write/admin shares are allowed.
		// user1 has access to all these projects, so it can be the assignee.
		// Mirrors the v1 archived/permission behaviour and the comment CREATE matrix.
		t.Run("Permissions check", func(t *testing.T) {
			// task 34 is owned by user13 — user1 has no access at all.
			t.Run("Forbidden no access", func(t *testing.T) {
				_, err := onTaskAs("34", &testuser1).testCreateWithUser(nil, nil, `{"user_id":1}`)
				require.Error(t, err)
				assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
			})

			// Read-only shares: create forbidden.
			forbiddenCreate := map[string]string{
				"Shared Via Team readonly":                "15",
				"Shared Via User readonly":                "18",
				"Shared Via Parent Project Team readonly": "21",
				"Shared Via Parent Project User readonly": "24",
			}
			for name, taskID := range forbiddenCreate {
				t.Run(name, func(t *testing.T) {
					_, err := onTaskAs(taskID, &testuser1).testCreateWithUser(nil, nil, `{"user_id":1}`)
					require.Error(t, err)
					assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
				})
			}

			// Write/admin shares: create allowed (8 positive cases).
			allowedCreate := map[string]string{
				"Shared Via Team write":                "16",
				"Shared Via Team admin":                "17",
				"Shared Via User write":                "19",
				"Shared Via User admin":                "20",
				"Shared Via Parent Project Team write": "22",
				"Shared Via Parent Project Team admin": "23",
				"Shared Via Parent Project User write": "25",
				"Shared Via Parent Project User admin": "26",
			}
			for name, taskID := range allowedCreate {
				t.Run(name, func(t *testing.T) {
					rec, err := onTaskAs(taskID, &testuser1).testCreateWithUser(nil, nil, `{"user_id":1}`)
					require.NoError(t, err)
					assert.Equal(t, http.StatusCreated, rec.Code)
					assert.Contains(t, rec.Body.String(), `"user_id":1`)
				})
			}
		})
	})

	t.Run("Delete", func(t *testing.T) {
		t.Run("Nonexisting assignee on writable task", func(t *testing.T) {
			// v1 parity: delete is permissive — removing a user who isn't assigned
			// to a task the caller can write still succeeds with no content.
			rec, err := onTask1.testDeleteWithUser(nil, map[string]string{"user": "9999"})
			require.NoError(t, err)
			assert.Equal(t, http.StatusNoContent, rec.Code)
			assert.Empty(t, rec.Body.String())
		})
		t.Run("Nonexisting task", func(t *testing.T) {
			_, err := onTaskAs("99999", &testuser1).testDeleteWithUser(nil, map[string]string{"user": "2"})
			require.Error(t, err)
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
			assertHandlerErrorCode(t, err, models.ErrCodeProjectDoesNotExist)
		})
		t.Run("Forbidden no access", func(t *testing.T) {
			// user6 has no write access to task 1.
			_, err := asUser6.testDeleteWithUser(nil, map[string]string{"user": "2"})
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})

		// Permission matrix: DELETE requires write access to the task, same as
		// create. Read-only shares 403 before touching the row; write/admin
		// succeed (the row need not exist — v1 delete is permissive).
		t.Run("Permissions check", func(t *testing.T) {
			t.Run("Forbidden no access", func(t *testing.T) {
				_, err := onTaskAs("34", &testuser1).testDeleteWithUser(nil, map[string]string{"user": "1"})
				require.Error(t, err)
				assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
			})

			forbiddenDelete := map[string]string{
				"Shared Via Team readonly":                "15",
				"Shared Via User readonly":                "18",
				"Shared Via Parent Project Team readonly": "21",
				"Shared Via Parent Project User readonly": "24",
			}
			for name, taskID := range forbiddenDelete {
				t.Run(name, func(t *testing.T) {
					_, err := onTaskAs(taskID, &testuser1).testDeleteWithUser(nil, map[string]string{"user": "1"})
					require.Error(t, err)
					assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
				})
			}

			allowedDelete := map[string]string{
				"Shared Via Team write":                "16",
				"Shared Via Team admin":                "17",
				"Shared Via User write":                "19",
				"Shared Via User admin":                "20",
				"Shared Via Parent Project Team write": "22",
				"Shared Via Parent Project Team admin": "23",
				"Shared Via Parent Project User write": "25",
				"Shared Via Parent Project User admin": "26",
			}
			for name, taskID := range allowedDelete {
				t.Run(name, func(t *testing.T) {
					rec, err := onTaskAs(taskID, &testuser1).testDeleteWithUser(nil, map[string]string{"user": "1"})
					require.NoError(t, err)
					assert.Equal(t, http.StatusNoContent, rec.Code)
				})
			}
		})

		t.Run("Normal", func(t *testing.T) {
			// Run last: removes user2 from task 30, a fixture row the ReadAll
			// cases above rely on.
			rec, err := onTask30.testDeleteWithUser(nil, map[string]string{"user": "2"})
			require.NoError(t, err)
			assert.Equal(t, http.StatusNoContent, rec.Code)
			assert.Empty(t, rec.Body.String())
		})
	})
}

// assigneeIDsFromReadAll extracts the user ids from a v2 paginated assignee list
// so the visible set can be asserted exactly rather than via substring matching.
func assigneeIDsFromReadAll(t *testing.T, body []byte) []int64 {
	t.Helper()
	var resp struct {
		Items []struct {
			ID int64 `json:"id"`
		} `json:"items"`
	}
	require.NoError(t, json.Unmarshal(body, &resp), "ReadAll body must be a paginated envelope: %s", string(body))
	ids := make([]int64, 0, len(resp.Items))
	for _, it := range resp.Items {
		ids = append(ids, it.ID)
	}
	return ids
}
