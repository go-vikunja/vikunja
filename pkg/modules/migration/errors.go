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

package migration

import (
	"fmt"
	"net/http"
	"time"

	"code.vikunja.io/api/pkg/web"
)

// upstreamErrorBodyLimit caps the response body kept in an ErrUpstreamRequestFailed: it ends up in
// user notifications and logs.
const upstreamErrorBodyLimit = 1024

// ErrUpstreamRequestFailed is returned when the service we migrate from answered with a non-2xx status.
type ErrUpstreamRequestFailed struct {
	Migrator   string
	StatusCode int
	Body       string
}

func NewErrUpstreamRequestFailed(migrator string, statusCode int, body string) *ErrUpstreamRequestFailed {
	if len(body) > upstreamErrorBodyLimit {
		body = body[:upstreamErrorBodyLimit]
	}
	return &ErrUpstreamRequestFailed{
		Migrator:   migrator,
		StatusCode: statusCode,
		Body:       body,
	}
}

func (err *ErrUpstreamRequestFailed) Error() string {
	if err.Body == "" {
		return fmt.Sprintf("%s api error: status code: %d", err.Migrator, err.StatusCode)
	}
	return fmt.Sprintf("%s api error: status code: %d, response was: %s", err.Migrator, err.StatusCode, err.Body)
}

// IsClientError reports whether the upstream blamed the request (expired token, missing item, rate limit)
// rather than itself.
func (err *ErrUpstreamRequestFailed) IsClientError() bool {
	return err.StatusCode >= 400 && err.StatusCode < 500
}

// ErrCodeUpstreamRequestFailed holds the unique world-error code of this error
const ErrCodeUpstreamRequestFailed = 14008

// HTTPError holds the http error description
func (err *ErrUpstreamRequestFailed) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusBadGateway,
		Code:     ErrCodeUpstreamRequestFailed,
		// The upstream response body may contain anything, don't hand it back over the api.
		Message: "The service you are migrating from returned an error.",
	}
}

// ErrMigrationAlreadyRunning includes the migrator holding the account-wide claim.
type ErrMigrationAlreadyRunning struct {
	StartedAt    time.Time
	MigratorName string
}

func (err *ErrMigrationAlreadyRunning) Error() string {
	if err.MigratorName == "" {
		return "Migration already running"
	}
	return "Migration already running: " + err.MigratorName
}

// ErrCodeMigrationAlreadyRunning holds the unique world-error code of this error
const ErrCodeMigrationAlreadyRunning = 14005

// HTTPError holds the http error description
func (err *ErrMigrationAlreadyRunning) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusPreconditionFailed,
		Code:     ErrCodeMigrationAlreadyRunning,
		Message:  err.Error(),
	}
}

// ErrImportRowLimitExceeded reports a CSV exceeding migration.maxcsvrows.
type ErrImportRowLimitExceeded struct {
	MaxRows int64
}

func (err *ErrImportRowLimitExceeded) Error() string {
	return "The import file contains too many rows"
}

// ErrCodeImportRowLimitExceeded holds the unique world-error code of this error
const ErrCodeImportRowLimitExceeded = 14006

// HTTPError holds the http error description
func (err *ErrImportRowLimitExceeded) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusBadRequest,
		Code:     ErrCodeImportRowLimitExceeded,
		Message:  "The import file contains more rows than the configured maximum.",
	}
}

// ErrNotAZipFile represents a "ErrNotAZipFile" kind of error.
type ErrNotAZipFile struct{}

func (err *ErrNotAZipFile) Error() string {
	return "The provided file is not a valid zip file"
}

// ErrCodeNotAZipFile holds the unique world-error code of this error
const ErrCodeNotAZipFile = 14011

// HTTPError holds the http error description
func (err *ErrNotAZipFile) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusBadRequest,
		Code:     ErrCodeNotAZipFile,
		Message:  "The provided file is not a valid zip file.",
	}
}

// ErrFileIsEmpty represents a "ErrFileIsEmpty" kind of error.
type ErrFileIsEmpty struct{}

