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

// TestLabelTask is the nested-path test for labels-on-a-task under
// /api/v2/tasks/{projecttask}/labels. It ports the full v1 model-level matrix
// from pkg/models/label_task_test.go so the v2 HTTP surface independently
// proves the permission contract once v1's routes are removed.
//
// Permission topology for testuser1 (see pkg/db/fixtures):
//   - task 1 (project 1): owned by user1 → write/admin. Has label #4 attached.
//   - task 14 (project 5): no access → forbidden.
//   - task 15 (project 6): shared via team 2 read-only → readable, not writable.
//   - task 16 (project 7): shared via team 3 with write → writable.
//   - task 34 (project 20): private to user13 → no access.
//
// Labels user1 may attach: #1 (own) and #4 (visible via accessible task 1).
// Label #9999 does not exist → no access → forbidden on attach.
func TestLabelTask(t *testing.T) {
	// task 1 is owned by testuser1.
	owned := webHandlerTestV2{
		user:     &testuser1,
		basePath: "/api/v2/tasks/1/labels",
		idParam:  "label",
		t:        t,
	}
	require.NoError(t, owned.ensureEnv())
	// Each setupTestEnv() regenerates the JWT signing secret, so every harness
	// below must reuse owned's Echo instance to keep its token valid.
	forbidden := webHandlerTestV2{
		user:     &testuser1,
		basePath: "/api/v2/tasks/14/labels",
		idParam:  "label",
		t:        t,
		e:        owned.e,
	}
	readShared := webHandlerTestV2{
		user:     &testuser1,
		basePath: "/api/v2/tasks/15/labels",
		idParam:  "label",
		t:        t,
		e:        owned.e,
	}
	writeShared := webHandlerTestV2{
		user:     &testuser1,
		basePath: "/api/v2/tasks/16/labels",
		idParam:  "label",
		t:        t,
		e:        owned.e,
	}
	private := webHandlerTestV2{
		user:     &testuser1,
		basePath: "/api/v2/tasks/34/labels",
		idParam:  "label",
		t:        t,
		e:        owned.e,
	}

	t.Run("ReadAll", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := owned.testReadAllWithUser(nil, nil)
			require.NoError(t, err)
			// task 1 has exactly label #4 attached.
			ids := labelTaskIDsFromReadAll(t, rec.Body.Bytes())
			assert.ElementsMatch(t, []int64{4}, ids,
				"ReadAll on task 1 must return exactly {4}; body: %s", rec.Body.String())
		})
		t.Run("Read-only share can list", func(t *testing.T) {
			// ReadAll only requires read access to the task; a read share suffices.
			_, err := readShared.testReadAllWithUser(nil, nil)
			require.NoError(t, err)
		})
		t.Run("Search", func(t *testing.T) {
			rec, err := owned.testReadAllWithUser(map[string][]string{"q": {"4"}}, nil)
			require.NoError(t, err)
			ids := labelTaskIDsFromReadAll(t, rec.Body.Bytes())
			assert.ElementsMatch(t, []int64{4}, ids)
		})
		t.Run("Forbidden", func(t *testing.T) {
			_, err := forbidden.testReadAllWithUser(nil, nil)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Nonexisting task", func(t *testing.T) {
			noTask := webHandlerTestV2{user: &testuser1, basePath: "/api/v2/tasks/9999/labels", idParam: "label", t: t, e: owned.e}
			_, err := noTask.testReadAllWithUser(nil, nil)
			require.Error(t, err)
			assertHandlerErrorCode(t, err, models.ErrCodeTaskDoesNotExist)
		})
	})

	t.Run("Create", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := owned.testCreateWithUser(nil, nil, `{"label_id":1}`)
			require.NoError(t, err)
			assert.Equal(t, http.StatusCreated, rec.Code)
			assert.Contains(t, rec.Body.String(), `"label_id":1`)
		})
		t.Run("Write share can attach", func(t *testing.T) {
			// task 16 is write-shared; user1 has access to label #1.
			rec, err := writeShared.testCreateWithUser(nil, nil, `{"label_id":1}`)
			require.NoError(t, err)
			assert.Equal(t, http.StatusCreated, rec.Code)
		})
		t.Run("Already on task", func(t *testing.T) {
			// label #4 is already attached to task 1.
			_, err := owned.testCreateWithUser(nil, nil, `{"label_id":4}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusBadRequest, getHTTPErrorCode(err))
			assertHandlerErrorCode(t, err, models.ErrCodeLabelIsAlreadyOnTask)
		})
		t.Run("Nonexisting label", func(t *testing.T) {
			// CanCreate looks the label up first, so a missing label surfaces as
			// 404 (label not found) rather than a permission error.
			_, err := owned.testCreateWithUser(nil, nil, `{"label_id":9999}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
			assertHandlerErrorCode(t, err, models.ErrCodeLabelDoesNotExist)
		})
		t.Run("Read-only share cannot attach", func(t *testing.T) {
			// task 15 is read-only shared → CanCreate needs write to the task.
			_, err := readShared.testCreateWithUser(nil, nil, `{"label_id":1}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Forbidden task", func(t *testing.T) {
			_, err := private.testCreateWithUser(nil, nil, `{"label_id":1}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
	})

	t.Run("Delete", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			// label #4 is attached to task 1.
			rec, err := owned.testDeleteWithUser(nil, map[string]string{"label": "4"})
			require.NoError(t, err)
			assert.Equal(t, http.StatusNoContent, rec.Code)
			assert.Empty(t, rec.Body.String())
		})
		t.Run("Nonexisting relation", func(t *testing.T) {
			// label #2 (own, never attached by any subtest) is not on task 1 →
			// CanDelete requires the relation to exist, so it refuses.
			_, err := owned.testDeleteWithUser(nil, map[string]string{"label": "2"})
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Read-only share cannot detach", func(t *testing.T) {
			_, err := readShared.testDeleteWithUser(nil, map[string]string{"label": "1"})
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Forbidden task", func(t *testing.T) {
			_, err := private.testDeleteWithUser(nil, map[string]string{"label": "1"})
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
	})
}

// labelTaskIDsFromReadAll extracts the label IDs from a v2 paginated list body
// so the attached set can be asserted exactly rather than via substring match.
func labelTaskIDsFromReadAll(t *testing.T, body []byte) []int64 {
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
