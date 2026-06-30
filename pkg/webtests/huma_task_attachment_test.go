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
	"strconv"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/routes"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// multipartFilesBody builds a multipart/form-data body with one or more files
// under the "files" field, matching the v2 upload handler's form schema.
func multipartFilesBody(t *testing.T, files map[string][]byte) (*bytes.Buffer, string) {
	t.Helper()
	buf := &bytes.Buffer{}
	w := multipart.NewWriter(buf)
	for filename, content := range files {
		fw, err := w.CreateFormFile("files", filename)
		require.NoError(t, err)
		_, err = fw.Write(content)
		require.NoError(t, err)
	}
	require.NoError(t, w.Close())
	return buf, w.FormDataContentType()
}

func uploadAttachmentRequest(t *testing.T, e *echo.Echo, taskID string, body *bytes.Buffer, contentType, token string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/v2/tasks/"+taskID+"/attachments", body)
	req.Header.Set("Content-Type", contentType)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

// uploadOneAttachment uploads a single file to task 1 and returns the created
// attachment id, so download/delete tests have a real file in storage to act on
// (setupTestEnv resets the mem storage, so fixture files have no bytes).
func uploadOneAttachment(t *testing.T, e *echo.Echo, token, filename string, content []byte) int64 {
	t.Helper()
	body, contentType := multipartFilesBody(t, map[string][]byte{filename: content})
	rec := uploadAttachmentRequest(t, e, "1", body, contentType, token)
	require.Equal(t, http.StatusCreated, rec.Code, "body: %s", rec.Body.String())

	var resp struct {
		Body struct {
			Success []*models.TaskAttachment `json:"success"`
			Errors  []struct {
				Message string `json:"message"`
			} `json:"errors"`
		}
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp.Body))
	require.Empty(t, resp.Body.Errors, "upload reported per-file errors: %+v", resp.Body.Errors)
	require.Len(t, resp.Body.Success, 1)
	require.NotZero(t, resp.Body.Success[0].ID)
	return resp.Body.Success[0].ID
}

