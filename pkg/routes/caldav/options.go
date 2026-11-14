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
	"fmt"
	"net/http"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"

	"github.com/labstack/echo/v5"
)

const (
	davCompliance  = "1, 3, calendar-access"
	allowReadWrite = "GET, HEAD, PUT, DELETE, OPTIONS, PROPFIND, REPORT"
	allowReadOnly  = "GET, HEAD, OPTIONS, PROPFIND, REPORT"
	// The home set only lists the collections in it.
	allowHomeSet = "OPTIONS, PROPFIND, REPORT"
)

func serveOptions(c *echo.Context, u *user.User, t target) error {
	allow, err := allowed(u, t)
	if err != nil {
		return writeError(c, err)
	}

	c.Response().Header().Set("DAV", davCompliance)
	c.Response().Header().Set("Allow", allow)
	return c.NoContent(http.StatusOK)
}

// Advertising PUT and DELETE on a collection the user cannot write to would make clients
// retry the 403 forever.
func collectionAllow(u *user.User, collection collectionTarget) (string, error) {
	s := db.NewSession()
	defer s.Close()

	canWrite, err := collection.canWrite(s, u)
	if err != nil {
		return "", err
	}
	if canWrite {
		return allowReadWrite, nil
	}

	// The collection was loaded without a permission check, so answering here would
	// leak which project and filter ids exist.
	if _, err := collection.checkRead(s, u); err != nil {
		return "", err
	}

	// A pseudo collection cannot be written to as a collection, but PUT and DELETE
	// of the tasks it aggregates go through to their real projects.
	if collection.isPseudo() {
		return allowReadWrite, nil
	}
	return allowReadOnly, nil
}

// allowed is what t answers to, for the Allow header.
func allowed(u *user.User, t target) (string, error) {
	switch t := t.(type) {
	case entryTarget, principalTarget:
		return allowReadOnly, nil
	case homeSetTarget:
		return allowHomeSet, nil
	case collectionTarget:
		return collectionAllow(u, t)
	case taskTarget:
		return collectionAllow(u, t.collection)
	}
	return "", fmt.Errorf("unhandled caldav target %T", t)
}

func methodNotAllowed(c *echo.Context, u *user.User, t target) error {
	allow, err := allowed(u, t)
	if err != nil {
		return writeError(c, err)
	}
	c.Response().Header().Set("Allow", allow)
	return c.NoContent(http.StatusMethodNotAllowed)
}
