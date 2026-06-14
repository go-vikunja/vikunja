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
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/routes"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// multipartFileBody builds a multipart body with a single file part under the
// given field name. CreateFormFile sets the part Content-Type to
// application/octet-stream, mirroring how many programmatic clients upload.
func multipartFileBody(t *testing.T, fieldName, filename string, content []byte) (*bytes.Buffer, string) {
	t.Helper()
	buf := &bytes.Buffer{}
	w := multipart.NewWriter(buf)
	fw, err := w.CreateFormFile(fieldName, filename)
	require.NoError(t, err)
	_, err = fw.Write(content)
	require.NoError(t, err)
	require.NoError(t, w.Close())
	return buf, w.FormDataContentType()
}

func uploadBackgroundRequest(t *testing.T, e *echo.Echo, project, token string, body *bytes.Buffer, contentType string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPut, "/api/v2/projects/"+project+"/backgrounds/upload", body)
	req.Header.Set("Content-Type", contentType)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func TestHumaProjectBackgroundUpload(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)

	t.Run("Owner uploads a background", func(t *testing.T) {
		// testuser1 owns project 1, which starts without a background.
		body, contentType := multipartFileBody(t, "background", "bg.png", pngBytes(t))
		rec := uploadBackgroundRequest(t, e, "1", humaTokenFor(t, &testuser1), body, contentType)
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())

		s := db.NewSession()
		defer s.Close()
		project := models.Project{ID: 1}
		has, err := s.Get(&project)
		require.NoError(t, err)
		require.True(t, has)
		assert.NotZero(t, project.BackgroundFileID, "the upload must set a background file id")
		assert.NotEmpty(t, project.BackgroundBlurHash, "the upload must compute a blur hash")
	})

	t.Run("Non-image rejected with 400", func(t *testing.T) {
		body, contentType := multipartFileBody(t, "background", "not-an-image.txt", []byte("this is plain text, not an image"))
		rec := uploadBackgroundRequest(t, e, "1", humaTokenFor(t, &testuser1), body, contentType)
		require.Equal(t, http.StatusBadRequest, rec.Code, "body: %s", rec.Body.String())
	})

	t.Run("Read-only user is forbidden", func(t *testing.T) {
		// testuser15 has read-only access to project 35.
		body, contentType := multipartFileBody(t, "background", "bg.png", pngBytes(t))
		rec := uploadBackgroundRequest(t, e, "35", humaTokenFor(t, &testuser15), body, contentType)
		assert.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
	})

	t.Run("No access at all is forbidden", func(t *testing.T) {
		// testuser1 has no access to project 35.
		body, contentType := multipartFileBody(t, "background", "bg.png", pngBytes(t))
		rec := uploadBackgroundRequest(t, e, "35", humaTokenFor(t, &testuser1), body, contentType)
		assert.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
	})

	t.Run("Unauthenticated", func(t *testing.T) {
		body, contentType := multipartFileBody(t, "background", "bg.png", pngBytes(t))
		rec := uploadBackgroundRequest(t, e, "1", "", body, contentType)
		require.Equal(t, http.StatusUnauthorized, rec.Code, "body: %s", rec.Body.String())
	})

	t.Run("Renders as multipart in the OpenAPI spec", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v2/openapi.json", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)

		var spec map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &spec))

		paths, _ := spec["paths"].(map[string]any)
		op, _ := paths["/projects/{project}/backgrounds/upload"].(map[string]any)
		put, ok := op["put"].(map[string]any)
		require.True(t, ok, "PUT /projects/{project}/backgrounds/upload must be in the spec")
		content, _ := put["requestBody"].(map[string]any)
		contentMap, _ := content["content"].(map[string]any)
		mp, ok := contentMap["multipart/form-data"].(map[string]any)
		require.True(t, ok, "background upload must be modeled as multipart/form-data")
		schema, _ := mp["schema"].(map[string]any)
		props, _ := schema["properties"].(map[string]any)
		bgProp, ok := props["background"].(map[string]any)
		require.True(t, ok, "the background field must appear in the multipart schema")
		assert.Equal(t, "binary", bgProp["format"], "background field must be a binary file in the spec")
	})
}

// TestHumaProjectBackgroundUploadDisabledByConfig verifies the upload route is
// absent (404) when the upload provider is disabled, even though backgrounds
// themselves are enabled.
func TestHumaProjectBackgroundUploadDisabledByConfig(t *testing.T) {
	_, err := setupTestEnv()
	require.NoError(t, err)

	config.BackgroundsUploadEnabled.Set(false)
	defer config.BackgroundsUploadEnabled.Set(true)

	e := routes.NewEcho()
	routes.RegisterRoutes(e)

	body, contentType := multipartFileBody(t, "background", "bg.png", pngBytes(t))
	rec := uploadBackgroundRequest(t, e, "1", humaTokenFor(t, &testuser1), body, contentType)
	assert.Equal(t, http.StatusNotFound, rec.Code, "route must be absent when background upload is disabled; body: %s", rec.Body.String())
}
