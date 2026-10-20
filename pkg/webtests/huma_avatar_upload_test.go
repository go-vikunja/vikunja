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
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"image"
	"image/color"
	"image/png"
	"io/fs"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/modules/imageutils"
	"code.vikunja.io/api/pkg/user"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const avatarUploadPath = "/api/v2/user/settings/avatar"

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

func multipartAvatarBodyWithPartType(t *testing.T, fieldName, filename, partContentType string, content []byte) (*bytes.Buffer, string) {
	t.Helper()
	buf := &bytes.Buffer{}
	w := multipart.NewWriter(buf)
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name=%q; filename=%q`, fieldName, filename))
	h.Set("Content-Type", partContentType)
	fw, err := w.CreatePart(h)
	require.NoError(t, err)
	_, err = fw.Write(content)
	require.NoError(t, err)
	require.NoError(t, w.Close())
	return buf, w.FormDataContentType()
}

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

		s := db.NewSession()
		defer s.Close()
		u, err := user.GetUserByID(s, testuser1.ID)
		require.NoError(t, err)
		assert.Equal(t, "upload", u.AvatarProvider)
		assert.NotZero(t, u.AvatarFileID)
	})

	t.Run("Real image content-type on the part", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, &testuser1)

		body, contentType := multipartAvatarBodyWithPartType(t, "avatar", "avatar.png", "image/png", pngBytes(t))
		rec := uploadAvatarRequest(t, e, body, contentType, token)
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), "uploaded successfully")

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

		s := db.NewSession()
		defer s.Close()
		u, err := user.GetUserByID(s, testuser1.ID)
		require.NoError(t, err)
		assert.NotEqual(t, "upload", u.AvatarProvider)
	})

	t.Run("Undecodable image (SVG) rejected with 400", func(t *testing.T) {
		// Sniffs as image/svg+xml but has no decoder: 400, not 500.
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, &testuser1)

		svg := []byte(`<?xml version="1.0" encoding="UTF-8"?>` +
			`<svg xmlns="http://www.w3.org/2000/svg" width="8" height="8"><rect width="8" height="8" fill="#abc"/></svg>`)
		body, contentType := multipartAvatarBody(t, "avatar", "avatar.svg", svg)
		rec := uploadAvatarRequest(t, e, body, contentType, token)
		require.Equal(t, http.StatusBadRequest, rec.Code, "body: %s", rec.Body.String())

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

		var spec map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &spec))

		paths, _ := spec["paths"].(map[string]any)
		op, _ := paths["/user/settings/avatar"].(map[string]any)
		put, ok := op["put"].(map[string]any)
		require.True(t, ok, "PUT /user/settings/avatar must be in the spec")
		content, _ := put["requestBody"].(map[string]any)
		assert.Equal(t, true, content["required"], "avatar upload body must be required")
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

func pngWithClaimedDimensionsPatch(t *testing.T, width, height int) []byte {
	t.Helper()
	data := pngBytes(t)
	binary.BigEndian.PutUint32(data[16:20], uint32(width))  // #nosec G115 - test constant
	binary.BigEndian.PutUint32(data[20:24], uint32(height)) // #nosec G115 - test constant
	crc := crc32.ChecksumIEEE(data[12:29])
	binary.BigEndian.PutUint32(data[29:33], crc)
	return data
}

func TestAvatarUploadRemovesReplacedFile(t *testing.T) {
	for _, path := range []string{avatarUploadPath, "/api/v1/user/settings/avatar/upload"} {
		t.Run(path, func(t *testing.T) {
			e, err := setupTestEnv()
			require.NoError(t, err)
			token := humaTokenFor(t, &testuser1)

			upload := func() int64 {
				body, contentType := multipartAvatarBody(t, "avatar", "avatar.png", pngBytes(t))
				req := httptest.NewRequest(http.MethodPut, path, body)
				req.Header.Set("Content-Type", contentType)
				req.Header.Set("Authorization", "Bearer "+token)
				rec := httptest.NewRecorder()
				e.ServeHTTP(rec, req)
				require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())

				s := db.NewSession()
				defer s.Close()
				u, err := user.GetUserByID(s, testuser1.ID)
				require.NoError(t, err)
				return u.AvatarFileID
			}

			oldID := upload()
			require.NotZero(t, oldID)
			_, err = files.FileStat(&files.File{ID: oldID})
			require.NoError(t, err)

			newID := upload()
			require.NotEqual(t, oldID, newID)
			db.AssertMissing(t, "files", map[string]interface{}{"id": oldID})
			_, err = files.FileStat(&files.File{ID: oldID})
			require.ErrorIs(t, err, fs.ErrNotExist)
		})
	}
}

// Guards image allocation bounds (GHSA-4vh2-39rq-rq8j).
func TestAvatarUploadImageBounds(t *testing.T) {
	t.Run("8000 by 8000 avatar upload is rejected", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, &testuser1)

		body, contentType := multipartAvatarBody(t, "avatar", "bomb.png", pngWithClaimedDimensionsPatch(t, 8000, 8000))
		rec := uploadAvatarRequest(t, e, body, contentType, token)
		assert.Equal(t, http.StatusBadRequest, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"code":%d`, imageutils.ErrCodeImageTooLarge))
	})

	t.Run("avatar reads are fitted, never upscaled", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, &testuser1)

		// Larger than the smallest allowed avatar size (16), so a fit is observable.
		src := &bytes.Buffer{}
		require.NoError(t, png.Encode(src, image.NewRGBA(image.Rect(0, 0, 64, 64))))
		body, contentType := multipartAvatarBody(t, "avatar", "avatar.png", src.Bytes())
		rec := uploadAvatarRequest(t, e, body, contentType, token)
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())

		get := func(t *testing.T, size string) image.Image {
			rec := humaRequest(t, e, http.MethodGet, "/api/v2/avatar/"+testuser1.Username+"?size="+size, "", token, "")
			require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
			img, _, err := image.Decode(bytes.NewReader(rec.Body.Bytes()))
			require.NoError(t, err)
			return img
		}

		t.Run("smaller size scales down", func(t *testing.T) {
			img := get(t, "16")
			assert.Equal(t, 16, img.Bounds().Dx())
			assert.Equal(t, 16, img.Bounds().Dy())
		})
		t.Run("larger size keeps the native size", func(t *testing.T) {
			img := get(t, "1024")
			assert.Equal(t, 64, img.Bounds().Dx(), "Fit must not upscale a small source")
			assert.Equal(t, 64, img.Bounds().Dy())
		})
	})
}

