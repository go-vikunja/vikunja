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
	"os"
	"strconv"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/modules/migration"
	migrationHandler "code.vikunja.io/api/pkg/modules/migration/handler"
	"code.vikunja.io/api/pkg/notifications"

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

// runQueuedImport runs the job for the import the preceding request queued:
// webtests dispatch into events.Fake(), so no listener runs on its own.
func runQueuedImport(t *testing.T) *migration.Status {
	t.Helper()
	dispatched := events.GetDispatchedEvents((&migrationHandler.FileMigrationRequestedEvent{}).Name())
	require.Len(t, dispatched, 1)
	event, ok := dispatched[0].(*migrationHandler.FileMigrationRequestedEvent)
	require.True(t, ok)

	events.TestListener(t, event, &migrationHandler.FileMigrationListener{})

	status, err := migration.GetMigrationStatusByID(event.MigrationStatusID)
	require.NoError(t, err)
	return status
}

func assertClaimReleased(t *testing.T, status *migration.Status) {
	t.Helper()
	assert.False(t, status.FinishedAt.IsZero(), "the finished import must not keep holding the claim")
	assert.Nil(t, status.ActiveUserID, "the finished import must not keep holding the claim")
}

// TestHumaMigrationFile covers the always-registered file migrators
// (vikunja-file, ticktick, wekan) status + migrate endpoints. There is no v1
// webtest for these handlers to mirror, so this is the parity baseline.
func TestHumaMigrationFile(t *testing.T) {
	migrators := []string{"vikunja-file", "ticktick", "wekan"}

	t.Run("status - never migrated", func(t *testing.T) {
		e := setupMigrationTestEnv(t)
		token := humaTokenFor(t, &testuser1)

		for _, name := range migrators {
			rec := humaRequest(t, e, http.MethodGet, "/api/v2/migration/"+name+"/status", "", token, "")
			require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
			assert.Contains(t, rec.Body.String(), `"started_at":"0001-01-01T00:00:00Z"`, "body: %s", rec.Body.String())
		}
	})

	// vikunja-file is the only migrator still validating in the request: reading
	// the zip central directory is cheap, while ticktick and wekan would have to
	// parse the whole upload to say anything about it.
	t.Run("vikunja-file rejects a non-zip upload", func(t *testing.T) {
		e := setupMigrationTestEnv(t)
		token := humaTokenFor(t, &testuser1)

		body, contentType := multipartImportBody(t, "bad.zip", []byte("not a zip archive"), nil)
		rec := migrationUploadRequest(t, e, "/api/v2/migration/vikunja-file/migrate", body, contentType, token)
		require.Equal(t, http.StatusBadRequest, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), strconv.Itoa(migration.ErrCodeNotAZipFile), "body: %s", rec.Body.String())
	})

	// payload is shaped to hit a *domain* rejection once parsed: a ticktick CSV
	// with no data rows is "empty", as is a wekan board with no title or cards.
	for name, payload := range map[string][]byte{
		"ticktick": []byte("Title,Content\n"),
		"wekan":    []byte(`{"title":"","cards":[]}`),
	} {
		t.Run(name+" queues an unusable upload and fails it in the job", func(t *testing.T) {
			e := setupMigrationTestEnv(t)
			token := humaTokenFor(t, &testuser1)
			notifications.Fake()
			t.Cleanup(notifications.Unfake)
			events.ClearDispatchedEvents()

			body, contentType := multipartImportBody(t, "bad."+name, payload, nil)
			rec := migrationUploadRequest(t, e, "/api/v2/migration/"+name+"/migrate", body, contentType, token)
			require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
			assert.Contains(t, rec.Body.String(), `"message":"Migration was started successfully."`)

			status := runQueuedImport(t)
			notifications.AssertSent(t, &migrationHandler.MigrationFailedNotification{})
			notifications.AssertNotSent(t, &migrationHandler.MigrationDoneNotification{})
			assertClaimReleased(t, status)
		})
	}
}

// TestHumaMigrationFile_QueuesTheImport proves a valid export is accepted and
// queued rather than imported inside the request, and that the claim it takes
// blocks a second import until the queued one finishes.
func TestHumaMigrationFile_QueuesTheImport(t *testing.T) {
	e := setupMigrationTestEnv(t)
	token := humaTokenFor(t, &testuser1)

	export, err := os.ReadFile("../modules/migration/vikunja-file/export.zip")
	require.NoError(t, err)

	body, contentType := multipartImportBody(t, "export.zip", export, nil)
	rec := migrationUploadRequest(t, e, "/api/v2/migration/vikunja-file/migrate", body, contentType, token)
	require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
	assert.Contains(t, rec.Body.String(), `"message":"Migration was started successfully."`)
	events.AssertDispatched(t, &migrationHandler.FileMigrationRequestedEvent{})

	body, contentType = multipartImportBody(t, "export.zip", export, nil)
	rec = migrationUploadRequest(t, e, "/api/v2/migration/vikunja-file/migrate", body, contentType, token)
	assert.Equal(t, http.StatusPreconditionFailed, rec.Code,
		"the queued import must keep holding the claim; body: %s", rec.Body.String())
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

// TestHumaMigrationFile_MalformedJSON proves a syntactically broken upload is
// classified as the user's problem, not a 500 forwarded to Sentry. See
// API-CLOUD-4B. wekan no longer validates in the request, so the classification
// now happens when the job maps the parse failure through asImportFileError.
func TestHumaMigrationFile_MalformedJSON(t *testing.T) {
	config.SentryEnabled.Set(true)
	defer config.SentryEnabled.Set(false)

	t.Cleanup(notifications.Unfake)

	for _, payload := range [][]byte{
		[]byte(`<html>not an export</html>`),
		[]byte(`{"title": `),
		append([]byte{0xEF, 0xBB, 0xBF}, []byte(`{"title": `)...),
	} {
		e := setupMigrationTestEnv(t)
		token := humaTokenFor(t, &testuser1)
		notifications.Fake()
		events.ClearDispatchedEvents()

		body, contentType := multipartImportBody(t, "board.json", payload, nil)
		rec := migrationUploadRequest(t, e, "/api/v2/migration/wekan/migrate", body, contentType, token)
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())

		status := runQueuedImport(t)
		notifications.AssertSent(t, &migrationHandler.MigrationFailedNotification{})
		notifications.AssertNotSent(t, &migrationHandler.MigrationFailedReportedNotification{})
		assertClaimReleased(t, status)
	}
}

// TestHumaMigrationFile_JSONWithBOM proves an otherwise valid export is not
// rejected just because the exporter prefixed it with a byte order mark.
func TestHumaMigrationFile_JSONWithBOM(t *testing.T) {
	e := setupMigrationTestEnv(t)
	token := humaTokenFor(t, &testuser1)
	notifications.Fake()
	t.Cleanup(notifications.Unfake)
	events.ClearDispatchedEvents()

	payload := append([]byte{0xEF, 0xBB, 0xBF}, []byte(`{"title":"BOM board","lists":[{"_id":"l1","title":"Todo"}],"cards":[{"_id":"c1","title":"A card","listId":"l1"}]}`)...)
	body, contentType := multipartImportBody(t, "board.json", payload, nil)
	rec := migrationUploadRequest(t, e, "/api/v2/migration/wekan/migrate", body, contentType, token)
	require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())

	status := runQueuedImport(t)
	notifications.AssertSent(t, &migrationHandler.MigrationDoneNotification{})
	assertClaimReleased(t, status)
	db.AssertExists(t, "projects", map[string]interface{}{
		"title":    "BOM board",
		"owner_id": testuser1.ID,
	}, false)
}
