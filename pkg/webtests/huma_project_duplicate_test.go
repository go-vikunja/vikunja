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
	"testing"

	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProjectDuplicateV2 covers POST /projects/{projectid}/duplicate. It drives
// the Echo+Huma stack directly (humaRequest/humaTokenFor) because
// webHandlerTestV2's buildURL only models base[/{id}] paths, not action sub-paths.
func TestProjectDuplicateV2(t *testing.T) {
	t.Run("duplicates an accessible project to the top level", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		// Duplicating copies the source project's task attachments, so the
		// referenced fixture file must exist in the (memory) file store.
		files.InitTestFileFixtures(t)
		token := humaTokenFor(t, &testuser1)

		// Project 1 is owned by testuser1.
		const sourceProjectID int64 = 1
		rec := humaRequest(t, e, http.MethodPost, "/api/v2/projects/1/duplicate", `{}`, token, "")
		require.Equal(t, http.StatusCreated, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"duplicated_project"`)

		var resp struct {
			DuplicatedProject struct {
				ID              int64  `json:"id"`
				Title           string `json:"title"`
				ParentProjectID int64  `json:"parent_project_id"`
			} `json:"duplicated_project"`
		}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.NotZero(t, resp.DuplicatedProject.ID, "duplicated project should have an id")
		assert.NotEqual(t, sourceProjectID, resp.DuplicatedProject.ID, "duplicated project must have a new id, not the source project's")
		assert.Contains(t, resp.DuplicatedProject.Title, "duplicate")
		assert.Zero(t, resp.DuplicatedProject.ParentProjectID, "top-level duplicate must have no parent")
	})

	t.Run("places the duplicate under parent_project_id from the body", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		files.InitTestFileFixtures(t)
		token := humaTokenFor(t, &testuser1)

		// testuser1 owns project 1, so it may both read the source and create
		// the copy underneath it.
		rec := humaRequest(t, e, http.MethodPost, "/api/v2/projects/1/duplicate", `{"parent_project_id":1}`, token, "")
		require.Equal(t, http.StatusCreated, rec.Code, "body: %s", rec.Body.String())

		var resp struct {
			DuplicatedProject struct {
				ID              int64 `json:"id"`
				ParentProjectID int64 `json:"parent_project_id"`
			} `json:"duplicated_project"`
		}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.NotZero(t, resp.DuplicatedProject.ID)
		assert.Equal(t, int64(1), resp.DuplicatedProject.ParentProjectID, "duplicate must land under the requested parent")
	})

	t.Run("nonexistent source project", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, &testuser1)

		rec := humaRequest(t, e, http.MethodPost, "/api/v2/projects/99999/duplicate", `{}`, token, "")
		// CanCreate loads the source via CanRead, which surfaces
		// ErrProjectDoesNotExist (404) for a missing project rather than a 403.
		require.Equal(t, http.StatusNotFound, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"code":%d`, models.ErrCodeProjectDoesNotExist), "body must surface ErrCodeProjectDoesNotExist; body: %s", rec.Body.String())
	})

	t.Run("no read on source project is forbidden", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		// testuser15 cannot read project 1 (owned by testuser1, no share).
		token := humaTokenFor(t, &testuser15)

		rec := humaRequest(t, e, http.MethodPost, "/api/v2/projects/1/duplicate", `{}`, token, "")
		require.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
	})
}
