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
	"strings"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/web"
)

// Generic

// ErrGenericForbidden represents a "UsernameAlreadyExists" kind of error.
type ErrGenericForbidden struct{}

// IsErrGenericForbidden checks if an error is a ErrGenericForbidden.
func IsErrGenericForbidden(err error) bool {
	_, ok := err.(ErrGenericForbidden)
	return ok
}

func (err ErrGenericForbidden) Error() string {
	return "Forbidden"
}

// ErrorCodeGenericForbidden holds the unique world-error code of this error
const ErrorCodeGenericForbidden = 0001

// HTTPError holds the http error description
func (err ErrGenericForbidden) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: http.StatusForbidden, Code: ErrorCodeGenericForbidden, Message: "You're not allowed to do this."}
}

// ===================
// Empty things errors
// ===================

// ErrIDCannotBeZero represents a "IDCannotBeZero" kind of error. Used if an ID (of something, not defined) is 0 where it should not.
type ErrIDCannotBeZero struct{}

// IsErrIDCannotBeZero checks if an error is a ErrIDCannotBeZero.
func IsErrIDCannotBeZero(err error) bool {
	_, ok := err.(ErrIDCannotBeZero)
	return ok
}

func (err ErrIDCannotBeZero) Error() string {
	return "ID cannot be empty or 0"
}

// ErrCodeIDCannotBeZero holds the unique world-error code of this error
const ErrCodeIDCannotBeZero = 2001

// HTTPError holds the http error description
func (err ErrIDCannotBeZero) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: http.StatusBadRequest, Code: ErrCodeIDCannotBeZero, Message: "The ID cannot be empty or 0."}
}

// ErrInvalidData represents a "ErrInvalidData" kind of error. Used when a struct is invalid -> validation failed.
type ErrInvalidData struct {
	Message string
}

// IsErrInvalidData checks if an error is a ErrIDCannotBeZero.
func IsErrInvalidData(err error) bool {
	_, ok := err.(ErrInvalidData)
	return ok
}

func (err ErrInvalidData) Error() string {
	return fmt.Sprintf("Struct is invalid. %s", err.Message)
}

// ErrCodeInvalidData holds the unique world-error code of this error
const ErrCodeInvalidData = 2002

// HTTPError holds the http error description
func (err ErrInvalidData) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: http.StatusBadRequest, Code: ErrCodeInvalidData, Message: err.Message}
}

// ValidationHTTPError is the http error when a validation fails
type ValidationHTTPError struct {
	web.HTTPError
	InvalidFields []string `json:"invalid_fields"`
}

// Error implements the Error type (so we can return it as type error)
func (err ValidationHTTPError) Error() string {
	theErr := ErrInvalidData{
		Message: err.Message,
	}
	return theErr.Error()
}

func InvalidFieldError(fields []string) error {
	return InvalidFieldErrorWithMessage(fields, "Invalid Data")
}

func InvalidFieldErrorWithMessage(fields []string, message string) error {
	return ValidationHTTPError{
		HTTPError: web.HTTPError{
			HTTPCode: http.StatusPreconditionFailed,
			Code:     ErrCodeInvalidData,
			Message:  message,
		},
		InvalidFields: fields,
	}
}

// ErrInvalidTimezone represents a "InvalidTimezone" kind of error.
type ErrInvalidTimezone struct {
	Name      string
	LoadError error
}

// IsErrInvalidTimezone checks if an error is a ErrInvalidTimezone.
func IsErrInvalidTimezone(err error) bool {
	_, ok := err.(ErrInvalidTimezone)
	return ok
}

func (err ErrInvalidTimezone) Error() string {
	return fmt.Sprintf("invalid timezone: %s, err: %v", err.Name, err.LoadError)
}

// ErrCodeInvalidTimezone holds the unique world-error code of this error
const ErrCodeInvalidTimezone = 2003

// HTTPError holds the http error description
func (err ErrInvalidTimezone) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusBadRequest,
		Code:     ErrCodeInvalidTimezone,
		Message:  fmt.Sprintf("The timezone '%s' is invalid", err.Name),
	}
}

// ===========
// Project errors
// ===========

// ErrProjectDoesNotExist represents a "ErrProjectDoesNotExist" kind of error. Used if the project does not exist.
type ErrProjectDoesNotExist struct {
	ID int64
}

// IsErrProjectDoesNotExist checks if an error is a ErrProjectDoesNotExist.
func IsErrProjectDoesNotExist(err error) bool {
	_, ok := err.(ErrProjectDoesNotExist)
	return ok
}

func (err ErrProjectDoesNotExist) Error() string {
	return fmt.Sprintf("Project does not exist [ID: %d]", err.ID)
}

// ErrCodeProjectDoesNotExist holds the unique world-error code of this error
const ErrCodeProjectDoesNotExist = 3001

// HTTPError holds the http error description
func (err ErrProjectDoesNotExist) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: http.StatusNotFound, Code: ErrCodeProjectDoesNotExist, Message: "This project does not exist."}
}

// ErrNeedToHaveProjectReadAccess represents an error, where the user dont has read access to that Project
type ErrNeedToHaveProjectReadAccess struct {
	ProjectID int64
	UserID    int64
}

// IsErrNeedToHaveProjectReadAccess checks if an error is a ErrProjectDoesNotExist.
func IsErrNeedToHaveProjectReadAccess(err error) bool {
	_, ok := err.(ErrNeedToHaveProjectReadAccess)
	return ok
}

func (err ErrNeedToHaveProjectReadAccess) Error() string {
	return fmt.Sprintf("User needs to have read access to that project [ProjectID: %d, ID: %d]", err.ProjectID, err.UserID)
}

// ErrCodeNeedToHaveProjectReadAccess holds the unique world-error code of this error
const ErrCodeNeedToHaveProjectReadAccess = 3004

// HTTPError holds the http error description
func (err ErrNeedToHaveProjectReadAccess) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: http.StatusForbidden, Code: ErrCodeNeedToHaveProjectReadAccess, Message: "You need to have read access to this project."}
}

// ErrProjectTitleCannotBeEmpty represents a "ErrProjectTitleCannotBeEmpty" kind of error. Used if the project does not exist.
type ErrProjectTitleCannotBeEmpty struct{}

// IsErrProjectTitleCannotBeEmpty checks if an error is a ErrProjectTitleCannotBeEmpty.
func IsErrProjectTitleCannotBeEmpty(err error) bool {
	_, ok := err.(ErrProjectTitleCannotBeEmpty)
	return ok
}

func (err ErrProjectTitleCannotBeEmpty) Error() string {
	return "Project title cannot be empty."
}

// ErrCodeProjectTitleCannotBeEmpty holds the unique world-error code of this error
const ErrCodeProjectTitleCannotBeEmpty = 3005

// HTTPError holds the http error description
func (err ErrProjectTitleCannotBeEmpty) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: http.StatusBadRequest, Code: ErrCodeProjectTitleCannotBeEmpty, Message: "You must provide at least a project title."}
}

// ErrProjectShareDoesNotExist represents a "ErrProjectShareDoesNotExist" kind of error. Used if the project share does not exist.
type ErrProjectShareDoesNotExist struct {
	ID   int64
	Hash string
}

// IsErrProjectShareDoesNotExist checks if an error is a ErrProjectShareDoesNotExist.
func IsErrProjectShareDoesNotExist(err error) bool {
	_, ok := err.(ErrProjectShareDoesNotExist)
	return ok
}

func (err ErrProjectShareDoesNotExist) Error() string {
	return "Project share does not exist."
}

// ErrCodeProjectShareDoesNotExist holds the unique world-error code of this error
const ErrCodeProjectShareDoesNotExist = 3006

// HTTPError holds the http error description
func (err ErrProjectShareDoesNotExist) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: http.StatusNotFound, Code: ErrCodeProjectShareDoesNotExist, Message: "The project share does not exist."}
}

// ErrProjectIdentifierIsNotUnique represents a "ErrProjectIdentifierIsNotUnique" kind of error. Used if the provided project identifier is not unique.
type ErrProjectIdentifierIsNotUnique struct {
	Identifier string
}

// IsErrProjectIdentifierIsNotUnique checks if an error is a ErrProjectIdentifierIsNotUnique.
func IsErrProjectIdentifierIsNotUnique(err error) bool {
	_, ok := err.(ErrProjectIdentifierIsNotUnique)
	return ok
}

func (err ErrProjectIdentifierIsNotUnique) Error() string {
	return "Project identifier is not unique."
}

// ErrCodeProjectIdentifierIsNotUnique holds the unique world-error code of this error
const ErrCodeProjectIdentifierIsNotUnique = 3007

// HTTPError holds the http error description
func (err ErrProjectIdentifierIsNotUnique) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusBadRequest,
		Code:     ErrCodeProjectIdentifierIsNotUnique,
		Message:  "A project with this identifier already exists.",
	}
}

// ErrProjectIsArchived represents an error, where a project is archived
type ErrProjectIsArchived struct {
	ProjectID int64
}

// IsErrProjectIsArchived checks if an error is a project is archived error.
func IsErrProjectIsArchived(err error) bool {
	_, ok := err.(ErrProjectIsArchived)
	return ok
}

