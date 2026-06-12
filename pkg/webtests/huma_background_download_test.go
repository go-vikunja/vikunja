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
	"net/http/httptest"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/routes"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getBackgroundRequest issues a GET against the background download route with an
// optional If-Modified-Since header (humaRequest can't set arbitrary headers).
func getBackgroundRequest(t *testing.T, e *echo.Echo, project, token, ifModifiedSince string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/v2/projects/"+project+"/background", nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if ifModifiedSince != "" {
		req.Header.Set("If-Modified-Since", ifModifiedSince)
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

// TestHumaProjectBackgroundDownload covers GET /projects/{project}/background. The
// fixture file row (project 35, background_file_id 1) carries no bytes, so the happy
// path uploads a real background first (the "upload-then-download" pattern) before
// fetching it back.
func TestHumaProjectBackgroundDownload(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)

	t.Run("Owner uploads then downloads the background", func(t *testing.T) {
		// testuser1 owns project 1, which starts without a background.
		body, contentType := multipartFileBody(t, "background", "bg.png", pngBytes(t))
		up := uploadBackgroundRequest(t, e, "1", humaTokenFor(t, &testuser1), body, contentType)
		require.Equal(t, http.StatusOK, up.Code, "upload body: %s", up.Body.String())

		rec := getBackgroundRequest(t, e, "1", humaTokenFor(t, &testuser1), "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Equal(t, "image/jpg", rec.Header().Get("Content-Type"))
		assert.Equal(t, "no-cache", rec.Header().Get("Cache-Control"))
		assert.NotEmpty(t, rec.Body.Bytes(), "the download must return the stored bytes")
	})

	t.Run("If-Modified-Since returns 304", func(t *testing.T) {
		// The in-memory test storage reports a zero modtime, so any valid
		// If-Modified-Since is not-before it and yields a 304.
		body, contentType := multipartFileBody(t, "background", "bg.png", pngBytes(t))
		up := uploadBackgroundRequest(t, e, "1", humaTokenFor(t, &testuser1), body, contentType)
		require.Equal(t, http.StatusOK, up.Code, "upload body: %s", up.Body.String())

		rec := getBackgroundRequest(t, e, "1", humaTokenFor(t, &testuser1), "Wed, 21 Oct 2015 07:28:00 GMT")
		assert.Equal(t, http.StatusNotModified, rec.Code, "body: %s", rec.Body.String())
		assert.Empty(t, rec.Body.Bytes(), "a 304 must not carry a body")
	})

	t.Run("Project without a background returns 404", func(t *testing.T) {
		// testuser1 owns project 21, which has no background and isn't uploaded to
		// by any other subtest (project 1 is, and subtests share this env).
		rec := getBackgroundRequest(t, e, "21", humaTokenFor(t, &testuser1), "")
		assert.Equal(t, http.StatusNotFound, rec.Code, "body: %s", rec.Body.String())
	})

	t.Run("Read-only user may download", func(t *testing.T) {
		// testuser6 owns project 35 and uploads a real background; testuser15 has
		// read-only access, which CanRead allows for the download. Uploading first
		// gives the file real bytes (the fixture row has none).
		body, contentType := multipartFileBody(t, "background", "bg.png", pngBytes(t))
		up := uploadBackgroundRequest(t, e, "35", humaTokenFor(t, &testuser6), body, contentType)
		require.Equal(t, http.StatusOK, up.Code, "upload body: %s", up.Body.String())

		rec := getBackgroundRequest(t, e, "35", humaTokenFor(t, &testuser15), "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.NotEmpty(t, rec.Body.Bytes(), "the read-only user must receive the bytes")
	})

	t.Run("No access at all is forbidden", func(t *testing.T) {
		// testuser1 has no access to project 35.
		rec := getBackgroundRequest(t, e, "35", humaTokenFor(t, &testuser1), "")
		assert.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
	})

	t.Run("Unauthenticated", func(t *testing.T) {
		rec := getBackgroundRequest(t, e, "35", "", "")
		assert.Equal(t, http.StatusUnauthorized, rec.Code, "body: %s", rec.Body.String())
	})
}

// TestHumaProjectBackgroundDownloadDisabledByConfig verifies the download route is
// absent (404) when project backgrounds are disabled.
func TestHumaProjectBackgroundDownloadDisabledByConfig(t *testing.T) {
	_, err := setupTestEnv()
	require.NoError(t, err)

	config.BackgroundsEnabled.Set(false)
	defer config.BackgroundsEnabled.Set(true)

	e := routes.NewEcho()
	routes.RegisterRoutes(e)

	rec := getBackgroundRequest(t, e, "35", humaTokenFor(t, &testuser6), "")
	assert.Equal(t, http.StatusNotFound, rec.Code, "route must be absent when backgrounds are disabled; body: %s", rec.Body.String())
}

// TestHumaUnsplashProxy covers the Unsplash image/thumb proxy routes' gating and auth.
// They only register when the unsplash provider is enabled (off by default), so the
// router is rebuilt with the flag on. The proxy's happy path needs the live Unsplash
// API and is therefore not covered here, matching v1 (which has no proxy tests).
func TestHumaUnsplashProxy(t *testing.T) {
	_, err := setupTestEnv()
	require.NoError(t, err)

	t.Run("Routes absent when unsplash is disabled", func(t *testing.T) {
		// Unsplash is disabled by default; the proxy routes must not exist.
		e := routes.NewEcho()
		routes.RegisterRoutes(e)

		rec := humaRequest(t, e, http.MethodGet, "/api/v2/backgrounds/unsplash/images/abc", "", humaTokenFor(t, &testuser1), "")
		assert.Equal(t, http.StatusNotFound, rec.Code, "image proxy must be absent when unsplash is disabled; body: %s", rec.Body.String())

		rec = humaRequest(t, e, http.MethodGet, "/api/v2/backgrounds/unsplash/images/abc/thumb", "", humaTokenFor(t, &testuser1), "")
		assert.Equal(t, http.StatusNotFound, rec.Code, "thumb proxy must be absent when unsplash is disabled; body: %s", rec.Body.String())
	})

	t.Run("Proxies require auth when unsplash is enabled", func(t *testing.T) {
		config.BackgroundsUnsplashEnabled.Set(true)
		defer config.BackgroundsUnsplashEnabled.Set(false)

		e := routes.NewEcho()
		routes.RegisterRoutes(e)

		rec := humaRequest(t, e, http.MethodGet, "/api/v2/backgrounds/unsplash/images/abc", "", "", "")
		assert.Equal(t, http.StatusUnauthorized, rec.Code, "image proxy body: %s", rec.Body.String())

		rec = humaRequest(t, e, http.MethodGet, "/api/v2/backgrounds/unsplash/images/abc/thumb", "", "", "")
		assert.Equal(t, http.StatusUnauthorized, rec.Code, "thumb proxy body: %s", rec.Body.String())
	})
}
