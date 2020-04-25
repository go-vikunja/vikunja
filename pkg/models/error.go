// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2020 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/web"
	"fmt"
	"net/http"
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
	return fmt.Sprintf("Forbidden")
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
	return fmt.Sprintf("ID cannot be empty or 0")
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

// ===========
// List errors
// ===========

// ErrListDoesNotExist represents a "ErrListDoesNotExist" kind of error. Used if the list does not exist.
type ErrListDoesNotExist struct {
	ID int64
}

// IsErrListDoesNotExist checks if an error is a ErrListDoesNotExist.
func IsErrListDoesNotExist(err error) bool {
	_, ok := err.(ErrListDoesNotExist)
	return ok
}

func (err ErrListDoesNotExist) Error() string {
	return fmt.Sprintf("List does not exist [ID: %d]", err.ID)
}

// ErrCodeListDoesNotExist holds the unique world-error code of this error
const ErrCodeListDoesNotExist = 3001

// HTTPError holds the http error description
func (err ErrListDoesNotExist) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: http.StatusNotFound, Code: ErrCodeListDoesNotExist, Message: "This list does not exist."}
}

// ErrNeedToHaveListReadAccess represents an error, where the user dont has read access to that List
type ErrNeedToHaveListReadAccess struct {
	ListID int64
	UserID int64
}

// IsErrNeedToHaveListReadAccess checks if an error is a ErrListDoesNotExist.
func IsErrNeedToHaveListReadAccess(err error) bool {
	_, ok := err.(ErrNeedToHaveListReadAccess)
	return ok
}

func (err ErrNeedToHaveListReadAccess) Error() string {
	return fmt.Sprintf("User needs to have read access to that list [ListID: %d, UserID: %d]", err.ListID, err.UserID)
}

// ErrCodeNeedToHaveListReadAccess holds the unique world-error code of this error
const ErrCodeNeedToHaveListReadAccess = 3004

// HTTPError holds the http error description
func (err ErrNeedToHaveListReadAccess) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: http.StatusForbidden, Code: ErrCodeNeedToHaveListReadAccess, Message: "You need to have read access to this list."}
}

// ErrListTitleCannotBeEmpty represents a "ErrListTitleCannotBeEmpty" kind of error. Used if the list does not exist.
type ErrListTitleCannotBeEmpty struct{}

// IsErrListTitleCannotBeEmpty checks if an error is a ErrListTitleCannotBeEmpty.
func IsErrListTitleCannotBeEmpty(err error) bool {
	_, ok := err.(ErrListTitleCannotBeEmpty)
	return ok
}

func (err ErrListTitleCannotBeEmpty) Error() string {
	return fmt.Sprintf("List title cannot be empty.")
}

// ErrCodeListTitleCannotBeEmpty holds the unique world-error code of this error
const ErrCodeListTitleCannotBeEmpty = 3005

// HTTPError holds the http error description
func (err ErrListTitleCannotBeEmpty) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: http.StatusBadRequest, Code: ErrCodeListTitleCannotBeEmpty, Message: "You must provide at least a list title."}
}

// ErrListShareDoesNotExist represents a "ErrListShareDoesNotExist" kind of error. Used if the list share does not exist.
type ErrListShareDoesNotExist struct {
	ID   int64
	Hash string
}

// IsErrListShareDoesNotExist checks if an error is a ErrListShareDoesNotExist.
func IsErrListShareDoesNotExist(err error) bool {
	_, ok := err.(ErrListShareDoesNotExist)
	return ok
}

func (err ErrListShareDoesNotExist) Error() string {
	return fmt.Sprintf("List share does not exist.")
}

// ErrCodeListShareDoesNotExist holds the unique world-error code of this error
const ErrCodeListShareDoesNotExist = 3006

// HTTPError holds the http error description
func (err ErrListShareDoesNotExist) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: http.StatusNotFound, Code: ErrCodeListShareDoesNotExist, Message: "The list share does not exist."}
}

// ErrListIdentifierIsNotUnique represents a "ErrListIdentifierIsNotUnique" kind of error. Used if the provided list identifier is not unique.
type ErrListIdentifierIsNotUnique struct {
	Identifier string
}

