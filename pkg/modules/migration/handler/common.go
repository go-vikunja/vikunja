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

package handler

import (
	"net/http"

	"code.vikunja.io/api/pkg/modules/migration"
	user2 "code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web/handler"
	"github.com/labstack/echo/v4"
)

func status(ms migration.MigratorName, c echo.Context) error {
	user, err := user2.GetCurrentUser(c)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	status, err := migration.GetMigrationStatus(ms, user)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	return c.JSON(http.StatusOK, status)
}
