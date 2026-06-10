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
	"net/http/httptest"
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHumaTask mirrors v1's TestTask so v2 contract parity is readable
// side-by-side. Read/update/delete address a task by its numeric id; create
// and by-index live on project-scoped paths that don't fit the harness's
// basePath/{id} shape, so those use humaRequest against a shared env.
//
// Fixture topology the matrix relies on (pkg/db/fixtures/tasks.yml +
// project shares):
//   - #1: user1's own task in project 1 (admin) — readable/updatable/deletable.
//   - #14: project shared read-only via team — forbidden to write/delete.
//   - #34: project 20, private to user13 — invisible to user1.
//   - project 6: shared read-only; project 7/8: shared write/admin via team.
func TestHumaTask(t *testing.T) {
	testHandler := webHandlerTestV2{
		user:     &testuser1,
		basePath: "/api/v2/tasks",
		idParam:  "projecttask",
		t:        t,
	}

	t.Run("ReadOne", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := testHandler.testReadOneWithUser(nil, map[string]string{"projecttask": "1"})
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"id":1`)
			assert.Contains(t, rec.Body.String(), `"title":"task #1"`)
			assert.Contains(t, rec.Body.String(), `"max_permission":2`) // owner = admin
			assert.NotEmpty(t, rec.Result().Header.Get("ETag"))
		})
		t.Run("Nonexisting", func(t *testing.T) {
			_, err := testHandler.testReadOneWithUser(nil, map[string]string{"projecttask": "99999"})
			require.Error(t, err)
			// CanRead resolves the task before the project check, so a missing
			// task surfaces as 404, not the 403 the label read uses.
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
		})
		t.Run("Forbidden - private project", func(t *testing.T) {
			// Task #34 lives in project 20, private to user13.
			_, err := testHandler.testReadOneWithUser(nil, map[string]string{"projecttask": "34"})
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
	})

	// The v2 harness loads fixtures once and reuses the env across subtests,
	// so each mutating subtest targets a distinct task to stay order-independent
	// (unlike v1's webHandlerTest, which reloads fixtures per request).
	t.Run("Update", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"projecttask": "3"}, `{"title":"Lorem Ipsum"}`)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
			assert.NotContains(t, rec.Body.String(), `"title":"task #3 high prio"`)
		})
		t.Run("Move to another project", func(t *testing.T) {
			rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"projecttask": "4"}, `{"project_id":7}`)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"project_id":7`)
		})
		t.Run("Nonexisting", func(t *testing.T) {
			_, err := testHandler.testUpdateWithUser(nil, map[string]string{"projecttask": "99999"}, `{"title":"x"}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
		})
		t.Run("Forbidden - read-only share", func(t *testing.T) {
			_, err := testHandler.testUpdateWithUser(nil, map[string]string{"projecttask": "14"}, `{"title":"x"}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Forbidden - move into a project the user can't write", func(t *testing.T) {
			_, err := testHandler.testUpdateWithUser(nil, map[string]string{"projecttask": "5"}, `{"project_id":20}`)
			require.Error(t, err)
			assertHandlerErrorCode(t, err, models.ErrorCodeGenericForbidden)
		})
	})

	t.Run("Delete", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := testHandler.testDeleteWithUser(nil, map[string]string{"projecttask": "2"})
			require.NoError(t, err)
			assert.Equal(t, http.StatusNoContent, rec.Code)
			assert.Empty(t, rec.Body.String())
		})
		t.Run("Nonexisting", func(t *testing.T) {
			_, err := testHandler.testDeleteWithUser(nil, map[string]string{"projecttask": "99999"})
			require.Error(t, err)
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
		})
		t.Run("Forbidden - read-only share", func(t *testing.T) {
			_, err := testHandler.testDeleteWithUser(nil, map[string]string{"projecttask": "14"})
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Shared via team write", func(t *testing.T) {
			rec, err := testHandler.testDeleteWithUser(nil, map[string]string{"projecttask": "16"})
			require.NoError(t, err)
			assert.Equal(t, http.StatusNoContent, rec.Code)
		})
	})
}

// TestHumaTask_Create covers the project-scoped create path, which the harness
// basePath shape can't express. Mirrors v1's TestTask/Create matrix.
func TestHumaTask_Create(t *testing.T) {
	h := webHandlerTestV2{user: &testuser1, t: t}
	require.NoError(t, h.ensureEnv())

	create := func(project, body string) *httptest.ResponseRecorder {
		return humaRequest(t, h.e, http.MethodPost, "/api/v2/projects/"+project+"/tasks", body, humaTokenFor(t, &testuser1), "")
	}

	t.Run("Normal", func(t *testing.T) {
		rec := create("1", `{"title":"Lorem Ipsum"}`)
		require.Equal(t, http.StatusCreated, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
		assert.Contains(t, rec.Body.String(), `"project_id":1`)
	})
	t.Run("Project id from body is ignored - URL wins", func(t *testing.T) {
		rec := create("1", `{"title":"url wins","project_id":7}`)
		require.Equal(t, http.StatusCreated, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"project_id":1`)
		assert.NotContains(t, rec.Body.String(), `"project_id":7`)
	})
	t.Run("Nonexisting project", func(t *testing.T) {
		rec := create("9999", `{"title":"x"}`)
		assert.Equal(t, http.StatusNotFound, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"code":%d`, models.ErrCodeProjectDoesNotExist))
	})
	t.Run("Forbidden - private project", func(t *testing.T) {
		rec := create("20", `{"title":"x"}`)
		assert.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
	})
	t.Run("Forbidden - read-only share", func(t *testing.T) {
		rec := create("6", `{"title":"x"}`)
		assert.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
	})
	t.Run("Shared via team write", func(t *testing.T) {
		rec := create("7", `{"title":"Lorem Ipsum"}`)
		require.Equal(t, http.StatusCreated, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
	})
	t.Run("Empty title is rejected", func(t *testing.T) {
		rec := create("1", `{"title":""}`)
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code, "body: %s", rec.Body.String())
	})
}

