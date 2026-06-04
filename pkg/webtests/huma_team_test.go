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
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// teamNamesFromListBody parses the v2 Paginated envelope and returns the names
// of the teams it contains, so tests can assert the EXACT result set rather
// than merely contains/not-contains. Mirrors v1's reflect-based length checks
// in pkg/models/teams_test.go but at the HTTP layer.
func teamNamesFromListBody(t *testing.T, body string) []string {
	t.Helper()
	var env struct {
		Items []struct {
			Name string `json:"name"`
		} `json:"items"`
	}
	require.NoError(t, json.Unmarshal([]byte(body), &env))
	names := make([]string, 0, len(env.Items))
	for _, it := range env.Items {
		names = append(names, it.Name)
	}
	return names
}

// TestHumaTeam mirrors v1's model TestTeam shape so v2 contract parity is
// readable side-by-side. Named TestHumaTeam to avoid clashing with the v1
// model test (pkg/models/teams_test.go TestTeam).
//
// This is a 1:1 port of the v1 model coverage (Team has no v1 webtest):
//   - pkg/models/teams_test.go: TestTeam_Create / ReadOne / ReadAll / Update /
//     Delete
//   - pkg/models/teams_permissions_test.go: TestTeam_CanDoSomething (the
//     owner/member/non-member permission matrix)
//
// Mapped to v2 HTTP semantics: 201 create, 204 delete, 403 forbidden, 404
// not-found, 422 validation. v2-only assertions (status codes, ETag) are kept
// on top of the v1 behaviours.
//
// Fixture facts (pkg/db/fixtures/team_members.yml, teams.yml): user1 is an
// ADMIN of team 1 and a non-admin member of teams 2-8. user2 is a non-admin
// member of team 1. Membership rows for teams 5/6/7 reference teams that don't
// exist in teams.yml, so they're dropped by the INNER JOIN — user1's effective
// teams are 1, 2, 3, 4 and 8 (exactly 5, matching v1's ReadAll count). Team 9
// (created by user 7) lists only user 2, so user1 is not a member. testteam13
// and testteam15 are public and list only user 10.
func TestHumaTeam(t *testing.T) {
	// Each subtest gets a fresh handler (and therefore a fresh setupTestEnv with
	// freshly loaded fixtures), mirroring v1's per-subtest db.LoadAndAssertFixtures.
	// Sharing one handler across subtests is unsafe here: setupTestEnv reruns
	// config.InitDefaultConfig, which rotates the random JWT secret, so a reused
	// echo.Echo would reject tokens signed after a sibling subtest reset the env.
	handlerFor := func(u *user.User) *webHandlerTestV2 {
		return &webHandlerTestV2{user: u, basePath: "/api/v2/teams", idParam: "team", t: t}
	}

	t.Run("ReadAll", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			testHandler := handlerFor(&testuser1)
			rec, err := testHandler.testReadAllWithUser(nil, nil)
			require.NoError(t, err)
			// User 1 is a member of teams 1-8.
			assert.Contains(t, rec.Body.String(), `testteam1`)
			// User 1 is not a member of team 9 (only user 2 is).
			assert.NotContains(t, rec.Body.String(), `testteam9`)
		})
		// Exact cardinality: v1's TestTeam_ReadAll/normal asserts len == 5.
		// Run against a pristine env so create/delete subtests can't perturb it.
		t.Run("Exact result set", func(t *testing.T) {
			h := handlerFor(&testuser1)
			rec, err := h.testReadAllWithUser(nil, nil)
			require.NoError(t, err)
			names := teamNamesFromListBody(t, rec.Body.String())
			assert.ElementsMatch(t, []string{
				"testteam1",
				"testteam2_read_only_on_project6",
				"testteam3_write_on_project7",
				"testteam4_admin_on_project8",
				"testteam8",
			}, names, "user1's teams are exactly 1,2,3,4,8 (5 total)")
		})
		// v1's TestTeam_ReadAll/search: q matches exactly one team (id 2).
		t.Run("Search", func(t *testing.T) {
			h := handlerFor(&testuser1)
			rec, err := h.testReadAllWithUser(url.Values{"q": []string{"READ_only_on_project6"}}, nil)
			require.NoError(t, err)
			names := teamNamesFromListBody(t, rec.Body.String())
			assert.Equal(t, []string{"testteam2_read_only_on_project6"}, names)
		})
		// testteam13 and testteam15 are public (teams.yml) and user1 is not a
		// member of either (team_members.yml only lists user 10 there).
		// v1's TestTeam_ReadAll/"public discovery disabled": with the instance
		// gate off, include_public is a no-op and the count stays 5.
		t.Run("Include public, but public teams disabled", func(t *testing.T) {
			// The config gate is off by default: include_public must be a no-op
			// so public teams the user is not a member of stay hidden.
			require.False(t, config.ServiceEnablePublicTeams.GetBool())

			h := handlerFor(&testuser1)

			// Without include_public.
			rec, err := h.testReadAllWithUser(nil, nil)
			require.NoError(t, err)
			assert.Len(t, teamNamesFromListBody(t, rec.Body.String()), 5)

			// With include_public the result is unchanged while the gate is off.
			rec, err = h.testReadAllWithUser(url.Values{"include_public": []string{"true"}}, nil)
			require.NoError(t, err)
			names := teamNamesFromListBody(t, rec.Body.String())
			assert.Len(t, names, 5)
			assert.Contains(t, names, "testteam1")
			assert.NotContains(t, names, "testteam13")
			assert.NotContains(t, names, "testteam15")
		})
		// v1's TestTeam_ReadAll/"public discovery enabled": with the gate on,
		// include_public surfaces the two public teams (count 5 -> 7).
		t.Run("Include public when public teams enabled", func(t *testing.T) {
			prev := config.ServiceEnablePublicTeams.GetBool()
			config.ServiceEnablePublicTeams.Set(true)
			defer config.ServiceEnablePublicTeams.Set(prev)

			h := handlerFor(&testuser1)

			// Without include_public the public teams stay hidden even with the
			// instance setting on; the count is still exactly 5.
			rec, err := h.testReadAllWithUser(nil, nil)
			require.NoError(t, err)
			names := teamNamesFromListBody(t, rec.Body.String())
			assert.Len(t, names, 5)
			assert.Contains(t, names, "testteam1")
			assert.NotContains(t, names, "testteam13")
			assert.NotContains(t, names, "testteam15")

			// With include_public=true the public teams the user is not a member
			// of are surfaced: 5 own teams + the 2 public teams = 7.
			rec, err = h.testReadAllWithUser(url.Values{"include_public": []string{"true"}}, nil)
			require.NoError(t, err)
			names = teamNamesFromListBody(t, rec.Body.String())
			assert.Len(t, names, 7)
			assert.Contains(t, names, "testteam1")
			assert.Contains(t, names, "testteam13")
			assert.Contains(t, names, "testteam15")
		})
	})

	t.Run("ReadOne", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			testHandler := handlerFor(&testuser1)
			rec, err := testHandler.testReadOneWithUser(nil, map[string]string{"team": "1"})
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"name":"testteam1"`)
			// v1's TestTeam_ReadOne also asserts the description and created_by.
			assert.Contains(t, rec.Body.String(), `"description":"Lorem Ipsum"`)
			assert.Contains(t, rec.Body.String(), `"created_by"`)
			assert.Contains(t, rec.Body.String(), `"max_permission":`)
			assert.NotEmpty(t, rec.Result().Header.Get("ETag"))
		})
		// v1's TestTeam_ReadOne/{invalid id, nonexisting} expects
		// ErrTeamDoesNotExist. At the HTTP layer CanRead refuses non-members
		// before existence is checked, so a missing team returns 403, not 404.
		t.Run("Nonexisting", func(t *testing.T) {
			testHandler := handlerFor(&testuser1)
			_, err := testHandler.testReadOneWithUser(nil, map[string]string{"team": "9999"})
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Permissions check", func(t *testing.T) {
			// Permission matrix from TestTeam_CanDoSomething: a non-member is
			// forbidden (CanRead == false).
			t.Run("Forbidden non-member", func(t *testing.T) {
				// Team 9: user1 is not a member.
				testHandler := handlerFor(&testuser1)
				_, err := testHandler.testReadOneWithUser(nil, map[string]string{"team": "9"})
				require.Error(t, err)
				assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
			})
			// CanRead is granted to ANY member, admin or not (Team.CanRead only
			// checks team membership). user2 is a non-admin member of team 1, so
			// the read must succeed even though update/delete below are denied.
			t.Run("Member but not admin can read", func(t *testing.T) {
				h := handlerFor(&testuser2)
				rec, err := h.testReadOneWithUser(nil, map[string]string{"team": "1"})
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"name":"testteam1"`)
			})
		})
	})

	t.Run("Create", func(t *testing.T) {
		// v1's TestTeam_Create/normal: creates a team and AssertExists with
		// is_public=false. Use a pristine env so we can read back + DB-assert.
		t.Run("Normal", func(t *testing.T) {
			h := handlerFor(&testuser1)
			rec, err := h.testCreateWithUser(nil, nil, `{"name":"Lorem","description":"Ipsum"}`)
			require.NoError(t, err)
			assert.Equal(t, http.StatusCreated, rec.Code)
			assert.Contains(t, rec.Body.String(), `"name":"Lorem"`)
			assert.Contains(t, rec.Body.String(), `"description":"Ipsum"`)
			// A freshly created team is private by default.
			assert.Contains(t, rec.Body.String(), `"is_public":false`)

			var created struct {
				ID int64 `json:"id"`
			}
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created))
			require.NotZero(t, created.ID)
			// DB persistence: mirrors v1's db.AssertExists.
			db.AssertExists(t, "teams", map[string]interface{}{
				"id":          created.ID,
				"name":        "Lorem",
				"description": "Ipsum",
				"is_public":   false,
			}, false)
		})
		// v1's TestTeam_Create/public: is_public=true must persist and read back.
		t.Run("Public", func(t *testing.T) {
			h := handlerFor(&testuser1)
			rec, err := h.testCreateWithUser(nil, nil, `{"name":"LoremPublic","description":"Ipsum","is_public":true}`)
			require.NoError(t, err)
			assert.Equal(t, http.StatusCreated, rec.Code)
			assert.Contains(t, rec.Body.String(), `"is_public":true`)

			var created struct {
				ID int64 `json:"id"`
			}
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created))
			require.NotZero(t, created.ID)
			db.AssertExists(t, "teams", map[string]interface{}{
				"id":        created.ID,
				"name":      "LoremPublic",
				"is_public": true,
			}, false)

			// Read it back through the API to prove the flag is served, too.
			rec, err = h.testReadOneWithUser(nil, map[string]string{"team": strconv.FormatInt(created.ID, 10)})
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"is_public":true`)
		})
		// v1's TestTeam_Create/"empty name" -> ErrTeamNameCannotBeEmpty. Name has
		// minLength:1, so Huma rejects an empty name with 422 before the model is
		// touched.
		t.Run("Empty name", func(t *testing.T) {
			testHandler := handlerFor(&testuser1)
			_, err := testHandler.testCreateWithUser(nil, nil, `{"name":""}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusUnprocessableEntity, getHTTPErrorCode(err))
		})
	})

	t.Run("Update", func(t *testing.T) {
		// v1's TestTeam_Update/normal: rename, then AssertExists. Use a pristine
		// env so we can DB-assert without other subtests interfering.
		t.Run("Normal", func(t *testing.T) {
			h := handlerFor(&testuser1)
			// Team 1: user1 is admin.
			rec, err := h.testUpdateWithUser(nil, map[string]string{"team": "1"}, `{"name":"SomethingNew"}`)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"name":"SomethingNew"`)
			// DB persistence: mirrors v1's db.AssertExists after Update.
			db.AssertExists(t, "teams", map[string]interface{}{
				"id":   1,
				"name": "SomethingNew",
			}, false)
		})
		// v1's TestTeam_Update/"empty name" -> ErrTeamNameCannotBeEmpty, this time
		// on an UPDATE (PUT) to an existing team the user can admin. Huma's
		// minLength:1 on the body rejects it with 422.
		t.Run("Empty name", func(t *testing.T) {
			h := handlerFor(&testuser1)
			_, err := h.testUpdateWithUser(nil, map[string]string{"team": "1"}, `{"name":""}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusUnprocessableEntity, getHTTPErrorCode(err))
		})
		// v1's TestTeam_Update/nonexisting -> ErrTeamDoesNotExist.
		// CanUpdate -> IsAdmin -> GetTeamByID surfaces it as 404.
		t.Run("Nonexisting", func(t *testing.T) {
			testHandler := handlerFor(&testuser1)
			_, err := testHandler.testUpdateWithUser(nil, map[string]string{"team": "9999"}, `{"name":"TestLoremIpsum"}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
		})
		t.Run("Permissions check", func(t *testing.T) {
			// Permission matrix: a member who is not admin cannot update
			// (CanUpdate == false).
			t.Run("Forbidden non-admin", func(t *testing.T) {
				// Team 2: user1 is a member but not an admin.
				testHandler := handlerFor(&testuser1)
				_, err := testHandler.testUpdateWithUser(nil, map[string]string{"team": "2"}, `{"name":"TestLoremIpsum"}`)
				require.Error(t, err)
				assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
			})
			// A non-member is likewise forbidden (team 9, user1 not a member).
			t.Run("Forbidden non-member", func(t *testing.T) {
				testHandler := handlerFor(&testuser1)
				_, err := testHandler.testUpdateWithUser(nil, map[string]string{"team": "9"}, `{"name":"TestLoremIpsum"}`)
				require.Error(t, err)
				assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
			})
		})
	})

	t.Run("Delete", func(t *testing.T) {
		// v1's TestTeam_Delete/normal: delete, then AssertMissing. Use a pristine
		// env so the DB-missing assertion is unambiguous.
		t.Run("Normal", func(t *testing.T) {
			h := handlerFor(&testuser1)
			// Team 1: user1 is admin.
			rec, err := h.testDeleteWithUser(nil, map[string]string{"team": "1"})
			require.NoError(t, err)
			// v2 delete is 204 No Content; v1 returned 200 + a message body.
			assert.Equal(t, http.StatusNoContent, rec.Code)
			assert.Empty(t, rec.Body.String())
			// DB persistence: mirrors v1's db.AssertMissing after Delete.
			db.AssertMissing(t, "teams", map[string]interface{}{"id": 1})
		})
		t.Run("Nonexisting", func(t *testing.T) {
			testHandler := handlerFor(&testuser1)
			_, err := testHandler.testDeleteWithUser(nil, map[string]string{"team": "9999"})
			require.Error(t, err)
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
		})
		t.Run("Permissions check", func(t *testing.T) {
			// Permission matrix: a member who is not admin cannot delete
			// (CanDelete == false).
			t.Run("Forbidden non-admin", func(t *testing.T) {
				// Team 2: user1 is a member but not an admin.
				testHandler := handlerFor(&testuser1)
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{"team": "2"})
				require.Error(t, err)
				assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
			})
			// A non-member is likewise forbidden (team 9, user1 not a member).
			t.Run("Forbidden non-member", func(t *testing.T) {
				testHandler := handlerFor(&testuser1)
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{"team": "9"})
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

func TestHumaTeam_ETagReflectsPermission(t *testing.T) {
	// Team 1: user1 is an admin (max_permission 2), user2 a non-admin member (0).
	// Same team, so the per-caller ETag must differ — else a 304 serves stale perms.
	e, err := setupTestEnv()
	require.NoError(t, err)

	admin := humaRequest(t, e, http.MethodGet, "/api/v2/teams/1", "", humaTokenFor(t, &testuser1), "")
	require.Equal(t, http.StatusOK, admin.Code, "body: %s", admin.Body.String())
	member := humaRequest(t, e, http.MethodGet, "/api/v2/teams/1", "", humaTokenFor(t, &testuser2), "")
	require.Equal(t, http.StatusOK, member.Code, "body: %s", member.Body.String())

	assert.NotEmpty(t, admin.Header().Get("ETag"))
	assert.NotEqual(t, admin.Header().Get("ETag"), member.Header().Get("ETag"),
		"same team, different caller permission must produce different ETags")
}
