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

// AutoPatch resubmits the GET response through PUT, so the subscription object
// embedded in task reads must validate against the write schema. With
// SubscriptionEntityType reflected as an integer this failed with
// "subscription.entity: expected integer" for subscribed users (issue #3316).
func TestHumaTaskPatchWithSubscription(t *testing.T) {
	patch := `[{"op":"replace","path":"/title","value":"patched"}]`

	t.Run("direct task subscription", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		// user1 is subscribed to task 2 via fixtures.
		rec := humaRequest(t, e, http.MethodPatch, "/api/v2/tasks/2", patch, humaTokenFor(t, &testuser1), "application/json-patch+json")
		assert.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"title":"patched"`)
		// The subscription block must be present, otherwise the AutoPatch
		// validation path this test guards was never exercised.
		assert.Contains(t, rec.Body.String(), `"entity":"task"`)
	})

	t.Run("inherited project subscription", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, &testuser1)
		rec := humaRequest(t, e, http.MethodPost, "/api/v2/subscriptions/project/1", "", token, "")
		require.Equal(t, http.StatusCreated, rec.Code, "body: %s", rec.Body.String())

		rec = humaRequest(t, e, http.MethodPatch, "/api/v2/tasks/1", patch, token, "application/json-patch+json")
		assert.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"title":"patched"`)
		// See above — inherited subscriptions surface with the parent's entity kind.
		assert.Contains(t, rec.Body.String(), `"entity":"project"`)
	})
}
