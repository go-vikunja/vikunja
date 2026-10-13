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
	"io"
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func useStorage(t *testing.T, s FileStorage) {
	t.Helper()
	original := storage
	t.Cleanup(func() { storage = original })
	storage = s
}

// The S3 backend is covered by TestFileStorageIntegration.
func TestFilePreview(t *testing.T) {
	backends := map[string]func(t *testing.T) FileStorage{
		"local":  func(t *testing.T) FileStorage { return newLocalStorage(t.TempDir()) },
		"memory": func(*testing.T) FileStorage { return newMemStorage() },
	}
	for name, newBackend := range backends {
		t.Run(name, func(t *testing.T) {
			assertStoresPreviews(t, newBackend(t))
		})
	}

	t.Run("a size that is not stored is rejected", func(t *testing.T) {
		useStorage(t, newMemStorage())
		file := &File{ID: 1}

		for _, size := range []PreviewSize{PreviewSizeUnknown, "../1"} {
			require.Error(t, file.SavePreview(size, []byte("preview")))
			_, _, err := file.OpenPreview(size)
			require.Error(t, err)
			require.NotErrorIs(t, err, fs.ErrNotExist)
		}
	})
}

func assertStoresPreviews(t *testing.T, backend FileStorage) {
	t.Helper()
	useStorage(t, backend)
	file := &File{ID: 42}
	// A real bucket outlives the test run.
	t.Cleanup(func() {
		for _, size := range storedPreviewSizes {
			_ = backend.Remove(file.previewPath(size))
		}
	})

	_, _, err := file.OpenPreview(PreviewMedium)
	require.ErrorIs(t, err, fs.ErrNotExist)

	for _, size := range storedPreviewSizes {
		require.NoError(t, file.SavePreview(size, []byte("preview "+size)))
	}
	require.NoError(t, file.SavePreview(PreviewMedium, []byte("overwritten")))

	for _, size := range storedPreviewSizes {
		want := []byte("preview " + size)
		if size == PreviewMedium {
			want = []byte("overwritten")
		}

		preview, length, err := file.OpenPreview(size)
		require.NoError(t, err)
		got, err := io.ReadAll(preview)
		require.NoError(t, err)
		require.NoError(t, preview.Close())
		assert.Equal(t, want, got)
		assert.Equal(t, int64(len(want)), length)
	}

	_, _, err = (&File{ID: 43}).OpenPreview(PreviewMedium)
	require.ErrorIs(t, err, fs.ErrNotExist, "previews are stored per file")
}
