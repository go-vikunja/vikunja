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

	apiv1 "code.vikunja.io/api/pkg/routes/api/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserTOTPLocalUser(t *testing.T) {
	t.Run("Enroll TOTP for local user", func(t *testing.T) {
		rec, err := newTestRequestWithUser(t, http.MethodPost, apiv1.UserTOTPEnroll, &testuser1, "", nil, nil)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `"secret"`)
		assert.Contains(t, rec.Body.String(), `"url"`)
		assert.Contains(t, rec.Body.String(), `"enabled":false`)
	})

	t.Run("Get TOTP QR Code for enrolled local user", func(t *testing.T) {
		rec, err := newTestRequestWithUser(t, http.MethodGet, apiv1.UserTOTPQrCode, &testuser1, "", nil, nil)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "image/jpeg", rec.Header().Get("Content-Type"))
	})

	t.Run("Get TOTP settings for enrolled local user", func(t *testing.T) {
		rec, err := newTestRequestWithUser(t, http.MethodGet, apiv1.UserTOTP, &testuser1, "", nil, nil)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `"secret"`)
		assert.Contains(t, rec.Body.String(), `"enabled":false`)
	})
}
