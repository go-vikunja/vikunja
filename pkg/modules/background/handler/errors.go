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

package handler

import (
	"errors"
	"fmt"
)

// ErrFileUnsupportedImageFormat defines an error where an uploaded image format is not supported
// by the imaging library
//
// This is returned when decoding the image fails because the format is unknown.
type ErrFileUnsupportedImageFormat struct {
	Mime string
}

// Error is the error implementation of ErrFileUnsupportedImageFormat
func (err ErrFileUnsupportedImageFormat) Error() string {
	return fmt.Sprintf("file is not a supported image format [Mime: %s]", err.Mime)
}

// IsErrFileUnsupportedImageFormat checks if an error is ErrFileUnsupportedImageFormat
func IsErrFileUnsupportedImageFormat(err error) bool {
	var errFileUnsupportedImageFormat ErrFileUnsupportedImageFormat
	ok := errors.As(err, &errFileUnsupportedImageFormat)
	return ok
}

// ErrFileIsNoImage is returned when an uploaded background does not sniff as an
// image at all (its detected mime type does not start with "image"). It is
// distinct from ErrFileUnsupportedImageFormat, which is a recognized image type
// the imaging library can't decode.
type ErrFileIsNoImage struct {
	Mime string
}

// Error is the error implementation of ErrFileIsNoImage
func (err ErrFileIsNoImage) Error() string {
	return fmt.Sprintf("uploaded file is not an image [Mime: %s]", err.Mime)
}

// IsErrFileIsNoImage checks if an error is ErrFileIsNoImage
func IsErrFileIsNoImage(err error) bool {
	var errFileIsNoImage ErrFileIsNoImage
	return errors.As(err, &errFileIsNoImage)
}
