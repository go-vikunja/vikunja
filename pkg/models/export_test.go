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
	"encoding/json"
	"os"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/notifications"
	"code.vikunja.io/api/pkg/user"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetUserDataExportStatus(t *testing.T) {
	t.Run("no export", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		status, err := GetUserDataExportStatus(s, &user.User{ID: 15})
		require.NoError(t, err)
		assert.Nil(t, status)
	})

	t.Run("with export", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		status, err := GetUserDataExportStatus(s, &user.User{ID: 1, ExportFileID: 1})
		require.NoError(t, err)
		require.NotNil(t, status)
		assert.Equal(t, int64(1), status.ID)
	})

	t.Run("export points at a missing file", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		// A dangling ExportFileID must read as "no export" rather than erroring,
		// matching the download path which 404s the same case.
		s := db.NewSession()
		defer s.Close()
		status, err := GetUserDataExportStatus(s, &user.User{ID: 15, ExportFileID: 9999})
		require.NoError(t, err)
		assert.Nil(t, status)
	})
}

func TestCleanupOldUserExports(t *testing.T) {
	t.Run("keeps fresh exports", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		fresh := &files.File{Name: "fresh", Size: 100, Created: time.Now(), CreatedByID: 2}
		_, err := s.Insert(fresh)
		require.NoError(t, err)

		// File 1 from the fixtures is from 2019 and therefore expired.
		_, err = s.ID(1).Cols("export_file_id").Update(&user.User{ExportFileID: 1})
		require.NoError(t, err)
		_, err = s.ID(2).Cols("export_file_id").Update(&user.User{ExportFileID: fresh.ID})
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		cleanupOldUserExports()

		db.AssertMissing(t, "users", map[string]interface{}{"id": 1, "export_file_id": 1})
		db.AssertMissing(t, "files", map[string]interface{}{"id": 1})
		db.AssertExists(t, "users", map[string]interface{}{"id": 2, "export_file_id": fresh.ID}, false)
		db.AssertExists(t, "files", map[string]interface{}{"id": fresh.ID}, false)
	})
}

func handleUserDataExport(t *testing.T, u *user.User) error {
	payload, err := json.Marshal(&UserDataExportRequestedEvent{User: u})
	require.NoError(t, err)
	return (&HandleUserDataExport{}).Handle(message.NewMessage("test", payload))
}

func assertBlobMissing(t *testing.T, id int64) {
	_, err := files.FileStat(&files.File{ID: id})
	assert.True(t, os.IsNotExist(err), "blob %d should be gone, stat err: %v", id, err)
}

// GHSA-7gmh-4fw6-8xc5
func TestHandleUserDataExport(t *testing.T) {
	t.Run("removes the previous export", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		files.InitTestFileFixtures(t)
		notifications.Fake()
		t.Cleanup(notifications.Unfake)

		s := db.NewSession()
		_, err := s.ID(4).Cols("export_file_id").Update(&user.User{ExportFileID: 1})
		require.NoError(t, err)
		u, err := user.GetUserByID(s, 4)
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		_ = s.Close()

		require.NoError(t, handleUserDataExport(t, u))

		db.AssertMissing(t, "users", map[string]interface{}{"id": 4, "export_file_id": 1})
		db.AssertMissing(t, "files", map[string]interface{}{"id": 1})
		assertBlobMissing(t, 1)
	})
	t.Run("removes the new blob when the export fails", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		s := db.NewSession()
		var maxID int64
		_, err := s.Table("files").Select("MAX(id)").Get(&maxID)
		require.NoError(t, err)
		_ = s.Close()

		// Notifying a user that doesn't exist fails after the blob was stored.
		err = handleUserDataExport(t, &user.User{ID: 9999, Username: "ghost"})
		require.Error(t, err)

		db.AssertMissing(t, "files", map[string]interface{}{"id": maxID + 1})
		assertBlobMissing(t, maxID+1)
	})
}
