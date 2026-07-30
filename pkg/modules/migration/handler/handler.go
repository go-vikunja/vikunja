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
	"github.com/labstack/echo/v5"
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
	URL string `json:"url" readOnly:"true" doc:"The OAuth authorization url the client should redirect the user to. After authorizing, the obtained code is passed back to the migrate endpoint."`
}

// RegisterMigrator registers all routes for migration. The /auth route is
// only registered for migrators using an OAuth flow - token-based migrators
// have no auth url to hand out.
func (mw *MigrationWeb) RegisterMigrator(g *echo.Group) {
	ms := mw.MigrationStruct()
	if _, isOAuth := ms.(migration.OAuthMigrator); isOAuth {
		g.GET("/"+ms.Name()+"/auth", mw.AuthURL)
	}
	g.GET("/"+ms.Name()+"/status", mw.Status)
	g.POST("/"+ms.Name()+"/migrate", mw.Migrate)
	registeredMigrators[ms.Name()] = mw
}

// AuthURL is the web handler to get the auth url
func (mw *MigrationWeb) AuthURL(c *echo.Context) error {
	ms, ok := mw.MigrationStruct().(migration.OAuthMigrator)
	if !ok {
		// Not reachable through the router - the route is only registered for
		// OAuth migrators - but guard against future direct calls.
		return echo.NewHTTPError(http.StatusNotFound, "This migrator does not use an auth url.")
	}
	return c.JSON(http.StatusOK, &AuthURL{URL: ms.AuthURL()})
}

// StartMigration kicks off a migration for the given user: it refuses with
// migration.ErrMigrationAlreadyRunning if one is already in progress, then
// dispatches the MigrationRequestedEvent that runs the migration asynchronously.
// The migrator must already carry its request payload (e.g. the OAuth code).
// Shared by the v1 and v2 HTTP layers so the orchestration lives in one place.
func StartMigration(ms migration.Migrator, u *user2.User) error {
	stats, err := migration.GetMigrationStatus(ms, u)
	if err != nil {
		return err
	}

	if !stats.StartedAt.IsZero() && stats.FinishedAt.IsZero() {
		return &migration.ErrMigrationAlreadyRunning{StartedAt: stats.StartedAt}
	}

	return events.Dispatch(&MigrationRequestedEvent{
		Migrator:     ms,
		MigratorKind: ms.Name(),
		User:         u,
	})
}

// Migrate calls the migration method
func (mw *MigrationWeb) Migrate(c *echo.Context) error {
	ms := mw.MigrationStruct()

	// Get the user from context
	user, err := user2.GetCurrentUser(c)
	if err != nil {
		return err
	}

	stats, err := migration.GetMigrationStatus(ms, user)
	if err != nil {
		return err
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
		return echo.NewHTTPError(http.StatusBadRequest, "No or invalid model provided: "+err.Error()).Wrap(err)
	}

	if err := StartMigration(ms, user); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, models.Message{Message: "Migration was started successfully."})
}

// Status returns whether or not a user has already done this migration
func (mw *MigrationWeb) Status(c *echo.Context) error {
	ms := mw.MigrationStruct()

	return status(ms, c)
}
