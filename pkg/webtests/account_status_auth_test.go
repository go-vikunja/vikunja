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
	"context"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setUserStatus(t *testing.T, id int64, status user.Status) {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	require.NoError(t, user.SetUserStatus(s, &user.User{ID: id}, status))
	require.NoError(t, s.Commit())
	events.DispatchPending(context.Background(), s)
}

// Admins freeze accounts by requiring email confirmation.
func TestEmailUnconfirmedAccountCannotAuthenticate(t *testing.T) {
	t.Run("api token", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		setUserStatus(t, 1, user.StatusEmailConfirmationRequired)

		rec := apiTokenReq(e, http.MethodGet, "/api/v1/tasks", "tk_2eef46f40ebab3304919ab2e7e39993f75f29d2e", "")
		assert.Equal(t, http.StatusUnauthorized, rec.Code, rec.Body.String())
	})

	t.Run("refresh token", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		setUserStatus(t, 1, user.StatusEmailConfirmationRequired)

		rec := refreshRequest(e, "testtoken_session1")
		assert.Equal(t, http.StatusPreconditionFailed, rec.Code, rec.Body.String())
		assert.NotContains(t, rec.Body.String(), `"token":"`)
		db.AssertMissing(t, "sessions", map[string]interface{}{"id": "550e8400-e29b-41d4-a716-446655440001"})
	})

	t.Run("oauth code exchange", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		codeVerifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
		h := sha256.Sum256([]byte(codeVerifier))
		code := getAuthorizationCode(t, e, base64.RawURLEncoding.EncodeToString(h[:]), "xyz")
		setUserStatus(t, 1, user.StatusEmailConfirmationRequired)

		rec := doTokenRequest(e, map[string]string{
			"grant_type":    "authorization_code",
			"code":          code,
			"client_id":     "vikunja",
			"redirect_uri":  "vikunja-flutter://callback",
			"code_verifier": codeVerifier,
		})
		assert.Equal(t, http.StatusPreconditionFailed, rec.Code, rec.Body.String())
		assert.NotContains(t, rec.Body.String(), "access_token")
	})
}
