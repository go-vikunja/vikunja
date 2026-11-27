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

package caldav

import (
	"errors"
	"net/http"

	caldavpkg "code.vikunja.io/api/pkg/caldav"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/richtext"
	"github.com/labstack/echo/v5"
)

var (
	// Also for resources the user cannot read: those must look missing.
	errNotFound      = errors.New("caldav: not found")
	errForbidden     = errors.New("caldav: forbidden")
	errMalformedBody = errors.New("caldav: malformed request body")
	errBodyTooLarge  = errors.New("caldav: request body too large")
)

func isError[T error](err error) bool {
	var target T
	return errors.As(err, &target)
}

func errorStatus(err error) int {
	switch {
	case errors.Is(err, errNotFound),
		isError[models.ErrProjectDoesNotExist](err),
		isError[models.ErrSavedFilterDoesNotExist](err),
		isError[models.ErrTaskDoesNotExist](err):
		return http.StatusNotFound
	case errors.Is(err, errForbidden),
		isError[models.ErrProjectIsArchived](err),
		// Pinned: caldav-go could not answer 400.
		isError[models.ValidationHTTPError](err),
		isError[models.ErrTaskCannotBeEmpty](err),
		errors.Is(err, richtext.ErrTooManyTableCells):
		return http.StatusForbidden
	case errors.Is(err, errBodyTooLarge),
		errors.Is(err, caldavpkg.ErrVCalendarTooLarge):
		return http.StatusRequestEntityTooLarge
	case errors.Is(err, errMalformedBody),
		errors.Is(err, caldavpkg.ErrTooManyPropertyParameters),
		isError[models.ErrInvalidData](err):
		return http.StatusBadRequest
	}
	return http.StatusInternalServerError
}

func writeError(c *echo.Context, err error) error {
	status := errorStatus(err)
	if status == http.StatusInternalServerError {
		log.Errorf("[CALDAV] %s %s: %v", c.Request().Method, c.Request().URL.Path, err)
		return echo.NewHTTPError(status, "internal server error").Wrap(err)
	}
	if c.Request().Method == http.MethodHead {
		return c.NoContent(status)
	}
	return c.String(status, http.StatusText(status))
}
