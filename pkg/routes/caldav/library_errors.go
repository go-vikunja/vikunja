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

	"github.com/samedi/caldav-go/errs"
)

// caldav-go answers 500 for anything but its own sentinels; both go with the library.
func toLibraryError(err error) error {
	switch errorStatus(err) {
	case http.StatusNotFound:
		return errs.ResourceNotFoundError
	case http.StatusForbidden:
		return errs.ForbiddenError
	}
	return err
}

func fromLibraryError(err error) error {
	switch {
	case errors.Is(err, errs.ResourceNotFoundError):
		return errNotFound
	case errors.Is(err, errs.ForbiddenError):
		return errForbidden
	}
	return err
}
