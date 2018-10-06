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
	return fmt.Sprintf("a user with this username does already exist [user id: %d, username: %s]", err.UserID, err.Username)
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
	return fmt.Sprintf("a user with this email does already exist [user id: %d, email: %s]", err.UserID, err.Email)
}

// ErrorCodeUserEmailExists holds the unique world-error code of this error
const ErrorCodeUserEmailExists = 1002

// HTTPError holds the http error description
func (err ErrUserEmailExists) HTTPError() HTTPError {
	return HTTPError{HTTPCode: http.StatusBadRequest, Code: ErrorCodeUserEmailExists, Message: "A user with this email address already exists."}
}

// ErrNoUsername represents a "UsernameAlreadyExists" kind of error.
type ErrNoUsername struct {
	UserID int64
}

// IsErrNoUsername checks if an error is a ErrUsernameExists.
func IsErrNoUsername(err error) bool {
	_, ok := err.(ErrNoUsername)
	return ok
}

func (err ErrNoUsername) Error() string {
	return fmt.Sprintf("you need to specify a username [user id: %d]", err.UserID)
}

// ErrCodeNoUsername holds the unique world-error code of this error
const ErrCodeNoUsername = 1003

// HTTPError holds the http error description
func (err ErrNoUsername) HTTPError() HTTPError {
	return HTTPError{HTTPCode: http.StatusBadRequest, Code: ErrCodeNoUsername, Message: "Please specify a username."}
}

// ErrNoUsernamePassword represents a "NoUsernamePassword" kind of error.
type ErrNoUsernamePassword struct{}

// IsErrNoUsernamePassword checks if an error is a ErrNoUsernamePassword.
func IsErrNoUsernamePassword(err error) bool {
	_, ok := err.(ErrNoUsernamePassword)
	return ok
}

func (err ErrNoUsernamePassword) Error() string {
	return fmt.Sprintf("you need to specify a username and a password")
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
	return fmt.Sprintf("this user does not exist [user id: %d]", err.UserID)
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
	return fmt.Sprintf("could not get user ID")
}

// ErrCodeCouldNotGetUserID holds the unique world-error code of this error
const ErrCodeCouldNotGetUserID = 1006

// HTTPError holds the http error description
func (err ErrCouldNotGetUserID) HTTPError() HTTPError {
	return HTTPError{HTTPCode: http.StatusBadRequest, Code: ErrCodeCouldNotGetUserID, Message: "Could not get user id."}
}

// ErrCannotDeleteLastUser represents a "ErrCannotDeleteLastUser" kind of error.
type ErrCannotDeleteLastUser struct{}

// IsErrCannotDeleteLastUser checks if an error is a ErrCannotDeleteLastUser.
func IsErrCannotDeleteLastUser(err error) bool {
	_, ok := err.(ErrCannotDeleteLastUser)
	return ok
}

func (err ErrCannotDeleteLastUser) Error() string {
	return fmt.Sprintf("cannot delete last user")
}

// ErrCodeCannotDeleteLastUser holds the unique world-error code of this error
const ErrCodeCannotDeleteLastUser = 1007

// HTTPError holds the http error description
func (err ErrCannotDeleteLastUser) HTTPError() HTTPError {
	return HTTPError{HTTPCode: http.StatusConflict, Code: ErrCodeCannotDeleteLastUser, Message: "Cannot delete the last user on the server."}
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
	return fmt.Sprintf("ID cannot be 0")
}

// ErrCodeIDCannotBeZero holds the unique world-error code of this error
const ErrCodeIDCannotBeZero = 2001

