---
date: "2019-02-12:00:00+02:00"
title: "Custom Errors"
draft: false
type: "doc"
menu:
  sidebar:
    parent: "development"
---

# Custom Errors

All custom errors are defined in `pkg/models/errors.go`.
You should add new ones in this file.

Custom errors usually have fields for the http return code, a [vikunja-specific error code]({{< ref "../usage/errors.md">}}) 
and a human-readable error message about what went wrong.

An error consists of multiple functions and definitions:

{{< highlight golang >}}
// This struct holds any information about this specific error.
// In this case, it contains the user ID of a nonexistand user.
// This type should always be a struct, even if it has no values in it.

// ErrUserDoesNotExist represents a "UserDoesNotExist" kind of error.
type ErrUserDoesNotExist struct {
	UserID int64
}

// This function is mostly used in unit tests to check if a returned error is of that type.
// Every error type should have one of these.
// The name should always start with IsErr... followed by the name of the error.

// IsErrUserDoesNotExist checks if an error is a ErrUserDoesNotExist.
func IsErrUserDoesNotExist(err error) bool {
	_, ok := err.(ErrUserDoesNotExist)
	return ok
}

// This is the definition of the actual error type.
// Your error type is _required_ to implement this in order to be able to be returned as an "error" from functions.
func (err ErrUserDoesNotExist) Error() string {
	return fmt.Sprintf("User does not exist [user id: %d]", err.UserID)
}

// This const holds the vikunja error code used to be able to identify this error without having to 
// rely on an error string.
// This needs to be unique, so you should check whether the error code exists or not.
// The general convention for error codes is as follows:
// * Every "group" errors lives in a thousend something. For example all user issues are 1000-something, all 
//   list errors are 3000-something and so on.
// * New error codes should be the current max error code + 1. Don't take free numbers to prevent old errors
//   which are depricated and removed from being "new ones". For example, if there are error codes 1001, 1002, 1004,
//   a new error should be 1005 and not 1003.

// ErrCodeUserDoesNotExist holds the unique world-error code of this error
const ErrCodeUserDoesNotExist = 1005

// This is the implementation which returns an http error which is then passed to the client.
// Here you define the http status code with which one the error will be returned, the vikunja error code and 
// a human-readable error message.

// HTTPError holds the http error description
func (err ErrUserDoesNotExist) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusNotFound, 
		Code: ErrCodeUserDoesNotExist, 
		Message: "The user does not exist.",
    }
}
{{< /highlight >}}