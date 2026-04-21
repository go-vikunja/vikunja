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

	"code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/web"

	"github.com/danielgtaylor/huma/v2"
)

// authFromCtx retrieves the authed user from a Huma handler context,
// surfacing lookup failures as 401 instead of falling through to 500.
func authFromCtx(ctx context.Context) (web.Auth, error) {
	a, err := auth.GetAuthFromContext(ctx)
	if err != nil {
		return nil, huma.Error401Unauthorized(err.Error())
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
		return huma.NewError(details.HTTPCode, msg)
	}
	return err
}
