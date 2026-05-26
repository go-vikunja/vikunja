// Package output defines stable error codes and the JSON envelope
// veans uses for non-zero exits.
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type Code string

const (
	CodeNotFound            Code = "NOT_FOUND"
	CodeConflict            Code = "CONFLICT"
	CodeValidation          Code = "VALIDATION_ERROR"
	CodeAuth                Code = "AUTH_ERROR"
	CodeRateLimited         Code = "RATE_LIMITED"
	CodeBotUsersUnavailable Code = "BOT_USERS_UNAVAILABLE"
	CodeNotConfigured       Code = "NOT_CONFIGURED"
	CodeUnknown             Code = "UNKNOWN"
)

// Error is the structured error type used for both internal flow and the
// `--json` mutation envelope. Callers wrap underlying errors with codes via
// Wrap; the cobra runner converts unmapped errors to CodeUnknown.
type Error struct {
	Code    Code   `json:"code"`
	Message string `json:"error"`
	Cause   error  `json:"-"`
}

func (e *Error) Error() string { return e.Message }
func (e *Error) Unwrap() error { return e.Cause }

func New(code Code, format string, args ...any) *Error {
	return &Error{Code: code, Message: fmt.Sprintf(format, args...)}
}

func Wrap(code Code, cause error, format string, args ...any) *Error {
	return &Error{Code: code, Message: fmt.Sprintf(format, args...), Cause: cause}
}

// AsError extracts an *Error from any error chain, returning a CodeUnknown
// wrapper for plain errors.
func AsError(err error) *Error {
	if err == nil {
		return nil
	}
	if e, ok := err.(*Error); ok {
		return e
	}
	return &Error{Code: CodeUnknown, Message: err.Error(), Cause: err}
}

// EmitError writes the JSON envelope when --json is set, or a plain message
// otherwise. Always to stderr so stdout stays parseable.
func EmitError(jsonMode bool, err error, w io.Writer) {
	if w == nil {
		w = os.Stderr
	}
	e := AsError(err)
	if jsonMode {
		_ = json.NewEncoder(w).Encode(e)
		return
	}
	fmt.Fprintf(w, "veans: %s: %s\n", e.Code, e.Message)
}
