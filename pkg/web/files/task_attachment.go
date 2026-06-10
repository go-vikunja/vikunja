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

// Package files holds the HTTP-layer glue for serving task attachments —
// the upload-result DTOs and the download response writer — shared by the
// v1 and v2 handlers. The domain logic stays in pkg/models; this package
// only translates it to and from the wire.
package files

import (
	"io"
	"mime"
	"net/http"
	"strconv"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/web"
)

// AttachmentUploadError is a per-file upload failure.
type AttachmentUploadError struct {
	Code    int    `json:"code,omitempty" doc:"Vikunja numeric error code, when the failure carries one."`
	Message string `json:"message" doc:"A human-readable description of why this file failed."`
}

// AttachmentUploadResult is the outcome of an attachment upload: files are
// processed independently, so a per-file failure lands in Errors while the
// rest still succeed.
type AttachmentUploadResult struct {
	Errors  []AttachmentUploadError  `json:"errors" doc:"Per-file failures. A file that fails here does not fail the whole request; the others still upload."`
	Success []*models.TaskAttachment `json:"success" doc:"The attachments that were created successfully."`
}

// BuildUploadResult turns the domain function's plain return values into the
// wire DTO, mapping each failure to its numeric code when it carries one.
func BuildUploadResult(success []*models.TaskAttachment, failures []error) *AttachmentUploadResult {
	r := &AttachmentUploadResult{Success: success}
	for _, err := range failures {
		r.Errors = append(r.Errors, toAttachmentUploadError(err))
	}
	return r
}

func toAttachmentUploadError(err error) AttachmentUploadError {
	if httpErr, ok := err.(web.HTTPErrorProcessor); ok {
		details := httpErr.HTTPError()
		return AttachmentUploadError{Code: details.Code, Message: details.Message}
	}
	return AttachmentUploadError{Message: err.Error()}
}

// WriteAttachmentDownload streams the attachment (or its preview) to the response:
// http.ServeContent for seekable local files (Range + If-Modified-Since for free),
// a manual 304 + io.Copy otherwise. It closes the file reader.
func WriteAttachmentDownload(w http.ResponseWriter, r *http.Request, ta *models.TaskAttachment, preview []byte) {
	defer func() { _ = ta.File.File.Close() }()

	if preview != nil {
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Content-Length", strconv.Itoa(len(preview)))
		_, _ = w.Write(preview)
		return
	}

	mimeToReturn := ta.File.Mime
	if mimeToReturn == "" {
		mimeToReturn = "application/octet-stream"
	}
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": ta.File.Name}))
	w.Header().Set("Content-Type", mimeToReturn)
	w.Header().Set("Content-Length", strconv.FormatUint(ta.File.Size, 10))
	w.Header().Set("Last-Modified", ta.File.Created.UTC().Format(http.TimeFormat))
	// Override the global no-store directive so browsers can cache attachments.
	w.Header().Set("Cache-Control", "no-cache")

	// Local files are *os.File (seekable), so ServeContent gives Range +
	// If-Modified-Since for free; s3 (and the in-memory test storage) return a
	// non-seekable reader, so check If-Modified-Since manually and io.Copy.
	if seeker, ok := ta.File.File.(io.ReadSeeker); ok {
		http.ServeContent(w, r, ta.File.Name, ta.File.Created, seeker)
		return
	}

	if ifModSince := r.Header.Get("If-Modified-Since"); ifModSince != "" {
		if t, parseErr := http.ParseTime(ifModSince); parseErr == nil && !ta.File.Created.UTC().After(t) {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}
	_, _ = io.Copy(w, ta.File.File)
}
