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
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/models"

	"github.com/stretchr/testify/assert"
)

func TestBuildUploadResult(t *testing.T) {
	t.Run("maps a domain error to its numeric code", func(t *testing.T) {
		// ErrTaskAttachmentIsTooLarge is an HTTPErrorProcessor, so its Code must surface.
		r := BuildUploadResult(nil, []error{models.ErrTaskAttachmentIsTooLarge{Size: 99}})
		assert.Empty(t, r.Success)
		if assert.Len(t, r.Errors, 1) {
			assert.Equal(t, models.ErrCodeTaskAttachmentIsTooLarge, r.Errors[0].Code)
			assert.NotEmpty(t, r.Errors[0].Message)
		}
	})

	t.Run("maps a wrapped domain error to its numeric code", func(t *testing.T) {
		err := fmt.Errorf("could not upload file: %w", models.ErrTaskAttachmentIsTooLarge{Size: 99})
		r := BuildUploadResult(nil, []error{err})
		if assert.Len(t, r.Errors, 1) {
			assert.Equal(t, models.ErrCodeTaskAttachmentIsTooLarge, r.Errors[0].Code)
		}
	})

	t.Run("plain error has no code, just the message", func(t *testing.T) {
		r := BuildUploadResult(nil, []error{errors.New("boom")})
		if assert.Len(t, r.Errors, 1) {
			assert.Zero(t, r.Errors[0].Code)
			assert.Equal(t, "boom", r.Errors[0].Message)
		}
	})

	t.Run("preserves success and failure order", func(t *testing.T) {
		success := []*models.TaskAttachment{{ID: 1}, {ID: 2}}
		r := BuildUploadResult(success, []error{errors.New("first"), errors.New("second")})
		assert.Equal(t, success, r.Success)
		if assert.Len(t, r.Errors, 2) {
			assert.Equal(t, "first", r.Errors[0].Message)
			assert.Equal(t, "second", r.Errors[1].Message)
		}
	})
}

type trackedCloser struct {
	io.Reader
	closed int
}

func (c *trackedCloser) Close() error {
	c.closed++
	return nil
}

func TestWriteAttachmentDownload(t *testing.T) {
	t.Run("preview", func(t *testing.T) {
		content := &trackedCloser{Reader: bytes.NewReader([]byte("preview"))}
		rec := httptest.NewRecorder()

		WriteAttachmentDownload(rec, httptest.NewRequest(http.MethodGet, "/", nil), &models.TaskAttachmentPreview{
			ReadCloser: content,
			Mime:       "image/jpeg",
			Size:       7,
		})

		assert.Equal(t, "preview", rec.Body.String())
		assert.Equal(t, "image/jpeg", rec.Header().Get("Content-Type"))
		assert.Equal(t, "7", rec.Header().Get("Content-Length"))
		assert.Equal(t, "no-cache", rec.Header().Get("Cache-Control"))
		assert.Equal(t, 1, content.closed)
	})

	t.Run("original", func(t *testing.T) {
		content := &trackedCloser{Reader: bytes.NewReader([]byte("original"))}
		rec := httptest.NewRecorder()

		WriteAttachmentDownload(rec, httptest.NewRequest(http.MethodGet, "/", nil), &models.TaskAttachmentOriginal{
			File: &files.File{
				Name: "original.txt",
				Mime: "text/plain",
				Size: 8,
				File: content,
			},
		})

		assert.Equal(t, "original", rec.Body.String())
		assert.Equal(t, "text/plain", rec.Header().Get("Content-Type"))
		assert.Contains(t, rec.Header().Get("Content-Disposition"), "original.txt")
		assert.Equal(t, 1, content.closed)
	})
}
