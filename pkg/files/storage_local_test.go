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
	"path/filepath"
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
