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

package planka

import (
	"net/http"

	"code.vikunja.io/api/pkg/web"
)

// ErrInvalidCredentials is returned when Planka rejects the token or username/password.
type ErrInvalidCredentials struct{}

func (err *ErrInvalidCredentials) Error() string {
	return "Planka rejected the provided credentials"
}

// ErrCodeInvalidCredentials holds the unique world-error code of this error
const ErrCodeInvalidCredentials = 14201

// HTTPError holds the http error description
func (err *ErrInvalidCredentials) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusBadRequest,
		Code:     ErrCodeInvalidCredentials,
		Message:  "Planka rejected the provided credentials. Check the url and the token or username and password.",
	}
}

// ErrLoginStepRequired is returned when Planka's password login answers with a pending step
// (e.g. two-factor verification or accepting the terms) instead of a session.
type ErrLoginStepRequired struct {
	Step string
}

func (err *ErrLoginStepRequired) Error() string {
	return "The Planka account requires an extra login step: " + err.Step
}

// ErrCodeLoginStepRequired holds the unique world-error code of this error
const ErrCodeLoginStepRequired = 14202

// HTTPError holds the http error description
func (err *ErrLoginStepRequired) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode:   http.StatusBadRequest,
		Code:       ErrCodeLoginStepRequired,
		Message:    "The Planka account requires an extra login step (two-factor code or accepting the terms). Log in to Planka in the browser once, or use an API key instead of username and password.",
		I18nParams: map[string]string{"step": err.Step},
	}
}

// ErrUnsupportedVersion is returned when the Planka instance does not look like v2.
type ErrUnsupportedVersion struct{}

func (err *ErrUnsupportedVersion) Error() string {
	return "Only Planka v2 is supported"
}

// ErrInvalidConfig is returned when the migration request misses the url or credentials.
type ErrInvalidConfig struct{}

func (err *ErrInvalidConfig) Error() string {
	return "A Planka url and either a token or username and password are required"
}

// ErrCodeInvalidConfig holds the unique world-error code of this error
const ErrCodeInvalidConfig = 14203

// HTTPError holds the http error description
func (err *ErrInvalidConfig) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusBadRequest,
		Code:     ErrCodeInvalidConfig,
		Message:  "A Planka url and either a token or username and password are required.",
	}
}

// ErrNoPlankaAtURL is returned when the url does not answer like a Planka api (unreachable, 404, ...).
type ErrNoPlankaAtURL struct {
	Reason string
}

func (err *ErrNoPlankaAtURL) Error() string {
	return "No Planka api found at the given url: " + err.Reason
}

// ErrCodeNoPlankaAtURL holds the unique world-error code of this error
const ErrCodeNoPlankaAtURL = 14204

// HTTPError holds the http error description
func (err *ErrNoPlankaAtURL) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusBadRequest,
		Code:     ErrCodeNoPlankaAtURL,
		Message:  "Could not reach a Planka API at the given url. Check the url and that Vikunja can reach the instance.",
	}
}
