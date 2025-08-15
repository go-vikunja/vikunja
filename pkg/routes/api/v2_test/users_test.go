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

package v2_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	v2models "code.vikunja.io/api/pkg/models/v2"
	"code.vikunja.io/api/pkg/routes"
	v2 "code.vikunja.io/api/pkg/routes/api/v2"
	"code.vikunja.io/api/pkg/user"
	"github.com/stretchr/testify/assert"
)

func TestGetUsers(t *testing.T) {
	db.NewTestDB()
	e := routes.NewEcho()

	// Create a user
	u, err := user.CreateUser(db.Get(), &user.User{
		Username: "test",
		Password: "password",
		Email:    "test@example.com",
	})
	assert.NoError(t, err)

	// Create another user
	_, err = user.CreateUser(db.Get(), &user.User{
		Username: "test2",
		Password: "password",
		Email:    "test2@example.com",
	})
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/v2/users", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", u)

	if assert.NoError(t, v2.GetUsers(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		var users []*v2models.User
		assert.NoError(t, rec.Result().Body.Close())
		assert.NoError(t, e.JSONSerializer.Deserialize(c, rec.Result().Body, &users))
		assert.Len(t, users, 2)
	}
}
