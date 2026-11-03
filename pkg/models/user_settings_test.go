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

package models

import (
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChangeUserPassword(t *testing.T) {
	t.Run("verifies the password of the authenticated account", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		require.NoError(t, user.UpdateUserPassword(s, &user.User{ID: 2}, "user2password"))

		// The username points at another account, e.g. a stale token claim.
		err := ChangeUserPassword(s, &user.User{ID: 1, Username: "user2"}, "user2password", "newpassword")
		require.Error(t, err)
		assert.True(t, user.IsErrWrongUsernameOrPassword(err))
	})
	t.Run("changes the password", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		require.NoError(t, ChangeUserPassword(s, &user.User{ID: 1}, "12345678", "newpassword"))
		require.NoError(t, s.Commit())

		u, err := user.GetUserByID(s, 1)
		require.NoError(t, err)
		require.NoError(t, user.CheckUserPassword(u, "newpassword"))
		db.AssertMissing(t, "sessions", map[string]any{"user_id": 1})
	})
}
