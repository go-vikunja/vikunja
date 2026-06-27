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

func mustJSON(s string) string {
	b, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func decodeLabel(t *testing.T, raw []byte) (id int64, description string) {
	t.Helper()
	var l struct {
		ID          int64  `json:"id"`
		Description string `json:"description"`
	}
	require.NoError(t, json.Unmarshal(raw, &l))
	return l.ID, l.Description
}

func TestHumaRichText_FormatDocumented(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)

	rec := humaRequest(t, e, http.MethodGet, "/api/v2/openapi.json", "", "", "")
	require.Equal(t, http.StatusOK, rec.Code)

	type param struct {
		Name string `json:"name"`
		In   string `json:"in"`
	}
	var spec struct {
		Info struct {
			Description string `json:"description"`
		} `json:"info"`
		Paths map[string]map[string]struct {
			Parameters []param `json:"parameters"`
		} `json:"paths"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &spec))

	hasParam := func(path, method, name, in string) bool {
		op, ok := spec.Paths[path][method]
		if !ok {
			return false
		}
		for _, p := range op.Parameters {
			if p.Name == name && p.In == in {
				return true
			}
		}
		return false
	}

	// Query param on the ops where it works (GET/POST/PUT), per entity.
	assert.True(t, hasParam("/labels/{id}", "get", "format", "query"), "labels read must document ?format")
	assert.True(t, hasParam("/labels", "post", "format", "query"), "labels create must document ?format")
	assert.True(t, hasParam("/tasks/{projecttask}", "put", "format", "query"), "tasks update must document ?format")

	// PATCH must NOT advertise ?format — AutoPatch strips the query at runtime, so
	// it would be a trap (markdown stored as HTML). Stripped by stripPatchFormatQuery.
	assert.False(t, hasParam("/labels/{id}", "patch", "format", "query"), "PATCH must not advertise ?format")

	// The X-Vikunja-Format header is documented centrally, not as a per-op param.
	assert.False(t, hasParam("/labels/{id}", "get", "X-Vikunja-Format", "header"))
	assert.False(t, hasParam("/labels/{id}", "patch", "X-Vikunja-Format", "header"))

	// Non-rich-text ops carry no format param.
	assert.False(t, hasParam("/tasks/{task}/comments/{commentid}", "delete", "format", "query"))

	// The cross-cutting behavior, including the PATCH header, is in the API description.
	assert.Contains(t, spec.Info.Description, "Rich-text fields")
	assert.Contains(t, spec.Info.Description, "CalDAV always exchanges")
	assert.Contains(t, spec.Info.Description, "X-Vikunja-Format")
}

func TestHumaRichText_Read(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	token := humaTokenFor(t, &testuser1)

	// Store a label with HTML directly (no format → verbatim).
	rec := humaRequest(t, e, http.MethodPost, "/api/v2/labels",
		`{"title":"rt","description":"<p>Hello <strong>world</strong></p>","hex_color":"112233"}`, token, "")
	require.Equal(t, http.StatusCreated, rec.Code, "body: %s", rec.Body.String())
	id, _ := decodeLabel(t, rec.Body.Bytes())

	t.Run("read as markdown converts html", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, fmt.Sprintf("/api/v2/labels/%d?format=markdown", id), "", token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		_, desc := decodeLabel(t, rec.Body.Bytes())
		assert.Equal(t, "Hello **world**", desc)
	})

	t.Run("read without param keeps html", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, fmt.Sprintf("/api/v2/labels/%d", id), "", token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		_, desc := decodeLabel(t, rec.Body.Bytes())
		assert.Equal(t, "<p>Hello <strong>world</strong></p>", desc)
	})

	t.Run("list converts every item", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/labels?format=markdown", "", token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		// The freshly created label's HTML must not appear; its markdown must.
		assert.NotContains(t, rec.Body.String(), "<strong>world</strong>")
		assert.Contains(t, rec.Body.String(), "Hello **world**")
	})
}

func decodeField(t *testing.T, raw []byte, field string) (id int64, value string) {
	t.Helper()
	var m map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(raw, &m))
	if v, ok := m["id"]; ok {
		_ = json.Unmarshal(v, &id)
	}
	if v, ok := m[field]; ok {
		_ = json.Unmarshal(v, &value)
	}
	return id, value
}

// TestHumaRichText_EveryEntity drives every rich-text entity through the real v2
// API: each is created with a markdown body and read back as both HTML and
// markdown. A handler that stops converting fails its row here.
func TestHumaRichText_EveryEntity(t *testing.T) {
	const md = "a **bold** note"
	const html = "<p>a <strong>bold</strong> note</p>"

	entities := []struct {
		name       string
		createPath string
		createBody string
		readPath   string // fmt verb %d for the created id
		field      string
	}{
		{"label", "/api/v2/labels", `{"title":"e-label","description":"a **bold** note"}`, "/api/v2/labels/%d", "description"},
		{"project", "/api/v2/projects", `{"title":"e-project","description":"a **bold** note"}`, "/api/v2/projects/%d", "description"},
		{"team", "/api/v2/teams", `{"name":"e-team","description":"a **bold** note"}`, "/api/v2/teams/%d", "description"},
		{"saved filter", "/api/v2/filters", `{"title":"e-filter","description":"a **bold** note","filters":{"filter":"done = true"}}`, "/api/v2/filters/%d", "description"},
		{"task", "/api/v2/projects/1/tasks", `{"title":"e-task","description":"a **bold** note"}`, "/api/v2/tasks/%d", "description"},
		{"task comment", "/api/v2/tasks/1/comments", `{"comment":"a **bold** note"}`, "/api/v2/tasks/1/comments/%d", "comment"},
	}

	e, err := setupTestEnv()
	require.NoError(t, err)
	token := humaTokenFor(t, &testuser1)

	for _, ent := range entities {
		t.Run(ent.name, func(t *testing.T) {
			// Markdown body converted to HTML on create.
			rec := humaRequest(t, e, http.MethodPost, ent.createPath+"?format=markdown", ent.createBody, token, "")
			require.Equal(t, http.StatusCreated, rec.Code, "body: %s", rec.Body.String())
			id, _ := decodeField(t, rec.Body.Bytes(), ent.field)
			require.NotZero(t, id)

			// Stored as canonical HTML (default read).
			rec = humaRequest(t, e, http.MethodGet, fmt.Sprintf(ent.readPath, id), "", token, "")
			require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
			_, stored := decodeField(t, rec.Body.Bytes(), ent.field)
			assert.Equal(t, html, stored, "%s write seam did not convert markdown to HTML", ent.name)

			// Read back as markdown.
			rec = humaRequest(t, e, http.MethodGet, fmt.Sprintf(ent.readPath, id)+"?format=markdown", "", token, "")
			require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
			_, asMarkdown := decodeField(t, rec.Body.Bytes(), ent.field)
			assert.Equal(t, md, asMarkdown, "%s read transformer did not convert HTML to markdown", ent.name)
		})
	}
}

// TestHumaRichText_KanbanNested proves the read conversion reaches tasks nested
// inside kanban buckets (Body.Items[].Tasks[].Description), which the explicit
// handler converts by looping the buckets.
func TestHumaRichText_KanbanNested(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	token := humaTokenFor(t, &testuser1)

	// Store a task with HTML directly (no format → verbatim) in project 1.
	rec := humaRequest(t, e, http.MethodPost, "/api/v2/projects/1/tasks",
		`{"title":"kanban task","description":"<p>kanban <strong>md</strong></p>"}`, token, "")
	require.Equal(t, http.StatusCreated, rec.Code, "body: %s", rec.Body.String())

	// View 4 is project 1's kanban view; its buckets/tasks response nests tasks.
	rec = humaRequest(t, e, http.MethodGet, "/api/v2/projects/1/views/4/buckets/tasks?format=markdown", "", token, "")
	require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
	assert.Contains(t, rec.Body.String(), "kanban **md**", "nested task description must be converted to markdown")
	assert.NotContains(t, rec.Body.String(), "<strong>md</strong>", "no HTML should leak from a nested task")
}

// TestHumaRichText_TaskExpandedNested proves expanded comments and related tasks
// are converted too, not just the top-level task description.
func TestHumaRichText_TaskExpandedNested(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	token := humaTokenFor(t, &testuser1)

	// A comment with HTML on task 1.
	rec := humaRequest(t, e, http.MethodPost, "/api/v2/tasks/1/comments",
		`{"comment":"<p>a <strong>bold</strong> comment</p>"}`, token, "")
	require.Equal(t, http.StatusCreated, rec.Code, "body: %s", rec.Body.String())

	// A subtask (related task) with an HTML description.
	rec = humaRequest(t, e, http.MethodPost, "/api/v2/projects/1/tasks",
		`{"title":"sub","description":"<p>sub <strong>desc</strong></p>"}`, token, "")
	require.Equal(t, http.StatusCreated, rec.Code, "body: %s", rec.Body.String())
	subID, _ := decodeField(t, rec.Body.Bytes(), "title")
	rec = humaRequest(t, e, http.MethodPost, "/api/v2/tasks/1/relations",
		fmt.Sprintf(`{"other_task_id":%d,"relation_kind":"subtask"}`, subID), token, "")
	require.Equal(t, http.StatusCreated, rec.Code, "body: %s", rec.Body.String())

	rec = humaRequest(t, e, http.MethodGet, "/api/v2/tasks/1?expand=comments&expand=subtasks&format=markdown", "", token, "")
	require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
	body := rec.Body.String()
	assert.Contains(t, body, "a **bold** comment", "expanded comment must be markdown")
	assert.Contains(t, body, "sub **desc**", "related task description must be markdown")
	assert.NotContains(t, body, "<strong>", "no nested HTML should leak")
}

func TestHumaRichText_Write(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	token := humaTokenFor(t, &testuser1)

	t.Run("markdown write is stored as html", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodPost, "/api/v2/labels?format=markdown",
			`{"title":"w1","description":"Hello **world**","hex_color":"112233"}`, token, "")
		require.Equal(t, http.StatusCreated, rec.Code, "body: %s", rec.Body.String())
		id, _ := decodeLabel(t, rec.Body.Bytes())

		// Read back without format → canonical HTML.
		rec = humaRequest(t, e, http.MethodGet, fmt.Sprintf("/api/v2/labels/%d", id), "", token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		_, desc := decodeLabel(t, rec.Body.Bytes())
		assert.Equal(t, "<p>Hello <strong>world</strong></p>", desc)
	})

	t.Run("default write stores body verbatim", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodPost, "/api/v2/labels",
			`{"title":"w2","description":"Hello **world**","hex_color":"112233"}`, token, "")
		require.Equal(t, http.StatusCreated, rec.Code, "body: %s", rec.Body.String())
		id, _ := decodeLabel(t, rec.Body.Bytes())

		rec = humaRequest(t, e, http.MethodGet, fmt.Sprintf("/api/v2/labels/%d", id), "", token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		_, desc := decodeLabel(t, rec.Body.Bytes())
		assert.Equal(t, "Hello **world**", desc, "without the param the body is stored unconverted")
	})

	t.Run("mention is rebuilt on markdown write", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodPost, "/api/v2/labels?format=markdown",
			`{"title":"w3","description":"ping @user1","hex_color":"112233"}`, token, "")
		require.Equal(t, http.StatusCreated, rec.Code, "body: %s", rec.Body.String())
		id, _ := decodeLabel(t, rec.Body.Bytes())

		rec = humaRequest(t, e, http.MethodGet, fmt.Sprintf("/api/v2/labels/%d", id), "", token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		_, desc := decodeLabel(t, rec.Body.Bytes())
		assert.Contains(t, desc, `<mention-user data-id="user1"`)
	})

	t.Run("markdown round trip is stable", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodPost, "/api/v2/labels?format=markdown",
			`{"title":"w4","description":"- [x] done\n- [ ] todo","hex_color":"112233"}`, token, "")
		require.Equal(t, http.StatusCreated, rec.Code, "body: %s", rec.Body.String())
		id, _ := decodeLabel(t, rec.Body.Bytes())

		// GET as markdown → PUT it back as markdown → GET as markdown must match.
		rec = humaRequest(t, e, http.MethodGet, fmt.Sprintf("/api/v2/labels/%d?format=markdown", id), "", token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		_, md1 := decodeLabel(t, rec.Body.Bytes())

		put := fmt.Sprintf(`{"title":"w4","description":%s,"hex_color":"112233"}`, mustJSON(md1))
		rec = humaRequest(t, e, http.MethodPut, fmt.Sprintf("/api/v2/labels/%d?format=markdown", id), put, token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())

		rec = humaRequest(t, e, http.MethodGet, fmt.Sprintf("/api/v2/labels/%d?format=markdown", id), "", token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		_, md2 := decodeLabel(t, rec.Body.Bytes())
		assert.Equal(t, md1, md2, "markdown projection must be stable across a round trip")
	})

	t.Run("patch honours markdown via header", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodPost, "/api/v2/labels",
			`{"title":"w5","description":"<p>old</p>","hex_color":"112233"}`, token, "")
		require.Equal(t, http.StatusCreated, rec.Code, "body: %s", rec.Body.String())
		id, _ := decodeLabel(t, rec.Body.Bytes())

		// AutoPatch strips the query string but forwards headers, so PATCH markdown
		// support rides on X-Vikunja-Format.
		req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v2/labels/%d", id),
			strings.NewReader(`{"description":"new **bold**"}`))
		req.Header.Set("Content-Type", "application/merge-patch+json")
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("X-Vikunja-Format", "markdown")
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())

		rec = humaRequest(t, e, http.MethodGet, fmt.Sprintf("/api/v2/labels/%d", id), "", token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		_, desc := decodeLabel(t, rec.Body.Bytes())
		assert.Equal(t, "<p>new <strong>bold</strong></p>", desc)
	})
}
