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
	"net/http"
	"testing"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHumaTaskAssigneeBulk proves the v2 bulk-assignee replace contract:
// PUT /tasks/{projecttask}/assignees/bulk swaps the task's full assignee set
// for the posted list. Like the single-assignee test it gates on write access
// to the task's project (CanCreate → canDoTaskAssingee → project.CanUpdate).
//
// Fixture topology (pkg/db/fixtures/task_assignees.yml, tasks.yml, projects.yml,
// users_projects.yml):
//   - task 30 (project 1, owned by user1): assignees user1 (#1) and user2 (#2).
//     user2 is a fixture row only; user2 has NO access to project 1, so it can
//     be removed but never freshly added — replace cases here only remove it.
//   - tasks 16/19 (shared to user1 with write): user1 has project access, so
//     it is a valid assignee there — used for the add-from-empty case.
//   - tasks 15/18: shared read-only — write is forbidden.
//   - task 34 (project 20, user13): user1 has no access at all.
func TestHumaTaskAssigneeBulk(t *testing.T) {
	// One Echo env shared across users; setupTestEnv rotates the JWT secret per
	// call, so a second env would 401 tokens minted against the first.
	base := &webHandlerTestV2{user: &testuser1, t: t}
	require.NoError(t, base.ensureEnv())

	bulkPut := func(taskID string, u *user.User, payload string) (ids []int64, err error) {
		h := &webHandlerTestV2{user: u, basePath: "/api/v2/tasks/" + taskID + "/assignees/bulk", t: t, e: base.e}
		rec, err := h.serve(http.MethodPut, h.basePath, payload)
		if err != nil {
			return nil, err
		}
		// PUT defaults to 200 from the Register wrapper for a non-create verb.
		assert.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		return assigneeIDsFromReadAll(t, rec.Body.Bytes()), nil
	}
	// readAssignees fetches the current assignee set so a replace is verified
	// against persisted state, not just the response echo.
	readAssignees := func(taskID string, u *user.User) []int64 {
		h := &webHandlerTestV2{user: u, basePath: "/api/v2/tasks/" + taskID + "/assignees", idParam: "user", t: t, e: base.e}
		rec, err := h.testReadAllWithUser(nil, nil)
		require.NoError(t, err)
		return assigneeIDsFromReadAll(t, rec.Body.Bytes())
	}

	t.Run("Replace removes assignees not in the list", func(t *testing.T) {
		// task 30 starts as {1,2}; replacing with {1} must drop user2.
		require.ElementsMatch(t, []int64{1, 2}, readAssignees("30", &testuser1))
		_, err := bulkPut("30", &testuser1, `{"assignees":[{"id":1}]}`)
		require.NoError(t, err)
		assert.ElementsMatch(t, []int64{1}, readAssignees("30", &testuser1),
			"user2 must be unassigned after the replace")
	})

	t.Run("Empty list unassigns everyone", func(t *testing.T) {
		// task 30 now holds {1}; an empty array clears it entirely.
		_, err := bulkPut("30", &testuser1, `{"assignees":[]}`)
		require.NoError(t, err)
		assert.Empty(t, readAssignees("30", &testuser1),
			"an empty assignees array must remove all assignees")
	})

	t.Run("Replace adds new assignees", func(t *testing.T) {
		// task 16 is shared to user1 with write access and starts with no
		// assignees; user1 has project access, so it is a valid new assignee.
		require.Empty(t, readAssignees("16", &testuser1))
		_, err := bulkPut("16", &testuser1, `{"assignees":[{"id":1}]}`)
		require.NoError(t, err)
		assert.ElementsMatch(t, []int64{1}, readAssignees("16", &testuser1),
			"user1 must be assigned after the replace")
	})

	t.Run("Forbidden - read-only share", func(t *testing.T) {
		// task 18 is shared to user1 read-only; bulk replace needs write.
		_, err := bulkPut("18", &testuser1, `{"assignees":[{"id":1}]}`)
		require.Error(t, err)
		assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
	})

	t.Run("Forbidden - no access at all", func(t *testing.T) {
		// task 34 belongs to user13's private project 20.
		_, err := bulkPut("34", &testuser1, `{"assignees":[{"id":1}]}`)
		require.Error(t, err)
		assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
	})

	t.Run("Forbidden - user without project access", func(t *testing.T) {
		// user6 has no access to project 1, so it cannot write task 1.
		_, err := bulkPut("1", &testuser6, `{"assignees":[{"id":6}]}`)
		require.Error(t, err)
		assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
	})

	t.Run("Nonexisting task", func(t *testing.T) {
		// The write check resolves the project from the task, so a missing task
		// surfaces project-does-not-exist as a 404.
		_, err := bulkPut("99999", &testuser1, `{"assignees":[{"id":1}]}`)
		require.Error(t, err)
		assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
		assertHandlerErrorCode(t, err, models.ErrCodeProjectDoesNotExist)
	})
}
