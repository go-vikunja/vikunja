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
	"strconv"
	"testing"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHumaCalDAVToken covers the v2 CalDAV token lifecycle. All calls share one
// echo env because setupTestEnv rotates the JWT signing key per call, which would
// 401 a token minted against an earlier env.
//
// Fixture (pkg/db/fixtures/user_tokens.yml): token id 6, kind 4 (CalDAV),
// belongs to user10. user1 starts with no CalDAV tokens.
func TestHumaCalDAVToken(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)

	user1Token := humaTokenFor(t, &testuser1)
	user10Token := humaTokenFor(t, &testuser10)

	t.Run("Create returns the clear-text token", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodPost, "/api/v2/user/settings/token/caldav", "", user1Token, "")
		require.Equal(t, http.StatusCreated, rec.Code, "body: %s", rec.Body.String())

		var created struct {
			ID    int64  `json:"id"`
			Token string `json:"token"`
		}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created), "body: %s", rec.Body.String())
		assert.NotZero(t, created.ID)
		assert.NotEmpty(t, created.Token, "the clear-text token must be returned on create")
	})

	t.Run("List omits the token value", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/user/settings/token/caldav", "", user1Token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())

		ids := caldavTokenIDsFromList(t, rec.Body.Bytes())
		assert.NotEmpty(t, ids, "the token created above must show up in the list")
		assert.Empty(t, caldavTokenValuesFromList(t, rec.Body.Bytes()),
			"the clear-text token must never appear in the list; body: %s", rec.Body.String())
	})

	t.Run("List is scoped to the current user", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/user/settings/token/caldav", "", user10Token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, caldavTokenIDsFromList(t, rec.Body.Bytes()), int64(6),
			"user10's fixture token #6 must be listed; body: %s", rec.Body.String())
	})

	t.Run("Delete removes the token", func(t *testing.T) {
		listRec := humaRequest(t, e, http.MethodGet, "/api/v2/user/settings/token/caldav", "", user1Token, "")
		require.Equal(t, http.StatusOK, listRec.Code, "body: %s", listRec.Body.String())
		ids := caldavTokenIDsFromList(t, listRec.Body.Bytes())
		require.NotEmpty(t, ids)

		del := humaRequest(t, e, http.MethodDelete, "/api/v2/user/settings/token/caldav/"+strconv.FormatInt(ids[0], 10), "", user1Token, "")
		require.Equal(t, http.StatusNoContent, del.Code, "body: %s", del.Body.String())

		afterRec := humaRequest(t, e, http.MethodGet, "/api/v2/user/settings/token/caldav", "", user1Token, "")
		require.Equal(t, http.StatusOK, afterRec.Code, "body: %s", afterRec.Body.String())
		assert.NotContains(t, caldavTokenIDsFromList(t, afterRec.Body.Bytes()), ids[0],
			"the deleted token must be gone; body: %s", afterRec.Body.String())
	})

	t.Run("Delete is scoped to the current user", func(t *testing.T) {
		// Token #6 belongs to user10; user1 deleting it is a no-op (204), not an error.
		del := humaRequest(t, e, http.MethodDelete, "/api/v2/user/settings/token/caldav/6", "", user1Token, "")
		require.Equal(t, http.StatusNoContent, del.Code, "body: %s", del.Body.String())

		rec := humaRequest(t, e, http.MethodGet, "/api/v2/user/settings/token/caldav", "", user10Token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, caldavTokenIDsFromList(t, rec.Body.Bytes()), int64(6),
			"user10's token #6 must survive a delete attempt by another user; body: %s", rec.Body.String())
	})
}

// TestHumaCalDAVToken_LinkShareForbidden ports v1's implicit guard: a link share
// is not a user, so create / list / delete all refuse it (403).
func TestHumaCalDAVToken_LinkShareForbidden(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	token, err := auth.NewLinkShareJWTAuthtoken(&models.LinkSharing{
		ID:          1,
		Hash:        "test",
		ProjectID:   1,
		Permission:  models.PermissionRead,
		SharingType: models.SharingTypeWithoutPassword,
		SharedByID:  1,
	})
	require.NoError(t, err)

	t.Run("create", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodPost, "/api/v2/user/settings/token/caldav", "", token, "")
		assert.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
	})
	t.Run("list", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/user/settings/token/caldav", "", token, "")
		assert.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
	})
	t.Run("delete", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodDelete, "/api/v2/user/settings/token/caldav/6", "", token, "")
		assert.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
	})
}

func caldavTokenIDsFromList(t *testing.T, body []byte) []int64 {
	t.Helper()
	items := caldavTokenItemsFromList(t, body)
	ids := make([]int64, 0, len(items))
	for _, it := range items {
		ids = append(ids, it.ID)
	}
	return ids
}

func caldavTokenValuesFromList(t *testing.T, body []byte) []string {
	t.Helper()
	values := []string{}
	for _, it := range caldavTokenItemsFromList(t, body) {
		if it.Token != "" {
			values = append(values, it.Token)
		}
	}
	return values
}

func caldavTokenItemsFromList(t *testing.T, body []byte) []struct {
	ID    int64  `json:"id"`
	Token string `json:"token"`
} {
	t.Helper()
	var resp struct {
		Items []struct {
			ID    int64  `json:"id"`
			Token string `json:"token"`
		} `json:"items"`
	}
	require.NoError(t, json.Unmarshal(body, &resp), "list body must be a paginated envelope: %s", string(body))
	return resp.Items
}