func (err ErrProjectIsArchived) Error() string {
	return fmt.Sprintf("Project is archived [ProjectID: %d]", err.ProjectID)
}

// ErrCodeProjectIsArchived holds the unique world-error code of this error
const ErrCodeProjectIsArchived = 3008

// HTTPError holds the http error description
func (err ErrProjectIsArchived) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: http.StatusPreconditionFailed, Code: ErrCodeProjectIsArchived, Message: "This project is archived. Editing or creating new tasks is not possible."}
}

// ErrProjectCannotBelongToAPseudoParentProject represents an error where a project cannot belong to a pseudo project
type ErrProjectCannotBelongToAPseudoParentProject struct {
	ProjectID       int64
	ParentProjectID int64
}

// IsErrProjectCannotBelongToAPseudoParentProject checks if an error is a project is archived error.
func IsErrProjectCannotBelongToAPseudoParentProject(err error) bool {
	_, ok := err.(*ErrProjectCannotBelongToAPseudoParentProject)
	return ok
}

func (err *ErrProjectCannotBelongToAPseudoParentProject) Error() string {
	return fmt.Sprintf("Project cannot belong to a pseudo parent project [ProjectID: %d, ParentProjectID: %d]", err.ProjectID, err.ParentProjectID)
}

// ErrCodeProjectCannotBelongToAPseudoParentProject holds the unique world-error code of this error
const ErrCodeProjectCannotBelongToAPseudoParentProject = 3009

// HTTPError holds the http error description
func (err *ErrProjectCannotBelongToAPseudoParentProject) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusPreconditionFailed,
		Code:     ErrCodeProjectCannotBelongToAPseudoParentProject,
		Message:  "This project cannot belong a dynamically generated project.",
	}
}

// ErrProjectCannotBeChildOfItself represents an error where a project cannot become a child of its own
type ErrProjectCannotBeChildOfItself struct {
	ProjectID int64
}

// IsErrProjectCannotBeChildOfItsOwn checks if an error is a project is archived error.
func IsErrProjectCannotBeChildOfItsOwn(err error) bool {
	_, ok := err.(*ErrProjectCannotBeChildOfItself)
	return ok
}

func (err *ErrProjectCannotBeChildOfItself) Error() string {
	return fmt.Sprintf("Project cannot be made a child of itself [ProjectID: %d]", err.ProjectID)
}

// ErrCodeProjectCannotBeChildOfItself holds the unique world-error code of this error
const ErrCodeProjectCannotBeChildOfItself = 3010

// HTTPError holds the http error description
func (err *ErrProjectCannotBeChildOfItself) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusPreconditionFailed,
		Code:     ErrCodeProjectCannotBeChildOfItself,
		Message:  "This project cannot be a child of itself.",
	}
}

// ErrProjectCannotHaveACyclicRelationship represents an error where a project cannot have a cyclic parent relationship
type ErrProjectCannotHaveACyclicRelationship struct {
	ProjectID int64
	CycleIDs  []int64
}

// IsErrProjectCannotHaveACyclicRelationship checks if an error is a project is archived error.
func IsErrProjectCannotHaveACyclicRelationship(err error) bool {
	_, ok := err.(*ErrProjectCannotHaveACyclicRelationship)
	return ok
}

func (err *ErrProjectCannotHaveACyclicRelationship) CycleString() string {
	var cycle string
	for _, projectID := range err.CycleIDs {
		cycle += fmt.Sprintf("%d -> ", projectID)
	}
	return strings.TrimSuffix(cycle, " -> ")
}

func (err *ErrProjectCannotHaveACyclicRelationship) Error() string {
	return fmt.Sprintf("Project cannot have a cyclic relationship [ProjectID: %d]", err.ProjectID)
}

// ErrCodeProjectCannotHaveACyclicRelationship holds the unique world-error code of this error
const ErrCodeProjectCannotHaveACyclicRelationship = 3011

// HTTPError holds the http error description
func (err *ErrProjectCannotHaveACyclicRelationship) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusPreconditionFailed,
		Code:     ErrCodeProjectCannotHaveACyclicRelationship,
		Message:  "This project cannot have a cyclic relationship to a parent project.",
	}
}

// ErrCannotDeleteDefaultProject represents an error where the default project is being deleted
type ErrCannotDeleteDefaultProject struct {
	ProjectID int64
}

// IsErrCannotDeleteDefaultProject checks if an error is a project is archived error.
func IsErrCannotDeleteDefaultProject(err error) bool {
	_, ok := err.(*ErrCannotDeleteDefaultProject)
	return ok
}

func (err *ErrCannotDeleteDefaultProject) Error() string {
	return fmt.Sprintf("Default project cannot be deleted [ProjectID: %d]", err.ProjectID)
}

// ErrCodeCannotDeleteDefaultProject holds the unique world-error code of this error
const ErrCodeCannotDeleteDefaultProject = 3012

// HTTPError holds the http error description
func (err *ErrCannotDeleteDefaultProject) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusPreconditionFailed,
		Code:     ErrCodeCannotDeleteDefaultProject,
		Message:  "This project cannot be deleted because it is the default project of a user.",
	}
}

// ErrCannotArchiveDefaultProject represents an error where the default project is being deleted
type ErrCannotArchiveDefaultProject struct {
	ProjectID int64
}

// IsErrCannotArchiveDefaultProject checks if an error is a project is archived error.
func IsErrCannotArchiveDefaultProject(err error) bool {
	_, ok := err.(*ErrCannotArchiveDefaultProject)
	return ok
}

func (err *ErrCannotArchiveDefaultProject) Error() string {
	return fmt.Sprintf("Default project cannot be archived [ProjectID: %d]", err.ProjectID)
}

// ErrCodeCannotArchiveDefaultProject holds the unique world-error code of this error
const ErrCodeCannotArchiveDefaultProject = 3013

// HTTPError holds the http error description
func (err *ErrCannotArchiveDefaultProject) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusPreconditionFailed,
		Code:     ErrCodeCannotArchiveDefaultProject,
		Message:  "This project cannot be archived because it is the default project of a user.",
	}
}

// ErrProjectViewDoesNotExist represents an error where the default project is being deleted
type ErrProjectViewDoesNotExist struct {
	ProjectViewID int64
}

// IsErrProjectViewDoesNotExist checks if an error is a project is archived error.
func IsErrProjectViewDoesNotExist(err error) bool {
	_, ok := err.(*ErrProjectViewDoesNotExist)
	return ok
}

func (err *ErrProjectViewDoesNotExist) Error() string {
	return fmt.Sprintf("Project view does not exist [ProjectViewID: %d]", err.ProjectViewID)
}

// ErrCodeProjectViewDoesNotExist holds the unique world-error code of this error
const ErrCodeProjectViewDoesNotExist = 3014

// HTTPError holds the http error description
func (err *ErrProjectViewDoesNotExist) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusNotFound,
		Code:     ErrCodeProjectViewDoesNotExist,
		Message:  "This project view does not exist.",
	}
}

// ==============
// Task errors
// ==============

// ErrTaskCannotBeEmpty represents a "ErrTaskCannotBeEmpty" kind of error.
type ErrTaskCannotBeEmpty struct{}

// IsErrTaskCannotBeEmpty checks if an error is a ErrProjectDoesNotExist.
func IsErrTaskCannotBeEmpty(err error) bool {
	_, ok := err.(ErrTaskCannotBeEmpty)
	return ok
}

func (err ErrTaskCannotBeEmpty) Error() string {
	return "Task title cannot be empty."
}

// ErrCodeTaskCannotBeEmpty holds the unique world-error code of this error
const ErrCodeTaskCannotBeEmpty = 4001

// HTTPError holds the http error description
func (err ErrTaskCannotBeEmpty) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: http.StatusBadRequest, Code: ErrCodeTaskCannotBeEmpty, Message: "You must provide at least a task title."}
}

// ErrTaskDoesNotExist represents a "ErrProjectDoesNotExist" kind of error. Used if the project does not exist.
type ErrTaskDoesNotExist struct {
	ID int64
}

// IsErrTaskDoesNotExist checks if an error is a ErrProjectDoesNotExist.
func IsErrTaskDoesNotExist(err error) bool {
	_, ok := err.(ErrTaskDoesNotExist)
	return ok
}

func (err ErrTaskDoesNotExist) Error() string {
	return fmt.Sprintf("The task does not exist. [ID: %d]", err.ID)
}

// ErrCodeTaskDoesNotExist holds the unique world-error code of this error
const ErrCodeTaskDoesNotExist = 4002

// HTTPError holds the http error description
func (err ErrTaskDoesNotExist) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: http.StatusNotFound, Code: ErrCodeTaskDoesNotExist, Message: "This task does not exist"}
}

// ErrBulkTasksMustBeInSameProject represents a "ErrBulkTasksMustBeInSameProject" kind of error.
type ErrBulkTasksMustBeInSameProject struct {
	ShouldBeID int64
	IsID       int64
}