// IsErrListIdentifierIsNotUnique checks if an error is a ErrListIdentifierIsNotUnique.
func IsErrListIdentifierIsNotUnique(err error) bool {
	_, ok := err.(ErrListIdentifierIsNotUnique)
	return ok
}

func (err ErrListIdentifierIsNotUnique) Error() string {
	return fmt.Sprintf("List identifier is not unique.")
}

// ErrCodeListIdentifierIsNotUnique holds the unique world-error code of this error
const ErrCodeListIdentifierIsNotUnique = 3007

// HTTPError holds the http error description
func (err ErrListIdentifierIsNotUnique) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusBadRequest,
		Code:     ErrCodeListIdentifierIsNotUnique,
		Message:  "A list with this identifier already exists.",
	}
}

// ErrListIsArchived represents an error, where a list is archived
type ErrListIsArchived struct {
	ListID int64
}

// IsErrListIsArchived checks if an error is a .
func IsErrListIsArchived(err error) bool {
	_, ok := err.(ErrListIsArchived)
	return ok
}

func (err ErrListIsArchived) Error() string {
	return fmt.Sprintf("List is archived [ListID: %d]", err.ListID)
}

// ErrCodeListIsArchived holds the unique world-error code of this error
const ErrCodeListIsArchived = 3008

// HTTPError holds the http error description
func (err ErrListIsArchived) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: http.StatusPreconditionFailed, Code: ErrCodeListIsArchived, Message: "This lists is archived. Editing or creating new tasks is not possible."}
}

// ================
// List task errors
// ================

// ErrTaskCannotBeEmpty represents a "ErrListDoesNotExist" kind of error. Used if the list does not exist.
type ErrTaskCannotBeEmpty struct{}

// IsErrTaskCannotBeEmpty checks if an error is a ErrListDoesNotExist.
func IsErrTaskCannotBeEmpty(err error) bool {
	_, ok := err.(ErrTaskCannotBeEmpty)
	return ok
}

func (err ErrTaskCannotBeEmpty) Error() string {
	return fmt.Sprintf("List task text cannot be empty.")
}

// ErrCodeTaskCannotBeEmpty holds the unique world-error code of this error
const ErrCodeTaskCannotBeEmpty = 4001

// HTTPError holds the http error description
func (err ErrTaskCannotBeEmpty) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: http.StatusBadRequest, Code: ErrCodeTaskCannotBeEmpty, Message: "You must provide at least a list task text."}
}

// ErrTaskDoesNotExist represents a "ErrListDoesNotExist" kind of error. Used if the list does not exist.
type ErrTaskDoesNotExist struct {
	ID int64
}

// IsErrTaskDoesNotExist checks if an error is a ErrListDoesNotExist.
func IsErrTaskDoesNotExist(err error) bool {
	_, ok := err.(ErrTaskDoesNotExist)
	return ok
}

func (err ErrTaskDoesNotExist) Error() string {
	return fmt.Sprintf("List task does not exist. [ID: %d]", err.ID)
}

// ErrCodeTaskDoesNotExist holds the unique world-error code of this error
const ErrCodeTaskDoesNotExist = 4002

// HTTPError holds the http error description
func (err ErrTaskDoesNotExist) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: http.StatusNotFound, Code: ErrCodeTaskDoesNotExist, Message: "This task does not exist"}
}

// ErrBulkTasksMustBeInSameList represents a "ErrBulkTasksMustBeInSameList" kind of error.
type ErrBulkTasksMustBeInSameList struct {
	ShouldBeID int64
	IsID       int64
}

// IsErrBulkTasksMustBeInSameList checks if an error is a ErrBulkTasksMustBeInSameList.
func IsErrBulkTasksMustBeInSameList(err error) bool {
	_, ok := err.(ErrBulkTasksMustBeInSameList)
	return ok
}

func (err ErrBulkTasksMustBeInSameList) Error() string {
	return fmt.Sprintf("All bulk editing tasks must be in the same list. [Should be: %d, is: %d]", err.ShouldBeID, err.IsID)
}

// ErrCodeBulkTasksMustBeInSameList holds the unique world-error code of this error
const ErrCodeBulkTasksMustBeInSameList = 4003

