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
	"time"

	"code.vikunja.io/api/pkg/db"
	apiv1 "code.vikunja.io/api/pkg/routes/api/v1"
	"code.vikunja.io/api/pkg/user"

	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserTOTPLocalUser(t *testing.T) {
	t.Run("Enroll TOTP for local user", func(t *testing.T) {
		// Use testuser15 who has no TOTP enrollment in fixtures
		rec, err := newTestRequestWithUser(t, http.MethodPost, apiv1.UserTOTPEnroll, &testuser15, "", nil, nil)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `"secret"`)
		assert.Contains(t, rec.Body.String(), `"url"`)
		assert.Contains(t, rec.Body.String(), `"enabled":false`)
	})

	t.Run("Get TOTP QR Code for enrolled local user", func(t *testing.T) {
		// user1 has TOTP enrolled (but not enabled) via fixtures
		rec, err := newTestRequestWithUser(t, http.MethodGet, apiv1.UserTOTPQrCode, &testuser1, "", nil, nil)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "image/jpeg", rec.Header().Get("Content-Type"))
	})

	t.Run("Get TOTP settings for enrolled local user", func(t *testing.T) {
		// user1 has TOTP enrolled (but not enabled) via fixtures
		rec, err := newTestRequestWithUser(t, http.MethodGet, apiv1.UserTOTP, &testuser1, "", nil, nil)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `"secret"`)
		assert.Contains(t, rec.Body.String(), `"enabled":false`)
	})
}

// Guards enabled provisioning secrets (GHSA-88f6-4rjv-x774).
func TestUserTOTPSecretHiddenWhenEnabled(t *testing.T) {
	t.Run("Get TOTP settings hides secret and url once enabled", func(t *testing.T) {
		rec, err := newTestRequestWithUser(t, http.MethodGet, apiv1.UserTOTP, &testuser10, "", nil, nil)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `"enabled":true`)
		assert.NotContains(t, rec.Body.String(), `JBSWY3DPEHPK3PXP`, "the secret must not be disclosed once enabled")
		assert.NotContains(t, rec.Body.String(), `otpauth://`, "the url must not be disclosed once enabled")
	})

	t.Run("Get TOTP qrcode is refused once enabled", func(t *testing.T) {
		_, err := newTestRequestWithUser(t, http.MethodGet, apiv1.UserTOTPQrCode, &testuser10, "", nil, nil)
		require.Error(t, err)
		assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
	})

	t.Run("Enabled TOTP login validation still works with the full secret", func(t *testing.T) {
		// The login flow needs the secret even though it is no longer exposed.
		passcode, err := totp.GenerateCode("JBSWY3DPEHPK3PXP", time.Now())
		require.NoError(t, err)
		s := db.NewSession()
		defer s.Close()
		tt, err := user.GetTOTPForUser(s, &user.User{ID: 10})
		require.NoError(t, err)
		assert.Equal(t, "JBSWY3DPEHPK3PXP", tt.Secret, "GetTOTPForUser must keep returning the full object")
		_, err = user.ValidateTOTPPasscode(s, &user.TOTPPasscode{User: &user.User{ID: 10}, Passcode: passcode})
		require.NoError(t, err)
	})
}
