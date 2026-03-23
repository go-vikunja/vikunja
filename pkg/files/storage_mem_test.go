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
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemStorage_WriteAndOpen(t *testing.T) {
	s := newMemStorage()
	content := []byte("hello world")

	err := s.Write("test/file.txt", bytes.NewReader(content), uint64(len(content)))
	require.NoError(t, err)

	rc, err := s.Open("test/file.txt")
	require.NoError(t, err)
	defer rc.Close()

	got, err := io.ReadAll(rc)
	require.NoError(t, err)
	assert.Equal(t, content, got)
}

func TestMemStorage_Stat(t *testing.T) {
	s := newMemStorage()
	content := []byte("hello world")

	err := s.Write("test/file.txt", bytes.NewReader(content), uint64(len(content)))
	require.NoError(t, err)

	info, err := s.Stat("test/file.txt")
	require.NoError(t, err)
	assert.Equal(t, "file.txt", info.Name())
	assert.Equal(t, int64(len(content)), info.Size())
	assert.False(t, info.IsDir())
}

func TestMemStorage_StatNotFound(t *testing.T) {
	s := newMemStorage()

	_, err := s.Stat("nonexistent")
	require.Error(t, err)
	assert.True(t, os.IsNotExist(err))
}

func TestMemStorage_Remove(t *testing.T) {
	s := newMemStorage()
	content := []byte("hello")

	err := s.Write("test/file.txt", bytes.NewReader(content), uint64(len(content)))
	require.NoError(t, err)

	err = s.Remove("test/file.txt")
	require.NoError(t, err)

	_, err = s.Open("test/file.txt")
	require.Error(t, err)
	assert.True(t, os.IsNotExist(err))
}

func TestMemStorage_OpenNotFound(t *testing.T) {
	s := newMemStorage()

	_, err := s.Open("nonexistent")
	require.Error(t, err)
	assert.True(t, os.IsNotExist(err))
}

func TestMemStorage_RemoveNotFound(t *testing.T) {
	s := newMemStorage()

	err := s.Remove("nonexistent")
	require.Error(t, err)
	assert.True(t, os.IsNotExist(err))
}

func TestMemStorage_MkdirAll(t *testing.T) {
	s := newMemStorage()
	// Should be a no-op, no error
	err := s.MkdirAll("/some/path", 0755)
	require.NoError(t, err)
}
