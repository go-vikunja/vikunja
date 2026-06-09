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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLabelTaskBulk_V2 ports the v1 bulk-replace matrix
// (pkg/webtests/label_task_test.go) onto PUT /api/v2/tasks/{projecttask}/labels/bulk.
// The body is the full target label set; the call adds missing labels and
// removes any not listed.
//
// Permission topology for testuser1 (see pkg/db/fixtures):
//   - task 1 (project 1): owned by user1 → write. Has label #4 attached.
//   - task 15 (project 6): shared via team 2 read-only → no write.
//   - task 16 (project 7): shared via team 3 with write.
//   - task 34 (project 20): private to user13 → no access.
//
// Labels: #1 own; #3 (user2, attached to no visible task) is invisible to user1.
func TestLabelTaskBulk_V2(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	token := humaTokenFor(t, &testuser1)

	put := func(taskID, body string) (*v2ProblemJSON, []int64, int) {
		t.Helper()
		rec := humaRequest(t, e, http.MethodPut, "/api/v2/tasks/"+taskID+"/labels/bulk", body, token, "")
		if rec.Code >= 400 {
			var p v2ProblemJSON
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &p), "error body: %s", rec.Body.String())
			return &p, nil, rec.Code
		}
		var resp struct {
			Labels []struct {
				ID int64 `json:"id"`
			} `json:"labels"`
		}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp), "body: %s", rec.Body.String())
		ids := make([]int64, 0, len(resp.Labels))
		for _, l := range resp.Labels {
			ids = append(ids, l.ID)
		}
		return nil, ids, rec.Code
	}

	t.Run("Replace adds and removes", func(t *testing.T) {
		// task 1 starts with label #4; replacing with [#1] must add #1 and drop #4.
		p, ids, code := put("1", `{"labels":[{"id":1}]}`)
		require.Nil(t, p)
		assert.Equal(t, http.StatusOK, code)
		assert.ElementsMatch(t, []int64{1}, ids,
			"task 1's labels must be exactly {1} after replace")
	})
	t.Run("Empty list clears all labels", func(t *testing.T) {
		// task 16 (write-shared) gets a label, then an empty replace removes it.
		_, ids, code := put("16", `{"labels":[{"id":1}]}`)
		assert.Equal(t, http.StatusOK, code)
		assert.ElementsMatch(t, []int64{1}, ids)

		p, ids, code := put("16", `{"labels":[]}`)
		require.Nil(t, p)
		assert.Equal(t, http.StatusOK, code)
		assert.Empty(t, ids, "empty replace must remove every label")
	})
	t.Run("Write share can replace", func(t *testing.T) {
		_, ids, code := put("16", `{"labels":[{"id":1}]}`)
		assert.Equal(t, http.StatusOK, code)
		assert.ElementsMatch(t, []int64{1}, ids)
	})
	t.Run("Read-only share is forbidden", func(t *testing.T) {
		p, _, code := put("15", `{"labels":[{"id":1}]}`)
		assert.Equal(t, http.StatusForbidden, code)
		require.NotNil(t, p)
	})
	t.Run("Forbidden task", func(t *testing.T) {
		// task 34 is private to user13.
		p, _, code := put("34", `{"labels":[{"id":1}]}`)
		assert.Equal(t, http.StatusForbidden, code)
		require.NotNil(t, p)
	})
	t.Run("Nonexisting task", func(t *testing.T) {
		p, _, code := put("9999", `{"labels":[{"id":1}]}`)
		assert.Equal(t, http.StatusNotFound, code)
		require.NotNil(t, p)
		assert.Equal(t, models.ErrCodeTaskDoesNotExist, p.Code)
	})
	t.Run("Label the user cannot see is rejected", func(t *testing.T) {
		// label #3 (user2's, attached to no task user1 can see) is invisible to
		// user1; attaching it to a writable task must be refused.
		p, _, code := put("1", `{"labels":[{"id":3}]}`)
		assert.Equal(t, http.StatusForbidden, code)
		require.NotNil(t, p)
		assert.Equal(t, models.ErrCodeUserHasNoAccessToLabel, p.Code)
	})
}
