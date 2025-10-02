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

package services

import (
	"testing"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPITokenService_Create(t *testing.T) {

	t.Run("create token successfully", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewAPITokenService(testEngine)
		u := &user.User{ID: 1}

		token := &models.APIToken{
			Title: "Test Token",
			APIPermissions: models.APIPermissions{
				"tasks": []string{"read_all"},
			},
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}

		err := service.Create(s, token, u)
		require.NoError(t, err)
		assert.NotZero(t, token.ID)
		assert.NotEmpty(t, token.Token)
		assert.True(t, len(token.Token) > 8)
		assert.NotEmpty(t, token.TokenSalt)
		assert.NotEmpty(t, token.TokenHash)
		assert.NotEmpty(t, token.TokenLastEight)
		assert.Equal(t, int64(1), token.OwnerID)
		assert.Contains(t, token.Token, models.APITokenPrefix)
	})

	t.Run("create token with nil user", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewAPITokenService(testEngine)

		token := &models.APIToken{
			Title: "Test Token",
			APIPermissions: models.APIPermissions{
				"tasks": []string{"read_all"},
			},
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}

		err := service.Create(s, token, nil)
		require.Error(t, err)
		assert.Equal(t, ErrAccessDenied, err)
	})

	t.Run("create token with invalid permissions", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewAPITokenService(testEngine)
		u := &user.User{ID: 1}

		token := &models.APIToken{
			Title: "Test Token",
			APIPermissions: models.APIPermissions{
				"invalid_group": []string{"invalid_permission"},
			},
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}

		err := service.Create(s, token, u)
		require.Error(t, err)
	})

	t.Run("token ID is reset on create", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewAPITokenService(testEngine)
		u := &user.User{ID: 1}

		token := &models.APIToken{
			ID:    999, // Should be reset
			Title: "Test Token",
			APIPermissions: models.APIPermissions{
				"tasks": []string{"read_all"},
			},
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}

		err := service.Create(s, token, u)
		require.NoError(t, err)
		assert.NotEqual(t, int64(999), token.ID)
	})
}

func TestAPITokenService_Get(t *testing.T) {

	t.Run("get own token", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewAPITokenService(testEngine)
		u := &user.User{ID: 1}

		token, err := service.Get(s, 1, u)
		require.NoError(t, err)
		assert.Equal(t, int64(1), token.ID)
		assert.Equal(t, "Token 1", token.Title)
		assert.Equal(t, int64(1), token.OwnerID)
	})

	t.Run("get token of another user", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewAPITokenService(testEngine)
		u := &user.User{ID: 2}

		_, err := service.Get(s, 1, u) // Token 1 belongs to user 1
		require.Error(t, err)
		assert.True(t, models.IsErrAPITokenDoesNotExist(err))
	})

	t.Run("get nonexistent token", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewAPITokenService(testEngine)
		u := &user.User{ID: 1}

		_, err := service.Get(s, 999, u)
		require.Error(t, err)
		assert.True(t, models.IsErrAPITokenDoesNotExist(err))
	})

	t.Run("get with nil user", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewAPITokenService(testEngine)

		_, err := service.Get(s, 1, nil)
		require.Error(t, err)
		assert.Equal(t, ErrAccessDenied, err)
	})
}

func TestAPITokenService_GetAll(t *testing.T) {

	t.Run("get all tokens for user", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewAPITokenService(testEngine)
		u := &user.User{ID: 1}

		tokens, count, total, err := service.GetAll(s, u, "", 1, 50)
		require.NoError(t, err)
		assert.Len(t, tokens, 2)
		assert.Equal(t, 2, count)
		assert.Equal(t, int64(2), total)
		assert.Equal(t, int64(1), tokens[0].ID)
		assert.Equal(t, int64(2), tokens[1].ID)
	})

	t.Run("user only sees their own tokens", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewAPITokenService(testEngine)
		u := &user.User{ID: 2}

		tokens, count, total, err := service.GetAll(s, u, "", 1, 50)
		require.NoError(t, err)
		assert.Len(t, tokens, 0)
		assert.Equal(t, 0, count)
		assert.Equal(t, int64(0), total)
	})

	t.Run("search tokens by title", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewAPITokenService(testEngine)
		u := &user.User{ID: 1}

		tokens, count, total, err := service.GetAll(s, u, "Token 1", 1, 50)
		require.NoError(t, err)
		assert.Len(t, tokens, 1)
		assert.Equal(t, 1, count)
		assert.Equal(t, int64(1), total)
		assert.Equal(t, "Token 1", tokens[0].Title)
	})

	t.Run("pagination", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewAPITokenService(testEngine)
		u := &user.User{ID: 1}

		// Get first page with 1 item per page
		tokens, count, total, err := service.GetAll(s, u, "", 1, 1)
		require.NoError(t, err)
		assert.Len(t, tokens, 1)
		assert.Equal(t, 1, count)
		assert.Equal(t, int64(2), total)
		assert.Equal(t, int64(1), tokens[0].ID)

		// Get second page
		tokens, count, total, err = service.GetAll(s, u, "", 2, 1)
		require.NoError(t, err)
		assert.Len(t, tokens, 1)
		assert.Equal(t, 1, count)
		assert.Equal(t, int64(2), total)
		assert.Equal(t, int64(2), tokens[0].ID)
	})

	t.Run("get all with nil user", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewAPITokenService(testEngine)

		_, _, _, err := service.GetAll(s, nil, "", 1, 50)
		require.Error(t, err)
		assert.Equal(t, ErrAccessDenied, err)
	})
}

