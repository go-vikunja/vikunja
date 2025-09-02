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

package upload

import (
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
			ID:           999999, // Use a high ID to avoid conflicts
			AvatarFileID: 0,      // No avatar file ID to avoid actual file operations
		}

		// Simulate corrupted cached data by storing a string instead of CachedAvatar
		cacheKey := CacheKeyPrefix + "999999_64"
		err := keyvalue.Put(cacheKey, "corrupted_string_data")
		require.NoError(t, err)

		// This should not panic but should handle the type assertion gracefully
		// and return an error (since there's no actual avatar file)
		avatar, mimeType, err := provider.GetAvatar(testUser, 64)

		// The function should handle the type assertion failure gracefully
		// and attempt to regenerate the avatar (which will fail due to no file)
		require.Error(t, err)
		assert.Nil(t, avatar)
		assert.Empty(t, mimeType)
	})

	t.Run("handles valid cached type", func(t *testing.T) {
		provider := &Provider{}

		// Create a test user
		testUser := &user.User{
			ID:           888888, // Use a different ID to avoid cache conflicts
			AvatarFileID: 0,
		}

		// Store a valid cached avatar
		cacheKey := CacheKeyPrefix + "888888_32"
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
}
