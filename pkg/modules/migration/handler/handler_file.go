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
	"io"
	"net/http"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/migration"
	user2 "code.vikunja.io/api/pkg/user"
	"github.com/labstack/echo/v5"
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

// RunFileMigration records the migration's start, runs the file migrator and
// records its finish. Shared by the v1 and v2 HTTP layers so the orchestration
// lives in one place; the caller supplies the already-opened upload.
func RunFileMigration(ms migration.FileMigrator, u *user2.User, file io.ReaderAt, size int64) error {
	m, err := migration.StartMigration(ms, u)
	if err != nil {
		return err
	}

	if err := ms.Migrate(u, file, size); err != nil {
		return err
	}

	return migration.FinishMigration(m)
}

// Migrate calls the migration method
func (fw *FileMigratorWeb) Migrate(c *echo.Context) error {
	ms := fw.MigrationStruct()

	// Get the user from context
	user, err := user2.GetCurrentUser(c)
	if err != nil {
		return err
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

	if err := RunFileMigration(ms, user, src, file.Size); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, models.Message{Message: "Everything was migrated successfully."})
}

// Status returns whether or not a user has already done this migration
func (fw *FileMigratorWeb) Status(c *echo.Context) error {
	ms := fw.MigrationStruct()

	return status(ms, c)
}
