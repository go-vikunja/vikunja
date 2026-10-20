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
	"encoding/binary"
	"hash/crc32"
	"image"
	"image/color"
	"image/png"
	"math/rand"
	"os"
	"strconv"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/modules/imageutils"
	"code.vikunja.io/api/pkg/modules/keyvalue"
	"code.vikunja.io/api/pkg/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
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

func newUserWithAvatarFile(t *testing.T) (*xorm.Session, *user.User) {
	t.Helper()
	db.LoadAndAssertFixtures(t)
	files.InitTestFileFixtures(t)

	// Committed so the test's rollback keeps the link.
	setup := db.NewSession()
	_, err := setup.ID(1).Cols("avatar_file_id", "avatar_provider").Update(&user.User{AvatarFileID: 1, AvatarProvider: "upload"})
	require.NoError(t, err)
	require.NoError(t, setup.Commit())
	setup.Close()

	s := db.NewSession()
	t.Cleanup(func() { s.Close() })
	u := &user.User{ID: 1}
	_, err = s.ID(1).Get(u)
	require.NoError(t, err)
	require.Equal(t, int64(1), u.AvatarFileID)
	require.Equal(t, "upload", u.AvatarProvider)
	return s, u
}

func TestStoreAvatarFile(t *testing.T) {
	t.Run("keeps the old avatar when creating the new file fails", func(t *testing.T) {
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

		s, u := newUserWithAvatarFile(t)

		oldMax := config.FilesMaxSize.GetString()
		require.NoError(t, config.SetMaxFileSizeMBytesFromString("1MB"))
		t.Cleanup(func() { _ = config.SetMaxFileSizeMBytesFromString(oldMax) })

		err = StoreAvatarFile(t.Context(), s, u, "upload", src)
		require.Error(t, err)
		assert.True(t, files.IsErrFileIsTooLarge(err))

		assert.Equal(t, int64(1), u.AvatarFileID)
		require.NoError(t, s.Rollback())
		db.AssertExists(t, "users", map[string]interface{}{"id": 1, "avatar_file_id": 1}, false)
		_, err = files.FileStat(&files.File{ID: 1})
		require.NoError(t, err, "old avatar blob must still exist")
	})

	t.Run("stores a non-square upload as a centered square", func(t *testing.T) {
		s, u := newUserWithAvatarFile(t)

		require.NoError(t, StoreAvatarFile(t.Context(), s, u, "upload", bytes.NewReader(encodePNG(t, gradient(300, 200)))))
		require.NoError(t, s.Commit())

		f := &files.File{ID: u.AvatarFileID}
		require.NoError(t, f.LoadFileByID())
		defer f.File.Close()
		img, _, err := image.Decode(f.File)
		require.NoError(t, err)
		assert.Equal(t, image.Rect(0, 0, 200, 200), img.Bounds())
	})

	t.Run("replaces the old avatar on success", func(t *testing.T) {
		s, u := newUserWithAvatarFile(t)

		small := &bytes.Buffer{}
		require.NoError(t, png.Encode(small, image.NewNRGBA(image.Rect(0, 0, 8, 8))))

		require.NoError(t, StoreAvatarFile(t.Context(), s, u, "upload", small))
		assert.NotEqual(t, int64(1), u.AvatarFileID)
		assert.NotZero(t, u.AvatarFileID)
		require.NoError(t, s.Commit())
		db.AssertExists(t, "users", map[string]interface{}{"id": 1, "avatar_file_id": u.AvatarFileID}, false)
		db.AssertMissing(t, "files", map[string]interface{}{"id": 1})
		_, err := files.FileStat(&files.File{ID: 1})
		require.Error(t, err, "old avatar blob must be gone")
		assert.True(t, os.IsNotExist(err))
	})
}

func encodePNG(t *testing.T, img image.Image) []byte {
	t.Helper()
	buf := &bytes.Buffer{}
	require.NoError(t, png.Encode(buf, img))
	return buf.Bytes()
}

func gradient(width, height int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			img.Set(x, y, color.NRGBA{
				R: uint8(x), //nolint:gosec // test sizes stay below 256
				G: uint8(y), //nolint:gosec // test sizes stay below 256
				B: 42,
				A: 255,
			})
		}
	}
	return img
}

