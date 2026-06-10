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

	"code.vikunja.io/api/pkg/user"

	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testuser14 is a non-local (OIDC) account; totp is local-only, so every totp
// route must refuse it. See pkg/db/fixtures/users.yml.
var testuser14 = user.User{ID: 14, Username: "user14", Issuer: "https://some.service.com"}

// TestHumaTOTP mirrors v1's TestUserTOTPLocalUser and adds the enable/disable
// flows plus the local-account-only guard. The QR-code endpoint is not ported
// to v2 (binary streaming, later wave), so there is no test for it here.
//
// Fixture topology (pkg/db/fixtures/totp.yml + users.yml):
//   - user1:  totp enrolled, not enabled (secret HXDMVJEC…).
//   - user10: totp enabled (secret JBSWY3DP…), local, password 12345678.
//   - user15: local, no totp enrollment.
//   - user14: non-local (OIDC) account.
func TestHumaTOTP(t *testing.T) {
	t.Run("Get status for enrolled user", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/user/settings/totp", "", humaTokenFor(t, &testuser1), "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"secret"`)
		assert.Contains(t, rec.Body.String(), `"enabled":false`)
	})

	t.Run("Get status without enrollment returns 412", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/user/settings/totp", "", humaTokenFor(t, &testuser15), "")
		require.Equal(t, http.StatusPreconditionFailed, rec.Code, "body: %s", rec.Body.String())
	})

	t.Run("Enroll a fresh user", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		// user15 has no totp enrollment in the fixtures.
		rec := humaRequest(t, e, http.MethodPost, "/api/v2/user/settings/totp/enroll", "", humaTokenFor(t, &testuser15), "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"secret"`)
		assert.Contains(t, rec.Body.String(), `"url"`)
		assert.Contains(t, rec.Body.String(), `"enabled":false`)
	})

	t.Run("Enroll when already enrolled returns 412", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		rec := humaRequest(t, e, http.MethodPost, "/api/v2/user/settings/totp/enroll", "", humaTokenFor(t, &testuser1), "")
		require.Equal(t, http.StatusPreconditionFailed, rec.Code, "body: %s", rec.Body.String())
	})

	t.Run("Enable with a valid passcode", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		// user1's fixture secret; generate a passcode that is valid right now.
		passcode, err := totp.GenerateCode("HXDMVJECJJWSRB3HWIZR4IFUGFTMXBOZ", time.Now())
		require.NoError(t, err)
		rec := humaRequest(t, e, http.MethodPost, "/api/v2/user/settings/totp/enable",
			fmt.Sprintf(`{"passcode":%q}`, passcode), humaTokenFor(t, &testuser1), "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), "enabled successfully")
	})

	t.Run("Enable with an invalid passcode returns 412", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		rec := humaRequest(t, e, http.MethodPost, "/api/v2/user/settings/totp/enable",
			`{"passcode":"000000"}`, humaTokenFor(t, &testuser1), "")
		require.Equal(t, http.StatusPreconditionFailed, rec.Code, "body: %s", rec.Body.String())
	})

	t.Run("Disable with the correct password", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		// user10 has totp enabled; 12345678 is their fixture password.
		rec := humaRequest(t, e, http.MethodPost, "/api/v2/user/settings/totp/disable",
			`{"password":"12345678"}`, humaTokenFor(t, &testuser10), "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), "disabled successfully")
	})

	t.Run("Disable with a wrong password is refused", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		rec := humaRequest(t, e, http.MethodPost, "/api/v2/user/settings/totp/disable",
			`{"password":"wrong-password"}`, humaTokenFor(t, &testuser10), "")
		require.NotEqual(t, http.StatusOK, rec.Code, "wrong password must not disable totp; body: %s", rec.Body.String())
	})

	t.Run("Non-local user is refused on every route", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, &testuser14)
		for _, tc := range []struct {
			method, path, body string
		}{
			{http.MethodGet, "/api/v2/user/settings/totp", ""},
			{http.MethodPost, "/api/v2/user/settings/totp/enroll", ""},
			{http.MethodPost, "/api/v2/user/settings/totp/enable", `{"passcode":"000000"}`},
			{http.MethodPost, "/api/v2/user/settings/totp/disable", `{"password":"12345678"}`},
		} {
			rec := humaRequest(t, e, tc.method, tc.path, tc.body, token, "")
			assert.Equal(t, http.StatusPreconditionFailed, rec.Code,
				"%s %s must refuse a non-local account; body: %s", tc.method, tc.path, rec.Body.String())
		}
	})
}
