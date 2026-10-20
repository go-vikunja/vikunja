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
	"image"
	"image/png"
	"io/fs"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/modules/imageutils"
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