// IsErrBulkTasksMustBeInSameProject checks if an error is a ErrBulkTasksMustBeInSameProject.
func IsErrBulkTasksMustBeInSameProject(err error) bool {
	_, ok := err.(ErrBulkTasksMustBeInSameProject)
	return ok
}

func (err ErrBulkTasksMustBeInSameProject) Error() string {
	return fmt.Sprintf("All bulk editing tasks must be in the same project. [Should be: %d, is: %d]", err.ShouldBeID, err.IsID)
}

// ErrCodeBulkTasksMustBeInSameProject holds the unique world-error code of this error
const ErrCodeBulkTasksMustBeInSameProject = 4003

// HTTPError holds the http error description
func (err ErrBulkTasksMustBeInSameProject) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: http.StatusBadRequest, Code: ErrCodeBulkTasksMustBeInSameProject, Message: "All tasks must be in the same project."}
}

// ErrBulkTasksNeedAtLeastOne represents a "ErrBulkTasksNeedAtLeastOne" kind of error.
type ErrBulkTasksNeedAtLeastOne struct{}

// IsErrBulkTasksNeedAtLeastOne checks if an error is a ErrBulkTasksNeedAtLeastOne.
func IsErrBulkTasksNeedAtLeastOne(err error) bool {
	_, ok := err.(ErrBulkTasksNeedAtLeastOne)
	return ok
}

func (err ErrBulkTasksNeedAtLeastOne) Error() string {
	return "Need at least one task when bulk editing tasks"
}

// ErrCodeBulkTasksNeedAtLeastOne holds the unique world-error code of this error
const ErrCodeBulkTasksNeedAtLeastOne = 4004

// HTTPError holds the http error description
func (err ErrBulkTasksNeedAtLeastOne) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: http.StatusBadRequest, Code: ErrCodeBulkTasksNeedAtLeastOne, Message: "Need at least one tasks to do bulk editing."}
}

// ErrNoPermissionToSeeTask represents an error where a user does not have the permission to see a task
type ErrNoPermissionToSeeTask struct {
	TaskID int64
	UserID int64
}

// IsErrNoPermissionToSeeTask checks if an error is ErrNoPermissionToSeeTask.
func IsErrNoPermissionToSeeTask(err error) bool {
	_, ok := err.(ErrNoPermissionToSeeTask)
	return ok
}

func (err ErrNoPermissionToSeeTask) Error() string {
	return fmt.Sprintf("User does not have the permission to see the task [TaskID: %v, ID: %v]", err.TaskID, err.UserID)
}

// ErrCodeNoRightToSeeTask holds the unique world-error code of this error
const ErrCodeNoRightToSeeTask = 4005

// HTTPError holds the http error description
func (err ErrNoPermissionToSeeTask) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusForbidden,
		Code:     ErrCodeNoRightToSeeTask,
		Message:  "You don't have the permission to see this task.",
	}
}

// ErrParentTaskCannotBeTheSame represents an error where the user tries to set a tasks parent as the same
type ErrParentTaskCannotBeTheSame struct {
	TaskID int64
}

// IsErrParentTaskCannotBeTheSame checks if an error is ErrParentTaskCannotBeTheSame.
func IsErrParentTaskCannotBeTheSame(err error) bool {
	_, ok := err.(ErrParentTaskCannotBeTheSame)
	return ok
}

func (err ErrParentTaskCannotBeTheSame) Error() string {
	return fmt.Sprintf("Tried to set a parents task as the same [TaskID: %v]", err.TaskID)
}

// ErrCodeParentTaskCannotBeTheSame holds the unique world-error code of this error
const ErrCodeParentTaskCannotBeTheSame = 4006

// HTTPError holds the http error description
func (err ErrParentTaskCannotBeTheSame) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusForbidden,
		Code:     ErrCodeParentTaskCannotBeTheSame,
		Message:  "You cannot set a parent task to the task itself.",
	}
}

// ErrInvalidRelationKind represents an error where the user tries to use an invalid relation kind
type ErrInvalidRelationKind struct {
	Kind RelationKind
}

// IsErrInvalidRelationKind checks if an error is ErrInvalidRelationKind.
func IsErrInvalidRelationKind(err error) bool {
	_, ok := err.(ErrInvalidRelationKind)
	return ok
}

func (err ErrInvalidRelationKind) Error() string {
	return fmt.Sprintf("Invalid task relation kind [Kind: %v]", err.Kind)
}

// ErrCodeInvalidRelationKind holds the unique world-error code of this error
const ErrCodeInvalidRelationKind = 4007

// HTTPError holds the http error description
func (err ErrInvalidRelationKind) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusBadRequest,
		Code:     ErrCodeInvalidRelationKind,
		Message:  "The task relation is invalid.",
	}
}

// ErrRelationAlreadyExists represents an error where the user tries to create an already existing relation
type ErrRelationAlreadyExists struct {
	Kind        RelationKind
	TaskID      int64
	OtherTaskID int64
}

// IsErrRelationAlreadyExists checks if an error is ErrRelationAlreadyExists.
func IsErrRelationAlreadyExists(err error) bool {
	_, ok := err.(ErrRelationAlreadyExists)
	return ok
}

func (err ErrRelationAlreadyExists) Error() string {
	return fmt.Sprintf("Task relation already exists [TaskID: %v, OtherTaskID: %v, Kind: %v]", err.TaskID, err.OtherTaskID, err.Kind)
}

// ErrCodeRelationAlreadyExists holds the unique world-error code of this error
const ErrCodeRelationAlreadyExists = 4008

// HTTPError holds the http error description
func (err ErrRelationAlreadyExists) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusConflict,
		Code:     ErrCodeRelationAlreadyExists,
		Message:  "The task relation already exists.",
	}
}

// ErrRelationDoesNotExist represents an error where a task relation does not exist.
type ErrRelationDoesNotExist struct {
	Kind        RelationKind
	TaskID      int64
	OtherTaskID int64
}

// IsErrRelationDoesNotExist checks if an error is ErrRelationDoesNotExist.
func IsErrRelationDoesNotExist(err error) bool {
	_, ok := err.(ErrRelationDoesNotExist)
	return ok
}

func (err ErrRelationDoesNotExist) Error() string {
	return fmt.Sprintf("Task relation does not exist [TaskID: %v, OtherTaskID: %v, Kind: %v]", err.TaskID, err.OtherTaskID, err.Kind)
}

// ErrCodeRelationDoesNotExist holds the unique world-error code of this error
const ErrCodeRelationDoesNotExist = 4009

// HTTPError holds the http error description
func (err ErrRelationDoesNotExist) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusNotFound,
		Code:     ErrCodeRelationDoesNotExist,
		Message:  "The task relation does not exist.",
	}
}

// ErrRelationTasksCannotBeTheSame represents an error where the user tries to relate a task with itself
type ErrRelationTasksCannotBeTheSame struct {
	TaskID      int64
	OtherTaskID int64
}

// IsErrRelationTasksCannotBeTheSame checks if an error is ErrRelationTasksCannotBeTheSame.
func IsErrRelationTasksCannotBeTheSame(err error) bool {
	_, ok := err.(ErrRelationTasksCannotBeTheSame)
	return ok
}

func (err ErrRelationTasksCannotBeTheSame) Error() string {
	return fmt.Sprintf("Tried to relate a task with itself [TaskID: %v, OtherTaskID: %v]", err.TaskID, err.OtherTaskID)
}

// ErrCodeRelationTasksCannotBeTheSame holds the unique world-error code of this error
const ErrCodeRelationTasksCannotBeTheSame = 4010

// HTTPError holds the http error description
func (err ErrRelationTasksCannotBeTheSame) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusBadRequest,
		Code:     ErrCodeRelationTasksCannotBeTheSame,
		Message:  "You cannot relate a task with itself",
	}
}

// ErrTaskAttachmentDoesNotExist represents an error where the user tries to relate a task with itself
type ErrTaskAttachmentDoesNotExist struct {
	TaskID       int64
	AttachmentID int64
	FileID       int64
}

// IsErrTaskAttachmentDoesNotExist checks if an error is ErrTaskAttachmentDoesNotExist.
func IsErrTaskAttachmentDoesNotExist(err error) bool {
	_, ok := err.(ErrTaskAttachmentDoesNotExist)
	return ok
}

func (err ErrTaskAttachmentDoesNotExist) Error() string {
	return fmt.Sprintf("Task attachment does not exist [TaskID: %d, AttachmentID: %d, FileID: %d]", err.TaskID, err.AttachmentID, err.FileID)
}

// ErrCodeTaskAttachmentDoesNotExist holds the unique world-error code of this error
const ErrCodeTaskAttachmentDoesNotExist = 4011

// HTTPError holds the http error description
func (err ErrTaskAttachmentDoesNotExist) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusNotFound,
		Code:     ErrCodeTaskAttachmentDoesNotExist,
		Message:  "This task attachment does not exist.",
	}
}

// ErrTaskAttachmentIsTooLarge represents an error where the user tries to relate a task with itself
type ErrTaskAttachmentIsTooLarge struct {
	Size uint64
}

