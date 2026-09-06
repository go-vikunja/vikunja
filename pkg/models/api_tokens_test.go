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
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/modules/keyvalue"
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
	assert.Len(t, tokens, 3)
	assert.Len(t, tokens, count)
	assert.Equal(t, int64(3), total)
	assert.Equal(t, int64(1), tokens[0].ID)
	assert.Equal(t, int64(2), tokens[1].ID)
	assert.Equal(t, int64(9), tokens[2].ID)
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

// nonUserAuth is a web.Auth that is neither *user.User nor *models.LinkSharing.
// It proves the API-token guard rejects by principal type, not by matching the
// concrete link-share struct (GHSA-vvcv-vpph-h844).
type nonUserAuth struct {
	id int64
}

func (a *nonUserAuth) GetID() int64 { return a.id }

func TestAPIToken_RejectsNonUserPrincipal(t *testing.T) {
	// ID 2 collides with user 2, who owns token 3.
	a := &nonUserAuth{id: 2}

	t.Run("CanCreate", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()
		db.LoadAndAssertFixtures(t)

		can, err := (&APIToken{}).CanCreate(s, a)
		require.Error(t, err)
		assert.False(t, can)
	})
	t.Run("CanDelete", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()
		db.LoadAndAssertFixtures(t)

		can, err := (&APIToken{ID: 3}).CanDelete(s, a)
		require.Error(t, err)
		assert.False(t, can)

		exists, err := s.Where("id = ?", 3).Exist(&APIToken{})
		require.NoError(t, err)
		assert.True(t, exists, "token must be retained")
	})
	t.Run("Create", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()
		db.LoadAndAssertFixtures(t)

		err := (&APIToken{}).Create(s, a)
		require.Error(t, err)

		exists, err := s.Where("owner_id = ?", 2).Count(&APIToken{})
		require.NoError(t, err)
		assert.Equal(t, int64(1), exists, "no token must be created for the colliding id")
	})
	t.Run("ReadAll", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()
		db.LoadAndAssertFixtures(t)

		_, _, _, err := (&APIToken{}).ReadAll(s, a, "", 1, 50)
		require.Error(t, err)
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
	// Fixture values from pkg/db/fixtures/api_tokens.yml
	const (
		token1Hash = "a1813a558185d99f5197d2d549e4dd91292376aa00210229d70f77b57e165f6613fd12c1f790aa6493548cb9bceff33b45b4"
		token3Hash = "da4b9c3aa72633274c37ab3419fbfbe4c5b79310b76027ac36f85e4c5ad0c2342a1d9e1c9b72ca07ec0a66ad2ee3505539af"
	)

	t.Run("valid token", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()
		db.LoadAndAssertFixtures(t)
		const raw = "tk_2eef46f40ebab3304919ab2e7e39993f75f29d2e" // Token 1
		key := verifiedAPITokenKey(raw)
		t.Cleanup(func() { _ = keyvalue.Del(key) })

		token, err := GetTokenFromTokenString(s, raw)

		require.NoError(t, err)
		assert.Equal(t, int64(1), token.ID)
	})
	t.Run("verified token is served from the cache until it is deleted", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()
		db.LoadAndAssertFixtures(t)
		const raw = "tk_2eef46f40ebab3304919ab2e7e39993f75f29d2e" // Token 1
		key := verifiedAPITokenKey(raw)
		t.Cleanup(func() { _ = keyvalue.Del(key) })
		require.NoError(t, keyvalue.Del(key))

		_, err := GetTokenFromTokenString(s, raw)
		require.NoError(t, err)
		var v verifiedAPIToken
		cached, err := keyvalue.GetWithValue(key, &v)
		require.NoError(t, err)
		assert.True(t, cached)
		assert.Equal(t, token1Hash, v.Hash)

		token, err := GetTokenFromTokenString(s, raw)
		require.NoError(t, err)
		assert.Equal(t, int64(1), token.ID)

		_, err = s.Where("id = ?", 1).Delete(&APIToken{})
		require.NoError(t, err)
		_, err = GetTokenFromTokenString(s, raw)
		require.Error(t, err)
		assert.True(t, IsErrAPITokenInvalid(err))
		cached, err = keyvalue.GetWithValue(key, &v)
		require.NoError(t, err)
		assert.False(t, cached)
	})
	t.Run("cache hit is served without pbkdf2", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()
		db.LoadAndAssertFixtures(t)
		const raw = "tk_2eef46f40ebab3304919ab2e7e39993f75f29d2e" // Token 1
		key := verifiedAPITokenKey(raw)
		t.Cleanup(func() { _ = keyvalue.Del(key) })
		// Only the cache hit can return token 3 here; hashing the raw string would find token 1.
		require.NoError(t, keyvalue.PutWithTTL(key, verifiedAPIToken{Hash: token3Hash, Tag: verifiedAPITokenTag(raw, token3Hash)}, verifiedAPITokenTTL))

		token, err := GetTokenFromTokenString(s, raw)
		require.NoError(t, err)
		assert.Equal(t, int64(3), token.ID)
	})
	t.Run("cache hit returns the current row, not a cached copy", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()
		db.LoadAndAssertFixtures(t)
		const raw = "tk_2eef46f40ebab3304919ab2e7e39993f75f29d2e" // Token 1
		key := verifiedAPITokenKey(raw)
		t.Cleanup(func() { _ = keyvalue.Del(key) })
		require.NoError(t, keyvalue.Del(key))

		_, err := GetTokenFromTokenString(s, raw)
		require.NoError(t, err)

		updated := APIPermissions{"tasks": {"read_one"}}
		_, err = s.Where("id = ?", 1).Cols("permissions").Update(&APIToken{APIPermissions: updated})
		require.NoError(t, err)

		token, err := GetTokenFromTokenString(s, raw)
		require.NoError(t, err)
		assert.Equal(t, updated, token.APIPermissions)
	})
	t.Run("tampered or stale cache entry falls back to the full lookup", func(t *testing.T) {
		const raw = "tk_2eef46f40ebab3304919ab2e7e39993f75f29d2e" // Token 1
		otherToken := "tk_" + strings.Repeat("0", 40)
		for _, tc := range []struct {
			name  string
			entry verifiedAPIToken
		}{
			{"copied from another token", verifiedAPIToken{Hash: token3Hash, Tag: verifiedAPITokenTag(otherToken, token3Hash)}},
			{"stale hash", verifiedAPIToken{Hash: "not a hash", Tag: verifiedAPITokenTag(raw, "not a hash")}},
		} {
			t.Run(tc.name, func(t *testing.T) {
				s := db.NewSession()
				defer s.Close()
				db.LoadAndAssertFixtures(t)
				key := verifiedAPITokenKey(raw)
				t.Cleanup(func() { _ = keyvalue.Del(key) })
				require.NoError(t, keyvalue.PutWithTTL(key, tc.entry, verifiedAPITokenTTL))

				token, err := GetTokenFromTokenString(s, raw)
				require.NoError(t, err)
				assert.Equal(t, int64(1), token.ID)
				var v verifiedAPIToken
				cached, err := keyvalue.GetWithValue(key, &v)
				require.NoError(t, err)
				assert.True(t, cached)
				assert.Equal(t, token1Hash, v.Hash)
			})
		}
	})
	t.Run("key derivation cannot forge a tag", func(t *testing.T) {
		key := strings.TrimPrefix(verifiedAPITokenKey("A|B"), "api_token_verified_")
		assert.NotEqual(t, verifiedAPITokenTag("A", "B"), key)
	})
	t.Run("invalid token", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()
		db.LoadAndAssertFixtures(t)

		_, err := GetTokenFromTokenString(s, "tk_loremipsum")

		require.Error(t, err)
		assert.True(t, IsErrAPITokenInvalid(err))
	})
	t.Run("token shorter than prefix+8 does not panic", func(t *testing.T) {
		for _, short := range []string{"", "tk_", "tk_a", "tk_abc", "tk_1234567"} {
			s := db.NewSession()
			db.LoadAndAssertFixtures(t)

			token, err := GetTokenFromTokenString(s, short)

			require.Errorf(t, err, "short token %q must be rejected", short)
			assert.True(t, IsErrAPITokenInvalid(err))
			assert.Nil(t, token)
			s.Close()
		}
	})
}
