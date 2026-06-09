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

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHumaTaskUnreadStatus ports v1's POST /tasks/:projecttask/read (no v1
// webtest exists). The action deletes the caller's unread entry for the task;
// there is no fixture file for task_unread_statuses, so the table starts empty
// and the test seeds the row it expects to clear.
//
// Note on the permission model: the v1 handler enforces nothing — CanUpdate is
// a hardcoded true and Update is an unconditional DELETE on (task_id, user_id).
// A task the caller can't see (or doesn't exist) therefore has no row to clear
// and the call succeeds as a no-op. The only thing actually gated is auth, so
// that is what the negative case covers.
func TestHumaTaskUnreadStatus(t *testing.T) {
	t.Run("Normal - clears the caller's unread entry", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		s := db.NewSession()
		_, err = s.Insert(&models.TaskUnreadStatus{TaskID: 1, UserID: testuser1.ID})
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		require.NoError(t, s.Close())

		token := humaTokenFor(t, &testuser1)
		rec := humaRequest(t, e, http.MethodPut, "/api/v2/tasks/1/read", "", token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"message":"success"`)

		db.AssertMissing(t, "task_unread_statuses", map[string]interface{}{
			"task_id": 1,
			"user_id": testuser1.ID,
		})
	})

	t.Run("No-op - already read, no entry to clear", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		token := humaTokenFor(t, &testuser1)
		rec := humaRequest(t, e, http.MethodPut, "/api/v2/tasks/1/read", "", token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"message":"success"`)
	})

	t.Run("No-op - nonexistent task (unenforced, mirrors v1)", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		token := humaTokenFor(t, &testuser1)
		rec := humaRequest(t, e, http.MethodPut, "/api/v2/tasks/99999/read", "", token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
	})

	t.Run("Anonymous request is rejected with 401", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		rec := humaRequest(t, e, http.MethodPut, "/api/v2/tasks/1/read", "", "", "")
		assert.Equal(t, http.StatusUnauthorized, rec.Code, "anonymous must get 401; body: %s", rec.Body.String())
	})
}
