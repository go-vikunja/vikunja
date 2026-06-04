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
// readable side-by-side. Labels has no v1 webtest; coverage is ported 1:1
// from the model-level matrix in pkg/models/label_test.go so the v2 HTTP
// surface independently proves the full visibility/permission contract once
// v1's routes and tests are removed.
//
// Fixture topology the matrix relies on (see pkg/db/fixtures/labels.yml and
// label_tasks.yml):
//   - #1, #2: owned by user1, no task attachment.
//   - #3: owned by user2, no task attachment — invisible to user1.
//   - #4: owned by user2, attached to task #1 in project 1 (user1 is admin),
//     so user1 can READ it (visible via an accessible task) but must NOT be
//     able to update/delete it (not the owner).
//   - #5: owned by user2, attached only to task #35 (inaccessible to user1) —
//     invisible to user1.
//   - #6: owned by user13, attached only to task #34 in private project 20
//     (GHSA-hj5c-mhh2-g7jq regression fixture) — invisible to user1.
//   - #7: owned by user1, no task attachment — readable by its creator.
//   - #8: owned by user1, attached only to inaccessible task #34 — still
//     readable via the creator branch.
func TestHumaLabel(t *testing.T) {
	testHandler := webHandlerTestV2{
		user:     &testuser1,
		basePath: "/api/v2/labels",
		idParam:  "label",
		t:        t,
	}

	t.Run("ReadAll", func(t *testing.T) {
		t.Run("Normal - exact visible set for user1", func(t *testing.T) {
			rec, err := testHandler.testReadAllWithUser(nil, nil)
			require.NoError(t, err)

			ids := labelIDsFromReadAll(t, rec.Body.Bytes())
			// Exact set: user1's own labels (#1, #2, #7, #8) plus #4 which is
			// visible because it is attached to an accessible task. Assert the
			// full set so the cardinality is pinned, not just contains/absent.
			assert.ElementsMatch(t, []int64{1, 2, 4, 7, 8}, ids,
				"ReadAll must return exactly {1,2,4,7,8}; body: %s", rec.Body.String())
			// #5 (other owner, only on inaccessible task) and #6 (GHSA private
			// fixture) must be absent — assert explicitly beyond the set match.
			assert.NotContains(t, ids, int64(3), "label #3 (other owner, unattached) must be hidden")
			assert.NotContains(t, ids, int64(5), "label #5 (other owner, inaccessible task) must be hidden")
			assert.NotContains(t, ids, int64(6), "label #6 (GHSA private fixture) must be hidden")
		})
	})

	t.Run("ReadOne", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := testHandler.testReadOneWithUser(nil, map[string]string{"label": "1"})
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"title":"Label #1"`)
			assert.Contains(t, rec.Body.String(), `"max_permission":`)
			assert.NotEmpty(t, rec.Result().Header.Get("ETag"))
		})
		t.Run("Nonexisting", func(t *testing.T) {
			// Missing labels return 403, not 404 — the CanRead branch refuses to disclose existence.
			_, err := testHandler.testReadOneWithUser(nil, map[string]string{"label": "9999"})
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Permissions check", func(t *testing.T) {
			t.Run("Forbidden - other owner, unattached (#3)", func(t *testing.T) {
				// Label #3: user2's label with no task attachment. user1 is
				// neither owner nor has a task path to it.
				_, err := testHandler.testReadOneWithUser(nil, map[string]string{"label": "3"})
				require.Error(t, err)
				assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
			})
			t.Run("Forbidden - GHSA private label only on unreachable task (#6)", func(t *testing.T) {
				// Label #6: user13's private label, reachable only via task #34
				// in private project 20. GHSA-hj5c-mhh2-g7jq: must stay hidden.
				_, err := testHandler.testReadOneWithUser(nil, map[string]string{"label": "6"})
				require.Error(t, err)
				assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
			})
			t.Run("Allowed - other owner but visible via accessible task (#4)", func(t *testing.T) {
				// GHSA-hj5c-mhh2-g7jq read-vs-write case: #4 is owned by user2
				// but attached to task #1 in a project user1 administers, so
				// READ must succeed even though user1 is not the owner.
				rec, err := testHandler.testReadOneWithUser(nil, map[string]string{"label": "4"})
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Label #4 - visible via other task"`)
				assert.Contains(t, rec.Body.String(), `"id":4`)
			})
			t.Run("Allowed - own label, no task attachment (#7)", func(t *testing.T) {
				// Creator of an unattached label can read it.
				rec, err := testHandler.testReadOneWithUser(nil, map[string]string{"label": "7"})
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Label #7 - created by user 1, no task attachment"`)
			})
			t.Run("Allowed - own label only on inaccessible task (#8)", func(t *testing.T) {
				// Access comes from the creator branch: #8's only label_tasks
				// row points at inaccessible task #34, yet the owner can read it.
				rec, err := testHandler.testReadOneWithUser(nil, map[string]string{"label": "8"})
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Label #8 - user 1 creator, only attached to inaccessible task"`)
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
		t.Run("Hex color is normalized", func(t *testing.T) {
			// NormalizeHex strips a leading '#' (and truncates to 6 chars).
			// Send a non-normalized value and assert the stored/returned form.
			rec, err := testHandler.testCreateWithUser(nil, nil, `{"title":"Normalized","hex_color":"#aabbcc"}`)
			require.NoError(t, err)
			assert.Equal(t, http.StatusCreated, rec.Code)
			assert.Contains(t, rec.Body.String(), `"hex_color":"aabbcc"`,
				"leading '#' must be normalized away; body: %s", rec.Body.String())
			assert.NotContains(t, rec.Body.String(), `#aabbcc`)
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
			t.Run("Forbidden - other owner, unattached (#3)", func(t *testing.T) {
				// Only the owner may update; #3 belongs to user2.
				_, err := testHandler.testUpdateWithUser(nil, map[string]string{"label": "3"}, `{"title":"TestLoremIpsum"}`)
				require.Error(t, err)
				assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
			})
			t.Run("Forbidden - other owner but readable via task (#4)", func(t *testing.T) {
				// GHSA-hj5c-mhh2-g7jq read-vs-write case: #4 is READABLE by user1
				// (visible via an accessible task) but must NOT be updatable —
				// update requires ownership, which user1 does not have.
				_, err := testHandler.testUpdateWithUser(nil, map[string]string{"label": "4"}, `{"title":"TestLoremIpsum"}`)
				require.Error(t, err)
				assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
			})
			t.Run("Forbidden - GHSA private label (#6)", func(t *testing.T) {
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
			t.Run("Forbidden - other owner, unattached (#3)", func(t *testing.T) {
				// Only the owner may delete; #3 belongs to user2.
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{"label": "3"})
				require.Error(t, err)
				assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
			})
			t.Run("Forbidden - other owner but readable via task (#4)", func(t *testing.T) {
				// GHSA-hj5c-mhh2-g7jq read-vs-write case: #4 is READABLE but
				// must NOT be deletable by the non-owner.
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{"label": "4"})
				require.Error(t, err)
				assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
			})
			t.Run("Forbidden - GHSA private label (#6)", func(t *testing.T) {
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{"label": "6"})
				require.Error(t, err)
				assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
			})
		})
	})
}

