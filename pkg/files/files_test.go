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
	"os"
	"testing"

	"code.vikunja.io/api/pkg/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		createdFile, err := Create(tf, "testfile", 100, ta)
		require.NoError(t, err)

		// Check the file was created correctly
		file := &File{ID: createdFile.ID}
		err = file.LoadFileMetaByID()
		require.NoError(t, err)
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
		require.Error(t, err)
		assert.True(t, IsErrFileIsTooLarge(err))
	})
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
		assert.True(t, os.IsNotExist(err))
	})
}

func TestFile_LoadFileMetaByID(t *testing.T) {
	t.Run("Normal", func(t *testing.T) {
		initFixtures(t)
		f := &File{ID: 1}
		err := f.LoadFileMetaByID()
		require.NoError(t, err)
		assert.Equal(t, "test", f.Name)
	})
	t.Run("Nonexisting", func(t *testing.T) {
		initFixtures(t)
		f := &File{ID: 9999}
		err := f.LoadFileMetaByID()
		require.Error(t, err)
		assert.True(t, IsErrFileDoesNotExist(err))
	})
}
