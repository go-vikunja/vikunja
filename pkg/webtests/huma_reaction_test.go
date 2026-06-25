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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// reactionMapFromBody decodes the v2 reactions list body — a map keyed by
// reaction value, each value the list of users who reacted with it.
func reactionMapFromBody(t *testing.T, body []byte) map[string][]struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
} {
	t.Helper()
	var m map[string][]struct {
		ID       int64  `json:"id"`
		Username string `json:"username"`
	}
	require.NoError(t, json.Unmarshal(body, &m), "list body must be a reaction map: %s", string(body))
	return m
}

// TestHumaReaction exercises the v2 reaction surface, mirroring the v1
// model-level matrix in pkg/models/reaction_test.go. Fixture reactions.yml
// seeds reaction #1: user1 reacted "👋" on task #1.
func TestHumaReaction(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	token := humaTokenFor(t, &testuser1)

	t.Run("List returns the map with the reacting user", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/tasks/1/reactions", "", token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		m := reactionMapFromBody(t, rec.Body.Bytes())
		require.Len(t, m["👋"], 1, "fixture reaction must be present; body: %s", rec.Body.String())
		assert.Equal(t, int64(1), m["👋"][0].ID, "the reacting user is user1")
	})

	t.Run("Create then list reflects the new reaction", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodPost, "/api/v2/tasks/1/reactions", `{"value":"🦙"}`, token, "")
		require.Equal(t, http.StatusCreated, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"value":"🦙"`)

		rec = humaRequest(t, e, http.MethodGet, "/api/v2/tasks/1/reactions", "", token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		m := reactionMapFromBody(t, rec.Body.Bytes())
		require.Len(t, m["🦙"], 1, "created reaction must appear in the list; body: %s", rec.Body.String())
		assert.Equal(t, int64(1), m["🦙"][0].ID)
	})

	t.Run("Delete removes the reaction", func(t *testing.T) {
		// Remove the fixture reaction (user1's "👋" on task #1) and confirm via a follow-up list.
		rec := humaRequest(t, e, http.MethodPost, "/api/v2/tasks/1/reactions/delete", `{"value":"👋"}`, token, "")
		require.Equal(t, http.StatusOK, rec.Code, "delete is POST-with-body returning 200; body: %s", rec.Body.String())

		rec = humaRequest(t, e, http.MethodGet, "/api/v2/tasks/1/reactions", "", token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		m := reactionMapFromBody(t, rec.Body.Bytes())
		assert.NotContains(t, m, "👋", "deleted reaction must be gone; body: %s", rec.Body.String())
	})

	t.Run("Invalid entitykind is rejected", func(t *testing.T) {
		// The enum tag on the path param makes Huma reject unknown kinds before the handler runs.
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/loremipsum/1/reactions", "", token, "")
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code, "body: %s", rec.Body.String())
	})

	t.Run("Forbidden - no access to the entity", func(t *testing.T) {
		// Task #34 lives in a private project user1 cannot see.
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/tasks/34/reactions", "", token, "")
		assert.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
	})

	t.Run("Nonexistent entity", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/tasks/9999999/reactions", "", token, "")
		assert.Equal(t, http.StatusNotFound, rec.Code, "body: %s", rec.Body.String())
	})

	t.Run("Create forbidden - no access to the entity", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodPost, "/api/v2/tasks/34/reactions", `{"value":"🦙"}`, token, "")
		assert.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
	})
}
