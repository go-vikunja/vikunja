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

package models

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskAttachment_ReadOne(t *testing.T) {
	u := &user.User{ID: 1}

	t.Run("Normal File", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		files.InitTestFileFixtures(t)
		ta := &TaskAttachment{
			ID: 1,
		}
		err := ta.ReadOne(s, u)
		require.NoError(t, err)
		assert.NotNil(t, ta.File)
		assert.True(t, ta.File.ID == ta.FileID && ta.FileID != 0)

		// Load the actual attachment file and check its content
		err = ta.File.LoadFileByID()
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(config.ServiceRootpath.GetString(), "files", "1"), ta.File.File.Name())
		content := make([]byte, 9)
		read, err := ta.File.File.Read(content)
		require.NoError(t, err)
		assert.Equal(t, 9, read)
		assert.Equal(t, []byte("testfile1"), content)
	})
	t.Run("Nonexisting Attachment", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		files.InitTestFileFixtures(t)
		ta := &TaskAttachment{
			ID: 9999,
		}
		err := ta.ReadOne(s, u)
		require.Error(t, err)
		assert.True(t, IsErrTaskAttachmentDoesNotExist(err))
	})
	t.Run("Existing Attachment, Nonexisting File", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		files.InitTestFileFixtures(t)
		ta := &TaskAttachment{
			ID: 2,
		}
		err := ta.ReadOne(s, u)
		require.Error(t, err)
		require.EqualError(t, err, "file 9999 does not exist")
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
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	files.InitTestFileFixtures(t)
	// Assert the file is being stored correctly
	ta := TaskAttachment{
		TaskID: 1,
	}
	tf := &testfile{
		content: []byte("testingstuff"),
	}
	testuser := &user.User{ID: 1}

	err := ta.NewAttachment(s, tf, "testfile", 100, testuser)
	require.NoError(t, err)
	assert.NotEqual(t, 0, ta.FileID)
	_, err = files.FileStat(ta.File)
	require.NoError(t, err)
	assert.False(t, os.IsNotExist(err))
	assert.Equal(t, testuser.ID, ta.CreatedByID)

	// Check the file was inserted correctly
	ta.File = &files.File{ID: ta.FileID}
	err = ta.File.LoadFileMetaByID()
	require.NoError(t, err)
	assert.Equal(t, testuser.ID, ta.File.CreatedByID)
	assert.Equal(t, "testfile", ta.File.Name)
	assert.Equal(t, uint64(100), ta.File.Size)

	// Extra test for max size test
}

func TestTaskAttachment_ReadAll(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	files.InitTestFileFixtures(t)
	ta := &TaskAttachment{TaskID: 1}
	as, _, _, err := ta.ReadAll(s, &user.User{ID: 1}, "", 0, 50)
	attachments, _ := as.([]*TaskAttachment)
	require.NoError(t, err)
	assert.Len(t, attachments, 3)
	assert.Equal(t, "test", attachments[0].File.Name)
	for _, a := range attachments {
		assert.NotNil(t, a.CreatedBy)
	}
	assert.Equal(t, int64(-2), attachments[2].CreatedByID)
	assert.Equal(t, int64(-2), attachments[2].CreatedBy.ID)
}

func TestTaskAttachment_Delete(t *testing.T) {
	u := &user.User{ID: 1}

	t.Run("Normal", func(t *testing.T) {
		files.InitTestFileFixtures(t)
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		ta := &TaskAttachment{ID: 1}
		err := ta.Delete(s, u)
		require.NoError(t, err)
		// Check if the file itself was deleted
		_, err = files.FileStat(ta.File) // The new file has the id 2 since it's the second attachment
		require.Error(t, err)
		assert.True(t, os.IsNotExist(err))
	})
	t.Run("Nonexisting", func(t *testing.T) {
		files.InitTestFileFixtures(t)
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		ta := &TaskAttachment{ID: 9999}
		err := ta.Delete(s, u)
		require.Error(t, err)
		assert.True(t, IsErrTaskAttachmentDoesNotExist(err))
	})
	t.Run("Existing attachment, nonexisting file", func(t *testing.T) {
		files.InitTestFileFixtures(t)
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		ta := &TaskAttachment{ID: 2}
		err := ta.Delete(s, u)
		require.NoError(t, err)
	})
}

func TestTaskAttachment_Permissions(t *testing.T) {
	u := &user.User{ID: 1}
	t.Run("Can Read", func(t *testing.T) {
		t.Run("Allowed", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			ta := &TaskAttachment{TaskID: 1}
			can, _, err := ta.CanRead(s, u)
			require.NoError(t, err)
			assert.True(t, can)
		})
		t.Run("Forbidden", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			ta := &TaskAttachment{TaskID: 14}
			can, _, err := ta.CanRead(s, u)
			require.NoError(t, err)
			assert.False(t, can)
		})
	})
	t.Run("Can Delete", func(t *testing.T) {
		t.Run("Allowed", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			ta := &TaskAttachment{TaskID: 1}
			can, err := ta.CanDelete(s, u)
			require.NoError(t, err)
			assert.True(t, can)
		})
		t.Run("Forbidden, no access", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			ta := &TaskAttachment{TaskID: 14}
			can, err := ta.CanDelete(s, u)
			require.NoError(t, err)
			assert.False(t, can)
		})
		t.Run("Forbidden, shared read only", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			ta := &TaskAttachment{TaskID: 15}
			can, err := ta.CanDelete(s, u)
			require.NoError(t, err)
			assert.False(t, can)
		})
	})
	t.Run("Can Create", func(t *testing.T) {
		t.Run("Allowed", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			ta := &TaskAttachment{TaskID: 1}
			can, err := ta.CanCreate(s, u)
			require.NoError(t, err)
			assert.True(t, can)
		})
		t.Run("Forbidden, no access", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			ta := &TaskAttachment{TaskID: 14}
			can, err := ta.CanCreate(s, u)
			require.NoError(t, err)
			assert.False(t, can)
		})
		t.Run("Forbidden, shared read only", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			ta := &TaskAttachment{TaskID: 15}
			can, err := ta.CanCreate(s, u)
			require.NoError(t, err)
			assert.False(t, can)
		})
	})
}
