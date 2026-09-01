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

	t.Run("Get qr code for enrolled user", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/user/settings/totp/qrcode", "", humaTokenFor(t, &testuser1), "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Equal(t, "image/jpeg", rec.Header().Get("Content-Type"))
		assert.NotEmpty(t, rec.Body.Bytes(), "the qr code jpeg must have bytes")
	})

	t.Run("Enroll a fresh user", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
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
			{http.MethodGet, "/api/v2/user/settings/totp/qrcode", ""},
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

// Guards enabled provisioning secrets on v2 (GHSA-88f6-4rjv-x774).
func TestHumaTOTPSecretHiddenWhenEnabled(t *testing.T) {
	t.Run("Get settings hides secret and url once enabled", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/user/settings/totp", "", humaTokenFor(t, &testuser10), "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"enabled":true`)
		assert.NotContains(t, rec.Body.String(), `JBSWY3DPEHPK3PXP`, "the secret must not be disclosed once enabled")
		assert.NotContains(t, rec.Body.String(), `otpauth://`, "the url must not be disclosed once enabled")
	})

	t.Run("Get qrcode is refused once enabled", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/user/settings/totp/qrcode", "", humaTokenFor(t, &testuser10), "")
		assert.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
	})
}
