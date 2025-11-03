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

package user

import (
	"fmt"
	"net/http"

	"code.vikunja.io/api/pkg/web"
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
	return "No username and password provided"
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
	return "Could not get user ID"
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
	return fmt.Sprintf("No token to reset a password [ID: %d]", err.UserID)
}

// ErrCodeNoPasswordResetToken holds the unique world-error code of this error
const ErrCodeNoPasswordResetToken = 1008

// HTTPError holds the http error description
func (err ErrNoPasswordResetToken) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: http.StatusPreconditionFailed, Code: ErrCodeNoPasswordResetToken, Message: "No token to reset a user's password provided."}
}

// IsErrNoPasswordResetToken checks if an error is ErrNoPasswordResetToken
func IsErrNoPasswordResetToken(err error) bool {
	_, ok := err.(ErrNoPasswordResetToken)
	return ok
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
	return "Wrong username or password"
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
	return fmt.Sprintf("Email is not confirmed [ID: %d]", err.UserID)
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
	return "New password is empty"
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
	return "Old password is empty"
}

// ErrCodeEmptyOldPassword holds the unique world-error code of this error
const ErrCodeEmptyOldPassword = 1014

// HTTPError holds the http error description
func (err ErrEmptyOldPassword) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: http.StatusPreconditionFailed, Code: ErrCodeEmptyOldPassword, Message: "Please specify old password."}
}

// ErrTOTPAlreadyEnabled represents a "TOTPAlreadyEnabled" kind of error.
type ErrTOTPAlreadyEnabled struct{}

// IsErrTOTPAlreadyEnabled checks if an error is a ErrTOTPAlreadyEnabled.
func IsErrTOTPAlreadyEnabled(err error) bool {
	_, ok := err.(ErrTOTPAlreadyEnabled)
	return ok
}

func (err ErrTOTPAlreadyEnabled) Error() string {
	return "Totp is already enabled for this user"
}

// ErrCodeTOTPAlreadyEnabled holds the unique world-error code of this error
const ErrCodeTOTPAlreadyEnabled = 1015

// HTTPError holds the http error description
func (err ErrTOTPAlreadyEnabled) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusPreconditionFailed,
		Code:     ErrCodeTOTPAlreadyEnabled,
		Message:  "Totp is already enabled for this user, but not activated.",
	}
}

// ErrTOTPNotEnabled represents a "TOTPNotEnabled" kind of error.
type ErrTOTPNotEnabled struct{}

// IsErrTOTPNotEnabled checks if an error is a ErrTOTPNotEnabled.
func IsErrTOTPNotEnabled(err error) bool {
	_, ok := err.(ErrTOTPNotEnabled)
	return ok
}

func (err ErrTOTPNotEnabled) Error() string {
	return "Totp is not enabled for this user"
}

// ErrCodeTOTPNotEnabled holds the unique world-error code of this error
const ErrCodeTOTPNotEnabled = 1016

// HTTPError holds the http error description
func (err ErrTOTPNotEnabled) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusPreconditionFailed,
		Code:     ErrCodeTOTPNotEnabled,
		Message:  "Totp is not enabled for this user.",
	}
}

// ErrInvalidTOTPPasscode represents a "InvalidTOTPPasscode" kind of error.
type ErrInvalidTOTPPasscode struct {
	Passcode string
}

// IsErrInvalidTOTPPasscode checks if an error is a ErrInvalidTOTPPasscode.
func IsErrInvalidTOTPPasscode(err error) bool {
	_, ok := err.(ErrInvalidTOTPPasscode)
	return ok
}

func (err ErrInvalidTOTPPasscode) Error() string {
	return "Invalid totp passcode"
}

// ErrCodeInvalidTOTPPasscode holds the unique world-error code of this error
const ErrCodeInvalidTOTPPasscode = 1017

// HTTPError holds the http error description
func (err ErrInvalidTOTPPasscode) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusPreconditionFailed,
		Code:     ErrCodeInvalidTOTPPasscode,
		Message:  "Invalid totp passcode.",
	}
}

// ErrInvalidAvatarProvider represents a "InvalidAvatarProvider" kind of error.
type ErrInvalidAvatarProvider struct {
	AvatarProvider string
}

// IsErrInvalidAvatarProvider checks if an error is a ErrInvalidAvatarProvider.
func IsErrInvalidAvatarProvider(err error) bool {
	_, ok := err.(ErrInvalidAvatarProvider)
	return ok
}

func (err ErrInvalidAvatarProvider) Error() string {
	return "Invalid avatar provider"
}

// ErrCodeInvalidAvatarProvider holds the unique world-error code of this error
const ErrCodeInvalidAvatarProvider = 1018

// HTTPError holds the http error description
func (err ErrInvalidAvatarProvider) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusPreconditionFailed,
		Code:     ErrCodeInvalidAvatarProvider,
		Message:  "Invalid avatar provider setting. See docs for valid types.",
	}
}

// ErrNoOpenIDEmailProvided represents a "NoEmailProvided" kind of error.
type ErrNoOpenIDEmailProvided struct {
}

