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
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const csvTestFile = `Title,Description,Done,Priority
Task 1,Description 1,true,high
Task 2,Description 2,false,low`

const csvTestConfig = `{"delimiter":",","quote_char":"\"","date_format":"2006-01-02","mapping":[` +
	`{"column_index":0,"column_name":"Title","attribute":"title"},` +
	`{"column_index":1,"column_name":"Description","attribute":"description"},` +
	`{"column_index":2,"column_name":"Done","attribute":"done"},` +
	`{"column_index":3,"column_name":"Priority","attribute":"priority"}]}`

// TestHumaMigrationCSV covers the generic CSV importer's v2 endpoints:
// status, detect, preview and migrate. No v1 webtest exists to mirror.
func TestHumaMigrationCSV(t *testing.T) {
	e := setupMigrationTestEnv(t)
	token := humaTokenFor(t, &testuser1)

	t.Run("status - never migrated", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/migration/csv/status", "", token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"started_at":"0001-01-01T00:00:00Z"`, "body: %s", rec.Body.String())
	})

	t.Run("detect returns columns and a suggested mapping", func(t *testing.T) {
		body, contentType := multipartImportBody(t, "import.csv", []byte(csvTestFile), nil)
		rec := migrationUploadRequest(t, e, "/api/v2/migration/csv/detect", body, contentType, token)
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"columns"`)
		assert.Contains(t, rec.Body.String(), `"suggested_mapping"`)
		assert.Contains(t, rec.Body.String(), "Title")
	})

	t.Run("preview returns tasks without importing", func(t *testing.T) {
		body, contentType := multipartImportBody(t, "import.csv", []byte(csvTestFile), map[string]string{"config": csvTestConfig})
		rec := migrationUploadRequest(t, e, "/api/v2/migration/csv/preview", body, contentType, token)
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"tasks"`)
		assert.Contains(t, rec.Body.String(), "Task 1")
	})

	t.Run("migrate imports the file", func(t *testing.T) {
		body, contentType := multipartImportBody(t, "import.csv", []byte(csvTestFile), map[string]string{"config": csvTestConfig})
		rec := migrationUploadRequest(t, e, "/api/v2/migration/csv/migrate", body, contentType, token)
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"message":"Everything was migrated successfully."`)

		// The status now reflects a finished migration.
		rec = humaRequest(t, e, http.MethodGet, "/api/v2/migration/csv/status", "", token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.NotContains(t, rec.Body.String(), `"started_at":"0001-01-01T00:00:00Z"`,
			"after migrating, the status must carry a real started_at; body: %s", rec.Body.String())
	})
}

// TestHumaMigrationCSV_BadInput covers the negative paths: missing config,
// malformed config JSON, and an empty file.
func TestHumaMigrationCSV_BadInput(t *testing.T) {
	e := setupMigrationTestEnv(t)
	token := humaTokenFor(t, &testuser1)

	t.Run("missing config is rejected with 422", func(t *testing.T) {
		// The config form value is required:"true", so Huma's multipart
		// validation refuses the request before the handler runs.
		body, contentType := multipartImportBody(t, "import.csv", []byte(csvTestFile), nil)
		rec := migrationUploadRequest(t, e, "/api/v2/migration/csv/migrate", body, contentType, token)
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code, "body: %s", rec.Body.String())
	})

	t.Run("malformed config JSON is rejected with 400", func(t *testing.T) {
		body, contentType := multipartImportBody(t, "import.csv", []byte(csvTestFile), map[string]string{"config": "{not json"})
		rec := migrationUploadRequest(t, e, "/api/v2/migration/csv/migrate", body, contentType, token)
		assert.Equal(t, http.StatusBadRequest, rec.Code, "body: %s", rec.Body.String())
	})

	t.Run("empty file is rejected with a domain error", func(t *testing.T) {
		body, contentType := multipartImportBody(t, "empty.csv", []byte{}, map[string]string{"config": csvTestConfig})
		rec := migrationUploadRequest(t, e, "/api/v2/migration/csv/migrate", body, contentType, token)
		assert.Equal(t, http.StatusBadRequest, rec.Code, "body: %s", rec.Body.String())
	})
}

// TestHumaMigrationCSV_Unauthenticated proves all CSV ops require auth.
func TestHumaMigrationCSV_Unauthenticated(t *testing.T) {
	e := setupMigrationTestEnv(t)

	t.Run("status", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/migration/csv/status", "", "", "")
		assert.Equal(t, http.StatusUnauthorized, rec.Code, "body: %s", rec.Body.String())
	})
	t.Run("detect", func(t *testing.T) {
		body, contentType := multipartImportBody(t, "import.csv", []byte(csvTestFile), nil)
		rec := migrationUploadRequest(t, e, "/api/v2/migration/csv/detect", body, contentType, "")
		assert.Equal(t, http.StatusUnauthorized, rec.Code, "body: %s", rec.Body.String())
	})
	t.Run("migrate", func(t *testing.T) {
		body, contentType := multipartImportBody(t, "import.csv", []byte(csvTestFile), map[string]string{"config": csvTestConfig})
		rec := migrationUploadRequest(t, e, "/api/v2/migration/csv/migrate", body, contentType, "")
		assert.Equal(t, http.StatusUnauthorized, rec.Code, "body: %s", rec.Body.String())
	})
}
