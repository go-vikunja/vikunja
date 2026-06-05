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

const (
	sessionUser1A = "550e8400-e29b-41d4-a716-446655440001"
	sessionUser1B = "550e8400-e29b-41d4-a716-446655440002"
	sessionUser2  = "550e8400-e29b-41d4-a716-446655440003"
)

// TestHumaSession mirrors v1's TestSessions session-CRUD matrix (list own vs
// others', delete own, non-owner forbidden) so v2 parity is readable
// side-by-side. The login/refresh auth-flow cases in TestSessions are not a
// session CRUD surface and stay on v1.
func TestHumaSession(t *testing.T) {
	testHandler := webHandlerTestV2{
		user:     &testuser1,
		basePath: "/api/v2/user/sessions",
		idParam:  "session",
		t:        t,
	}

	t.Run("ReadAll", func(t *testing.T) {
		t.Run("Normal - exact visible set for user1", func(t *testing.T) {
			rec, err := testHandler.testReadAllWithUser(nil, nil)
			require.NoError(t, err)

			ids := sessionIDsFromReadAll(t, rec.Body.Bytes())
			// User1 owns exactly sessions A and B; user2's session must never appear.
			assert.ElementsMatch(t, []string{sessionUser1A, sessionUser1B}, ids,
				"ReadAll must return exactly user1's two sessions; body: %s", rec.Body.String())
			assert.NotContains(t, ids, sessionUser2, "user2's session must be hidden")
		})
	})

	t.Run("Delete", func(t *testing.T) {
		t.Run("Normal - own session", func(t *testing.T) {
			rec, err := testHandler.testDeleteWithUser(nil, map[string]string{"session": sessionUser1B})
			require.NoError(t, err)
			// v2 delete is 204 No Content; v1 returned 200 + a message body.
			assert.Equal(t, http.StatusNoContent, rec.Code)
			assert.Empty(t, rec.Body.String())
		})
		t.Run("Nonexisting", func(t *testing.T) {
			_, err := testHandler.testDeleteWithUser(nil, map[string]string{"session": "00000000-0000-0000-0000-000000000000"})
			require.Error(t, err)
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
		})
		t.Run("Forbidden - other user's session", func(t *testing.T) {
			_, err := testHandler.testDeleteWithUser(nil, map[string]string{"session": sessionUser2})
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
	})
}

// sessionIDsFromReadAll extracts the session UUIDs from a v2 paginated list
// body so the visible set can be asserted exactly.
func sessionIDsFromReadAll(t *testing.T, body []byte) []string {
	t.Helper()
	var resp struct {
		Items []struct {
			ID string `json:"id"`
		} `json:"items"`
	}
	require.NoError(t, json.Unmarshal(body, &resp), "ReadAll body must be a paginated envelope: %s", string(body))
	ids := make([]string, 0, len(resp.Items))
	for _, it := range resp.Items {
		ids = append(ids, it.ID)
	}
	return ids
}
