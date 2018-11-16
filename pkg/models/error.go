package models

import (
	"fmt"
	"net/http"
)

// HTTPErrorProcessor is executed when the defined error is thrown, it will make sure the user sees an appropriate error message and http status code
type HTTPErrorProcessor interface {
	HTTPError() HTTPError
}

// HTTPError holds informations about an http error
type HTTPError struct {
	HTTPCode int    `json:"-"`
	Code     int    `json:"code"`
	Message  string `json:"message"`
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
func (err ErrUsernameExists) HTTPError() HTTPError {
	return HTTPError{HTTPCode: http.StatusBadRequest, Code: ErrorCodeUsernameExists, Message: "A user with this username already exists."}
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
func (err ErrUserEmailExists) HTTPError() HTTPError {
	return HTTPError{HTTPCode: http.StatusBadRequest, Code: ErrorCodeUserEmailExists, Message: "A user with this email address already exists."}
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
func (err ErrNoUsernamePassword) HTTPError() HTTPError {
	return HTTPError{HTTPCode: http.StatusBadRequest, Code: ErrCodeNoUsernamePassword, Message: "Please specify a username and a password."}
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
func (err ErrUserDoesNotExist) HTTPError() HTTPError {
	return HTTPError{HTTPCode: http.StatusNotFound, Code: ErrCodeUserDoesNotExist, Message: "The user does not exist."}
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
func (err ErrCouldNotGetUserID) HTTPError() HTTPError {
	return HTTPError{HTTPCode: http.StatusBadRequest, Code: ErrCodeCouldNotGetUserID, Message: "Could not get user id."}
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
func (err ErrNoPasswordResetToken) HTTPError() HTTPError {
	return HTTPError{HTTPCode: http.StatusPreconditionFailed, Code: ErrCodeNoPasswordResetToken, Message: "No token to reset a user's password provided."}
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
func (err ErrInvalidPasswordResetToken) HTTPError() HTTPError {
	return HTTPError{HTTPCode: http.StatusPreconditionFailed, Code: ErrCodeInvalidPasswordResetToken, Message: "Invalid token to reset a user's password."}
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
func (err ErrInvalidEmailConfirmToken) HTTPError() HTTPError {
	return HTTPError{HTTPCode: http.StatusPreconditionFailed, Code: ErrCodeInvalidEmailConfirmToken, Message: "Invalid email confirm token."}
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
func (err ErrWrongUsernameOrPassword) HTTPError() HTTPError {
	return HTTPError{HTTPCode: http.StatusPreconditionFailed, Code: ErrCodeWrongUsernameOrPassword, Message: "Wrong username or password."}
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
func (err ErrEmailNotConfirmed) HTTPError() HTTPError {
	return HTTPError{HTTPCode: http.StatusPreconditionFailed, Code: ErrCodeEmailNotConfirmed, Message: "Please confirm your email address."}
}

// IsErrEmailNotConfirmed checks if an error is a IsErrEmailNotConfirmed.
func IsErrEmailNotConfirmed(err error) bool {
	_, ok := err.(ErrEmailNotConfirmed)
	return ok
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
func (err ErrIDCannotBeZero) HTTPError() HTTPError {
	return HTTPError{HTTPCode: http.StatusBadRequest, Code: ErrCodeIDCannotBeZero, Message: "The ID cannot be empty or 0."}
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
func (err ErrInvalidData) HTTPError() HTTPError {
	return HTTPError{HTTPCode: http.StatusBadRequest, Code: ErrCodeInvalidData, Message: err.Message}
}

// ValidationHTTPError is the http error when a validation fails
type ValidationHTTPError struct {
	HTTPError
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
func (err ErrListDoesNotExist) HTTPError() HTTPError {
	return HTTPError{HTTPCode: http.StatusNotFound, Code: ErrCodeListDoesNotExist, Message: "This list does not exist."}
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
func (err ErrNeedToHaveListReadAccess) HTTPError() HTTPError {
	return HTTPError{HTTPCode: http.StatusForbidden, Code: ErrCodeNeedToHaveListReadAccess, Message: "You need to have read access to this list."}
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
func (err ErrListTitleCannotBeEmpty) HTTPError() HTTPError {
	return HTTPError{HTTPCode: http.StatusBadRequest, Code: ErrCodeListTitleCannotBeEmpty, Message: "You must provide at least a list title."}
}

// ================
// List task errors
// ================

// ErrListTaskCannotBeEmpty represents a "ErrListDoesNotExist" kind of error. Used if the list does not exist.
type ErrListTaskCannotBeEmpty struct{}

// IsErrListTaskCannotBeEmpty checks if an error is a ErrListDoesNotExist.
func IsErrListTaskCannotBeEmpty(err error) bool {
	_, ok := err.(ErrListTaskCannotBeEmpty)
	return ok
}

func (err ErrListTaskCannotBeEmpty) Error() string {
	return fmt.Sprintf("List task text cannot be empty.")
}

// ErrCodeListTaskCannotBeEmpty holds the unique world-error code of this error
const ErrCodeListTaskCannotBeEmpty = 4001

// HTTPError holds the http error description
func (err ErrListTaskCannotBeEmpty) HTTPError() HTTPError {
	return HTTPError{HTTPCode: http.StatusBadRequest, Code: ErrCodeListTaskCannotBeEmpty, Message: "You must provide at least a list task text."}
}

// ErrListTaskDoesNotExist represents a "ErrListDoesNotExist" kind of error. Used if the list does not exist.
type ErrListTaskDoesNotExist struct {
	ID int64
}

// IsErrListTaskDoesNotExist checks if an error is a ErrListDoesNotExist.
func IsErrListTaskDoesNotExist(err error) bool {
	_, ok := err.(ErrListTaskDoesNotExist)
	return ok
}

func (err ErrListTaskDoesNotExist) Error() string {
	return fmt.Sprintf("List task does not exist. [ID: %d]", err.ID)
}

// ErrCodeListTaskDoesNotExist holds the unique world-error code of this error
const ErrCodeListTaskDoesNotExist = 4002

// HTTPError holds the http error description
func (err ErrListTaskDoesNotExist) HTTPError() HTTPError {
	return HTTPError{HTTPCode: http.StatusNotFound, Code: ErrCodeListTaskDoesNotExist, Message: "This list task does not exist"}
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
func (err ErrNamespaceDoesNotExist) HTTPError() HTTPError {
	return HTTPError{HTTPCode: http.StatusNotFound, Code: ErrCodeNamespaceDoesNotExist, Message: "Namespace not found."}
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
func (err ErrUserDoesNotHaveAccessToNamespace) HTTPError() HTTPError {
	return HTTPError{HTTPCode: http.StatusForbidden, Code: ErrCodeUserDoesNotHaveAccessToNamespace, Message: "This user does not have access to the namespace."}
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
func (err ErrNamespaceNameCannotBeEmpty) HTTPError() HTTPError {
	return HTTPError{HTTPCode: http.StatusBadRequest, Code: ErrCodeNamespaceNameCannotBeEmpty, Message: "The namespace name cannot be empty."}
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
func (err ErrNeedToHaveNamespaceReadAccess) HTTPError() HTTPError {
	return HTTPError{HTTPCode: http.StatusForbidden, Code: ErrCodeNeedToHaveNamespaceReadAccess, Message: "You need to have namespace read access to do this."}
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
func (err ErrTeamDoesNotHaveAccessToNamespace) HTTPError() HTTPError {
	return HTTPError{HTTPCode: http.StatusForbidden, Code: ErrCodeTeamDoesNotHaveAccessToNamespace, Message: "You need to have access to this namespace to do this."}
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
func (err ErrUserAlreadyHasNamespaceAccess) HTTPError() HTTPError {
	return HTTPError{HTTPCode: http.StatusConflict, Code: ErrCodeUserAlreadyHasNamespaceAccess, Message: "This user already has access to this namespace."}
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
func (err ErrTeamNameCannotBeEmpty) HTTPError() HTTPError {
	return HTTPError{HTTPCode: http.StatusBadRequest, Code: ErrCodeTeamNameCannotBeEmpty, Message: "The team name cannot be empty"}
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
func (err ErrTeamDoesNotExist) HTTPError() HTTPError {
	return HTTPError{HTTPCode: http.StatusNotFound, Code: ErrCodeTeamDoesNotExist, Message: "This team does not exist."}
}

// ErrInvalidTeamRight represents an error where a team right is invalid
type ErrInvalidTeamRight struct {
	Right TeamRight
}

// IsErrInvalidTeamRight checks if an error is ErrInvalidTeamRight.
func IsErrInvalidTeamRight(err error) bool {
	_, ok := err.(ErrInvalidTeamRight)
	return ok
}

func (err ErrInvalidTeamRight) Error() string {
	return fmt.Sprintf("Team right invalid [Right: %d]", err.Right)
}

// ErrCodeInvalidTeamRight holds the unique world-error code of this error
const ErrCodeInvalidTeamRight = 6003

// HTTPError holds the http error description
func (err ErrInvalidTeamRight) HTTPError() HTTPError {
	return HTTPError{HTTPCode: http.StatusBadRequest, Code: ErrCodeInvalidTeamRight, Message: "The team right is invalid."}
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
func (err ErrTeamAlreadyHasAccess) HTTPError() HTTPError {
	return HTTPError{HTTPCode: http.StatusConflict, Code: ErrCodeTeamAlreadyHasAccess, Message: "This team already has access."}
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
func (err ErrUserIsMemberOfTeam) HTTPError() HTTPError {
	return HTTPError{HTTPCode: http.StatusConflict, Code: ErrCodeUserIsMemberOfTeam, Message: "This user is already a member of that team."}
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
func (err ErrCannotDeleteLastTeamMember) HTTPError() HTTPError {
	return HTTPError{HTTPCode: http.StatusBadRequest, Code: ErrCodeCannotDeleteLastTeamMember, Message: "You cannot delete the last member of a team."}
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
func (err ErrTeamDoesNotHaveAccessToList) HTTPError() HTTPError {
	return HTTPError{HTTPCode: http.StatusForbidden, Code: ErrCodeTeamDoesNotHaveAccessToList, Message: "This team does not have access to the list."}
}

// ====================
// User <-> List errors
// ====================

// ErrInvalidUserRight represents an error where a user right is invalid
type ErrInvalidUserRight struct {
	Right UserRight
}

// IsErrInvalidUserRight checks if an error is ErrInvalidUserRight.
func IsErrInvalidUserRight(err error) bool {
	_, ok := err.(ErrInvalidUserRight)
	return ok
}

func (err ErrInvalidUserRight) Error() string {
	return fmt.Sprintf("User right is invalid [Right: %d]", err.Right)
}

// ErrCodeInvalidUserRight holds the unique world-error code of this error
const ErrCodeInvalidUserRight = 7001

// HTTPError holds the http error description
func (err ErrInvalidUserRight) HTTPError() HTTPError {
	return HTTPError{HTTPCode: http.StatusBadRequest, Code: ErrCodeInvalidUserRight, Message: "The user right is invalid."}
}

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
func (err ErrUserAlreadyHasAccess) HTTPError() HTTPError {
	return HTTPError{HTTPCode: http.StatusConflict, Code: ErrCodeUserAlreadyHasAccess, Message: "This user already has access to this list."}
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
func (err ErrUserDoesNotHaveAccessToList) HTTPError() HTTPError {
	return HTTPError{HTTPCode: http.StatusForbidden, Code: ErrCodeUserDoesNotHaveAccessToList, Message: "This user does not have access to the list."}
}
