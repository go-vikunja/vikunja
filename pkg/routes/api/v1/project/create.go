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

package project

import (
	"net/http"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/services"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/labstack/echo/v4"
)

// CreateProject is the handler to create a project
func CreateProject(c echo.Context) error {
	p := new(models.Project)
	if err := c.Bind(p); err != nil {
		return err
	}

	u, err := user.GetCurrentUser(c)
	if err != nil {
		return err
	}

	s := db.NewSession()
	defer s.Close()

	projectService := services.Project{DB: db.GetEngine()}
	createdProject, err := projectService.Create(s, p, u)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	return c.JSON(http.StatusCreated, createdProject)
}
