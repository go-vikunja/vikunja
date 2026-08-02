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
	"fmt"
	"net/http"
	"strconv"
	"testing"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The Favorites pseudo-project and the project backing saved filter 1, both of
// which testuser1 sees in a project listing.
const (
	favoritesProjectID   = models.FavoritesPseudoProjectID
	savedFilterProjectID = -2
)

// assertParentProjectIDIsANumber fails if any project in raw omits
// parent_project_id or sends it as null. Clients parse it as a plain int, so a
// missing key breaks the whole response (go-vikunja/app#295).
func assertParentProjectIDIsANumber(t *testing.T, raw json.RawMessage) map[int64]int64 {
	var projects []map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(raw, &projects))
	require.NotEmpty(t, projects)

	parents := make(map[int64]int64, len(projects))
	for _, p := range projects {
		id, err := strconv.ParseInt(string(p["id"]), 10, 64)
		require.NoError(t, err)

		parent, ok := p["parent_project_id"]
		require.True(t, ok, "project %d is missing parent_project_id", id)
		parentID, err := strconv.ParseInt(string(parent), 10, 64)
		require.NoError(t, err, "project %d has a non-numeric parent_project_id %s", id, parent)
		parents[id] = parentID
	}
	return parents
}

// TestProjectPseudoParentProjectIDV1 guards the raw JSON of v1's project routes:
// the Favorites and saved-filter pseudo-projects are struct literals that never
// set ParentProjectID, so they used to drop the key entirely.
func TestProjectPseudoParentProjectIDV1(t *testing.T) {
	testHandler := webHandlerTest{
		user:    &testuser1,
		strFunc: func() handler.CObject { return &models.Project{} },
		t:       t,
	}

	t.Run("ReadAll", func(t *testing.T) {
		rec, err := testHandler.testReadAllWithUser(nil, nil)
		require.NoError(t, err)

		parents := assertParentProjectIDIsANumber(t, rec.Body.Bytes())
		require.Contains(t, parents, int64(favoritesProjectID))
		require.Contains(t, parents, int64(savedFilterProjectID))
		assert.Equal(t, int64(0), parents[favoritesProjectID])
		assert.Equal(t, int64(0), parents[savedFilterProjectID])
	})

	for _, id := range []int64{favoritesProjectID, savedFilterProjectID} {
		t.Run(fmt.Sprintf("ReadOne %d", id), func(t *testing.T) {
			rec, err := testHandler.testReadOneWithUser(nil, map[string]string{"project": strconv.FormatInt(id, 10)})
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"parent_project_id":0`)
		})
	}
}

// TestProjectPseudoParentProjectIDV2 is the same guard for the Huma-backed v2
// routes, which serve the same model through a paginated envelope.
func TestProjectPseudoParentProjectIDV2(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	token := humaTokenFor(t, &testuser1)

	rec := humaRequest(t, e, http.MethodGet, "/api/v2/projects?per_page=100", "", token, "")
	require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())

	var paginated struct {
		Items json.RawMessage `json:"items"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &paginated))

	parents := assertParentProjectIDIsANumber(t, paginated.Items)
	require.Contains(t, parents, int64(favoritesProjectID))
	require.Contains(t, parents, int64(savedFilterProjectID))
	assert.Equal(t, int64(0), parents[favoritesProjectID])
	assert.Equal(t, int64(0), parents[savedFilterProjectID])

	for _, id := range []int64{favoritesProjectID, savedFilterProjectID} {
		t.Run(fmt.Sprintf("ReadOne %d", id), func(t *testing.T) {
			rec := humaRequest(t, e, http.MethodGet, fmt.Sprintf("/api/v2/projects/%d", id), "", token, "")
			require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
			assert.Contains(t, rec.Body.String(), `"parent_project_id":0`)
		})
	}
}

// TestProjectOmittedParentProjectIDOverHTTP is the request-side counterpart:
// dropping the omitempty tag must not turn parent_project_id into a field that
// defaults to 0 on write, or a partial update would detach a project from its
// parent — the escalation GHSA-44v6-7fxq-vgf4 gates behind Admin.
func TestProjectOmittedParentProjectIDOverHTTP(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	// testuser1 has Write (not Admin) on project 43 through its parent 10.
	token := humaTokenFor(t, &testuser1)

	t.Run("omitted parent keeps the parent", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodPut, "/api/v2/projects/43",
			`{"title":"renamed without a parent"}`, token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())

		rec = humaRequest(t, e, http.MethodGet, "/api/v2/projects/43", "", token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"parent_project_id":10`)
	})

	t.Run("explicit 0 without Admin is still forbidden", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodPut, "/api/v2/projects/43",
			`{"title":"detached","parent_project_id":0}`, token, "")
		assert.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
	})
}
