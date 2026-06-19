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

// TestTaskBucketV2 covers PUT /projects/{project}/views/{view}/buckets/{bucket}/tasks.
// It drives the Echo+Huma stack directly (humaRequest/humaTokenFor) because the
// route is an action sub-path webHandlerTestV2's buildURL doesn't model. Fixtures
// (project 1, view 4): bucket 1 default, bucket 2 "Doing" limit 3 (full), bucket 3 done.
func TestTaskBucketV2(t *testing.T) {
	const path = "/api/v2/projects/1/views/4/buckets/%d/tasks"

	t.Run("moves a task into a bucket", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, &testuser1)

		// Task 3 starts in bucket 2; move it into bucket 1 (neither full nor done).
		rec := humaRequest(t, e, http.MethodPut, fmt.Sprintf(path, 1), `{"task_id":3}`, token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"task_id":3`)

		db.AssertExists(t, "task_buckets", map[string]interface{}{
			"task_id":   3,
			"bucket_id": 1,
		}, false)
	})

	t.Run("moving a task into the done bucket marks it done", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, &testuser1)

		// Bucket 3 is the done bucket on view 4; task 1 is not yet done.
		rec := humaRequest(t, e, http.MethodPut, fmt.Sprintf(path, 3), `{"task_id":1}`, token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"done":true`)

		db.AssertExists(t, "tasks", map[string]interface{}{
			"id":   1,
			"done": true,
		}, false)
		db.AssertExists(t, "task_buckets", map[string]interface{}{
			"task_id":   1,
			"bucket_id": 3,
		}, false)
	})

	t.Run("moving a task out of the done bucket un-marks it done", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, &testuser1)

		// Task 2 starts in bucket 3 (done) and is done; move it to bucket 1.
		rec := humaRequest(t, e, http.MethodPut, fmt.Sprintf(path, 1), `{"task_id":2}`, token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"done":false`)

		db.AssertExists(t, "tasks", map[string]interface{}{
			"id":   2,
			"done": false,
		}, false)
		db.AssertExists(t, "task_buckets", map[string]interface{}{
			"task_id":   2,
			"bucket_id": 1,
		}, false)
	})

	t.Run("full bucket is rejected", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, &testuser1)

		// Bucket 2 already holds 3 tasks and has a limit of 3.
		rec := humaRequest(t, e, http.MethodPut, fmt.Sprintf(path, 2), `{"task_id":1}`, token, "")
		require.Equal(t, http.StatusPreconditionFailed, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"code":%d`, models.ErrCodeBucketLimitExceeded))
	})

	t.Run("bucket on another view is rejected", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, &testuser1)

		// Bucket 4 lives on view 8 (project 2), so under view 4 / project 1 the
		// permission check resolves the bucket's own view scoped by the path
		// project and finds none → 404 before the move's own 400 can fire.
		rec := humaRequest(t, e, http.MethodPut, fmt.Sprintf(path, 4), `{"task_id":1}`, token, "")
		require.Equal(t, http.StatusNotFound, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"code":%d`, models.ErrCodeProjectViewDoesNotExist))
	})

	t.Run("no write access is forbidden", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		// testuser15 has no access to project 1.
		token := humaTokenFor(t, &testuser15)

		rec := humaRequest(t, e, http.MethodPut, fmt.Sprintf(path, 1), `{"task_id":1}`, token, "")
		require.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
	})
}
