// Copyright2018-2020 Vikunja and contriubtors. All rights reserved.
//
// This file is part of Vikunja.
//
// Vikunja is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Vikunja is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Vikunja.  If not, see <https://www.gnu.org/licenses/>.

package user

import (
	"code.vikunja.io/web"
	"fmt"
	"net/http"
)

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

// ErrCouldNotGetUserID represents a "ErrCouldNotGetuser_id" kind of error.
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
