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

package initials

import (
	"image"
	"os"
	"testing"

	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/modules/keyvalue"
	"code.vikunja.io/api/pkg/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMain initializes the test environment
func TestMain(m *testing.M) {
	// Initialize logger for tests
	log.InitLogger()

	os.Exit(m.Run())
}

func TestGetAvatar(t *testing.T) {
	// Initialize storage for testing
	keyvalue.InitStorage()

	t.Run("handles invalid cached type", func(t *testing.T) {
		provider := &Provider{}

		// Create a test user
		testUser := &user.User{
			ID:       999999, // Use a high ID to avoid conflicts
			Name:     "Test User",
			Username: "testuser",
		}

		// Simulate corrupted cached data by storing a string instead of CachedAvatar
		cacheKey := getCacheKey("resized", testUser.ID, 64)
		err := keyvalue.Put(cacheKey, "corrupted_string_data")
		require.NoError(t, err)

		// This should not panic but should handle the type assertion gracefully
		// and regenerate the avatar
		avatar, mimeType, err := provider.GetAvatar(testUser, 64)

		// The function should handle the type assertion failure gracefully
		// and regenerate the avatar successfully
		require.NoError(t, err)
		assert.NotNil(t, avatar)
		assert.Equal(t, "image/png", mimeType)
		assert.NotEmpty(t, avatar, "Avatar should contain image data")
	})

	t.Run("handles valid cached type", func(t *testing.T) {
		provider := &Provider{}

		// Create a test user
		testUser := &user.User{
			ID:       888888, // Use a different ID to avoid cache conflicts
			Name:     "Valid User",
			Username: "validuser",
		}

		// Store a valid cached avatar
		cacheKey := getCacheKey("resized", testUser.ID, 32)
		validCachedAvatar := CachedAvatar{
			Content:  []byte("fake_image_data"),
			MimeType: "image/png",
		}
		err := keyvalue.Put(cacheKey, validCachedAvatar)
		require.NoError(t, err)

		// This should work correctly with the valid cached data
		avatar, mimeType, err := provider.GetAvatar(testUser, 32)

		// Should return the cached data successfully
		require.NoError(t, err)
		assert.Equal(t, []byte("fake_image_data"), avatar)
		assert.Equal(t, "image/png", mimeType)
	})

	t.Run("generates valid initials", func(t *testing.T) {
		provider := &Provider{}

		// Test with name
		testUser1 := &user.User{
			ID:       555555,
			Name:     "John Doe",
			Username: "johndoe",
		}

		avatar1, mimeType1, err1 := provider.GetAvatar(testUser1, 128)
		require.NoError(t, err1)
		assert.NotNil(t, avatar1)
		assert.Equal(t, "image/png", mimeType1)
		assert.NotEmpty(t, avatar1)

		// Test with username when name is empty
		testUser2 := &user.User{
			ID:       444444,
			Name:     "", // Empty name
			Username: "jane_smith",
		}

		avatar2, mimeType2, err2 := provider.GetAvatar(testUser2, 128)
		require.NoError(t, err2)
		assert.NotNil(t, avatar2)
		assert.Equal(t, "image/png", mimeType2)
		assert.NotEmpty(t, avatar2)
	})
}

func TestGetAvatarForUser(t *testing.T) {
	// Initialize storage for testing
	keyvalue.InitStorage()

	t.Run("handles invalid cached type", func(t *testing.T) {
		// Create a test user
		testUser := &user.User{
			ID:       777777, // Use another unique ID
			Name:     "Full Size Test User",
			Username: "fullsizeuser",
		}

		// Simulate corrupted cached data by storing a string instead of image.RGBA64
		cacheKey := getCacheKey("full", testUser.ID)
		err := keyvalue.Put(cacheKey, "corrupted_image_data")
		require.NoError(t, err)

		// This should not panic but should handle the type assertion gracefully
		// and regenerate the full size avatar
		fullAvatar, err := getAvatarForUser(testUser)

		// The function should handle the type assertion failure gracefully
		// and generate a new avatar successfully
		require.NoError(t, err)
		assert.NotNil(t, fullAvatar)
		assert.IsType(t, &image.RGBA64{}, fullAvatar)
	})

	t.Run("handles valid cached type", func(t *testing.T) {
		// Create a test user
		testUser := &user.User{
			ID:       666666, // Use another unique ID
			Name:     "Valid Full Size User",
			Username: "validfulluser",
		}

		// Create a valid image.RGBA64 for caching
		validImage := image.NewRGBA64(image.Rect(0, 0, 64, 64))
		cacheKey := getCacheKey("full", testUser.ID)
		err := keyvalue.Put(cacheKey, *validImage)
		require.NoError(t, err)

		// This should work correctly with the valid cached data
		fullAvatar, err := getAvatarForUser(testUser)

		// Should return the cached image successfully
		require.NoError(t, err)
		assert.NotNil(t, fullAvatar)
		assert.IsType(t, &image.RGBA64{}, fullAvatar)
	})
}
