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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHumaAPITokenAutoPatch guards #3528: AutoPatch's internal GET leg made every
// PATCH additionally require read_one, and pins both legs to the PATCH's own route.
func TestHumaAPITokenAutoPatch(t *testing.T) {
	// Token 1 (pkg/db/fixtures/api_tokens.yml), owned by user1, scoped to
	// tasks: read_all + update — deliberately without read_one.
	const updateToken = "tk_2eef46f40ebab3304919ab2e7e39993f75f29d2e"
	// Token 9, also owned by user1, scoped to tasks: read_all only — so a
	// missing update scope is the only thing that can reject a patch with it.
	const readOnlyToken = "tk_readonly_tasks_user1_00000000abcd1234"

	t.Run("patch works with update but no read_one", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		rec := humaRequest(t, e, http.MethodPatch, "/api/v2/tasks/1", `{"done":true}`, updateToken, "application/merge-patch+json")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"done":true`)
		assert.Contains(t, rec.Body.String(), `"title":"task #1"`,
			"the internal GET leg must have loaded the existing task so the merge keeps untouched fields")
	})

	t.Run("read_one stays enforced for client GETs", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		rec := humaRequest(t, e, http.MethodGet, "/api/v2/tasks/1", "", updateToken, "")
		assert.Equal(t, http.StatusUnauthorized, rec.Code, "body: %s", rec.Body.String())
	})

	t.Run("patch still rejected without update", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		rec := humaRequest(t, e, http.MethodPatch, "/api/v2/tasks/1", `{"done":true}`, readOnlyToken, "application/merge-patch+json")
		assert.Equal(t, http.StatusUnauthorized, rec.Code, "body: %s", rec.Body.String())
	})

	// Echo routes on the raw path while autopatch re-dispatches the decoded one,
	// so an encoded slash can land the legs on a deeper route than the PATCH was
	// authorised against.
	t.Run("encoded slash must not pivot into a route the token lacks", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		rec := humaRequest(t, e, http.MethodPatch, "/api/v2/tasks/1%2Fcomments%2F1", `{"comment":"PWNED"}`, updateToken, "application/merge-patch+json")
		assert.Equal(t, http.StatusUnauthorized, rec.Code, "body: %s", rec.Body.String())

		rec = humaRequest(t, e, http.MethodGet, "/api/v2/tasks/1/comments/1", "", humaTokenFor(t, &testuser1), "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), "Lorem Ipsum Dolor Sit Amet", "the comment must be untouched")
	})

	// An encoded ? smuggles a query string onto the internal legs, which would
	// otherwise run with the scope check skipped.
	t.Run("encoded query must not widen what the legs may read", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		rec := humaRequest(t, e, http.MethodPatch, "/api/v2/tasks/1%3Fexpand=comments", `{"title":"qs"}`, updateToken, "application/merge-patch+json")
		assert.Equal(t, http.StatusUnauthorized, rec.Code, "body: %s", rec.Body.String())
		assert.NotContains(t, rec.Body.String(), "Lorem Ipsum Dolor Sit Amet",
			"the smuggled expand must not leak comments the token cannot read")
	})
}