// HTTPError holds the http error description
func (err ErrBulkTasksMustBeInSameList) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: http.StatusBadRequest, Code: ErrCodeBulkTasksMustBeInSameList, Message: "All tasks must be in the same list."}
}

// ErrBulkTasksNeedAtLeastOne represents a "ErrBulkTasksNeedAtLeastOne" kind of error.
type ErrBulkTasksNeedAtLeastOne struct{}

// IsErrBulkTasksNeedAtLeastOne checks if an error is a ErrBulkTasksNeedAtLeastOne.
func IsErrBulkTasksNeedAtLeastOne(err error) bool {
	_, ok := err.(ErrBulkTasksNeedAtLeastOne)
	return ok
}

func (err ErrBulkTasksNeedAtLeastOne) Error() string {
	return fmt.Sprintf("Need at least one task when bulk editing tasks")
}

// ErrCodeBulkTasksNeedAtLeastOne holds the unique world-error code of this error
const ErrCodeBulkTasksNeedAtLeastOne = 4004

// HTTPError holds the http error description
func (err ErrBulkTasksNeedAtLeastOne) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: http.StatusBadRequest, Code: ErrCodeBulkTasksNeedAtLeastOne, Message: "Need at least one tasks to do bulk editing."}
}

// ErrNoRightToSeeTask represents an error where a user does not have the right to see a task
type ErrNoRightToSeeTask struct {
	TaskID int64
	UserID int64
}

// IsErrNoRightToSeeTask checks if an error is ErrNoRightToSeeTask.
func IsErrNoRightToSeeTask(err error) bool {
	_, ok := err.(ErrNoRightToSeeTask)
	return ok
}

func (err ErrNoRightToSeeTask) Error() string {
	return fmt.Sprintf("User does not have the right to see the task [TaskID: %v, UserID: %v]", err.TaskID, err.UserID)
}

// ErrCodeNoRightToSeeTask holds the unique world-error code of this error
const ErrCodeNoRightToSeeTask = 4005

// HTTPError holds the http error description
func (err ErrNoRightToSeeTask) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusForbidden,
		Code:     ErrCodeNoRightToSeeTask,
		Message:  "You don't have the right to see this task.",
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

// =================
// Namespace errors
// =================

// ErrNamespaceDoesNotExist represents a "ErrNamespaceDoesNotExist" kind of error. Used if the namespace does not exist.
type ErrNamespaceDoesNotExist struct {
	ID int64
}

// IsErrNamespaceDoesNotExist checks if an error is a ErrNamespaceDoesNotExist.
func IsErrNamespaceDoesNotExist(err error) bool {
	_, ok := err.(ErrNamespaceDoesNotExist)
	return ok
}

func (err ErrNamespaceDoesNotExist) Error() string {
	return fmt.Sprintf("Namespace does not exist [ID: %d]", err.ID)
}

// ErrCodeNamespaceDoesNotExist holds the unique world-error code of this error
const ErrCodeNamespaceDoesNotExist = 5001

// HTTPError holds the http error description
func (err ErrNamespaceDoesNotExist) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: http.StatusNotFound, Code: ErrCodeNamespaceDoesNotExist, Message: "Namespace not found."}
}

// ErrUserDoesNotHaveAccessToNamespace represents an error, where the user is not the owner of that namespace (used i.e. when deleting a namespace)
type ErrUserDoesNotHaveAccessToNamespace struct {
	NamespaceID int64
	UserID      int64
}

// IsErrUserDoesNotHaveAccessToNamespace checks if an error is a ErrNamespaceDoesNotExist.
func IsErrUserDoesNotHaveAccessToNamespace(err error) bool {
	_, ok := err.(ErrUserDoesNotHaveAccessToNamespace)
	return ok
}

func (err ErrUserDoesNotHaveAccessToNamespace) Error() string {
	return fmt.Sprintf("User does not have access to the namespace [NamespaceID: %d, UserID: %d]", err.NamespaceID, err.UserID)
}

// ErrCodeUserDoesNotHaveAccessToNamespace holds the unique world-error code of this error
const ErrCodeUserDoesNotHaveAccessToNamespace = 5003

