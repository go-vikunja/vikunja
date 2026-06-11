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
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/routes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHumaInfo covers the public instance-info endpoint. It needs no auth and
// always reports the running version.
func TestHumaInfo(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)

	rec := humaRequest(t, e, http.MethodGet, "/api/v2/info", "", "", "")
	require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Contains(t, body, "version")
	assert.Contains(t, body, "auth")
	assert.Contains(t, body, "available_migrators")
}

// TestHumaWebhookEvents covers the available-webhook-events listing. The route
// is only registered when webhooks are enabled (the test config default).
func TestHumaWebhookEvents(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)

	t.Run("Returns the events", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/webhooks/events", "", humaTokenFor(t, &testuser1), "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())

		var events []string
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &events))
		assert.ElementsMatch(t, models.GetAvailableWebhookEvents(), events)
	})
	t.Run("Unauthenticated", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/webhooks/events", "", "", "")
		assert.Equal(t, http.StatusUnauthorized, rec.Code, "body: %s", rec.Body.String())
	})
}

// TestHumaProjectBackgroundDelete covers removing a project background. It
// mirrors the v1 background_test.go matrix: the owner clears the background
// (and keeps the title), a read-only user is refused.
func TestHumaProjectBackgroundDelete(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)

	t.Run("Owner clears the background, title preserved", func(t *testing.T) {
		// testuser6 owns project 35 (title "Test35 with background", background_file_id 1).
		rec := humaRequest(t, e, http.MethodDelete, "/api/v2/projects/35/background", "", humaTokenFor(t, &testuser6), "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())

		s := db.NewSession()
		defer s.Close()
		project := models.Project{ID: 35}
		has, err := s.Get(&project)
		require.NoError(t, err)
		require.True(t, has)
		assert.Equal(t, "Test35 with background", project.Title)
		assert.Equal(t, int64(0), project.BackgroundFileID)
	})
	t.Run("Read-only user is forbidden", func(t *testing.T) {
		// testuser15 has read-only (permission 0) access to project 35.
		rec := humaRequest(t, e, http.MethodDelete, "/api/v2/projects/35/background", "", humaTokenFor(t, &testuser15), "")
		assert.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
	})
	t.Run("No access at all is forbidden", func(t *testing.T) {
		// testuser1 has no access to project 35.
		rec := humaRequest(t, e, http.MethodDelete, "/api/v2/projects/35/background", "", humaTokenFor(t, &testuser1), "")
		assert.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
	})
}

// TestHumaBackgroundDisabledByConfig verifies the registrar early-returns when
// project backgrounds are disabled: the DELETE route is then absent (404).
func TestHumaBackgroundDisabledByConfig(t *testing.T) {
	_, err := setupTestEnv()
	require.NoError(t, err)

	config.BackgroundsEnabled.Set(false)
	defer config.BackgroundsEnabled.Set(true)

	e := routes.NewEcho()
	routes.RegisterRoutes(e)

	rec := humaRequest(t, e, http.MethodDelete, "/api/v2/projects/35/background", "", humaTokenFor(t, &testuser6), "")
	assert.Equal(t, http.StatusNotFound, rec.Code, "route must be absent when backgrounds are disabled; body: %s", rec.Body.String())
}

// TestHumaUnsplashBackground covers the Unsplash routes' auth and permission
// gates. They are only registered when the unsplash provider is enabled (off by
// default), so the router is rebuilt with the flag on. The set route's
// permission check runs before any Unsplash network call, so the negative cases
// are exercised without hitting the real API; the happy path needs the network
// and is therefore not covered here (matching v1).
func TestHumaUnsplashBackground(t *testing.T) {
	_, err := setupTestEnv()
	require.NoError(t, err)

	config.BackgroundsEnabled.Set(true)
	config.BackgroundsUnsplashEnabled.Set(true)
	defer config.BackgroundsUnsplashEnabled.Set(false)

	e := routes.NewEcho()
	routes.RegisterRoutes(e)

	t.Run("Search requires auth", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/backgrounds/unsplash/search?q=mountain", "", "", "")
		assert.Equal(t, http.StatusUnauthorized, rec.Code, "body: %s", rec.Body.String())
	})
	t.Run("Set requires auth", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodPut, "/api/v2/projects/35/backgrounds/unsplash", `{"id":"abc"}`, "", "")
		assert.Equal(t, http.StatusUnauthorized, rec.Code, "body: %s", rec.Body.String())
	})
	t.Run("Set forbidden for read-only user", func(t *testing.T) {
		// testuser15 has read-only access to project 35; CanUpdate fails before
		// p.Set reaches Unsplash.
		rec := humaRequest(t, e, http.MethodPut, "/api/v2/projects/35/backgrounds/unsplash", `{"id":"abc"}`, humaTokenFor(t, &testuser15), "")
		assert.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
	})
}
