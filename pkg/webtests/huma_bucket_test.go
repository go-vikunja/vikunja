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

	"code.vikunja.io/api/pkg/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBucket covers the nested kanban-bucket CRUD on /api/v2. Buckets live under
// /projects/{project}/views/{view}/buckets, so the harness binds the project and
// view in basePath and idParam picks {bucket}.
//
// Permission model — Bucket.Can{Create,Update,Delete} all delegate to
// Project.CanUpdate, which resolves to write access (not admin). Bucket.ReadAll
// only needs the view's read access. So write is the boundary for mutation,
// unlike project views where admin is required.
//
// Fixture topology (see pkg/db/fixtures):
//   - project 1 (owned by testuser1), kanban view 4: buckets 1, 2, 3.
//   - project 2 (owned by user3, no share to testuser1), kanban view 8:
//     buckets 4, 40 — the forbidden / non-member negatives.
//   - projects 9/10/11 are owned by user6 and shared to testuser1 read/write/admin;
//     their kanban views 36/40/44 carry buckets {9,25}/{10,26}/{11,27}. The same
//     user exercises every rung by switching the parent path.
func TestHumaBucket(t *testing.T) {
	// project 1 is owned by testuser1.
	owned := webHandlerTestV2{
		user:     &testuser1,
		basePath: "/api/v2/projects/1/views/4/buckets",
		idParam:  "bucket",
		t:        t,
	}
	require.NoError(t, owned.ensureEnv())
	// project 2 is owned by user3; testuser1 has no access. Share owned's Echo
	// instance: each setupTestEnv() regenerates the global JWT signing secret,
	// so two independent harnesses would invalidate each other's tokens.
	forbidden := webHandlerTestV2{
		user:     &testuser1,
		basePath: "/api/v2/projects/2/views/8/buckets",
		idParam:  "bucket",
		t:        t,
		e:        owned.e,
	}
	// project 9 is shared to testuser1 read-only — enough to list, below the
	// write bar mutation requires.
	readShared := webHandlerTestV2{
		user:     &testuser1,
		basePath: "/api/v2/projects/9/views/36/buckets",
		idParam:  "bucket",
		t:        t,
		e:        owned.e,
	}
	// project 10 is shared with write — the rung that clears Project.CanUpdate,
	// so it can create/update/delete buckets.
	writeShared := webHandlerTestV2{
		user:     &testuser1,
		basePath: "/api/v2/projects/10/views/40/buckets",
		idParam:  "bucket",
		t:        t,
		e:        owned.e,
	}
	// project 11 is shared with admin — write access is a subset, so it can do
	// everything too.
	adminShared := webHandlerTestV2{
		user:     &testuser1,
		basePath: "/api/v2/projects/11/views/44/buckets",
		idParam:  "bucket",
		t:        t,
		e:        owned.e,
	}

	t.Run("ReadAll", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := owned.testReadAllWithUser(nil, nil)
			require.NoError(t, err)
			// view 4 has exactly buckets 1, 2, 3 in position order.
			ids, viewIDs := bucketsFromReadAll(t, rec.Body.Bytes())
			assert.ElementsMatch(t, []int64{1, 2, 3}, ids)
			for _, vid := range viewIDs {
				assert.Equal(t, int64(4), vid, "every returned bucket must belong to view 4")
			}
			assert.Contains(t, rec.Body.String(), `"total":3`)
		})
		t.Run("Read-only share can list", func(t *testing.T) {
			// ReadAll only needs the view's read access; a read share suffices.
			rec, err := readShared.testReadAllWithUser(nil, nil)
			require.NoError(t, err)
			ids, _ := bucketsFromReadAll(t, rec.Body.Bytes())
			assert.ElementsMatch(t, []int64{9, 25}, ids)
		})
		t.Run("Write share can list", func(t *testing.T) {
			rec, err := writeShared.testReadAllWithUser(nil, nil)
			require.NoError(t, err)
			ids, _ := bucketsFromReadAll(t, rec.Body.Bytes())
			assert.ElementsMatch(t, []int64{10, 26}, ids)
		})
		t.Run("Forbidden", func(t *testing.T) {
			_, err := forbidden.testReadAllWithUser(nil, nil)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
	})

	t.Run("Create", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := owned.testCreateWithUser(nil, nil, `{"title":"New bucket","limit":5}`)
			require.NoError(t, err)
			assert.Equal(t, http.StatusCreated, rec.Code)
			assert.Contains(t, rec.Body.String(), `"title":"New bucket"`)
			assert.Contains(t, rec.Body.String(), `"limit":5`)
			// ownership: the view from the URL wins over the body.
			assert.Contains(t, rec.Body.String(), `"project_view_id":4`)
		})
		t.Run("Write share can create", func(t *testing.T) {
			// write access clears Project.CanUpdate → Bucket.CanCreate passes.
			rec, err := writeShared.testCreateWithUser(nil, nil, `{"title":"Write made"}`)
			require.NoError(t, err)
			assert.Equal(t, http.StatusCreated, rec.Code)
			assert.Contains(t, rec.Body.String(), `"title":"Write made"`)
			assert.Contains(t, rec.Body.String(), `"project_view_id":40`)
		})
		t.Run("Admin share can create", func(t *testing.T) {
			rec, err := adminShared.testCreateWithUser(nil, nil, `{"title":"Admin made"}`)
			require.NoError(t, err)
			assert.Equal(t, http.StatusCreated, rec.Code)
			assert.Contains(t, rec.Body.String(), `"title":"Admin made"`)
			assert.Contains(t, rec.Body.String(), `"project_view_id":44`)
		})
		t.Run("Read share cannot create", func(t *testing.T) {
			// read share is below the write bar Bucket.CanCreate enforces.
			_, err := readShared.testCreateWithUser(nil, nil, `{"title":"Nope"}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Forbidden", func(t *testing.T) {
			_, err := forbidden.testCreateWithUser(nil, nil, `{"title":"Nope"}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Empty title", func(t *testing.T) {
			// Title has valid:"required" / minLength:"1" → 422 before the model.
			_, err := owned.testCreateWithUser(nil, nil, `{"title":""}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusUnprocessableEntity, getHTTPErrorCode(err))
		})
	})

	t.Run("Update", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := owned.testUpdateWithUser(nil, map[string]string{"bucket": "1"}, `{"title":"Renamed bucket"}`)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"title":"Renamed bucket"`)
			assert.Contains(t, rec.Body.String(), `"id":1`)
			// Only the sent fields are written: the server-managed creator and the
			// view scoping from the URL are preserved, not clobbered to zero.
			db.AssertExists(t, "buckets", map[string]interface{}{
				"id":              1,
				"title":           "Renamed bucket",
				"project_view_id": 4,
				"created_by_id":   1,
			}, false)
		})
		t.Run("Write share can update", func(t *testing.T) {
			// bucket 10 belongs to view 40 (project 10, write share).
			rec, err := writeShared.testUpdateWithUser(nil, map[string]string{"bucket": "10"}, `{"title":"Write renamed"}`)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"title":"Write renamed"`)
			assert.Contains(t, rec.Body.String(), `"id":10`)
		})
		t.Run("Read share cannot update", func(t *testing.T) {
			// bucket 9 belongs to view 36 (project 9, read share) → needs write.
			_, err := readShared.testUpdateWithUser(nil, map[string]string{"bucket": "9"}, `{"title":"x"}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Nonexisting", func(t *testing.T) {
			_, err := owned.testUpdateWithUser(nil, map[string]string{"bucket": "9999"}, `{"title":"x"}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
		})
		t.Run("Forbidden", func(t *testing.T) {
			// bucket 4 belongs to view 8 (project 2) — testuser1 has no access.
			_, err := forbidden.testUpdateWithUser(nil, map[string]string{"bucket": "4"}, `{"title":"x"}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
	})

	t.Run("Delete", func(t *testing.T) {
		t.Run("Read share cannot delete", func(t *testing.T) {
			// bucket 25 belongs to view 36 (project 9, read share) → needs write.
			_, err := readShared.testDeleteWithUser(nil, map[string]string{"bucket": "25"})
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Write share can delete", func(t *testing.T) {
			// bucket 26 belongs to view 40 (project 10, write share); view 40 still
			// has bucket 10 (plus the one created above), so it isn't the last.
			rec, err := writeShared.testDeleteWithUser(nil, map[string]string{"bucket": "26"})
			require.NoError(t, err)
			assert.Equal(t, http.StatusNoContent, rec.Code)
			assert.Empty(t, rec.Body.String())
		})
		t.Run("Forbidden", func(t *testing.T) {
			_, err := forbidden.testDeleteWithUser(nil, map[string]string{"bucket": "40"})
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Normal", func(t *testing.T) {
			// view 4 has buckets 1, 2, 3 (plus the one created above), so deleting
			// bucket 2 leaves more than one behind.
			rec, err := owned.testDeleteWithUser(nil, map[string]string{"bucket": "2"})
			require.NoError(t, err)
			assert.Equal(t, http.StatusNoContent, rec.Code)
			assert.Empty(t, rec.Body.String())
			db.AssertMissing(t, "buckets", map[string]interface{}{"id": 2})
		})
		t.Run("Nonexisting", func(t *testing.T) {
			_, err := owned.testDeleteWithUser(nil, map[string]string{"bucket": "9999"})
			require.Error(t, err)
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
		})
	})
}

// bucketsFromReadAll extracts the bucket ids and their project_view_ids from a v2
// paginated list body so the visible set can be asserted exactly.
func bucketsFromReadAll(t *testing.T, body []byte) (ids []int64, viewIDs []int64) {
	t.Helper()
	var resp struct {
		Items []struct {
			ID            int64 `json:"id"`
			ProjectViewID int64 `json:"project_view_id"`
		} `json:"items"`
	}
	require.NoError(t, json.Unmarshal(body, &resp), "ReadAll body must be a paginated envelope: %s", string(body))
	ids = make([]int64, 0, len(resp.Items))
	viewIDs = make([]int64, 0, len(resp.Items))
	for _, it := range resp.Items {
		ids = append(ids, it.ID)
		viewIDs = append(viewIDs, it.ProjectViewID)
	}
	return ids, viewIDs
}
