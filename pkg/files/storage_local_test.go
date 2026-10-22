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
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalStorage_WriteAndOpen(t *testing.T) {
	dir := t.TempDir()
	s := newLocalStorage(dir)

	content := []byte("hello local")

	err := s.Write("testfile", bytes.NewReader(content), uint64(len(content)))
	require.NoError(t, err)

	rc, err := s.Open("testfile")
	require.NoError(t, err)
	defer rc.Close()

	got, err := io.ReadAll(rc)
	require.NoError(t, err)
	assert.Equal(t, content, got)
}

// GHSA-9rfw-7hv4-3vv3
func TestLocalStorage_Permissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix permissions")
	}

	basePath := filepath.Join(t.TempDir(), "files")
	s := newLocalStorage(basePath)
	require.NoError(t, s.Ensure())
	require.NoError(t, s.Write("testfile", bytes.NewReader([]byte("secret")), 6))

	for path, want := range map[string]os.FileMode{
		basePath:                              0700,
		filepath.Join(basePath, LocalTempDir): 0700,
		filepath.Join(basePath, "testfile"):   0600,
	} {
		info, err := os.Stat(path)
		require.NoError(t, err)
		assert.Equal(t, want, info.Mode().Perm(), path)
	}
}

func TestLocalStorage_OpenWithSize(t *testing.T) {
	s := newLocalStorage(t.TempDir())
	content := []byte("hello local")
	require.NoError(t, s.Write("testfile", bytes.NewReader(content), uint64(len(content))))

	rc, size, err := s.OpenWithSize("testfile")
	require.NoError(t, err)
	defer rc.Close()

	got, err := io.ReadAll(rc)
	require.NoError(t, err)
	assert.Equal(t, content, got)
	assert.Equal(t, int64(len(content)), size)

	_, _, err = s.OpenWithSize("nonexistent")
	require.ErrorIs(t, err, os.ErrNotExist)
}

type failingReader struct {
	io.ReadSeeker
}

func (failingReader) Read([]byte) (int, error) {
	return 0, errors.New("read failed")
}

func TestLocalStorage_WriteIsAtomic(t *testing.T) {
	t.Run("a failed write leaves no file", func(t *testing.T) {
		dir := t.TempDir()
		s := newLocalStorage(dir)

		err := s.Write("testfile", failingReader{ReadSeeker: bytes.NewReader([]byte("content"))}, 7)
		require.Error(t, err)

		_, err = os.Stat(filepath.Join(dir, "testfile"))
		require.ErrorIs(t, err, os.ErrNotExist)
		assert.Empty(t, tempFiles(t, dir))
	})

	t.Run("a failed write keeps the previous content", func(t *testing.T) {
		s := newLocalStorage(t.TempDir())
		content := []byte("previous")
		require.NoError(t, s.Write("testfile", bytes.NewReader(content), uint64(len(content))))

		err := s.Write("testfile", failingReader{ReadSeeker: bytes.NewReader([]byte("new"))}, 3)
		require.Error(t, err)

		rc, err := s.Open("testfile")
		require.NoError(t, err)
		defer rc.Close()
		got, err := io.ReadAll(rc)
		require.NoError(t, err)
		assert.Equal(t, content, got)
	})

	t.Run("overwrites and leaves no temporary file", func(t *testing.T) {
		dir := t.TempDir()
		s := newLocalStorage(dir)
		require.NoError(t, s.Write("testfile", bytes.NewReader([]byte("previous")), 8))
		require.NoError(t, s.Write("testfile", bytes.NewReader([]byte("new")), 3))

		got, err := os.ReadFile(filepath.Join(dir, "testfile"))
		require.NoError(t, err)
		assert.Equal(t, []byte("new"), got)

		assert.Empty(t, tempFiles(t, dir))
	})

	t.Run("writes the temporary file into the temporary directory", func(t *testing.T) {
		dir := t.TempDir()
		s := newLocalStorage(dir)
		content := &inspectingReader{
			ReadSeeker: bytes.NewReader([]byte("content")),
			inspect: func() {
				assert.Len(t, tempFiles(t, dir), 1)
			},
		}

		require.NoError(t, s.MkdirAll("previews", 0755))
		require.NoError(t, s.Write("previews/testfile", content, 7))

		assert.True(t, content.inspected)
		entries, err := os.ReadDir(filepath.Join(dir, "previews"))
		require.NoError(t, err)
		assert.Len(t, entries, 1)
	})
}