// IsErrTaskAttachmentIsTooLarge checks if an error is ErrTaskAttachmentIsTooLarge.
func IsErrTaskAttachmentIsTooLarge(err error) bool {
	_, ok := err.(ErrTaskAttachmentIsTooLarge)
	return ok
}

func (err ErrTaskAttachmentIsTooLarge) Error() string {
	return fmt.Sprintf("Task attachment is too large [Size: %d]", err.Size)
}

// ErrCodeTaskAttachmentIsTooLarge holds the unique world-error code of this error
const ErrCodeTaskAttachmentIsTooLarge = 4012

// HTTPError holds the http error description
func (err ErrTaskAttachmentIsTooLarge) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusBadRequest,
		Code:     ErrCodeTaskAttachmentIsTooLarge,
		Message:  fmt.Sprintf("The task attachment exceeds the configured file size of %d bytes, filesize was %d", config.FilesMaxSize.GetInt64(), err.Size),
	}
}

// ErrInvalidSortParam represents an error where the provided sort param is invalid
type ErrInvalidSortParam struct {
	SortBy string
}

// IsErrInvalidSortParam checks if an error is ErrInvalidSortParam.
func IsErrInvalidSortParam(err error) bool {
	_, ok := err.(ErrInvalidSortParam)
	return ok
}

func (err ErrInvalidSortParam) Error() string {
	return fmt.Sprintf("Sort param is invalid [SortBy: %s]", err.SortBy)
}

// ErrCodeInvalidSortParam holds the unique world-error code of this error
const ErrCodeInvalidSortParam = 4013

// HTTPError holds the http error description
func (err ErrInvalidSortParam) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusBadRequest,
		Code:     ErrCodeInvalidSortParam,
		Message:  fmt.Sprintf("The task sort param '%s' is invalid.", err.SortBy),
	}
}

// ErrInvalidSortOrder represents an error where the provided sort order is invalid
type ErrInvalidSortOrder struct {
	OrderBy sortOrder
}

// IsErrInvalidSortOrder checks if an error is ErrInvalidSortOrder.
func IsErrInvalidSortOrder(err error) bool {
	_, ok := err.(ErrInvalidSortOrder)
	return ok
}

func (err ErrInvalidSortOrder) Error() string {
	return fmt.Sprintf("Sort order is invalid [OrderBy: %s]", err.OrderBy)
}

// ErrCodeInvalidSortOrder holds the unique world-error code of this error
const ErrCodeInvalidSortOrder = 4014

// HTTPError holds the http error description
func (err ErrInvalidSortOrder) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusBadRequest,
		Code:     ErrCodeInvalidSortOrder,
		Message:  fmt.Sprintf("The task sort order '%s' is invalid. Allowed is either asc or desc.", err.OrderBy),
	}
}

// ErrTaskCommentDoesNotExist represents an error where a task comment does not exist
type ErrTaskCommentDoesNotExist struct {
	ID     int64
	TaskID int64
}

// IsErrTaskCommentDoesNotExist checks if an error is ErrTaskCommentDoesNotExist.
func IsErrTaskCommentDoesNotExist(err error) bool {
	_, ok := err.(ErrTaskCommentDoesNotExist)
	return ok
}

func (err ErrTaskCommentDoesNotExist) Error() string {
	return fmt.Sprintf("Task comment does not exist [ID: %d, TaskID: %d]", err.ID, err.TaskID)
}

// ErrCodeTaskCommentDoesNotExist holds the unique world-error code of this error
const ErrCodeTaskCommentDoesNotExist = 4015

// HTTPError holds the http error description
func (err ErrTaskCommentDoesNotExist) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusNotFound,
		Code:     ErrCodeTaskCommentDoesNotExist,
		Message:  "This task comment does not exist",
	}
}

// ErrInvalidTaskField represents an error where the provided task field is invalid
type ErrInvalidTaskField struct {
	TaskField string
}

// IsErrInvalidTaskField checks if an error is ErrInvalidTaskField.
func IsErrInvalidTaskField(err error) bool {
	_, ok := err.(ErrInvalidTaskField)
	return ok
}

func (err ErrInvalidTaskField) Error() string {
	return fmt.Sprintf("Task Field is invalid [TaskField: %s]", err.TaskField)
}

// ErrCodeInvalidTaskField holds the unique world-error code of this error
const ErrCodeInvalidTaskField = 4016

// HTTPError holds the http error description
func (err ErrInvalidTaskField) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusBadRequest,
		Code:     ErrCodeInvalidTaskField,
		Message:  fmt.Sprintf("The task field '%s' is invalid.", err.TaskField),
	}
}

// ErrInvalidTaskFilterComparator represents an error where the provided task field is invalid
type ErrInvalidTaskFilterComparator struct {
	Comparator taskFilterComparator
}

// IsErrInvalidTaskFilterComparator checks if an error is ErrInvalidTaskFilterComparator.
func IsErrInvalidTaskFilterComparator(err error) bool {
	_, ok := err.(ErrInvalidTaskFilterComparator)
	return ok
}

func (err ErrInvalidTaskFilterComparator) Error() string {
	return fmt.Sprintf("Task filter comparator is invalid [Comparator: %s]", err.Comparator)
}

// ErrCodeInvalidTaskFilterComparator holds the unique world-error code of this error
const ErrCodeInvalidTaskFilterComparator = 4017

// HTTPError holds the http error description
func (err ErrInvalidTaskFilterComparator) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusBadRequest,
		Code:     ErrCodeInvalidTaskFilterComparator,
		Message:  fmt.Sprintf("The task filter comparator '%s' is invalid.", err.Comparator),
	}
}

// ErrInvalidTaskFilterConcatinator represents an error where the provided task field is invalid
type ErrInvalidTaskFilterConcatinator struct {
	Concatinator taskFilterConcatinator
}

// IsErrInvalidTaskFilterConcatinator checks if an error is ErrInvalidTaskFilterConcatinator.
func IsErrInvalidTaskFilterConcatinator(err error) bool {
	_, ok := err.(ErrInvalidTaskFilterConcatinator)
	return ok
}

func (err ErrInvalidTaskFilterConcatinator) Error() string {
	return fmt.Sprintf("Task filter concatinator is invalid [Concatinator: %s]", err.Concatinator)
}

// ErrCodeInvalidTaskFilterConcatinator holds the unique world-error code of this error
const ErrCodeInvalidTaskFilterConcatinator = 4018

// HTTPError holds the http error description
func (err ErrInvalidTaskFilterConcatinator) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusBadRequest,
		Code:     ErrCodeInvalidTaskFilterConcatinator,
		Message:  fmt.Sprintf("The task filter concatinator '%s' is invalid.", err.Concatinator),
	}
}

// ErrInvalidTaskFilterValue represents an error where the provided task filter value is invalid
type ErrInvalidTaskFilterValue struct {
	Value interface{}
	Field string
}

// IsErrInvalidTaskFilterValue checks if an error is ErrInvalidTaskFilterValue.
func IsErrInvalidTaskFilterValue(err error) bool {
	_, ok := err.(ErrInvalidTaskFilterValue)
	return ok
}

func (err ErrInvalidTaskFilterValue) Error() string {
	return fmt.Sprintf("Task filter value is invalid [Value: %v, Field: %s]", err.Value, err.Field)
}

// ErrCodeInvalidTaskFilterValue holds the unique world-error code of this error
const ErrCodeInvalidTaskFilterValue = 4019

// HTTPError holds the http error description
func (err ErrInvalidTaskFilterValue) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusBadRequest,
		Code:     ErrCodeInvalidTaskFilterValue,
		Message:  fmt.Sprintf("The task filter value '%s' for field '%s' is invalid.", err.Value, err.Field),
	}
}

// ErrAttachmentDoesNotBelongToTask represents an error where the provided task cover attachment does not belong to the same task
type ErrAttachmentDoesNotBelongToTask struct {
	TaskID       int64
	AttachmentID int64
}

// IsErrAttachmentAndCoverMustBelongToTheSameTask checks if an error is ErrAttachmentDoesNotBelongToTask.
func IsErrAttachmentAndCoverMustBelongToTheSameTask(err error) bool {
	_, ok := err.(ErrAttachmentDoesNotBelongToTask)
	return ok
}

func (err ErrAttachmentDoesNotBelongToTask) Error() string {
	return fmt.Sprintf("Task attachment and cover image do not belong to the same task [TaskID: %d, AttachmentID: %d]", err.TaskID, err.AttachmentID)
}

// ErrCodeAttachmentDoesNotBelongToTask holds the unique world-error code of this error
const ErrCodeAttachmentDoesNotBelongToTask = 4020

// HTTPError holds the http error description
func (err ErrAttachmentDoesNotBelongToTask) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusBadRequest,
		Code:     ErrCodeAttachmentDoesNotBelongToTask,
		Message:  "This attachment does not belong to that task.",
	}
}

// ErrUserAlreadyAssigned represents an error where the user is already assigned to this task
type ErrUserAlreadyAssigned struct {
	TaskID int64
	UserID int64
}

// IsErrUserAlreadyAssigned checks if an error is ErrUserAlreadyAssigned.
func IsErrUserAlreadyAssigned(err error) bool {
	_, ok := err.(ErrUserAlreadyAssigned)
	return ok
}

