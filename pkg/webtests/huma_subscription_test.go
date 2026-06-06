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

// TestHumaSubscription ports the model-level matrix in
// pkg/models/subscription_test.go to the v2 HTTP surface: subscribing requires
// access to the target entity, an invalid entity kind is rejected, and
// inaccessible entities are forbidden. Subscriptions has no v1 webtest, so this
// proves the contract independently.
//
// Fixture facts the matrix relies on (see pkg/db/fixtures):
//   - user1 has read access to task 1 and project 1.
//   - user1 is already subscribed to task 2 (subscriptions.yml id 1).
//   - user1 cannot see task 14 or project 20.
func TestHumaSubscription(t *testing.T) {
	token := func(t *testing.T) string { return humaTokenFor(t, &testuser1) }

	t.Run("Create", func(t *testing.T) {
		t.Run("task - normal", func(t *testing.T) {
			e, err := setupTestEnv()
			require.NoError(t, err)
			rec := humaRequest(t, e, http.MethodPost, "/api/v2/subscriptions/task/1", "", token(t), "")
			assert.Equal(t, http.StatusCreated, rec.Code, "body: %s", rec.Body.String())
			assert.Contains(t, rec.Body.String(), `"entity":"task"`)
			assert.Contains(t, rec.Body.String(), `"entity_id":1`)
		})
		t.Run("project - normal", func(t *testing.T) {
			e, err := setupTestEnv()
			require.NoError(t, err)
			rec := humaRequest(t, e, http.MethodPost, "/api/v2/subscriptions/project/1", "", token(t), "")
			assert.Equal(t, http.StatusCreated, rec.Code, "body: %s", rec.Body.String())
			assert.Contains(t, rec.Body.String(), `"entity":"project"`)
		})
		t.Run("already exists", func(t *testing.T) {
			e, err := setupTestEnv()
			require.NoError(t, err)
			// user1 is already subscribed to task 2.
			rec := humaRequest(t, e, http.MethodPost, "/api/v2/subscriptions/task/2", "", token(t), "")
			assert.Equal(t, http.StatusPreconditionFailed, rec.Code, "body: %s", rec.Body.String())
		})
		t.Run("invalid entity kind", func(t *testing.T) {
			e, err := setupTestEnv()
			require.NoError(t, err)
			// The enum on the path param makes Huma reject unknown kinds before the handler.
			rec := humaRequest(t, e, http.MethodPost, "/api/v2/subscriptions/bogus/1", "", token(t), "")
			assert.Equal(t, http.StatusUnprocessableEntity, rec.Code, "body: %s", rec.Body.String())
		})
		t.Run("nonexisting task", func(t *testing.T) {
			e, err := setupTestEnv()
			require.NoError(t, err)
			rec := humaRequest(t, e, http.MethodPost, "/api/v2/subscriptions/task/9999999", "", token(t), "")
			assert.Equal(t, http.StatusNotFound, rec.Code, "body: %s", rec.Body.String())
		})
		t.Run("nonexisting project", func(t *testing.T) {
			e, err := setupTestEnv()
			require.NoError(t, err)
			rec := humaRequest(t, e, http.MethodPost, "/api/v2/subscriptions/project/9999999", "", token(t), "")
			assert.Equal(t, http.StatusNotFound, rec.Code, "body: %s", rec.Body.String())
		})
		t.Run("forbidden - no access to task", func(t *testing.T) {
			e, err := setupTestEnv()
			require.NoError(t, err)
			// task 14 is not accessible to user1.
			rec := humaRequest(t, e, http.MethodPost, "/api/v2/subscriptions/task/14", "", token(t), "")
			assert.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
		})
		t.Run("forbidden - no access to project", func(t *testing.T) {
			e, err := setupTestEnv()
			require.NoError(t, err)
			// project 20 is not accessible to user1.
			rec := humaRequest(t, e, http.MethodPost, "/api/v2/subscriptions/project/20", "", token(t), "")
			assert.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
		})
	})

	t.Run("Delete", func(t *testing.T) {
		t.Run("normal", func(t *testing.T) {
			e, err := setupTestEnv()
			require.NoError(t, err)
			// user1 is subscribed to task 2.
			rec := humaRequest(t, e, http.MethodDelete, "/api/v2/subscriptions/task/2", "", token(t), "")
			assert.Equal(t, http.StatusNoContent, rec.Code, "body: %s", rec.Body.String())
			assert.Empty(t, rec.Body.String())
		})
		t.Run("not subscribed - forbidden", func(t *testing.T) {
			e, err := setupTestEnv()
			require.NoError(t, err)
			// CanDelete returns false when no subscription exists, so the generic
			// handler refuses with 403 (mirrors v1's DeleteWeb).
			rec := humaRequest(t, e, http.MethodDelete, "/api/v2/subscriptions/task/1", "", token(t), "")
			assert.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
		})
		t.Run("invalid entity kind", func(t *testing.T) {
			e, err := setupTestEnv()
			require.NoError(t, err)
			rec := humaRequest(t, e, http.MethodDelete, "/api/v2/subscriptions/bogus/2", "", token(t), "")
			assert.Equal(t, http.StatusUnprocessableEntity, rec.Code, "body: %s", rec.Body.String())
		})
	})
}
