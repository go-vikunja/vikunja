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
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupIdentityTestEnv(t *testing.T) *echo.Echo {
	t.Helper()
	config.AuthOpenIDEnabled.Set(true)
	t.Cleanup(func() { config.AuthOpenIDEnabled.Set(false) })
	e, err := setupTestEnv()
	require.NoError(t, err)
	return e
}

func TestHumaUserIdentities(t *testing.T) {
	const identitiesPath = "/api/v2/user/settings/identities"

	t.Run("not registered when OpenID is disabled", func(t *testing.T) {
		config.AuthOpenIDEnabled.Set(false)
		e, err := setupTestEnv()
		require.NoError(t, err)

		rec := humaRequest(t, e, http.MethodGet, identitiesPath, "", humaTokenFor(t, &testuser1), "")
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("list", func(t *testing.T) {
		t.Run("returns the own identities", func(t *testing.T) {
			e := setupIdentityTestEnv(t)

			rec := humaRequest(t, e, http.MethodGet, identitiesPath, "", humaTokenFor(t, &testuser1), "")
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

			var body struct {
				Items []map[string]any `json:"items"`
				Total int64            `json:"total"`
			}
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
			require.Len(t, body.Items, 2)
			assert.Equal(t, int64(2), body.Total)
			assert.InDelta(t, 1, body.Items[0]["id"], 0)
			assert.Equal(t, "https://some.service.com", body.Items[0]["issuer"])
			assert.Contains(t, body.Items[0], "created")
			assert.Empty(t, body.Items[0]["provider_key"])
			assert.Empty(t, body.Items[0]["provider_name"])
			assert.NotContains(t, body.Items[0], "subject")
			assert.NotContains(t, body.Items[0], "user_id")
		})
		t.Run("is empty for a user without identities", func(t *testing.T) {
			e := setupIdentityTestEnv(t)

			rec := humaRequest(t, e, http.MethodGet, identitiesPath, "", humaTokenFor(t, &testuser2), "")
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			assert.Contains(t, rec.Body.String(), `"items":[]`)
		})
	})

	t.Run("link", func(t *testing.T) {
		t.Run("refuses a non-local user", func(t *testing.T) {
			e := setupIdentityTestEnv(t)
			oidcUser := &user.User{
				ID:       14,
				Username: "user14",
			}

			body := `{"provider":"p","code":"code","redirect_url":"https://localhost","password":"12345678"}`
			rec := humaRequest(t, e, http.MethodPost, identitiesPath, body, humaTokenFor(t, oidcUser), "")
			assert.Equal(t, http.StatusPreconditionFailed, rec.Code, rec.Body.String())
			assert.Equal(t, user.ErrCodeAccountIsNotLocal, problemCode(t, rec))
		})
	})

	t.Run("unlink", func(t *testing.T) {
		t.Run("removes the own identity", func(t *testing.T) {
			e := setupIdentityTestEnv(t)

			rec := humaRequest(t, e, http.MethodPost, identitiesPath+"/1/unlink", `{"password":"12345678"}`, humaTokenFor(t, &testuser1), "")
			assert.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())
			db.AssertMissing(t, "user_identities", map[string]interface{}{
				"id": 1,
			})
		})
		t.Run("another user's identity is not found", func(t *testing.T) {
			e := setupIdentityTestEnv(t)

			rec := humaRequest(t, e, http.MethodPost, identitiesPath+"/1/unlink", `{"password":"12345678"}`, humaTokenFor(t, &testuser6), "")
			assert.Equal(t, http.StatusNotFound, rec.Code, rec.Body.String())
			assert.Equal(t, models.ErrCodeUserIdentityDoesNotExist, problemCode(t, rec))
			db.AssertExists(t, "user_identities", map[string]interface{}{
				"id": 1,
			}, false)
		})
	})
}
