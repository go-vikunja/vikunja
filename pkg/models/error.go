//  Vikunja is a todo-list application to facilitate your life.
//  Copyright 2018 Vikunja and contributors. All web.Rights reserved.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU General Public License as published by
//  the Free Software Foundation, either version 3 of the License, or
//  (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU General Public License for more details.
//
//  You should have received a copy of the GNU General Public License
//  along with this program.  If not, see <https://www.gnu.org/licenses/>.

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

// =====================
// User Operation Errors
// =====================

// ErrUsernameExists represents a "UsernameAlreadyExists" kind of error.
type ErrUsernameExists struct {
	UserID   int64
	Username string
}

// IsErrUsernameExists checks if an error is a ErrUsernameExists.
func IsErrUsernameExists(err error) bool {
	_, ok := err.(ErrUsernameExists)
	return ok
}

func (err ErrUsernameExists) Error() string {
	return fmt.Sprintf("User with that username already exists [user id: %d, username: %s]", err.UserID, err.Username)
}

// ErrorCodeUsernameExists holds the unique world-error code of this error
const ErrorCodeUsernameExists = 1001

// HTTPError holds the http error description
func (err ErrUsernameExists) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: http.StatusBadRequest, Code: ErrorCodeUsernameExists, Message: "A user with this username already exists."}
}

// ErrUserEmailExists represents a "UserEmailExists" kind of error.
type ErrUserEmailExists struct {
	UserID int64
	Email  string
}

// IsErrUserEmailExists checks if an error is a ErrUserEmailExists.
func IsErrUserEmailExists(err error) bool {
	_, ok := err.(ErrUserEmailExists)
	return ok
}

func (err ErrUserEmailExists) Error() string {
	return fmt.Sprintf("User with that email already exists [user id: %d, email: %s]", err.UserID, err.Email)
}

// ErrorCodeUserEmailExists holds the unique world-error code of this error
const ErrorCodeUserEmailExists = 1002

// HTTPError holds the http error description
func (err ErrUserEmailExists) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: http.StatusBadRequest, Code: ErrorCodeUserEmailExists, Message: "A user with this email address already exists."}
}

// ErrNoUsernamePassword represents a "NoUsernamePassword" kind of error.
type ErrNoUsernamePassword struct{}

// IsErrNoUsernamePassword checks if an error is a ErrNoUsernamePassword.
func IsErrNoUsernamePassword(err error) bool {
	_, ok := err.(ErrNoUsernamePassword)
	return ok
}

func (err ErrNoUsernamePassword) Error() string {
	return fmt.Sprintf("No username and password provided")
}

// ErrCodeNoUsernamePassword holds the unique world-error code of this error
const ErrCodeNoUsernamePassword = 1004

// HTTPError holds the http error description
func (err ErrNoUsernamePassword) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: http.StatusBadRequest, Code: ErrCodeNoUsernamePassword, Message: "Please specify a username and a password."}
}

// ErrUserDoesNotExist represents a "UserDoesNotExist" kind of error.
type ErrUserDoesNotExist struct {
	UserID int64
}

// IsErrUserDoesNotExist checks if an error is a ErrUserDoesNotExist.
func IsErrUserDoesNotExist(err error) bool {
	_, ok := err.(ErrUserDoesNotExist)
	return ok
}

func (err ErrUserDoesNotExist) Error() string {
	return fmt.Sprintf("User does not exist [user id: %d]", err.UserID)
}

// ErrCodeUserDoesNotExist holds the unique world-error code of this error
const ErrCodeUserDoesNotExist = 1005

// HTTPError holds the http error description
func (err ErrUserDoesNotExist) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: http.StatusNotFound, Code: ErrCodeUserDoesNotExist, Message: "The user does not exist."}
}

// ErrCouldNotGetUserID represents a "ErrCouldNotGetUserID" kind of error.
type ErrCouldNotGetUserID struct{}

// IsErrCouldNotGetUserID checks if an error is a ErrCouldNotGetUserID.
func IsErrCouldNotGetUserID(err error) bool {
	_, ok := err.(ErrCouldNotGetUserID)
	return ok
}

func (err ErrCouldNotGetUserID) Error() string {
	return fmt.Sprintf("Could not get user ID")
}

// ErrCodeCouldNotGetUserID holds the unique world-error code of this error
const ErrCodeCouldNotGetUserID = 1006

// HTTPError holds the http error description
func (err ErrCouldNotGetUserID) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: http.StatusBadRequest, Code: ErrCodeCouldNotGetUserID, Message: "Could not get user id."}
}

// ErrNoPasswordResetToken represents an error where no password reset token exists for that user
type ErrNoPasswordResetToken struct {
	UserID int64
}

func (err ErrNoPasswordResetToken) Error() string {
	return fmt.Sprintf("No token to reset a password [UserID: %d]", err.UserID)
}

// ErrCodeNoPasswordResetToken holds the unique world-error code of this error
const ErrCodeNoPasswordResetToken = 1008

// HTTPError holds the http error description
func (err ErrNoPasswordResetToken) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: http.StatusPreconditionFailed, Code: ErrCodeNoPasswordResetToken, Message: "No token to reset a user's password provided."}
}

// ErrInvalidPasswordResetToken is an error where the password reset token is invalid
type ErrInvalidPasswordResetToken struct {
	Token string
}

