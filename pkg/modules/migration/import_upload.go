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
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/cron"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/user"

	"xorm.io/xorm"
)

var errNoImportUpload = errors.New("the migration has no stored upload")

// StoreImportUpload keeps the upload in the configured file storage rather than on this instance's
// disk, so whichever instance picks up the queued import can read it.
func StoreImportUpload(status *Status, u *user.User, src io.ReaderAt, size int64) error {
	s := db.NewSession()
	defer s.Close()

	stored, err := files.CreateWithMimeAndSession(s, io.NewSectionReader(src, 0, size), status.MigratorName+"-import", 0, u, "application/octet-stream", false)
	if err != nil {
		_ = s.Rollback()
		if stored != nil && stored.ID != 0 {
			removeOrphanedBlob(stored.ID)
		}
		return fmt.Errorf("could not store the import upload: %w", err)
	}

	status.UploadFileID = &stored.ID
	if _, err := s.Where("id = ?", status.ID).Cols("upload_file_id").Update(status); err != nil {
		_ = s.Rollback()
		removeOrphanedBlob(stored.ID)
		return fmt.Errorf("could not record the import upload: %w", err)
	}
	if err := s.Commit(); err != nil {
		removeOrphanedBlob(stored.ID)
		return fmt.Errorf("could not record the import upload: %w", err)
	}
	return nil
}

// A rolled-back file row leaves its blob behind, since storage writes happen before commit.
func removeOrphanedBlob(fileID int64) {
	if err := files.DeleteBlob(fileID); err != nil && !errors.Is(err, fs.ErrNotExist) {
		log.Errorf("[Migration] Could not remove the blob of rolled back import upload %d: %s", fileID, err)
	}
}

// OpenImportUpload copies the stored upload to a local file: the zip importer needs random access,
// which an S3 object body cannot provide. closeFile removes the copy.
func OpenImportUpload(status *Status) (file io.ReaderAt, size int64, closeFile func(), err error) {
	if status.UploadFileID == nil {
		return nil, 0, nil, errNoImportUpload
	}

	stored := &files.File{ID: *status.UploadFileID}
	if err := stored.LoadFileByID(); err != nil {
		return nil, 0, nil, err
	}
	defer stored.File.Close()

	local, err := os.CreateTemp("", "vikunja-import-*")
	if err != nil {
		return nil, 0, nil, fmt.Errorf("could not create a local copy of the import upload: %w", err)
	}
	closeFile = func() {
		_ = local.Close()
		_ = os.Remove(local.Name())
	}

	size, err = io.Copy(local, stored.File)
	if err != nil {
		closeFile()
		return nil, 0, nil, fmt.Errorf("could not copy the import upload: %w", err)
	}
	return local, size, closeFile, nil
}

// RemoveImportUpload deletes the stored upload once its import is over. Failures are only logged:
// the import already succeeded or failed, and the cleanup cron retries.
func RemoveImportUpload(statusID, userID int64) {
	s := db.NewSession()
	defer s.Close()

	status := &Status{}
	has, err := s.Where("id = ? AND user_id = ?", statusID, userID).Get(status)
	if err != nil {
		log.Errorf("[Migration] Could not load migration %d to remove its upload: %s", statusID, err)
		return
	}
	if !has || status.UploadFileID == nil {
		return
	}

	if err := removeImportUpload(s, status); err != nil {
		_ = s.Rollback()
		log.Errorf("[Migration] Could not remove the upload of migration %d: %s", statusID, err)
		return
	}
	if err := s.Commit(); err != nil {
		log.Errorf("[Migration] Could not remove the upload of migration %d: %s", statusID, err)
	}
}

func removeImportUpload(s *xorm.Session, status *Status) error {
	err := (&files.File{ID: *status.UploadFileID}).Delete(s)
	if err != nil && !files.IsErrFileDoesNotExist(err) {
		return err
	}

	status.UploadFileID = nil
	_, err = s.Where("id = ?", status.ID).Cols("upload_file_id").Update(status)
	return err
}

// cleanupImportUploads removes uploads left behind by a job that never ran to its own cleanup:
// the event was lost on restart, or the instance died mid-import.
func cleanupImportUploads() {
	s := db.NewSession()
	query := s.Where("upload_file_id IS NOT NULL")
	if timeout := config.MigrationClaimTimeout.GetDuration(); timeout > 0 {
		// Same timezone binding gotcha as releaseStaleClaims.
		staleBefore := time.Now().Add(-timeout).In(config.GetTimeZone())
		query = query.And("active_user_id IS NULL OR COALESCE(heartbeat_at, started_at) < ?", staleBefore)
	} else {
		query = query.And("active_user_id IS NULL")
	}

	statuses := []*Status{}
	err := query.Find(&statuses)
	_ = s.Close()
	if err != nil {
		log.Errorf("[Migration] Could not find leftover import uploads: %s", err)
		return
	}

	for _, status := range statuses {
		RemoveImportUpload(status.ID, status.UserID)
	}
}

func RegisterImportUploadCleanupCron() {
	if err := cron.Schedule("0 * * * *", cleanupImportUploads); err != nil {
		log.Fatalf("[Migration] Could not register the import upload cleanup cron: %s", err)
	}
}
