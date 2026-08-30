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
	"code.vikunja.io/api/pkg/log"
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

// RegisterMigrator registers all routes for migration
func (mw *MigrationWeb) RegisterMigrator(g *echo.Group) {
	ms := mw.MigrationStruct()
	g.GET("/"+ms.Name()+"/auth", mw.AuthURL)
	g.GET("/"+ms.Name()+"/status", mw.Status)
	g.POST("/"+ms.Name()+"/migrate", mw.Migrate)
	RegisterMigratorForEvents(mw.MigrationStruct)
}

// RegisterMigratorForEvents makes a migrator known to the migration listener without exposing v1 routes.
func RegisterMigratorForEvents(factory func() migration.Migrator) {
	registeredMigrators[factory().Name()] = &MigrationWeb{MigrationStruct: factory}
}

// AuthURL is the web handler to get the auth url
func (mw *MigrationWeb) AuthURL(c *echo.Context) error {
	ms := mw.MigrationStruct()
	return c.JSON(http.StatusOK, &AuthURL{URL: ms.AuthURL()})
}

// StartMigration validates credentials and dispatches a migration while holding its claim.
func StartMigration(ms migration.Migrator, u *user2.User) error {
	status, err := migration.ClaimMigration(ms, u)
	if err != nil {
		return err
	}

	if cc, ok := ms.(migration.CredentialsChecker); ok {
		if err := cc.CheckCredentials(); err != nil {
			releaseClaim(status, u, "failed credential check")
			return err
		}
	}

	if err := events.Dispatch(&MigrationRequestedEvent{
		Migrator:          ms,
		MigratorKind:      ms.Name(),
		User:              u,
		MigrationStatusID: status.ID,
	}); err != nil {
		releaseClaim(status, u, "failed event dispatch")
		return err
	}

	return nil
}

func releaseClaim(status *migration.Status, u *user2.User, reason string) {
	if ferr := migration.FinishMigration(status); ferr != nil {
		log.Errorf("[Migration] Could not release claim of migration %d for user %d after %s: %s", status.ID, u.ID, reason, ferr)
	}
}

// Migrate calls the migration method
func (mw *MigrationWeb) Migrate(c *echo.Context) error {
	ms := mw.MigrationStruct()

	// Get the user from context
	user, err := user2.GetCurrentUser(c)
	if err != nil {
		return err
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
