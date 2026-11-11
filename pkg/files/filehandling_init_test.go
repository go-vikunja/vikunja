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

package files

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"code.vikunja.io/api/pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// resetFilesStateAfterTest restores the config and storage backend TestMain set up.
func resetFilesStateAfterTest(t *testing.T) {
	originalStorage := storage

	t.Cleanup(func() {
		config.ResetForTests()
		setDefaultLocalConfig()
		storage = originalStorage
	})
}

// setupLocalStorageTest points the file config at a not-yet-existing directory
// inside a temp dir.
func setupLocalStorageTest(t *testing.T) (basePath string) {
	resetFilesStateAfterTest(t)

	basePath = filepath.Join(t.TempDir(), "files")
	config.FilesType.Set("local")
	config.FilesBasePath.Set(basePath)
	storage = nil

	return basePath
}

func TestInitStorageBackend_DoesNotCreateBasePath(t *testing.T) {
	basePath := setupLocalStorageTest(t)

	require.NoError(t, InitStorageBackend(context.Background()))

	local, ok := storage.(*localStorage)
	require.True(t, ok)
	assert.Equal(t, basePath, local.basePath)

	_, err := os.Stat(basePath)
	require.ErrorIs(t, err, os.ErrNotExist, "InitStorageBackend must not create the base path")
}

func TestInitFileHandler_CreatesBasePath(t *testing.T) {
	basePath := setupLocalStorageTest(t)

	require.NoError(t, InitFileHandler(context.Background()))

	info, err := os.Stat(basePath)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestValidateFileStorage_MissingBasePathFails(t *testing.T) {
	basePath := setupLocalStorageTest(t)

	require.NoError(t, InitStorageBackend(context.Background()))

	require.ErrorContains(t, ValidateFileStorage(context.Background()), "failed to access file storage directory")

	_, err := os.Stat(basePath)
	require.ErrorIs(t, err, os.ErrNotExist)
}

func TestValidateFileStorage_BasePathIsAFile(t *testing.T) {
	basePath := setupLocalStorageTest(t)

	require.NoError(t, os.WriteFile(basePath, []byte("not a directory"), 0600))
	require.NoError(t, InitStorageBackend(context.Background()))

	require.ErrorContains(t, ValidateFileStorage(context.Background()), "file storage path exists but is not a directory")
}
