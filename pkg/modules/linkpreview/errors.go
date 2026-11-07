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

package linkpreview

import (
	"errors"
	"fmt"
	"net"
	"net/http"

	"code.vikunja.io/api/pkg/web"
)

type ErrInvalidURL struct{}

func (err *ErrInvalidURL) Error() string {
	return "Link previews require an external HTTP(S) URL without credentials"
}

func (err *ErrInvalidURL) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusUnprocessableEntity,
		Message:  err.Error(),
	}
}

type ErrPreviewUnavailable struct {
	Err error
}

func (err *ErrPreviewUnavailable) Error() string {
	return fmt.Sprintf("Link preview unavailable: %s", err.Err)
}

func (err *ErrPreviewUnavailable) Unwrap() error { return err.Err }

func (err *ErrPreviewUnavailable) HTTPError() web.HTTPError {
	status := http.StatusBadGateway
	var networkError net.Error
	if errors.As(err.Err, &networkError) && networkError.Timeout() {
		status = http.StatusGatewayTimeout
	}
	return web.HTTPError{
		HTTPCode: status,
		Message:  "Link preview unavailable",
	}
}