func TestAPITokenService_Delete(t *testing.T) {

	t.Run("delete own token", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewAPITokenService(testEngine)
		u := &user.User{ID: 1}

		err := service.Delete(s, 1, u)
		require.NoError(t, err)

		// Verify token is deleted
		_, err = service.Get(s, 1, u)
		require.Error(t, err)
		assert.True(t, models.IsErrAPITokenDoesNotExist(err))
	})

	t.Run("delete token of another user", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewAPITokenService(testEngine)
		u := &user.User{ID: 2}

		err := service.Delete(s, 1, u) // Token 1 belongs to user 1
		require.Error(t, err)
		assert.True(t, models.IsErrAPITokenDoesNotExist(err))
	})

	t.Run("delete nonexistent token", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewAPITokenService(testEngine)
		u := &user.User{ID: 1}

		err := service.Delete(s, 999, u)
		require.Error(t, err)
		assert.True(t, models.IsErrAPITokenDoesNotExist(err))
	})

	t.Run("delete with nil user", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewAPITokenService(testEngine)

		err := service.Delete(s, 1, nil)
		require.Error(t, err)
		assert.Equal(t, ErrAccessDenied, err)
	})
}

func TestAPITokenService_GetTokenFromTokenString(t *testing.T) {

	t.Run("valid token string", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewAPITokenService(testEngine)

		token, err := service.GetTokenFromTokenString(s, "tk_2eef46f40ebab3304919ab2e7e39993f75f29d2e")
		require.NoError(t, err)
		assert.Equal(t, int64(1), token.ID)
		assert.Equal(t, "Token 1", token.Title)
	})

	t.Run("invalid token string", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewAPITokenService(testEngine)

		_, err := service.GetTokenFromTokenString(s, "tk_invalidtoken")
		require.Error(t, err)
		assert.True(t, models.IsErrAPITokenInvalid(err))
	})

	t.Run("token string with wrong hash", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewAPITokenService(testEngine)

		// Use correct last 8 chars but wrong prefix
		_, err := service.GetTokenFromTokenString(s, "tk_wronghash0000000000000075f29d2e")
		require.Error(t, err)
		assert.True(t, models.IsErrAPITokenInvalid(err))
	})
}

func TestAPITokenService_ValidateToken(t *testing.T) {

	t.Run("validate valid token", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewAPITokenService(testEngine)

		token, u, err := service.ValidateToken(s, "tk_2eef46f40ebab3304919ab2e7e39993f75f29d2e", "/api/v1/tasks", "GET")
		require.NoError(t, err)
		assert.Equal(t, int64(1), token.ID)
		assert.Equal(t, int64(1), u.ID)
		assert.Equal(t, "user1", u.Username)
	})

	t.Run("validate expired token", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewAPITokenService(testEngine)

		// Token 2 is expired (see fixtures)
		_, _, err := service.ValidateToken(s, "tk_a5e6f92ddbad68f49ee2c63e52174db0235008c8", "/api/v1/tasks", "GET")
		require.Error(t, err)
		assert.True(t, models.IsErrAPITokenExpired(err))
	})

	t.Run("validate invalid token", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewAPITokenService(testEngine)

		_, _, err := service.ValidateToken(s, "tk_invalidtoken", "/api/v1/tasks", "GET")
		require.Error(t, err)
		assert.True(t, models.IsErrAPITokenInvalid(err))
	})
}

func TestAPITokenService_CanDelete(t *testing.T) {

	t.Run("can delete own token", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewAPITokenService(testEngine)
		u := &user.User{ID: 1}

		can, err := service.CanDelete(s, 1, u)
		require.NoError(t, err)
		assert.True(t, can)
	})

	t.Run("cannot delete token of another user", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewAPITokenService(testEngine)
		u := &user.User{ID: 2}

		can, err := service.CanDelete(s, 1, u)
		require.NoError(t, err)
		assert.False(t, can)
	})

	t.Run("cannot delete nonexistent token", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewAPITokenService(testEngine)
		u := &user.User{ID: 1}

		can, err := service.CanDelete(s, 999, u)
		require.NoError(t, err)
		assert.False(t, can)
	})

	t.Run("cannot delete with nil user", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewAPITokenService(testEngine)

		can, err := service.CanDelete(s, 1, nil)
		require.NoError(t, err)
		assert.False(t, can)
	})
}

func TestAPITokenService_HashToken(t *testing.T) {
	service := NewAPITokenService(testEngine)

	t.Run("consistent hashing", func(t *testing.T) {
		token := "tk_testtoken123"
		salt := "testsalt"

		hash1 := service.hashToken(token, salt)
		hash2 := service.hashToken(token, salt)

		assert.Equal(t, hash1, hash2)
		assert.NotEmpty(t, hash1)
	})

	t.Run("different salts produce different hashes", func(t *testing.T) {
		token := "tk_testtoken123"
		salt1 := "salt1"
		salt2 := "salt2"

		hash1 := service.hashToken(token, salt1)
		hash2 := service.hashToken(token, salt2)

		assert.NotEqual(t, hash1, hash2)
	})

	t.Run("different tokens produce different hashes", func(t *testing.T) {
		token1 := "tk_testtoken1"
		token2 := "tk_testtoken2"
		salt := "testsalt"

		hash1 := service.hashToken(token1, salt)
		hash2 := service.hashToken(token2, salt)

		assert.NotEqual(t, hash1, hash2)
	})
}
