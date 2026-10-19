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
	"bytes"
	"image"
	"image/png"
	"math/rand"
	"os"
	"strconv"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/files"
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

	files.InitTests()
	user.InitTests()
	// Both Init*Tests replace the fixture set; we need users and files.
	if err := db.InitTestFixtures("users", "files"); err != nil {
		log.Fatal(err)
	}

	os.Exit(m.Run())
}

func TestFlushCache(t *testing.T) {
	keyvalue.InitStorage()

	provider := &Provider{}

	testUser := &user.User{
		ID: 777777,
	}

	// Populate cache with multiple size variants (matching the real key format from line 65)
	sizes := []int64{32, 64, 128, 250}
	for _, size := range sizes {
		cacheKey := CacheKeyPrefix + strconv.Itoa(int(testUser.ID)) + "_" + strconv.FormatInt(size, 10)
		err := keyvalue.Put(cacheKey, CachedAvatar{
			Content:  []byte("fake_avatar_data"),
			MimeType: "image/png",
		})
		require.NoError(t, err)
	}

	// Verify all entries exist before flush
	for _, size := range sizes {
		cacheKey := CacheKeyPrefix + strconv.Itoa(int(testUser.ID)) + "_" + strconv.FormatInt(size, 10)
		_, exists, err := keyvalue.Get(cacheKey)
		require.NoError(t, err)
		assert.True(t, exists, "cache entry for size %d should exist before flush", size)
	}

	// Flush cache
	err := provider.FlushCache(testUser)
	require.NoError(t, err)

	// Verify ALL size variants are removed
	for _, size := range sizes {
		cacheKey := CacheKeyPrefix + strconv.Itoa(int(testUser.ID)) + "_" + strconv.FormatInt(size, 10)
		_, exists, err := keyvalue.Get(cacheKey)
		require.NoError(t, err)
		assert.False(t, exists, "cache entry for size %d should be removed after flush", size)
	}
}

func TestGetAvatar(t *testing.T) {
	// Initialize storage for testing
	keyvalue.InitStorage()

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

	t.Run("avatar file missing from storage", func(t *testing.T) {
		files.InitTestFileHandler()

		provider := &Provider{}
		testUser := &user.User{
			ID:           999999,
			AvatarFileID: 424242,
		}

		_, _, err := provider.GetAvatar(testUser, 32)
		require.Error(t, err)
		assert.True(t, files.IsErrFileDoesNotExist(err))
	})
}

func TestStoreAvatarFile(t *testing.T) {
	// Incompressible, so the encoded avatar exceeds the 1MB limit below.
	noise := image.NewNRGBA(image.Rect(0, 0, 1024, 1024))
	rng := rand.New(rand.NewSource(1)) //nolint:gosec // deterministic test data
	_, err := rng.Read(noise.Pix)
	require.NoError(t, err)
	for i := 3; i < len(noise.Pix); i += 4 {
		noise.Pix[i] = 255
	}
	src := &bytes.Buffer{}
	require.NoError(t, png.Encode(src, noise))

	t.Run("keeps the old avatar when creating the new file fails", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		files.InitTestFileFixtures(t)

		s := db.NewSession()
		defer s.Close()

		// File 1 is a fixture with content in storage.
		_, err := s.ID(1).Cols("avatar_file_id").Update(&user.User{AvatarFileID: 1})
		require.NoError(t, err)
		u := &user.User{ID: 1}
		_, err = s.ID(1).Get(u)
		require.NoError(t, err)
		require.Equal(t, int64(1), u.AvatarFileID)

		oldMax := config.FilesMaxSize.GetString()
		require.NoError(t, config.SetMaxFileSizeMBytesFromString("1MB"))
		t.Cleanup(func() { _ = config.SetMaxFileSizeMBytesFromString(oldMax) })

		err = StoreAvatarFile(t.Context(), s, u, bytes.NewReader(src.Bytes()))
		require.Error(t, err)
		assert.True(t, files.IsErrFileIsTooLarge(err))

		assert.Equal(t, int64(1), u.AvatarFileID)
		_, err = files.FileStat(&files.File{ID: 1})
		require.NoError(t, err, "old avatar blob must still exist")
	})

	t.Run("replaces the old avatar on success", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		files.InitTestFileFixtures(t)

		s := db.NewSession()
		defer s.Close()

		_, err := s.ID(1).Cols("avatar_file_id").Update(&user.User{AvatarFileID: 1})
		require.NoError(t, err)
		u := &user.User{ID: 1}
		_, err = s.ID(1).Get(u)
		require.NoError(t, err)

		small := &bytes.Buffer{}
		require.NoError(t, png.Encode(small, image.NewNRGBA(image.Rect(0, 0, 8, 8))))

		require.NoError(t, StoreAvatarFile(t.Context(), s, u, small))
		assert.NotEqual(t, int64(1), u.AvatarFileID)
		assert.NotZero(t, u.AvatarFileID)
		_, err = files.FileStat(&files.File{ID: 1})
		require.Error(t, err, "old avatar blob must be gone")
	})
}
