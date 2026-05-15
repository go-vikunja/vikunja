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

// Package output defines stable error codes and the JSON envelope
// veans uses for non-zero exits.
package output

import (
	"encoding/json"
	"errors"
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
	var e *Error
	if errors.As(err, &e) {
		return e
	}
	return &Error{Code: CodeUnknown, Message: err.Error(), Cause: err}
}

// EmitError encodes the error as a JSON envelope `{code, error}` to w
// (default stderr). All veans commands share this shape so callers can
// branch on `code` without sniffing the output format.
func EmitError(err error, w io.Writer) {
	if w == nil {
		w = os.Stderr
	}
	e := AsError(err)
	if encErr := json.NewEncoder(w).Encode(e); encErr != nil {
		fmt.Fprintf(os.Stderr, "veans: failed to encode error envelope: %v\n", encErr)
	}
}
