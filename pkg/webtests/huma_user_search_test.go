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

// TestHumaUserSearch covers the global user search. Emails must never leak.
func TestHumaUserSearch(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	token := humaTokenFor(t, &testuser1)

	t.Run("Search by username", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/users?q=user2", "", token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())

		usernames, emails := usersFromSearch(t, rec.Body.Bytes())
		assert.Contains(t, usernames, "user2")
		for _, em := range emails {
			assert.Empty(t, em, "user search must never return email addresses")
		}
	})
	t.Run("Unauthenticated", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/users?q=user2", "", "", "")
		assert.Equal(t, http.StatusUnauthorized, rec.Code, "body: %s", rec.Body.String())
	})
}

// TestHumaProjectUserSearch covers the per-project user search used for share
// autocomplete. It requires read access to the project.
func TestHumaProjectUserSearch(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	token := humaTokenFor(t, &testuser1)

	t.Run("Owned project", func(t *testing.T) {
		// testuser1 owns project 1.
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/projects/1/users/search", "", token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"items"`)
	})
	t.Run("Forbidden - no access", func(t *testing.T) {
		// project 2 is owned by user3; testuser1 has no access.
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/projects/2/users/search", "", token, "")
		assert.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
	})
	t.Run("Nonexistent project", func(t *testing.T) {
		// CanRead surfaces ErrProjectDoesNotExist (404), not a bare forbidden.
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/projects/99999/users/search", "", token, "")
		assert.Equal(t, http.StatusNotFound, rec.Code, "body: %s", rec.Body.String())
	})
}

func usersFromSearch(t *testing.T, body []byte) (usernames, emails []string) {
	t.Helper()
	var resp struct {
		Items []struct {
			Username string `json:"username"`
			Email    string `json:"email"`
		} `json:"items"`
	}
	require.NoError(t, json.Unmarshal(body, &resp), "search body must be a paginated envelope: %s", string(body))
	for _, it := range resp.Items {
		usernames = append(usernames, it.Username)
		emails = append(emails, it.Email)
	}
	return usernames, emails
}
