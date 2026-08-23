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

func uploadCardBackgroundRequest(t *testing.T, e *echo.Echo, project, token string, body *bytes.Buffer, contentType string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPut, "/api/v2/projects/"+project+"/card-backgrounds/upload", body)
	req.Header.Set("Content-Type", contentType)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func getCardBackgroundRequest(t *testing.T, e *echo.Echo, project, token string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/v2/projects/"+project+"/card-background", nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

// TestHumaProjectCardBackgroundDelete mirrors TestHumaProjectBackgroundDelete but
// also asserts the key independence guarantee: deleting the card image must not
// touch the project's separate full background (project 35 has both, on distinct
// files: background_file_id 1, card_background_file_id 2).
func TestHumaProjectCardBackgroundDelete(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)

	t.Run("Owner clears the card image, background untouched", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodDelete, "/api/v2/projects/35/card-background", "", humaTokenFor(t, &testuser6), "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())

		s := db.NewSession()
		defer s.Close()
		project := models.Project{ID: 35}
		has, err := s.Get(&project)
		require.NoError(t, err)
		require.True(t, has)
		assert.Equal(t, "Test35 with background", project.Title)
		assert.Equal(t, int64(0), project.CardBackgroundFileID)
		assert.Equal(t, int64(1), project.BackgroundFileID, "the full background must be unaffected")
	})
	t.Run("Read-only user is forbidden", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodDelete, "/api/v2/projects/35/card-background", "", humaTokenFor(t, &testuser15), "")
		assert.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
	})
	t.Run("No access at all is forbidden", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodDelete, "/api/v2/projects/35/card-background", "", humaTokenFor(t, &testuser1), "")
		assert.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
	})
}

// TestHumaCardBackgroundDisabledByConfig mirrors TestHumaBackgroundDisabledByConfig:
// card routes reuse the same backgrounds.enabled gate.
func TestHumaCardBackgroundDisabledByConfig(t *testing.T) {
	_, err := setupTestEnv()
	require.NoError(t, err)

	config.BackgroundsEnabled.Set(false)
	defer config.BackgroundsEnabled.Set(true)

	e := routes.NewEcho()
	routes.RegisterRoutes(e)

	rec := humaRequest(t, e, http.MethodDelete, "/api/v2/projects/35/card-background", "", humaTokenFor(t, &testuser6), "")
	assert.Equal(t, http.StatusNotFound, rec.Code, "route must be absent when backgrounds are disabled; body: %s", rec.Body.String())
}

// TestHumaProjectCardBackgroundUpload mirrors TestHumaProjectBackgroundUpload and
// additionally verifies uploading a card image on project 35 does not disturb its
// existing full background.
func TestHumaProjectCardBackgroundUpload(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)

	t.Run("Owner uploads a card image", func(t *testing.T) {
		// testuser1 owns project 1, which starts without a card image.
		body, contentType := multipartFileBody(t, "background", "card.png", pngBytes(t))
		rec := uploadCardBackgroundRequest(t, e, "1", humaTokenFor(t, &testuser1), body, contentType)
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())

		s := db.NewSession()
		defer s.Close()
		project := models.Project{ID: 1}
		has, err := s.Get(&project)
		require.NoError(t, err)
		require.True(t, has)
		assert.NotZero(t, project.CardBackgroundFileID, "the upload must set a card image file id")
		assert.NotEmpty(t, project.CardBackgroundBlurHash, "the upload must compute a blur hash")
	})

	t.Run("Uploading a card image leaves the existing background untouched", func(t *testing.T) {
		body, contentType := multipartFileBody(t, "background", "card.png", pngBytes(t))
		rec := uploadCardBackgroundRequest(t, e, "35", humaTokenFor(t, &testuser6), body, contentType)
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())

		s := db.NewSession()
		defer s.Close()
		project := models.Project{ID: 35}
		has, err := s.Get(&project)
		require.NoError(t, err)
		require.True(t, has)
		assert.Equal(t, int64(1), project.BackgroundFileID, "the full background must be unaffected")
	})

	t.Run("Non-image rejected with 400", func(t *testing.T) {
		body, contentType := multipartFileBody(t, "background", "not-an-image.txt", []byte("this is plain text, not an image"))
		rec := uploadCardBackgroundRequest(t, e, "1", humaTokenFor(t, &testuser1), body, contentType)
		require.Equal(t, http.StatusBadRequest, rec.Code, "body: %s", rec.Body.String())
	})

	t.Run("Read-only user is forbidden", func(t *testing.T) {
		body, contentType := multipartFileBody(t, "background", "card.png", pngBytes(t))
		rec := uploadCardBackgroundRequest(t, e, "35", humaTokenFor(t, &testuser15), body, contentType)
		assert.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
	})

	t.Run("Unauthenticated", func(t *testing.T) {
		body, contentType := multipartFileBody(t, "background", "card.png", pngBytes(t))
		rec := uploadCardBackgroundRequest(t, e, "1", "", body, contentType)
		require.Equal(t, http.StatusUnauthorized, rec.Code, "body: %s", rec.Body.String())
	})
}

