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

package avatar

import (
	"bytes"
	"context"
	"image"
	"image/png"
	"io/fs"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/modules/avatar/upload"
	"code.vikunja.io/api/pkg/modules/imageutils"
	"code.vikunja.io/api/pkg/modules/keyvalue"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// GHSA-3x3m-2j93-w46j
func TestNormalizeAvatarSize(t *testing.T) {
	oldMax := config.ServiceMaxAvatarSize.GetInt64()
	t.Cleanup(func() { config.ServiceMaxAvatarSize.Set(oldMax) })

	config.ServiceMaxAvatarSize.Set(1024)
	for in, want := range map[int64]int64{
		-1000: 256,
		-1:    256,
		0:     256,
		1:     16,
		16:    16,
		17:    32,
		49:    64,
		250:   256,
		1000:  1024,
		1024:  1024,
		99999: 1024,
	} {
		assert.Equal(t, want, normalizeAvatarSize(in), "size %d", in)
	}

	t.Run("capped at the configured max", func(t *testing.T) {
		config.ServiceMaxAvatarSize.Set(100)
		assert.Equal(t, int64(64), normalizeAvatarSize(64))
		assert.Equal(t, int64(100), normalizeAvatarSize(65))
		assert.Equal(t, int64(100), normalizeAvatarSize(99999))
	})

	t.Run("max above the largest allowed size", func(t *testing.T) {
		config.ServiceMaxAvatarSize.Set(2048)
		assert.Equal(t, int64(1024), normalizeAvatarSize(1000))
		assert.Equal(t, int64(2048), normalizeAvatarSize(1500))
	})
}

func TestSyncFromProvider(t *testing.T) {
	raw := &bytes.Buffer{}
	require.NoError(t, png.Encode(raw, image.NewNRGBA(image.Rect(0, 0, 8, 4))))

	t.Run("stores provider and file", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		require.NoError(t, SyncFromProvider(t.Context(), 1, "ldap", raw.Bytes()))

		s := db.NewSession()
		defer s.Close()
		stored, err := user.GetUserByID(s, 1)
		require.NoError(t, err)
		assert.Equal(t, "ldap", stored.AvatarProvider)
		require.NotZero(t, stored.AvatarFileID)
		db.AssertExists(t, "files", map[string]interface{}{
			"id":            stored.AvatarFileID,
			"name":          "avatar.png",
			"mime":          "image/png",
			"created_by_id": 1,
		}, false)

		f := &files.File{ID: stored.AvatarFileID}
		require.NoError(t, f.LoadFileByID())
		defer f.File.Close()
		cfg, err := png.DecodeConfig(f.File)
		require.NoError(t, err)
		assert.Equal(t, 4, cfg.Width)
		assert.Equal(t, 4, cfg.Height)
	})

	t.Run("removes the replaced avatar file and blob", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		files.InitTestFileFixtures(t)

		setup := db.NewSession()
		_, err := setup.ID(1).Cols("avatar_file_id", "avatar_provider").Update(&user.User{AvatarFileID: 1, AvatarProvider: "upload"})
		require.NoError(t, err)
		require.NoError(t, setup.Commit())
		setup.Close()

		_, err = files.FileStat(&files.File{ID: 1})
		require.NoError(t, err)

		require.NoError(t, SyncFromProvider(t.Context(), 1, "ldap", raw.Bytes()))

		db.AssertMissing(t, "files", map[string]interface{}{"id": 1})
		_, err = files.FileStat(&files.File{ID: 1})
		require.Error(t, err)
		assert.ErrorIs(t, err, fs.ErrNotExist)
	})

	t.Run("busy decoder leaves the avatar unchanged", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		t.Cleanup(imageutils.ExhaustBudgetForTests(10 * time.Millisecond))

		err := SyncFromProvider(t.Context(), 19, "openid", raw.Bytes())
		require.Error(t, err)
		assert.True(t, imageutils.IsErrDecodeBusy(err))

		s := db.NewSession()
		defer s.Close()
		stored, err := user.GetUserByID(s, 19)
		require.NoError(t, err)
		assert.Equal(t, "openid", stored.AvatarProvider)
		assert.Equal(t, int64(0), stored.AvatarFileID)
	})
}

