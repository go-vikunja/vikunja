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
	"code.vikunja.io/api/pkg/modules/migration/ticktick"
	vikunja_file "code.vikunja.io/api/pkg/modules/migration/vikunja-file"
	"code.vikunja.io/api/pkg/modules/migration/wekan"
	"code.vikunja.io/api/pkg/user"

	"github.com/danielgtaylor/huma/v2"
)

// fileMigrateInput is the multipart upload body shared by every file migrator's
// migrate endpoint.
type fileMigrateInput struct {
	RawBody huma.MultipartFormFiles[struct {
		Import huma.FormFile `form:"import" required:"true" doc:"The export file to import. Its expected format depends on the migrator (e.g. a Vikunja export zip, a TickTick CSV, a WeKan JSON export)."`
	}]
}

// RegisterMigrationFileRoutes wires the file-based migrators (Vikunja export,
// TickTick, WeKan) onto the Huma API. Unlike the OAuth migrators these have no
// config flag in v1, so they are always registered.
func RegisterMigrationFileRoutes(api huma.API) {
	registerFileMigrator(api, func() migration.FileMigrator { return &vikunja_file.FileMigrator{} })
	registerFileMigrator(api, func() migration.FileMigrator { return &ticktick.Migrator{} })
	registerFileMigrator(api, func() migration.FileMigrator { return &wekan.Migrator{} })
}

func init() { AddRouteRegistrar(RegisterMigrationFileRoutes) }

// registerFileMigrator registers status + migrate for a single file migrator.
// factory produces a fresh migrator instance per request, matching v1's
// MigrationStruct func so concurrent requests never share mutable state.
func registerFileMigrator(api huma.API, factory func() migration.FileMigrator) {
	name := factory().Name()
	tags := []string{"migration"}

	Register(api, huma.Operation{
		OperationID: "migration-" + name + "-status",
		Summary:     "Get the migration status for " + name,
		Description: "Returns the migration status of the authenticated user for this service, i.e. whether and when they last migrated.",
		Method:      http.MethodGet,
		Path:        "/migration/" + name + "/status",
		Tags:        tags,
	}, func(ctx context.Context, _ *struct{}) (*migrationStatusBody, error) {
		return migrationFileStatus(ctx, factory)
	})

	Register(api, huma.Operation{
		OperationID: "migration-" + name + "-migrate",
		Summary:     "Migrate from " + name,
		Description: "Imports the authenticated user's data from an uploaded export file into Vikunja. Send the file under the multipart \"import\" field. The import runs synchronously and returns once it has finished.",
		Method:      http.MethodPost,
		Path:        "/migration/" + name + "/migrate",
		// POST runs an import rather than creating a REST resource, so it
		// returns 200 with a confirmation, not the wrapper's 201.
		DefaultStatus: http.StatusOK,
		Tags:          tags,
		// +2 MB mirrors Echo's global BodyLimit overhead so a max-sized file isn't rejected by multipart boundary/header bytes.
		// #nosec G115 - configured value won't exceed int64 max in practice.
		MaxBodyBytes: (int64(config.GetMaxFileSizeInMBytes()) + 2) * 1024 * 1024,
	}, func(ctx context.Context, in *fileMigrateInput) (*migrationStartedBody, error) {
		return migrationFileMigrate(ctx, factory, in)
	})
}

func migrationFileStatus(ctx context.Context, factory func() migration.FileMigrator) (*migrationStatusBody, error) {
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

func migrationFileMigrate(ctx context.Context, factory func() migration.FileMigrator, in *fileMigrateInput) (*migrationStartedBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	u, err := user.GetFromAuth(a)
	if err != nil {
		return nil, translateDomainError(err)
	}

	src := in.RawBody.Data().Import
	defer func() { _ = src.Close() }()

	if err := migrationHandler.RunFileMigration(factory(), u, src, src.Size); err != nil {
		return nil, translateDomainError(err)
	}

	out := &migrationStartedBody{}
	out.Body.Message = "Everything was migrated successfully."
	return out, nil
}
