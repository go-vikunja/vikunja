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
	"fmt"
	"net/http"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"

	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// GHSA-4hv6-xc92-j86g
func TestSessionRevocationDispatchesEvent(t *testing.T) {
	const sid = "550e8400-e29b-41d4-a716-446655440001" // user 1

	revoked := func() []*models.SessionsRevokedEvent {
		var out []*models.SessionsRevokedEvent
		for _, e := range events.GetDispatchedEvents((&models.SessionsRevokedEvent{}).Name()) {
			out = append(out, e.(*models.SessionsRevokedEvent))
		}
		return out
	}
	tokenWithSession := func(t *testing.T) string {
		token, err := auth.NewUserJWTAuthtoken(&testuser1, sid)
		require.NoError(t, err)
		return token
	}

	for _, api := range []struct {
		version, sessions, logout, totpEnable, password string
		deleteSessionStatus                             int
	}{
		{"v1", "/api/v1/user/sessions/", "/api/v1/user/logout", "/api/v1/user/settings/totp/enable", "/api/v1/user/password", http.StatusOK},
		{"v2", "/api/v2/user/sessions/", "/api/v2/logout", "/api/v2/user/settings/totp/enable", "/api/v2/user/password", http.StatusNoContent},
	} {
		t.Run(api.version, func(t *testing.T) {
			t.Run("deleting a session", func(t *testing.T) {
				e, err := setupTestEnv()
				require.NoError(t, err)
				rec := apiTokenReq(e, http.MethodDelete, api.sessions+sid, tokenWithSession(t), "")
				require.Equal(t, api.deleteSessionStatus, rec.Code, rec.Body.String())
				assert.Equal(t, []*models.SessionsRevokedEvent{{UserID: 1, SessionID: sid}}, revoked())
			})
			t.Run("logging out", func(t *testing.T) {
				e, err := setupTestEnv()
				require.NoError(t, err)
				rec := apiTokenReq(e, http.MethodPost, api.logout, tokenWithSession(t), "")
				require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
				assert.Equal(t, []*models.SessionsRevokedEvent{{UserID: 1, SessionID: sid}}, revoked())
			})
			t.Run("enabling TOTP", func(t *testing.T) {
				e, err := setupTestEnv()
				require.NoError(t, err)
				passcode, err := totp.GenerateCode("HXDMVJECJJWSRB3HWIZR4IFUGFTMXBOZ", time.Now())
				require.NoError(t, err)
				rec := apiTokenReq(e, http.MethodPost, api.totpEnable, tokenWithSession(t), fmt.Sprintf(`{"passcode":%q}`, passcode))
				require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
				assert.Equal(t, []*models.SessionsRevokedEvent{{UserID: 1}}, revoked())
			})
			t.Run("changing the password", func(t *testing.T) {
				e, err := setupTestEnv()
				require.NoError(t, err)
				rec := apiTokenReq(e, http.MethodPost, api.password, tokenWithSession(t), `{"old_password":"12345678","new_password":"123456789"}`)
				require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
				assert.Equal(t, []*models.SessionsRevokedEvent{{UserID: 1}}, revoked())
			})
			t.Run("resetting the password", func(t *testing.T) {
				e, err := setupTestEnv()
				require.NoError(t, err)
				rec := humaRequest(t, e, http.MethodPost, api.password+"/reset", `{"token":"passwordresettesttoken","new_password":"12345678"}`, "", "")
				require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
				assert.Equal(t, []*models.SessionsRevokedEvent{{UserID: 3}}, revoked())
			})
		})
	}
}