// HTTPError holds the http error description
func (err ErrUserDoesNotHaveAccessToNamespace) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: http.StatusForbidden, Code: ErrCodeUserDoesNotHaveAccessToNamespace, Message: "This user does not have access to the namespace."}
}

// ErrNamespaceNameCannotBeEmpty represents an error, where a namespace name is empty.
type ErrNamespaceNameCannotBeEmpty struct {
	NamespaceID int64
	UserID      int64
}

// IsErrNamespaceNameCannotBeEmpty checks if an error is a ErrNamespaceDoesNotExist.
func IsErrNamespaceNameCannotBeEmpty(err error) bool {
	_, ok := err.(ErrNamespaceNameCannotBeEmpty)
	return ok
}

func (err ErrNamespaceNameCannotBeEmpty) Error() string {
	return fmt.Sprintf("Namespace name cannot be empty [NamespaceID: %d, UserID: %d]", err.NamespaceID, err.UserID)
}

// ErrCodeNamespaceNameCannotBeEmpty holds the unique world-error code of this error
const ErrCodeNamespaceNameCannotBeEmpty = 5006

// HTTPError holds the http error description
func (err ErrNamespaceNameCannotBeEmpty) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: http.StatusBadRequest, Code: ErrCodeNamespaceNameCannotBeEmpty, Message: "The namespace name cannot be empty."}
}

// ErrNeedToHaveNamespaceReadAccess represents an error, where the user is not the owner of that namespace (used i.e. when deleting a namespace)
type ErrNeedToHaveNamespaceReadAccess struct {
	NamespaceID int64
	UserID      int64
}

// IsErrNeedToHaveNamespaceReadAccess checks if an error is a ErrNamespaceDoesNotExist.
func IsErrNeedToHaveNamespaceReadAccess(err error) bool {
	_, ok := err.(ErrNeedToHaveNamespaceReadAccess)
	return ok
}

func (err ErrNeedToHaveNamespaceReadAccess) Error() string {
	return fmt.Sprintf("User does not have access to that namespace [NamespaceID: %d, UserID: %d]", err.NamespaceID, err.UserID)
}

// ErrCodeNeedToHaveNamespaceReadAccess holds the unique world-error code of this error
const ErrCodeNeedToHaveNamespaceReadAccess = 5009

// HTTPError holds the http error description
func (err ErrNeedToHaveNamespaceReadAccess) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: http.StatusForbidden, Code: ErrCodeNeedToHaveNamespaceReadAccess, Message: "You need to have namespace read access to do this."}
}

// ErrTeamDoesNotHaveAccessToNamespace represents an error, where the Team is not the owner of that namespace (used i.e. when deleting a namespace)
type ErrTeamDoesNotHaveAccessToNamespace struct {
	NamespaceID int64
	TeamID      int64
}

// IsErrTeamDoesNotHaveAccessToNamespace checks if an error is a ErrNamespaceDoesNotExist.
func IsErrTeamDoesNotHaveAccessToNamespace(err error) bool {
	_, ok := err.(ErrTeamDoesNotHaveAccessToNamespace)
	return ok
}

func (err ErrTeamDoesNotHaveAccessToNamespace) Error() string {
	return fmt.Sprintf("Team does not have access to that namespace [NamespaceID: %d, TeamID: %d]", err.NamespaceID, err.TeamID)
}

// ErrCodeTeamDoesNotHaveAccessToNamespace holds the unique world-error code of this error
const ErrCodeTeamDoesNotHaveAccessToNamespace = 5010

// HTTPError holds the http error description
func (err ErrTeamDoesNotHaveAccessToNamespace) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: http.StatusForbidden, Code: ErrCodeTeamDoesNotHaveAccessToNamespace, Message: "You need to have access to this namespace to do this."}
}

// ErrUserAlreadyHasNamespaceAccess represents an error where a user already has access to a namespace
type ErrUserAlreadyHasNamespaceAccess struct {
	UserID      int64
	NamespaceID int64
}

// IsErrUserAlreadyHasNamespaceAccess checks if an error is ErrUserAlreadyHasNamespaceAccess.
func IsErrUserAlreadyHasNamespaceAccess(err error) bool {
	_, ok := err.(ErrUserAlreadyHasNamespaceAccess)
	return ok
}

