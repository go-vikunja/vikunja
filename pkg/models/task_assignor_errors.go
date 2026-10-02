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

package models

import (
	"net/http"

	"code.vikunja.io/api/pkg/web"
)

// Errors about the assignor (the user a task is stored as created by).

// ErrAssignorNotActive is returned when the assignor exists but is disabled, locked or unconfirmed.
type ErrAssignorNotActive struct{}

// IsErrAssignorNotActive checks if an error is ErrAssignorNotActive.
func IsErrAssignorNotActive(err error) bool {
	_, ok := err.(ErrAssignorNotActive)
	return ok
}

func (err ErrAssignorNotActive) Error() string {
	return "The assignor is not active"
}

// ErrCodeAssignorNotActive holds the unique world-error code of this error
const ErrCodeAssignorNotActive = 4045

// HTTPError holds the http error description
func (err ErrAssignorNotActive) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusConflict,
		Code:     ErrCodeAssignorNotActive,
		Message:  "The assignor is disabled, locked or not confirmed, or the project owner used as the default assignor is a bot. Pass an assignor_email.",
	}
}

// ErrAssignorNotFound is returned when no user has the assignor's email address.
type ErrAssignorNotFound struct{}

// IsErrAssignorNotFound checks if an error is ErrAssignorNotFound.
func IsErrAssignorNotFound(err error) bool {
	_, ok := err.(ErrAssignorNotFound)
	return ok
}

func (err ErrAssignorNotFound) Error() string {
	return "No user has the assignor's email address"
}

// ErrCodeAssignorNotFound holds the unique world-error code of this error
const ErrCodeAssignorNotFound = 4046

// HTTPError holds the http error description
func (err ErrAssignorNotFound) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusNotFound,
		Code:     ErrCodeAssignorNotFound,
		Message:  "No user with the assignor's email address exists.",
	}
}

// ErrAssignorAmbiguous is returned when more than one user has the assignor's email address.
type ErrAssignorAmbiguous struct{}

// IsErrAssignorAmbiguous checks if an error is ErrAssignorAmbiguous.
func IsErrAssignorAmbiguous(err error) bool {
	_, ok := err.(ErrAssignorAmbiguous)
	return ok
}

func (err ErrAssignorAmbiguous) Error() string {
	return "More than one user has the assignor's email address"
}

// ErrCodeAssignorAmbiguous holds the unique world-error code of this error
const ErrCodeAssignorAmbiguous = 4047

// HTTPError holds the http error description
func (err ErrAssignorAmbiguous) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusConflict,
		Code:     ErrCodeAssignorAmbiguous,
		Message:  "More than one user has the assignor's email address, so the assignor cannot be told apart.",
	}
}