type inspectingReader struct {
	io.ReadSeeker
	inspect   func()
	inspected bool
}

func (r *inspectingReader) Read(p []byte) (int, error) {
	if !r.inspected {
		r.inspected = true
		r.inspect()
	}
	return r.ReadSeeker.Read(p)
}

func tempFiles(t *testing.T, basePath string) []os.DirEntry {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(basePath, LocalTempDir))
	require.NoError(t, err)
	return entries
}

func TestLocalStorage_Stat(t *testing.T) {
	dir := t.TempDir()
	s := newLocalStorage(dir)

	content := []byte("stat me")

	err := s.Write("statfile", bytes.NewReader(content), uint64(len(content)))
	require.NoError(t, err)

	info, err := s.Stat("statfile")
	require.NoError(t, err)
	assert.Equal(t, int64(len(content)), info.Size())
}

func TestLocalStorage_Remove(t *testing.T) {
	dir := t.TempDir()
	s := newLocalStorage(dir)

	content := []byte("remove me")

	err := s.Write("removefile", bytes.NewReader(content), uint64(len(content)))
	require.NoError(t, err)

	err = s.Remove("removefile")
	require.NoError(t, err)

	_, err = s.Stat("removefile")
	require.Error(t, err)
	assert.True(t, os.IsNotExist(err))
}

func TestLocalStorage_MkdirAll(t *testing.T) {
	dir := t.TempDir()
	s := newLocalStorage(dir)

	err := s.MkdirAll(filepath.Join("a", "b", "c"), 0755)
	require.NoError(t, err)

	info, err := os.Stat(filepath.Join(dir, "a", "b", "c"))
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestLocalStorage_Ensure(t *testing.T) {
	t.Run("creates a missing base path", func(t *testing.T) {
		basePath := filepath.Join(t.TempDir(), "nested", "files")
		s := newLocalStorage(basePath)

		require.NoError(t, s.Ensure())

		info, err := os.Stat(basePath)
		require.NoError(t, err)
		assert.True(t, info.IsDir())
	})

	t.Run("is idempotent for an existing base path", func(t *testing.T) {
		s := newLocalStorage(t.TempDir())

		require.NoError(t, s.Ensure())
		require.NoError(t, s.Ensure())
	})

	t.Run("removes leftover temporary files", func(t *testing.T) {
		basePath := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(basePath, LocalTempDir), 0755))
		require.NoError(t, os.WriteFile(filepath.Join(basePath, LocalTempDir, "leftover"), []byte("partial"), 0600))
		require.NoError(t, os.WriteFile(filepath.Join(basePath, "1"), []byte("stored"), 0600))

		require.NoError(t, newLocalStorage(basePath).Ensure())

		_, err := os.Stat(filepath.Join(basePath, LocalTempDir, "leftover"))
		require.ErrorIs(t, err, os.ErrNotExist)
		_, err = os.Stat(filepath.Join(basePath, "1"))
		require.NoError(t, err)
	})

	t.Run("leaves a base path that is a file to ValidateBasePath", func(t *testing.T) {
		basePath := filepath.Join(t.TempDir(), "files")
		require.NoError(t, os.WriteFile(basePath, []byte("not a directory"), 0600))
		s := newLocalStorage(basePath)

		require.NoError(t, s.Ensure())
	})
}

func TestLocalStorage_ValidateBasePath(t *testing.T) {
	t.Run("accepts an existing directory", func(t *testing.T) {
		s := newLocalStorage(t.TempDir())

		require.NoError(t, s.ValidateBasePath())
	})

	t.Run("rejects a missing base path without creating it", func(t *testing.T) {
		basePath := filepath.Join(t.TempDir(), "files")
		s := newLocalStorage(basePath)

		require.ErrorContains(t, s.ValidateBasePath(), "failed to access file storage directory")

		_, err := os.Stat(basePath)
		require.ErrorIs(t, err, os.ErrNotExist)
	})

	t.Run("rejects a base path that is a regular file", func(t *testing.T) {
		basePath := filepath.Join(t.TempDir(), "files")
		require.NoError(t, os.WriteFile(basePath, []byte("not a directory"), 0600))
		s := newLocalStorage(basePath)

		require.ErrorContains(t, s.ValidateBasePath(), "file storage path exists but is not a directory")
	})
}
