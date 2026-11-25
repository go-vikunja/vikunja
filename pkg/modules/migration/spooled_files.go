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
	"io"
	"net/http"
	"os"

	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/models"
)

// SpooledFiles is a FileProvider which keeps downloaded attachments on disk, not in memory.
// Downloading up front keeps the slow requests out of the insert transaction.
type SpooledFiles struct {
	dir         string
	attachments map[*models.TaskAttachment]spooledFile
}

type spooledFile struct {
	path string
	size int64
}

func NewSpooledFiles() (*SpooledFiles, error) {
	dir, err := os.MkdirTemp("", "vikunja-import-*")
	if err != nil {
		return nil, &ErrSpoolFailed{Err: err}
	}
	return &SpooledFiles{
		dir:         dir,
		attachments: make(map[*models.TaskAttachment]spooledFile),
	}, nil
}

func (s *SpooledFiles) Close() error {
	return os.RemoveAll(s.dir)
}

// DownloadAttachment returns files.ErrFileIsTooLarge for a file above files.maxsize.
func (s *SpooledFiles) DownloadAttachment(hc *http.Client, url string, headers http.Header, name, mime string) (*models.TaskAttachment, error) {
	f, err := os.CreateTemp(s.dir, "attachment-*")
	if err != nil {
		return nil, &ErrSpoolFailed{Err: err}
	}

	size, err := downloadTo(spoolWriter{f: f}, hc, url, headers, MaxDownloadSize())
	if closeErr := f.Close(); err == nil && closeErr != nil {
		err = &ErrSpoolFailed{Err: closeErr}
	}
	if err != nil {
		_ = os.Remove(f.Name())
		return nil, err
	}

	attachment := &models.TaskAttachment{
		File: &files.File{
			Name: name,
			Mime: mime,
			Size: uint64(size), //nolint:gosec // size is never negative
		},
	}
	// Empty files are skipped, like InsertFromStructure does without a provider.
	if size == 0 {
		_ = os.Remove(f.Name())
		return attachment, nil
	}
	s.attachments[attachment] = spooledFile{path: f.Name(), size: size}
	return attachment, nil
}

// spoolWriter marks file write errors, which io.Copy would return like body read errors.
type spoolWriter struct {
	f *os.File
}

func (w spoolWriter) Write(p []byte) (int, error) {
	n, err := w.f.Write(p)
	if err != nil {
		return n, &ErrSpoolFailed{Err: err}
	}
	return n, nil
}

func (s *SpooledFiles) OpenAttachment(attachment *models.TaskAttachment) (io.ReadSeekCloser, int64, error) {
	spooled, has := s.attachments[attachment]
	if !has {
		return nil, 0, nil
	}

	f, err := os.Open(spooled.path)
	if err != nil {
		return nil, 0, &ErrSpoolFailed{Err: err}
	}
	return RemoveOnClose{File: f}, spooled.size, nil
}

func (s *SpooledFiles) OpenBackground(_ *models.ProjectWithTasksAndBuckets) (io.ReadSeekCloser, int64, error) {
	return nil, 0, nil
}

type RemoveOnClose struct {
	*os.File
}

func (f RemoveOnClose) Close() error {
	err := f.File.Close()
	_ = os.Remove(f.Name())
	return err
}
