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

	"code.vikunja.io/api/pkg/user"

	"github.com/labstack/echo/v5"
	"xorm.io/xorm"
)

func serveDelete(c *echo.Context, u *user.User, t target) error {
	switch t := t.(type) {
	case entryTarget, principalTarget, homeSetTarget:
		return methodNotAllowed(c, u, t)
	case collectionTarget:
		return methodNotAllowed(c, u, t)
	case taskTarget:
		return deleteTaskTarget(c, u, t)
	}
	return writeError(c, fmt.Errorf("unhandled caldav target %T", t))
}

func deleteTaskTarget(c *echo.Context, u *user.User, t taskTarget) error {
	preconditionFailed := false
	err := writeTransaction(func(s *xorm.Session) error {
		collection, err := t.collection.checkRead(s, u)
		if err != nil {
			return err
		}
		task, err := findTask(s, u, t.collection, t.uid)
		if err != nil {
			return err
		}
		if task == nil {
			return errNotFound
		}
		if !preconditionsOf(c.Request().Header).allowDelete(taskETag(task)) {
			preconditionFailed = true
			return nil
		}
		return deleteTask(s, u, collection, task)
	})
	if err != nil {
		return writeError(c, err)
	}
	if preconditionFailed {
		return c.NoContent(http.StatusPreconditionFailed)
	}
	return c.NoContent(http.StatusNoContent)
}
