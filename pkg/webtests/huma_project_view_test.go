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
// Fixtures: project 1 (owned by testuser1) has views 1-4; project 2 (owned by
// user3, no share to testuser1) has views 5-8 — used for the forbidden and
// wrong-parent negatives.
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
		t.Run("Forbidden", func(t *testing.T) {
			_, err := forbidden.testCreateWithUser(nil, nil, `{"title":"Nope","view_kind":"list"}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
	})

	t.Run("Update", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := owned.testUpdateWithUser(nil, map[string]string{"view": "1"}, `{"title":"Renamed list","view_kind":"list"}`)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"title":"Renamed list"`)
			assert.Contains(t, rec.Body.String(), `"id":1`)
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
	})

	t.Run("Delete", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := owned.testDeleteWithUser(nil, map[string]string{"view": "2"})
			require.NoError(t, err)
			assert.Equal(t, http.StatusNoContent, rec.Code)
			assert.Empty(t, rec.Body.String())
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