// HTTPError holds the http error description
func (err ErrIDCannotBeZero) HTTPError() HTTPError {
	return HTTPError{HTTPCode: http.StatusBadRequest, Code: ErrCodeIDCannotBeZero, Message: "The ID cannot be empty or 0."}
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

// ErrNeedToBeListAdmin represents an error, where the user is not the owner of that list (used i.e. when deleting a list)
type ErrNeedToBeListAdmin struct {
	ListID int64
	UserID int64
}

// IsErrNeedToBeListAdmin checks if an error is a ErrListDoesNotExist.
func IsErrNeedToBeListAdmin(err error) bool {
	_, ok := err.(ErrNeedToBeListAdmin)
	return ok
}

func (err ErrNeedToBeListAdmin) Error() string {
	return fmt.Sprintf("You need to be list owner to do that [ListID: %d, UserID: %d]", err.ListID, err.UserID)
}

// ErrCodeNeedToBeListAdmin holds the unique world-error code of this error
const ErrCodeNeedToBeListAdmin = 3002

// HTTPError holds the http error description
func (err ErrNeedToBeListAdmin) HTTPError() HTTPError {
	return HTTPError{HTTPCode: http.StatusForbidden, Code: ErrCodeNeedToBeListAdmin, Message: "You need to be the list admin of that list."}
}

// ErrNeedToBeListWriter represents an error, where the user is not the owner of that list (used i.e. when deleting a list)
type ErrNeedToBeListWriter struct {
	ListID int64
	UserID int64
}

// IsErrNeedToBeListWriter checks if an error is a ErrListDoesNotExist.
func IsErrNeedToBeListWriter(err error) bool {
	_, ok := err.(ErrNeedToBeListWriter)
	return ok
}

func (err ErrNeedToBeListWriter) Error() string {
	return fmt.Sprintf("You need to have write acces to the list to do that [ListID: %d, UserID: %d]", err.ListID, err.UserID)
}

// ErrCodeNeedToBeListWriter holds the unique world-error code of this error
const ErrCodeNeedToBeListWriter = 3003

// HTTPError holds the http error description
func (err ErrNeedToBeListWriter) HTTPError() HTTPError {
	return HTTPError{HTTPCode: http.StatusForbidden, Code: ErrCodeNeedToBeListWriter, Message: "You need to have write access on that list."}
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
	return fmt.Sprintf("You need to be List owner to do that [ListID: %d, UserID: %d]", err.ListID, err.UserID)
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
	return fmt.Sprintf("List task text cannot be empty.")
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

// ErrNeedToBeTaskOwner represents an error, where the user is not the owner of that task (used i.e. when deleting a list)
type ErrNeedToBeTaskOwner struct {
	TaskID int64
	UserID int64
}

// IsErrNeedToBeTaskOwner checks if an error is a ErrNeedToBeTaskOwner.
func IsErrNeedToBeTaskOwner(err error) bool {
	_, ok := err.(ErrNeedToBeTaskOwner)
	return ok
}

func (err ErrNeedToBeTaskOwner) Error() string {
	return fmt.Sprintf("You need to be task owner to do that [TaskID: %d, UserID: %d]", err.TaskID, err.UserID)
}

// ErrCodeNeedToBeTaskOwner holds the unique world-error code of this error
const ErrCodeNeedToBeTaskOwner = 4003

// HTTPError holds the http error description
func (err ErrNeedToBeTaskOwner) HTTPError() HTTPError {
	return HTTPError{HTTPCode: http.StatusForbidden, Code: ErrCodeNeedToBeTaskOwner, Message: "You need to be task owner to do that."}
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

// ErrNeedToBeNamespaceOwner represents an error, where the user is not the owner of that namespace (used i.e. when deleting a namespace)
type ErrNeedToBeNamespaceOwner struct {
	NamespaceID int64
	UserID      int64
}

// IsErrNeedToBeNamespaceOwner checks if an error is a ErrNamespaceDoesNotExist.
func IsErrNeedToBeNamespaceOwner(err error) bool {
	_, ok := err.(ErrNeedToBeNamespaceOwner)
	return ok
}

func (err ErrNeedToBeNamespaceOwner) Error() string {
	return fmt.Sprintf("You need to be namespace owner to do that [NamespaceID: %d, UserID: %d]", err.NamespaceID, err.UserID)
}

// ErrCodeNeedToBeNamespaceOwner holds the unique world-error code of this error
const ErrCodeNeedToBeNamespaceOwner = 5002

// HTTPError holds the http error description
func (err ErrNeedToBeNamespaceOwner) HTTPError() HTTPError {
	return HTTPError{HTTPCode: http.StatusForbidden, Code: ErrCodeNeedToBeNamespaceOwner, Message: "You need to be namespace owner to do this."}
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
	return fmt.Sprintf("You need to have access to this namespace to do that [NamespaceID: %d, UserID: %d]", err.NamespaceID, err.UserID)
}

// ErrCodeUserDoesNotHaveAccessToNamespace holds the unique world-error code of this error
const ErrCodeUserDoesNotHaveAccessToNamespace = 5003

// HTTPError holds the http error description
func (err ErrUserDoesNotHaveAccessToNamespace) HTTPError() HTTPError {
	return HTTPError{HTTPCode: http.StatusForbidden, Code: ErrCodeUserDoesNotHaveAccessToNamespace, Message: "This user does not have access to the namespace."}
}

// ErrUserNeedsToBeNamespaceAdmin represents an error, where the user is not the owner of that namespace (used i.e. when deleting a namespace)
type ErrUserNeedsToBeNamespaceAdmin struct {
	NamespaceID int64
	UserID      int64
}

// IsErrUserNeedsToBeNamespaceAdmin checks if an error is a ErrNamespaceDoesNotExist.
func IsErrUserNeedsToBeNamespaceAdmin(err error) bool {
	_, ok := err.(ErrUserNeedsToBeNamespaceAdmin)
	return ok
}

func (err ErrUserNeedsToBeNamespaceAdmin) Error() string {
	return fmt.Sprintf("You need to be namespace admin to do that [NamespaceID: %d, UserID: %d]", err.NamespaceID, err.UserID)
}

// ErrCodeUserNeedsToBeNamespaceAdmin holds the unique world-error code of this error
const ErrCodeUserNeedsToBeNamespaceAdmin = 5004

// HTTPError holds the http error description
func (err ErrUserNeedsToBeNamespaceAdmin) HTTPError() HTTPError {
	return HTTPError{HTTPCode: http.StatusForbidden, Code: ErrCodeUserNeedsToBeNamespaceAdmin, Message: "You need to be namespace admin to do this."}
}

// ErrUserDoesNotHaveWriteAccessToNamespace represents an error, where the user is not the owner of that namespace (used i.e. when deleting a namespace)
type ErrUserDoesNotHaveWriteAccessToNamespace struct {
	NamespaceID int64
	UserID      int64
}

// IsErrUserDoesNotHaveWriteAccessToNamespace checks if an error is a ErrNamespaceDoesNotExist.
func IsErrUserDoesNotHaveWriteAccessToNamespace(err error) bool {
	_, ok := err.(ErrUserDoesNotHaveWriteAccessToNamespace)
	return ok
}

func (err ErrUserDoesNotHaveWriteAccessToNamespace) Error() string {
	return fmt.Sprintf("You need to have write access to this namespace to do that [NamespaceID: %d, UserID: %d]", err.NamespaceID, err.UserID)
}

// ErrCodeUserDoesNotHaveWriteAccessToNamespace holds the unique world-error code of this error
const ErrCodeUserDoesNotHaveWriteAccessToNamespace = 5005

// HTTPError holds the http error description
func (err ErrUserDoesNotHaveWriteAccessToNamespace) HTTPError() HTTPError {
	return HTTPError{HTTPCode: http.StatusForbidden, Code: ErrCodeUserDoesNotHaveWriteAccessToNamespace, Message: "You need to have write access to this namespace to do this."}
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

// ErrNamespaceOwnerCannotBeEmpty represents an error, where a namespace owner is empty.
type ErrNamespaceOwnerCannotBeEmpty struct {
	NamespaceID int64
	UserID      int64
}

// IsErrNamespaceOwnerCannotBeEmpty checks if an error is a ErrNamespaceDoesNotExist.
func IsErrNamespaceOwnerCannotBeEmpty(err error) bool {
	_, ok := err.(ErrNamespaceOwnerCannotBeEmpty)
	return ok
}

func (err ErrNamespaceOwnerCannotBeEmpty) Error() string {
	return fmt.Sprintf("Namespace owner cannot be empty [NamespaceID: %d, UserID: %d]", err.NamespaceID, err.UserID)
}

// ErrCodeNamespaceOwnerCannotBeEmpty holds the unique world-error code of this error
const ErrCodeNamespaceOwnerCannotBeEmpty = 5007

// HTTPError holds the http error description
func (err ErrNamespaceOwnerCannotBeEmpty) HTTPError() HTTPError {
	return HTTPError{HTTPCode: http.StatusBadRequest, Code: ErrCodeNamespaceOwnerCannotBeEmpty, Message: "The namespace owner cannot be empty."}
}

// ErrNeedToBeNamespaceAdmin represents an error, where the user is not the admin of that namespace (used i.e. when deleting a namespace)
type ErrNeedToBeNamespaceAdmin struct {
	NamespaceID int64
	UserID      int64
}

// IsErrNeedToBeNamespaceAdmin checks if an error is a ErrNamespaceDoesNotExist.
func IsErrNeedToBeNamespaceAdmin(err error) bool {
	_, ok := err.(ErrNeedToBeNamespaceAdmin)
	return ok
}

func (err ErrNeedToBeNamespaceAdmin) Error() string {
	return fmt.Sprintf("You need to be namespace owner to do that [NamespaceID: %d, UserID: %d]", err.NamespaceID, err.UserID)
}

// ErrCodeNeedToBeNamespaceAdmin holds the unique world-error code of this error
const ErrCodeNeedToBeNamespaceAdmin = 5008

// HTTPError holds the http error description
func (err ErrNeedToBeNamespaceAdmin) HTTPError() HTTPError {
	return HTTPError{HTTPCode: http.StatusForbidden, Code: ErrCodeNeedToBeNamespaceAdmin, Message: "You need to be namespace owner to do this."}
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
	return fmt.Sprintf("You need to have namespace read access to do that [NamespaceID: %d, UserID: %d]", err.NamespaceID, err.UserID)
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
	return fmt.Sprintf("You need to have access to this namespace to do that [NamespaceID: %d, TeamID: %d]", err.NamespaceID, err.TeamID)
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
	return fmt.Sprintf("This user already has access to that namespace. [User ID: %d, Namespace ID: %d]", err.UserID, err.NamespaceID)
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
	return fmt.Sprintf("The right is invalid [Right: %d]", err.Right)
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
	return fmt.Sprintf("This team already has access. [Team ID: %d, ID: %d]", err.TeamID, err.ID)
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
	return fmt.Sprintf("This user is already a member of that team. [Team ID: %d, User ID: %d]", err.TeamID, err.UserID)
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
	return fmt.Sprintf("You need to have access to this List to do that [ListID: %d, TeamID: %d]", err.ListID, err.TeamID)
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
	return fmt.Sprintf("The right is invalid [Right: %d]", err.Right)
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
	return fmt.Sprintf("This user already has access to that list. [User ID: %d, List ID: %d]", err.UserID, err.ListID)
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
	return fmt.Sprintf("You need to have access to this List to do that [ListID: %d, UserID: %d]", err.ListID, err.UserID)
}

// ErrCodeUserDoesNotHaveAccessToList holds the unique world-error code of this error
const ErrCodeUserDoesNotHaveAccessToList = 7003

// HTTPError holds the http error description
func (err ErrUserDoesNotHaveAccessToList) HTTPError() HTTPError {
	return HTTPError{HTTPCode: http.StatusForbidden, Code: ErrCodeUserDoesNotHaveAccessToList, Message: "This user does not have access to the list."}
}
