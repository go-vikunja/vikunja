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
	"code.vikunja.io/api/pkg/user"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRequestResetPasswordToken(t *testing.T) {
	t.Run("Normal requesting a password reset token", func(t *testing.T) {
		rec, err := newTestRequest(t, http.MethodPost, apiv1.UserRequestResetPasswordToken, `{"email": "user1@example.com"}`, nil, nil)
		require.NoError(t, err)
		assert.Contains(t, rec.Body.String(), `Token was sent.`)
	})
	t.Run("Empty payload", func(t *testing.T) {
		_, err := newTestRequest(t, http.MethodPost, apiv1.UserRequestResetPasswordToken, `{}`, nil, nil)
		require.Error(t, err)
		assertHandlerErrorCode(t, err, user.ErrCodeNoUsernamePassword)
	})
	t.Run("Invalid email address", func(t *testing.T) {
		_, err := newTestRequest(t, http.MethodPost, apiv1.UserRequestResetPasswordToken, `{"email": "user1example.com"}`, nil, nil)
		require.Error(t, err)
		assert.Equal(t, http.StatusBadRequest, err.(*echo.HTTPError).Code)
	})
	t.Run("No user with that email address", func(t *testing.T) {
		_, err := newTestRequest(t, http.MethodPost, apiv1.UserRequestResetPasswordToken, `{"email": "user1000@example.com"}`, nil, nil)
		require.Error(t, err)
		assertHandlerErrorCode(t, err, user.ErrCodeUserDoesNotExist)
	})
}