func TestSyncFromProviderAsync(t *testing.T) {
	raw := &bytes.Buffer{}
	require.NoError(t, png.Encode(raw, image.NewNRGBA(image.Rect(0, 0, 8, 4))))

	storedUser := func(t *testing.T) *user.User {
		t.Helper()
		s := db.NewSession()
		defer s.Close()
		u, err := user.GetUserByID(s, 1)
		require.NoError(t, err)
		return u
	}

	t.Run("stores provider and file", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		SyncFromProviderAsync(t.Context(), 1, "ldap", func(context.Context) ([]byte, error) {
			return raw.Bytes(), nil
		})
		WaitForProviderSyncsForTests()

		stored := storedUser(t)
		assert.Equal(t, "ldap", stored.AvatarProvider)
		require.NotZero(t, stored.AvatarFileID)
		db.AssertExists(t, "files", map[string]interface{}{
			"id":            stored.AvatarFileID,
			"name":          "avatar.png",
			"created_by_id": 1,
		}, false)
	})

	t.Run("recovers a panicking fetch", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		panicking := func(context.Context) ([]byte, error) {
			panic("boom")
		}

		SyncFromProviderAsync(t.Context(), 1, "ldap", panicking)
		WaitForProviderSyncsForTests()

		stored := storedUser(t)
		assert.NotEqual(t, "ldap", stored.AvatarProvider)
		assert.Zero(t, stored.AvatarFileID)

		err := fetchAndSyncFromProvider(t.Context(), 1, "ldap", panicking)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "panic: boom")
	})

	t.Run("outlives the request context", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		ctx, cancel := context.WithCancel(t.Context())
		release := make(chan struct{})

		SyncFromProviderAsync(ctx, 1, "ldap", func(ctx context.Context) ([]byte, error) {
			<-release
			return raw.Bytes(), ctx.Err()
		})
		cancel()
		close(release)
		WaitForProviderSyncsForTests()

		stored := storedUser(t)
		assert.Equal(t, "ldap", stored.AvatarProvider)
		assert.NotZero(t, stored.AvatarFileID)
	})
}

func TestSaveUploaded(t *testing.T) {
	raw := &bytes.Buffer{}
	require.NoError(t, png.Encode(raw, image.NewNRGBA(image.Rect(0, 0, 4, 4))))
	prepared, err := PrepareUpload(t.Context(), raw)
	require.NoError(t, err)

	const cacheKey = upload.CacheKeyPrefix + "1_1_32"

	setup := func(t *testing.T) *user.User {
		t.Helper()
		db.LoadAndAssertFixtures(t)
		files.InitTestFileFixtures(t)
		keyvalue.InitStorage()

		s := db.NewSession()
		defer s.Close()
		_, err := s.ID(1).Cols("avatar_file_id", "avatar_provider").Update(&user.User{AvatarFileID: 1, AvatarProvider: "upload"})
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		u, err := user.GetUserByID(s, 1)
		require.NoError(t, err)
		require.Equal(t, int64(1), u.AvatarFileID)

		require.NoError(t, keyvalue.Put(cacheKey, upload.CachedAvatar{
			Content:  []byte("old"),
			MimeType: "image/png",
		}))
		return u
	}

	cached := func(t *testing.T) bool {
		t.Helper()
		_, exists, err := keyvalue.Get(cacheKey)
		require.NoError(t, err)
		return exists
	}

	t.Run("keeps the cache when the session rolls back", func(t *testing.T) {
		u := setup(t)
		s := db.NewSession()
		defer s.Close()

		require.NoError(t, SaveUploaded(s, u, prepared))
		assert.True(t, cached(t))
		require.NoError(t, s.Rollback())
		events.CleanupPending(s)
		events.DispatchPending(t.Context(), s)

		assert.True(t, cached(t))
	})

	t.Run("flushes the cache once the commit is dispatched", func(t *testing.T) {
		u := setup(t)
		s := db.NewSession()
		defer s.Close()

		require.NoError(t, SaveUploaded(s, u, prepared))
		require.NoError(t, s.Commit())
		assert.True(t, cached(t))
		events.DispatchPending(t.Context(), s)

		assert.False(t, cached(t))
	})
}
