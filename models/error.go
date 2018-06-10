package models

import "fmt"

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

// ErrAuthorCannotBeEmpty represents a "AuthorCannotBeEmpty" kind of error.
type ErrAuthorCannotBeEmpty struct{}

// IsErrAuthorCannotBeEmpty checks if an error is a ErrAuthorCannotBeEmpty.
func IsErrAuthorCannotBeEmpty(err error) bool {
	_, ok := err.(ErrAuthorCannotBeEmpty)
	return ok
}

func (err ErrAuthorCannotBeEmpty) Error() string {
	return fmt.Sprintf("author cannot be empty")
}

// ErrItemTitleCannotBeEmpty represents a "ErrItemTitleCannotBeEmpty" kind of error.
type ErrItemTitleCannotBeEmpty struct{}

// IsErrItemTitleCannotBeEmpty checks if an error is a ErrItemTitleCannotBeEmpty.
func IsErrItemTitleCannotBeEmpty(err error) bool {
	_, ok := err.(ErrItemTitleCannotBeEmpty)
	return ok
}

func (err ErrItemTitleCannotBeEmpty) Error() string {
	return fmt.Sprintf("title cannot be empty")
}

// ErrBookTitleCannotBeEmpty represents a "ErrBookTitleCannotBeEmpty" kind of error.
type ErrBookTitleCannotBeEmpty struct{}

// IsErrBookTitleCannotBeEmpty checks if an error is a ErrBookTitleCannotBeEmpty.
func IsErrBookTitleCannotBeEmpty(err error) bool {
	_, ok := err.(ErrBookTitleCannotBeEmpty)
	return ok
}

func (err ErrBookTitleCannotBeEmpty) Error() string {
	return fmt.Sprintf("the book should at least have a title")
}

// ErrNoPublisherName represents a "ErrNoPublisherName" kind of error.
type ErrNoPublisherName struct{}

// IsErrNoPublisherName checks if an error is a ErrNoPublisherName.
func IsErrNoPublisherName(err error) bool {
	_, ok := err.(ErrNoPublisherName)
	return ok
}

func (err ErrNoPublisherName) Error() string {
	return fmt.Sprintf("you need at least a name to insert a new publisher")
}
