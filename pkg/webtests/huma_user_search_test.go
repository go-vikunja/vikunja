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

	"code.vikunja.io/api/pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHumaUserSearch covers the global user search. Emails must not leak
// unless service.enableglobalusersearch is on.
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

type userSearchResponse struct {
	Items []struct {
		ID       int64  `json:"id"`
		Username string `json:"username"`
		Email    string `json:"email"`
	} `json:"items"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	TotalPages int64 `json:"total_pages"`
}

func TestHumaUserSearch_GlobalFlag(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	token := humaTokenFor(t, &testuser1)

	search := func(t *testing.T, q string) userSearchResponse {
		t.Helper()
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/users?q="+q, "", token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		var body userSearchResponse
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		return body
	}

	setFlag := func(t *testing.T, enabled bool) {
		prev := config.ServiceEnableGlobalUserSearch.GetBool()
		config.ServiceEnableGlobalUserSearch.Set(enabled)
		t.Cleanup(func() { config.ServiceEnableGlobalUserSearch.Set(prev) })
	}

	t.Run("flag off", func(t *testing.T) {
		setFlag(t, false)

		body := search(t, "one%20else")
		assert.Empty(t, body.Items)
		assert.Zero(t, body.Total)

		body = search(t, "user7")
		require.Len(t, body.Items, 1)
		assert.Empty(t, body.Items[0].Email)
		assert.EqualValues(t, 1, body.Total)
		assert.Equal(t, 25, body.PerPage)
	})
	t.Run("flag on", func(t *testing.T) {
		setFlag(t, true)

		body := search(t, "one%20else")
		require.Len(t, body.Items, 1)
		assert.Equal(t, "user11", body.Items[0].Username)
		assert.Equal(t, "user11@example.com", body.Items[0].Email)
		assert.EqualValues(t, 1, body.Total)
		assert.Equal(t, 1, body.Page)
		assert.Equal(t, 25, body.PerPage)
		assert.EqualValues(t, 1, body.TotalPages)

		body = search(t, "u")
		assert.Empty(t, body.Items)
		assert.Zero(t, body.Total)
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
