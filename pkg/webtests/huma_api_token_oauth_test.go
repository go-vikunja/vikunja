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
	"slices"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHumaAPITokenListKind(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	user1Token := humaTokenFor(t, &testuser1)

	rec := humaRequest(t, e, http.MethodPost, "/api/v2/tokens",
		`{"title":"made by hand","permissions":{"mcp":["access"]},"expires_at":"2099-01-01T00:00:00Z"}`,
		user1Token, "")
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	var manual struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &manual))
	first := storedToken(t, insertOAuthToken(t, mcpResourceURL, projectsReadPermissions())).ID
	second := storedToken(t, insertOAuthToken(t, mcpResourceURL, projectsReadPermissions())).ID
	fixtures := []int64{
		1,
		2,
		9,
		10,
		11,
		12,
	}

	for _, tc := range []struct {
		query string
		want  []int64
	}{
		{
			query: "kind=manual",
			want:  slices.Concat(fixtures, []int64{manual.ID}),
		},
		{
			query: "kind=oauth",
			want: []int64{
				first,
				second,
			},
		},
		{
			query: "",
			want: slices.Concat(fixtures, []int64{
				manual.ID,
				first,
				second,
			}),
		},
	} {
		t.Run(tc.query, func(t *testing.T) {
			rec := humaRequest(t, e, http.MethodGet, "/api/v2/tokens?per_page=100&"+tc.query, "", user1Token, "")
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			assert.Equal(t, tc.want, apiTokenIDsFromReadAll(t, rec.Body.Bytes()))
			assert.Equal(t, int64(len(tc.want)), apiTokenTotalFromReadAll(t, rec.Body.Bytes()))
		})
	}

	t.Run("paginates after filtering", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/tokens?kind=oauth&per_page=1&page=2", "", user1Token, "")
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		assert.Equal(t, []int64{second}, apiTokenIDsFromReadAll(t, rec.Body.Bytes()))
		assert.Equal(t, int64(2), apiTokenTotalFromReadAll(t, rec.Body.Bytes()))
	})
	t.Run("other users see none of them", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/tokens?kind=oauth", "", humaTokenFor(t, &testuser2), "")
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		assert.Empty(t, apiTokenIDsFromReadAll(t, rec.Body.Bytes()))
	})
	t.Run("unknown kind", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/tokens?kind=bogus", "", user1Token, "")
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code, rec.Body.String())
	})
}

func TestHumaAPITokenRevokeOAuthGrant(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	cleartext := insertOAuthToken(t, mcpResourceURL, projectsReadPermissions())
	path := "/api/v2/tokens/" + strconv.FormatInt(storedToken(t, cleartext).ID, 10)

	rec := humaRequest(t, e, http.MethodDelete, path, "", humaTokenFor(t, &testuser2), "")
	assert.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())
	assert.Equal(t, http.StatusOK, initializeMCP(e, cleartext).Code, "a denied revoke must leave the grant working")

	rec = humaRequest(t, e, http.MethodDelete, path, "", humaTokenFor(t, &testuser1), "")
	require.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())
	assert.Equal(t, http.StatusUnauthorized, initializeMCP(e, cleartext).Code)
}
