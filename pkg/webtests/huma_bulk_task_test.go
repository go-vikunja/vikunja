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

// TestBulkTaskV2 covers PUT /tasks/bulk. It drives the Echo+Huma stack directly
// (humaRequest/humaTokenFor) because webHandlerTestV2's buildURL only models
// base[/{id}] paths, not action sub-paths.
func TestBulkTaskV2(t *testing.T) {
	t.Run("updates multiple tasks the user can write", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, &testuser1)

		// Tasks 1 and 2 both live in project 1, which testuser1 owns.
		rec := humaRequest(t, e, http.MethodPut, "/api/v2/tasks/bulk",
			`{"task_ids":[1,2],"fields":["title"],"values":{"title":"bulkupdated"}}`, token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())

		db.AssertExists(t, "tasks", map[string]interface{}{"id": 1, "title": "bulkupdated"}, false)
		db.AssertExists(t, "tasks", map[string]interface{}{"id": 2, "title": "bulkupdated"}, false)
	})

	t.Run("forbidden when missing write on one involved project", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, &testuser1)

		// Task 1 is in project 1 (owned), task 32 in project 3 (read-only share).
		// CanUpdate fans the write check across both projects, so the whole
		// request is rejected and neither task changes.
		rec := humaRequest(t, e, http.MethodPut, "/api/v2/tasks/bulk",
			`{"task_ids":[1,32],"fields":["title"],"values":{"title":"shouldnothappen"}}`, token, "")
		require.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())

		db.AssertMissing(t, "tasks", map[string]interface{}{"title": "shouldnothappen"})
	})

	t.Run("empty task_ids is rejected", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, &testuser1)

		rec := humaRequest(t, e, http.MethodPut, "/api/v2/tasks/bulk",
			`{"task_ids":[],"fields":["title"],"values":{"title":"bulkupdated"}}`, token, "")
		require.Equal(t, http.StatusBadRequest, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"code":%d`, models.ErrCodeBulkTasksNeedAtLeastOne), "body: %s", rec.Body.String())
	})
}