// IsErrNoEmailProvided checks if an error is a ErrNoOpenIDEmailProvided.
func IsErrNoEmailProvided(err error) bool {
	_, ok := err.(*ErrNoOpenIDEmailProvided)
	return ok
}

func (err *ErrNoOpenIDEmailProvided) Error() string {
	return "No email provided"
}

// ErrCodeNoOpenIDEmailProvided holds the unique world-error code of this error
const ErrCodeNoOpenIDEmailProvided = 1019

// HTTPError holds the http error description
func (err *ErrNoOpenIDEmailProvided) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusPreconditionFailed,
		Code:     ErrCodeNoOpenIDEmailProvided,
		Message:  "No email address available. Please make sure the openid provider publicly provides an email address for your account.",
	}
}

// ErrNoOpenIDEmailProvided represents a "NoEmailProvided" kind of error.
type ErrOpenIDCustomScopeMalformed struct {
}

// IsErrNoEmailProvided checks if an error is a ErrNoOpenIDEmailProvided.
func IsErrOpenIDCustomScopeMalformed(err error) bool {
	_, ok := err.(*ErrOpenIDCustomScopeMalformed)
	return ok
}

func (err *ErrOpenIDCustomScopeMalformed) Error() string {
	return "Custom Scope malformed"
}

// ErrCodeNoOpenIDEmailProvided holds the unique world-error code of this error
const ErrCodeOpenIDCustomScopeMalformed = 1022

// HTTPError holds the http error description
func (err *ErrOpenIDCustomScopeMalformed) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusPreconditionFailed,
		Code:     ErrCodeOpenIDCustomScopeMalformed,
		Message:  "The custom scope set by the OIDC provider is malformed. Please make sure the openid provider sets the data correctly for your scope. Check especially to have set an oidcID",
	}
}

// ErrAccountDisabled represents a "AccountDisabled" kind of error.
type ErrAccountDisabled struct {
	UserID int64
}

// IsErrAccountDisabled checks if an error is a ErrAccountDisabled.
func IsErrAccountDisabled(err error) bool {
	_, ok := err.(*ErrAccountDisabled)
	return ok
}

func (err *ErrAccountDisabled) Error() string {
	return "Account is disabled"
}

// ErrCodeAccountDisabled holds the unique world-error code of this error
const ErrCodeAccountDisabled = 1020

// HTTPError holds the http error description
func (err *ErrAccountDisabled) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusPreconditionFailed,
		Code:     ErrCodeAccountDisabled,
		Message:  "This account is disabled. Check your emails or ask your administrator.",
	}
}

// ErrAccountIsNotLocal represents a "AccountIsNotLocal" kind of error.
type ErrAccountIsNotLocal struct {
	UserID int64
}

// IsErrAccountIsNotLocal checks if an error is a ErrAccountIsNotLocal.
func IsErrAccountIsNotLocal(err error) bool {
	_, ok := err.(*ErrAccountIsNotLocal)
	return ok
}

func (err *ErrAccountIsNotLocal) Error() string {
	return "Account is not local"
}

// ErrCodeAccountIsNotLocal holds the unique world-error code of this error
const ErrCodeAccountIsNotLocal = 1021

// HTTPError holds the http error description
func (err *ErrAccountIsNotLocal) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusPreconditionFailed,
		Code:     ErrCodeAccountIsNotLocal,
		Message:  "This account is managed by a third-party authentication provider.",
	}
}

// ErrUsernameMustNotContainSpaces represents a "UsernameMustNotContainSpaces" kind of error.
type ErrUsernameMustNotContainSpaces struct {
	Username string
}

// IsErrUsernameMustNotContainSpaces checks if an error is a ErrUsernameMustNotContainSpaces.
func IsErrUsernameMustNotContainSpaces(err error) bool {
	_, ok := err.(*ErrUsernameMustNotContainSpaces)
	return ok
}

func (err *ErrUsernameMustNotContainSpaces) Error() string {
	return "username must not contain spaces"
}

// ErrCodeUsernameMustNotContainSpaces holds the unique world-error code of this error
const ErrCodeUsernameMustNotContainSpaces = 1022

// HTTPError holds the http error description
func (err *ErrUsernameMustNotContainSpaces) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusPreconditionFailed,
		Code:     ErrCodeUsernameMustNotContainSpaces,
		Message:  "The username must not contain spaces.",
	}
}

// ErrMustNotBeLinkShare represents a "MustNotBeLinkShare" kind of error.
type ErrMustNotBeLinkShare struct{}

// IsErrMustNotBeLinkShare checks if an error is a ErrMustNotBeLinkShare.
func IsErrMustNotBeLinkShare(err error) bool {
	_, ok := err.(*ErrMustNotBeLinkShare)
	return ok
}

func (err *ErrMustNotBeLinkShare) Error() string {
	return "user must be a *User, not a *models.LinkSharing"
}

// ErrCodeMustNotBeLinkShare holds the unique world-error code of this error
const ErrCodeMustNotBeLinkShare = 1023