func (err ErrInvalidPasswordResetToken) Error() string {
	return fmt.Sprintf("Invalid token to reset a password [Token: %s]", err.Token)
}

// ErrCodeInvalidPasswordResetToken holds the unique world-error code of this error
const ErrCodeInvalidPasswordResetToken = 1009

// HTTPError holds the http error description
func (err ErrInvalidPasswordResetToken) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: http.StatusPreconditionFailed, Code: ErrCodeInvalidPasswordResetToken, Message: "Invalid token to reset a user's password."}
}

// IsErrInvalidPasswordResetToken checks if an error is a ErrInvalidPasswordResetToken.
func IsErrInvalidPasswordResetToken(err error) bool {
	_, ok := err.(ErrInvalidPasswordResetToken)
	return ok
}

// ErrInvalidEmailConfirmToken is an error where the email confirm token is invalid
type ErrInvalidEmailConfirmToken struct {
	Token string
}

func (err ErrInvalidEmailConfirmToken) Error() string {
	return fmt.Sprintf("Invalid email confirm token [Token: %s]", err.Token)
}

// ErrCodeInvalidEmailConfirmToken holds the unique world-error code of this error
const ErrCodeInvalidEmailConfirmToken = 1010

// HTTPError holds the http error description
func (err ErrInvalidEmailConfirmToken) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: http.StatusPreconditionFailed, Code: ErrCodeInvalidEmailConfirmToken, Message: "Invalid email confirm token."}
}

// IsErrInvalidEmailConfirmToken checks if an error is a ErrInvalidEmailConfirmToken.
func IsErrInvalidEmailConfirmToken(err error) bool {
	_, ok := err.(ErrInvalidEmailConfirmToken)
	return ok
}

// ErrWrongUsernameOrPassword is an error where the email was not confirmed
type ErrWrongUsernameOrPassword struct {
}

func (err ErrWrongUsernameOrPassword) Error() string {
	return fmt.Sprintf("Wrong username or password")
}

// ErrCodeWrongUsernameOrPassword holds the unique world-error code of this error
const ErrCodeWrongUsernameOrPassword = 1011

// HTTPError holds the http error description
func (err ErrWrongUsernameOrPassword) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: http.StatusPreconditionFailed, Code: ErrCodeWrongUsernameOrPassword, Message: "Wrong username or password."}
}

// IsErrWrongUsernameOrPassword checks if an error is a IsErrEmailNotConfirmed.
func IsErrWrongUsernameOrPassword(err error) bool {
	_, ok := err.(ErrWrongUsernameOrPassword)
	return ok
}

// ErrEmailNotConfirmed is an error where the email was not confirmed
type ErrEmailNotConfirmed struct {
	UserID int64
}

func (err ErrEmailNotConfirmed) Error() string {
	return fmt.Sprintf("Email is not confirmed [UserID: %d]", err.UserID)
}

// ErrCodeEmailNotConfirmed holds the unique world-error code of this error
const ErrCodeEmailNotConfirmed = 1012

// HTTPError holds the http error description
func (err ErrEmailNotConfirmed) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: http.StatusPreconditionFailed, Code: ErrCodeEmailNotConfirmed, Message: "Please confirm your email address."}
}

// IsErrEmailNotConfirmed checks if an error is a IsErrEmailNotConfirmed.
func IsErrEmailNotConfirmed(err error) bool {
	_, ok := err.(ErrEmailNotConfirmed)
	return ok
}

// ErrEmptyNewPassword represents a "EmptyNewPassword" kind of error.
type ErrEmptyNewPassword struct{}

// IsErrEmptyNewPassword checks if an error is a ErrEmptyNewPassword.
func IsErrEmptyNewPassword(err error) bool {
	_, ok := err.(ErrEmptyNewPassword)
	return ok
}

func (err ErrEmptyNewPassword) Error() string {
	return fmt.Sprintf("New password is empty")
}

// ErrCodeEmptyNewPassword holds the unique world-error code of this error
const ErrCodeEmptyNewPassword = 1013

// HTTPError holds the http error description
func (err ErrEmptyNewPassword) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: http.StatusPreconditionFailed, Code: ErrCodeEmptyNewPassword, Message: "Please specify new password."}
}

// ErrEmptyOldPassword represents a "EmptyOldPassword" kind of error.
type ErrEmptyOldPassword struct{}

// IsErrEmptyOldPassword checks if an error is a ErrEmptyOldPassword.
func IsErrEmptyOldPassword(err error) bool {
	_, ok := err.(ErrEmptyOldPassword)
	return ok
}

func (err ErrEmptyOldPassword) Error() string {
	return fmt.Sprintf("Old password is empty")
}

// ErrCodeEmptyOldPassword holds the unique world-error code of this error
const ErrCodeEmptyOldPassword = 1014

// HTTPError holds the http error description
func (err ErrEmptyOldPassword) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: http.StatusPreconditionFailed, Code: ErrCodeEmptyOldPassword, Message: "Please specify old password."}
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
	return web.HTTPError{HTTPCode: http.StatusNotFound, Code: ErrCodeTaskDoesNotExist, Message: "This list task does not exist"}
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
//  Rights
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
	return fmt.Sprintf(" right invalid [Right: %d]", err.Right)
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