func (err ErrUserAlreadyHasNamespaceAccess) Error() string {
	return fmt.Sprintf("User already has access to that namespace. [User ID: %d, Namespace ID: %d]", err.UserID, err.NamespaceID)
}

// ErrCodeUserAlreadyHasNamespaceAccess holds the unique world-error code of this error
const ErrCodeUserAlreadyHasNamespaceAccess = 5011

// HTTPError holds the http error description
func (err ErrUserAlreadyHasNamespaceAccess) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: http.StatusConflict, Code: ErrCodeUserAlreadyHasNamespaceAccess, Message: "This user already has access to this namespace."}
}

// ErrNamespaceIsArchived represents an error where a namespace is archived
type ErrNamespaceIsArchived struct {
	NamespaceID int64
}

// IsErrNamespaceIsArchived checks if an error is a .
func IsErrNamespaceIsArchived(err error) bool {
	_, ok := err.(ErrNamespaceIsArchived)
	return ok
}

func (err ErrNamespaceIsArchived) Error() string {
	return fmt.Sprintf("Namespace is archived [NamespaceID: %d]", err.NamespaceID)
}

// ErrCodeNamespaceIsArchived holds the unique world-error code of this error
const ErrCodeNamespaceIsArchived = 5012

// HTTPError holds the http error description
func (err ErrNamespaceIsArchived) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: http.StatusPreconditionFailed, Code: ErrCodeNamespaceIsArchived, Message: "This namespaces is archived. Editing or creating new lists is not possible."}
}

// ============
// Team errors
// ============

// ErrTeamNameCannotBeEmpty represents an error where a team name is empty.
type ErrTeamNameCannotBeEmpty struct {
	TeamID int64
}

// IsErrTeamNameCannotBeEmpty checks if an error is a ErrNamespaceDoesNotExist.
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

// ErrTeamDoesNotExist represents an error where a team does not exist
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

// ErrTeamAlreadyHasAccess represents an error where a team already has access to a list/namespace
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

// ErrTeamDoesNotHaveAccessToList represents an error, where the Team is not the owner of that List (used i.e. when deleting a List)
type ErrTeamDoesNotHaveAccessToList struct {
	ListID int64
	TeamID int64
}

// IsErrTeamDoesNotHaveAccessToList checks if an error is a ErrListDoesNotExist.
func IsErrTeamDoesNotHaveAccessToList(err error) bool {
	_, ok := err.(ErrTeamDoesNotHaveAccessToList)
	return ok
}

func (err ErrTeamDoesNotHaveAccessToList) Error() string {
	return fmt.Sprintf("Team does not have access to the list [ListID: %d, TeamID: %d]", err.ListID, err.TeamID)
}

// ErrCodeTeamDoesNotHaveAccessToList holds the unique world-error code of this error
const ErrCodeTeamDoesNotHaveAccessToList = 6007

// HTTPError holds the http error description
func (err ErrTeamDoesNotHaveAccessToList) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: http.StatusForbidden, Code: ErrCodeTeamDoesNotHaveAccessToList, Message: "This team does not have access to the list."}
}

// ====================
// User <-> List errors
// ====================

// ErrUserAlreadyHasAccess represents an error where a user already has access to a list/namespace
type ErrUserAlreadyHasAccess struct {
	UserID int64
	ListID int64
}

// IsErrUserAlreadyHasAccess checks if an error is ErrUserAlreadyHasAccess.
func IsErrUserAlreadyHasAccess(err error) bool {
	_, ok := err.(ErrUserAlreadyHasAccess)
	return ok
}

func (err ErrUserAlreadyHasAccess) Error() string {
	return fmt.Sprintf("User already has access to that list. [User ID: %d, List ID: %d]", err.UserID, err.ListID)
}

// ErrCodeUserAlreadyHasAccess holds the unique world-error code of this error
const ErrCodeUserAlreadyHasAccess = 7002

// HTTPError holds the http error description
func (err ErrUserAlreadyHasAccess) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: http.StatusConflict, Code: ErrCodeUserAlreadyHasAccess, Message: "This user already has access to this list."}
}

