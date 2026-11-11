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
	"net/url"
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testuser22 is the second bot owner from pkg/db/fixtures/users.yml; user22
// owns bot 24. Paired with testuser21 to assert bot-owner isolation: each
// owner sees and acts on their own bots' resources, never the other's.
var testuser22 = user.User{ID: 22, Username: "user_bot_owner_b", Issuer: "local"}

// Paired to assert a bot inherits its own owner's labels and nobody else's.
var testbot23 = user.User{ID: 23, Username: "bot-owner-a-assistant", Issuer: "local", BotOwnerID: 21}
var testbot24 = user.User{ID: 24, Username: "bot-owner-b-assistant", Issuer: "local", BotOwnerID: 22}

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
//   - #5: owned by user2, attached to task #35 in project 21, which user1 owns
//     but has archived — archiving does not hide a project's tasks, so user1
//     can read it.
//   - #6: owned by user13, attached only to task #34 in private project 20
//     (GHSA-hj5c-mhh2-g7jq regression fixture) — invisible to user1.
//   - #7: owned by user1, no task attachment — readable by its creator.
//   - #8: owned by user1, attached only to inaccessible task #34 — still
//     readable via the creator branch.
//   - #10: owned by user6, attached only to task #25 in project 16, a child of
//     project 33 which is shared to team 1 (user1 is a member) — visible via
//     the inherited child-project access.
//   - #13: owned by user13, attached only to the soft-deleted task #51 in
//     project 1 — invisible to user1 despite the accessible project.
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
			assert.ElementsMatch(t, []int64{1, 2, 4, 5, 7, 8, 10}, ids,
				"ReadAll must return exactly {1,2,4,5,7,8,10}; body: %s", rec.Body.String())
		})

		t.Run("Pagination - full page forces the Count() fallback", func(t *testing.T) {
			rec, err := testHandler.testReadAllWithUser(url.Values{"page": {"1"}, "per_page": {"3"}}, nil)
			require.NoError(t, err)
			ids := labelIDsFromReadAll(t, rec.Body.Bytes())
			assert.Equal(t, []int64{1, 2, 4}, ids,
				"a full page 1 of 3 must slice the first 3 ids in order; body: %s", rec.Body.String())
			assert.EqualValues(t, 7, totalFromReadAll(t, rec.Body.Bytes()),
				"a full unfiltered page must report the true total via Count(); body: %s", rec.Body.String())
		})
		t.Run("Pagination - partial last page uses the start+len shortcut", func(t *testing.T) {
			rec, err := testHandler.testReadAllWithUser(url.Values{"page": {"3"}, "per_page": {"3"}}, nil)
			require.NoError(t, err)
			ids := labelIDsFromReadAll(t, rec.Body.Bytes())
			assert.Equal(t, []int64{10}, ids,
				"the short last page must contain only the trailing id; body: %s", rec.Body.String())
			assert.EqualValues(t, 7, totalFromReadAll(t, rec.Body.Bytes()),
				"a short last page must still report the true total; body: %s", rec.Body.String())
		})
		t.Run("Pagination - page past the end still reports the true total", func(t *testing.T) {
			rec, err := testHandler.testReadAllWithUser(url.Values{"page": {"4"}, "per_page": {"3"}}, nil)
			require.NoError(t, err)
			ids := labelIDsFromReadAll(t, rec.Body.Bytes())
			assert.Empty(t, ids, "a page past the end must return no items; body: %s", rec.Body.String())
			assert.EqualValues(t, 7, totalFromReadAll(t, rec.Body.Bytes()),
				"an empty page with a non-zero offset must fall back to Count(), not report 0 or the offset; body: %s", rec.Body.String())
		})
		t.Run("Pagination - full non-last page respects the q filter", func(t *testing.T) {
			// q keeps 4 labels, so a full page 1 (2 of 4) isn't the last page - start+len would wrongly report 2.
			rec, err := testHandler.testReadAllWithUser(url.Values{"q": {"1,2,4,5"}, "page": {"1"}, "per_page": {"2"}}, nil)
			require.NoError(t, err)
			ids := labelIDsFromReadAll(t, rec.Body.Bytes())
			assert.Equal(t, []int64{1, 2}, ids,
				"a full filtered page 1 of 2 must slice the first 2 filtered ids; body: %s", rec.Body.String())
			assert.EqualValues(t, 4, totalFromReadAll(t, rec.Body.Bytes()),
				"a full non-last page must report the true filtered total, not start+len; body: %s", rec.Body.String())
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
			t.Run("Forbidden - only on a soft-deleted task in an accessible project (#13)", func(t *testing.T) {
				// Task #51 is soft-deleted but in project 1 (user1's), so this only passes if the deleted-task filter runs.
				_, err := testHandler.testReadOneWithUser(nil, map[string]string{"label": "13"})
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

// TestHumaLabel_BotOwner asserts that bot owners can read, update, and delete
// labels that were created by bots they own. Fixture label #9 is owned by
// bot 23, whose owner is user 21 (testuser21); user 22 owns a different bot
// and must not see or touch it.
func TestHumaLabel_BotOwner(t *testing.T) {
	botOwner := webHandlerTestV2{
		user:     &testuser21,
		basePath: "/api/v2/labels",
		idParam:  "label",
		t:        t,
	}
	require.NoError(t, botOwner.ensureEnv())
	otherOwner := webHandlerTestV2{
		user:     &testuser22,
		basePath: "/api/v2/labels",
		idParam:  "label",
		t:        t,
		e:        botOwner.e,
	}

	t.Run("ReadOne - bot owner can read label created by their bot", func(t *testing.T) {
		rec, err := botOwner.testReadOneWithUser(nil, map[string]string{"label": "9"})
		require.NoError(t, err)
		assert.Contains(t, rec.Body.String(), `"title":"Label #9 - created by bot 23 owned by user 21"`)
	})
	t.Run("ReadOne - non-owner cannot read another owner's bot's label", func(t *testing.T) {
		_, err := otherOwner.testReadOneWithUser(nil, map[string]string{"label": "9"})
		require.Error(t, err)
		assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
	})
	t.Run("ReadAll - bot owner's listing surfaces their bot's labels", func(t *testing.T) {
		rec, err := botOwner.testReadAllWithUser(nil, nil)
		require.NoError(t, err)
		ids := labelIDsFromReadAll(t, rec.Body.Bytes())
		assert.Contains(t, ids, int64(9), "label #9 (created by user 21's bot) must be listed")
	})
	t.Run("Update - bot owner can update label created by their bot", func(t *testing.T) {
		rec, err := botOwner.testUpdateWithUser(nil, map[string]string{"label": "9"}, `{"title":"renamed by owner"}`)
		require.NoError(t, err)
		assert.Contains(t, rec.Body.String(), `"title":"renamed by owner"`)
	})
	t.Run("Update - non-owner cannot update another owner's bot's label", func(t *testing.T) {
		_, err := otherOwner.testUpdateWithUser(nil, map[string]string{"label": "9"}, `{"title":"hijack"}`)
		require.Error(t, err)
		assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
	})
	t.Run("Delete - non-owner cannot delete another owner's bot's label", func(t *testing.T) {
		_, err := otherOwner.testDeleteWithUser(nil, map[string]string{"label": "9"})
		require.Error(t, err)
		assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
	})
	t.Run("Delete - bot owner can delete label created by their bot", func(t *testing.T) {
		// Run last so the earlier subtests still have label #9 to operate on.
		rec, err := botOwner.testDeleteWithUser(nil, map[string]string{"label": "9"})
		require.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, rec.Code)
	})
}

// Unattached labels #11 and #12 isolate identity-based access (#3592).
func TestHumaLabel_BotUsesOwnerLabel(t *testing.T) {
	bot := webHandlerTestV2{
		user:     &testbot23,
		basePath: "/api/v2/labels",
		idParam:  "label",
		t:        t,
	}
	require.NoError(t, bot.ensureEnv())
	otherBot := webHandlerTestV2{
		user:     &testbot24,
		basePath: "/api/v2/labels",
		idParam:  "label",
		t:        t,
		e:        bot.e,
	}

	t.Run("ReadAll - bot's listing surfaces its owner's and its siblings' unattached labels", func(t *testing.T) {
		rec, err := bot.testReadAllWithUser(nil, nil)
		require.NoError(t, err)
		ids := labelIDsFromReadAll(t, rec.Body.Bytes())
		assert.ElementsMatch(t, []int64{9, 11, 12}, ids,
			"bot's ReadAll must return exactly {9,11,12}; body: %s", rec.Body.String())
	})
	t.Run("ReadOne - bot can read a label created by a sibling bot", func(t *testing.T) {
		rec, err := bot.testReadOneWithUser(nil, map[string]string{"label": "12"})
		require.NoError(t, err)
		assert.Contains(t, rec.Body.String(), `"title":"Label #12 - created by bot 25, sibling of bot 23"`)
	})
	t.Run("ReadOne - a different owner's bot cannot read the sibling's label", func(t *testing.T) {
		_, err := otherBot.testReadOneWithUser(nil, map[string]string{"label": "12"})
		require.Error(t, err)
		assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
	})
	t.Run("ReadOne - bot can read its owner's unattached label", func(t *testing.T) {
		rec, err := bot.testReadOneWithUser(nil, map[string]string{"label": "11"})
		require.NoError(t, err)
		assert.Contains(t, rec.Body.String(), `"title":"Label #11 - created by user 21, owner of bot 23, no task attachment"`)
	})
	t.Run("ReadOne - a different owner's bot cannot read it", func(t *testing.T) {
		_, err := otherBot.testReadOneWithUser(nil, map[string]string{"label": "11"})
		require.Error(t, err)
		assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
	})
	t.Run("ReadAll - a different owner's bot's listing does not surface it", func(t *testing.T) {
		rec, err := otherBot.testReadAllWithUser(nil, nil)
		require.NoError(t, err)
		ids := labelIDsFromReadAll(t, rec.Body.Bytes())
		assert.Empty(t, ids, "other owner's bot must see no labels; body: %s", rec.Body.String())
	})
	t.Run("Update - bot cannot rename its owner's label", func(t *testing.T) {
		_, err := bot.testUpdateWithUser(nil, map[string]string{"label": "11"}, `{"title":"renamed by bot"}`)
		require.Error(t, err)
		assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
	})

	attach := webHandlerTestV2{
		user:     &testbot23,
		basePath: "/api/v2/tasks/52/labels",
		idParam:  "label",
		t:        t,
		e:        bot.e,
	}
	t.Run("Create - bot can attach its owner's never-used label", func(t *testing.T) {
		rec, err := attach.testCreateWithUser(nil, nil, `{"label_id":11}`)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Contains(t, rec.Body.String(), `"label_id":11`)
	})
	t.Run("Create - bot can attach a never-used label created by a sibling bot", func(t *testing.T) {
		rec, err := attach.testCreateWithUser(nil, nil, `{"label_id":12}`)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Contains(t, rec.Body.String(), `"label_id":12`)
	})
	t.Run("Create - bot cannot attach an unrelated user's label", func(t *testing.T) {
		_, err := attach.testCreateWithUser(nil, nil, `{"label_id":6}`)
		require.Error(t, err)
		assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
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

func totalFromReadAll(t *testing.T, body []byte) int64 {
	t.Helper()
	var resp struct {
		Total int64 `json:"total"`
	}
	require.NoError(t, json.Unmarshal(body, &resp), "ReadAll body must be a paginated envelope: %s", string(body))
	return resp.Total
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
