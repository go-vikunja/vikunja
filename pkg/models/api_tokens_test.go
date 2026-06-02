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

func TestAPIToken_ReadAll(t *testing.T) {
	u := &user.User{ID: 1}
	token := &APIToken{}
	s := db.NewSession()
	defer s.Close()
	db.LoadAndAssertFixtures(t)

	// Checking if the user only sees their own tokens

	result, count, total, err := token.ReadAll(s, u, "", 1, 50)
	require.NoError(t, err)
	tokens, is := result.([]*APIToken)
	assert.Truef(t, is, "tokens are not of type []*APIToken")
	assert.Len(t, tokens, 2)
	assert.Len(t, tokens, count)
	assert.Equal(t, int64(2), total)
	assert.Equal(t, int64(1), tokens[0].ID)
	assert.Equal(t, int64(2), tokens[1].ID)
}

func TestAPIToken_CanDelete(t *testing.T) {
	t.Run("own token", func(t *testing.T) {
		u := &user.User{ID: 1}
		token := &APIToken{ID: 1}
		s := db.NewSession()
		defer s.Close()
		db.LoadAndAssertFixtures(t)

		can, err := token.CanDelete(s, u)
		require.NoError(t, err)
		assert.True(t, can)
	})
	t.Run("noneixsting token", func(t *testing.T) {
		u := &user.User{ID: 1}
		token := &APIToken{ID: 999}
		s := db.NewSession()
		defer s.Close()
		db.LoadAndAssertFixtures(t)

		can, err := token.CanDelete(s, u)
		require.NoError(t, err)
		assert.False(t, can)
	})
	t.Run("token of another user", func(t *testing.T) {
		u := &user.User{ID: 2}
		token := &APIToken{ID: 1}
		s := db.NewSession()
		defer s.Close()
		db.LoadAndAssertFixtures(t)

		can, err := token.CanDelete(s, u)
		require.NoError(t, err)
		assert.False(t, can)
	})
}

func TestAPIToken_Create(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		u := &user.User{ID: 1}
		token := &APIToken{}
		s := db.NewSession()
		defer s.Close()
		db.LoadAndAssertFixtures(t)

		err := token.Create(s, u)
		require.NoError(t, err)
	})
}

func TestAPIToken_HasCaldavAccess(t *testing.T) {
	t.Run("has caldav access", func(t *testing.T) {
		token := &APIToken{
			APIPermissions: APIPermissions{"caldav": {"access"}},
		}
		assert.True(t, token.HasCaldavAccess())
	})
	t.Run("no caldav group", func(t *testing.T) {
		token := &APIToken{
			APIPermissions: APIPermissions{"tasks": {"read_all"}},
		}
		assert.False(t, token.HasCaldavAccess())
	})
	t.Run("caldav group but wrong permission", func(t *testing.T) {
		token := &APIToken{
			APIPermissions: APIPermissions{"caldav": {"read_all"}},
		}
		assert.False(t, token.HasCaldavAccess())
	})
	t.Run("caldav access among other permissions", func(t *testing.T) {
		token := &APIToken{
			APIPermissions: APIPermissions{
				"tasks":  {"read_all", "update"},
				"caldav": {"access"},
			},
		}
		assert.True(t, token.HasCaldavAccess())
	})
}

func TestAPIToken_HasFeedsAccess(t *testing.T) {
	t.Run("has feeds access", func(t *testing.T) {
		token := &APIToken{
			APIPermissions: APIPermissions{"feeds": {"access"}},
		}
		assert.True(t, token.HasFeedsAccess())
	})
	t.Run("no feeds group", func(t *testing.T) {
		token := &APIToken{
			APIPermissions: APIPermissions{"tasks": {"read_all"}},
		}
		assert.False(t, token.HasFeedsAccess())
	})
	t.Run("feeds group but wrong permission", func(t *testing.T) {
		token := &APIToken{
			APIPermissions: APIPermissions{"feeds": {"read_all"}},
		}
		assert.False(t, token.HasFeedsAccess())
	})
	t.Run("feeds access among other permissions", func(t *testing.T) {
		token := &APIToken{
			APIPermissions: APIPermissions{
				"tasks": {"read_all", "update"},
				"feeds": {"access"},
			},
		}
		assert.True(t, token.HasFeedsAccess())
	})
}

func TestAPIToken_GetTokenFromTokenString(t *testing.T) {
	t.Run("valid token", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()
		db.LoadAndAssertFixtures(t)

		token, err := GetTokenFromTokenString(s, "tk_2eef46f40ebab3304919ab2e7e39993f75f29d2e") // Token 1

		require.NoError(t, err)
		assert.Equal(t, int64(1), token.ID)
	})
	t.Run("invalid token", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()
		db.LoadAndAssertFixtures(t)

		_, err := GetTokenFromTokenString(s, "tk_loremipsum")

		require.Error(t, err)
		assert.True(t, IsErrAPITokenInvalid(err))
	})
}
