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
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"

	"github.com/labstack/echo/v5"
)

const (
	contentTypeCollection = "text/calendar"
	contentTypeTask       = "text/calendar; component=vcalendar"
)

func serveGet(c *echo.Context, u *user.User, t target) error {
	switch t := t.(type) {
	case entryTarget, principalTarget:
		// Pinned: caldav-go answers these as an empty calendar object.
		return writeCalendar(c, contentTypeTask, "", time.Time{}, "")
	case homeSetTarget:
		return methodNotAllowed(c, u, t)
	case collectionTarget:
		return getCollection(c, u, t)
	case taskTarget:
		return getTask(c, u, t)
	}
	return writeError(c, fmt.Errorf("unhandled caldav target %T", t))
}

func getCollection(c *echo.Context, u *user.User, collection collectionTarget) error {
	s := db.NewSession()
	defer s.Close()

	readable, err := collection.checkRead(s, u)
	if err != nil {
		return writeError(c, err)
	}
	tasks, err := readCollectionTasks(s, u, readable)
	if err != nil {
		return writeError(c, err)
	}
	if err := s.Commit(); err != nil {
		return writeError(c, err)
	}

	// Pinned: a collection without tasks answers an empty body, not an empty calendar.
	var body string
	if tasks != nil {
		body = renderCollection(collection.project, tasks)
	}
	return writeCalendar(c, contentTypeCollection, collectionETag(collection.project, tasks), collection.project.Updated, body)
}

func getTask(c *echo.Context, u *user.User, t taskTarget) error {
	s := db.NewSession()
	defer s.Close()

	task, err := findTask(s, u, t.collection, t.uid)
	if err != nil {
		return writeError(c, err)
	}
	if task == nil {
		return writeError(c, errNotFound)
	}
	if err := s.Commit(); err != nil {
		return writeError(c, err)
	}

	return writeCalendar(c, contentTypeTask, taskETag(task), task.Updated, renderTask(task))
}

func writeCalendar(c *echo.Context, contentType, etag string, modified time.Time, body string) error {
	h := c.Response().Header()
	h.Set("ETag", etag)
	if etag != "" && preconditionsOf(c.Request().Header).notModified(etag) {
		return c.NoContent(http.StatusNotModified)
	}
	h.Set("Last-Modified", lastModified(modified))
	if c.Request().Method == http.MethodHead {
		h.Set("Content-Type", contentType)
		return c.NoContent(http.StatusOK)
	}
	return c.Blob(http.StatusOK, contentType, []byte(body))
}
