// Copyright 2019 Vikunja and contriubtors. All rights reserved.
//
// This file is part of Vikunja.
//
// Vikunja is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Vikunja is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Vikunja.  If not, see <https://www.gnu.org/licenses/>.

package files

import (
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"testing"
)

type testfile struct {
	content []byte
	done    bool
}

func (t *testfile) Read(p []byte) (n int, err error) {
	if t.done {
		return 0, io.EOF
	}
	copy(p, t.content)
	t.done = true
	return len(p), nil
}

func (t *testfile) Close() error {
	return nil
}

type testauth struct {
	id int64
}

func (a *testauth) GetID() int64 {
	return a.id
}

func TestCreate(t *testing.T) {
	t.Run("Normal", func(t *testing.T) {
		initFixtures(t)
		tf := &testfile{
			content: []byte("testfile"),
		}
		ta := &testauth{id: 1}
		_, err := Create(tf, "testfile", 100, ta)
		assert.NoError(t, err)

		// Check the file was created correctly
		file := &File{ID: 2}
		err = file.LoadFileMetaByID()
		assert.NoError(t, err)
		assert.Equal(t, int64(1), file.CreatedByID)
		assert.Equal(t, "testfile", file.Name)
		assert.Equal(t, uint64(100), file.Size)

	})
	t.Run("Too Large", func(t *testing.T) {
		initFixtures(t)
		tf := &testfile{
			content: []byte("testfile"),
		}
		ta := &testauth{id: 1}
		_, err := Create(tf, "testfile", 99999999999, ta)
		assert.Error(t, err)
		assert.True(t, IsErrFileIsTooLarge(err))
	})
}

func TestFile_Delete(t *testing.T) {
	t.Run("Normal", func(t *testing.T) {
		initFixtures(t)
		f := &File{ID: 1}
		err := f.Delete()
		assert.NoError(t, err)
	})
	t.Run("Nonexisting", func(t *testing.T) {
		initFixtures(t)
		f := &File{ID: 9999}
		err := f.Delete()
		assert.Error(t, err)
		assert.True(t, IsErrFileDoesNotExist(err))
	})
}

func TestFile_LoadFileByID(t *testing.T) {
	t.Run("Normal", func(t *testing.T) {
		initFixtures(t)
		f := &File{ID: 1}
		err := f.LoadFileByID()
		assert.NoError(t, err)
	})
	t.Run("Nonexisting", func(t *testing.T) {
		initFixtures(t)
		f := &File{ID: 9999}
		err := f.LoadFileByID()
		assert.Error(t, err)
		assert.True(t, os.IsNotExist(err))
	})
}

func TestFile_LoadFileMetaByID(t *testing.T) {
	t.Run("Normal", func(t *testing.T) {
		initFixtures(t)
		f := &File{ID: 1}
		err := f.LoadFileMetaByID()
		assert.NoError(t, err)
		assert.Equal(t, "test", f.Name)
	})
	t.Run("Nonexisting", func(t *testing.T) {
		initFixtures(t)
		f := &File{ID: 9999}
		err := f.LoadFileMetaByID()
		assert.Error(t, err)
		assert.True(t, IsErrFileDoesNotExist(err))
	})
}
