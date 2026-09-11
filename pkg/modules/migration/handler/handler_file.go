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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/migration"
	user2 "code.vikunja.io/api/pkg/user"
	"github.com/labstack/echo/v5"
)

var registeredFileMigrators map[string]func() migration.FileMigrator

func init() {
	registeredFileMigrators = make(map[string]func() migration.FileMigrator)
}

type FileMigratorWeb struct {
	MigrationStruct func() migration.FileMigrator
}

// RegisterRoutes registers all routes for migration
func (fw *FileMigratorWeb) RegisterRoutes(g *echo.Group) {
	ms := fw.MigrationStruct()
	g.GET("/"+ms.Name()+"/status", fw.Status)
	g.PUT("/"+ms.Name()+"/migrate", fw.Migrate)
	RegisterFileMigrator(fw.MigrationStruct)
}

// RegisterFileMigrator makes a file migrator resolvable by the background listener.
func RegisterFileMigrator(factory func() migration.FileMigrator) {
	registeredFileMigrators[factory().Name()] = factory
}

// It returns as soon as the job is queued: an import of a large export runs for
// minutes, far longer than a reverse proxy will hold a request open, and a
// client that gives up waiting cannot abort the import it started.
func StartFileMigration(ms migration.FileMigrator, u *user2.User, file io.ReaderAt, size int64, options []byte) error {
	// The listener applies these again on its own instance; doing it here too
	// turns an unusable config into a failed request instead of a failed job.
	if err := applyMigratorOptions(ms, options); err != nil {
		return err
	}

	// Validating before the claim means a wrong file doesn't occupy the slot.
	if v, ok := ms.(migration.FileValidator); ok {
		if err := v.ValidateFile(file, size); err != nil {
			return asImportFileError(err)
		}
	}

	status, err := migration.ClaimMigration(ms, u)
	if err != nil {
		return err
	}

	uploadName, uploadSize, err := migration.SpoolUpload(io.NewSectionReader(file, 0, size))
	if err != nil {
		releaseClaim(status, u, "failed upload spooling")
		return err
	}
	if uploadSize != size {
		migration.RemoveSpooledUpload(uploadName)
		releaseClaim(status, u, "short upload spooling")
		return fmt.Errorf("spooled %d bytes of the %d byte upload", uploadSize, size)
	}

	if err := events.Dispatch(&FileMigrationRequestedEvent{
		User:              u,
		MigratorKind:      ms.Name(),
		MigrationStatusID: status.ID,
		UploadName:        uploadName,
		UploadSize:        uploadSize,
		Options:           options,
	}); err != nil {
		migration.RemoveSpooledUpload(uploadName)
		releaseClaim(status, u, "failed event dispatch")
		return err
	}

	return nil
}

func applyMigratorOptions(ms migration.FileMigrator, options []byte) error {
	if len(options) == 0 {
		return nil
	}
	o, ok := ms.(migration.FileMigratorOptions)
	if !ok {
		return fmt.Errorf("migrator %s does not accept options", ms.Name())
	}
	return o.SetOptions(options)
}

// asImportFileError maps a decode failure to a 400: a file migrator only ever
// parses user-supplied data, so a broken document is never a server fault.
func asImportFileError(err error) error {
	var syntaxErr *json.SyntaxError
	var typeErr *json.UnmarshalTypeError
	// A truncated document surfaces as io.ErrUnexpectedEOF rather than a SyntaxError.
	if errors.As(err, &syntaxErr) || errors.As(err, &typeErr) ||
		errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, io.EOF) {
		return &migration.ErrInvalidImportFile{Err: err}
	}
	return err
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

	if err := StartFileMigration(ms, user, src, file.Size, nil); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, models.Message{Message: "Migration was started successfully."})
}

// Status returns whether or not a user has already done this migration
func (fw *FileMigratorWeb) Status(c *echo.Context) error {
	ms := fw.MigrationStruct()

	return status(ms, c)
}