func TestTaskAttachmentsV2(t *testing.T) {
	t.Run("Upload single file", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, &testuser1)

		body, contentType := multipartFilesBody(t, map[string][]byte{"hello.txt": []byte("hello world")})
		rec := uploadAttachmentRequest(t, e, "1", body, contentType, token)
		require.Equal(t, http.StatusCreated, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), "hello.txt")
		assert.Contains(t, rec.Body.String(), `"success"`)
	})

	t.Run("Upload multiple files", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, &testuser1)

		body, contentType := multipartFilesBody(t, map[string][]byte{
			"one.txt": []byte("first file"),
			"two.txt": []byte("second file"),
		})
		rec := uploadAttachmentRequest(t, e, "1", body, contentType, token)
		require.Equal(t, http.StatusCreated, rec.Code, "body: %s", rec.Body.String())

		var resp struct {
			Success []*models.TaskAttachment `json:"success"`
		}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Len(t, resp.Success, 2)
	})

	t.Run("List", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, &testuser1)

		// Upload first so there is at least one attachment with a real file row.
		uploadOneAttachment(t, e, token, "listed.txt", []byte("listed content"))

		rec := humaRequest(t, e, http.MethodGet, "/api/v2/tasks/1/attachments", "", token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())

		var resp struct {
			Items []*models.TaskAttachment `json:"items"`
			Total int64                    `json:"total"`
		}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.NotEmpty(t, resp.Items)
		assert.Positive(t, resp.Total)
	})

	t.Run("Download returns bytes and content type", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, &testuser1)

		content := []byte("downloadable content")
		id := uploadOneAttachment(t, e, token, "download.txt", content)

		rec := humaRequest(t, e, http.MethodGet, "/api/v2/tasks/1/attachments/"+strconv.FormatInt(id, 10), "", token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Equal(t, content, rec.Body.Bytes(), "the streamed file bytes must match the original")
		assert.NotEmpty(t, rec.Header().Get("Content-Type"))
		assert.Contains(t, rec.Header().Get("Content-Disposition"), "download.txt")
		// Caching headers mirror v1: a concrete length and a cacheable directive.
		assert.Equal(t, strconv.Itoa(len(content)), rec.Header().Get("Content-Length"))
		assert.Equal(t, "no-cache", rec.Header().Get("Cache-Control"))
		assert.NotEmpty(t, rec.Header().Get("Last-Modified"))
	})

	t.Run("Delete", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, &testuser1)

		id := uploadOneAttachment(t, e, token, "todelete.txt", []byte("bye"))

		rec := humaRequest(t, e, http.MethodDelete, "/api/v2/tasks/1/attachments/"+strconv.FormatInt(id, 10), "", token, "")
		require.Equal(t, http.StatusNoContent, rec.Code, "body: %s", rec.Body.String())

		// The download must now 404.
		rec = humaRequest(t, e, http.MethodGet, "/api/v2/tasks/1/attachments/"+strconv.FormatInt(id, 10), "", token, "")
		assert.Equal(t, http.StatusNotFound, rec.Code, "body: %s", rec.Body.String())
	})

	t.Run("Upload forbidden on inaccessible task", func(t *testing.T) {
		// Task 34 is owned by user 13 and inaccessible to testuser1 (see the v1 IDOR test).
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, &testuser1)

		body, contentType := multipartFilesBody(t, map[string][]byte{"nope.txt": []byte("nope")})
		rec := uploadAttachmentRequest(t, e, "34", body, contentType, token)
		assert.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
	})

	t.Run("List empty returns 200 not 500", func(t *testing.T) {
		// Regression: listing attachments on a task with zero attachments
		// returned HTTP 500 because ReadAll returned nil instead of an empty slice.
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, &testuser1)

		// Task 2 exists in project 1 (owned by testuser1) and has no attachment fixtures.
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/tasks/2/attachments", "", token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())

		var resp struct {
			Items []*models.TaskAttachment `json:"items"`
			Total int64                    `json:"total"`
		}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Empty(t, resp.Items)
		assert.Zero(t, resp.Total)
	})

	t.Run("List forbidden on inaccessible task", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, &testuser1)

		rec := humaRequest(t, e, http.MethodGet, "/api/v2/tasks/34/attachments", "", token, "")
		assert.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
	})

	t.Run("Download nonexistent attachment", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, &testuser1)

		rec := humaRequest(t, e, http.MethodGet, "/api/v2/tasks/1/attachments/99999", "", token, "")
		assert.Equal(t, http.StatusNotFound, rec.Code, "body: %s", rec.Body.String())
	})

	t.Run("Cannot download attachment that does not belong to the task in the path", func(t *testing.T) {
		// Mirrors the v1 IDOR test: attachment 4 belongs to task 34, not task 1.
		// Requesting it under task 1 (accessible) must 404, not leak the file.
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, &testuser1)

		rec := humaRequest(t, e, http.MethodGet, "/api/v2/tasks/1/attachments/4", "", token, "")
		assert.Equal(t, http.StatusNotFound, rec.Code, "body: %s", rec.Body.String())
	})

	t.Run("Unauthenticated upload is rejected", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		body, contentType := multipartFilesBody(t, map[string][]byte{"x.txt": []byte("x")})
		rec := uploadAttachmentRequest(t, e, "1", body, contentType, "")
		assert.Equal(t, http.StatusUnauthorized, rec.Code, "body: %s", rec.Body.String())
	})
}

// TestTaskAttachmentsV2_PreviewSize covers the preview_size query param: a non-image
// attachment ignores it and returns the original bytes (the v1 behaviour).
func TestTaskAttachmentsV2_PreviewSize(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	token := humaTokenFor(t, &testuser1)

	content := []byte("not an image, just text")
	id := uploadOneAttachment(t, e, token, "notimage.txt", content)

	rec := humaRequest(t, e, http.MethodGet, "/api/v2/tasks/1/attachments/"+strconv.FormatInt(id, 10)+"?preview_size=md", "", token, "")
	require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
	assert.Equal(t, content, rec.Body.Bytes(), "preview_size on a non-image must return the original file")
}

// TestTaskAttachmentsV2_Disabled proves the resource is absent when the
// service.enabletaskattachments config flag is off.
func TestTaskAttachmentsV2_Disabled(t *testing.T) {
	_, err := setupTestEnv()
	require.NoError(t, err)

	oldValue := config.ServiceEnableTaskAttachments.GetBool()
	config.ServiceEnableTaskAttachments.Set(false)
	defer config.ServiceEnableTaskAttachments.Set(oldValue)

	// Rebuild the router so RegisterAll re-evaluates the (now disabled) flag.
	e := routes.NewEcho()
	routes.RegisterRoutes(e)
	token := humaTokenFor(t, &testuser1)

	rec := humaRequest(t, e, http.MethodGet, "/api/v2/tasks/1/attachments", "", token, "")
	assert.Equal(t, http.StatusNotFound, rec.Code,
		"attachment routes must not be registered when the flag is off; body: %s", rec.Body.String())
}
