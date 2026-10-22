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
	"encoding/json"
	"net/http"

	"code.vikunja.io/api/pkg/modules/migration"
	migrationHandler "code.vikunja.io/api/pkg/modules/migration/handler"
	"code.vikunja.io/api/pkg/user"

	"github.com/danielgtaylor/huma/v2"
)

type migrationStatusBody struct {
	Body *migration.Status
}

// migrationStartedBody confirms the migration was kicked off; the actual work
// runs asynchronously.
type migrationStartedBody struct {
	Body struct {
		Message string `json:"message" readOnly:"true" doc:"A confirmation message."`
	}
}

func registerMigrationStatus(api huma.API, name string, tags []string, factory func() migration.Migrator) {
	Register(api, huma.Operation{
		OperationID: "migration-" + name + "-status",
		Summary:     "Get the migration status for " + name,
		Description: "Returns the migration status of the authenticated user for this service, i.e. whether and when they last migrated. Used to prevent starting a second migration while one is running.",
		Method:      http.MethodGet,
		Path:        "/migration/" + name + "/status",
		Tags:        tags,
	}, func(ctx context.Context, _ *struct{}) (*migrationStatusBody, error) {
		return migrationStatus(ctx, factory)
	})
}

func migrationStatus(ctx context.Context, factory func() migration.Migrator) (*migrationStatusBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	u, err := user.GetFromAuth(a)
	if err != nil {
		return nil, translateDomainError(err)
	}

	status, err := migration.GetMigrationStatus(factory(), u)
	if err != nil {
		return nil, translateDomainError(err)
	}
	return &migrationStatusBody{Body: status}, nil
}

func migrationMigrate(ctx context.Context, factory func() migration.Migrator, body any) (*migrationStartedBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	u, err := user.GetFromAuth(a)
	if err != nil {
		return nil, translateDomainError(err)
	}

	ms := factory()
	// Apply the request payload onto the concrete migrator the same way v1's
	// c.Bind does, so migrator-specific field names (e.g. Trello's Token,
	// json:"code") bind transparently.
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(raw, ms); err != nil {
		return nil, huma.Error400BadRequest("invalid migration payload", err)
	}

	if err := migrationHandler.StartMigration(ms, u); err != nil {
		return nil, translateDomainError(err)
	}

	out := &migrationStartedBody{}
	out.Body.Message = "Migration was started successfully."
	return out, nil
}

// registerMigrationMigrate registers the migrate operation with B as its request body and makes the
// async event listener aware of the migrator.
func registerMigrationMigrate[B any](api huma.API, name string, tags []string, description string, factory func() migration.Migrator) {
	migrationHandler.RegisterMigratorForEvents(factory)

	Register(api, huma.Operation{
		OperationID: "migration-" + name + "-migrate",
		Summary:     "Migrate from " + name,
		Description: description,
		Method:      http.MethodPost,
		Path:        "/migration/" + name + "/migrate",
		// POST kicks off a job rather than creating a REST resource, so it
		// returns 200 with a confirmation, not the wrapper's 201.
		DefaultStatus: http.StatusOK,
		Tags:          tags,
	}, func(ctx context.Context, in *struct{ Body B }) (*migrationStartedBody, error) {
		return migrationMigrate(ctx, factory, in.Body)
	})
}
