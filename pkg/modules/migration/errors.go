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
	"net/http"

	"code.vikunja.io/api/pkg/web"
)

// ErrNotAZipFile represents a "ErrNotAZipFile" kind of error.
type ErrNotAZipFile struct{}

func (err *ErrNotAZipFile) Error() string {
	return "The provided file is not a valid zip file"
}

// ErrCodeNotAZipFile holds the unique world-error code of this error
const ErrCodeNotAZipFile = 14001

// HTTPError holds the http error description
func (err *ErrNotAZipFile) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusBadRequest,
		Code:     ErrCodeNotAZipFile,
		Message:  "The provided file is not a valid zip file.",
	}
}

// ErrFileIsEmpty represents a "ErrFileIsEmpty" kind of error.
type ErrFileIsEmpty struct{}

func (err *ErrFileIsEmpty) Error() string {
	return "The provided file does not contain any data"
}

// ErrCodeFileIsEmpty holds the unique world-error code of this error
const ErrCodeFileIsEmpty = 14002

// HTTPError holds the http error description
func (err *ErrFileIsEmpty) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusBadRequest,
		Code:     ErrCodeFileIsEmpty,
		Message:  "The provided file does not contain any data.",
	}
}

// ErrNotACSVFile represents a "ErrNotACSVFile" kind of error.
type ErrNotACSVFile struct{}

func (err *ErrNotACSVFile) Error() string {
	return "The provided file is not a valid CSV file"
}

// ErrCodeNotACSVFile holds the unique world-error code of this error
const ErrCodeNotACSVFile = 14003

// HTTPError holds the http error description
func (err *ErrNotACSVFile) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusBadRequest,
		Code:     ErrCodeNotACSVFile,
		Message:  "The provided file is not a valid CSV file.",
	}
}
