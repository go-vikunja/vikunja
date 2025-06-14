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

import "fmt"

// ErrFileDoesNotExist defines an error where a file does not exist in the db
type ErrFileDoesNotExist struct {
	FileID int64
}

// Error is the error implementation of ErrFileDoesNotExist
func (err ErrFileDoesNotExist) Error() string {
	return fmt.Sprintf("file %d does not exist", err.FileID)
}

// IsErrFileDoesNotExist checks if an error is ErrFileDoesNotExist
func IsErrFileDoesNotExist(err error) bool {
	_, ok := err.(ErrFileDoesNotExist)
	return ok
}

// ErrFileIsTooLarge defines an error where a file is larger than the configured limit
type ErrFileIsTooLarge struct {
	Size uint64
}

// Error is the error implementation of ErrFileIsTooLarge
func (err ErrFileIsTooLarge) Error() string {
	return fmt.Sprintf("file is too large [Size: %d]", err.Size)
}

// IsErrFileIsTooLarge checks if an error is ErrFileIsTooLarge
func IsErrFileIsTooLarge(err error) bool {
	_, ok := err.(ErrFileIsTooLarge)
	return ok
}

// ErrFileIsNotUnsplashFile defines an error where a file is not downloaded from unsplash.
// Used in cases whenever unsplash information about a file is requested, but the file was not downloaded from unsplash.
type ErrFileIsNotUnsplashFile struct {
	FileID int64
}

// Error is the error implementation of ErrFileIsNotUnsplashFile
func (err ErrFileIsNotUnsplashFile) Error() string {
	return fmt.Sprintf("file was not downloaded from unsplash [FileID: %d]", err.FileID)
}

// IsErrFileIsNotUnsplashFile checks if an error is ErrFileIsNotUnsplashFile
func IsErrFileIsNotUnsplashFile(err error) bool {
	_, ok := err.(ErrFileIsNotUnsplashFile)
	return ok
}
