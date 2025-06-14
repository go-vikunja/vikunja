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

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/migration"
	user2 "code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web/handler"
	"github.com/labstack/echo/v4"
)

type FileMigratorWeb struct {
	MigrationStruct func() migration.FileMigrator
}

// RegisterRoutes registers all routes for migration
func (fw *FileMigratorWeb) RegisterRoutes(g *echo.Group) {
	ms := fw.MigrationStruct()
	g.GET("/"+ms.Name()+"/status", fw.Status)
	g.PUT("/"+ms.Name()+"/migrate", fw.Migrate)
}

// Migrate calls the migration method
func (fw *FileMigratorWeb) Migrate(c echo.Context) error {
	ms := fw.MigrationStruct()

	// Get the user from context
	user, err := user2.GetCurrentUser(c)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	file, err := c.FormFile("import")
	if err != nil {
		return err
	}
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	m, err := migration.StartMigration(ms, user)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	// Do the migration
	err = ms.Migrate(user, src, file.Size)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	err = migration.FinishMigration(m)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	return c.JSON(http.StatusOK, models.Message{Message: "Everything was migrated successfully."})
}

// Status returns whether or not a user has already done this migration
func (fw *FileMigratorWeb) Status(c echo.Context) error {
	ms := fw.MigrationStruct()

	return status(ms, c)
}
