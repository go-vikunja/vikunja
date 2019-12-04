// Vikunja is a todo-list application to facilitate your life.
// Copyright 2019 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package integrations

import (
	"code.vikunja.io/api/pkg/models"
	apiv1 "code.vikunja.io/api/pkg/routes/api/v1"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestUserChangePassword(t *testing.T) {
	t.Run("Normal test", func(t *testing.T) {
		rec, err := newTestRequestWithUser(t, http.MethodPost, apiv1.UserChangePassword, &testuser1, `{
  "new_password": "12345",
  "old_password": "1234"
}`, nil, nil)
		assert.NoError(t, err)
		assert.Contains(t, rec.Body.String(), `The password was updated successfully.`)
	})
	t.Run("Wrong old password", func(t *testing.T) {
		_, err := newTestRequestWithUser(t, http.MethodPost, apiv1.UserChangePassword, &testuser1, `{
  "new_password": "12345",
  "old_password": "invalid"
}`, nil, nil)
		assert.Error(t, err)
		assertHandlerErrorCode(t, err, models.ErrCodeWrongUsernameOrPassword)
	})
	t.Run("Empty old password", func(t *testing.T) {
		_, err := newTestRequestWithUser(t, http.MethodPost, apiv1.UserChangePassword, &testuser1, `{
  "new_password": "12345",
  "old_password": ""
}`, nil, nil)
		assert.Error(t, err)
		assertHandlerErrorCode(t, err, models.ErrCodeEmptyOldPassword)
	})
	t.Run("Empty new password", func(t *testing.T) {
		_, err := newTestRequestWithUser(t, http.MethodPost, apiv1.UserChangePassword, &testuser1, `{
  "new_password": "",
  "old_password": "1234"
}`, nil, nil)
		assert.Error(t, err)
		assertHandlerErrorCode(t, err, models.ErrCodeEmptyNewPassword)
	})
}