func (err ErrUserAlreadyAssigned) Error() string {
	return fmt.Sprintf("User is already assigned to task [TaskID: %d, ID: %d]", err.TaskID, err.UserID)
}

// ErrCodeUserAlreadyAssigned holds the unique world-error code of this error
const ErrCodeUserAlreadyAssigned = 4021

// HTTPError holds the http error description
func (err ErrUserAlreadyAssigned) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusBadRequest,
		Code:     ErrCodeUserAlreadyAssigned,
		Message:  "This user is already assigned to that task.",
	}
}

// ErrReminderRelativeToMissing represents an error where a task has a relative reminder without reference date
type ErrReminderRelativeToMissing struct {
	TaskID int64
}

// IsErrReminderRelativeToMissing checks if an error is ErrReminderRelativeToMissing.
func IsErrReminderRelativeToMissing(err error) bool {
	_, ok := err.(ErrReminderRelativeToMissing)
	return ok
}

func (err ErrReminderRelativeToMissing) Error() string {
	return fmt.Sprintf("Task [TaskID: %v] has a relative reminder without relative_to", err.TaskID)
}

// ErrCodeRelationDoesNotExist holds the unique world-error code of this error
const ErrCodeReminderRelativeToMissing = 4022

// HTTPError holds the http error description
func (err ErrReminderRelativeToMissing) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusBadRequest,
		Code:     ErrCodeReminderRelativeToMissing,
		Message:  "Please provide what the reminder date is relative to",
	}
}

// ErrTaskRelationCycle represents an error where the user tries to create an already existing relation
type ErrTaskRelationCycle struct {
	Kind        RelationKind
	TaskID      int64
	OtherTaskID int64
}

// IsErrTaskRelationCycle checks if an error is ErrTaskRelationCycle.
func IsErrTaskRelationCycle(err error) bool {
	_, ok := err.(ErrTaskRelationCycle)
	return ok
}

func (err ErrTaskRelationCycle) Error() string {
	return fmt.Sprintf("Task relation cycle detectetd [TaskID: %v, OtherTaskID: %v, Kind: %v]", err.TaskID, err.OtherTaskID, err.Kind)
}

// ErrCodeTaskRelationCycle holds the unique world-error code of this error
const ErrCodeTaskRelationCycle = 4023

// HTTPError holds the http error description
func (err ErrTaskRelationCycle) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusConflict,
		Code:     ErrCodeTaskRelationCycle,
		Message:  "This task relation would create a cycle.",
	}
}

// ErrInvalidFilterExpression represents an error where the task filter expression was invalid
type ErrInvalidFilterExpression struct {
	Expression      string
	ExpressionError error
}

// IsErrInvalidFilterExpression checks if an error is ErrInvalidFilterExpression.
func IsErrInvalidFilterExpression(err error) bool {
	_, ok := err.(ErrInvalidFilterExpression)
	return ok
}

func (err ErrInvalidFilterExpression) Error() string {
	return fmt.Sprintf("Task filter expression '%s' is invalid [ExpressionError: %v]", err.Expression, err.ExpressionError)
}

// ErrCodeInvalidFilterExpression holds the unique world-error code of this error
const ErrCodeInvalidFilterExpression = 4024

// HTTPError holds the http error description
func (err ErrInvalidFilterExpression) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusBadRequest,
		Code:     ErrCodeInvalidFilterExpression,
		Message:  fmt.Sprintf("The filter expression '%s' is invalid: %v", err.Expression, err.ExpressionError),
	}
}

// ErrInvalidReactionEntityKind represents an error where the reaction kind is invalid
type ErrInvalidReactionEntityKind struct {
	Kind string
}

// IsErrInvalidReactionEntityKind checks if an error is ErrInvalidReactionEntityKind.
func IsErrInvalidReactionEntityKind(err error) bool {
	_, ok := err.(ErrInvalidReactionEntityKind)
	return ok
}

func (err ErrInvalidReactionEntityKind) Error() string {
	return fmt.Sprintf("Reaction kind %s is invalid", err.Kind)
}

// ErrCodeInvalidReactionEntityKind holds the unique world-error code of this error
const ErrCodeInvalidReactionEntityKind = 4025

// HTTPError holds the http error description
func (err ErrInvalidReactionEntityKind) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusBadRequest,
		Code:     ErrCodeInvalidReactionEntityKind,
		Message:  fmt.Sprintf("The reaction kind '%s' is invalid.", err.Kind),
	}
}

// ErrMustHaveProjectViewToSortByPosition represents an error where no project view id was supplied
type ErrMustHaveProjectViewToSortByPosition struct{}

func (err ErrMustHaveProjectViewToSortByPosition) Error() string {
	return "You must provide a project view ID when sorting by position"
}

// ErrCodeMustHaveProjectViewToSortByPosition holds the unique world-error code of this error
const ErrCodeMustHaveProjectViewToSortByPosition = 4026

// HTTPError holds the http error description
func (err ErrMustHaveProjectViewToSortByPosition) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusBadRequest,
		Code:     ErrCodeMustHaveProjectViewToSortByPosition,
		Message:  "You must provide a project view ID when sorting by position",
	}
}

// ============
// Team errors
// ============

// ErrTeamNameCannotBeEmpty represents an error where a team name is empty.
type ErrTeamNameCannotBeEmpty struct {
	TeamID int64
}

// IsErrTeamNameCannotBeEmpty checks if an error is a ErrTeamNameCannotBeEmpty.
func IsErrTeamNameCannotBeEmpty(err error) bool {
	_, ok := err.(ErrTeamNameCannotBeEmpty)
	return ok
}

func (err ErrTeamNameCannotBeEmpty) Error() string {
	return fmt.Sprintf("Team name cannot be empty [Team ID: %d]", err.TeamID)
}

// ErrCodeTeamNameCannotBeEmpty holds the unique world-error code of this error
const ErrCodeTeamNameCannotBeEmpty = 6001

// HTTPError holds the http error description
func (err ErrTeamNameCannotBeEmpty) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: http.StatusBadRequest, Code: ErrCodeTeamNameCannotBeEmpty, Message: "The team name cannot be empty"}
}

type ErrTeamDoesNotExist struct {
	TeamID int64
}

// IsErrTeamDoesNotExist checks if an error is ErrTeamDoesNotExist.
func IsErrTeamDoesNotExist(err error) bool {
	_, ok := err.(ErrTeamDoesNotExist)
	return ok
}

func (err ErrTeamDoesNotExist) Error() string {
	return fmt.Sprintf("Team does not exist [Team ID: %d]", err.TeamID)
}

// ErrCodeTeamDoesNotExist holds the unique world-error code of this error
const ErrCodeTeamDoesNotExist = 6002

// HTTPError holds the http error description
func (err ErrTeamDoesNotExist) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: http.StatusNotFound, Code: ErrCodeTeamDoesNotExist, Message: "This team does not exist."}
}

// ErrTeamAlreadyHasAccess represents an error where a team already has access to a project
type ErrTeamAlreadyHasAccess struct {
	TeamID int64
	ID     int64
}

// IsErrTeamAlreadyHasAccess checks if an error is ErrTeamAlreadyHasAccess.
func IsErrTeamAlreadyHasAccess(err error) bool {
	_, ok := err.(ErrTeamAlreadyHasAccess)
	return ok
}

func (err ErrTeamAlreadyHasAccess) Error() string {
	return fmt.Sprintf("Team already has access. [Team ID: %d, ID: %d]", err.TeamID, err.ID)
}

// ErrCodeTeamAlreadyHasAccess holds the unique world-error code of this error
const ErrCodeTeamAlreadyHasAccess = 6004

// HTTPError holds the http error description
func (err ErrTeamAlreadyHasAccess) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: http.StatusConflict, Code: ErrCodeTeamAlreadyHasAccess, Message: "This team already has access."}
}

// ErrUserIsMemberOfTeam represents an error where a user is already member of a team.
type ErrUserIsMemberOfTeam struct {
	TeamID int64
	UserID int64
}

// IsErrUserIsMemberOfTeam checks if an error is ErrUserIsMemberOfTeam.
func IsErrUserIsMemberOfTeam(err error) bool {
	_, ok := err.(ErrUserIsMemberOfTeam)
	return ok
}

func (err ErrUserIsMemberOfTeam) Error() string {
	return fmt.Sprintf("User is already a member of that team. [Team ID: %d, User ID: %d]", err.TeamID, err.UserID)
}

// ErrCodeUserIsMemberOfTeam holds the unique world-error code of this error
const ErrCodeUserIsMemberOfTeam = 6005

// HTTPError holds the http error description
func (err ErrUserIsMemberOfTeam) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: http.StatusConflict, Code: ErrCodeUserIsMemberOfTeam, Message: "This user is already a member of that team."}
}

// ErrCannotDeleteLastTeamMember represents an error where a user wants to delete the last member of a team (probably himself)
type ErrCannotDeleteLastTeamMember struct {
	TeamID int64
	UserID int64
}