// TestHumaProjectCardBackgroundDownload covers GET .../card-background, mirroring
// TestHumaProjectBackgroundDownload's upload-then-download pattern.
func TestHumaProjectCardBackgroundDownload(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)

	t.Run("Owner uploads then downloads the card image", func(t *testing.T) {
		body, contentType := multipartFileBody(t, "background", "card.png", pngBytes(t))
		up := uploadCardBackgroundRequest(t, e, "1", humaTokenFor(t, &testuser1), body, contentType)
		require.Equal(t, http.StatusOK, up.Code, "upload body: %s", up.Body.String())

		rec := getCardBackgroundRequest(t, e, "1", humaTokenFor(t, &testuser1))
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Equal(t, "image/jpg", rec.Header().Get("Content-Type"))
		assert.NotEmpty(t, rec.Body.Bytes(), "the download must return the stored bytes")
	})

	t.Run("Project without a card image returns 404", func(t *testing.T) {
		// testuser1 owns project 21, which has neither a background nor a card image.
		rec := getCardBackgroundRequest(t, e, "21", humaTokenFor(t, &testuser1))
		assert.Equal(t, http.StatusNotFound, rec.Code, "body: %s", rec.Body.String())
	})

	t.Run("No access at all is forbidden", func(t *testing.T) {
		rec := getCardBackgroundRequest(t, e, "35", humaTokenFor(t, &testuser1))
		assert.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
	})

	t.Run("Unauthenticated", func(t *testing.T) {
		rec := getCardBackgroundRequest(t, e, "35", "")
		assert.Equal(t, http.StatusUnauthorized, rec.Code, "body: %s", rec.Body.String())
	})
}

// TestHumaCardBackgroundUnsplash covers the card-image Unsplash "set" route's auth
// and permission gates, mirroring TestHumaUnsplashBackground. The shared
// /backgrounds/unsplash/search and proxy routes are already covered there.
func TestHumaCardBackgroundUnsplash(t *testing.T) {
	_, err := setupTestEnv()
	require.NoError(t, err)

	config.BackgroundsEnabled.Set(true)
	config.BackgroundsUnsplashEnabled.Set(true)
	defer config.BackgroundsUnsplashEnabled.Set(false)

	e := routes.NewEcho()
	routes.RegisterRoutes(e)

	t.Run("Set requires auth", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodPut, "/api/v2/projects/35/card-backgrounds/unsplash", `{"id":"abc"}`, "", "")
		assert.Equal(t, http.StatusUnauthorized, rec.Code, "body: %s", rec.Body.String())
	})
	t.Run("Set forbidden for read-only user", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodPut, "/api/v2/projects/35/card-backgrounds/unsplash", `{"id":"abc"}`, humaTokenFor(t, &testuser15), "")
		assert.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
	})
}
