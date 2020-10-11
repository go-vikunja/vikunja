// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2020 Vikunja and contributors. All rights reserved.
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
	"net/http"
	"testing"

	apiv1 "code.vikunja.io/api/pkg/routes/api/v1"
	"code.vikunja.io/api/pkg/user"
	"github.com/stretchr/testify/assert"
)

func TestLogin(t *testing.T) {
	t.Run("Normal login", func(t *testing.T) {
		rec, err := newTestRequest(t, http.MethodPost, apiv1.Login, `{
  "username": "user1",
  "password": "1234"
}`)
		assert.NoError(t, err)
		assert.Contains(t, rec.Body.String(), "token")
	})
	t.Run("Empty payload", func(t *testing.T) {
		_, err := newTestRequest(t, http.MethodPost, apiv1.Login, `{}`)
		assert.Error(t, err)
		assertHandlerErrorCode(t, err, user.ErrCodeNoUsernamePassword)
	})
	t.Run("Not existing user", func(t *testing.T) {
		_, err := newTestRequest(t, http.MethodPost, apiv1.Login, `{
  "username": "userWichDoesNotExist",
  "password": "1234"
}`)
		assert.Error(t, err)
		assertHandlerErrorCode(t, err, user.ErrCodeWrongUsernameOrPassword)
	})
	t.Run("Wrong password", func(t *testing.T) {
		_, err := newTestRequest(t, http.MethodPost, apiv1.Login, `{
  "username": "user1",
  "password": "wrong"
}`)
		assert.Error(t, err)
		assertHandlerErrorCode(t, err, user.ErrCodeWrongUsernameOrPassword)
	})
	t.Run("user with unconfirmed email", func(t *testing.T) {
		_, err := newTestRequest(t, http.MethodPost, apiv1.Login, `{
  "username": "user5",
  "password": "1234"
}`)
		assert.Error(t, err)
		assertHandlerErrorCode(t, err, user.ErrCodeEmailNotConfirmed)
	})
}
