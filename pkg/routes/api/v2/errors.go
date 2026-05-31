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

package apiv2

import (
	"context"
	"errors"
	"net/http"

	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/web"

	"github.com/danielgtaylor/huma/v2"
)

// authFromCtx retrieves the authed user from a Huma handler context,
// surfacing lookup failures as 401 instead of falling through to 500.
func authFromCtx(ctx context.Context) (web.Auth, error) {
	a, err := auth.GetAuthFromContext(ctx)
	if err != nil {
		// The underlying error can carry internal adapter/config detail
		// (e.g. a missing Echo context — a programming error, since the
		// token middleware authenticates before the handler runs). Log it
		// and return a generic 401 so nothing internal leaks to clients.
		log.Errorf("v2: could not resolve auth from context: %s", err)
		return nil, huma.Error401Unauthorized("invalid or missing authentication")
	}
	return a, nil
}

// translateDomainError maps a Vikunja domain error (web.HTTPErrorProcessor)
// onto Huma's status-error type so the response carries the right code
// and an RFC 9457 body. Errors without HTTP semantics fall through, which
// Huma treats as 500.
func translateDomainError(err error) error {
	if err == nil {
		return nil
	}
	var hp web.HTTPErrorProcessor
	if errors.As(err, &hp) {
		details := hp.HTTPError()
		msg := details.Message
		if msg == "" {
			msg = err.Error()
		}
		se := huma.NewError(details.HTTPCode, msg)
		// Preserve Vikunja's numeric domain error code (the value the
		// error docs key off) on the problem+json body. v1 exposes it as
		// `code`; without this v2 clients always read 0.
		if vm, ok := se.(*vikunjaErrorModel); ok {
			vm.Code = details.Code
		}
		return se
	}
	return err
}

// vikunjaErrorModel extends Huma's RFC 9457 body with Vikunja's numeric
// domain error code, preserving the v1 error-code contract on v2. Wired in
// as the global error type via the huma.NewError override in init().
type vikunjaErrorModel struct {
	huma.ErrorModel
	Code int `json:"code,omitempty" doc:"Vikunja numeric error code; see https://vikunja.io/docs/errors/"`
}

func init() {
	// Replace Huma's default error constructor so both the generated
	// OpenAPI schema and runtime responses use vikunjaErrorModel. Huma
	// derives the error-response schema from NewError(0, "") at register
	// time and routes runtime errors through the same constructor, so the
	// `code` field stays consistent between spec and wire.
	huma.NewError = func(status int, msg string, errs ...error) huma.StatusError {
		details := make([]*huma.ErrorDetail, 0, len(errs))
		for _, e := range errs {
			if e == nil {
				continue
			}
			if d, ok := e.(huma.ErrorDetailer); ok {
				details = append(details, d.ErrorDetail())
			} else {
				details = append(details, &huma.ErrorDetail{Message: e.Error()})
			}
		}
		return &vikunjaErrorModel{ErrorModel: huma.ErrorModel{
			Status: status,
			Title:  http.StatusText(status),
			Detail: msg,
			Errors: details,
		}}
	}

	// Strip internal detail from server errors. Huma's handler-error path
	// wraps a raw error as NewErrorWithContext(ctx, 500, "unexpected error
	// occurred", err) and — because the humaecho5 adapter writes the
	// response itself — bypasses Vikunja's CreateHTTPErrorHandler, which for
	// v1 returns a generic 500 with no detail. Without this override a raw
	// DB/driver error (SQL, table, column names) would leak into the
	// problem+json `errors[]`. Log the real cause, return a generic body.
	huma.NewErrorWithContext = func(_ huma.Context, status int, msg string, errs ...error) huma.StatusError {
		if status >= 500 {
			for _, e := range errs {
				if e != nil {
					log.Errorf("v2: internal server error: %s", e)
				}
			}
			errs = nil
		}
		return huma.NewError(status, msg, errs...)
	}
}
