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

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// GHSA-4qvp-948j-4cgv
func TestUserJWTRequiresLiveSession(t *testing.T) {
	const sid = "550e8400-e29b-41d4-a716-446655440001" // user 1
	jwtFor := func(t *testing.T, u *user.User, sessionID string) string {
		token, err := auth.NewUserJWTAuthtoken(u, sessionID)
		require.NoError(t, err)
		return token
	}

	for _, api := range []struct{ version, logout string }{
		{"v1", "/api/v1/user/logout"},
		{"v2", "/api/v2/logout"},
	} {
		t.Run(api.version, func(t *testing.T) {
			userRoute := "/api/" + api.version + "/user"

			t.Run("live session is accepted", func(t *testing.T) {
				e, err := setupTestEnv()
				require.NoError(t, err)
				rec := humaRequest(t, e, http.MethodGet, userRoute, "", jwtFor(t, &testuser1, sid), "")
				assert.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			})

			t.Run("logout revokes the bearer", func(t *testing.T) {
				e, err := setupTestEnv()
				require.NoError(t, err)
				token := jwtFor(t, &testuser1, sid)
				// Warm the session cache so the revocation has to bust it.
				require.Equal(t, http.StatusOK, humaRequest(t, e, http.MethodGet, userRoute, "", token, "").Code)

				rec := humaRequest(t, e, http.MethodPost, api.logout, "", token, "")
				require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

				rec = humaRequest(t, e, http.MethodGet, userRoute, "", token, "")
				assert.Equal(t, http.StatusUnauthorized, rec.Code, rec.Body.String())
				assert.Contains(t, rec.Body.String(), `"code":11`)
			})

			t.Run("revoked bearer cannot mint an oauth code", func(t *testing.T) {
				e, err := setupTestEnv()
				require.NoError(t, err)
				token := jwtFor(t, &testuser1, sid)
				require.Equal(t, http.StatusOK, humaRequest(t, e, http.MethodPost, api.logout, "", token, "").Code)

				body := authorizeRequestBody("code", "vikunja", "vikunja-flutter://callback", "abc123", "S256", "s")
				rec := humaRequest(t, e, http.MethodPost, "/api/"+api.version+"/oauth/authorize", string(body), token, "")
				assert.Equal(t, http.StatusUnauthorized, rec.Code, rec.Body.String())
			})

			t.Run("unknown session is rejected", func(t *testing.T) {
				e, err := setupTestEnv()
				require.NoError(t, err)
				rec := humaRequest(t, e, http.MethodGet, userRoute, "", jwtFor(t, &testuser1, "no-such-session"), "")
				assert.Equal(t, http.StatusUnauthorized, rec.Code, rec.Body.String())
			})

			t.Run("another user's session is rejected", func(t *testing.T) {
				e, err := setupTestEnv()
				require.NoError(t, err)
				rec := humaRequest(t, e, http.MethodGet, userRoute, "", jwtFor(t, &testuser2, sid), "")
				assert.Equal(t, http.StatusUnauthorized, rec.Code, rec.Body.String())
			})

			t.Run("frozen account is rejected", func(t *testing.T) {
				e, err := setupTestEnv()
				require.NoError(t, err)
				token := jwtFor(t, &testuser1, sid)
				require.Equal(t, http.StatusOK, humaRequest(t, e, http.MethodGet, userRoute, "", token, "").Code)
				setUserStatus(t, 1, user.StatusEmailConfirmationRequired)

				rec := humaRequest(t, e, http.MethodGet, userRoute, "", token, "")
				assert.Equal(t, http.StatusUnauthorized, rec.Code, rec.Body.String())
			})

			t.Run("link share tokens carry no session", func(t *testing.T) {
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
				rec := humaRequest(t, e, http.MethodGet, "/api/"+api.version+"/projects/1", "", token, "")
				assert.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			})
		})
	}
}
