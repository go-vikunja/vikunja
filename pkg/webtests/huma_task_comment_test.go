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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHumaTaskComment is the nested feature-gated reference test for /api/v2.
// Comments live under /tasks/{task}/comments/{commentid}, so the harness binds
// two path params: basePath carries the literal {task} and idParam picks
// {commentid}.
//
// The resource is gated behind config.ServiceEnableTaskComments, which
// InitDefaultConfig (called by setupTestEnv) defaults to true — so the
// registrar registers the routes and these tests can reach them.
//
// Fixtures: task 1 (project 1, owned by testuser1) has comment 1 (author 1).
// task 35 (project 21, owned by testuser1) has comment 15 (author 1) and
// comment 17 (author -2, a link share) — used for the link-share read case.
func TestHumaTaskComment(t *testing.T) {
	// task 1 belongs to project 1, owned by testuser1.
	onTask1 := webHandlerTestV2{
		user:     &testuser1,
		basePath: "/api/v2/tasks/1/comments",
		idParam:  "commentid",
		t:        t,
	}
	require.NoError(t, onTask1.ensureEnv())
	// task 35 also belongs to testuser1; share the Echo instance so the JWT
	// signing secret stays valid (each setupTestEnv() regenerates it).
	onTask35 := webHandlerTestV2{
		user:     &testuser1,
		basePath: "/api/v2/tasks/35/comments",
		idParam:  "commentid",
		t:        t,
		e:        onTask1.e,
	}
	// task 2 also belongs to project 1; used for the wrong-parent negative.
	onTask2 := webHandlerTestV2{
		user:     &testuser1,
		basePath: "/api/v2/tasks/2/comments",
		idParam:  "commentid",
		t:        t,
		e:        onTask1.e,
	}
	// user6 has no access to project 1, so it is neither author nor writer on
	// task 1's comment 1 — used for the no-access forbidden negatives.
	asUser6 := webHandlerTestV2{
		user:     &testuser6,
		basePath: "/api/v2/tasks/1/comments",
		idParam:  "commentid",
		t:        t,
		e:        onTask1.e,
	}
	// task 16 belongs to project 7, which testuser1 can write to via team 3.
	// Comment 4 on task 16 is authored by user 6, so testuser1 has write access
	// to the task but is *not* the comment author — this is what genuinely
	// proves the author-only update/delete restriction (as opposed to plain
	// access denial, which asUser6 covers).
	asWriterNonAuthor := webHandlerTestV2{
		user:     &testuser1,
		basePath: "/api/v2/tasks/16/comments",
		idParam:  "commentid",
		t:        t,
		e:        onTask1.e,
	}

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
			// order_by is an exposed query param; just assert it is accepted.
			rec, err := onTask35.testReadAllWithUser(url.Values{"order_by": []string{"desc"}}, nil)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `comment 15`)
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
		t.Run("Forbidden", func(t *testing.T) {
			_, err := asUser6.testCreateWithUser(nil, nil, `{"comment":"Nope"}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
	})

	t.Run("Update", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := onTask1.testUpdateWithUser(nil, map[string]string{"commentid": "1"}, `{"comment":"Edited comment"}`)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"comment":"Edited comment"`)
		})
		t.Run("Comment from another task", func(t *testing.T) {
			// comment 1 is on task 1; updating it under task 2 must 404.
			_, err := onTask2.testUpdateWithUser(nil, map[string]string{"commentid": "1"}, `{"comment":"x"}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
		})
		t.Run("Forbidden non-author with write access", func(t *testing.T) {
			// testuser1 can write to task 16 but did not author comment 4
			// (user 6 did), so the author-only restriction must still 403 — this
			// is the case that actually exercises authorship, not access.
			_, err := asWriterNonAuthor.testUpdateWithUser(nil, map[string]string{"commentid": "4"}, `{"comment":"x"}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Forbidden no access", func(t *testing.T) {
			// user6 has no access to task 1 at all.
			_, err := asUser6.testUpdateWithUser(nil, map[string]string{"commentid": "1"}, `{"comment":"x"}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
	})

	t.Run("Delete", func(t *testing.T) {
		t.Run("Forbidden non-author with write access", func(t *testing.T) {
			// testuser1 can write to task 16 but did not author comment 4,
			// so deleting another user's comment must 403.
			_, err := asWriterNonAuthor.testDeleteWithUser(nil, map[string]string{"commentid": "4"})
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Forbidden no access", func(t *testing.T) {
			_, err := asUser6.testDeleteWithUser(nil, map[string]string{"commentid": "1"})
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Comment from another task", func(t *testing.T) {
			_, err := onTask2.testDeleteWithUser(nil, map[string]string{"commentid": "1"})
			require.Error(t, err)
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
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
