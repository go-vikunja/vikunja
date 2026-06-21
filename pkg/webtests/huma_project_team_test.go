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
	"net/url"
	"testing"

	"code.vikunja.io/api/pkg/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProjectTeam ports v1's TeamProject model coverage
// (pkg/models/project_team_test.go) to the nested /api/v2 path. There is no
// read-one (v1 has none), so list is the only read path.
//
// Every mutation needs project admin (Can* -> Project.IsAdmin); list only needs
// read. Projects 9/10/11 are shared to testuser1 read/write/admin, so the same
// user walks the whole gradient by switching the parent path.
func TestProjectTeam(t *testing.T) {
	// project 1 is owned by testuser1 and starts with no team shares.
	owned := webHandlerTestV2{
		user:     &testuser1,
		basePath: "/api/v2/projects/1/teams",
		idParam:  "team",
		t:        t,
	}
	require.NoError(t, owned.ensureEnv())
	// Every other harness shares owned's Echo: setupTestEnv() rotates the global
	// JWT secret, so independent harnesses would invalidate each other's tokens.
	//
	// project 3 has team 1 shared read-only; testuser1 reads it via team 1.
	readableShared := webHandlerTestV2{user: &testuser1, basePath: "/api/v2/projects/3/teams", idParam: "team", t: t, e: owned.e}
	// project 2 (owner user3) is not shared to testuser1 at all.
	forbidden := webHandlerTestV2{user: &testuser1, basePath: "/api/v2/projects/2/teams", idParam: "team", t: t, e: owned.e}
	// read/write shares: can list, below the admin bar every mutation requires.
	readShared := webHandlerTestV2{user: &testuser1, basePath: "/api/v2/projects/9/teams", idParam: "team", t: t, e: owned.e}
	writeShared := webHandlerTestV2{user: &testuser1, basePath: "/api/v2/projects/10/teams", idParam: "team", t: t, e: owned.e}
	adminShared := webHandlerTestV2{user: &testuser1, basePath: "/api/v2/projects/11/teams", idParam: "team", t: t, e: owned.e}

	t.Run("ReadAll", func(t *testing.T) {
		// project 3 has exactly one team shared.
		t.Run("Normal", func(t *testing.T) {
			rec, err := readableShared.testReadAllWithUser(nil, nil)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"name":"testteam1"`)
			assert.Contains(t, rec.Body.String(), `"permission":`)

			var env struct {
				Items []struct {
					ID   int64  `json:"id"`
					Name string `json:"name"`
				} `json:"items"`
				Total int64 `json:"total"`
			}
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &env))
			assert.Len(t, env.Items, 1)
			assert.Equal(t, int64(1), env.Total)
			assert.Equal(t, int64(1), env.Items[0].ID)
		})
		// project 19: q "TEAM9" matches exactly team 9 of its three shares.
		t.Run("Search", func(t *testing.T) {
			h := webHandlerTestV2{user: &testuser1, basePath: "/api/v2/projects/19/teams", idParam: "team", t: t, e: owned.e}
			rec, err := h.testReadAllWithUser(url.Values{"q": []string{"TEAM9"}}, nil)
			require.NoError(t, err)
			var env struct {
				Items []struct {
					ID int64 `json:"id"`
				} `json:"items"`
			}
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &env))
			require.Len(t, env.Items, 1)
			assert.Equal(t, int64(9), env.Items[0].ID)
		})
		t.Run("Read-only share can list", func(t *testing.T) {
			_, err := readShared.testReadAllWithUser(nil, nil)
			require.NoError(t, err)
		})
		t.Run("Write share can list", func(t *testing.T) {
			_, err := writeShared.testReadAllWithUser(nil, nil)
			require.NoError(t, err)
		})
		// No read access -> ErrNeedToHaveProjectReadAccess (403), not 404.
		t.Run("Forbidden", func(t *testing.T) {
			h := webHandlerTestV2{user: &testuser1, basePath: "/api/v2/projects/5/teams", idParam: "team", t: t, e: owned.e}
			_, err := h.testReadAllWithUser(nil, nil)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Nonexisting project", func(t *testing.T) {
			h := webHandlerTestV2{user: &testuser1, basePath: "/api/v2/projects/99999/teams", idParam: "team", t: t, e: owned.e}
			_, err := h.testReadAllWithUser(nil, nil)
			require.Error(t, err)
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
		})
	})

	t.Run("Create", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := owned.testCreateWithUser(nil, nil, `{"team_id":1,"permission":2}`)
			require.NoError(t, err)
			assert.Equal(t, http.StatusCreated, rec.Code)
			assert.Contains(t, rec.Body.String(), `"team_id":1`)
			assert.Contains(t, rec.Body.String(), `"permission":2`)
			db.AssertExists(t, "team_projects", map[string]interface{}{
				"team_id":    1,
				"project_id": 1,
				"permission": 2,
			}, false)
		})
		t.Run("Admin share can create", func(t *testing.T) {
			rec, err := adminShared.testCreateWithUser(nil, nil, `{"team_id":2,"permission":0}`)
			require.NoError(t, err)
			assert.Equal(t, http.StatusCreated, rec.Code)
			db.AssertExists(t, "team_projects", map[string]interface{}{
				"team_id":    2,
				"project_id": 11,
			}, false)
		})
		// Re-shares team 1 added by Normal above -> ErrTeamAlreadyHasAccess (409).
		t.Run("Team already has access", func(t *testing.T) {
			_, err := owned.testCreateWithUser(nil, nil, `{"team_id":1,"permission":1}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusConflict, getHTTPErrorCode(err))
		})
		t.Run("Read share cannot create", func(t *testing.T) {
			_, err := readShared.testCreateWithUser(nil, nil, `{"team_id":1,"permission":0}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Write share cannot create", func(t *testing.T) {
			_, err := writeShared.testCreateWithUser(nil, nil, `{"team_id":1,"permission":0}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Forbidden", func(t *testing.T) {
			_, err := forbidden.testCreateWithUser(nil, nil, `{"team_id":1,"permission":0}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Nonexisting team", func(t *testing.T) {
			_, err := owned.testCreateWithUser(nil, nil, `{"team_id":9999,"permission":0}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
		})
		// -1 parses (maps to PermissionUnknown), then isValid rejects it -> 400.
		t.Run("Invalid permission", func(t *testing.T) {
			_, err := owned.testCreateWithUser(nil, nil, `{"team_id":8,"permission":-1}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusBadRequest, getHTTPErrorCode(err))
		})
		// Out of the -1..2 enum, so the body fails to parse -> 422, before the model.
		t.Run("Unparseable permission", func(t *testing.T) {
			_, err := owned.testCreateWithUser(nil, nil, `{"team_id":8,"permission":500}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusUnprocessableEntity, getHTTPErrorCode(err))
		})
	})

	t.Run("Update", func(t *testing.T) {
		// Share team 3 first so the update has a row to flip.
		t.Run("Normal", func(t *testing.T) {
			_, err := owned.testCreateWithUser(nil, nil, `{"team_id":3,"permission":0}`)
			require.NoError(t, err)
			rec, err := owned.testUpdateWithUser(nil, map[string]string{"team": "3"}, `{"permission":2}`)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"permission":2`)
			db.AssertExists(t, "team_projects", map[string]interface{}{
				"team_id":    3,
				"project_id": 1,
				"permission": 2,
			}, false)
		})
		t.Run("Read share cannot update", func(t *testing.T) {
			_, err := readShared.testUpdateWithUser(nil, map[string]string{"team": "1"}, `{"permission":2}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Write share cannot update", func(t *testing.T) {
			_, err := writeShared.testUpdateWithUser(nil, map[string]string{"team": "1"}, `{"permission":2}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Forbidden", func(t *testing.T) {
			_, err := forbidden.testUpdateWithUser(nil, map[string]string{"team": "1"}, `{"permission":2}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		// -1 parses, then isValid rejects it -> ErrInvalidPermission (400).
		t.Run("Invalid permission", func(t *testing.T) {
			_, err := owned.testUpdateWithUser(nil, map[string]string{"team": "3"}, `{"permission":-1}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusBadRequest, getHTTPErrorCode(err))
		})
		t.Run("Unparseable permission", func(t *testing.T) {
			_, err := owned.testUpdateWithUser(nil, map[string]string{"team": "3"}, `{"permission":500}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusUnprocessableEntity, getHTTPErrorCode(err))
		})
	})

	t.Run("Delete", func(t *testing.T) {
		// Share team 4 first so there's a row to remove.
		t.Run("Normal", func(t *testing.T) {
			_, err := owned.testCreateWithUser(nil, nil, `{"team_id":4,"permission":0}`)
			require.NoError(t, err)
			rec, err := owned.testDeleteWithUser(nil, map[string]string{"team": "4"})
			require.NoError(t, err)
			assert.Equal(t, http.StatusNoContent, rec.Code)
			assert.Empty(t, rec.Body.String())
			db.AssertMissing(t, "team_projects", map[string]interface{}{
				"team_id":    4,
				"project_id": 1,
			})
		})
		t.Run("Read share cannot delete", func(t *testing.T) {
			_, err := readShared.testDeleteWithUser(nil, map[string]string{"team": "1"})
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Write share cannot delete", func(t *testing.T) {
			_, err := writeShared.testDeleteWithUser(nil, map[string]string{"team": "1"})
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Forbidden", func(t *testing.T) {
			_, err := forbidden.testDeleteWithUser(nil, map[string]string{"team": "1"})
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Nonexisting team", func(t *testing.T) {
			_, err := owned.testDeleteWithUser(nil, map[string]string{"team": "9999"})
			require.Error(t, err)
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
		})
		// Team 8 exists but was never shared -> ErrTeamDoesNotHaveAccessToProject (403).
		t.Run("Team not on project", func(t *testing.T) {
			_, err := owned.testDeleteWithUser(nil, map[string]string{"team": "8"})
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
	})
}
