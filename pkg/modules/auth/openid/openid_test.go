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

package openid

import (
	"testing"

	"code.vikunja.io/api/pkg/db"
	"github.com/stretchr/testify/assert"
)

func TestGetOrCreateUser(t *testing.T) {
	t.Run("new user", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		cl := &claims{
			Email:             "test@example.com",
			PreferredUsername: "someUserWhoDoesNotExistYet",
		}
		u, err := getOrCreateUser(cl, "https://some.issuer", "12345")
		assert.NoError(t, err)
		db.AssertExists(t, "users", map[string]interface{}{
			"id":       u.ID,
			"email":    cl.Email,
			"username": "someUserWhoDoesNotExistYet",
		}, false)
	})
	t.Run("new user, no username provided", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		cl := &claims{
			Email:             "test@example.com",
			PreferredUsername: "",
		}
		u, err := getOrCreateUser(cl, "https://some.issuer", "12345")
		assert.NoError(t, err)
		assert.NotEmpty(t, u.Username)
		db.AssertExists(t, "users", map[string]interface{}{
			"id":    u.ID,
			"email": cl.Email,
		}, false)
	})
	t.Run("new user, no email address", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		cl := &claims{
			Email: "",
		}
		_, err := getOrCreateUser(cl, "https://some.issuer", "12345")
		assert.Error(t, err)
	})
	t.Run("existing user, different email address", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		cl := &claims{
			Email: "other-email-address@some.service.com",
		}
		u, err := getOrCreateUser(cl, "https://some.service.com", "12345")
		assert.NoError(t, err)
		db.AssertExists(t, "users", map[string]interface{}{
			"id":    u.ID,
			"email": cl.Email,
		}, false)
	})
}
