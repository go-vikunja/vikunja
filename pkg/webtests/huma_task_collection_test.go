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
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// decodePaginatedTaskItems pulls the items slice out of a Paginated[*Task]
// response so length assertions don't have to regex over nested task JSON.
func decodePaginatedTaskItems(t *testing.T, rec *httptest.ResponseRecorder) []json.RawMessage {
	t.Helper()
	var body struct {
		Items []json.RawMessage `json:"items"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	return body.Items
}

// TestHumaTaskCollection covers the v2 task-list endpoints. v2 splits v1's
// single polymorphic /tasks endpoint into flat-task endpoints (always []*Task,
// paginated) and a dedicated buckets-with-tasks endpoint (always []*Bucket).
// Mirrors v1's TestTaskCollection where the surface overlaps.
func TestHumaTaskCollection(t *testing.T) {
	h := webHandlerTestV2{user: &testuser1, t: t}
	require.NoError(t, h.ensureEnv())
	tok := humaTokenFor(t, &testuser1)

	get := func(path string) *httptest.ResponseRecorder {
		return humaRequest(t, h.e, http.MethodGet, path, "", tok, "")
	}

	t.Run("project-scoped", func(t *testing.T) {
		t.Run("returns the project's tasks", func(t *testing.T) {
			rec := get("/api/v2/projects/1/tasks")
			require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
			body := rec.Body.String()
			assert.Contains(t, body, `"items":[`)
			assert.Contains(t, body, `task #1`)
			assert.Contains(t, body, `task #12`)
			assert.NotContains(t, body, `task #13`) // other project
			assert.NotContains(t, body, `task #14`)
		})
		t.Run("forbidden project", func(t *testing.T) {
			// Project 2 is inaccessible to user1.
			rec := get("/api/v2/projects/2/tasks")
			assert.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
		})
		t.Run("nonexistent project", func(t *testing.T) {
			rec := get("/api/v2/projects/99999/tasks")
			assert.Equal(t, http.StatusNotFound, rec.Code, "body: %s", rec.Body.String())
		})
		t.Run("pagination", func(t *testing.T) {
			rec := get("/api/v2/projects/1/tasks?page=1&per_page=2")
			require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
			assert.Len(t, decodePaginatedTaskItems(t, rec), 2, "per_page caps the page to two tasks")
			body := rec.Body.String()
			assert.Contains(t, body, `"page":1`)
			assert.Contains(t, body, `"per_page":2`)
		})
		t.Run("filter", func(t *testing.T) {
			rec := get("/api/v2/projects/1/tasks?filter=" +
				"start_date%20%3E%20%272018-12-11T03%3A46%3A40%2B00%3A00%27%20%7C%7C%20" +
				"end_date%20%3C%20%272018-12-13T11%3A20%3A01%2B00%3A00%27%20%7C%7C%20" +
				"due_date%20%3E%20%272018-11-29T14%3A00%3A00%2B00%3A00%27")
			require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
			body := rec.Body.String()
			assert.NotContains(t, body, `task #1`)
			assert.Contains(t, body, `task #5 `)
			assert.Contains(t, body, `task #6 `)
			assert.NotContains(t, body, `task #10`)
		})
		t.Run("invalid filter value", func(t *testing.T) {
			// ErrInvalidTaskFilterValue carries an explicit 400; only govalidator
			// failures map to 422 in v2.
			rec := get("/api/v2/projects/1/tasks?filter=due_date%20%3E%20invalid")
			assert.Equal(t, http.StatusBadRequest, rec.Code, "body: %s", rec.Body.String())
		})
	})

	t.Run("search via q", func(t *testing.T) {
		// Only task #6 has the word "unique" in its description.
		rec := get("/api/v2/projects/1/tasks?q=unique")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		body := rec.Body.String()
		assert.Contains(t, body, `task #6 `)
		assert.NotContains(t, body, `task #1`)
		assert.NotContains(t, body, `task #2 `)
	})

	t.Run("sort by repeated params", func(t *testing.T) {
		// Two sort_by + two order_by prove ,explode binds every value.
		rec := get("/api/v2/projects/1/tasks?sort_by=priority&sort_by=id&order_by=desc&order_by=asc")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		// task #3 has priority 100, the highest; desc puts it first.
		assert.Regexp(t, `"items":\[\{"id":3,`, rec.Body.String())
	})

	t.Run("invalid sort field", func(t *testing.T) {
		// A 400 (not 200) proves sort_by binds: the model validated the field
		// and rejected it. ErrInvalidTaskField carries an explicit 400.
		rec := get("/api/v2/projects/1/tasks?sort_by=loremipsum")
		assert.Equal(t, http.StatusBadRequest, rec.Code, "body: %s", rec.Body.String())
	})

	t.Run("cross-project", func(t *testing.T) {
		// /tasks returns tasks from every project the user can see, including
		// shared ones, but not tasks in projects they have no access to.
		rec := get("/api/v2/tasks")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		body := rec.Body.String()
		assert.Contains(t, body, `task #1`)     // own project
		assert.Contains(t, body, `task #15`)    // shared via team readonly
		assert.Contains(t, body, `task #21`)    // shared via parent project team
		assert.NotContains(t, body, `task #13`) // no access
		assert.NotContains(t, body, `task #14`)
	})

	t.Run("view-scoped", func(t *testing.T) {
		t.Run("list view returns flat tasks", func(t *testing.T) {
			rec := get("/api/v2/projects/1/views/1/tasks")
			require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
			body := rec.Body.String()
			assert.Contains(t, body, `task #1`)
			assert.NotContains(t, body, `testbucket`) // not buckets
		})
		t.Run("kanban view still returns flat tasks", func(t *testing.T) {
			// View 4 is project 1's kanban view. v1 would return buckets here;
			// v2's tasks endpoint forces flat tasks.
			rec := get("/api/v2/projects/1/views/4/tasks")
			require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
			body := rec.Body.String()
			assert.Contains(t, body, `"items":[`)
			assert.Contains(t, body, `task #1`)
			assert.NotContains(t, body, `testbucket`)
		})
		t.Run("forbidden view", func(t *testing.T) {
			// Project 2 (and its view 8) is inaccessible to user1.
			rec := get("/api/v2/projects/2/views/8/tasks")
			assert.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
		})
	})

	t.Run("saved filter project", func(t *testing.T) {
		// Project -2 maps to saved filter #1, whose stored filter matches the
		// date-range tasks. Recurses inside the model.
		rec := get("/api/v2/projects/-2/tasks")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		body := rec.Body.String()
		assert.Contains(t, body, `task #5 `)
		assert.Contains(t, body, `task #6 `)
		assert.NotContains(t, body, `task #1`)
		assert.NotContains(t, body, `task #10`)
	})
}

