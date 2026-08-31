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

package apiv2

import (
	"context"
	"net/http"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/modules/migration"
	migrationHandler "code.vikunja.io/api/pkg/modules/migration/handler"
	microsofttodo "code.vikunja.io/api/pkg/modules/migration/microsoft-todo"
	"code.vikunja.io/api/pkg/modules/migration/todoist"
	"code.vikunja.io/api/pkg/modules/migration/trello"

	"github.com/danielgtaylor/huma/v2"
)

// migrationAuthURLBody is the response for the OAuth auth-url endpoint.
type migrationAuthURLBody struct {
	Body migrationHandler.AuthURL
}

// migrationMigrateBody carries the OAuth code obtained from the auth url back
// to the server. It is applied onto the concrete migrator (whose field carries
// json:"code") so it works across migrators regardless of their field name.
type migrationMigrateBody struct {
	Code string `json:"code" doc:"The OAuth code obtained after authorizing against the auth url."`
}

// RegisterMigrationOAuthRoutes wires the OAuth-based migrators (Todoist, Trello,
// Microsoft To-Do) onto the Huma API. Each migrator is gated behind its static
// config flag and exposes the same three operations, so registration is driven
// by one generic helper instead of three copy-pasted blocks.
func RegisterMigrationOAuthRoutes(api huma.API) {
	registerOAuthMigrator(api, config.MigrationTodoistEnable.GetBool(), func() migration.Migrator { return &todoist.Migration{} })
	registerOAuthMigrator(api, config.MigrationTrelloEnable.GetBool(), func() migration.Migrator { return &trello.Migration{} })
	registerOAuthMigrator(api, config.MigrationMicrosoftTodoEnable.GetBool(), func() migration.Migrator { return &microsofttodo.Migration{} })
}

func init() { AddRouteRegistrar(RegisterMigrationOAuthRoutes) }

// registerOAuthMigrator registers auth/status/migrate for a single OAuth
// migrator. enabled gates the whole migrator (config early-return, no
// middleware); factory produces a fresh migrator instance per request, matching
// v1's MigrationStruct func so concurrent requests never share mutable state.
func registerOAuthMigrator(api huma.API, enabled bool, factory func() migration.Migrator) {
	if !enabled {
		return
	}

	name := factory().Name()
	tags := []string{"migration"}

	Register(api, huma.Operation{
		OperationID: "migration-" + name + "-auth",
		Summary:     "Get the auth url for " + name,
		Description: "Returns the OAuth url the user needs to authenticate against. The code obtained there is passed back to the migrate endpoint.",
		Method:      http.MethodGet,
		Path:        "/migration/" + name + "/auth",
		Tags:        tags,
	}, func(_ context.Context, _ *struct{}) (*migrationAuthURLBody, error) {
		return &migrationAuthURLBody{Body: migrationHandler.AuthURL{URL: factory().AuthURL()}}, nil
	})

	registerMigrationStatus(api, name, tags, factory)
	registerMigrationMigrate[migrationMigrateBody](api, name, tags,
		"Starts a migration of the authenticated user's data from this service into Vikunja. The migration runs asynchronously; this returns once it has been queued. Refuses with 412 if a migration for this service is already running.",
		factory)
}
