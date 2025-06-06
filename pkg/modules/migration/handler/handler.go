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

	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/migration"
	user2 "code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web/handler"
	"github.com/labstack/echo/v4"
)

var registeredMigrators map[string]*MigrationWeb

func init() {
	registeredMigrators = make(map[string]*MigrationWeb)
}

// MigrationWeb holds the web migration handler
type MigrationWeb struct {
	MigrationStruct func() migration.Migrator
}

// AuthURL is returned to the user when requesting the auth url
type AuthURL struct {
	URL string `json:"url"`
}

// RegisterMigrator registers all routes for migration
func (mw *MigrationWeb) RegisterMigrator(g *echo.Group) {
	ms := mw.MigrationStruct()
	g.GET("/"+ms.Name()+"/auth", mw.AuthURL)
	g.GET("/"+ms.Name()+"/status", mw.Status)
	g.POST("/"+ms.Name()+"/migrate", mw.Migrate)
	registeredMigrators[ms.Name()] = mw
}

// AuthURL is the web handler to get the auth url
func (mw *MigrationWeb) AuthURL(c echo.Context) error {
	ms := mw.MigrationStruct()
	return c.JSON(http.StatusOK, &AuthURL{URL: ms.AuthURL()})
}

// Migrate calls the migration method
func (mw *MigrationWeb) Migrate(c echo.Context) error {
	ms := mw.MigrationStruct()

	// Get the user from context
	user, err := user2.GetCurrentUser(c)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	stats, err := migration.GetMigrationStatus(ms, user)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	if !stats.StartedAt.IsZero() && stats.FinishedAt.IsZero() {
		return c.JSON(http.StatusPreconditionFailed, map[string]string{
			"message":       "Migration already running",
			"running_since": stats.StartedAt.String(),
		})
	}

	// Bind user request stuff
	err = c.Bind(ms)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "No or invalid model provided: "+err.Error()).SetInternal(err)
	}

	err = events.Dispatch(&MigrationRequestedEvent{
		Migrator:     ms,
		MigratorKind: ms.Name(),
		User:         user,
	})
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	return c.JSON(http.StatusOK, models.Message{Message: "Migration was started successfully."})
}

// Status returns whether or not a user has already done this migration
func (mw *MigrationWeb) Status(c echo.Context) error {
	ms := mw.MigrationStruct()

	return status(ms, c)
}
