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
	"context"
	"io"
	"os"
)

// FileStorage abstracts file storage operations across local, S3, and in-memory backends.
type FileStorage interface {
	Open(path string) (io.ReadCloser, error)
	Write(path string, content io.ReadSeeker, size uint64) error
	Stat(path string) (os.FileInfo, error)
	Remove(path string) error
	MkdirAll(path string, perm os.FileMode) error

	// Ensure prepares the backend for use. It is the only place allowed to create storage.
	Ensure() error
	ValidateBasePath() error
}

// contextStorage is implemented by backends doing network IO, so callers can bound
// them with a deadline. Local and in-memory writes cannot block indefinitely.
type contextStorage interface {
	writeContext(ctx context.Context, path string, content io.ReadSeeker, size uint64) error
	removeContext(ctx context.Context, path string) error
}