// ErrUserDoesNotHaveAccessToList represents an error, where the user is not the owner of that List (used i.e. when deleting a List)
type ErrUserDoesNotHaveAccessToList struct {
	ListID int64
	UserID int64
}

// IsErrUserDoesNotHaveAccessToList checks if an error is a ErrListDoesNotExist.
func IsErrUserDoesNotHaveAccessToList(err error) bool {
	_, ok := err.(ErrUserDoesNotHaveAccessToList)
	return ok
}

func (err ErrUserDoesNotHaveAccessToList) Error() string {
	return fmt.Sprintf("User does not have access to the list [ListID: %d, UserID: %d]", err.ListID, err.UserID)
}

// ErrCodeUserDoesNotHaveAccessToList holds the unique world-error code of this error
const ErrCodeUserDoesNotHaveAccessToList = 7003

// HTTPError holds the http error description
func (err ErrUserDoesNotHaveAccessToList) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: http.StatusForbidden, Code: ErrCodeUserDoesNotHaveAccessToList, Message: "This user does not have access to the list."}
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

// ErrUserHasNoAccessToLabel represents an error where a user does not have the right to see a label
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
	return fmt.Sprintf("The user does not have access to this label [LabelID: %v, UserID: %v]", err.LabelID, err.UserID)
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
// Rights
// ========

// ErrInvalidRight represents an error where a right is invalid
type ErrInvalidRight struct {
	Right Right
}

// IsErrInvalidRight checks if an error is ErrInvalidRight.
func IsErrInvalidRight(err error) bool {
	_, ok := err.(ErrInvalidRight)
	return ok
}

func (err ErrInvalidRight) Error() string {
	return fmt.Sprintf("Right invalid [Right: %d]", err.Right)
}

// ErrCodeInvalidRight holds the unique world-error code of this error
const ErrCodeInvalidRight = 9001

// HTTPError holds the http error description
func (err ErrInvalidRight) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusBadRequest,
		Code:     ErrCodeInvalidRight,
		Message:  "The right is invalid.",
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

// ErrBucketDoesNotBelongToList represents an error where a kanban bucket does not belong to a list
type ErrBucketDoesNotBelongToList struct {
	BucketID int64
	ListID   int64
}

// IsErrBucketDoesNotBelongToList checks if an error is ErrBucketDoesNotBelongToList.
func IsErrBucketDoesNotBelongToList(err error) bool {
	_, ok := err.(ErrBucketDoesNotBelongToList)
	return ok
}

func (err ErrBucketDoesNotBelongToList) Error() string {
	return fmt.Sprintf("Bucket does not not belong to list [BucketID: %d, ListID: %d]", err.BucketID, err.ListID)
}

// ErrCodeBucketDoesNotBelongToList holds the unique world-error code of this error
const ErrCodeBucketDoesNotBelongToList = 10002

// HTTPError holds the http error description
func (err ErrBucketDoesNotBelongToList) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusBadRequest,
		Code:     ErrCodeBucketDoesNotBelongToList,
		Message:  "This bucket does not belong to that list.",
	}
}

// ErrCannotRemoveLastBucket represents an error where a kanban bucket is the last on a list and thus cannot be removed.
type ErrCannotRemoveLastBucket struct {
	BucketID int64
	ListID   int64
}

// IsErrCannotRemoveLastBucket checks if an error is ErrCannotRemoveLastBucket.
func IsErrCannotRemoveLastBucket(err error) bool {
	_, ok := err.(ErrCannotRemoveLastBucket)
	return ok
}

func (err ErrCannotRemoveLastBucket) Error() string {
	return fmt.Sprintf("Cannot remove last bucket of list [BucketID: %d, ListID: %d]", err.BucketID, err.ListID)
}

// ErrCodeCannotRemoveLastBucket holds the unique world-error code of this error
const ErrCodeCannotRemoveLastBucket = 10003

// HTTPError holds the http error description
func (err ErrCannotRemoveLastBucket) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusPreconditionFailed,
		Code:     ErrCodeCannotRemoveLastBucket,
		Message:  "You cannot remove the last bucket on this list.",
	}
}
