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
	"net/http"
	"net/url"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHumaTaskComment ports the v1 webtest coverage (TestTaskComments +
// TestTaskCommentIDOR) to /api/v2, plus v2-specific HTTP assertions (status
// codes, ETag). It re-proves the full permission/sharing matrix independently
// because the v1 routes and their tests will be removed.
//
// The crux of the author-only rule: across tasks 15–26, testuser1 is granted
// access through every share kind but never authored the comments (user 5/6
// did), so a 403 there exercises authorship rather than plain access denial.
func TestHumaTaskComment(t *testing.T) {
	// task 1 belongs to project 1, owned by testuser1.
	onTask1 := webHandlerTestV2{
		user:     &testuser1,
		basePath: "/api/v2/tasks/1/comments",
		idParam:  "commentid",
		t:        t,
	}
	require.NoError(t, onTask1.ensureEnv())
	// onTaskAs reuses the one Echo instance (and its single fixture load) for a
	// different task. v2 does not reload fixtures per request, so the subtests
	// are ordered to avoid clobbering each other's rows.
	onTaskAs := func(taskID string, u *user.User) *webHandlerTestV2 {
		return &webHandlerTestV2{
			user:     u,
			basePath: "/api/v2/tasks/" + taskID + "/comments",
			idParam:  "commentid",
			t:        t,
			e:        onTask1.e,
		}
	}
	// task 35 also belongs to testuser1.
	onTask35 := onTaskAs("35", &testuser1)
	// task 2 also belongs to project 1; used for the wrong-parent negative.
	onTask2 := onTaskAs("2", &testuser1)
	// user6 has no access to project 1, so it is neither author nor writer on
	// task 1's comment 1 — used for the no-access forbidden negatives.
	asUser6 := onTaskAs("1", &testuser6)

	t.Run("ReadAll", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := onTask1.testReadAllWithUser(nil, nil)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `Lorem Ipsum Dolor Sit Amet`)
			// comments from other tasks must not leak in.
			assert.NotContains(t, rec.Body.String(), `comment 2`)
		})
		t.Run("Link share author resolves", func(t *testing.T) {
			rec, err := onTask35.testReadAllWithUser(nil, nil)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `comment 15`)
			assert.Contains(t, rec.Body.String(), `comment 17`)
		})
		t.Run("order_by desc", func(t *testing.T) {
			rec, err := onTask35.testReadAllWithUser(url.Values{"order_by": []string{"desc"}}, nil)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `comment 15`)
		})
		t.Run("Search filter", func(t *testing.T) {
			// Mirrors the v1 model ReadAll search test: search is case-insensitive.
			rec, err := onTask35.testReadAllWithUser(url.Values{"q": []string{"COMMENT 15"}}, nil)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `comment 15`)
			assert.NotContains(t, rec.Body.String(), `comment 17`)
		})
		t.Run("Forbidden", func(t *testing.T) {
			// user6 cannot read task 1.
			_, err := asUser6.testReadAllWithUser(nil, nil)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
	})

	t.Run("ReadOne", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := onTask1.testReadOneWithUser(nil, map[string]string{"commentid": "1"})
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `Lorem Ipsum Dolor Sit Amet`)
			assert.Contains(t, rec.Body.String(), `"id":1`)
			assert.NotEmpty(t, rec.Result().Header.Get("ETag"))
		})
		t.Run("Nonexisting", func(t *testing.T) {
			_, err := onTask1.testReadOneWithUser(nil, map[string]string{"commentid": "9999"})
			require.Error(t, err)
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
		})
		t.Run("Comment from another task", func(t *testing.T) {
			// comment 1 belongs to task 1; reading it under task 2 must 404.
			_, err := onTask2.testReadOneWithUser(nil, map[string]string{"commentid": "1"})
			require.Error(t, err)
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
		})
		t.Run("IDOR via accessible task", func(t *testing.T) {
			// Port of v1 TestTaskCommentIDOR: comment 18 belongs to task 34
			// (owned by user 13, inaccessible to testuser1). Task 1 is
			// accessible to testuser1. Requesting it under task 1 must 404 with
			// the comment-does-not-exist code (not leak the inaccessible row).
			_, err := onTask1.testReadOneWithUser(nil, map[string]string{"commentid": "18"})
			require.Error(t, err)
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
			assertHandlerErrorCode(t, err, models.ErrCodeTaskCommentDoesNotExist)
		})
		t.Run("Forbidden", func(t *testing.T) {
			_, err := asUser6.testReadOneWithUser(nil, map[string]string{"commentid": "1"})
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
	})

	t.Run("Create", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := onTask1.testCreateWithUser(nil, nil, `{"comment":"A brand new comment"}`)
			require.NoError(t, err)
			assert.Equal(t, http.StatusCreated, rec.Code)
			assert.Contains(t, rec.Body.String(), `"comment":"A brand new comment"`)
			// author is set server-side from the authenticated user.
			assert.Contains(t, rec.Body.String(), `"username":"user1"`)
		})
		t.Run("Nonexisting task", func(t *testing.T) {
			// Creating a comment on a task that does not exist surfaces the
			// task-does-not-exist domain error as a 404.
			onMissing := onTaskAs("9999", &testuser1)
			_, err := onMissing.testCreateWithUser(nil, nil, `{"comment":"Lorem Ipsum"}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
			assertHandlerErrorCode(t, err, models.ErrCodeTaskDoesNotExist)
		})
		t.Run("Forbidden", func(t *testing.T) {
			// user6 has no write access to task 1.
			_, err := asUser6.testCreateWithUser(nil, nil, `{"comment":"Nope"}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})

		// Permission matrix: CREATE requires write access to the task, so
		// read-only shares are forbidden while write/admin shares are allowed.
		// These mirror v1 TestTaskComments/Create/Permissions_check exactly.
		t.Run("Permissions check", func(t *testing.T) {
			// task 34 is owned by user 13 — testuser1 has no access at all.
			t.Run("Forbidden", func(t *testing.T) {
				_, err := onTaskAs("34", &testuser1).testCreateWithUser(nil, nil, `{"comment":"Lorem Ipsum"}`)
				require.Error(t, err)
				assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
			})

			// Read-only shares: create forbidden.
			forbiddenCreate := map[string]string{
				"Shared Via Team readonly":                "15",
				"Shared Via User readonly":                "18",
				"Shared Via Parent Project Team readonly": "21",
				"Shared Via Parent Project User readonly": "24",
			}
			for name, taskID := range forbiddenCreate {
				t.Run(name, func(t *testing.T) {
					_, err := onTaskAs(taskID, &testuser1).testCreateWithUser(nil, nil, `{"comment":"Lorem Ipsum"}`)
					require.Error(t, err)
					assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
				})
			}

			// Write/admin shares: create allowed (8 positive cases).
			allowedCreate := map[string]string{
				"Shared Via Team write":                "16",
				"Shared Via Team admin":                "17",
				"Shared Via User write":                "19",
				"Shared Via User admin":                "20",
				"Shared Via Parent Project Team write": "22",
				"Shared Via Parent Project Team admin": "23",
				"Shared Via Parent Project User write": "25",
				"Shared Via Parent Project User admin": "26",
			}
			for name, taskID := range allowedCreate {
				t.Run(name, func(t *testing.T) {
					rec, err := onTaskAs(taskID, &testuser1).testCreateWithUser(nil, nil, `{"comment":"Lorem Ipsum"}`)
					require.NoError(t, err)
					assert.Equal(t, http.StatusCreated, rec.Code)
					assert.Contains(t, rec.Body.String(), `"comment":"Lorem Ipsum"`)
				})
			}
		})

		t.Run("Link Share", func(t *testing.T) {
			// Port of v1 TestTaskComments/Create/Link_Share: link share id 2 has
			// write access to project 2 (task 13). The created comment is
			// attributed to the synthetic link-share user (author_id == -2,
			// i.e. share.ID * -1). Driven through the full Huma stack with a
			// real link-share JWT so v2's auth bridging is exercised too.
			token, err := auth.NewLinkShareJWTAuthtoken(&models.LinkSharing{
				ID:          2,
				Hash:        "test2",
				ProjectID:   2,
				Permission:  models.PermissionWrite,
				SharingType: models.SharingTypeWithoutPassword,
				SharedByID:  1,
			})
			require.NoError(t, err)
			rec := humaRequest(t, onTask1.e, http.MethodPost, "/api/v2/tasks/13/comments", `{"comment":"Lorem Ipsum"}`, token, "")
			require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
			assert.Contains(t, rec.Body.String(), `"comment":"Lorem Ipsum"`)
			db.AssertExists(t, "task_comments", map[string]interface{}{
				"task_id":   13,
				"comment":   "Lorem Ipsum",
				"author_id": -2,
			}, false)
		})
	})

	t.Run("Update", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := onTask1.testUpdateWithUser(nil, map[string]string{"commentid": "1"}, `{"comment":"Edited comment"}`)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"comment":"Edited comment"`)
		})
		t.Run("Nonexisting task", func(t *testing.T) {
			// v1 used task 99999 / comment 9999: no access to the task yields
			// the task-does-not-exist error rather than leaking the comment.
			onMissing := onTaskAs("99999", &testuser1)
			_, err := onMissing.testUpdateWithUser(nil, map[string]string{"commentid": "9999"}, `{"comment":"Lorem Ipsum"}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
			assertHandlerErrorCode(t, err, models.ErrCodeTaskDoesNotExist)
		})
		t.Run("Nonexisting comment on accessible task", func(t *testing.T) {
			// commentid 9999 under task 1 (writable, owned by testuser1):
			// the comment lookup fails after the write check, so this 404s with
			// the comment-does-not-exist code.
			_, err := onTask1.testUpdateWithUser(nil, map[string]string{"commentid": "9999"}, `{"comment":"x"}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
			assertHandlerErrorCode(t, err, models.ErrCodeTaskCommentDoesNotExist)
		})
		t.Run("Comment from another task", func(t *testing.T) {
			// comment 1 is on task 1; updating it under task 2 must 404.
			_, err := onTask2.testUpdateWithUser(nil, map[string]string{"commentid": "1"}, `{"comment":"x"}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
		})
		t.Run("Forbidden no access", func(t *testing.T) {
			// user6 has no access to task 1 at all.
			_, err := asUser6.testUpdateWithUser(nil, map[string]string{"commentid": "1"}, `{"comment":"x"}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})

		// Author-only matrix: even a user WITH write/admin access who is NOT the
		// comment author is forbidden from updating it. testuser1 has access to
		// every share kind below (read/write/admin) but authored none of
		// comments 3–14, so every case must 403. This is the distinctive
		// author-only rule and the heart of the v1 Update/Permissions matrix.
		t.Run("Permissions check (author-only)", func(t *testing.T) {
			cases := map[string]struct {
				task    string
				comment string
			}{
				"Forbidden":                               {"14", "2"},
				"Shared Via Team readonly":                {"15", "3"},
				"Shared Via Team write":                   {"16", "4"},
				"Shared Via Team admin":                   {"17", "5"},
				"Shared Via User readonly":                {"18", "6"},
				"Shared Via User write":                   {"19", "7"},
				"Shared Via User admin":                   {"20", "8"},
				"Shared Via Parent Project Team readonly": {"21", "9"},
				"Shared Via Parent Project Team write":    {"22", "10"},
				"Shared Via Parent Project Team admin":    {"23", "11"},
				"Shared Via Parent Project User readonly": {"24", "12"},
				"Shared Via Parent Project User write":    {"25", "13"},
				"Shared Via Parent Project User admin":    {"26", "14"},
			}
			for name, c := range cases {
				t.Run(name, func(t *testing.T) {
					_, err := onTaskAs(c.task, &testuser1).testUpdateWithUser(nil, map[string]string{"commentid": c.comment}, `{"comment":"Lorem Ipsum"}`)
					require.Error(t, err)
					assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
				})
			}
		})
	})

	t.Run("Delete", func(t *testing.T) {
		t.Run("Nonexisting task", func(t *testing.T) {
			onMissing := onTaskAs("99999", &testuser1)
			_, err := onMissing.testDeleteWithUser(nil, map[string]string{"commentid": "9999"})
			require.Error(t, err)
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
			assertHandlerErrorCode(t, err, models.ErrCodeTaskDoesNotExist)
		})
		t.Run("Nonexisting comment on accessible task", func(t *testing.T) {
			_, err := onTask1.testDeleteWithUser(nil, map[string]string{"commentid": "9999"})
			require.Error(t, err)
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
			assertHandlerErrorCode(t, err, models.ErrCodeTaskCommentDoesNotExist)
		})
		t.Run("Comment from another task", func(t *testing.T) {
			_, err := onTask2.testDeleteWithUser(nil, map[string]string{"commentid": "1"})
			require.Error(t, err)
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
		})
		t.Run("Forbidden no access", func(t *testing.T) {
			_, err := asUser6.testDeleteWithUser(nil, map[string]string{"commentid": "1"})
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})

		// Author-only matrix for delete, mirroring the update matrix above:
		// write/admin access is not enough — only the author may delete.
		t.Run("Permissions check (author-only)", func(t *testing.T) {
			cases := map[string]struct {
				task    string
				comment string
			}{
				"Forbidden":                               {"14", "2"},
				"Shared Via Team readonly":                {"15", "3"},
				"Shared Via Team write":                   {"16", "4"},
				"Shared Via Team admin":                   {"17", "5"},
				"Shared Via User readonly":                {"18", "6"},
				"Shared Via User write":                   {"19", "7"},
				"Shared Via User admin":                   {"20", "8"},
				"Shared Via Parent Project Team readonly": {"21", "9"},
				"Shared Via Parent Project Team write":    {"22", "10"},
				"Shared Via Parent Project Team admin":    {"23", "11"},
				"Shared Via Parent Project User readonly": {"24", "12"},
				"Shared Via Parent Project User write":    {"25", "13"},
				"Shared Via Parent Project User admin":    {"26", "14"},
			}
			for name, c := range cases {
				t.Run(name, func(t *testing.T) {
					_, err := onTaskAs(c.task, &testuser1).testDeleteWithUser(nil, map[string]string{"commentid": c.comment})
					require.Error(t, err)
					assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
				})
			}
		})

		t.Run("Normal", func(t *testing.T) {
			// Run last: comment 1 is the author's own, so this succeeds and
			// removes the fixture row used by the read/update cases above.
			rec, err := onTask1.testDeleteWithUser(nil, map[string]string{"commentid": "1"})
			require.NoError(t, err)
			assert.Equal(t, http.StatusNoContent, rec.Code)
			assert.Empty(t, rec.Body.String())
		})
	})
}
