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

package models

import (
	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/files"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"strconv"
	"testing"
)

func TestTaskAttachment_ReadOne(t *testing.T) {
	t.Run("Normal File", func(t *testing.T) {
		files.InitTestFileFixtures(t)
		ta := &TaskAttachment{
			ID: 1,
		}
		err := ta.ReadOne()
		assert.NoError(t, err)
		assert.NotNil(t, ta.File)
		assert.True(t, ta.File.ID == ta.FileID && ta.FileID != 0)

		// Load the actual attachment file and check its content
		err = ta.File.LoadFileByID()
		assert.NoError(t, err)
		assert.Equal(t, config.FilesBasePath.GetString()+"/1", ta.File.File.Name())
		content := make([]byte, 9)
		read, err := ta.File.File.Read(content)
		assert.NoError(t, err)
		assert.Equal(t, 9, read)
		assert.Equal(t, []byte("testfile1"), content)
	})
	t.Run("Nonexisting Attachment", func(t *testing.T) {
		files.InitTestFileFixtures(t)
		ta := &TaskAttachment{
			ID: 9999,
		}
		err := ta.ReadOne()
		assert.Error(t, err)
		assert.True(t, IsErrTaskAttachmentDoesNotExist(err))
	})
	t.Run("Existing Attachment, Nonexisting File", func(t *testing.T) {
		files.InitTestFileFixtures(t)
		ta := &TaskAttachment{
			ID: 2,
		}
		err := ta.ReadOne()
		assert.Error(t, err)
		assert.EqualError(t, err, "file 9999 does not exist")
	})
}

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

func TestTaskAttachment_NewAttachment(t *testing.T) {
	files.InitTestFileFixtures(t)
	// Assert the file is being stored correctly
	ta := TaskAttachment{
		TaskID: 1,
	}
	tf := &testfile{
		content: []byte("testingstuff"),
	}
	testuser := &User{ID: 1}

	err := ta.NewAttachment(tf, "testfile", 100, testuser)
	assert.NoError(t, err)
	assert.NotEqual(t, 0, ta.FileID)
	_, err = files.FileStat("files/" + strconv.FormatInt(ta.FileID, 10))
	assert.NoError(t, err)
	assert.False(t, os.IsNotExist(err))
	assert.Equal(t, testuser.ID, ta.CreatedByID)

	// Check the file was inserted correctly
	ta.File = &files.File{ID: ta.FileID}
	err = ta.File.LoadFileMetaByID()
	assert.NoError(t, err)
	assert.Equal(t, testuser.ID, ta.File.CreatedByID)
	assert.Equal(t, "testfile", ta.File.Name)
	assert.Equal(t, uint64(100), ta.File.Size)

	// Extra test for max size test
}

func TestTaskAttachment_ReadAll(t *testing.T) {
	files.InitTestFileFixtures(t)
	ta := &TaskAttachment{TaskID: 1}
	as, _, _, err := ta.ReadAll(&User{ID: 1}, "", 0, 50)
	attachments, _ := as.([]*TaskAttachment)
	assert.NoError(t, err)
	assert.Len(t, attachments, 3)
	assert.Equal(t, "test", attachments[0].File.Name)
}

func TestTaskAttachment_Delete(t *testing.T) {
	files.InitTestFileFixtures(t)
	t.Run("Normal", func(t *testing.T) {
		ta := &TaskAttachment{ID: 1}
		err := ta.Delete()
		assert.NoError(t, err)
		// Check if the file itself was deleted
		_, err = files.FileStat("/1") // The new file has the id 2 since it's the second attachment
		assert.True(t, os.IsNotExist(err))
	})
	t.Run("Nonexisting", func(t *testing.T) {
		files.InitTestFileFixtures(t)
		ta := &TaskAttachment{ID: 9999}
		err := ta.Delete()
		assert.Error(t, err)
		assert.True(t, IsErrTaskAttachmentDoesNotExist(err))
	})
	t.Run("Existing attachment, nonexisting file", func(t *testing.T) {
		files.InitTestFileFixtures(t)
		ta := &TaskAttachment{ID: 2}
		err := ta.Delete()
		assert.NoError(t, err)
	})
}

func TestTaskAttachment_Rights(t *testing.T) {
	u := &User{ID: 1}
	t.Run("Can Read", func(t *testing.T) {
		t.Run("Allowed", func(t *testing.T) {
			ta := &TaskAttachment{TaskID: 1}
			can, err := ta.CanRead(u)
			assert.NoError(t, err)
			assert.True(t, can)
		})
		t.Run("Forbidden", func(t *testing.T) {
			ta := &TaskAttachment{TaskID: 14}
			can, err := ta.CanRead(u)
			assert.NoError(t, err)
			assert.False(t, can)
		})
	})
	t.Run("Can Delete", func(t *testing.T) {
		t.Run("Allowed", func(t *testing.T) {
			ta := &TaskAttachment{TaskID: 1}
			can, err := ta.CanDelete(u)
			assert.NoError(t, err)
			assert.True(t, can)
		})
		t.Run("Forbidden, no access", func(t *testing.T) {
			ta := &TaskAttachment{TaskID: 14}
			can, err := ta.CanDelete(u)
			assert.NoError(t, err)
			assert.False(t, can)
		})
		t.Run("Forbidden, shared read only", func(t *testing.T) {
			ta := &TaskAttachment{TaskID: 15}
			can, err := ta.CanDelete(u)
			assert.NoError(t, err)
			assert.False(t, can)
		})
	})
	t.Run("Can Create", func(t *testing.T) {
		t.Run("Allowed", func(t *testing.T) {
			ta := &TaskAttachment{TaskID: 1}
			can, err := ta.CanCreate(u)
			assert.NoError(t, err)
			assert.True(t, can)
		})
		t.Run("Forbidden, no access", func(t *testing.T) {
			ta := &TaskAttachment{TaskID: 14}
			can, err := ta.CanCreate(u)
			assert.NoError(t, err)
			assert.False(t, can)
		})
		t.Run("Forbidden, shared read only", func(t *testing.T) {
			ta := &TaskAttachment{TaskID: 15}
			can, err := ta.CanCreate(u)
			assert.NoError(t, err)
			assert.False(t, can)
		})
	})
}
