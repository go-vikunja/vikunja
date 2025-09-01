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
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLinkShareService(t *testing.T) {
	service := NewLinkShareService(db.GetEngine())

	assert.NotNil(t, service)
	assert.NotNil(t, service.DB)
}

func TestLinkShareService_Create(t *testing.T) {
	db.LoadAndAssertFixtures(t)

	t.Run("normal create", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewLinkShareService(db.GetEngine())
		user1 := &user.User{ID: 1}

		share := &models.LinkSharing{
			ProjectID:  1,
			Permission: models.PermissionRead,
			Name:       "Test Share",
		}

		err := service.Create(s, share, user1)
		require.NoError(t, err)

		assert.NotZero(t, share.ID)
		assert.NotEmpty(t, share.Hash)
		assert.Equal(t, models.SharingTypeWithoutPassword, share.SharingType)
		assert.Equal(t, int64(1), share.SharedByID)
		assert.Empty(t, share.Password) // Should be cleared in response
	})

	t.Run("create with password", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewLinkShareService(db.GetEngine())
		user1 := &user.User{ID: 1}

		share := &models.LinkSharing{
			ProjectID:  1,
			Permission: models.PermissionRead,
			Name:       "Test Share With Password",
			Password:   "testpassword",
		}

		err := service.Create(s, share, user1)
		require.NoError(t, err)

		assert.NotZero(t, share.ID)
		assert.NotEmpty(t, share.Hash)
		assert.Equal(t, models.SharingTypeWithPassword, share.SharingType)
		assert.Empty(t, share.Password) // Should be cleared in response
	})

	t.Run("create with invalid permission", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewLinkShareService(db.GetEngine())
		user1 := &user.User{ID: 1}

		share := &models.LinkSharing{
			ProjectID:  1,
			Permission: models.Permission(999), // Invalid permission
			Name:       "Test Share",
		}

		err := service.Create(s, share, user1)
		require.Error(t, err)
		assert.IsType(t, &models.ErrInvalidPermission{}, err)
	})

	t.Run("create on non-existent project", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewLinkShareService(db.GetEngine())
		user1 := &user.User{ID: 1}

		share := &models.LinkSharing{
			ProjectID:  999999,
			Permission: models.PermissionRead,
			Name:       "Test Share",
		}

		err := service.Create(s, share, user1)
		require.Error(t, err)
	})
}

func TestLinkShareService_GetByHash(t *testing.T) {
	db.LoadAndAssertFixtures(t)

	t.Run("existing hash", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewLinkShareService(db.GetEngine())

		// Use existing fixture hash
		share, err := service.GetByHash(s, "test")
		require.NoError(t, err)

		assert.NotNil(t, share)
		assert.Equal(t, "test", share.Hash)
		assert.Empty(t, share.Password) // Password should be cleared
	})

	t.Run("non-existent hash", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewLinkShareService(db.GetEngine())

		_, err := service.GetByHash(s, "nonexistent")
		require.Error(t, err)
		assert.IsType(t, &models.ErrProjectShareDoesNotExist{}, err)
	})
}

func TestLinkShareService_VerifyPassword(t *testing.T) {
	db.LoadAndAssertFixtures(t)

	service := NewLinkShareService(db.GetEngine())

	t.Run("no password required", func(t *testing.T) {
		share := &models.LinkSharing{
			SharingType: models.SharingTypeWithoutPassword,
		}

		err := service.VerifyPassword(share, "")
		require.NoError(t, err)

		err = service.VerifyPassword(share, "somepassword")
		require.NoError(t, err)
	})

	t.Run("correct password", func(t *testing.T) {
		// Create a link share with a hashed password
		hashedPassword, _ := user.HashPassword("testpassword")
		share := &models.LinkSharing{
			SharingType: models.SharingTypeWithPassword,
			Password:    hashedPassword,
		}

		err := service.VerifyPassword(share, "testpassword")
		require.NoError(t, err)
	})

	t.Run("wrong password", func(t *testing.T) {
		hashedPassword, _ := user.HashPassword("testpassword")
		share := &models.LinkSharing{
			ID:          1,
			SharingType: models.SharingTypeWithPassword,
			Password:    hashedPassword,
		}

		err := service.VerifyPassword(share, "wrongpassword")
		require.Error(t, err)
		assert.IsType(t, &models.ErrLinkSharePasswordInvalid{}, err)
	})

	t.Run("no password provided when required", func(t *testing.T) {
		share := &models.LinkSharing{
			ID:          1,
			SharingType: models.SharingTypeWithPassword,
			Password:    "hashedpassword",
		}

		err := service.VerifyPassword(share, "")
		require.Error(t, err)
		assert.IsType(t, &models.ErrLinkSharePasswordRequired{}, err)
	})
}

