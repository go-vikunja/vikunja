// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2021 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/notifications"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
)

func TestDeleteUser(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		notifications.Fake()

		u := &user.User{ID: 6}
		err := DeleteUser(s, u)

		assert.NoError(t, err)
		db.AssertMissing(t, "users", map[string]interface{}{"id": u.ID})
		db.AssertMissing(t, "lists", map[string]interface{}{"id": 24}) // only user6 had access to this list
		db.AssertExists(t, "lists", map[string]interface{}{"id": 6}, false)
		db.AssertExists(t, "lists", map[string]interface{}{"id": 7}, false)
		db.AssertExists(t, "lists", map[string]interface{}{"id": 8}, false)
		db.AssertExists(t, "lists", map[string]interface{}{"id": 9}, false)
		db.AssertExists(t, "lists", map[string]interface{}{"id": 10}, false)
		db.AssertExists(t, "lists", map[string]interface{}{"id": 11}, false)
	})
	t.Run("user with no namespaces", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		notifications.Fake()

		u := &user.User{ID: 4}
		err := DeleteUser(s, u)

		assert.NoError(t, err)
		// No assertions for deleted lists and namespaces since that user doesn't have any
	})
}
