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
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/routes/api/shared"

	"github.com/danielgtaylor/huma/v2"
)

// testingReplaceInput is the request for resetting a single table. The
// Authorization header carries the configured testing token (not a JWT or API
// token); the endpoint is public and checks it in-handler like v1.
type testingReplaceInput struct {
	Table string `path:"table" doc:"The table to reset."`
	// String (not bool) so absent is distinguishable from an explicit "false":
	// like v1, an absent truncate parameter means truncate. Huma does not
	// support *bool params, and a bool with default:"true" silently ignores an
	// explicit ?truncate=false, so the parameter is read as a raw string and
	// interpreted in the handler exactly like v1 does.
	Truncate      string           `query:"truncate" enum:"true,false" doc:"Empty the table (and its dependents) before inserting the rows. Defaults to true; pass false to restore on top of existing data."`
	Authorization string           `header:"Authorization" doc:"The configured testing token."`
	Body          []map[string]any `doc:"The rows to write into the table. Free-form objects matching the table's columns."`
}

type testingReplaceBody struct {
	Body []map[string]any `doc:"The table's contents after the reset."`
}

type testingTruncateAllInput struct {
	Authorization string `header:"Authorization" doc:"The configured testing token."`
}

type testingTruncateAllBody struct {
	Body struct {
		Message string `json:"message" doc:"Always \"ok\" on success."`
	}
}

// RegisterTestingRoutes wires the e2e testing-support endpoints onto the Huma
// API. They are only mounted when the testing token is configured, matching v1.
func RegisterTestingRoutes(api huma.API) {
	if config.ServiceTestingtoken.GetString() == "" {
		return
	}

	tags := []string{"testing"}
	// Public: opt out of the globally-applied JWT/API-token auth — these
	// authenticate with the testing token via the Authorization header
	// instead. Their paths are also listed in unauthenticatedAPIPaths so the
	// token middleware lets them through.
	noAuth := []map[string][]string{}

	Register(api, huma.Operation{
		OperationID: "testing-truncate-all",
		Summary:     "Truncate all tables",
		Description: "Removes all data from every Vikunja table. Used by e2e tests to ensure a clean state before each test. Authenticates with the configured testing token via the Authorization header, not a JWT or API token.",
		Method:      http.MethodDelete,
		Path:        "/test/all",
		Tags:        tags,
		Security:    noAuth,
		// v1 returns 200 with a body rather than the 204 a DELETE would default to.
		DefaultStatus: http.StatusOK,
	}, testingTruncateAll)

	Register(api, huma.Operation{
		OperationID: "testing-replace-table",
		Summary:     "Reset a table to a defined state",
		Description: "Replaces the contents of the named table with the rows in the payload and returns the resulting contents. Used by e2e tests to seed fixtures. Authenticates with the configured testing token via the Authorization header, not a JWT or API token.",
		Method:      http.MethodPut,
		Path:        "/test/{table}",
		Tags:        tags,
		Security:    noAuth,
		// Mirror v1's 201 for a successful reset.
		DefaultStatus: http.StatusCreated,
	}, testingReplaceTable)
}

func init() { AddRouteRegistrar(RegisterTestingRoutes) }

func testingReplaceTable(_ context.Context, in *testingReplaceInput) (*testingReplaceBody, error) {
	if in.Authorization != config.ServiceTestingtoken.GetString() {
		return nil, huma.Error403Forbidden("forbidden")
	}

	// Mirror v1: absent or "true" truncates; only an explicit "false" appends.
	truncate := in.Truncate == "true" || in.Truncate == ""
	data, err := shared.ReplaceTableContents(in.Table, in.Body, truncate)
	if err != nil {
		log.Errorf("Error replacing table data: %v", err)
		return nil, huma.Error500InternalServerError("could not replace table data")
	}

	return &testingReplaceBody{Body: data}, nil
}

func testingTruncateAll(_ context.Context, in *testingTruncateAllInput) (*testingTruncateAllBody, error) {
	if in.Authorization != config.ServiceTestingtoken.GetString() {
		return nil, huma.Error403Forbidden("forbidden")
	}

	if err := shared.TruncateAllTestingTables(); err != nil {
		log.Errorf("Error truncating all tables: %v", err)
		return nil, huma.Error500InternalServerError("could not truncate tables")
	}

	out := &testingTruncateAllBody{}
	out.Body.Message = "ok"
	return out, nil
}
