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

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHumaTeamMember ports the v1 model coverage (pkg/models/team_members_test.go:
// TestTeamMember_Create / _Delete / _Update) plus the permission matrix from
// CanCreate/CanDelete/CanUpdate to /api/v2. It drives the Echo+Huma stack
// directly (humaRequest/humaTokenFor) because these are action sub-paths under a
// team that webHandlerTestV2's buildURL does not model.
//
// Fixture facts (team_members.yml, teams.yml): user1 is an ADMIN of team 1 and a
// non-admin member of teams 2-8. user2 is a non-admin member of team 1. Team 9
// lists only user2. Teams 2/3/4 each list only user1 (non-admin). Teams 14/15
// are external (external_id set) — team 14 lists user10 and user11.
func TestHumaTeamMember(t *testing.T) {
	// user11 is a member of external team 14; not pre-defined in integrations.go.
	testuser11 := user.User{
		ID:       11,
		Username: "user11",
		Email:    "user11@example.com",
		Issuer:   "local",
	}

	t.Run("Add", func(t *testing.T) {
		// v1's TestTeamMember_Create/normal: team admin adds a user by username.
		t.Run("Normal", func(t *testing.T) {
			e, err := setupTestEnv()
			require.NoError(t, err)
			token := humaTokenFor(t, &testuser1) // admin of team 1

			rec := humaRequest(t, e, http.MethodPost, "/api/v2/teams/1/members", `{"username":"user3"}`, token, "")
			require.Equal(t, http.StatusCreated, rec.Code, "body: %s", rec.Body.String())
			assert.Contains(t, rec.Body.String(), `"username":"user3"`)
			// user3 has id 3.
			db.AssertExists(t, "team_members", map[string]interface{}{
				"team_id": 1,
				"user_id": 3,
			}, false)
		})
		// v1's TestTeamMember_Create/normal also allows seeding an admin member.
		t.Run("As admin", func(t *testing.T) {
			e, err := setupTestEnv()
			require.NoError(t, err)
			token := humaTokenFor(t, &testuser1)

			rec := humaRequest(t, e, http.MethodPost, "/api/v2/teams/1/members", `{"username":"user3","admin":true}`, token, "")
			require.Equal(t, http.StatusCreated, rec.Code, "body: %s", rec.Body.String())
			assert.Contains(t, rec.Body.String(), `"admin":true`)
			db.AssertExists(t, "team_members", map[string]interface{}{
				"team_id": 1,
				"user_id": 3,
				"admin":   true,
			}, false)
		})
		// v1's TestTeamMember_Create/"already existing" -> ErrUserIsMemberOfTeam (409).
		t.Run("Already a member", func(t *testing.T) {
			e, err := setupTestEnv()
			require.NoError(t, err)
			token := humaTokenFor(t, &testuser1)

			rec := humaRequest(t, e, http.MethodPost, "/api/v2/teams/1/members", `{"username":"user1"}`, token, "")
			require.Equal(t, http.StatusConflict, rec.Code, "body: %s", rec.Body.String())
		})
		// v1's TestTeamMember_Create/"nonexisting user" -> ErrUserDoesNotExist (404).
		t.Run("Nonexisting user", func(t *testing.T) {
			e, err := setupTestEnv()
			require.NoError(t, err)
			token := humaTokenFor(t, &testuser1)

			rec := humaRequest(t, e, http.MethodPost, "/api/v2/teams/1/members", `{"username":"nonexistinguser"}`, token, "")
			require.Equal(t, http.StatusNotFound, rec.Code, "body: %s", rec.Body.String())
		})
		// v1's TestTeamMember_Create/"empty username": required by valid: tag -> 422.
		t.Run("Empty username", func(t *testing.T) {
			e, err := setupTestEnv()
			require.NoError(t, err)
			token := humaTokenFor(t, &testuser1)

			rec := humaRequest(t, e, http.MethodPost, "/api/v2/teams/1/members", `{"username":""}`, token, "")
			require.Equal(t, http.StatusUnprocessableEntity, rec.Code, "body: %s", rec.Body.String())
		})
		t.Run("Permissions check", func(t *testing.T) {
			// CanCreate -> IsAdmin: a non-admin member cannot add members.
			t.Run("Forbidden non-admin member", func(t *testing.T) {
				e, err := setupTestEnv()
				require.NoError(t, err)
				token := humaTokenFor(t, &testuser2) // member of team 1, not admin

				rec := humaRequest(t, e, http.MethodPost, "/api/v2/teams/1/members", `{"username":"user3"}`, token, "")
				require.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
			})
			// A non-member is likewise forbidden (team 9, user1 not a member).
			t.Run("Forbidden non-member", func(t *testing.T) {
				e, err := setupTestEnv()
				require.NoError(t, err)
				token := humaTokenFor(t, &testuser1)

				rec := humaRequest(t, e, http.MethodPost, "/api/v2/teams/9/members", `{"username":"user3"}`, token, "")
				require.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
			})
		})
	})

	t.Run("Remove", func(t *testing.T) {
		// v1's TestTeamMember_Delete/normal: team admin removes a member.
		t.Run("Normal", func(t *testing.T) {
			e, err := setupTestEnv()
			require.NoError(t, err)
			token := humaTokenFor(t, &testuser1) // admin of team 1

			rec := humaRequest(t, e, http.MethodDelete, "/api/v2/teams/1/members/user2", "", token, "")
			require.Equal(t, http.StatusNoContent, rec.Code, "body: %s", rec.Body.String())
			assert.Empty(t, rec.Body.String())
			db.AssertMissing(t, "team_members", map[string]interface{}{
				"team_id": 1,
				"user_id": 2,
			})
		})
		// CanDelete grants self-removal even to a non-admin: user2 removes itself
		// from team 1 (which still has user1 afterwards, so not the last member).
		t.Run("Member can remove themselves", func(t *testing.T) {
			e, err := setupTestEnv()
			require.NoError(t, err)
			token := humaTokenFor(t, &testuser2) // non-admin member of team 1

			rec := humaRequest(t, e, http.MethodDelete, "/api/v2/teams/1/members/user2", "", token, "")
			require.Equal(t, http.StatusNoContent, rec.Code, "body: %s", rec.Body.String())
			db.AssertMissing(t, "team_members", map[string]interface{}{
				"team_id": 1,
				"user_id": 2,
			})
		})
		// Delete guard: removing the only member is refused. Team 2 lists just
		// user1; CanDelete passes (self), then the last-member guard rejects it.
		t.Run("Cannot remove last member", func(t *testing.T) {
			e, err := setupTestEnv()
			require.NoError(t, err)
			token := humaTokenFor(t, &testuser1)

			rec := humaRequest(t, e, http.MethodDelete, "/api/v2/teams/2/members/user1", "", token, "")
			require.Equal(t, http.StatusBadRequest, rec.Code, "body: %s", rec.Body.String())
		})
		// Delete guard: members of an external (oidc/ldap) team cannot be removed.
		// Team 14 has external_id set; user11 removing itself hits that guard.
		t.Run("Cannot remove from external team", func(t *testing.T) {
			e, err := setupTestEnv()
			require.NoError(t, err)
			token := humaTokenFor(t, &testuser11)

			rec := humaRequest(t, e, http.MethodDelete, "/api/v2/teams/14/members/user11", "", token, "")
			require.Equal(t, http.StatusPreconditionFailed, rec.Code, "body: %s", rec.Body.String())
		})
		t.Run("Permissions check", func(t *testing.T) {
			// CanDelete -> IsAdmin (when not self): a non-admin member cannot remove
			// someone else. user2 (non-admin) tries to remove user1 from team 1.
			t.Run("Forbidden non-admin removing another", func(t *testing.T) {
				e, err := setupTestEnv()
				require.NoError(t, err)
				token := humaTokenFor(t, &testuser2)

				rec := humaRequest(t, e, http.MethodDelete, "/api/v2/teams/1/members/user1", "", token, "")
				require.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
			})
			// A non-member is forbidden (team 9, user1 not a member).
			t.Run("Forbidden non-member", func(t *testing.T) {
				e, err := setupTestEnv()
				require.NoError(t, err)
				token := humaTokenFor(t, &testuser1)

				rec := humaRequest(t, e, http.MethodDelete, "/api/v2/teams/9/members/user2", "", token, "")
				require.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
			})
		})
	})

	t.Run("ToggleAdmin", func(t *testing.T) {
		// v1's TestTeamMember_Update/normal: the toggle flips the flag. user2 is a
		// non-admin member of team 1, so toggling makes them an admin.
		t.Run("Normal", func(t *testing.T) {
			e, err := setupTestEnv()
			require.NoError(t, err)
			token := humaTokenFor(t, &testuser1) // admin of team 1

			rec := humaRequest(t, e, http.MethodPost, "/api/v2/teams/1/members/user2/admin", "", token, "")
			require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
			assert.Contains(t, rec.Body.String(), `"admin":true`)
			db.AssertExists(t, "team_members", map[string]interface{}{
				"team_id": 1,
				"user_id": 2,
				"admin":   true,
			}, false)
		})
		// v1's TestTeamMember_Update/"explicitly false in payload": the body is
		// ignored and the flag toggled regardless — user1 (admin of team 1) becomes
		// a non-admin despite admin:false in the payload.
		t.Run("Body is ignored", func(t *testing.T) {
			e, err := setupTestEnv()
			require.NoError(t, err)
			token := humaTokenFor(t, &testuser1)

			rec := humaRequest(t, e, http.MethodPost, "/api/v2/teams/1/members/user1/admin", `{"admin":false}`, token, "")
			require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
			assert.Contains(t, rec.Body.String(), `"admin":false`)
			db.AssertExists(t, "team_members", map[string]interface{}{
				"team_id": 1,
				"user_id": 1,
				"admin":   false,
			}, false)
		})
		t.Run("Permissions check", func(t *testing.T) {
			// CanUpdate -> IsAdmin: a non-admin member cannot toggle admin status,
			// not even their own. The handler checks this explicitly (non-CRUD action).
			t.Run("Forbidden non-admin member", func(t *testing.T) {
				e, err := setupTestEnv()
				require.NoError(t, err)
				token := humaTokenFor(t, &testuser2) // non-admin member of team 1

				rec := humaRequest(t, e, http.MethodPost, "/api/v2/teams/1/members/user2/admin", "", token, "")
				require.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
			})
			// A non-member is forbidden (team 9, user1 not a member).
			t.Run("Forbidden non-member", func(t *testing.T) {
				e, err := setupTestEnv()
				require.NoError(t, err)
				token := humaTokenFor(t, &testuser1)

				rec := humaRequest(t, e, http.MethodPost, "/api/v2/teams/9/members/user2/admin", "", token, "")
				require.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
			})
		})
	})
}
