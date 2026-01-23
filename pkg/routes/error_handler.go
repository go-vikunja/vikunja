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

package routes

import (
	"encoding/json"
	"errors"
	"net/http"

	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/web"

	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo/v5"
)

// httpCodeGetter is an interface for errors that can provide their HTTP status code.
type httpCodeGetter interface {
	GetHTTPCode() int
}

// errorMessage is used to wrap string error messages in a consistent JSON structure.
type errorMessage struct {
	Message interface{} `json:"message"`
}

// CreateHTTPErrorHandler creates a centralized HTTP error handler that:
// 1. Converts all error types to proper HTTP responses
// 2. Preserves full error details (like ValidationHTTPError.InvalidFields)
// 3. Handles Sentry reporting for 5xx errors
// 4. Logs all errors appropriately
func CreateHTTPErrorHandler(e *echo.Echo, enableSentry bool) echo.HTTPErrorHandler {
	return func(c *echo.Context, err error) {

		var (
			code                = http.StatusInternalServerError
			message interface{} = http.StatusText(http.StatusInternalServerError)
		)

		// Keep track of the original error for logging/sentry
		originalErr := err

		// 1. Check if it implements HTTPStatusCoder (includes echo.ErrForbidden, etc.)
		// In Echo v5, predefined errors like ErrForbidden are *httpError (unexported),
		// not *HTTPError, so we must check the interface instead of the concrete type.
		var sc echo.HTTPStatusCoder
		if errors.As(err, &sc) {
			code = sc.StatusCode()
			// HTTPStatusCoder doesn't have Error(), so we use the status text
			message = http.StatusText(code)
		}

		// 2. If it's specifically an HTTPError, use its message for more details
		var he *echo.HTTPError
		if errors.As(err, &he) {
			code = he.Code
			if he.Message != "" {
				message = he.Message
			}
		}

		// 3. Special case: 413 body limit → convert to ErrFileIsTooLarge
		// Check both the code (if it was an HTTPError) and errors.Is for wrapped errors
		// In Echo v5, body limit errors during multipart parsing may be wrapped
		if code == http.StatusRequestEntityTooLarge || errors.Is(err, echo.ErrStatusRequestEntityTooLarge) {
			fileErr := files.ErrFileIsTooLarge{}
			errDetails := fileErr.HTTPError()
			code = errDetails.HTTPCode
			message = errDetails
		} else if _, isMarshaler := err.(json.Marshaler); isMarshaler {
			// 4. Check for json.Marshaler (preserves full struct like ValidationHTTPError)
			// This allows errors with extra fields (like InvalidFields) to be serialized correctly
			if codeGetter, hasCode := err.(httpCodeGetter); hasCode {
				code = codeGetter.GetHTTPCode()
			}
			message = err // Echo will serialize via MarshalJSON
		} else if hp, ok := err.(web.HTTPErrorProcessor); ok {
			// 5. Standard HTTPErrorProcessor (domain errors like ErrProjectDoesNotExist)
			errDetails := hp.HTTPError()
			code = errDetails.HTTPCode
			message = errDetails
		}
		// 6. For any other error type, we keep the defaults (500 with generic message)
		// or the echo.HTTPStatusCoder/HTTPError values if it was that type

		// Log the error
		log.Error(originalErr.Error())

		// Sentry reporting for 5xx errors
		if enableSentry && code >= 500 {
			reportToSentry(originalErr, c)
		}

		// Send response
		if c.Request().Method == http.MethodHead {
			err = c.NoContent(code)
		} else {
			// Wrap string messages in a struct to ensure consistent JSON format
			// e.g., "Forbidden" becomes {"message": "Forbidden"}
			if _, isString := message.(string); isString {
				message = errorMessage{Message: message}
			}
			err = c.JSON(code, message)
		}
		if err != nil {
			e.Logger.Error(err.Error())
		}
	}
}

// reportToSentry sends an error to Sentry with request context
func reportToSentry(err error, c *echo.Context) {
	hub := GetSentryHubFromContext(c)
	if hub != nil {
		hub.WithScope(func(scope *sentry.Scope) {
			scope.SetExtra("url", c.Request().URL)
			hub.CaptureException(err)
		})
	} else {
		sentry.CaptureException(err)
		log.Debugf("Could not add context for sending error '%s' to sentry", err.Error())
	}
	log.Debugf("Error '%s' sent to sentry", err.Error())
}
