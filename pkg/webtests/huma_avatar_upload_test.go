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
	"image"
	"image/color"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// avatarUploadPath is the v2 endpoint under test.
const avatarUploadPath = "/api/v2/user/settings/avatar"

// pngBytes builds a small valid PNG so StoreAvatarFile can decode + resize it.
func pngBytes(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	for x := 0; x < 8; x++ {
		for y := 0; y < 8; y++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 16), G: uint8(y * 16), B: 100, A: 255})
		}
	}
	buf := &bytes.Buffer{}
	require.NoError(t, png.Encode(buf, img))
	return buf.Bytes()
}

// multipartAvatarBody returns a multipart/form-data body with a single
// "avatar" file field plus the matching Content-Type header (with boundary).
func multipartAvatarBody(t *testing.T, fieldName, filename string, content []byte) (*bytes.Buffer, string) {
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

// uploadAvatarRequest dispatches a multipart avatar upload against a prepared echo.Echo.
func uploadAvatarRequest(t *testing.T, e *echo.Echo, body *bytes.Buffer, contentType, token string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPut, avatarUploadPath, body)
	req.Header.Set("Content-Type", contentType)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func TestAvatarUpload(t *testing.T) {
	t.Run("Normal", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, &testuser1)

		body, contentType := multipartAvatarBody(t, "avatar", "avatar.png", pngBytes(t))
		rec := uploadAvatarRequest(t, e, body, contentType, token)
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), "uploaded successfully")

		// The provider must be flipped to "upload" and an avatar file stored.
		s := db.NewSession()
		defer s.Close()
		u, err := user.GetUserByID(s, testuser1.ID)
		require.NoError(t, err)
		assert.Equal(t, "upload", u.AvatarProvider)
		assert.NotZero(t, u.AvatarFileID)
	})

	t.Run("Non-image rejected", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, &testuser1)

		body, contentType := multipartAvatarBody(t, "avatar", "not-an-image.txt", []byte("this is plain text, not an image"))
		rec := uploadAvatarRequest(t, e, body, contentType, token)
		require.Equal(t, http.StatusBadRequest, rec.Code, "body: %s", rec.Body.String())

		// The provider must NOT have been changed.
		s := db.NewSession()
		defer s.Close()
		u, err := user.GetUserByID(s, testuser1.ID)
		require.NoError(t, err)
		assert.NotEqual(t, "upload", u.AvatarProvider)
	})

	t.Run("Unauthenticated", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		body, contentType := multipartAvatarBody(t, "avatar", "avatar.png", pngBytes(t))
		rec := uploadAvatarRequest(t, e, body, contentType, "")
		require.Equal(t, http.StatusUnauthorized, rec.Code, "body: %s", rec.Body.String())
	})

	t.Run("Renders as multipart in the OpenAPI spec", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/api/v2/openapi.json", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)

		// Navigate the spec loosely: Huma can emit `type` as either a string
		// or an array, so avoid binding it to a concrete Go type.
		var spec map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &spec))

		paths, _ := spec["paths"].(map[string]any)
		op, _ := paths["/user/settings/avatar"].(map[string]any)
		put, ok := op["put"].(map[string]any)
		require.True(t, ok, "PUT /user/settings/avatar must be in the spec")
		content, _ := put["requestBody"].(map[string]any)
		contentMap, _ := content["content"].(map[string]any)
		mp, ok := contentMap["multipart/form-data"].(map[string]any)
		require.True(t, ok, "avatar upload must be modeled as multipart/form-data")
		schema, _ := mp["schema"].(map[string]any)
		props, _ := schema["properties"].(map[string]any)
		avatarProp, ok := props["avatar"].(map[string]any)
		require.True(t, ok, "the avatar field must appear in the multipart schema")
		assert.Equal(t, "binary", avatarProp["format"], "avatar field must be a binary file in the spec")
	})
}
