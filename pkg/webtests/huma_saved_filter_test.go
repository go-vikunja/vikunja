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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHumaSavedFilter ports the owner-only matrix from saved_filters_test.go
// onto the HTTP surface; v1 has no /filters webtest, so this is the only one.
// Fixture: filter #1 is owned by user1 (saved_filters.yml).
func TestHumaSavedFilter(t *testing.T) {
	testHandler := webHandlerTestV2{
		user:     &testuser1,
		basePath: "/api/v2/filters",
		idParam:  "filter",
		t:        t,
	}
	require.NoError(t, testHandler.ensureEnv())
	// Share the one Echo (and its single fixture load) as user2; v2 doesn't
	// reload fixtures per request, so Update/Delete of #1 are ordered last.
	otherUserHandler := webHandlerTestV2{
		user:     &testuser2,
		basePath: "/api/v2/filters",
		idParam:  "filter",
		t:        t,
		e:        testHandler.e,
	}

	t.Run("ReadOne", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := testHandler.testReadOneWithUser(nil, map[string]string{"filter": "1"})
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"title":"testfilter1"`)
			// Owner-only resource → admin permission for a successful read.
			assert.Contains(t, rec.Body.String(), `"max_permission":2`)
			assert.NotEmpty(t, rec.Result().Header.Get("ETag"))
		})
		t.Run("Nonexisting", func(t *testing.T) {
			// canDoFilter loads the filter, so a missing id surfaces 404.
			_, err := testHandler.testReadOneWithUser(nil, map[string]string{"filter": "9999"})
			require.Error(t, err)
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
		})
		t.Run("Forbidden - not owner", func(t *testing.T) {
			// #1 is user1's; saved filters are owner-only, so user2 is refused.
			_, err := otherUserHandler.testReadOneWithUser(nil, map[string]string{"filter": "1"})
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
	})

	t.Run("Create", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := testHandler.testCreateWithUser(nil, nil,
				`{"title":"Lorem","description":"Ipsum","filters":{"filter":"done = true"}}`)
			require.NoError(t, err)
			assert.Equal(t, http.StatusCreated, rec.Code)
			assert.Contains(t, rec.Body.String(), `"title":"Lorem"`)
			assert.Contains(t, rec.Body.String(), `"description":"Ipsum"`)
		})
		t.Run("Empty title", func(t *testing.T) {
			// 422 (not Huma's schema 400): central govalidator on `valid:"required"`.
			_, err := testHandler.testCreateWithUser(nil, nil,
				`{"title":"","filters":{"filter":"done = true"}}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusUnprocessableEntity, getHTTPErrorCode(err))
		})
		t.Run("Invalid filter string", func(t *testing.T) {
			// 400 from the model's filter parser, not the 422 validation path.
			_, err := testHandler.testCreateWithUser(nil, nil,
				`{"title":"BadFilter","filters":{"filter":"foo = bar"}}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusBadRequest, getHTTPErrorCode(err))
		})
	})

	t.Run("Update", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"filter": "1"},
				`{"title":"NewTitle","filters":{"filter":"done = true"}}`)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"title":"NewTitle"`)
		})
		t.Run("Nonexisting", func(t *testing.T) {
			_, err := testHandler.testUpdateWithUser(nil, map[string]string{"filter": "9999"},
				`{"title":"NewTitle","filters":{"filter":"done = true"}}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
		})
		t.Run("Empty title", func(t *testing.T) {
			_, err := testHandler.testUpdateWithUser(nil, map[string]string{"filter": "1"},
				`{"title":"","filters":{"filter":"done = true"}}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusUnprocessableEntity, getHTTPErrorCode(err))
		})
		t.Run("Forbidden - not owner", func(t *testing.T) {
			_, err := otherUserHandler.testUpdateWithUser(nil, map[string]string{"filter": "1"},
				`{"title":"NewTitle","filters":{"filter":"done = true"}}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
	})

	// Normal is last: it removes #1, which the negatives above still need.
	t.Run("Delete", func(t *testing.T) {
		t.Run("Nonexisting", func(t *testing.T) {
			_, err := testHandler.testDeleteWithUser(nil, map[string]string{"filter": "9999"})
			require.Error(t, err)
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
		})
		t.Run("Forbidden - not owner", func(t *testing.T) {
			_, err := otherUserHandler.testDeleteWithUser(nil, map[string]string{"filter": "1"})
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Normal", func(t *testing.T) {
			rec, err := testHandler.testDeleteWithUser(nil, map[string]string{"filter": "1"})
			require.NoError(t, err)
			assert.Equal(t, http.StatusNoContent, rec.Code)
			assert.Empty(t, rec.Body.String())
		})
	})
}

// v2-only behaviour, no v1 counterpart: ETag/304 and AutoPatch.

func TestHumaSavedFilter_ETagReturns304(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	token := humaTokenFor(t, &testuser1)

	rec := humaRequest(t, e, http.MethodGet, "/api/v2/filters/1", "", token, "")
	require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
	etag := rec.Header().Get("ETag")
	require.NotEmpty(t, etag, "GET must return an ETag header")

	req := httptest.NewRequest(http.MethodGet, "/api/v2/filters/1", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("If-None-Match", etag)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNotModified, rec.Code, "body: %s", rec.Body.String())
}

func TestHumaSavedFilter_PATCHMergePatch(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	token := humaTokenFor(t, &testuser1)

	// PATCH only the title; AutoPatch must leave the description alone.
	rec := humaRequest(t, e, http.MethodPatch, "/api/v2/filters/1",
		`{"title":"patched"}`, token, "application/merge-patch+json")
	require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())

	rec = humaRequest(t, e, http.MethodGet, "/api/v2/filters/1", "", token, "")
	require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
	var after struct {
		Title string `json:"title"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &after))
	assert.Equal(t, "patched", after.Title)
}
