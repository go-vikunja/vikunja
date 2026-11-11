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
	"code.vikunja.io/api/pkg/modules/migration"
	"code.vikunja.io/api/pkg/modules/migration/planka"

	"github.com/danielgtaylor/huma/v2"
)

// migrationCredentialsBody carries the connection details for migrators that pull data live from
// another instance the user has an account on. Either token or username + password must be given.
type migrationCredentialsBody struct {
	URL      string `json:"url" doc:"The base url of the instance to import from, e.g. https://planka.example.com."`
	Token    string `json:"token,omitempty" doc:"An API key or access token for the instance. Alternative to username and password."`
	Username string `json:"username,omitempty" doc:"The username or email to log in with. Only used when no token is given."`
	Password string `json:"password,omitempty" doc:"The password to log in with. Only used when no token is given."`
}

// RegisterMigrationCredentialsRoutes wires the credential-based migrators (Planka) onto the Huma API.
// They are always enabled since they need no server-side API keys.
func RegisterMigrationCredentialsRoutes(api huma.API) {
	registerCredentialsMigrator(api, func() migration.Migrator { return &planka.Migrator{} })
}

func init() { AddRouteRegistrar(RegisterMigrationCredentialsRoutes) }

// registerCredentialsMigrator registers status/migrate for a single credentials migrator. Unlike the
// OAuth migrators there is no auth url; the credentials travel in the migrate body.
func registerCredentialsMigrator(api huma.API, factory func() migration.Migrator) {
	name := factory().Name()
	tags := []string{"migration"}

	registerMigrationStatus(api, name, tags, factory)
	registerMigrationMigrate[migrationCredentialsBody](api, name, tags,
		"Starts a migration of the authenticated user's data from the given instance into Vikunja. The credentials are verified synchronously and rejected with 400 if the instance refuses them; the migration itself runs asynchronously. Refuses with 412 if a migration for this service is already running.",
		factory)
}
