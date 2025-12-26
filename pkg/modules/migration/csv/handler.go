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

package csv

import (
	"encoding/json"
	"net/http"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/migration"
	user2 "code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web/handler"
	"github.com/labstack/echo/v4"
)

// MigratorWeb handles CSV migration HTTP routes
type MigratorWeb struct{}

// RegisterRoutes registers all CSV migration routes
func (c *MigratorWeb) RegisterRoutes(g *echo.Group) {
	g.GET("/csv/status", c.Status)
	g.PUT("/csv/detect", c.Detect)
	g.PUT("/csv/preview", c.Preview)
	g.PUT("/csv/migrate", c.Migrate)
}

// Status returns the migration status
// @Summary Get CSV migration status
// @Description Returns if the current user already did the CSV migration or not.
// @tags migration
// @Produce json
// @Security JWTKeyAuth
// @Success 200 {object} migration.Status "The migration status"
// @Failure 500 {object} models.Message "Internal server error"
// @Router /migration/csv/status [get]
func (c *MigratorWeb) Status(ctx echo.Context) error {
	u, err := user2.GetCurrentUser(ctx)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	m := &Migrator{}
	s, err := migration.GetMigrationStatus(m, u)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	return ctx.JSON(http.StatusOK, s)
}

// Detect analyzes a CSV file and returns detection results
// @Summary Detect CSV structure
// @Description Analyzes a CSV file and returns auto-detected columns, delimiter, quote character, and date format with suggested column mappings.
// @tags migration
// @Accept multipart/form-data
// @Produce json
// @Security JWTKeyAuth
// @Param import formData file true "The CSV file to analyze"
// @Success 200 {object} DetectionResult "Detection results with suggested mappings"
// @Failure 400 {object} models.Message "Invalid CSV file"
// @Failure 500 {object} models.Message "Internal server error"
// @Router /migration/csv/detect [put]
func (c *MigratorWeb) Detect(ctx echo.Context) error {
	_, err := user2.GetCurrentUser(ctx)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	file, err := ctx.FormFile("import")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "No file provided")
	}

	src, err := file.Open()
	if err != nil {
		return handler.HandleHTTPError(err)
	}
	defer src.Close()

	result, err := DetectCSVStructure(src, file.Size)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	return ctx.JSON(http.StatusOK, result)
}

// Preview generates a preview of the import
// @Summary Preview CSV import
// @Description Generates a preview of the first 5 tasks that would be imported with the given configuration.
// @tags migration
// @Accept multipart/form-data
// @Produce json
// @Security JWTKeyAuth
// @Param import formData file true "The CSV file to preview"
// @Param config formData string true "The import configuration JSON"
// @Success 200 {object} PreviewResult "Preview of tasks to import"
// @Failure 400 {object} models.Message "Invalid CSV file or configuration"
// @Failure 500 {object} models.Message "Internal server error"
// @Router /migration/csv/preview [put]
func (c *MigratorWeb) Preview(ctx echo.Context) error {
	_, err := user2.GetCurrentUser(ctx)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	file, err := ctx.FormFile("import")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "No file provided")
	}

	configStr := ctx.FormValue("config")
	if configStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "No configuration provided")
	}

	var config ImportConfig
	if err := json.Unmarshal([]byte(configStr), &config); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid configuration: "+err.Error())
	}

	src, err := file.Open()
	if err != nil {
		return handler.HandleHTTPError(err)
	}
	defer src.Close()

	result, err := PreviewImport(src, file.Size, config)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	return ctx.JSON(http.StatusOK, result)
}

// Migrate imports the CSV file
// @Summary Import CSV file
// @Description Imports tasks from a CSV file into Vikunja with the provided configuration.
// @tags migration
// @Accept multipart/form-data
// @Produce json
// @Security JWTKeyAuth
// @Param import formData file true "The CSV file to import"
// @Param config formData string true "The import configuration JSON"
// @Success 200 {object} models.Message "A message telling you everything was migrated successfully."
// @Failure 400 {object} models.Message "Invalid CSV file or configuration"
// @Failure 500 {object} models.Message "Internal server error"
// @Router /migration/csv/migrate [put]
func (c *MigratorWeb) Migrate(ctx echo.Context) error {
	u, err := user2.GetCurrentUser(ctx)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	file, err := ctx.FormFile("import")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "No file provided")
	}

	configStr := ctx.FormValue("config")
	if configStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "No configuration provided")
	}

	var config ImportConfig
	if err := json.Unmarshal([]byte(configStr), &config); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid configuration: "+err.Error())
	}

	src, err := file.Open()
	if err != nil {
		return handler.HandleHTTPError(err)
	}
	defer src.Close()

	m := &Migrator{}
	status, err := migration.StartMigration(m, u)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	err = MigrateWithConfig(u, src, file.Size, config)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	err = migration.FinishMigration(status)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	return ctx.JSON(http.StatusOK, models.Message{Message: "Everything was migrated successfully."})
}
