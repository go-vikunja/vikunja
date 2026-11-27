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
	case entryTarget, principalTarget, homeSetTarget:
		return methodNotAllowed(c, u, t)
	case collectionTarget:
		return putOnCollection(c, u, t)
	case taskTarget:
		body, err := readPutBody(c)
		if err != nil {
			return writeError(c, err)
		}
		return put(c, u, t.collection, body, t.uid)
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

// Deliberate legacy support: a VCALENDAR PUT onto a collection writes the task its UID names.
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

	return put(c, u, collection, body, "")
}

func put(c *echo.Context, u *user.User, collection collectionTarget, body, urlUID string) error {
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
		var vTask *models.Task
		var props caldav2.ParsedVTODOProperties
		uid := urlUID
		if uid != "" {
			vTask, props, err = parseTaskBody(body)
			if err != nil {
				return err
			}
		} else {
			vTask, props, err = caldav2.ParseTaskFromVTODO(body)
			if err != nil {
				log.Warningf("[CALDAV] Failed to parse task: %v", err)
				return models.ErrInvalidData{Message: "Invalid task"}
			}
			uid = vTask.UID
		}

		// A body without UID names no stored task; creating it gives it one.
		var stored *models.Task
		if uid != "" {
			stored, err = findTask(s, u, collection, uid)
			if err != nil {
				return err
			}
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

		if !conditions.allowUpdate(taskETag(stored)) {
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