// TestHumaTask_ReadByIndex covers the by-index route, including the textual
// project-identifier resolution that v1 does in echo middleware. Mirrors v1's
// TestTaskByProjectIndex and TestTask/ReadOneByIndex.
func TestHumaTask_ReadByIndex(t *testing.T) {
	h := webHandlerTestV2{user: &testuser1, t: t}
	require.NoError(t, h.ensureEnv())

	get := func(project, index string) *httptest.ResponseRecorder {
		return humaRequest(t, h.e, http.MethodGet,
			fmt.Sprintf("/api/v2/projects/%s/tasks/by-index/%s", project, index), "", humaTokenFor(t, &testuser1), "")
	}

	t.Run("By numeric project id", func(t *testing.T) {
		rec := get("1", "1")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"id":1`)
		assert.Contains(t, rec.Body.String(), `"index":1`)
	})
	t.Run("By textual project identifier", func(t *testing.T) {
		// Project 1 has identifier "TEST1".
		rec := get("TEST1", "1")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"id":1`)
	})
	t.Run("Identifier match is case-insensitive", func(t *testing.T) {
		rec := get("test1", "1")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"id":1`)
	})
	t.Run("Unknown identifier returns ErrProjectDoesNotExist", func(t *testing.T) {
		rec := get("does-not-exist", "1")
		assert.Equal(t, http.StatusNotFound, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"code":%d`, models.ErrCodeProjectDoesNotExist))
	})
	t.Run("Nonexistent index returns 404", func(t *testing.T) {
		rec := get("1", "99999")
		assert.Equal(t, http.StatusNotFound, rec.Code, "body: %s", rec.Body.String())
	})
	t.Run("No permission returns 403", func(t *testing.T) {
		// Project 2 is inaccessible to user1; must be 403, not a 404 oracle.
		rec := get("2", "1")
		assert.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
	})
}

// TestHumaTask_Expand asserts the expand query param populates the extra,
// more expensive fields, is repeatable (explode), and rejects unknown values.
// comment_count and reactions are genuinely gated on the flag, so they prove
// the param is wired through; subtasks-as-related-tasks load regardless.
func TestHumaTask_Expand(t *testing.T) {
	h := webHandlerTestV2{user: &testuser1, t: t}
	require.NoError(t, h.ensureEnv())
	tok := humaTokenFor(t, &testuser1)

	t.Run("absent leaves expand-gated fields empty", func(t *testing.T) {
		rec := humaRequest(t, h.e, http.MethodGet, "/api/v2/tasks/1", "", tok, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.NotContains(t, rec.Body.String(), `"comment_count":`)
		assert.NotContains(t, rec.Body.String(), `"reactions":{`)
	})
	t.Run("comment_count", func(t *testing.T) {
		rec := humaRequest(t, h.e, http.MethodGet, "/api/v2/tasks/1?expand=comment_count", "", tok, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"comment_count":`, "comment_count must be present: %s", rec.Body.String())
	})
	t.Run("reactions", func(t *testing.T) {
		// Task #1 has reaction fixture #1.
		rec := humaRequest(t, h.e, http.MethodGet, "/api/v2/tasks/1?expand=reactions", "", tok, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"reactions":{`, "reactions must be embedded: %s", rec.Body.String())
	})
	t.Run("repeated param applies every value", func(t *testing.T) {
		// explode binding: both ?expand= values take effect, not just the first.
		rec := humaRequest(t, h.e, http.MethodGet, "/api/v2/tasks/1?expand=comment_count&expand=reactions", "", tok, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"comment_count":`)
		assert.Contains(t, rec.Body.String(), `"reactions":{`)
	})
	t.Run("invalid value is rejected", func(t *testing.T) {
		rec := humaRequest(t, h.e, http.MethodGet, "/api/v2/tasks/1?expand=bogus", "", tok, "")
		// enum on the query param makes Huma reject it as a 422 before the handler.
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code, "body: %s", rec.Body.String())
	})
}

// TestHumaTask_ETagReturns304 covers the v2-only conditional-read behaviour.
func TestHumaTask_ETagReturns304(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	token := humaTokenFor(t, &testuser1)

	rec := humaRequest(t, e, http.MethodGet, "/api/v2/tasks/1", "", token, "")
	require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
	etag := rec.Header().Get("ETag")
	require.NotEmpty(t, etag)

	req := httptest.NewRequest(http.MethodGet, "/api/v2/tasks/1", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("If-None-Match", etag)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNotModified, rec.Code, "body: %s", rec.Body.String())
}
