//   Vikunja is a todo-list application to facilitate your life.
//   Copyright 2019 Vikunja and contributors. All rights reserved.
//
//   This program is free software: you can redistribute it and/or modify
//   it under the terms of the GNU General Public License as published by
//   the Free Software Foundation, either version 3 of the License, or
//   (at your option) any later version.
//
//   This program is distributed in the hope that it will be useful,
//   but WITHOUT ANY WARRANTY; without even the implied warranty of
//   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//   GNU General Public License for more details.
//
//   You should have received a copy of the GNU General Public License
//   along with this program.  If not, see <https://www.gnu.org/licenses/>.

package integrations

import (
	"code.vikunja.io/api/pkg/models"
	apiv1 "code.vikunja.io/api/pkg/routes/api/v1"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestUserConfirmEmail(t *testing.T) {
	t.Run("Normal test", func(t *testing.T) {
		rec, err := newTestRequest(t, http.MethodPost, apiv1.UserConfirmEmail, `{"token": "tiepiQueed8ahc7zeeFe1eveiy4Ein8osooxegiephauph2Ael"}`)
		assert.NoError(t, err)
		assert.Contains(t, rec.Body.String(), `The email was confirmed successfully.`)
	})
	t.Run("Empty payload", func(t *testing.T) {
		_, err := newTestRequest(t, http.MethodPost, apiv1.UserConfirmEmail, `{}`)
		assert.Error(t, err)
		assert.Equal(t, http.StatusPreconditionFailed, err.(*echo.HTTPError).Code)
		assertHandlerErrorCode(t, err, models.ErrCodeInvalidEmailConfirmToken)
	})
	t.Run("Empty token", func(t *testing.T) {
		_, err := newTestRequest(t, http.MethodPost, apiv1.UserConfirmEmail, `{"token": ""}`)
		assert.Error(t, err)
		assertHandlerErrorCode(t, err, models.ErrCodeInvalidEmailConfirmToken)
	})
	t.Run("Invalid token", func(t *testing.T) {
		_, err := newTestRequest(t, http.MethodPost, apiv1.UserConfirmEmail, `{"token": "invalidToken"}`)
		assert.Error(t, err)
		assertHandlerErrorCode(t, err, models.ErrCodeInvalidEmailConfirmToken)
	})
}
