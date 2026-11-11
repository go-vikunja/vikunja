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
	"errors"
	"image"
	"image/png"
	"io"
	"io/fs"
	"net/http"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testauth struct {
	id int64
}

func (a *testauth) GetID() int64 {
	return a.id
}

func TestCreate(t *testing.T) {
	t.Run("Normal", func(t *testing.T) {
		initFixtures(t)
		ta := &testauth{id: 1}
		content := []byte("testfile")
		createdFile, err := Create(bytes.NewReader(content), "testfile", uint64(len(content)), ta)
		require.NoError(t, err)

		// Check the file was created correctly
		s := db.NewSession()
		defer s.Close()
		file := &File{ID: createdFile.ID}
		err = file.LoadFileMetaByID(s)
		require.NoError(t, err)
		assert.Equal(t, int64(1), file.CreatedByID)
		assert.Equal(t, "testfile", file.Name)
		assert.Equal(t, uint64(len(content)), file.Size)

	})
	t.Run("Too Large", func(t *testing.T) {
		initFixtures(t)
		// initFixtures resets FilesMaxSize, so override afterwards.
		oldMax := config.FilesMaxSize.GetString()
		require.NoError(t, config.SetMaxFileSizeMBytesFromString("1MB"))
		t.Cleanup(func() { _ = config.SetMaxFileSizeMBytesFromString(oldMax) })

		ta := &testauth{id: 1}
		_, err := Create(bytes.NewReader(make([]byte, 2*1024*1024)), "big.bin", 0, ta)
		require.Error(t, err)
		assert.True(t, IsErrFileIsTooLarge(err))
	})
}

func TestCreate_HardenedAgainstLyingCaller(t *testing.T) {
	// Regression: GHSA-qh78-rvg3-cv54. Caller's claimed size must not
	// influence the limit check or the stored size.

	// setLimit must run AFTER initFixtures since initFixtures resets it.
	setLimit := func(t *testing.T, size string) {
		t.Helper()
		oldMax := config.FilesMaxSize.GetString()
		require.NoError(t, config.SetMaxFileSizeMBytesFromString(size))
		t.Cleanup(func() {
			_ = config.SetMaxFileSizeMBytesFromString(oldMax)
		})
	}

	ta := &testauth{id: 1}

	t.Run("content over limit, claimed size under limit", func(t *testing.T) {
		initFixtures(t)
		setLimit(t, "1MB")
		payload := make([]byte, 2*1024*1024)
		_, err := Create(bytes.NewReader(payload), "sneaky.bin", 1, ta)
		require.Error(t, err)
		assert.True(t, IsErrFileIsTooLarge(err))
	})

	t.Run("content under limit, claimed size over limit", func(t *testing.T) {
		initFixtures(t)
		setLimit(t, "1MB")
		payload := []byte("hi")
		f, err := Create(bytes.NewReader(payload), "honest.txt", 99999999999, ta)
		require.NoError(t, err)
		assert.Equal(t, uint64(len(payload)), f.Size)
	})

	t.Run("content exactly at limit", func(t *testing.T) {
		initFixtures(t)
		setLimit(t, "1MB")
		payload := make([]byte, 1*1024*1024)
		f, err := Create(bytes.NewReader(payload), "edge.bin", 0, ta)
		require.NoError(t, err)
		assert.Equal(t, uint64(1*1024*1024), f.Size)
	})
}

func TestCreateDetectsMimeType(t *testing.T) {
	initFixtures(t)
	ta := &testauth{id: 1}

	// Minimal valid PNG (1x1 pixel)
	pngData := createMinimalPNG(t)

	f, err := Create(bytes.NewReader(pngData), "test.png", uint64(len(pngData)), ta)
	require.NoError(t, err)
	assert.Equal(t, "image/png", f.Mime)
}

func TestCreateDetectsMimeTypePlainText(t *testing.T) {
	initFixtures(t)
	ta := &testauth{id: 1}

	textData := []byte("hello world this is plain text")

	f, err := Create(bytes.NewReader(textData), "readme.txt", uint64(len(textData)), ta)
	require.NoError(t, err)
	assert.Equal(t, "text/plain; charset=utf-8", f.Mime)
}

func createMinimalPNG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	buf := &bytes.Buffer{}
	require.NoError(t, png.Encode(buf, img))
	return buf.Bytes()
}

