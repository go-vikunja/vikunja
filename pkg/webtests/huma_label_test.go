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
	"code.vikunja.io/api/pkg/user"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// labelTokenFor issues a JWT for a test user via the real auth flow — used
// only by the v2-only supplementary tests below.
func labelTokenFor(t *testing.T, u *user.User) string {
	t.Helper()
	tok, err := auth.NewUserJWTAuthtoken(u, "test-session-id")
	require.NoError(t, err)
	return tok
}

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

// TestHumaLabel mirrors the v1 webtest shape (see project_test.go's TestProject)
// so the v2 contract can be read side-by-side with the v1 coverage. The goal
// is to prove v2 is behaviourally compatible with v1 modulo the documented
// verb/error-shape changes.
//
// Per the PR review: "The tests should mirror *exactly* the existing tests.
// We want to make sure that the api is fully compatible to v1 (minus the
// changed verbs and error responses)." Labels has no v1 webtest today, so the
// coverage below is patterned after pkg/models/label_test.go (the
// model-level coverage).
func TestHumaLabel(t *testing.T) {
	testHandler := webHandlerTestV2{
		user:     &testuser1,
		basePath: "/api/v2/labels",
		idParam:  "label",
		t:        t,
	}

	t.Run("ReadAll", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := testHandler.testReadAllWithUser(nil, nil)
			require.NoError(t, err)
			// User 1 owns labels #1 and #2 per fixtures — both must show up.
			assert.Contains(t, rec.Body.String(), `Label #1`)
			assert.Contains(t, rec.Body.String(), `Label #2`)
			// Label #3 is owned by user2 with no shared task, user1 must not
			// see it (direct mirror of the model-level negative case).
			assert.NotContains(t, rec.Body.String(), `Label #3 - other user`)
			// Label #6 is the GHSA regression fixture — private to user13.
			assert.NotContains(t, rec.Body.String(), `Label #6 - private`)
		})
	})

	t.Run("ReadOne", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := testHandler.testReadOneWithUser(nil, map[string]string{"label": "1"})
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"title":"Label #1"`)
			// v2 read carries an ETag header derived from id+updated — part
			// of the "changed response shape" allowed by the review gate.
			assert.NotEmpty(t, rec.Result().Header.Get("ETag"))
		})
		t.Run("Nonexisting", func(t *testing.T) {
			// v1 convention: missing labels return 403 not 404 (refusal to
			// disclose existence via the CanRead branch). v2 preserves this.
			_, err := testHandler.testReadOneWithUser(nil, map[string]string{"label": "9999"})
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Permissions check", func(t *testing.T) {
			t.Run("Forbidden", func(t *testing.T) {
				// Label 6 is owned by user13 and attached only to a private
				// task — user1 must be rejected.
				_, err := testHandler.testReadOneWithUser(nil, map[string]string{"label": "6"})
				require.Error(t, err)
				assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
			})
		})
	})

	t.Run("Create", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := testHandler.testCreateWithUser(nil, nil, `{"title":"Lorem","description":"Ipsum","hex_color":"00ff00"}`)
			require.NoError(t, err)
			assert.Equal(t, http.StatusCreated, rec.Code)
			assert.Contains(t, rec.Body.String(), `"title":"Lorem"`)
			assert.Contains(t, rec.Body.String(), `"description":"Ipsum"`)
			assert.Contains(t, rec.Body.String(), `"hex_color":"00ff00"`)
		})
		t.Run("Empty title", func(t *testing.T) {
			// v2 validation error shape differs from v1 (RFC 9457 with 422,
			// vs v1's domain ValidationHTTPError at 400). Content checked in
			// TestHumaLabel_ErrorShapeIsRFC9457 once — here we only assert
			// the request was rejected.
			_, err := testHandler.testCreateWithUser(nil, nil, `{"title":""}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusUnprocessableEntity, getHTTPErrorCode(err))
		})
	})

	t.Run("Update", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"label": "1"}, `{"title":"TestLoremIpsum"}`)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
		})
		t.Run("Nonexisting", func(t *testing.T) {
			// Update: CanUpdate → isLabelOwner → getLabelByIDSimple returns
			// ErrLabelDoesNotExist (404) for a missing label. v1's label
			// model behaves identically; only the model_test set
			// `wantForbidden: true` because it tested the `allowed bool`
			// separately from the error. The real pipeline surfaces 404.
			_, err := testHandler.testUpdateWithUser(nil, map[string]string{"label": "9999"}, `{"title":"TestLoremIpsum"}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
		})
		t.Run("Permissions check", func(t *testing.T) {
			t.Run("Forbidden", func(t *testing.T) {
				// Label 6 is owned by user13 — user1 cannot update it.
				_, err := testHandler.testUpdateWithUser(nil, map[string]string{"label": "6"}, `{"title":"TestLoremIpsum"}`)
				require.Error(t, err)
				assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
			})
		})
	})

	t.Run("Delete", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := testHandler.testDeleteWithUser(nil, map[string]string{"label": "2"})
			require.NoError(t, err)
			// v2 delete returns 204 No Content (changed from v1's 200 +
			// message body — explicitly part of the response-shape
			// differences allowed by the review gate).
			assert.Equal(t, http.StatusNoContent, rec.Code)
			assert.Empty(t, rec.Body.String())
		})
		t.Run("Nonexisting", func(t *testing.T) {
			// Delete: CanDelete → isLabelOwner → getLabelByIDSimple returns
			// ErrLabelDoesNotExist (404) — same rationale as Update above.
			_, err := testHandler.testDeleteWithUser(nil, map[string]string{"label": "9999"})
			require.Error(t, err)
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
		})
		t.Run("Permissions check", func(t *testing.T) {
			t.Run("Forbidden", func(t *testing.T) {
				// Label 6 is owned by user13 — user1 cannot delete it.
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{"label": "6"})
				require.Error(t, err)
				assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
			})
		})
	})
}

// The tests below cover v2-only behaviour with no v1 counterpart: ETag +
// conditional requests, AutoPatch (merge-patch+json), the OpenAPI spec, and
// the RFC 9457 error-shape guarantee. They live under separate top-level
// TestFunc names so they're clearly supplementary to the v1-parity coverage
// in TestHumaLabel above.

// humaRequest is a one-shot dispatch helper that reuses an already-bootstrapped
// echo.Echo. Used by the v2-only supplementary tests below to avoid
// re-loading fixtures between chained calls (create → patch → get).
func humaRequest(t *testing.T, e *echo.Echo, method, path, body, token, contentType string) *httptest.ResponseRecorder {
	t.Helper()
	var reader *strings.Reader
	if body != "" {
		reader = strings.NewReader(body)
	} else {
		reader = strings.NewReader("")
	}
	req := httptest.NewRequest(method, path, reader)
	if contentType == "" {
		contentType = "application/json"
	}
	req.Header.Set("Content-Type", contentType)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func TestHumaLabel_ETagReturns304(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	token := labelTokenFor(t, &testuser1)

	// First GET to capture the ETag. Label 1 belongs to user1.
	rec := humaRequest(t, e, http.MethodGet, "/api/v2/labels/1", "", token, "")
	require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
	etag := rec.Header().Get("ETag")
	require.NotEmpty(t, etag, "GET must return an ETag header")

	// Second GET with If-None-Match must return 304.
	req := httptest.NewRequest(http.MethodGet, "/api/v2/labels/1", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("If-None-Match", etag)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNotModified, rec.Code, "body: %s", rec.Body.String())
}

func TestHumaLabel_PATCHMergePatch(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	token := labelTokenFor(t, &testuser1)

	// Create a label we can mutate without stomping fixtures.
	rec := humaRequest(t, e, http.MethodPost, "/api/v2/labels",
		`{"title":"before","description":"keep me","hex_color":"112233"}`, token, "")
	require.Equal(t, http.StatusCreated, rec.Code, "body: %s", rec.Body.String())
	var created struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created))

	// PATCH only the title. AutoPatch should leave description + hex_color
	// alone. Reuses the same echo.Echo so the create above isn't wiped by
	// a fixture reload.
	rec = humaRequest(t, e, http.MethodPatch, fmt.Sprintf("/api/v2/labels/%d", created.ID),
		`{"title":"after"}`, token, "application/merge-patch+json")
	require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())

	// Verify via a direct GET that only title changed.
	rec = humaRequest(t, e, http.MethodGet, fmt.Sprintf("/api/v2/labels/%d", created.ID), "", token, "")
	require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
	var after struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		HexColor    string `json:"hex_color"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &after))
	assert.Equal(t, "after", after.Title, "title must reflect the PATCH")
	assert.Equal(t, "keep me", after.Description, "description must survive the PATCH untouched")
	assert.Equal(t, "112233", after.HexColor, "hex_color must survive the PATCH untouched")
}

