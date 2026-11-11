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

package migration

import (
	"strings"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func claimWithUpload(t *testing.T, userID int64, content string) *Status {
	t.Helper()
	u := getTestUser(t, userID)
	status, err := ClaimMigration(&testMigrator{name: "upload-test"}, u)
	require.NoError(t, err)
	require.NoError(t, StoreImportUpload(status, u, strings.NewReader(content), int64(len(content))))
	require.NotNil(t, status.UploadFileID)
	return status
}

func TestImportUpload(t *testing.T) {
	t.Run("round trip through the file storage", func(t *testing.T) {
		clearMigrationStatus(t)
		status := claimWithUpload(t, 1, "some export")

		persisted, err := GetMigrationStatusByID(status.ID)
		require.NoError(t, err)
		require.NotNil(t, persisted.UploadFileID)
		assert.Equal(t, *status.UploadFileID, *persisted.UploadFileID)

		file, size, closeFile, err := OpenImportUpload(persisted)
		require.NoError(t, err)
		defer closeFile()

		assert.Equal(t, int64(len("some export")), size)
		content := make([]byte, size)
		_, err = file.ReadAt(content, 0)
		require.NoError(t, err)
		assert.Equal(t, "some export", string(content))
	})

	t.Run("a status without an upload cannot be opened", func(t *testing.T) {
		_, _, _, err := OpenImportUpload(&Status{})
		require.ErrorIs(t, err, errNoImportUpload)
	})

	t.Run("remove deletes the file and forgets it", func(t *testing.T) {
		clearMigrationStatus(t)
		status := claimWithUpload(t, 1, "gone soon")
		fileID := *status.UploadFileID

		RemoveImportUpload(status.ID, status.UserID)

		db.AssertMissing(t, "files", map[string]interface{}{"id": fileID})
		after, err := GetMigrationStatusByID(status.ID)
		require.NoError(t, err)
		assert.Nil(t, after.UploadFileID)
	})

	t.Run("another user cannot remove the upload", func(t *testing.T) {
		clearMigrationStatus(t)
		status := claimWithUpload(t, 1, "not yours")

		RemoveImportUpload(status.ID, 2)

		db.AssertExists(t, "files", map[string]interface{}{"id": *status.UploadFileID}, false)
	})

	t.Run("cleanup removes uploads of finished and abandoned imports only", func(t *testing.T) {
		clearMigrationStatus(t)

		finished := claimWithUpload(t, 1, "finished")
		require.NoError(t, FinishMigration(finished))
		running := claimWithUpload(t, 2, "running")
		abandoned := claimWithUpload(t, 3, "abandoned")

		s := db.NewSession()
		_, err := s.Where("id = ?", abandoned.ID).Cols("started_at").Update(&Status{StartedAt: time.Now().Add(-time.Hour)})
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		require.NoError(t, s.Close())

		cleanupImportUploads()

		db.AssertMissing(t, "files", map[string]interface{}{"id": *finished.UploadFileID})
		db.AssertExists(t, "files", map[string]interface{}{"id": *running.UploadFileID}, false)
		db.AssertMissing(t, "files", map[string]interface{}{"id": *abandoned.UploadFileID})
	})
}
