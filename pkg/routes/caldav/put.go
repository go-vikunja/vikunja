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
	"io"
	"net/http"
	"strings"

	caldav2 "code.vikunja.io/api/pkg/caldav"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/labstack/echo/v5"
	"xorm.io/xorm"
)

func servePut(c *echo.Context, u *user.User, t target) error {
	switch t := t.(type) {
	case entryTarget, principalTarget:
		// Pinned: caldav-go finds an empty resource here, then refuses to update it.
		if !preconditionsOf(c.Request().Header).allowUpdate("") {
			return c.NoContent(http.StatusPreconditionFailed)
		}
		return writeError(c, errForbidden)
	case homeSetTarget:
		return serveProjects(c, u, t)
	case collectionTarget:
		return putOnCollection(c, u, t)
	case taskTarget:
		body, err := readPutBody(c)
		if err != nil {
			return writeError(c, err)
		}
		return put(c, u, t.collection, body, taskPut{uid: t.uid})
	}
	return writeError(c, fmt.Errorf("unhandled caldav target %T", t))
}

func readPutBody(c *echo.Context) (string, error) {
	body, err := io.ReadAll(io.LimitReader(c.Request().Body, caldav2.MaxVCalendarSize+1))
	if err != nil {
		return "", fmt.Errorf("%w: %w", errMalformedBody, err)
	}
	logRequest(c.Request().Header, body)
	return string(body), nil
}

// Deliberate legacy support: a VCALENDAR PUT onto a collection URL creates or updates the
// task its UID names.
func putOnCollection(c *echo.Context, u *user.User, collection collectionTarget) error {
	body, err := readPutBody(c)
	if err != nil {
		return writeError(c, err)
	}

	if !strings.HasPrefix(body, "BEGIN:VCALENDAR") {
		s := db.NewSession()
		defer s.Close()
		if _, err := collection.checkRead(s, u); err != nil {
			return writeError(c, err)
		}
		return c.NoContent(http.StatusPreconditionFailed)
	}

	return put(c, u, collection, body, collectionPut{})
}

// putAddressing is how a PUT names its task: by the URL, or by the body's UID on a collection.
type putAddressing interface {
	parse(body string) (*models.Task, caldav2.ParsedVTODOProperties, error)
	taskUID(parsed *models.Task) string
	etag(stored, parsed *models.Task) string
}

type taskPut struct{ uid string }

func (taskPut) parse(body string) (*models.Task, caldav2.ParsedVTODOProperties, error) {
	return parseTaskBody(body)
}

func (p taskPut) taskUID(*models.Task) string { return p.uid }

func (taskPut) etag(stored, _ *models.Task) string { return taskETag(stored) }

type collectionPut struct{}

func (collectionPut) parse(body string) (*models.Task, caldav2.ParsedVTODOProperties, error) {
	vTask, props, err := caldav2.ParseTaskFromVTODO(body)
	if err != nil {
		log.Warningf("[CALDAV] Failed to parse task: %v", err)
		return nil, props, models.ErrInvalidData{Message: "Invalid task"}
	}
	return vTask, props, nil
}

func (collectionPut) taskUID(parsed *models.Task) string { return parsed.UID }

// Pinned: caldav-go compared If-Match against an ETag carrying the body's DTSTAMP.
func (collectionPut) etag(stored, parsed *models.Task) string {
	if parsed.Updated.Unix() > 0 {
		withDTStamp := *stored
		withDTStamp.Updated = parsed.Updated
		return taskETag(&withDTStamp)
	}
	return taskETag(stored)
}

func put(c *echo.Context, u *user.User, collection collectionTarget, body string, addressing putAddressing) error {
	conditions := preconditionsOf(c.Request().Header)
	status := http.StatusPreconditionFailed
	var etag string

	err := writeTransaction(func(s *xorm.Session) error {
		readable, err := collection.checkRead(s, u)
		if err != nil {
			return err
		}
		if err := caldav2.CheckVCalendarLimits(body); err != nil {
			return err
		}
		vTask, props, err := addressing.parse(body)
		if err != nil {
			return err
		}

		stored, err := findTask(s, u, collection, addressing.taskUID(vTask))
		if err != nil {
			return err
		}

		if stored == nil {
			if !conditions.allowCreate() {
				return nil
			}
			if err := createTask(s, u, readable, vTask); err != nil {
				return err
			}
			status, etag = http.StatusCreated, taskETag(vTask)
			return nil
		}

		if !conditions.allowUpdate(addressing.etag(stored, vTask)) {
			return nil
		}
		updated, err := updateTask(s, u, readable, stored, vTask, props)
		if err != nil {
			return err
		}
		status, etag = http.StatusNoContent, taskETag(updated)
		return nil
	})
	if err != nil {
		return writeError(c, err)
	}

	if etag != "" {
		c.Response().Header().Set("ETag", etag)
	}
	return c.NoContent(status)
}
