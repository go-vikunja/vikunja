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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTaskDuplicateV2 mirrors the v1 TestTaskDuplicate webtest against the
// /api/v2 custom action POST /tasks/{projecttask}/duplicate. The harness
// (webHandlerTestV2.buildURL) only models base[/{id}] paths, so this action
// endpoint drives the Echo+Huma stack directly via humaRequest/humaTokenFor.
func TestTaskDuplicateV2(t *testing.T) {
	t.Run("duplicates an accessible task", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, &testuser1)

		// Task 2 lives in project 1, which testuser1 owns.
		rec := humaRequest(t, e, http.MethodPost, "/api/v2/tasks/2/duplicate", `{}`, token, "")
		require.Equal(t, http.StatusCreated, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"duplicated_task"`)
		assert.Contains(t, rec.Body.String(), `"title":"task #2 done"`)
	})

	t.Run("nonexistent source task", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, &testuser1)

		rec := humaRequest(t, e, http.MethodPost, "/api/v2/tasks/99999/duplicate", `{}`, token, "")
		// A genuinely missing source task surfaces ErrTaskDoesNotExist (404),
		// mirroring v1's TestTaskDuplicate "nonexistent task" case.
		require.Equal(t, http.StatusNotFound, rec.Code, "body: %s", rec.Body.String())
	})

	t.Run("no read on source task is forbidden", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		// testuser15 cannot read task 1 (project 1, owned by testuser1).
		token := humaTokenFor(t, &testuser15)

		rec := humaRequest(t, e, http.MethodPost, "/api/v2/tasks/1/duplicate", `{}`, token, "")
		require.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
	})

	t.Run("read but no write on source project is forbidden", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		// Task 32 lives in project 3, on which testuser1 has read-only access:
		// CanRead passes, CanUpdate on the project fails, so CanCreate denies.
		token := humaTokenFor(t, &testuser1)

		rec := humaRequest(t, e, http.MethodPost, "/api/v2/tasks/32/duplicate", `{}`, token, "")
		require.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
	})
}