// IsErrCannotDeleteLastTeamMember checks if an error is ErrCannotDeleteLastTeamMember.
func IsErrCannotDeleteLastTeamMember(err error) bool {
	_, ok := err.(ErrCannotDeleteLastTeamMember)
	return ok
}

func (err ErrCannotDeleteLastTeamMember) Error() string {
	return fmt.Sprintf("Cannot delete last team member. [Team ID: %d, User ID: %d]", err.TeamID, err.UserID)
}

// ErrCodeCannotDeleteLastTeamMember holds the unique world-error code of this error
const ErrCodeCannotDeleteLastTeamMember = 6006

// HTTPError holds the http error description
func (err ErrCannotDeleteLastTeamMember) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: http.StatusBadRequest, Code: ErrCodeCannotDeleteLastTeamMember, Message: "You cannot delete the last member of a team."}
}

// ErrTeamDoesNotHaveAccessToProject represents an error, where the Team is not the owner of that Project (used i.e. when deleting a Project)
type ErrTeamDoesNotHaveAccessToProject struct {
	ProjectID int64
	TeamID    int64
}

// IsErrTeamDoesNotHaveAccessToProject checks if an error is a ErrProjectDoesNotExist.
func IsErrTeamDoesNotHaveAccessToProject(err error) bool {
	_, ok := err.(ErrTeamDoesNotHaveAccessToProject)
	return ok
}

func (err ErrTeamDoesNotHaveAccessToProject) Error() string {
	return fmt.Sprintf("Team does not have access to the project [ProjectID: %d, TeamID: %d]", err.ProjectID, err.TeamID)
}

// ErrCodeTeamDoesNotHaveAccessToProject holds the unique world-error code of this error
const ErrCodeTeamDoesNotHaveAccessToProject = 6007

// HTTPError holds the http error description
func (err ErrTeamDoesNotHaveAccessToProject) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: http.StatusForbidden, Code: ErrCodeTeamDoesNotHaveAccessToProject, Message: "This team does not have access to the project."}
}

// ErrExternalTeamDoesNotExist represents an error where a team with specified oidcId property does not exist for a given issuer
type ErrExternalTeamDoesNotExist struct {
	ExternalID string
	Issuer     string
}

// IsErrExternalTeamDoesNotExist checks if an error is ErrExternalTeamDoesNotExist.
func IsErrExternalTeamDoesNotExist(err error) bool {
	_, ok := err.(ErrExternalTeamDoesNotExist)
	return ok
}

// ErrTeamDoesNotExist represents an error where a team does not exist
func (err ErrExternalTeamDoesNotExist) Error() string {
	return fmt.Sprintf("No team could be found for the given oidcId and issuer. [OIDC ID : %v] [Issuer: %v] ", err.ExternalID, err.Issuer)
}

// ErrCodeTeamDoesNotExist holds the unique world-error code of this error
const ErrCodeOIDCTeamDoesNotExist = 6008

// HTTPError holds the http error description
func (err ErrExternalTeamDoesNotExist) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusNotFound,
		Code:     ErrCodeTeamDoesNotExist,
		Message:  "No team could be found for the given OIDC ID and issuer.",
	}
}

// ErrOIDCTeamsDoNotExistForUser represents an error where an oidcTeam does not exist for the user
type ErrOIDCTeamsDoNotExistForUser struct {
	UserID int64
}

// IsErrOIDCTeamsDoNotExistForUser checks if an error is ErrOIDCTeamsDoNotExistForUser.
func IsErrOIDCTeamsDoNotExistForUser(err error) bool {
	_, ok := err.(ErrOIDCTeamsDoNotExistForUser)
	return ok
}

func (err ErrOIDCTeamsDoNotExistForUser) Error() string {
	return fmt.Sprintf("No teams with property oidcId could be found for user [User ID: %d]", err.UserID)
}

// ErrCodeTeamDoesNotExist holds the unique world-error code of this error
const ErrCodeOIDCTeamsDoNotExistForUser = 6009

// HTTPError holds the http error description
func (err ErrOIDCTeamsDoNotExistForUser) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusNotFound,
		Code:     ErrCodeOIDCTeamsDoNotExistForUser,
		Message:  "No Teams with property oidcId could be found for User.",
	}
}

// ErrCannotRemoveUserFromExternalTeam represents an error where an oidcTeam does not exist for the user
type ErrCannotRemoveUserFromExternalTeam struct {
	TeamID int64
}

func (err ErrCannotRemoveUserFromExternalTeam) Error() string {
	return fmt.Sprintf("Users cannot be removed from an external team [Team ID: %d]", err.TeamID)
}

const ErrCodeCannotLeaveExternalTeam = 6010

// HTTPError holds the http error description
func (err ErrCannotRemoveUserFromExternalTeam) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusPreconditionFailed,
		Code:     ErrCodeCannotLeaveExternalTeam,
		Message:  "Users cannot be removed from an external team.",
	}
}

// ====================
// User <-> Project errors
// ====================

// ErrUserAlreadyHasAccess represents an error where a user already has access to a project
type ErrUserAlreadyHasAccess struct {
	UserID    int64
	ProjectID int64
}

// IsErrUserAlreadyHasAccess checks if an error is ErrUserAlreadyHasAccess.
func IsErrUserAlreadyHasAccess(err error) bool {
	_, ok := err.(ErrUserAlreadyHasAccess)
	return ok
}

func (err ErrUserAlreadyHasAccess) Error() string {
	return fmt.Sprintf("User already has access to that project. [User ID: %d, Project ID: %d]", err.UserID, err.ProjectID)
}

// ErrCodeUserAlreadyHasAccess holds the unique world-error code of this error
const ErrCodeUserAlreadyHasAccess = 7002

// HTTPError holds the http error description
func (err ErrUserAlreadyHasAccess) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: http.StatusConflict, Code: ErrCodeUserAlreadyHasAccess, Message: "This user already has access to this project."}
}

// ErrUserDoesNotHaveAccessToProject represents an error, where the user is not the owner of that Project (used i.e. when deleting a Project)
type ErrUserDoesNotHaveAccessToProject struct {
	ProjectID int64
	UserID    int64
}

// IsErrUserDoesNotHaveAccessToProject checks if an error is a ErrProjectDoesNotExist.
func IsErrUserDoesNotHaveAccessToProject(err error) bool {
	_, ok := err.(ErrUserDoesNotHaveAccessToProject)
	return ok
}

func (err ErrUserDoesNotHaveAccessToProject) Error() string {
	return fmt.Sprintf("User does not have access to the project [ProjectID: %d, ID: %d]", err.ProjectID, err.UserID)
}

// ErrCodeUserDoesNotHaveAccessToProject holds the unique world-error code of this error
const ErrCodeUserDoesNotHaveAccessToProject = 7003

// HTTPError holds the http error description
func (err ErrUserDoesNotHaveAccessToProject) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: http.StatusForbidden, Code: ErrCodeUserDoesNotHaveAccessToProject, Message: "This user does not have access to the project."}
}

// =============
// Label errors
// =============

// ErrLabelIsAlreadyOnTask represents an error where a label is already bound to a task
type ErrLabelIsAlreadyOnTask struct {
	LabelID int64
	TaskID  int64
}

// IsErrLabelIsAlreadyOnTask checks if an error is ErrLabelIsAlreadyOnTask.
func IsErrLabelIsAlreadyOnTask(err error) bool {
	_, ok := err.(ErrLabelIsAlreadyOnTask)
	return ok
}

func (err ErrLabelIsAlreadyOnTask) Error() string {
	return fmt.Sprintf("Label already exists on task [TaskID: %v, LabelID: %v]", err.TaskID, err.LabelID)
}

// ErrCodeLabelIsAlreadyOnTask holds the unique world-error code of this error
const ErrCodeLabelIsAlreadyOnTask = 8001

// HTTPError holds the http error description
func (err ErrLabelIsAlreadyOnTask) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusBadRequest,
		Code:     ErrCodeLabelIsAlreadyOnTask,
		Message:  "This label already exists on the task.",
	}
}

// ErrLabelDoesNotExist represents an error where a label does not exist
type ErrLabelDoesNotExist struct {
	LabelID int64
}

// IsErrLabelDoesNotExist checks if an error is ErrLabelDoesNotExist.
func IsErrLabelDoesNotExist(err error) bool {
	_, ok := err.(ErrLabelDoesNotExist)
	return ok
}

func (err ErrLabelDoesNotExist) Error() string {
	return fmt.Sprintf("Label does not exist [LabelID: %v]", err.LabelID)
}

// ErrCodeLabelDoesNotExist holds the unique world-error code of this error
const ErrCodeLabelDoesNotExist = 8002

// HTTPError holds the http error description
func (err ErrLabelDoesNotExist) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusNotFound,
		Code:     ErrCodeLabelDoesNotExist,
		Message:  "This label does not exist.",
	}
}

// ErrUserHasNoAccessToLabel represents an error where a user does not have the permission to see a label
type ErrUserHasNoAccessToLabel struct {
	LabelID int64
	UserID  int64
}

// IsErrUserHasNoAccessToLabel checks if an error is ErrUserHasNoAccessToLabel.
func IsErrUserHasNoAccessToLabel(err error) bool {
	_, ok := err.(ErrUserHasNoAccessToLabel)
	return ok
}