func (err *ErrFileIsEmpty) Error() string {
	return "The provided file does not contain any data"
}

// ErrCodeFileIsEmpty holds the unique world-error code of this error
const ErrCodeFileIsEmpty = 14012

// HTTPError holds the http error description
func (err *ErrFileIsEmpty) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusBadRequest,
		Code:     ErrCodeFileIsEmpty,
		Message:  "The provided file does not contain any data.",
	}
}

// ErrNoDataFileInZip represents a "ErrNoDataFileInZip" kind of error.
type ErrNoDataFileInZip struct{}

func (err *ErrNoDataFileInZip) Error() string {
	return "The provided zip file does not contain a Vikunja data file"
}

// ErrCodeNoDataFileInZip holds the unique world-error code of this error
const ErrCodeNoDataFileInZip = 14009

// HTTPError holds the http error description
func (err *ErrNoDataFileInZip) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusBadRequest,
		Code:     ErrCodeNoDataFileInZip,
		Message:  "The provided zip file does not contain a Vikunja data file.",
	}
}

// ErrCSVConfigRequired represents an error when the CSV migration endpoint
// is called without the required configuration. The CSV migrator requires
// a mapping configuration and must be used via /migration/csv/migrate with
// a config form field.
type ErrCSVConfigRequired struct{}

func (err *ErrCSVConfigRequired) Error() string {
	return "CSV import requires a configuration with column mappings. Use the /migration/csv/detect endpoint to get suggested mappings, then call /migration/csv/migrate with a config form field."
}

// ErrCodeCSVConfigRequired holds the unique world-error code of this error
const ErrCodeCSVConfigRequired = 14004

// HTTPError holds the http error description
func (err *ErrCSVConfigRequired) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusBadRequest,
		Code:     ErrCodeCSVConfigRequired,
		Message:  "CSV import requires a configuration with column mappings. Use the /migration/csv/detect endpoint to get suggested mappings, then call /migration/csv/migrate with a config form field.",
	}
}

// ErrNotACSVFile represents a "ErrNotACSVFile" kind of error.
type ErrNotACSVFile struct{}

func (err *ErrNotACSVFile) Error() string {
	return "The provided file is not a valid CSV file"
}

// ErrCodeNotACSVFile holds the unique world-error code of this error
const ErrCodeNotACSVFile = 14003

// HTTPError holds the http error description
func (err *ErrNotACSVFile) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusBadRequest,
		Code:     ErrCodeNotACSVFile,
		Message:  "The provided file is not a valid CSV file.",
	}
}

// ErrInvalidImportFile represents an import file which could not be parsed.
type ErrInvalidImportFile struct {
	Err error
}

func (err *ErrInvalidImportFile) Error() string {
	return "The provided file could not be parsed: " + err.Err.Error()
}

func (err *ErrInvalidImportFile) Unwrap() error {
	return err.Err
}

// ErrCodeInvalidImportFile holds the unique world-error code of this error
const ErrCodeInvalidImportFile = 14010

// HTTPError holds the http error description
func (err *ErrInvalidImportFile) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusBadRequest,
		Code:     ErrCodeInvalidImportFile,
		Message:  "The provided file could not be parsed. Please make sure it is a valid export file.",
	}
}

// ErrImportFromUnsupportedVersion represents an export created by a Vikunja version we can no longer read.
type ErrImportFromUnsupportedVersion struct {
	DumpVersion string
	MinVersion  string
}

func (err *ErrImportFromUnsupportedVersion) Error() string {
	return "export was created with an older version " + err.DumpVersion + ", need at least " + err.MinVersion
}

// ErrCodeImportFromUnsupportedVersion holds the unique world-error code of this error
const ErrCodeImportFromUnsupportedVersion = 14013

// HTTPError holds the http error description
func (err *ErrImportFromUnsupportedVersion) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusBadRequest,
		Code:     ErrCodeImportFromUnsupportedVersion,
		Message:  "The export was created with a Vikunja version that is too old to import. Please create a new export with a more recent version.",
	}
}
