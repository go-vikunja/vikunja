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

// Errors of the "create a task and assign it by email" integration endpoint.

// ErrAssigneeEmailNotFound is returned when no active user has the email address.
type ErrAssigneeEmailNotFound struct{}

// IsErrAssigneeEmailNotFound checks if an error is ErrAssigneeEmailNotFound.
func IsErrAssigneeEmailNotFound(err error) bool {
	_, ok := err.(ErrAssigneeEmailNotFound)
	return ok
}

func (err ErrAssigneeEmailNotFound) Error() string {
	return "No user has this email address"
}

// ErrCodeAssigneeEmailNotFound holds the unique world-error code of this error
const ErrCodeAssigneeEmailNotFound = 4040

// HTTPError holds the http error description
func (err ErrAssigneeEmailNotFound) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusNotFound,
		Code:     ErrCodeAssigneeEmailNotFound,
		Message:  "No user with this email address exists.",
	}
}

// ErrAssigneeNotActive is returned when the user exists but cannot be assigned tasks.
type ErrAssigneeNotActive struct{}

// IsErrAssigneeNotActive checks if an error is ErrAssigneeNotActive.
func IsErrAssigneeNotActive(err error) bool {
	_, ok := err.(ErrAssigneeNotActive)
	return ok
}

func (err ErrAssigneeNotActive) Error() string {
	return "The user with this email address is not active"
}

// ErrCodeAssigneeNotActive holds the unique world-error code of this error
const ErrCodeAssigneeNotActive = 4041

// HTTPError holds the http error description
func (err ErrAssigneeNotActive) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusConflict,
		Code:     ErrCodeAssigneeNotActive,
		Message:  "The user with this email address is disabled, locked or not confirmed and cannot be assigned tasks.",
	}
}

// ErrAssigneeAmbiguous is returned when more than one user has the email address.
type ErrAssigneeAmbiguous struct{}

// IsErrAssigneeAmbiguous checks if an error is ErrAssigneeAmbiguous.
func IsErrAssigneeAmbiguous(err error) bool {
	_, ok := err.(ErrAssigneeAmbiguous)
	return ok
}

func (err ErrAssigneeAmbiguous) Error() string {
	return "More than one user has this email address"
}

// ErrCodeAssigneeAmbiguous holds the unique world-error code of this error
const ErrCodeAssigneeAmbiguous = 4042

// HTTPError holds the http error description
func (err ErrAssigneeAmbiguous) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusConflict,
		Code:     ErrCodeAssigneeAmbiguous,
		Message:  "More than one user has this email address, so the assignee cannot be told apart.",
	}
}

// ErrNoUsableInbox is returned when no project was given and the assignee has no inbox the task
// can be put into.
type ErrNoUsableInbox struct {
	UserID int64
}

// IsErrNoUsableInbox checks if an error is ErrNoUsableInbox.
func IsErrNoUsableInbox(err error) bool {
	_, ok := err.(ErrNoUsableInbox)
	return ok
}

func (err ErrNoUsableInbox) Error() string {
	return "The assignee has no usable inbox project"
}

// ErrCodeNoUsableInbox holds the unique world-error code of this error
const ErrCodeNoUsableInbox = 4043

// HTTPError holds the http error description
func (err ErrNoUsableInbox) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusConflict,
		Code:     ErrCodeNoUsableInbox,
		Message:  "The assignee's default project is not an inbox they own, or it is archived. Pass a project_id instead.",
	}
}

// ErrExternalIDConflict is returned when an external id was used before for a different assignee.
type ErrExternalIDConflict struct{}

// IsErrExternalIDConflict checks if an error is ErrExternalIDConflict.
func IsErrExternalIDConflict(err error) bool {
	_, ok := err.(ErrExternalIDConflict)
	return ok
}

func (err ErrExternalIDConflict) Error() string {
	return "The external id was already used for a different assignee"
}

// ErrCodeExternalIDConflict holds the unique world-error code of this error
const ErrCodeExternalIDConflict = 4044

// HTTPError holds the http error description
func (err ErrExternalIDConflict) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusConflict,
		Code:     ErrCodeExternalIDConflict,
		Message:  "This external_id was already used to create a task for a different assignee.",
	}
}
