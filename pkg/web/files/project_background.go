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
	"net/http"
	"os"

	"code.vikunja.io/api/pkg/files"
)

// WriteProjectBackground streams a project's background file (its .File reader must be
// open) to the response, shared by the v1 and v2 background handlers. It does not close
// the reader; the caller owns it.
//
// The wire shape differs from WriteFileDownload on purpose and must stay byte-identical
// to v1: backgrounds are always served as image/jpg (no Content-Disposition, no
// Content-Length), with a cache-revalidation Last-Modified from the storage modtime
// rather than the file's DB Created timestamp.
func WriteProjectBackground(w http.ResponseWriter, r *http.Request, bgFile *files.File, stat os.FileInfo) {
	// Override the global no-store directive so browsers can cache background images.
	// no-cache allows caching but requires revalidation via If-Modified-Since.
	w.Header().Set("Cache-Control", "no-cache")

	if stat != nil {
		modTime := stat.ModTime().UTC()
		w.Header().Set("Last-Modified", modTime.Format(http.TimeFormat))

		if ifModSince := r.Header.Get("If-Modified-Since"); ifModSince != "" {
			if t, err := http.ParseTime(ifModSince); err == nil && !modTime.After(t) {
				w.WriteHeader(http.StatusNotModified)
				return
			}
		}
	}

	w.Header().Set("Content-Type", "image/jpg")
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, bgFile.File)
}
