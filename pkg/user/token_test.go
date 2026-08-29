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

package user

import (
	"testing"

	"code.vikunja.io/api/pkg/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserTokenHashing(t *testing.T) {
	t.Run("stores only a hash", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		token, err := generateToken(s, &User{ID: 1}, TokenPasswordReset)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		assert.NotEmpty(t, token.ClearTextToken)
		assert.NotEqual(t, token.ClearTextToken, token.Token)
		assert.Regexp(t, "^[0-9a-f]{64}$", token.Token)

		db.AssertExists(t, "user_tokens", map[string]interface{}{
			"id":    token.ID,
			"token": token.Token,
		}, false)
		db.AssertMissing(t, "user_tokens", map[string]interface{}{
			"token": token.ClearTextToken,
		})
	})
	t.Run("raw token round-trips", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		created, err := generateToken(s, &User{ID: 1}, TokenPasswordReset)
		require.NoError(t, err)

		got, err := getToken(s, created.ClearTextToken, TokenPasswordReset)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, created.ID, got.ID)
	})
	t.Run("wrong token fails", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		got, err := getToken(s, "somethingelse", TokenPasswordReset)
		require.NoError(t, err)
		assert.Nil(t, got)
	})
}
