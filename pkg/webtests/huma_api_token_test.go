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

	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Bot owner from pkg/db/fixtures/users.yml: user21 owns bot 23. No api_tokens
// fixture is owned by a bot, so listing a caller-owned bot's tokens returns an
// authorized (but empty) list.
var testuser21 = user.User{ID: 21, Username: "user_bot_owner_a", Issuer: "local"}

// TestHumaAPIToken ports the v1 model-level permission matrix
// (pkg/models/api_tokens_test.go) onto the v2 HTTP surface so /api/v2/tokens
// independently proves the full owner/bot-owner contract once v1's routes are
// gone. /tokens supports list/create/delete only — no ReadOne, no Update.
//
// Every request runs against one shared echo.Echo: setupTestEnv re-randomizes
// the global JWT secret on each call, so a second env would invalidate the
// first's tokens. One env, one secret, all callers via humaTokenFor.
//
// Fixture topology (pkg/db/fixtures/api_tokens.yml):
//   - tokens #1, #2: owned by user1.
//   - token #3: owned by user2 — never visible to user1, never deletable by user1.
//   - tokens #4, #5: owned by disabled/locked users 17/18.
//   - tokens #6, #7: owned by user15.
//   - token #8: owned by user13.
func TestHumaAPIToken(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	user1Token := humaTokenFor(t, &testuser1)
	user2Token := humaTokenFor(t, &testuser2)
	botOwnerToken := humaTokenFor(t, &testuser21)

	t.Run("ReadAll", func(t *testing.T) {
		t.Run("Normal - exact owned set for user1", func(t *testing.T) {
			rec := humaRequest(t, e, http.MethodGet, "/api/v2/tokens", "", user1Token, "")
			require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
			ids := apiTokenIDsFromReadAll(t, rec.Body.Bytes())
			// user1 owns exactly tokens #1, #2 and the MCP tokens #9-#11;
			// cardinality is pinned.
			assert.ElementsMatch(t, []int64{1, 2, 9, 10, 11}, ids,
				"ReadAll must return exactly {1,2,9,10,11}; body: %s", rec.Body.String())
			assert.Equal(t, int64(5), apiTokenTotalFromReadAll(t, rec.Body.Bytes()))
			assert.NotContains(t, ids, int64(3), "token #3 (owned by user2) must be hidden")
		})
		t.Run("Isolation - user2 sees only its own token", func(t *testing.T) {
			rec := humaRequest(t, e, http.MethodGet, "/api/v2/tokens", "", user2Token, "")
			require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
			ids := apiTokenIDsFromReadAll(t, rec.Body.Bytes())
			assert.ElementsMatch(t, []int64{3}, ids,
				"user2 must see only token #3; body: %s", rec.Body.String())
			assert.NotContains(t, ids, int64(1), "token #1 (owned by user1) must be hidden from user2")
			assert.NotContains(t, ids, int64(2), "token #2 (owned by user1) must be hidden from user2")
		})
		t.Run("Search by title", func(t *testing.T) {
			rec := humaRequest(t, e, http.MethodGet, "/api/v2/tokens?q=test+token+1", "", user1Token, "")
			require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
			ids := apiTokenIDsFromReadAll(t, rec.Body.Bytes())
			assert.ElementsMatch(t, []int64{1}, ids,
				"search must narrow to token #1; body: %s", rec.Body.String())
		})
		t.Run("owner_id - caller's own bot returns an authorized (empty) list", func(t *testing.T) {
			rec := humaRequest(t, e, http.MethodGet, "/api/v2/tokens?owner_id=23", "", botOwnerToken, "")
			require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
			ids := apiTokenIDsFromReadAll(t, rec.Body.Bytes())
			assert.Empty(t, ids, "bot 23 has no tokens; list must be empty but authorized; body: %s", rec.Body.String())
		})
		t.Run("owner_id - forbidden when not the bot's owner", func(t *testing.T) {
			// user1 is not the owner of bot 23 (owned by user21).
			rec := humaRequest(t, e, http.MethodGet, "/api/v2/tokens?owner_id=23", "", user1Token, "")
			assert.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
		})
		t.Run("owner_id - forbidden when the target is not a bot", func(t *testing.T) {
			// user2 is a real user, not a bot owned by user1.
			rec := humaRequest(t, e, http.MethodGet, "/api/v2/tokens?owner_id=2", "", user1Token, "")
			assert.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
		})
	})

	t.Run("Create", func(t *testing.T) {
		t.Run("Normal - returns cleartext token once and sets owner", func(t *testing.T) {
			rec := humaRequest(t, e, http.MethodPost, "/api/v2/tokens",
				`{"title":"new token","permissions":{"tasks":["read_all"]},"expires_at":"2099-01-01T00:00:00Z"}`,
				user1Token, "")
			require.Equal(t, http.StatusCreated, rec.Code, "body: %s", rec.Body.String())

			var created struct {
				ID      int64  `json:"id"`
				Token   string `json:"token"`
				Title   string `json:"title"`
				OwnerID int64  `json:"owner_id"`
			}
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created))
			assert.Equal(t, "new token", created.Title)
			assert.NotEmpty(t, created.Token, "create response must include the cleartext token exactly once")
			require.Greater(t, len(created.Token), 3)
			assert.Equal(t, "tk_", created.Token[:3], "cleartext token must carry the tk_ prefix; got %q", created.Token)
			assert.Equal(t, int64(1), created.OwnerID, "owner must default to the authenticated user")
		})
		t.Run("Empty title", func(t *testing.T) {
			// v2 enforces required title at the schema level → 422.
			rec := humaRequest(t, e, http.MethodPost, "/api/v2/tokens",
				`{"title":"","permissions":{"tasks":["read_all"]},"expires_at":"2099-01-01T00:00:00Z"}`,
				user1Token, "")
			assert.Equal(t, http.StatusUnprocessableEntity, rec.Code, "body: %s", rec.Body.String())
		})
		t.Run("owner_id - forbidden when the target is not a caller-owned bot", func(t *testing.T) {
			// user1 cannot mint a token for user2 (not a bot they own).
			rec := humaRequest(t, e, http.MethodPost, "/api/v2/tokens",
				`{"title":"sneaky","owner_id":2,"permissions":{"tasks":["read_all"]},"expires_at":"2099-01-01T00:00:00Z"}`,
				user1Token, "")
			assert.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
		})
	})

	t.Run("Delete", func(t *testing.T) {
		t.Run("Forbidden - token of another user", func(t *testing.T) {
			// Token #3 belongs to user2; user1 must not be able to delete it.
			rec := humaRequest(t, e, http.MethodDelete, "/api/v2/tokens/3", "", user1Token, "")
			assert.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
		})
		t.Run("Nonexisting", func(t *testing.T) {
			// CanDelete returns (false, nil) for a missing token → generic forbidden.
			rec := humaRequest(t, e, http.MethodDelete, "/api/v2/tokens/9999", "", user1Token, "")
			assert.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
		})
		t.Run("Normal - own token", func(t *testing.T) {
			// Runs last so deleting #1 doesn't disturb the ReadAll cardinality assertions.
			rec := humaRequest(t, e, http.MethodDelete, "/api/v2/tokens/1", "", user1Token, "")
			require.Equal(t, http.StatusNoContent, rec.Code, "body: %s", rec.Body.String())
			assert.Empty(t, rec.Body.String())
		})
	})
}

func apiTokenIDsFromReadAll(t *testing.T, body []byte) []int64 {
	t.Helper()
	var resp struct {
		Items []struct {
			ID int64 `json:"id"`
		} `json:"items"`
	}
	require.NoError(t, json.Unmarshal(body, &resp), "ReadAll body must be a paginated envelope: %s", string(body))
	ids := make([]int64, 0, len(resp.Items))
	for _, it := range resp.Items {
		ids = append(ids, it.ID)
	}
	return ids
}

func apiTokenTotalFromReadAll(t *testing.T, body []byte) int64 {
	t.Helper()
	var resp struct {
		Total int64 `json:"total"`
	}
	require.NoError(t, json.Unmarshal(body, &resp), "ReadAll body must be a paginated envelope: %s", string(body))
	return resp.Total
}