func TestLinkShareService_Authenticate(t *testing.T) {
	db.LoadAndAssertFixtures(t)

	t.Run("valid authentication without password", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewLinkShareService(db.GetEngine())

		share, err := service.Authenticate(s, "test", "")
		require.NoError(t, err)

		assert.NotNil(t, share)
		assert.Equal(t, "test", share.Hash)
	})

	t.Run("valid authentication with correct password", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewLinkShareService(db.GetEngine())

		share, err := service.Authenticate(s, "testWithPassword", "12345678")
		require.NoError(t, err)

		assert.NotNil(t, share)
		assert.Equal(t, "testWithPassword", share.Hash)
	})

	t.Run("authentication with wrong password", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewLinkShareService(db.GetEngine())

		_, err := service.Authenticate(s, "testWithPassword", "wrongpassword")
		require.Error(t, err)
		assert.IsType(t, &models.ErrLinkSharePasswordInvalid{}, err)
	})

	t.Run("authentication with non-existent hash", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewLinkShareService(db.GetEngine())

		_, err := service.Authenticate(s, "nonexistent", "")
		require.Error(t, err)
		assert.IsType(t, &models.ErrProjectShareDoesNotExist{}, err)
	})
}

func TestLinkShareService_ToUser(t *testing.T) {
	service := NewLinkShareService(db.GetEngine())

	t.Run("convert link share to user", func(t *testing.T) {
		share := &models.LinkSharing{
			ID:   5,
			Name: "Test Share",
		}

		user := service.ToUser(share)

		assert.Equal(t, int64(-5), user.ID) // Negative of share ID
		assert.Equal(t, "Test Share (Link Share)", user.Name)
		assert.Equal(t, "link-share-5", user.Username)
	})

	t.Run("convert link share without name to user", func(t *testing.T) {
		share := &models.LinkSharing{
			ID:   10,
			Name: "",
		}

		user := service.ToUser(share)

		assert.Equal(t, int64(-10), user.ID)
		assert.Equal(t, "Link Share", user.Name)
		assert.Equal(t, "link-share-10", user.Username)
	})
}

func TestLinkShareService_CreateJWTToken(t *testing.T) {
	service := NewLinkShareService(db.GetEngine())

	share := &models.LinkSharing{
		ID:         1,
		Hash:       "testhash",
		ProjectID:  5,
		Permission: models.PermissionRead,
		SharedByID: 10,
	}

	token, err := service.CreateJWTToken(share)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Token should be a valid JWT string with 3 parts separated by dots
	parts := len(strings.Split(token, "."))
	assert.Equal(t, 3, parts)
}

func TestLinkShareService_GetUsersOrLinkSharesFromIDs(t *testing.T) {
	db.LoadAndAssertFixtures(t)

	t.Run("get mixed users and link shares", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewLinkShareService(db.GetEngine())

		// Mix of user IDs (positive) and link share IDs (negative)
		ids := []int64{1, -1, 2, -2} // Users 1,2 and link shares 1,2

		users, err := service.GetUsersOrLinkSharesFromIDs(s, ids)
		require.NoError(t, err)

		// Should have 4 entries
		assert.Len(t, users, 4)

		// Check positive IDs (regular users)
		assert.Contains(t, users, int64(1))
		assert.Contains(t, users, int64(2))

		// Check negative IDs (link shares converted to users)
		assert.Contains(t, users, int64(-1))
		assert.Contains(t, users, int64(-2))

		// Verify link share users have correct properties
		linkShareUser1 := users[-1]
		assert.Contains(t, linkShareUser1.Username, "link-share")
		assert.Contains(t, linkShareUser1.Name, "Link Share")
	})

	t.Run("empty ID list", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewLinkShareService(db.GetEngine())

		users, err := service.GetUsersOrLinkSharesFromIDs(s, []int64{})
		require.NoError(t, err)
		assert.Empty(t, users)
	})
}