func (err ErrUserHasNoAccessToLabel) Error() string {
	return fmt.Sprintf("The user does not have access to this label [LabelID: %v, ID: %v]", err.LabelID, err.UserID)
}

// ErrCodeUserHasNoAccessToLabel holds the unique world-error code of this error
const ErrCodeUserHasNoAccessToLabel = 8003

// HTTPError holds the http error description
func (err ErrUserHasNoAccessToLabel) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusForbidden,
		Code:     ErrCodeUserHasNoAccessToLabel,
		Message:  "You don't have access to this label.",
	}
}

// ========
// Permissions
// ========

// ErrInvalidPermission represents an error where a permission is invalid
type ErrInvalidPermission struct {
	Permission Permission
}

// IsErrInvalidPermission checks if an error is ErrInvalidPermission.
func IsErrInvalidPermission(err error) bool {
	_, ok := err.(ErrInvalidPermission)
	return ok
}

func (err ErrInvalidPermission) Error() string {
	return fmt.Sprintf("Permission invalid [Permission: %d]", err.Permission)
}

// ErrCodeInvalidRight holds the unique world-error code of this error
const ErrCodeInvalidRight = 9001

// HTTPError holds the http error description
func (err ErrInvalidPermission) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusBadRequest,
		Code:     ErrCodeInvalidRight,
		Message:  "The permission is invalid.",
	}
}

// ========
// Kanban
// ========

// ErrBucketDoesNotExist represents an error where a kanban bucket does not exist
type ErrBucketDoesNotExist struct {
	BucketID int64
}

// IsErrBucketDoesNotExist checks if an error is ErrBucketDoesNotExist.
func IsErrBucketDoesNotExist(err error) bool {
	_, ok := err.(ErrBucketDoesNotExist)
	return ok
}

func (err ErrBucketDoesNotExist) Error() string {
	return fmt.Sprintf("Bucket does not exist [BucketID: %d]", err.BucketID)
}

// ErrCodeBucketDoesNotExist holds the unique world-error code of this error
const ErrCodeBucketDoesNotExist = 10001

// HTTPError holds the http error description
func (err ErrBucketDoesNotExist) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusNotFound,
		Code:     ErrCodeBucketDoesNotExist,
		Message:  "This bucket does not exist.",
	}
}

// ErrBucketDoesNotBelongToProjectView represents an error where a kanban bucket does not belong to a project
type ErrBucketDoesNotBelongToProjectView struct {
	BucketID      int64
	ProjectViewID int64
}

// IsErrBucketDoesNotBelongToProject checks if an error is ErrBucketDoesNotBelongToProjectView.
func IsErrBucketDoesNotBelongToProject(err error) bool {
	_, ok := err.(ErrBucketDoesNotBelongToProjectView)
	return ok
}

func (err ErrBucketDoesNotBelongToProjectView) Error() string {
	return fmt.Sprintf("Bucket does not not belong to project view [BucketID: %d, ProjectViewID: %d]", err.BucketID, err.ProjectViewID)
}

// ErrCodeBucketDoesNotBelongToProject holds the unique world-error code of this error
const ErrCodeBucketDoesNotBelongToProject = 10002

// HTTPError holds the http error description
func (err ErrBucketDoesNotBelongToProjectView) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusBadRequest,
		Code:     ErrCodeBucketDoesNotBelongToProject,
		Message:  "This bucket does not belong to that project.",
	}
}

// ErrCannotRemoveLastBucket represents an error where a kanban bucket is the last on a project and thus cannot be removed.
type ErrCannotRemoveLastBucket struct {
	BucketID      int64
	ProjectViewID int64
}

// IsErrCannotRemoveLastBucket checks if an error is ErrCannotRemoveLastBucket.
func IsErrCannotRemoveLastBucket(err error) bool {
	_, ok := err.(ErrCannotRemoveLastBucket)
	return ok
}

func (err ErrCannotRemoveLastBucket) Error() string {
	return fmt.Sprintf("Cannot remove last bucket of project view [BucketID: %d, ProjectViewID: %d]", err.BucketID, err.ProjectViewID)
}

// ErrCodeCannotRemoveLastBucket holds the unique world-error code of this error
const ErrCodeCannotRemoveLastBucket = 10003

// HTTPError holds the http error description
func (err ErrCannotRemoveLastBucket) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusPreconditionFailed,
		Code:     ErrCodeCannotRemoveLastBucket,
		Message:  "You cannot remove the last bucket on this project view.",
	}
}

// ErrBucketLimitExceeded represents an error where a task is being created or moved to a bucket which has its limit already exceeded.
type ErrBucketLimitExceeded struct {
	BucketID int64
	Limit    int64
	TaskID   int64 // may be 0
}

// IsErrBucketLimitExceeded checks if an error is ErrBucketLimitExceeded.
func IsErrBucketLimitExceeded(err error) bool {
	_, ok := err.(ErrBucketLimitExceeded)
	return ok
}

func (err ErrBucketLimitExceeded) Error() string {
	return fmt.Sprintf("Cannot add a task to this bucket because it would exceed the limit [BucketID: %d, Limit: %d, TaskID: %d]", err.BucketID, err.Limit, err.TaskID)
}

// ErrCodeBucketLimitExceeded holds the unique world-error code of this error
const ErrCodeBucketLimitExceeded = 10004

// HTTPError holds the http error description
func (err ErrBucketLimitExceeded) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusPreconditionFailed,
		Code:     ErrCodeBucketLimitExceeded,
		Message:  "You cannot add the task to this bucket as it already exceeded the limit of tasks it can hold.",
	}
}

// ErrOnlyOneDoneBucketPerProject represents an error where a bucket is set to the done bucket but one already exists for its project.
type ErrOnlyOneDoneBucketPerProject struct {
	BucketID     int64
	ProjectID    int64
	DoneBucketID int64
}

// IsErrOnlyOneDoneBucketPerProject checks if an error is ErrBucketLimitExceeded.
func IsErrOnlyOneDoneBucketPerProject(err error) bool {
	_, ok := err.(*ErrOnlyOneDoneBucketPerProject)
	return ok
}

func (err *ErrOnlyOneDoneBucketPerProject) Error() string {
	return fmt.Sprintf("There can be only one done bucket per project [BucketID: %d, ProjectID: %d]", err.BucketID, err.ProjectID)
}

// ErrCodeOnlyOneDoneBucketPerProject holds the unique world-error code of this error
const ErrCodeOnlyOneDoneBucketPerProject = 10005

// HTTPError holds the http error description
func (err *ErrOnlyOneDoneBucketPerProject) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusPreconditionFailed,
		Code:     ErrCodeOnlyOneDoneBucketPerProject,
		Message:  "There can be only one done bucket per project.",
	}
}

// ErrTaskAlreadyExistsInBucket represents an error where a task already exists in a bucket for a project view
type ErrTaskAlreadyExistsInBucket struct {
	TaskID        int64
	ProjectViewID int64
}

// IsErrTaskAlreadyExistsInBucket checks if an error is ErrTaskAlreadyExistsInBucket.
func IsErrTaskAlreadyExistsInBucket(err error) bool {
	_, ok := err.(ErrTaskAlreadyExistsInBucket)
	return ok
}

func (err ErrTaskAlreadyExistsInBucket) Error() string {
	return fmt.Sprintf("Task already exists in a bucket for this project view [TaskID: %d, ProjectViewID: %d]", err.TaskID, err.ProjectViewID)
}

// ErrCodeTaskAlreadyExistsInBucket holds the unique world-error code of this error
const ErrCodeTaskAlreadyExistsInBucket = 10006

// HTTPError holds the http error description
func (err ErrTaskAlreadyExistsInBucket) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusBadRequest,
		Code:     ErrCodeTaskAlreadyExistsInBucket,
		Message:  "This task already exists in a bucket for this project view.",
	}
}

// =============
// Saved Filters
// =============

// ErrSavedFilterDoesNotExist represents an error where a kanban bucket does not exist
type ErrSavedFilterDoesNotExist struct {
	SavedFilterID int64
}

// IsErrSavedFilterDoesNotExist checks if an error is ErrSavedFilterDoesNotExist.
func IsErrSavedFilterDoesNotExist(err error) bool {
	_, ok := err.(ErrSavedFilterDoesNotExist)
	return ok
}

func (err ErrSavedFilterDoesNotExist) Error() string {
	return fmt.Sprintf("Saved filter does not exist [SavedFilterID: %d]", err.SavedFilterID)
}

// ErrCodeSavedFilterDoesNotExist holds the unique world-error code of this error
const ErrCodeSavedFilterDoesNotExist = 11001

// HTTPError holds the http error description
func (err ErrSavedFilterDoesNotExist) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusNotFound,
		Code:     ErrCodeSavedFilterDoesNotExist,
		Message:  "This saved filter does not exist.",
	}
}

// ErrSavedFilterNotAvailableForLinkShare represents an error where a kanban bucket does not exist
type ErrSavedFilterNotAvailableForLinkShare struct {
	SavedFilterID int64
	LinkShareID   int64
}