func TestAvatarDecodeIsBounded(t *testing.T) {
	t.Run("busy upload is a 503 and keeps the previous avatar", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, &testuser1)

		body, contentType := multipartAvatarBody(t, "avatar", "avatar.png", pngBytes(t))
		rec := uploadAvatarRequest(t, e, body, contentType, token)
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())

		release := imageutils.ExhaustBudgetForTests(10 * time.Millisecond)
		t.Cleanup(release)
		body, contentType = multipartAvatarBody(t, "avatar", "avatar.png", pngBytes(t))
		rec = uploadAvatarRequest(t, e, body, contentType, token)
		assert.Equal(t, http.StatusServiceUnavailable, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"code":%d`, imageutils.ErrCodeDecodeBusy))

		release()
		rec = humaRequest(t, e, http.MethodGet, "/api/v2/avatar/"+testuser1.Username+"?size=7", "", token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		_, _, err = image.Decode(bytes.NewReader(rec.Body.Bytes()))
		require.NoError(t, err, "the previous avatar must still be readable")
	})

	t.Run("queued upload holds no transaction", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, &testuser1)

		release := imageutils.ExhaustBudgetForTests(time.Minute)
		t.Cleanup(release)

		ctx := &doneSignalCtx{Context: t.Context(), waiting: make(chan struct{})}
		body, contentType := multipartAvatarBody(t, "avatar", "avatar.png", pngBytes(t))
		req := httptest.NewRequest(http.MethodPut, avatarUploadPath, body).WithContext(ctx)
		req.Header.Set("Content-Type", contentType)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		finished := make(chan struct{})
		go func() {
			defer close(finished)
			e.ServeHTTP(rec, req)
		}()

		select {
		case <-ctx.waiting:
		case <-time.After(5 * time.Second):
			require.FailNow(t, "the upload never queued for decode capacity")
		}
		s := db.NewSession()
		_, err = s.ID(testuser2.ID).Cols("name").Update(&user.User{Name: "written while an upload is queued"})
		if err == nil {
			err = s.Commit()
		}
		s.Close()
		require.NoError(t, err)
		db.AssertExists(t, "users", map[string]interface{}{
			"id":   testuser2.ID,
			"name": "written while an upload is queued",
		}, false)

		release()
		select {
		case <-finished:
		case <-time.After(10 * time.Second):
			require.FailNow(t, "the upload did not finish after the budget was released")
		}
		assert.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
	})

	t.Run("busy avatar read is a 503", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, &testuser1)

		body, contentType := multipartAvatarBody(t, "avatar", "avatar.png", pngBytes(t))
		rec := uploadAvatarRequest(t, e, body, contentType, token)
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())

		t.Cleanup(imageutils.ExhaustBudgetForTests(10 * time.Millisecond))
		rec = humaRequest(t, e, http.MethodGet, "/api/v2/avatar/"+testuser1.Username+"?size=6", "", token, "")
		assert.Equal(t, http.StatusServiceUnavailable, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"code":%d`, imageutils.ErrCodeDecodeBusy))
	})
}
