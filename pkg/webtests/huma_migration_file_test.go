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
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// multipartImportBody builds a multipart/form-data body with the file under the
// "import" field plus any extra string form values (e.g. the CSV "config"),
// matching the v2 file/CSV migrator form schemas.
func multipartImportBody(t *testing.T, filename string, content []byte, values map[string]string) (*bytes.Buffer, string) {
	t.Helper()
	buf := &bytes.Buffer{}
	w := multipart.NewWriter(buf)
	fw, err := w.CreateFormFile("import", filename)
	require.NoError(t, err)
	_, err = fw.Write(content)
	require.NoError(t, err)
	for k, v := range values {
		require.NoError(t, w.WriteField(k, v))
	}
	require.NoError(t, w.Close())
	return buf, w.FormDataContentType()
}

func migrationUploadRequest(t *testing.T, e *echo.Echo, path string, body *bytes.Buffer, contentType, token string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, body)
	req.Header.Set("Content-Type", contentType)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

// TestHumaMigrationFile covers the always-registered file migrators
// (vikunja-file, ticktick, wekan) status + migrate endpoints. There is no v1
// webtest for these handlers to mirror, so this is the parity baseline.
func TestHumaMigrationFile(t *testing.T) {
	e := setupMigrationTestEnv(t)
	token := humaTokenFor(t, &testuser1)

	// payload is shaped per migrator to hit a *domain* rejection (4xx) rather
	// than a raw parse error: a wekan board with no title/cards is "empty", a
	// ticktick CSV with no data rows is "empty", and a vikunja-file that isn't
	// a zip is rejected as such. (Syntactically-malformed input would surface a
	// raw json/zip error that maps to 500 in both v1 and v2 alike.)
	migrators := map[string][]byte{
		"vikunja-file": []byte("not a zip archive"),
		"ticktick":     []byte("Title,Content\n"),
		"wekan":        []byte(`{"title":"","cards":[]}`),
	}

	for name, payload := range migrators {
		t.Run(name+" status - never migrated", func(t *testing.T) {
			rec := humaRequest(t, e, http.MethodGet, "/api/v2/migration/"+name+"/status", "", token, "")
			require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
			// A user who never migrated has a zero-value status.
			assert.Contains(t, rec.Body.String(), `"started_at":"0001-01-01T00:00:00Z"`, "body: %s", rec.Body.String())
		})

		t.Run(name+" migrate maps a rejected file to a 4xx domain error", func(t *testing.T) {
			// Drives the request through the multipart binding and into the
			// migrator, which rejects it with a domain error that
			// translateDomainError turns into a 4xx — proving the v2 plumbing
			// (bind, run, error bridge) is wired, not the parsing itself.
			body, contentType := multipartImportBody(t, "bad."+name, payload, nil)
			rec := migrationUploadRequest(t, e, "/api/v2/migration/"+name+"/migrate", body, contentType, token)
			assert.GreaterOrEqual(t, rec.Code, http.StatusBadRequest, "body: %s", rec.Body.String())
			assert.Less(t, rec.Code, http.StatusInternalServerError,
				"a rejected upload must map to a 4xx domain error, not a 500; body: %s", rec.Body.String())
		})
	}
}

// TestHumaMigrationFile_Unauthenticated proves the file migrator ops require auth.
func TestHumaMigrationFile_Unauthenticated(t *testing.T) {
	e := setupMigrationTestEnv(t)

	t.Run("status", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/migration/ticktick/status", "", "", "")
		assert.Equal(t, http.StatusUnauthorized, rec.Code, "body: %s", rec.Body.String())
	})
	t.Run("migrate", func(t *testing.T) {
		body, contentType := multipartImportBody(t, "x.csv", []byte("x"), nil)
		rec := migrationUploadRequest(t, e, "/api/v2/migration/ticktick/migrate", body, contentType, "")
		assert.Equal(t, http.StatusUnauthorized, rec.Code, "body: %s", rec.Body.String())
	})
}

// TestHumaMigrationFile_MissingFile proves the required "import" form field is
// enforced by Huma's multipart validation (422), not a 500.
func TestHumaMigrationFile_MissingFile(t *testing.T) {
	e := setupMigrationTestEnv(t)
	token := humaTokenFor(t, &testuser1)

	buf := &bytes.Buffer{}
	w := multipart.NewWriter(buf)
	require.NoError(t, w.Close())

	rec := migrationUploadRequest(t, e, "/api/v2/migration/ticktick/migrate", buf, w.FormDataContentType(), token)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code, "body: %s", rec.Body.String())
}
