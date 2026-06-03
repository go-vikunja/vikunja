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
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHumaTeam mirrors v1's model TestTeam shape so v2 contract parity is
// readable side-by-side. Named TestHumaTeam to avoid clashing with the v1
// model test (pkg/models/teams_test.go TestTeam).
//
// Fixture facts (pkg/db/fixtures/team_members.yml): testuser1 is an admin of
// team 1 and a non-admin member of teams 2-8. Team 9 (created by user 7) has
// only user 2 as a member, so user1 is not a member at all.
func TestHumaTeam(t *testing.T) {
	testHandler := webHandlerTestV2{
		user:     &testuser1,
		basePath: "/api/v2/teams",
		idParam:  "team",
		t:        t,
	}

	t.Run("ReadAll", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := testHandler.testReadAllWithUser(nil, nil)
			require.NoError(t, err)
			// User 1 is a member of teams 1-8.
			assert.Contains(t, rec.Body.String(), `testteam1`)
			// User 1 is not a member of team 9 (only user 2 is).
			assert.NotContains(t, rec.Body.String(), `testteam9`)
		})
	})

	t.Run("ReadOne", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := testHandler.testReadOneWithUser(nil, map[string]string{"team": "1"})
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"name":"testteam1"`)
			assert.NotEmpty(t, rec.Result().Header.Get("ETag"))
		})
		t.Run("Nonexisting", func(t *testing.T) {
			// CanRead refuses non-members before existence is checked, so a
			// missing team returns 403, not 404.
			_, err := testHandler.testReadOneWithUser(nil, map[string]string{"team": "9999"})
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Permissions check", func(t *testing.T) {
			t.Run("Forbidden non-member", func(t *testing.T) {
				// Team 9: user1 is not a member.
				_, err := testHandler.testReadOneWithUser(nil, map[string]string{"team": "9"})
				require.Error(t, err)
				assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
			})
		})
	})

	t.Run("Create", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := testHandler.testCreateWithUser(nil, nil, `{"name":"Lorem","description":"Ipsum"}`)
			require.NoError(t, err)
			assert.Equal(t, http.StatusCreated, rec.Code)
			assert.Contains(t, rec.Body.String(), `"name":"Lorem"`)
			assert.Contains(t, rec.Body.String(), `"description":"Ipsum"`)
		})
		t.Run("Empty name", func(t *testing.T) {
			// Name has minLength:1, so Huma rejects an empty name with 422
			// before the model is touched.
			_, err := testHandler.testCreateWithUser(nil, nil, `{"name":""}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusUnprocessableEntity, getHTTPErrorCode(err))
		})
	})

	t.Run("Update", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			// Team 1: user1 is admin.
			rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"team": "1"}, `{"name":"TestLoremIpsum"}`)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"name":"TestLoremIpsum"`)
		})
		t.Run("Nonexisting", func(t *testing.T) {
			// CanUpdate -> IsAdmin -> GetTeamByID surfaces ErrTeamDoesNotExist (404).
			_, err := testHandler.testUpdateWithUser(nil, map[string]string{"team": "9999"}, `{"name":"TestLoremIpsum"}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
		})
		t.Run("Permissions check", func(t *testing.T) {
			t.Run("Forbidden non-admin", func(t *testing.T) {
				// Team 2: user1 is a member but not an admin.
				_, err := testHandler.testUpdateWithUser(nil, map[string]string{"team": "2"}, `{"name":"TestLoremIpsum"}`)
				require.Error(t, err)
				assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
			})
		})
	})

	t.Run("Delete", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			// Team 1: user1 is admin.
			rec, err := testHandler.testDeleteWithUser(nil, map[string]string{"team": "1"})
			require.NoError(t, err)
			// v2 delete is 204 No Content; v1 returned 200 + a message body.
			assert.Equal(t, http.StatusNoContent, rec.Code)
			assert.Empty(t, rec.Body.String())
		})
		t.Run("Nonexisting", func(t *testing.T) {
			_, err := testHandler.testDeleteWithUser(nil, map[string]string{"team": "9999"})
			require.Error(t, err)
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
		})
		t.Run("Permissions check", func(t *testing.T) {
			t.Run("Forbidden non-admin", func(t *testing.T) {
				// Team 2: user1 is a member but not an admin.
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{"team": "2"})
				require.Error(t, err)
				assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
			})
		})
	})
}

// TestHumaTeam_ETagReturns304 covers the v2-only conditional-request behaviour
// (ETag + If-None-Match -> 304) with no v1 counterpart.
func TestHumaTeam_ETagReturns304(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	token := humaTokenFor(t, &testuser1)

	rec := humaRequest(t, e, http.MethodGet, "/api/v2/teams/1", "", token, "")
	require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
	etag := rec.Header().Get("ETag")
	require.NotEmpty(t, etag, "GET must return an ETag header")

	req := httptest.NewRequest(http.MethodGet, "/api/v2/teams/1", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("If-None-Match", etag)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNotModified, rec.Code, "body: %s", rec.Body.String())
}
