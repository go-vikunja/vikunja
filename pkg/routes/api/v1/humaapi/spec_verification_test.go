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

package humaapi_test

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"code.vikunja.io/api/pkg/modules/humaecho5"
	"code.vikunja.io/api/pkg/routes/api/v1/humaapi"

	"github.com/danielgtaylor/huma/v2"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSpecVerification_LabelOAS31 is the Phase F1 (file-based) smoke check.
// It builds the same Huma API the production routes wire up, registers the
// Label resource, then validates the generated spec is well-formed OAS 3.1
// with the expected paths/methods/security. On failure it dumps the spec
// to /tmp/huma-label-spec.json for human inspection.
func TestSpecVerification_LabelOAS31(t *testing.T) {
	e := echo.New()
	cfg := huma.DefaultConfig("Vikunja API (OAS 3.1 spike)", "0.0.1")
	cfg.OpenAPIPath = "/openapi"
	cfg.FieldsOptionalByDefault = true
	api := humaecho5.New(e, cfg)
	humaapi.Install()
	humaapi.RegisterLabelRoutes(api)

	// Render to JSON and round-trip into a generic map so we can assert on
	// shape without coupling to Huma's struct types.
	req := httptest.NewRequest("GET", "/openapi.json", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	require.Equalf(t, 200, rec.Code, "openapi spec endpoint failed: %s", rec.Body.String())

	// Persist for human inspection / external diffing.
	specPath := filepath.Join(t.TempDir(), "huma-label-spec.json")
	require.NoError(t, os.WriteFile(specPath, rec.Body.Bytes(), 0o600))

	var spec map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &spec))

	// 1. OAS version
	openapiVersion, _ := spec["openapi"].(string)
	assert.Truef(t, len(openapiVersion) >= 3 && openapiVersion[:3] == "3.1",
		"expected OAS 3.1.x, got %q", openapiVersion)
	t.Logf("openapi version: %s", openapiVersion)

	// 2. Paths exist
	paths, _ := spec["paths"].(map[string]any)
	require.NotNil(t, paths, "spec has no paths object")
	require.Contains(t, paths, "/labels", "missing /labels path")
	require.Contains(t, paths, "/labels/{id}", "missing /labels/{id} path")

	// 3. /labels has GET (list) + PUT (create)
	basePath, _ := paths["/labels"].(map[string]any)
	assert.Contains(t, basePath, "get", "/labels missing GET (list)")
	assert.Contains(t, basePath, "put", "/labels missing PUT (create)")

	// 4. /labels/{id} has GET (read) + POST (update) + DELETE
	itemPath, _ := paths["/labels/{id}"].(map[string]any)
	assert.Contains(t, itemPath, "get", "/labels/{id} missing GET (read)")
	assert.Contains(t, itemPath, "post", "/labels/{id} missing POST (update)")
	assert.Contains(t, itemPath, "delete", "/labels/{id} missing DELETE")

	// 5. Operations carry security (JWT) and tags
	listOp, _ := basePath["get"].(map[string]any)
	assert.Contains(t, listOp, "security", "list op missing security")
	assert.Contains(t, listOp, "tags", "list op missing tags")

	// 6. The {id} parameter is declared
	itemReadOp, _ := itemPath["get"].(map[string]any)
	params, _ := itemReadOp["parameters"].([]any)
	require.NotEmpty(t, params, "read-one op has no parameters")
	foundID := false
	for _, p := range params {
		pm, _ := p.(map[string]any)
		if pm["name"] == "id" && pm["in"] == "path" {
			foundID = true
			break
		}
	}
	assert.True(t, foundID, "read-one op missing path parameter 'id'")

	// 7. List op exposes paging query params
	listParams, _ := listOp["parameters"].([]any)
	queryNames := map[string]bool{}
	for _, p := range listParams {
		pm, _ := p.(map[string]any)
		if pm["in"] == "query" {
			if name, ok := pm["name"].(string); ok {
				queryNames[name] = true
			}
		}
	}
	for _, want := range []string{"page", "per_page", "s"} {
		assert.Truef(t, queryNames[want], "list op missing query param %q (got %v)", want, queryNames)
	}

	t.Logf("spec written to %s (%d bytes)", specPath, rec.Body.Len())
}
