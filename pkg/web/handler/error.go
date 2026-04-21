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

package handler

import (
	"net/http"

	"code.vikunja.io/api/pkg/web"
)

// ErrGenericForbidden indicates the authenticated caller lacks permission
// to perform the requested operation. It is framework-neutral: it implements
// web.HTTPErrorProcessor so any HTTP framework (Echo today, Huma in the
// upcoming /v2 migration) can translate it to the correct response via the
// central error handler.
//
// An optional Message overrides the default "Forbidden" text; this preserves
// the legacy per-site wording on sites like DoReadOne.
type ErrGenericForbidden struct {
	// Message overrides the default "Forbidden" text when non-empty.
	Message string
}

// Error implements the error interface.
func (e ErrGenericForbidden) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return "Forbidden"
}

// HTTPError implements web.HTTPErrorProcessor so the central error handler
// can translate the error to an HTTP 403 response regardless of which
// framework surfaces it.
func (e ErrGenericForbidden) HTTPError() web.HTTPError {
	msg := e.Message
	if msg == "" {
		msg = "Forbidden"
	}
	return web.HTTPError{HTTPCode: http.StatusForbidden, Message: msg}
}