// labelIDsFromReadAll extracts the label IDs from a v2 paginated list body so
// the visible set can be asserted exactly rather than via substring matching.
func labelIDsFromReadAll(t *testing.T, body []byte) []int64 {
	t.Helper()
	var resp struct {
		Items []struct {
			ID int64 `json:"id"`
		} `json:"items"`
	}
	require.NoError(t, json.Unmarshal(body, &resp), "ReadAll body must be a paginated envelope: %s", string(body))
	ids := make([]int64, 0, len(resp.Items))
	for _, it := range resp.Items {
		ids = append(ids, it.ID)
	}
	return ids
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

func TestHumaLabel_ETagReflectsPermission(t *testing.T) {
	// Label #4 is owned by user2 (admin) but readable by user1 only at read level;
	// same label, so the per-caller ETag must differ — else a 304 serves stale perms.
	e, err := setupTestEnv()
	require.NoError(t, err)

	reader := humaRequest(t, e, http.MethodGet, "/api/v2/labels/4", "", humaTokenFor(t, &testuser1), "")
	require.Equal(t, http.StatusOK, reader.Code, "body: %s", reader.Body.String())
	owner := humaRequest(t, e, http.MethodGet, "/api/v2/labels/4", "", humaTokenFor(t, &testuser2), "")
	require.Equal(t, http.StatusOK, owner.Code, "body: %s", owner.Body.String())

	assert.NotEmpty(t, reader.Header().Get("ETag"))
	assert.NotEqual(t, reader.Header().Get("ETag"), owner.Header().Get("ETag"),
		"same label, different caller permission must produce different ETags")
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
