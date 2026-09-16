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

package imageutils

import (
	"errors"
	"fmt"
	"net/http"

	"code.vikunja.io/api/pkg/web"
)

// ErrImageTooLarge rejects hostile dimensions before decoding (GHSA-4vh2-39rq-rq8j).
type ErrImageTooLarge struct {
	Width  int
	Height int
}

func (err *ErrImageTooLarge) Error() string {
	return fmt.Sprintf("image dimensions %dx%d are invalid or exceed the maximum of %d pixels", err.Width, err.Height, MaxPixels)
}

const ErrCodeImageTooLarge = 4032

func (err *ErrImageTooLarge) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusBadRequest,
		Code:     ErrCodeImageTooLarge,
		Message:  fmt.Sprintf("Image dimensions %dx%d are invalid or exceed the maximum of %d pixels.", err.Width, err.Height, MaxPixels),
	}
}

func IsErrImageTooLarge(err error) bool {
	var e *ErrImageTooLarge
	return errors.As(err, &e)
}
