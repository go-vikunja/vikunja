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

	"code.vikunja.io/api/pkg/modules/auth"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// humaErrorBody is the RFC 9457 problem+json shape Huma emits by default.
// Fields are a subset — we only assert what's load-bearing for the tests.
type humaErrorBody struct {
	Type   string `json:"type"`
	Title  string `json:"title"`
	Status int    `json:"status"`
	Detail string `json:"detail"`
	Errors []struct {
		Message  string `json:"message"`
		Location string `json:"location"`
	} `json:"errors"`
}

// labelResponse matches the Label struct fields the tests care about.
// Defined locally to avoid pulling models into a tight test fixture.
type labelResponse struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	HexColor    string `json:"hex_color"`
}

// labelListResponse mirrors the Paginated[*Label] envelope. We inline the
// items type to avoid importing apiv2 (which would create a test-package
// import loop nuisance).
type labelListResponse struct {
	Items      []labelResponse `json:"items"`
	Total      int64           `json:"total"`
	Page       int             `json:"page"`
	PerPage    int             `json:"per_page"`
	TotalPages int64           `json:"total_pages"`
}

// newAuthedV2Request builds a /api/v2 request with a JWT for the given
// test user.
func newAuthedV2Request(t *testing.T, method, path, body, token string) *http.Request {
	t.Helper()
	var reader *strings.Reader
	if body != "" {
		reader = strings.NewReader(body)
	} else {
		reader = strings.NewReader("")
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return req
}

// tokenForUser issues a JWT for the given test user.
func tokenForUser(t *testing.T, userID int) string {
	t.Helper()
	var u = &testuser1
	switch userID {
	case 1:
		u = &testuser1
	case 6:
		u = &testuser6
	case 10:
		u = &testuser10
	case 15:
		u = &testuser15
	default:
		t.Fatalf("tokenForUser: unsupported userID %d", userID)
	}
	tok, err := auth.NewUserJWTAuthtoken(u, "test-session-id")
	require.NoError(t, err)
	return tok
}

// serve dispatches the request through the full Echo handler chain and
// returns the recorder. Everything goes through routes.RegisterRoutes so
// the JWT middleware, normaliser, and Huma adapter all run.
func serve(t *testing.T, e *echo.Echo, req *http.Request) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func TestHumaLabel_Create_Read_Update_Delete(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	token := tokenForUser(t, 1)

	// Create
	payload := `{"title":"huma test label","description":"round-trip","hex_color":"00ff00"}`
	rec := serve(t, e, newAuthedV2Request(t, http.MethodPost, "/api/v2/labels", payload, token))
	require.Equal(t, http.StatusCreated, rec.Code, "body: %s", rec.Body.String())

	var created labelResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created))
	assert.NotZero(t, created.ID)
	assert.Equal(t, "huma test label", created.Title)
	assert.Equal(t, "round-trip", created.Description)
	assert.Equal(t, "00ff00", created.HexColor)

	// Read
	rec = serve(t, e, newAuthedV2Request(t, http.MethodGet, fmt.Sprintf("/api/v2/labels/%d", created.ID), "", token))
	require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
	var read labelResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &read))
	assert.Equal(t, created.ID, read.ID)
	assert.Equal(t, created.Title, read.Title)
	assert.NotEmpty(t, rec.Header().Get("ETag"), "read response must carry an ETag header")

	// Update (PUT — full replacement)
	updatePayload := `{"title":"renamed","description":"updated","hex_color":"ff0000"}`
	rec = serve(t, e, newAuthedV2Request(t, http.MethodPut, fmt.Sprintf("/api/v2/labels/%d", created.ID), updatePayload, token))
	require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
	var updated labelResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &updated))
	assert.Equal(t, "renamed", updated.Title)
	assert.Equal(t, "updated", updated.Description)
	assert.Equal(t, "ff0000", updated.HexColor)

	// Delete
	rec = serve(t, e, newAuthedV2Request(t, http.MethodDelete, fmt.Sprintf("/api/v2/labels/%d", created.ID), "", token))
	require.Equal(t, http.StatusNoContent, rec.Code, "body: %s", rec.Body.String())

	// Confirm it's gone — a follow-up read must return a non-success. The
	// model's CanRead check fails for missing labels with a 403 (the
	// Vikunja convention of refusing to disclose whether an ID exists); a
	// 404 is also acceptable. Either way, not 200.
	rec = serve(t, e, newAuthedV2Request(t, http.MethodGet, fmt.Sprintf("/api/v2/labels/%d", created.ID), "", token))
	assert.Contains(t, []int{http.StatusForbidden, http.StatusNotFound}, rec.Code,
		"post-delete GET must not return 200; body: %s", rec.Body.String())
}

// TestHumaLabel_List_ReturnsItems is the key regression catcher: the spike
// implementation quietly returned an empty array because the generic any
// slice couldn't be cast back to the concrete type. This test asserts that
// list returns an actual populated items array when fixtures contain
// labels the user can see.
func TestHumaLabel_List_ReturnsItems(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	token := tokenForUser(t, 1)

	rec := serve(t, e, newAuthedV2Request(t, http.MethodGet, "/api/v2/labels", "", token))
	require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())

	var list labelListResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &list), "body: %s", rec.Body.String())
	require.NotEmpty(t, list.Items, "expected at least one label in list response for user1 given the test fixtures")

	// User 1 owns labels #1 and #2 per fixtures — verify one of them shows
	// up, so the response is genuinely populated (not a half-broken cast).
	var titles []string
	for _, it := range list.Items {
		titles = append(titles, it.Title)
	}
	assert.Contains(t, titles, "Label #1", "user1 should see their own Label #1; got titles=%v", titles)
}

