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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Re-auth endpoints must not let a stolen session brute-force the password (GHSA-49pj-52x6-8gpj).
func TestReauthThrottle(t *testing.T) {
	const failureLimit = 10

	for _, tc := range []struct {
		name, method, path, wrong, right string
	}{
		{"v1 totp disable", http.MethodPost, "/api/v1/user/settings/totp/disable", `{"password":"wrong"}`, `{"password":"12345678"}`},
		{"v2 totp disable", http.MethodPost, "/api/v2/user/settings/totp/disable", `{"password":"wrong"}`, `{"password":"12345678"}`},
		{"v1 totp enroll", http.MethodPost, "/api/v1/user/settings/totp/enroll", `{"password":"wrong"}`, `{"password":"12345678"}`},
		{"v2 totp enroll", http.MethodPost, "/api/v2/user/settings/totp/enroll", `{"password":"wrong"}`, `{"password":"12345678"}`},
		{"v1 totp enable", http.MethodPost, "/api/v1/user/settings/totp/enable", `{"passcode":"000000","password":"wrong"}`, `{"passcode":"000000","password":"12345678"}`},
		{"v2 totp enable", http.MethodPost, "/api/v2/user/settings/totp/enable", `{"passcode":"000000","password":"wrong"}`, `{"passcode":"000000","password":"12345678"}`},
		{"v1 deletion request", http.MethodPost, "/api/v1/user/deletion/request", `{"password":"wrong"}`, `{"password":"12345678"}`},
		{"v2 deletion request", http.MethodPost, "/api/v2/user/deletion/request", `{"password":"wrong"}`, `{"password":"12345678"}`},
		{"v1 deletion cancel", http.MethodPost, "/api/v1/user/deletion/cancel", `{"password":"wrong"}`, `{"password":"12345678"}`},
		{"v2 deletion cancel", http.MethodPost, "/api/v2/user/deletion/cancel", `{"password":"wrong"}`, `{"password":"12345678"}`},
		{"v1 export request", http.MethodPost, "/api/v1/user/export/request", `{"password":"wrong"}`, `{"password":"12345678"}`},
		{"v2 export request", http.MethodPost, "/api/v2/user/export/request", `{"password":"wrong"}`, `{"password":"12345678"}`},
		{"v1 password change", http.MethodPost, "/api/v1/user/password", `{"old_password":"wrong","new_password":"newpassword1"}`, `{"old_password":"12345678","new_password":"newpassword1"}`},
		{"v2 password change", http.MethodPost, "/api/v2/user/password", `{"old_password":"wrong","new_password":"newpassword1"}`, `{"old_password":"12345678","new_password":"newpassword1"}`},
		{"v1 email change", http.MethodPost, "/api/v1/user/settings/email", `{"new_email":"new@example.com","password":"wrong"}`, `{"new_email":"new@example.com","password":"12345678"}`},
		{"v2 email change", http.MethodPut, "/api/v2/user/settings/email", `{"new_email":"new@example.com","password":"wrong"}`, `{"new_email":"new@example.com","password":"12345678"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, err := setupTestEnv()
			require.NoError(t, err)
			token := humaTokenFor(t, &testuser10)

			for i := 0; i < failureLimit; i++ {
				rec := humaRequest(t, e, tc.method, tc.path, tc.wrong, token, "")
				require.Equal(t, http.StatusForbidden, rec.Code, "attempt %d; body: %s", i, rec.Body.String())
			}

			rec := humaRequest(t, e, tc.method, tc.path, tc.right, token, "")
			assert.Equal(t, http.StatusTooManyRequests, rec.Code, "body: %s", rec.Body.String())
		})
	}

	t.Run("other users are not affected", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, &testuser10)
		for i := 0; i < failureLimit; i++ {
			humaRequest(t, e, http.MethodPost, "/api/v2/user/export/request", `{"password":"wrong"}`, token, "")
		}

		rec := humaRequest(t, e, http.MethodPost, "/api/v2/user/export/request", `{"password":"wrong"}`, humaTokenFor(t, &testuser1), "")
		assert.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
	})
}
