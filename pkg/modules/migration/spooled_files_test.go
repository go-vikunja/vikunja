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
	"bytes"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func serveFile(t *testing.T, content []byte) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(content)
	}))
	t.Cleanup(srv.Close)
	return srv.URL
}

func spooledFileCount(t *testing.T, spool *SpooledFiles) int {
	t.Helper()
	entries, err := os.ReadDir(spool.dir)
	require.NoError(t, err)
	return len(entries)
}

func TestSpooledFiles(t *testing.T) {
	hc := utils.NewSSRFSafeHTTPClient()

	t.Run("skips files above files.maxsize", func(t *testing.T) {
		require.NoError(t, config.SetMaxFileSizeMBytesFromString("1MB"))
		t.Cleanup(func() { _ = config.SetMaxFileSizeMBytesFromString(config.FilesMaxSize.GetString()) })
		spool, err := NewSpooledFiles()
		require.NoError(t, err)
		t.Cleanup(func() { _ = spool.Close() })

		_, err = spool.DownloadAttachment(hc, serveFile(t, bytes.Repeat([]byte("x"), 2*1024*1024)), nil, "big", "")
		assert.True(t, files.IsErrFileIsTooLarge(err), "got %v", err)
		assert.Equal(t, 0, spooledFileCount(t, spool))
	})

	t.Run("does not open empty attachments", func(t *testing.T) {
		spool, err := NewSpooledFiles()
		require.NoError(t, err)
		t.Cleanup(func() { _ = spool.Close() })

		empty, err := spool.DownloadAttachment(hc, serveFile(t, nil), nil, "empty", "")
		require.NoError(t, err)
		assert.Equal(t, 0, spooledFileCount(t, spool))
		file, _, err := spool.OpenAttachment(empty)
		require.NoError(t, err)
		assert.Nil(t, file)
	})

	t.Run("does not report an aborted download as a spool error", func(t *testing.T) {
		spool, err := NewSpooledFiles()
		require.NoError(t, err)
		t.Cleanup(func() { _ = spool.Close() })
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Length", "100")
			_, _ = w.Write([]byte("short"))
		}))
		t.Cleanup(srv.Close)

		_, err = spool.DownloadAttachment(hc, srv.URL, nil, "file.txt", "text/plain")
		require.Error(t, err)
		assert.False(t, IsErrSpoolFailed(err), "got %v", err)
		assert.Equal(t, 0, spooledFileCount(t, spool))
	})

	t.Run("reports a failed spool file write as a spool error", func(t *testing.T) {
		f, err := os.Create(filepath.Join(t.TempDir(), "spool"))
		require.NoError(t, err)
		require.NoError(t, f.Close())

		_, err = spoolWriter{f: f}.Write([]byte("x"))
		assert.True(t, IsErrSpoolFailed(err), "got %v", err)
	})

	t.Run("keeps the spool path out of the error message", func(t *testing.T) {
		spool, err := NewSpooledFiles()
		require.NoError(t, err)
		require.NoError(t, spool.Close())

		_, err = spool.DownloadAttachment(hc, serveFile(t, []byte("content")), nil, "file.txt", "text/plain")
		require.EqualError(t, err, "could not store a downloaded attachment on disk: no such file or directory")
		assert.ErrorIs(t, err, fs.ErrNotExist)
	})

	t.Run("keeps the spool path out of the open error message", func(t *testing.T) {
		spool, err := NewSpooledFiles()
		require.NoError(t, err)
		t.Cleanup(func() { _ = spool.Close() })

		attachment, err := spool.DownloadAttachment(hc, serveFile(t, []byte("content")), nil, "file.txt", "text/plain")
		require.NoError(t, err)
		entries, err := os.ReadDir(spool.dir)
		require.NoError(t, err)
		for _, entry := range entries {
			require.NoError(t, os.Remove(filepath.Join(spool.dir, entry.Name())))
		}

		_, _, err = spool.OpenAttachment(attachment)
		require.EqualError(t, err, "could not store a downloaded attachment on disk: no such file or directory")
	})

	t.Run("imports as a file provider", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		files.InitTestFileFixtures(t)
		spool, err := NewSpooledFiles()
		require.NoError(t, err)
		t.Cleanup(func() { _ = spool.Close() })

		attachment, err := spool.DownloadAttachment(hc, serveFile(t, []byte("content")), nil, "file.txt", "text/plain")
		require.NoError(t, err)
		attachment.ID = 42
		task := &models.TaskWithComments{Task: models.Task{
			Title:                  "spooled",
			Attachments:            []*models.TaskAttachment{attachment},
			CoverImageAttachmentID: 42,
		}}
		structure := []*models.ProjectWithTasksAndBuckets{{
			Project: models.Project{Title: "spooled attachments"},
			Tasks:   []*models.TaskWithComments{task},
		}}

		require.NoError(t, InsertFromStructureWithFileProvider(structure, &user.User{ID: 1}, spool))

		require.NotZero(t, attachment.ID)
		assert.Equal(t, attachment.ID, task.CoverImageAttachmentID)
		db.AssertExists(t, "task_attachments", map[string]interface{}{
			"id":      attachment.ID,
			"task_id": task.ID,
		}, false)

		stored := &files.File{ID: attachment.FileID}
		require.NoError(t, stored.LoadFileByID())
		t.Cleanup(func() { _ = stored.File.Close() })
		content, err := io.ReadAll(stored.File)
		require.NoError(t, err)
		assert.Equal(t, "content", string(content))
		assert.Equal(t, 0, spooledFileCount(t, spool))
	})
}