func TestHumaLabel_OpenAPISpecDescribesAllFive(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	// The spec is public — no token needed.
	rec := humaRequest(t, e, http.MethodGet, "/api/v2/openapi.json", "", "", "")
	require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())

	var spec struct {
		Paths map[string]map[string]any `json:"paths"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &spec), "body: %s", rec.Body.String())

	// /labels: GET (list), POST (create). /labels/{id}: GET (read),
	// PUT (update), DELETE (delete). Huma registers these relative to the
	// group so the spec paths are /labels and /labels/{id}.
	list, ok := spec.Paths["/labels"]
	require.True(t, ok, "spec must contain /labels path; paths=%v", pathKeys(spec.Paths))
	assert.Contains(t, list, "get", "/labels should have GET")
	assert.Contains(t, list, "post", "/labels should have POST")

	item, ok := spec.Paths["/labels/{id}"]
	require.True(t, ok, "spec must contain /labels/{id} path; paths=%v", pathKeys(spec.Paths))
	assert.Contains(t, item, "get", "/labels/{id} should have GET")
	assert.Contains(t, item, "put", "/labels/{id} should have PUT")
	assert.Contains(t, item, "delete", "/labels/{id} should have DELETE")

	total := len(list) + len(item)
	// The five hand-written operations plus any AutoPatch-added PATCH on
	// /labels/{id}. Assert at least five.
	assert.GreaterOrEqual(t, total, 5, "expected at least 5 Label operations in the spec; got %d (list=%v item=%v)", total, list, item)
}

// TestHumaLabel_ErrorShapeIsRFC9457 asserts once — across a 403 and a 422
// — that v2 errors use application/problem+json with a `status` field.
// This is the "changed error responses" deviation from v1, so the assertion
// lives in its own test rather than being duplicated at every call site.
func TestHumaLabel_ErrorShapeIsRFC9457(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	token := labelTokenFor(t, &testuser1)

	t.Run("403 Forbidden", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/labels/6", "", token, "")
		require.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())

		ct := rec.Header().Get("Content-Type")
		assert.Contains(t, ct, "application/problem+json", "forbidden response must use RFC 9457 content type; got %q", ct)

		var body humaErrorBody
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body), "body: %s", rec.Body.String())
		assert.Equal(t, http.StatusForbidden, body.Status)
		assert.NotEmpty(t, body.Title, "title is required by RFC 9457")
	})

	t.Run("422 Validation", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodPost, "/api/v2/labels", `{"title":""}`, token, "")
		require.Equal(t, http.StatusUnprocessableEntity, rec.Code, "body: %s", rec.Body.String())

		ct := rec.Header().Get("Content-Type")
		assert.Contains(t, ct, "application/problem+json", "validation response must use RFC 9457 content type; got %q", ct)

		var body humaErrorBody
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body), "body: %s", rec.Body.String())
		assert.Equal(t, http.StatusUnprocessableEntity, body.Status)
		require.NotEmpty(t, body.Errors, "validation errors must include structured per-field details")
		var foundTitleError bool
		for _, detail := range body.Errors {
			if strings.Contains(detail.Location, "title") {
				foundTitleError = true
				break
			}
		}
		assert.True(t, foundTitleError, "expected at least one error detail locating `title`; got %+v", body.Errors)
	})
}

func pathKeys(m map[string]map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
