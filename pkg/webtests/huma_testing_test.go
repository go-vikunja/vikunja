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

package webtests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/license"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/keyvalue"
	"code.vikunja.io/api/pkg/modules/migration"
	"code.vikunja.io/api/pkg/routes"
	"code.vikunja.io/api/pkg/user"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"src.techknowlogick.com/xormigrate"
)

const testingToken = "test-testing-token"

// setupTestingEnv mirrors setupTestEnv but sets the testing token before
// registering routes, so the config-gated /api/v2/test/* endpoints mount.
// When token is empty the endpoints stay unmounted (the disabled case).
func setupTestingEnv(t *testing.T, token string) *echo.Echo {
	t.Helper()
	config.InitDefaultConfig()
	config.ServicePublicURL.Set("https://localhost")
	config.ServiceTestingtoken.Set(token)
	t.Cleanup(func() { config.ServiceTestingtoken.Set("") })

	log.InitLogger()
	files.InitTests()
	user.InitTests()
	models.SetupTests()
	events.Fake()
	keyvalue.InitStorage()

	// models.SetupTests only syncs models + notifications tables, but
	// TruncateAllTables walks *every* registered table — including ones created
	// by migration in production (license_status, migration_status) plus
	// xormigrate's "migration" tracking table. Create them here so truncate-all
	// doesn't hit "no such table" (the same gap that kept v1 from testing it).
	engine, err := db.CreateTestEngine()
	require.NoError(t, err)
	extraTables := append(append([]any{new(xormigrate.Migration)}, license.GetTables()...), migration.GetTables()...)
	require.NoError(t, engine.Sync2(extraTables...))

	require.NoError(t, db.LoadFixtures())

	e := routes.NewEcho()
	routes.RegisterRoutes(e)
	return e
}

// testingRequest dispatches a request to a /api/v2/test/* endpoint, sending the
// raw token in the Authorization header (not a Bearer JWT).
func testingRequest(e *echo.Echo, method, path, body, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", token)
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func countRows(t *testing.T, table string) int {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	rows := []map[string]interface{}{}
	require.NoError(t, s.Table(table).Find(&rows))
	return len(rows)
}

func TestTesting(t *testing.T) {
	t.Run("replace table contents", func(t *testing.T) {
		e := setupTestingEnv(t, testingToken)
		t.Cleanup(func() { _ = db.LoadFixtures() })

		body := `[{"id":1,"title":"only label","created_by_id":1,"created":"2020-01-01T00:00:00Z","updated":"2020-01-01T00:00:00Z"}]`
		rec := testingRequest(e, http.MethodPut, "/api/v2/test/labels", body, testingToken)
		require.Equal(t, http.StatusCreated, rec.Code, "body: %s", rec.Body.String())

		var data []map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &data))
		require.Len(t, data, 1)
		assert.EqualValues(t, "only label", data[0]["title"])
		assert.Equal(t, 1, countRows(t, "labels"), "table should hold exactly the seeded rows")
	})

	t.Run("replace without truncate keeps existing rows", func(t *testing.T) {
		e := setupTestingEnv(t, testingToken)
		t.Cleanup(func() { _ = db.LoadFixtures() })

		before := countRows(t, "labels")
		require.Positive(t, before, "fixtures should seed some labels")

		body := `[{"id":9999,"title":"added label","created_by_id":1,"created":"2020-01-01T00:00:00Z","updated":"2020-01-01T00:00:00Z"}]`
		rec := testingRequest(e, http.MethodPut, "/api/v2/test/labels?truncate=false", body, testingToken)
		require.Equal(t, http.StatusCreated, rec.Code, "body: %s", rec.Body.String())

		assert.Equal(t, before+1, countRows(t, "labels"), "row should be added on top of existing data")
	})

	t.Run("truncate all tables", func(t *testing.T) {
		e := setupTestingEnv(t, testingToken)
		t.Cleanup(func() { _ = db.LoadFixtures() })

		require.Positive(t, countRows(t, "labels"))

		rec := testingRequest(e, http.MethodDelete, "/api/v2/test/all", "", testingToken)
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())

		var resp struct {
			Message string `json:"message"`
		}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, "ok", resp.Message)
		assert.Equal(t, 0, countRows(t, "labels"), "every table should be empty after truncate")
	})

	t.Run("wrong token is forbidden", func(t *testing.T) {
		e := setupTestingEnv(t, testingToken)

		rec := testingRequest(e, http.MethodPut, "/api/v2/test/labels", `[]`, "wrong-token")
		assert.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())

		rec = testingRequest(e, http.MethodDelete, "/api/v2/test/all", "", "wrong-token")
		assert.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
	})

	t.Run("missing token is forbidden", func(t *testing.T) {
		e := setupTestingEnv(t, testingToken)

		rec := testingRequest(e, http.MethodPut, "/api/v2/test/labels", `[]`, "")
		assert.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())

		rec = testingRequest(e, http.MethodDelete, "/api/v2/test/all", "", "")
		assert.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
	})
}

func TestTesting_DisabledConfig(t *testing.T) {
	e := setupTestingEnv(t, "")

	rec := testingRequest(e, http.MethodPut, "/api/v2/test/labels", `[]`, "")
	assert.Equal(t, http.StatusNotFound, rec.Code, "endpoint must be absent when no testing token is configured")

	rec = testingRequest(e, http.MethodDelete, "/api/v2/test/all", "", "")
	assert.Equal(t, http.StatusNotFound, rec.Code, "endpoint must be absent when no testing token is configured")
}

func TestTesting_BodySchemaIsArrayOfObjects(t *testing.T) {
	e := setupTestingEnv(t, testingToken)

	req := httptest.NewRequest(http.MethodGet, "/api/v2/openapi.json", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var spec map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &spec))

	paths, _ := spec["paths"].(map[string]any)
	op, _ := paths["/test/{table}"].(map[string]any)
	put, ok := op["put"].(map[string]any)
	require.True(t, ok, "PUT /test/{table} must be in the spec")

	reqBody, _ := put["requestBody"].(map[string]any)
	content, _ := reqBody["content"].(map[string]any)
	appJSON, _ := content["application/json"].(map[string]any)
	schema, _ := appJSON["schema"].(map[string]any)
	// FieldsOptionalByDefault makes the array nullable, so `type` may be the
	// string "array" or the list ["array","null"]. Either is honest; assert it
	// describes an array (not, say, a base64 string as json.RawMessage would).
	assert.Contains(t, schemaTypes(schema["type"]), "array", "request body must be modeled as an array")
}

// schemaTypes normalises an OpenAPI `type` value (a string or a list of
// strings when nullable) into a slice for assertion.
func schemaTypes(v any) []string {
	switch t := v.(type) {
	case string:
		return []string{t}
	case []any:
		out := make([]string, 0, len(t))
		for _, e := range t {
			if s, ok := e.(string); ok {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}