// TestHumaTaskCollection_Expand proves expand binds every repeated value
// (,explode) and routes through parseTaskExpand.
func TestHumaTaskCollection_Expand(t *testing.T) {
	h := webHandlerTestV2{user: &testuser1, t: t}
	require.NoError(t, h.ensureEnv())
	tok := humaTokenFor(t, &testuser1)

	get := func(path string) *httptest.ResponseRecorder {
		return humaRequest(t, h.e, http.MethodGet, path, "", tok, "")
	}

	t.Run("repeated expand applies every value", func(t *testing.T) {
		rec := get("/api/v2/projects/1/tasks?expand=comment_count&expand=reactions")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		body := rec.Body.String()
		assert.Contains(t, body, `"comment_count":`)
		assert.Contains(t, body, `"reactions":`)
	})
	t.Run("invalid expand rejected", func(t *testing.T) {
		rec := get("/api/v2/projects/1/tasks?expand=bogus")
		// enum on the query param makes Huma reject before the handler.
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code, "body: %s", rec.Body.String())
	})
}

// TestHumaTaskCollection_Buckets covers the dedicated buckets-with-tasks
// endpoint: a kanban view returns []*Bucket with each bucket's tasks populated,
// not paginated.
func TestHumaTaskCollection_Buckets(t *testing.T) {
	h := webHandlerTestV2{user: &testuser1, t: t}
	require.NoError(t, h.ensureEnv())
	tok := humaTokenFor(t, &testuser1)

	get := func(path string) *httptest.ResponseRecorder {
		return humaRequest(t, h.e, http.MethodGet, path, "", tok, "")
	}

	t.Run("kanban view returns buckets with tasks", func(t *testing.T) {
		rec := get("/api/v2/projects/1/views/4/buckets/tasks")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		body := rec.Body.String()
		assert.Contains(t, body, `testbucket1`)
		assert.Contains(t, body, `testbucket2`)
		assert.Contains(t, body, `testbucket3`)
		assert.NotContains(t, body, `testbucket4`) // belongs to project 2's view
		// Tasks are nested under their bucket, not at the top level.
		assert.Contains(t, body, `"tasks":[`)
		assert.Contains(t, body, `task #1`)
		// total counts buckets, not tasks.
		assert.Contains(t, body, `"total":3`)
	})

	t.Run("forbidden project", func(t *testing.T) {
		rec := get("/api/v2/projects/2/views/8/buckets/tasks")
		assert.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
	})

	t.Run("non-kanban view is a 400, not a 500", func(t *testing.T) {
		// View 1 is project 1's list view; it has no bucket configuration, so
		// the model returns flat tasks and the handler refuses cleanly.
		rec := get("/api/v2/projects/1/views/1/buckets/tasks")
		assert.Equal(t, http.StatusBadRequest, rec.Code, "body: %s", rec.Body.String())
	})

	t.Run("static tasks segment does not collide with the bucket-update route", func(t *testing.T) {
		// PUT .../buckets/{bucket}/tasks exists; GET .../buckets/tasks must hit
		// this handler, not parse "tasks" as a bucket id.
		rec := get("/api/v2/projects/1/views/4/buckets/tasks")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), `testbucket1`)
	})
}
