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

package task

import (
	"net/http"
	"strconv"

	"code.vikunja.io/api/pkg/services"
	"code.vikunja.io/api/pkg/user"
	"github.com/labstack/echo/v4"
	"xorm.io/xorm"
)

// Get is the handler to get a single task.
func Get(s *xorm.Session, u *user.User, c echo.Context) error {
	taskID, err := strconv.ParseInt(c.Param("projecttask"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid task ID").SetInternal(err)
	}

	taskService := services.NewTaskService(s.Engine())
	task, err := taskService.GetByID(s, taskID, u)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, task)
}
