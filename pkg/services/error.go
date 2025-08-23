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

package services

import (
	"fmt"
	"net/http"

	"code.vikunja.io/api/pkg/web"
)

// ErrProjectCannotBelongToAPseudoParentProject represents an error where a project cannot belong to a pseudo project
type ErrProjectCannotBelongToAPseudoParentProject struct {
	ProjectID       int64
	ParentProjectID int64
}

func (err *ErrProjectCannotBelongToAPseudoParentProject) Error() string {
	return fmt.Sprintf("Project cannot belong to a pseudo parent project [ProjectID: %d, ParentProjectID: %d]", err.ProjectID, err.ParentProjectID)
}

func (err *ErrProjectCannotBelongToAPseudoParentProject) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusPreconditionFailed,
		Message:  "This project cannot belong a dynamically generated project.",
	}
}

// ErrProjectCannotBeChildOfItself represents an error where a project cannot become a child of its own
type ErrProjectCannotBeChildOfItself struct {
	ProjectID int64
}

func (err *ErrProjectCannotBeChildOfItself) Error() string {
	return fmt.Sprintf("Project cannot be made a child of itself [ProjectID: %d]", err.ProjectID)
}

func (err *ErrProjectCannotBeChildOfItself) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusPreconditionFailed,
		Message:  "This project cannot be a child of itself.",
	}
}

// ErrProjectCannotHaveACyclicRelationship represents an error where a project cannot have a cyclic parent relationship
type ErrProjectCannotHaveACyclicRelationship struct {
	ProjectID int64
}

func (err *ErrProjectCannotHaveACyclicRelationship) Error() string {
	return fmt.Sprintf("Project cannot have a cyclic relationship [ProjectID: %d]", err.ProjectID)
}

func (err *ErrProjectCannotHaveACyclicRelationship) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusPreconditionFailed,
		Message:  "This project cannot have a cyclic relationship to a parent project.",
	}
}

// ErrProjectIdentifierIsNotUnique represents a "ErrProjectIdentifierIsNotUnique" kind of error. Used if the provided project identifier is not unique.
type ErrProjectIdentifierIsNotUnique struct {
	Identifier string
}

func (err ErrProjectIdentifierIsNotUnique) Error() string {
	return "Project identifier is not unique."
}

func (err ErrProjectIdentifierIsNotUnique) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusBadRequest,
		Message:  "A project with this identifier already exists.",
	}
}
