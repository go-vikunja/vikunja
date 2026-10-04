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
	"fmt"
	"net/http"

	"code.vikunja.io/api/pkg/web"
)

// Errors of the project risk register. Codes 20001 and up.

// ErrRiskDoesNotExist represents an error where a risk does not exist.
type ErrRiskDoesNotExist struct {
	ID int64
}

// IsErrRiskDoesNotExist checks if an error is ErrRiskDoesNotExist.
func IsErrRiskDoesNotExist(err error) bool {
	_, ok := err.(ErrRiskDoesNotExist)
	return ok
}

func (err ErrRiskDoesNotExist) Error() string {
	return fmt.Sprintf("Risk does not exist [ID: %d]", err.ID)
}

// ErrCodeRiskDoesNotExist holds the unique world-error code of this error
const ErrCodeRiskDoesNotExist = 20001

// HTTPError holds the http error description
func (err ErrRiskDoesNotExist) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusNotFound,
		Code:     ErrCodeRiskDoesNotExist,
		Message:  "This risk does not exist.",
	}
}

// ErrInvalidRiskStatus represents an error where a risk status or a status filter is not known.
type ErrInvalidRiskStatus struct {
	Status string
}

// IsErrInvalidRiskStatus checks if an error is ErrInvalidRiskStatus.
func IsErrInvalidRiskStatus(err error) bool {
	_, ok := err.(ErrInvalidRiskStatus)
	return ok
}

func (err ErrInvalidRiskStatus) Error() string {
	return fmt.Sprintf("Invalid risk status [Status: %s]", err.Status)
}

// ErrCodeInvalidRiskStatus holds the unique world-error code of this error
const ErrCodeInvalidRiskStatus = 20002

// HTTPError holds the http error description
func (err ErrInvalidRiskStatus) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusBadRequest,
		Code:     ErrCodeInvalidRiskStatus,
		Message:  "The risk status must be one of open, mitigating, accepted or closed.",
	}
}

// ErrRiskOwnerHasNoAccess represents an error where the chosen owner cannot be the owner of a risk of
// the project: not an active person, or someone who cannot read the project.
type ErrRiskOwnerHasNoAccess struct {
	UserID int64
}

// IsErrRiskOwnerHasNoAccess checks if an error is ErrRiskOwnerHasNoAccess.
func IsErrRiskOwnerHasNoAccess(err error) bool {
	_, ok := err.(ErrRiskOwnerHasNoAccess)
	return ok
}

func (err ErrRiskOwnerHasNoAccess) Error() string {
	return fmt.Sprintf("The user cannot own a risk of this project [UserID: %d]", err.UserID)
}

// ErrCodeRiskOwnerHasNoAccess holds the unique world-error code of this error
const ErrCodeRiskOwnerHasNoAccess = 20003

// HTTPError holds the http error description
func (err ErrRiskOwnerHasNoAccess) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusBadRequest,
		Code:     ErrCodeRiskOwnerHasNoAccess,
		Message:  "The owner must be an active user who can read this project.",
	}
}

// ErrRiskStatusConflict represents an error where the status of a risk was changed by somebody else
// at the same moment.
type ErrRiskStatusConflict struct {
	ID int64
}

// IsErrRiskStatusConflict checks if an error is ErrRiskStatusConflict.
func IsErrRiskStatusConflict(err error) bool {
	_, ok := err.(ErrRiskStatusConflict)
	return ok
}

func (err ErrRiskStatusConflict) Error() string {
	return fmt.Sprintf("The status of the risk was changed concurrently [ID: %d]", err.ID)
}

// ErrCodeRiskStatusConflict holds the unique world-error code of this error
const ErrCodeRiskStatusConflict = 20004

// HTTPError holds the http error description
func (err ErrRiskStatusConflict) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusConflict,
		Code:     ErrCodeRiskStatusConflict,
		Message:  "The status of this risk was changed by somebody else just now. Reload it and try again.",
	}
}
