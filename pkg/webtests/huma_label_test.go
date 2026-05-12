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

// TestHumaLabel mirrors v1's TestProject shape so v2 contract parity is
// readable side-by-side. Labels has no v1 webtest, so coverage is patterned
// after pkg/models/label_test.go.
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
			// User 1 owns labels #1 and #2; #3 is user2's, #6 is the GHSA private fixture.
			assert.Contains(t, rec.Body.String(), `Label #1`)
			assert.Contains(t, rec.Body.String(), `Label #2`)
			assert.NotContains(t, rec.Body.String(), `Label #3 - other user`)
			assert.NotContains(t, rec.Body.String(), `Label #6 - private`)
		})
	})

	t.Run("ReadOne", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := testHandler.testReadOneWithUser(nil, map[string]string{"label": "1"})
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"title":"Label #1"`)
			assert.NotEmpty(t, rec.Result().Header.Get("ETag"))
		})
		t.Run("Nonexisting", func(t *testing.T) {
			// Missing labels return 403, not 404 — the CanRead branch refuses to disclose existence.
			_, err := testHandler.testReadOneWithUser(nil, map[string]string{"label": "9999"})
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Permissions check", func(t *testing.T) {
			t.Run("Forbidden", func(t *testing.T) {
				// Label 6: user13's private label.
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
			// v2 returns 422, not v1's 400; full body shape asserted in TestHuma_ErrorShapeIsRFC9457.
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
			// Update/Delete surface 404 here (isLabelOwner → ErrLabelDoesNotExist),
			// unlike the read branch which returns 403 to hide existence.
			_, err := testHandler.testUpdateWithUser(nil, map[string]string{"label": "9999"}, `{"title":"TestLoremIpsum"}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
		})
		t.Run("Permissions check", func(t *testing.T) {
			t.Run("Forbidden", func(t *testing.T) {
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
			// v2 delete is 204 No Content; v1 returned 200 + a message body.
			assert.Equal(t, http.StatusNoContent, rec.Code)
			assert.Empty(t, rec.Body.String())
		})
		t.Run("Nonexisting", func(t *testing.T) {
			_, err := testHandler.testDeleteWithUser(nil, map[string]string{"label": "9999"})
			require.Error(t, err)
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
		})
		t.Run("Permissions check", func(t *testing.T) {
			t.Run("Forbidden", func(t *testing.T) {
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{"label": "6"})
				require.Error(t, err)
				assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
			})
		})
	})
}

// The two tests below cover v2-only behaviour with no v1 counterpart:
// ETag + conditional requests, and AutoPatch (merge-patch+json).

func TestHumaLabel_ETagReturns304(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	token := humaTokenFor(t, &testuser1)

	rec := humaRequest(t, e, http.MethodGet, "/api/v2/labels/1", "", token, "")
	require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
	etag := rec.Header().Get("ETag")
	require.NotEmpty(t, etag, "GET must return an ETag header")

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
	token := humaTokenFor(t, &testuser1)

	// Create a fresh label so we don't stomp fixtures.
	rec := humaRequest(t, e, http.MethodPost, "/api/v2/labels",
		`{"title":"before","description":"keep me","hex_color":"112233"}`, token, "")
	require.Equal(t, http.StatusCreated, rec.Code, "body: %s", rec.Body.String())
	var created struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created))

	// PATCH only title; AutoPatch must leave description + hex_color alone.
	// Reuses the same echo.Echo so the create above isn't wiped by a fixture reload.
	rec = humaRequest(t, e, http.MethodPatch, fmt.Sprintf("/api/v2/labels/%d", created.ID),
		`{"title":"after"}`, token, "application/merge-patch+json")
	require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())

	rec = humaRequest(t, e, http.MethodGet, fmt.Sprintf("/api/v2/labels/%d", created.ID), "", token, "")
	require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
	var after struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		HexColor    string `json:"hex_color"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &after))
	assert.Equal(t, "after", after.Title)
	assert.Equal(t, "keep me", after.Description, "description must survive the PATCH")
	assert.Equal(t, "112233", after.HexColor, "hex_color must survive the PATCH")
}