// HTTPError holds the http error description
func (err *ErrMustNotBeLinkShare) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusForbidden,
		Code:     ErrCodeMustNotBeLinkShare,
		Message:  "You can't do that as a link share.",
	}
}

// ErrInvalidClaimData represents a "InvalidClaimData" kind of error.
type ErrInvalidClaimData struct {
	Field string
	Type  string
}

// IsErrInvalidClaimData checks if an error is a ErrInvalidClaimData.
func IsErrInvalidClaimData(err error) bool {
	_, ok := err.(*ErrInvalidClaimData)
	return ok
}

func (err *ErrInvalidClaimData) Error() string {
	return fmt.Sprintf("invalid claim data for field %s of type %s", err.Field, err.Type)
}

// ErrCodeInvalidClaimData holds the unique world-error code of this error
const ErrCodeInvalidClaimData = 1024

// HTTPError holds the http error description
func (err *ErrInvalidClaimData) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusBadRequest,
		Code:     ErrCodeInvalidClaimData,
		Message:  fmt.Sprintf("Invalid claim data for field %s of type %s", err.Field, err.Type),
	}
}

// ErrInvalidTimezone represents an error where the provided timezone is invalid
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
	return fmt.Sprintf("Invalid timezone [Name: %s, Error: %s]", err.Name, err.LoadError)
}

// ErrorCodeInvalidTimezone holds the unique world-error code of this error
const ErrorCodeInvalidTimezone = 1025

// HTTPError holds the http error description
func (err ErrInvalidTimezone) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: http.StatusBadRequest, Code: ErrorCodeInvalidTimezone, Message: fmt.Sprintf("The timezone '%s' is invalid. Please select a valid timezone from the list.", err.Name)}
}

// ErrUsernameReserved represents a "UsernameReserved" kind of error.
type ErrUsernameReserved struct {
	Username string
}

// IsErrUsernameReserved checks if an error is a ErrUsernameReserved.
func IsErrUsernameReserved(err error) bool {
	_, ok := err.(ErrUsernameReserved)
	return ok
}

func (err ErrUsernameReserved) Error() string {
	return fmt.Sprintf("Username is reserved [Username: %s]", err.Username)
}

// ErrorCodeUsernameReserved holds the unique world-error code of this error
const ErrorCodeUsernameReserved = 1026

// HTTPError holds the http error description
func (err ErrUsernameReserved) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: http.StatusBadRequest, Code: ErrorCodeUsernameReserved, Message: "This username is reserved and cannot be used."}
}

// ErrInvalidUserContext represents an error where the user context is invalid or missing
type ErrInvalidUserContext struct {
	Reason string
}

// IsErrInvalidUserContext checks if an error is a ErrInvalidUserContext.
func IsErrInvalidUserContext(err error) bool {
	_, ok := err.(ErrInvalidUserContext)
	return ok
}

func (err ErrInvalidUserContext) Error() string {
	return fmt.Sprintf("Invalid user context: %s", err.Reason)
}

// ErrorCodeInvalidUserContext holds the unique world-error code of this error
const ErrorCodeInvalidUserContext = 1027

// HTTPError holds the http error description
func (err ErrInvalidUserContext) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusUnauthorized,
		Code:     ErrorCodeInvalidUserContext,
		Message:  "Invalid user context. Please make sure the passed token is valid and try again.",
	}
}

// ErrInvalidDeletionToken represents an error where the deletion token is invalid
type ErrInvalidDeletionToken struct {
	Token string
}

// IsErrInvalidDeletionToken checks if an error is a ErrInvalidDeletionToken.
func IsErrInvalidDeletionToken(err error) bool {
	_, ok := err.(ErrInvalidDeletionToken)
	return ok
}

func (err ErrInvalidDeletionToken) Error() string {
	return fmt.Sprintf("Invalid deletion token [Token: %s]", err.Token)
}

// ErrorCodeInvalidDeletionToken holds the unique world-error code of this error
const ErrorCodeInvalidDeletionToken = 1028

// HTTPError holds the http error description
func (err ErrInvalidDeletionToken) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusBadRequest,
		Code:     ErrorCodeInvalidDeletionToken,
		Message:  "Invalid deletion token provided.",
	}
}

// ErrTokenUserMismatch represents an error where the token doesn't belong to the user
type ErrTokenUserMismatch struct {
	TokenUserID int64
	UserID      int64
}

// IsErrTokenUserMismatch checks if an error is a ErrTokenUserMismatch.
func IsErrTokenUserMismatch(err error) bool {
	_, ok := err.(ErrTokenUserMismatch)
	return ok
}

func (err ErrTokenUserMismatch) Error() string {
	return fmt.Sprintf("Token user mismatch [Token User ID: %d, User ID: %d]", err.TokenUserID, err.UserID)
}

// ErrorCodeTokenUserMismatch holds the unique world-error code of this error
const ErrorCodeTokenUserMismatch = 1029

// HTTPError holds the http error description
func (err ErrTokenUserMismatch) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusForbidden,
		Code:     ErrorCodeTokenUserMismatch,
		Message:  "This deletion token does not belong to your account.",
	}
}
