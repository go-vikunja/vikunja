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

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/modules/migration"
	"code.vikunja.io/api/pkg/modules/migration/csv"
	"code.vikunja.io/api/pkg/user"

	"github.com/danielgtaylor/huma/v2"
)

// csvDetectInput is the detect upload: just the file.
type csvDetectInput struct {
	RawBody huma.MultipartFormFiles[struct {
		Import huma.FormFile `form:"import" required:"true" doc:"The CSV file to analyze."`
	}]
}

// csvImportInput is the preview/migrate upload: the file plus a JSON config
// blob carried as a multipart form value (mirrors v1's FormValue(\"config\")).
type csvImportInput struct {
	RawBody huma.MultipartFormFiles[struct {
		Import huma.FormFile `form:"import" required:"true" doc:"The CSV file to import."`
		Config string        `form:"config" required:"true" doc:"The import configuration as a JSON object (see the ImportConfig schema), passed as a multipart form value. Obtain a starting config from the detect endpoint."`
	}]
}

type csvDetectBody struct {
	Body *csv.DetectionResult
}

type csvPreviewBody struct {
	Body *csv.PreviewResult
}

// RegisterMigrationCSVRoutes wires the generic CSV importer onto the Huma API.
// Like the other file migrators it has no config flag in v1, so it is always
// registered.
func RegisterMigrationCSVRoutes(api huma.API) {
	tags := []string{"migration"}
	// +2 MB mirrors Echo's global BodyLimit overhead so a max-sized file isn't rejected by multipart boundary/header bytes.
	// #nosec G115 - configured value won't exceed int64 max in practice.
	maxBody := (int64(config.GetMaxFileSizeInMBytes()) + 2) * 1024 * 1024

	Register(api, huma.Operation{
		OperationID: "migration-csv-status",
		Summary:     "Get the CSV migration status",
		Description: "Returns the migration status of the authenticated user for the CSV importer, i.e. whether and when they last imported a CSV.",
		Method:      http.MethodGet,
		Path:        "/migration/csv/status",
		Tags:        tags,
	}, csvStatus)

	Register(api, huma.Operation{
		OperationID:   "migration-csv-detect",
		Summary:       "Detect a CSV file's structure",
		Description:   "Analyzes an uploaded CSV file and returns its detected columns, delimiter, quote character and date format, plus a suggested column-to-attribute mapping the client can edit before previewing or migrating. Read-only: nothing is imported.",
		Method:        http.MethodPost,
		Path:          "/migration/csv/detect",
		DefaultStatus: http.StatusOK,
		Tags:          tags,
		MaxBodyBytes:  maxBody,
	}, csvDetect)

	Register(api, huma.Operation{
		OperationID:   "migration-csv-preview",
		Summary:       "Preview a CSV import",
		Description:   "Returns the first few tasks that would be imported from the uploaded CSV file with the given config, without importing anything. Read-only.",
		Method:        http.MethodPost,
		Path:          "/migration/csv/preview",
		DefaultStatus: http.StatusOK,
		Tags:          tags,
		MaxBodyBytes:  maxBody,
	}, csvPreview)

	Register(api, huma.Operation{
		OperationID: "migration-csv-migrate",
		Summary:     "Import a CSV file",
		Description: "Imports the tasks from the uploaded CSV file into Vikunja using the given config. The import runs synchronously and returns once it has finished.",
		Method:      http.MethodPost,
		Path:        "/migration/csv/migrate",
		// POST runs an import rather than creating a REST resource, so it
		// returns 200 with a confirmation, not the wrapper's 201.
		DefaultStatus: http.StatusOK,
		Tags:          tags,
		MaxBodyBytes:  maxBody,
	}, csvMigrate)
}

func init() { AddRouteRegistrar(RegisterMigrationCSVRoutes) }

func csvStatus(ctx context.Context, _ *struct{}) (*migrationStatusBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	u, err := user.GetFromAuth(a)
	if err != nil {
		return nil, translateDomainError(err)
	}

	status, err := migration.GetMigrationStatus(&csv.Migrator{}, u)
	if err != nil {
		return nil, translateDomainError(err)
	}
	return &migrationStatusBody{Body: status}, nil
}

func csvDetect(ctx context.Context, in *csvDetectInput) (*csvDetectBody, error) {
	if _, err := authFromCtx(ctx); err != nil {
		return nil, err
	}

	src := in.RawBody.Data().Import
	defer func() { _ = src.Close() }()

	result, err := csv.DetectCSVStructure(src, src.Size)
	if err != nil {
		return nil, translateDomainError(err)
	}
	return &csvDetectBody{Body: result}, nil
}

func csvPreview(ctx context.Context, in *csvImportInput) (*csvPreviewBody, error) {
	if _, err := authFromCtx(ctx); err != nil {
		return nil, err
	}

	cfg, err := parseCSVImportConfig(in.RawBody.Data().Config)
	if err != nil {
		return nil, err
	}

	src := in.RawBody.Data().Import
	defer func() { _ = src.Close() }()

	result, err := csv.PreviewImport(src, src.Size, cfg)
	if err != nil {
		return nil, translateDomainError(err)
	}
	return &csvPreviewBody{Body: result}, nil
}

func csvMigrate(ctx context.Context, in *csvImportInput) (*migrationStartedBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	u, err := user.GetFromAuth(a)
	if err != nil {
		return nil, translateDomainError(err)
	}

	cfg, err := parseCSVImportConfig(in.RawBody.Data().Config)
	if err != nil {
		return nil, err
	}

	src := in.RawBody.Data().Import
	defer func() { _ = src.Close() }()

	if err := csv.RunMigration(u, src, src.Size, cfg); err != nil {
		return nil, translateDomainError(err)
	}

	out := &migrationStartedBody{}
	out.Body.Message = "Everything was migrated successfully."
	return out, nil
}

// parseCSVImportConfig unmarshals the JSON config form value, mirroring v1's
// json.Unmarshal of FormValue("config"). required:"true" guarantees presence,
// so only a malformed body needs guarding here.
func parseCSVImportConfig(raw string) (*csv.ImportConfig, error) {
	var cfg csv.ImportConfig
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return nil, huma.Error400BadRequest("Invalid configuration: " + err.Error())
	}
	return &cfg, nil
}
