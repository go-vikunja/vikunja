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
	"strconv"
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Bot ownership fixtures (pkg/db/fixtures/users.yml):
//   - user 21 (user_bot_owner_a) owns bot 23 (bot-owner-a-assistant).
//   - user 22 (user_bot_owner_b) owns bot 24 (bot-owner-b-assistant).
//
// These two owner/bot pairs give a clean matrix: every read/update/delete of a
// bot the caller does not own must be refused, and a caller's own bot must be
// reachable. Constructed locally (not in integrations.go) so this test is
// self-contained.
var (
	botOwnerA = user.User{ID: 21, Username: "user_bot_owner_a"}
	botOwnerB = user.User{ID: 22, Username: "user_bot_owner_b"}
)

// TestHumaBotUser ports the v1 bot-user permission matrix to the v2 HTTP
// surface 1:1 (the v1 coverage lives in pkg/models/bot_users_test.go; there is
// no v1 webtest). Unlike labels, ownership is verified by loading the user, so
// every unowned/nonexistent read/update/delete is refused with 403 — there is
// no 404 branch.
//
// One shared env (one fixture load, one signing secret) backs every request;
// the caller is swapped via h.user. A second env would regenerate the random
// service secret and invalidate the first env's JWTs.
func TestHumaBotUser(t *testing.T) {
	h := webHandlerTestV2{
		user:     &botOwnerA,
		basePath: "/api/v2/user/bots",
		idParam:  "bot",
		t:        t,
	}
	// asOwnerB runs fn with the caller temporarily switched to user 22.
	asOwnerB := func(fn func()) {
		h.user = &botOwnerB
		defer func() { h.user = &botOwnerA }()
		fn()
	}

	t.Run("ReadAll", func(t *testing.T) {
		t.Run("Normal - only own bots", func(t *testing.T) {
			rec, err := h.testReadAllWithUser(nil, nil)
			require.NoError(t, err)
			ids := botIDsFromReadAll(t, rec.Body.Bytes())
			// user 21 owns exactly bot 23; user 22's bot 24 must never leak.
			assert.ElementsMatch(t, []int64{23}, ids,
				"ReadAll must return exactly {23}; body: %s", rec.Body.String())
			assert.NotContains(t, ids, int64(24), "bot #24 (other owner) must be hidden")
		})
		t.Run("Search filters by username", func(t *testing.T) {
			rec, err := h.testReadAllWithUser(map[string][]string{"q": {"nomatch-xyz"}}, nil)
			require.NoError(t, err)
			ids := botIDsFromReadAll(t, rec.Body.Bytes())
			assert.Empty(t, ids, "a non-matching search must return no bots; body: %s", rec.Body.String())
		})
	})

	t.Run("ReadOne", func(t *testing.T) {
		t.Run("Normal - owner", func(t *testing.T) {
			rec, err := h.testReadOneWithUser(nil, map[string]string{"bot": "23"})
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"username":"bot-owner-a-assistant"`)
			assert.Contains(t, rec.Body.String(), `"bot_owner_id":21`)
			assert.Contains(t, rec.Body.String(), `"max_permission":`)
			assert.NotEmpty(t, rec.Result().Header.Get("ETag"))
		})
		t.Run("Forbidden - other owner (#24)", func(t *testing.T) {
			_, err := h.testReadOneWithUser(nil, map[string]string{"bot": "24"})
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Nonexisting refuses with 403", func(t *testing.T) {
			// Ownership is resolved by loading the user; a missing bot is
			// indistinguishable from one owned by someone else, so it is 403,
			// not 404 — existence is never disclosed.
			_, err := h.testReadOneWithUser(nil, map[string]string{"bot": "999999"})
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
	})

	t.Run("Create", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := h.testCreateWithUser(nil, nil, `{"username":"bot-create-success","name":"Created Bot"}`)
			require.NoError(t, err)
			assert.Equal(t, http.StatusCreated, rec.Code)
			assert.Contains(t, rec.Body.String(), `"username":"bot-create-success"`)
			assert.Contains(t, rec.Body.String(), `"name":"Created Bot"`)
			// The creating user becomes the owner.
			assert.Contains(t, rec.Body.String(), `"bot_owner_id":21`)
			// Bots are created active and carry no email.
			assert.Contains(t, rec.Body.String(), `"status":0`)
			assert.NotContains(t, rec.Body.String(), `"email":`)
		})
		t.Run("Missing bot- prefix", func(t *testing.T) {
			_, err := h.testCreateWithUser(nil, nil, `{"username":"no-prefix-bot"}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusBadRequest, getHTTPErrorCode(err))
		})
		t.Run("Empty username", func(t *testing.T) {
			// minLength:"1" makes Huma reject the body before the model runs (422).
			_, err := h.testCreateWithUser(nil, nil, `{"username":""}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusUnprocessableEntity, getHTTPErrorCode(err))
		})
		t.Run("Username with spaces", func(t *testing.T) {
			// ErrUsernameMustNotContainSpaces maps to 412, matching v1.
			_, err := h.testCreateWithUser(nil, nil, `{"username":"bot- with space"}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusPreconditionFailed, getHTTPErrorCode(err))
		})
		t.Run("Duplicate username", func(t *testing.T) {
			// bot-owner-a-assistant already exists (bot #23).
			_, err := h.testCreateWithUser(nil, nil, `{"username":"bot-owner-a-assistant"}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusBadRequest, getHTTPErrorCode(err))
		})
	})

	t.Run("Update", func(t *testing.T) {
		t.Run("Normal - rename owned bot", func(t *testing.T) {
			// Renames bot 23 but keeps it active so the Delete cases below can
			// still reach it (disabling poisons GetUserByID with a 412).
			rec, err := h.testUpdateWithUser(nil, map[string]string{"bot": "23"},
				`{"name":"Renamed Bot"}`)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"name":"Renamed Bot"`)
			assert.Contains(t, rec.Body.String(), `"status":0`)
		})
		t.Run("Rename owned bot's username", func(t *testing.T) {
			// A new username must keep the bot- prefix.
			rec, err := h.testUpdateWithUser(nil, map[string]string{"bot": "23"},
				`{"username":"bot-owner-a-renamed"}`)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"username":"bot-owner-a-renamed"`)
		})
		t.Run("Disable sets status; bot then resolves as disabled (412)", func(t *testing.T) {
			// Disabling is allowed, but once disabled GetUserByID surfaces
			// ErrAccountDisabled, so a follow-up read fails the precondition (412)
			// — same as v1. Use a throwaway bot so bot 23 stays usable.
			rec, err := h.testCreateWithUser(nil, nil, `{"username":"bot-to-disable"}`)
			require.NoError(t, err)
			id := botID(t, rec.Body.Bytes())

			rec, err = h.testUpdateWithUser(nil, map[string]string{"bot": id}, `{"status":2}`)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"status":2`)

			_, err = h.testReadOneWithUser(nil, map[string]string{"bot": id})
			require.Error(t, err)
			assert.Equal(t, http.StatusPreconditionFailed, getHTTPErrorCode(err))
		})
		t.Run("Forbidden - other owner (#24)", func(t *testing.T) {
			_, err := h.testUpdateWithUser(nil, map[string]string{"bot": "24"}, `{"name":"Nope"}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Nonexisting refuses with 403", func(t *testing.T) {
			_, err := h.testUpdateWithUser(nil, map[string]string{"bot": "999999"}, `{"name":"Nope"}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
	})

	t.Run("Delete", func(t *testing.T) {
		t.Run("Forbidden - other owner (#23)", func(t *testing.T) {
			// user 22 does not own bot 23.
			asOwnerB(func() {
				_, err := h.testDeleteWithUser(nil, map[string]string{"bot": "23"})
				require.Error(t, err)
				assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
			})
		})
		t.Run("Nonexisting refuses with 403", func(t *testing.T) {
			_, err := h.testDeleteWithUser(nil, map[string]string{"bot": "999999"})
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Normal", func(t *testing.T) {
			// Runs last so the deleted bot doesn't disturb the assertions above.
			rec, err := h.testDeleteWithUser(nil, map[string]string{"bot": "23"})
			require.NoError(t, err)
			assert.Equal(t, http.StatusNoContent, rec.Code)
			assert.Empty(t, rec.Body.String())
		})
	})
}

// The two tests below cover v2-only behaviour with no v1 counterpart:
// ETag + conditional requests, and AutoPatch (merge-patch+json).

func TestHumaBotUser_ETagReturns304(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	token := humaTokenFor(t, &botOwnerA)

	rec := humaRequest(t, e, http.MethodGet, "/api/v2/user/bots/23", "", token, "")
	require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
	etag := rec.Header().Get("ETag")
	require.NotEmpty(t, etag, "GET must return an ETag header")

	req := httptest.NewRequest(http.MethodGet, "/api/v2/user/bots/23", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("If-None-Match", etag)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNotModified, rec.Code, "body: %s", rec.Body.String())
}

func TestHumaBotUser_PATCHMergePatch(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	token := humaTokenFor(t, &botOwnerA)

	// Create a fresh bot so we don't stomp fixtures.
	rec := humaRequest(t, e, http.MethodPost, "/api/v2/user/bots",
		`{"username":"bot-patch-target","name":"keep me"}`, token, "")
	require.Equal(t, http.StatusCreated, rec.Code, "body: %s", rec.Body.String())
	id := botID(t, rec.Body.Bytes())

	// PATCH only the username; AutoPatch must leave name untouched.
	rec = humaRequest(t, e, http.MethodPatch, "/api/v2/user/bots/"+id,
		`{"username":"bot-patched"}`, token, "application/merge-patch+json")
	require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())

	rec = humaRequest(t, e, http.MethodGet, "/api/v2/user/bots/"+id, "", token, "")
	require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
	var after struct {
		Username string `json:"username"`
		Name     string `json:"name"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &after))
	assert.Equal(t, "bot-patched", after.Username)
	assert.Equal(t, "keep me", after.Name, "name must survive the PATCH")
}

// botID extracts the id from a single-bot response body as a path string.
func botID(t *testing.T, body []byte) string {
	t.Helper()
	var resp struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(body, &resp), "body must carry an id: %s", string(body))
	require.NotZero(t, resp.ID, "created bot must have an id: %s", string(body))
	return strconv.FormatInt(resp.ID, 10)
}

// botIDsFromReadAll extracts the bot user IDs from a v2 paginated list body so
// the owned set can be asserted exactly rather than via substring matching.
func botIDsFromReadAll(t *testing.T, body []byte) []int64 {
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