// IsErrSavedFilterNotAvailableForLinkShare checks if an error is ErrSavedFilterNotAvailableForLinkShare.
func IsErrSavedFilterNotAvailableForLinkShare(err error) bool {
	_, ok := err.(ErrSavedFilterNotAvailableForLinkShare)
	return ok
}

func (err ErrSavedFilterNotAvailableForLinkShare) Error() string {
	return fmt.Sprintf("Saved filters are not available for link shares [SavedFilterID: %d, LinkShareID: %d]", err.SavedFilterID, err.LinkShareID)
}

// ErrCodeSavedFilterNotAvailableForLinkShare holds the unique world-error code of this error
const ErrCodeSavedFilterNotAvailableForLinkShare = 11002

// HTTPError holds the http error description
func (err ErrSavedFilterNotAvailableForLinkShare) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusPreconditionFailed,
		Code:     ErrCodeSavedFilterNotAvailableForLinkShare,
		Message:  "Saved filters are not available for link shares.",
	}
}

// =============
// Subscriptions
// =============

// ErrUnknownSubscriptionEntityType represents an error where a subscription entity type is unknown
type ErrUnknownSubscriptionEntityType struct {
	EntityType SubscriptionEntityType
}

// IsErrUnknownSubscriptionEntityType checks if an error is ErrUnknownSubscriptionEntityType.
func IsErrUnknownSubscriptionEntityType(err error) bool {
	_, ok := err.(*ErrUnknownSubscriptionEntityType)
	return ok
}

func (err *ErrUnknownSubscriptionEntityType) Error() string {
	return fmt.Sprintf("Subscription entity type is unknown [EntityType: %d]", err.EntityType)
}

// ErrCodeUnknownSubscriptionEntityType holds the unique world-error code of this error
const ErrCodeUnknownSubscriptionEntityType = 12001

// HTTPError holds the http error description
func (err *ErrUnknownSubscriptionEntityType) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusPreconditionFailed,
		Code:     ErrCodeUnknownSubscriptionEntityType,
		Message:  "The subscription entity type is invalid.",
	}
}

// ErrSubscriptionAlreadyExists represents an error where a subscription entity already exists
type ErrSubscriptionAlreadyExists struct {
	EntityID   int64
	EntityType SubscriptionEntityType
	UserID     int64
}

// IsErrSubscriptionAlreadyExists checks if an error is ErrSubscriptionAlreadyExists.
func IsErrSubscriptionAlreadyExists(err error) bool {
	_, ok := err.(*ErrSubscriptionAlreadyExists)
	return ok
}

func (err *ErrSubscriptionAlreadyExists) Error() string {
	return fmt.Sprintf("Subscription for this (entity_id, entity_type, user_id) already exists [EntityType: %d, EntityID: %d, UserID: %d]", err.EntityType, err.EntityID, err.UserID)
}

// ErrCodeSubscriptionAlreadyExists holds the unique world-error code of this error
const ErrCodeSubscriptionAlreadyExists = 12002

// HTTPError holds the http error description
func (err *ErrSubscriptionAlreadyExists) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusPreconditionFailed,
		Code:     ErrCodeSubscriptionAlreadyExists,
		Message:  "You're already subscribed.",
	}
}

// ErrMustProvideUser represents an error where you need to provide a user to fetch subscriptions
type ErrMustProvideUser struct {
}

// IsErrMustProvideUser checks if an error is ErrMustProvideUser.
func IsErrMustProvideUser(err error) bool {
	_, ok := err.(*ErrMustProvideUser)
	return ok
}

func (err *ErrMustProvideUser) Error() string {
	return "no user provided while fetching subscriptions"
}

// ErrCodeMustProvideUser holds the unique world-error code of this error
const ErrCodeMustProvideUser = 12003

// HTTPError holds the http error description
func (err *ErrMustProvideUser) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusPreconditionFailed,
		Code:     ErrCodeMustProvideUser,
		Message:  "You must provide a user to fetch subscriptions",
	}
}

// =================
// Link Share errors
// =================

// ErrLinkSharePasswordRequired represents an error where a link share authentication requires a password and none was provided
type ErrLinkSharePasswordRequired struct {
	ShareID int64
}

// IsErrLinkSharePasswordRequired checks if an error is ErrLinkSharePasswordRequired.
func IsErrLinkSharePasswordRequired(err error) bool {
	_, ok := err.(*ErrLinkSharePasswordRequired)
	return ok
}

func (err *ErrLinkSharePasswordRequired) Error() string {
	return fmt.Sprintf("Link Share requires a password for authentication [ShareID: %d]", err.ShareID)
}

// ErrCodeLinkSharePasswordRequired holds the unique world-error code of this error
const ErrCodeLinkSharePasswordRequired = 13001

// HTTPError holds the http error description
func (err *ErrLinkSharePasswordRequired) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusPreconditionFailed,
		Code:     ErrCodeLinkSharePasswordRequired,
		Message:  "This link share requires a password for authentication, but none was provided.",
	}
}

// ErrLinkSharePasswordInvalid represents an error where a subscription entity type is unknown
type ErrLinkSharePasswordInvalid struct {
	ShareID int64
}

// IsErrLinkSharePasswordInvalid checks if an error is ErrLinkSharePasswordInvalid.
func IsErrLinkSharePasswordInvalid(err error) bool {
	_, ok := err.(*ErrLinkSharePasswordInvalid)
	return ok
}

func (err *ErrLinkSharePasswordInvalid) Error() string {
	return fmt.Sprintf("Provided Link Share password did not match the saved one [ShareID: %d]", err.ShareID)
}

// ErrCodeLinkSharePasswordInvalid holds the unique world-error code of this error
const ErrCodeLinkSharePasswordInvalid = 13002

// HTTPError holds the http error description
func (err *ErrLinkSharePasswordInvalid) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusForbidden,
		Code:     ErrCodeLinkSharePasswordInvalid,
		Message:  "The provided link share password is invalid.",
	}
}

// ErrLinkShareTokenInvalid represents an error where a link share token is invalid
type ErrLinkShareTokenInvalid struct {
}

// IsErrLinkShareTokenInvalid checks if an error is ErrLinkShareTokenInvalid.
func IsErrLinkShareTokenInvalid(err error) bool {
	_, ok := err.(*ErrLinkShareTokenInvalid)
	return ok
}

func (err *ErrLinkShareTokenInvalid) Error() string {
	return "Provided Link Share Token is invalid"
}

// ErrCodeLinkShareTokenInvalid holds the unique world-error code of this error
const ErrCodeLinkShareTokenInvalid = 13003

// HTTPError holds the http error description
func (err *ErrLinkShareTokenInvalid) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusBadRequest,
		Code:     ErrCodeLinkShareTokenInvalid,
		Message:  "The provided link share token is invalid.",
	}
}

// ================
// API Token Errors
// ================

// ErrAPITokenInvalid represents an error where an api token is invalid
type ErrAPITokenInvalid struct {
}

// IsErrAPITokenInvalid checks if an error is ErrAPITokenInvalid.
func IsErrAPITokenInvalid(err error) bool {
	_, ok := err.(*ErrAPITokenInvalid)
	return ok
}

func (err *ErrAPITokenInvalid) Error() string {
	return "Provided API token is invalid"
}

// ErrCodeAPITokenInvalid holds the unique world-error code of this error
const ErrCodeAPITokenInvalid = 14001

// HTTPError holds the http error description
func (err *ErrAPITokenInvalid) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusBadRequest,
		Code:     ErrCodeAPITokenInvalid,
		Message:  "The provided api token is invalid.",
	}
}

// ErrInvalidAPITokenPermission represents an error where an api token is invalid
type ErrInvalidAPITokenPermission struct {
	Group      string
	Permission string
}

// IsErrInvalidAPITokenPermission checks if an error is ErrInvalidAPITokenPermission.
func IsErrInvalidAPITokenPermission(err error) bool {
	_, ok := err.(*ErrInvalidAPITokenPermission)
	return ok
}

func (err *ErrInvalidAPITokenPermission) Error() string {
	return fmt.Sprintf("API token permission %s of group %s is invalid", err.Permission, err.Group)
}

// ErrCodeInvalidAPITokenPermission holds the unique world-error code of this error
const ErrCodeInvalidAPITokenPermission = 14002

// HTTPError holds the http error description
func (err *ErrInvalidAPITokenPermission) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusBadRequest,
		Code:     ErrCodeInvalidAPITokenPermission,
		Message:  fmt.Sprintf("The permission %s of group %s is invalid.", err.Permission, err.Group),
	}
}

// OIDC errors
const ErrCodeOpenIDError = 15001

type ErrOpenIDBadRequest struct {
	Message string
}

func (err *ErrOpenIDBadRequest) Error() string {
	return err.Message
}

func (err *ErrOpenIDBadRequest) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusBadRequest,
		Code:     ErrCodeOpenIDError,
		Message:  err.Message,
	}
}

type ErrOpenIDBadRequestWithDetails struct {
	Message string
	Details interface{}
}

func (err *ErrOpenIDBadRequestWithDetails) Error() string {
	return err.Message
}
