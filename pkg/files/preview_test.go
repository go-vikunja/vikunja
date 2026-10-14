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
	"bytes"
	"io"
	"io/fs"
	"testing"

	"code.vikunja.io/api/pkg/db"

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
func previewBackends() map[string]func(t *testing.T) FileStorage {
	return map[string]func(t *testing.T) FileStorage{
		"local":  func(t *testing.T) FileStorage { return newLocalStorage(t.TempDir()) },
		"memory": func(*testing.T) FileStorage { return newMemStorage() },
	}
}

func TestFilePreview(t *testing.T) {
	for name, newBackend := range previewBackends() {
		t.Run(name, func(t *testing.T) {
			assertStoresPreviews(t, newBackend(t))
		})
	}

	t.Run("a size that is not stored is rejected", func(t *testing.T) {
		useStorage(t, newMemStorage())
		file := &File{ID: 1}

		for _, size := range []PreviewSize{
			PreviewSizeUnknown,
			"../1",
		} {
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

func TestFilePreview_Cleanup(t *testing.T) {
	savePreviews := func(t *testing.T, file *File) {
		t.Helper()
		for _, size := range storedPreviewSizes {
			require.NoError(t, file.SavePreview(size, []byte("preview")))
		}
	}
	assertPreviews := func(t *testing.T, file *File, stored bool) {
		t.Helper()
		for _, size := range storedPreviewSizes {
			preview, _, err := file.OpenPreview(size)
			if !stored {
				require.ErrorIs(t, err, fs.ErrNotExist, "size %s", size)
				continue
			}
			require.NoError(t, err, "size %s", size)
			require.NoError(t, preview.Close())
		}
	}

	for name, newBackend := range previewBackends() {
		t.Run(name, func(t *testing.T) {
			t.Run("deleting a file removes its previews", func(t *testing.T) {
				useStorage(t, newBackend(t))
				initFixtures(t)
				s := db.NewSession()
				defer s.Close()
				file := &File{ID: 1}
				other := &File{ID: 11}
				savePreviews(t, file)
				savePreviews(t, other)

				require.NoError(t, file.Delete(s))

				assertPreviews(t, file, false)
				assertPreviews(t, other, true)
			})

			t.Run("deleting a file without previews is not an error", func(t *testing.T) {
				useStorage(t, newBackend(t))
				initFixtures(t)
				s := db.NewSession()
				defer s.Close()

				require.NoError(t, (&File{ID: 1}).Delete(s))
			})
		})
	}
}

func TestGetPreviewSizeFromString(t *testing.T) {
	for _, size := range storedPreviewSizes {
		assert.Equal(t, size, GetPreviewSizeFromString(string(size)))
	}
	for _, size := range []string{
		"",
		string(PreviewSizeUnknown),
		"../1",
	} {
		assert.Equal(t, PreviewSizeUnknown, GetPreviewSizeFromString(size), "size %q", size)
	}
}