func TestHumaLabel_ForbiddenErrorShape(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	// Label #6 belongs to user13 and is only visible to them per fixtures.
	// user1 must get a 403 on both read and delete.
	token := tokenForUser(t, 1)

	rec := serve(t, e, newAuthedV2Request(t, http.MethodGet, "/api/v2/labels/6", "", token))
	require.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())

	ct := rec.Header().Get("Content-Type")
	assert.Contains(t, ct, "application/problem+json", "forbidden response must use RFC 9457 content type; got %q", ct)

	var body humaErrorBody
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body), "body: %s", rec.Body.String())
	assert.Equal(t, http.StatusForbidden, body.Status)
	assert.NotEmpty(t, body.Title, "title is required by RFC 9457")
}

func TestHumaLabel_ValidationErrorShape(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	token := tokenForUser(t, 1)

	// Title is constrained minLength:1 on the Label struct. An empty
	// title must fail Huma's schema validation with RFC 9457 + detailed
	// per-field errors.
	rec := serve(t, e, newAuthedV2Request(t, http.MethodPost, "/api/v2/labels", `{"title":""}`, token))
	require.Equal(t, http.StatusUnprocessableEntity, rec.Code, "body: %s", rec.Body.String())

	ct := rec.Header().Get("Content-Type")
	assert.Contains(t, ct, "application/problem+json", "validation response must use RFC 9457 content type; got %q", ct)

	var body humaErrorBody
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body), "body: %s", rec.Body.String())
	assert.Equal(t, http.StatusUnprocessableEntity, body.Status)
	require.NotEmpty(t, body.Errors, "validation errors must include structured per-field details")
	// At least one structured error must point at `body.title`.
	var foundTitleError bool
	for _, detail := range body.Errors {
		if strings.Contains(detail.Location, "title") {
			foundTitleError = true
			break
		}
	}
	assert.True(t, foundTitleError, "expected at least one error detail locating `title`; got %+v", body.Errors)
}

func TestHumaLabel_ETagReturns304(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	token := tokenForUser(t, 1)

	// First GET to capture the ETag. Label 1 belongs to user1.
	rec := serve(t, e, newAuthedV2Request(t, http.MethodGet, "/api/v2/labels/1", "", token))
	require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
	etag := rec.Header().Get("ETag")
	require.NotEmpty(t, etag, "GET must return an ETag header")

	// Second GET with If-None-Match must return 304.
	req := newAuthedV2Request(t, http.MethodGet, "/api/v2/labels/1", "", token)
	req.Header.Set("If-None-Match", etag)
	rec = serve(t, e, req)
	require.Equal(t, http.StatusNotModified, rec.Code, "body: %s", rec.Body.String())
}

func TestHumaLabel_PATCHMergePatch(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	token := tokenForUser(t, 1)

	// Create a label we can mutate without stomping fixtures.
	rec := serve(t, e, newAuthedV2Request(t, http.MethodPost, "/api/v2/labels",
		`{"title":"before","description":"keep me","hex_color":"112233"}`, token))
	require.Equal(t, http.StatusCreated, rec.Code, "body: %s", rec.Body.String())
	var created labelResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created))

	// PATCH only the title. AutoPatch should leave description + hex_color
	// alone.
	req := newAuthedV2Request(t, http.MethodPatch, fmt.Sprintf("/api/v2/labels/%d", created.ID),
		`{"title":"after"}`, token)
	req.Header.Set("Content-Type", "application/merge-patch+json")
	rec = serve(t, e, req)
	require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())

	// Verify via a direct GET that only title changed.
	rec = serve(t, e, newAuthedV2Request(t, http.MethodGet, fmt.Sprintf("/api/v2/labels/%d", created.ID), "", token))
	require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
	var after labelResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &after))
	assert.Equal(t, "after", after.Title, "title must reflect the PATCH")
	assert.Equal(t, "keep me", after.Description, "description must survive the PATCH untouched")
	assert.Equal(t, "112233", after.HexColor, "hex_color must survive the PATCH untouched")
}

func TestHumaLabel_OpenAPISpecDescribesAllFive(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)

	// The spec is public — no token needed.
	req := httptest.NewRequest(http.MethodGet, "/api/v2/openapi.json", nil)
	rec := serve(t, e, req)
	require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())

	var spec struct {
		Paths map[string]map[string]any `json:"paths"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &spec), "body: %s", rec.Body.String())

	// /labels: GET (list), POST (create). /labels/{id}: GET (read),
	// PUT (update), DELETE (delete). Huma registers these relative to the
	// group so the spec paths are /labels and /labels/{id}.
	list, ok := spec.Paths["/labels"]
	require.True(t, ok, "spec must contain /labels path; paths=%v", keys(spec.Paths))
	assert.Contains(t, list, "get", "/labels should have GET")
	assert.Contains(t, list, "post", "/labels should have POST")

	item, ok := spec.Paths["/labels/{id}"]
	require.True(t, ok, "spec must contain /labels/{id} path; paths=%v", keys(spec.Paths))
	assert.Contains(t, item, "get", "/labels/{id} should have GET")
	assert.Contains(t, item, "put", "/labels/{id} should have PUT")
	assert.Contains(t, item, "delete", "/labels/{id} should have DELETE")

	total := len(list) + len(item)
	// The five hand-written operations plus any AutoPatch-added PATCH on
	// /labels/{id}. Assert at least five.
	assert.GreaterOrEqual(t, total, 5, "expected at least 5 Label operations in the spec; got %d (list=%v item=%v)", total, list, item)
}

func keys(m map[string]map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