func TestFile_Delete(t *testing.T) {
	t.Run("Normal", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()
		initFixtures(t)
		f := &File{ID: 1}
		err := f.Delete(s)
		require.NoError(t, err)
	})
	t.Run("Nonexisting", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()
		initFixtures(t)
		f := &File{ID: 9999}

		err := f.Delete(s)
		require.Error(t, err)
		assert.True(t, IsErrFileDoesNotExist(err))
	})
}

func TestFile_DeleteOnCommit(t *testing.T) {
	t.Run("removes the blob and previews once dispatched", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()
		initFixtures(t)
		f := &File{ID: 1}
		require.NoError(t, f.SavePreview(PreviewMedium, []byte("preview")))

		require.NoError(t, f.DeleteOnCommit(s))
		require.NoError(t, s.Commit())
		db.AssertMissing(t, "files", map[string]interface{}{"id": 1})
		_, err := FileStat(f)
		require.NoError(t, err)

		events.DispatchPending(t.Context(), s)
		_, err = FileStat(f)
		require.ErrorIs(t, err, fs.ErrNotExist)
		_, _, err = f.OpenPreview(PreviewMedium)
		require.ErrorIs(t, err, fs.ErrNotExist)
	})
	t.Run("keeps row and blob on rollback", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()
		initFixtures(t)
		f := &File{ID: 1}

		require.NoError(t, f.DeleteOnCommit(s))
		require.NoError(t, s.Rollback())
		events.CleanupPending(s)
		events.DispatchPending(t.Context(), s)

		db.AssertExists(t, "files", map[string]interface{}{"id": 1}, false)
		_, err := FileStat(f)
		require.NoError(t, err)
	})
	t.Run("Nonexisting", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()
		initFixtures(t)

		err := (&File{ID: 9999}).DeleteOnCommit(s)
		assert.True(t, IsErrFileDoesNotExist(err))
	})
}

func TestFile_LoadFileByID(t *testing.T) {
	t.Run("Normal", func(t *testing.T) {
		initFixtures(t)
		f := &File{ID: 1}
		err := f.LoadFileByID()
		require.NoError(t, err)
	})
	t.Run("Nonexisting", func(t *testing.T) {
		initFixtures(t)
		f := &File{ID: 9999}
		err := f.LoadFileByID()
		require.Error(t, err)
		assert.True(t, IsErrFileDoesNotExist(err))
		assert.Equal(t, http.StatusNotFound, err.(ErrFileDoesNotExist).HTTPError().HTTPCode)
	})
}

func TestFileSave_UsesStorage(t *testing.T) {
	originalStorage := storage
	t.Cleanup(func() {
		storage = originalStorage
	})

	mem := newMemStorage()
	storage = mem

	content := []byte("test-content")
	file := &File{ID: 123, Size: uint64(len(content))}

	err := file.Save(bytes.NewReader(content))
	require.NoError(t, err)

	rc, err := mem.Open(file.fileID())
	require.NoError(t, err)
	defer rc.Close()

	written, err := io.ReadAll(rc)
	require.NoError(t, err)
	assert.Equal(t, content, written)
}

type failingWriteStorage struct {
	*memStorage
}

func (failingWriteStorage) Write(string, io.ReadSeeker, uint64) error {
	return errors.New("write failed")
}

func TestCreateWithMimeAndSession_RemovesRowWhenSaveFails(t *testing.T) {
	initFixtures(t)
	useStorage(t, failingWriteStorage{newMemStorage()})

	s := db.NewSession()
	defer s.Close()

	file, err := CreateWithMimeAndSession(s, bytes.NewReader([]byte("content")), "failing.txt", 0, &testauth{id: 1}, "text/plain", true)
	require.Error(t, err)
	assert.Nil(t, file)

	require.NoError(t, s.Commit())
	db.AssertMissing(t, "files", map[string]interface{}{"name": "failing.txt"})
}

func TestFile_LoadFileMetaByID(t *testing.T) {
	t.Run("Normal", func(t *testing.T) {
		initFixtures(t)
		s := db.NewSession()
		defer s.Close()
		f := &File{ID: 1}
		err := f.LoadFileMetaByID(s)
		require.NoError(t, err)
		assert.Equal(t, "test", f.Name)
	})
	t.Run("Nonexisting", func(t *testing.T) {
		initFixtures(t)
		s := db.NewSession()
		defer s.Close()
		f := &File{ID: 9999}
		err := f.LoadFileMetaByID(s)
		require.Error(t, err)
		assert.True(t, IsErrFileDoesNotExist(err))
	})
}