func syncAvatar(t *testing.T, raw []byte) image.Image {
	t.Helper()
	db.LoadAndAssertFixtures(t)
	files.InitTestFileFixtures(t)

	s := db.NewSession()
	defer s.Close()
	u, err := user.GetUserByID(s, 1)
	require.NoError(t, err)

	require.NoError(t, StoreSyncedAvatar(t.Context(), s, u, "ldap", raw))
	require.NoError(t, s.Commit())

	require.NotZero(t, u.AvatarFileID)
	db.AssertExists(t, "users", map[string]interface{}{
		"id":              1,
		"avatar_provider": "ldap",
		"avatar_file_id":  u.AvatarFileID,
	}, false)

	f := &files.File{ID: u.AvatarFileID}
	require.NoError(t, f.LoadFileByID())
	defer f.File.Close()
	img, _, err := image.Decode(f.File)
	require.NoError(t, err)
	return img
}

func TestStoreSyncedAvatar(t *testing.T) {
	t.Run("crops a non-square image centered", func(t *testing.T) {
		img := syncAvatar(t, encodePNG(t, gradient(8, 4)))

		assert.Equal(t, image.Rect(0, 0, 4, 4), img.Bounds())
		assert.Equal(t, color.NRGBA{
			R: 2,
			G: 0,
			B: 42,
			A: 255,
		}, color.NRGBAModel.Convert(img.At(0, 0)))
	})

	t.Run("scales a large image down to the max side", func(t *testing.T) {
		img := syncAvatar(t, encodePNG(t, image.NewNRGBA(image.Rect(0, 0, 3000, 2000))))

		assert.Equal(t, image.Rect(0, 0, maxAvatarSide, maxAvatarSide), img.Bounds())
	})

	t.Run("keeps an already square image", func(t *testing.T) {
		original := gradient(4, 4)
		img := syncAvatar(t, encodePNG(t, original))

		require.Equal(t, original.Bounds(), img.Bounds())
		for x := 0; x < 4; x++ {
			for y := 0; y < 4; y++ {
				assert.Equal(t, original.At(x, y), color.NRGBAModel.Convert(img.At(x, y)), "pixel %d,%d", x, y)
			}
		}
	})

	t.Run("rejects a decompression bomb", func(t *testing.T) {
		// GHSA-4vh2-39rq-rq8j
		bomb := encodePNG(t, gradient(4, 4))
		binary.BigEndian.PutUint32(bomb[16:20], uint32(8000))
		binary.BigEndian.PutUint32(bomb[20:24], uint32(8000))
		binary.BigEndian.PutUint32(bomb[29:33], crc32.ChecksumIEEE(bomb[12:29]))

		s := db.NewSession()
		defer s.Close()
		err := StoreSyncedAvatar(t.Context(), s, &user.User{ID: 1}, "ldap", bomb)
		require.Error(t, err)
		assert.True(t, imageutils.IsErrImageTooLarge(err), "err: %v", err)
	})

	t.Run("restores the user when the update fails", func(t *testing.T) {
		s, u := newUserWithAvatarFile(t)
		oldFileID, oldProvider := u.AvatarFileID, u.AvatarProvider

		err := StoreSyncedAvatar(t.Context(), s, u, "bogus", encodePNG(t, gradient(2, 2)))
		var invalid *user.ErrInvalidAvatarProvider
		require.ErrorAs(t, err, &invalid)
		assert.Equal(t, oldFileID, u.AvatarFileID)
		assert.Equal(t, oldProvider, u.AvatarProvider)

		require.NoError(t, s.Rollback())
		db.AssertExists(t, "users", map[string]interface{}{
			"id":              1,
			"avatar_file_id":  oldFileID,
			"avatar_provider": oldProvider,
		}, false)
	})
}

func TestCenteredSquare(t *testing.T) {
	thin := image.NewGray(image.Rect(0, 0, 3, 40))

	sq := centeredSquare(thin)

	assert.Equal(t, image.Rect(0, 18, 3, 21), sq.Bounds())
	gray, ok := sq.(*image.Gray)
	require.True(t, ok)
	assert.Same(t, &thin.Pix[thin.PixOffset(0, 18)], &gray.Pix[0], "must be a view, not a copy")
}
