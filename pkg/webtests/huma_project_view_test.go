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
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProjectView is the nested-path reference test for /api/v2. Views live
// under /projects/{project}/views/{view}, so the harness binds two path params:
// basePath carries the literal {project} and idParam picks {view}.
//
// Fixtures (see pkg/db/fixtures): project 1 (owned by testuser1) has the four
// default views 1-4 plus the filtered view 161 (five total); project 2 (owned
// by user3, no share to testuser1) has views 5-8 — used for the forbidden and
// wrong-parent negatives.
//
// Permission gradient — ProjectView.Can* delegate to Project.CanRead (reads)
// and Project.IsAdmin (create/update/delete). Projects 9/10/11 are owned by
// user6 and shared to testuser1 with read/write/admin respectively, so the
// same user exercises every rung of the gradient by switching the parent path:
//   - read share  (project 9, views 33-36): CAN list/read, CANNOT create/update/delete
//   - write share (project 10, views 37-40): CAN list/read, CANNOT create/update/delete
//   - admin share (project 11, views 41-44): CAN do everything
func TestProjectView(t *testing.T) {
	// project 1 is owned by testuser1.
	owned := webHandlerTestV2{
		user:     &testuser1,
		basePath: "/api/v2/projects/1/views",
		idParam:  "view",
		t:        t,
	}
	require.NoError(t, owned.ensureEnv())
	// project 2 is owned by user3; testuser1 has no access. Share owned's Echo
	// instance: each setupTestEnv() regenerates the global JWT signing secret,
	// so two independent harnesses would invalidate each other's tokens.
	forbidden := webHandlerTestV2{
		user:     &testuser1,
		basePath: "/api/v2/projects/2/views",
		idParam:  "view",
		t:        t,
		e:        owned.e,
	}
	// project 9 is shared to testuser1 read-only; project 10 with write — both
	// below the admin bar Can{Create,Update,Delete} require. They reuse owned's
	// Echo so the JWT secret stays valid.
	readShared := webHandlerTestV2{
		user:     &testuser1,
		basePath: "/api/v2/projects/9/views",
		idParam:  "view",
		t:        t,
		e:        owned.e,
	}
	writeShared := webHandlerTestV2{
		user:     &testuser1,
		basePath: "/api/v2/projects/10/views",
		idParam:  "view",
		t:        t,
		e:        owned.e,
	}
	// project 11 is shared to testuser1 with admin — the only share rung that
	// clears the Project.IsAdmin gate the write methods enforce.
	adminShared := webHandlerTestV2{
		user:     &testuser1,
		basePath: "/api/v2/projects/11/views",
		idParam:  "view",
		t:        t,
		e:        owned.e,
	}

	t.Run("ReadAll", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := owned.testReadAllWithUser(nil, nil)
			require.NoError(t, err)
			// project 1's four default views, none from project 2.
			assert.Contains(t, rec.Body.String(), `"title":"List"`)
			assert.Contains(t, rec.Body.String(), `"title":"Gantt"`)
			assert.Contains(t, rec.Body.String(), `"title":"Table"`)
			assert.Contains(t, rec.Body.String(), `"title":"Kanban"`)
			assert.Contains(t, rec.Body.String(), `"project_id":1`)
			assert.NotContains(t, rec.Body.String(), `"project_id":2`)

			// Exact cardinality: project 1 has five views in the fixtures (the
			// four defaults 1-4 plus the filtered view 161). The list must
			// surface every one of them and nothing else.
			var env struct {
				Items []struct {
					ID        int64  `json:"id"`
					ProjectID int64  `json:"project_id"`
					Title     string `json:"title"`
				} `json:"items"`
				Total int64 `json:"total"`
			}
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &env))
			assert.Len(t, env.Items, 5)
			assert.Equal(t, int64(5), env.Total)
			ids := make([]int64, 0, len(env.Items))
			for _, v := range env.Items {
				assert.Equal(t, int64(1), v.ProjectID, "every returned view must belong to project 1")
				ids = append(ids, v.ID)
			}
			assert.ElementsMatch(t, []int64{1, 2, 3, 4, 161}, ids)
		})
		t.Run("Read-only share can list", func(t *testing.T) {
			// CanRead delegates to Project.CanRead; a read share is enough.
			rec, err := readShared.testReadAllWithUser(nil, nil)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"project_id":9`)
		})
		t.Run("Write share can list", func(t *testing.T) {
			rec, err := writeShared.testReadAllWithUser(nil, nil)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"project_id":10`)
		})
		t.Run("Forbidden", func(t *testing.T) {
			_, err := forbidden.testReadAllWithUser(nil, nil)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
	})

	t.Run("ReadOne", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := owned.testReadOneWithUser(nil, map[string]string{"view": "1"})
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"title":"List"`)
			assert.Contains(t, rec.Body.String(), `"id":1`)
			assert.NotEmpty(t, rec.Result().Header.Get("ETag"))
		})
		t.Run("Read-only share can read", func(t *testing.T) {
			// view 33 belongs to project 9 (read share) — CanRead must allow it.
			rec, err := readShared.testReadOneWithUser(nil, map[string]string{"view": "33"})
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"id":33`)
			assert.Contains(t, rec.Body.String(), `"project_id":9`)
		})
		t.Run("Write share can read", func(t *testing.T) {
			// view 37 belongs to project 10 (write share).
			rec, err := writeShared.testReadOneWithUser(nil, map[string]string{"view": "37"})
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"id":37`)
			assert.Contains(t, rec.Body.String(), `"project_id":10`)
		})
		t.Run("Nonexisting", func(t *testing.T) {
			_, err := owned.testReadOneWithUser(nil, map[string]string{"view": "9999"})
			require.Error(t, err)
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
		})
		t.Run("View from another project", func(t *testing.T) {
			// view 5 belongs to project 2; reading it under project 1 must 404.
			_, err := owned.testReadOneWithUser(nil, map[string]string{"view": "5"})
			require.Error(t, err)
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
		})
		t.Run("Forbidden", func(t *testing.T) {
			// view 5 read under its real parent (project 2) — no access.
			_, err := forbidden.testReadOneWithUser(nil, map[string]string{"view": "5"})
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
	})

	t.Run("Create", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := owned.testCreateWithUser(nil, nil, `{"title":"New view","view_kind":"list"}`)
			require.NoError(t, err)
			assert.Equal(t, http.StatusCreated, rec.Code)
			assert.Contains(t, rec.Body.String(), `"title":"New view"`)
			// ownership: the parent project from the URL wins.
			assert.Contains(t, rec.Body.String(), `"project_id":1`)
		})
		t.Run("Admin share can create", func(t *testing.T) {
			// project 11 admin share clears Project.IsAdmin → CanCreate passes.
			rec, err := adminShared.testCreateWithUser(nil, nil, `{"title":"Admin made","view_kind":"list"}`)
			require.NoError(t, err)
			assert.Equal(t, http.StatusCreated, rec.Code)
			assert.Contains(t, rec.Body.String(), `"title":"Admin made"`)
			assert.Contains(t, rec.Body.String(), `"project_id":11`)
		})
		t.Run("Read share cannot create", func(t *testing.T) {
			// project 9 read share → below the IsAdmin bar CanCreate enforces.
			_, err := readShared.testCreateWithUser(nil, nil, `{"title":"Nope","view_kind":"list"}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Write share cannot create", func(t *testing.T) {
			// project 10 write share is still short of admin.
			_, err := writeShared.testCreateWithUser(nil, nil, `{"title":"Nope","view_kind":"list"}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Forbidden", func(t *testing.T) {
			_, err := forbidden.testCreateWithUser(nil, nil, `{"title":"Nope","view_kind":"list"}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Empty title", func(t *testing.T) {
			// Title has valid:"required,runelength(1|250)" → minLength:"1"; Huma's
			// schema validation rejects the empty string with 422 before the model.
			_, err := owned.testCreateWithUser(nil, nil, `{"title":"","view_kind":"list"}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusUnprocessableEntity, getHTTPErrorCode(err))
		})
		t.Run("Title too long", func(t *testing.T) {
			// runelength(1|250) → maxLength:"250"; 251 runes trips Huma's 422.
			_, err := owned.testCreateWithUser(nil, nil, fmt.Sprintf(`{"title":%q,"view_kind":"list"}`, strings.Repeat("a", 251)))
			require.Error(t, err)
			assert.Equal(t, http.StatusUnprocessableEntity, getHTTPErrorCode(err))
		})
		t.Run("Invalid filter", func(t *testing.T) {
			// createProjectView validates Filter.Filter via getTaskFiltersFromFilterString;
			// "foo = value" references an unknown task field → ErrInvalidTaskField (400).
			_, err := owned.testCreateWithUser(nil, nil, `{"title":"Bad filter","view_kind":"list","filter":{"filter":"foo = value"}}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusBadRequest, getHTTPErrorCode(err))
		})
	})

	t.Run("Update", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := owned.testUpdateWithUser(nil, map[string]string{"view": "1"}, `{"title":"Renamed list","view_kind":"list"}`)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"title":"Renamed list"`)
			assert.Contains(t, rec.Body.String(), `"id":1`)
		})
		t.Run("Admin share can update", func(t *testing.T) {
			// view 41 belongs to project 11 (admin share).
			rec, err := adminShared.testUpdateWithUser(nil, map[string]string{"view": "41"}, `{"title":"Admin renamed","view_kind":"list"}`)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"title":"Admin renamed"`)
			assert.Contains(t, rec.Body.String(), `"id":41`)
		})
		t.Run("Read share cannot update", func(t *testing.T) {
			// view 33 belongs to project 9 (read share) → CanUpdate needs admin.
			_, err := readShared.testUpdateWithUser(nil, map[string]string{"view": "33"}, `{"title":"x","view_kind":"list"}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Write share cannot update", func(t *testing.T) {
			// view 37 belongs to project 10 (write share) — still below admin.
			_, err := writeShared.testUpdateWithUser(nil, map[string]string{"view": "37"}, `{"title":"x","view_kind":"list"}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("View from another project", func(t *testing.T) {
			_, err := owned.testUpdateWithUser(nil, map[string]string{"view": "5"}, `{"title":"x","view_kind":"list"}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
		})
		t.Run("Forbidden", func(t *testing.T) {
			_, err := forbidden.testUpdateWithUser(nil, map[string]string{"view": "5"}, `{"title":"x","view_kind":"list"}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Empty title", func(t *testing.T) {
			_, err := owned.testUpdateWithUser(nil, map[string]string{"view": "1"}, `{"title":"","view_kind":"list"}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusUnprocessableEntity, getHTTPErrorCode(err))
		})
		t.Run("Title too long", func(t *testing.T) {
			_, err := owned.testUpdateWithUser(nil, map[string]string{"view": "1"}, fmt.Sprintf(`{"title":%q,"view_kind":"list"}`, strings.Repeat("a", 251)))
			require.Error(t, err)
			assert.Equal(t, http.StatusUnprocessableEntity, getHTTPErrorCode(err))
		})
		t.Run("Invalid filter", func(t *testing.T) {
			// ProjectView.Update validates Filter.Filter the same way Create does.
			_, err := owned.testUpdateWithUser(nil, map[string]string{"view": "1"}, `{"title":"Bad filter","view_kind":"list","filter":{"filter":"foo = value"}}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusBadRequest, getHTTPErrorCode(err))
		})
	})

	t.Run("Delete", func(t *testing.T) {
		t.Run("Read share cannot delete", func(t *testing.T) {
			// view 34 belongs to project 9 (read share) → CanDelete needs admin.
			_, err := readShared.testDeleteWithUser(nil, map[string]string{"view": "34"})
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Write share cannot delete", func(t *testing.T) {
			// view 38 belongs to project 10 (write share) — still below admin.
			_, err := writeShared.testDeleteWithUser(nil, map[string]string{"view": "38"})
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Admin share can delete", func(t *testing.T) {
			// view 42 belongs to project 11 (admin share).
			rec, err := adminShared.testDeleteWithUser(nil, map[string]string{"view": "42"})
			require.NoError(t, err)
			assert.Equal(t, http.StatusNoContent, rec.Code)
			assert.Empty(t, rec.Body.String())
		})
		t.Run("Normal", func(t *testing.T) {
			rec, err := owned.testDeleteWithUser(nil, map[string]string{"view": "2"})
			require.NoError(t, err)
			assert.Equal(t, http.StatusNoContent, rec.Code)
			assert.Empty(t, rec.Body.String())
		})
		t.Run("View from another project", func(t *testing.T) {
			// view 5 belongs to project 2. testuser1 admins the path project (1),
			// so CanDelete passes — but the view isn't under project 1. The Delete
			// guard must 404 before touching project 2's buckets/positions, which
			// the old code wiped via a project_view_id-only delete.
			_, err := owned.testDeleteWithUser(nil, map[string]string{"view": "5"})
			require.Error(t, err)
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
		})
		t.Run("Forbidden", func(t *testing.T) {
			_, err := forbidden.testDeleteWithUser(nil, map[string]string{"view": "5"})
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
	})
}

// TestProjectView_ETagReturns304 covers v2-only conditional-request behaviour.
func TestProjectView_ETagReturns304(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	token := humaTokenFor(t, &testuser1)

	rec := humaRequest(t, e, http.MethodGet, "/api/v2/projects/1/views/1", "", token, "")
	require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
	etag := rec.Header().Get("ETag")
	require.NotEmpty(t, etag, "GET must return an ETag header")

	req := httptest.NewRequest(http.MethodGet, "/api/v2/projects/1/views/1", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("If-None-Match", etag)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNotModified, rec.Code, "body: %s", rec.Body.String())
}

// TestProjectView_PATCHMergePatch confirms AutoPatch synthesises a PATCH that
// only touches supplied fields.
func TestProjectView_PATCHMergePatch(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	token := humaTokenFor(t, &testuser1)

	// Create a fresh view so we don't stomp fixtures.
	rec := humaRequest(t, e, http.MethodPost, "/api/v2/projects/1/views",
		`{"title":"before","view_kind":"table"}`, token, "")
	require.Equal(t, http.StatusCreated, rec.Code, "body: %s", rec.Body.String())
	var created struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created))

	// PATCH only title; view_kind must survive.
	rec = humaRequest(t, e, http.MethodPatch, fmt.Sprintf("/api/v2/projects/1/views/%d", created.ID),
		`{"title":"after"}`, token, "application/merge-patch+json")
	require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())

	rec = humaRequest(t, e, http.MethodGet, fmt.Sprintf("/api/v2/projects/1/views/%d", created.ID), "", token, "")
	require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
	var after struct {
		Title    string `json:"title"`
		ViewKind string `json:"view_kind"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &after))
	assert.Equal(t, "after", after.Title)
	assert.Equal(t, "table", after.ViewKind, "view_kind must survive the PATCH")
}
