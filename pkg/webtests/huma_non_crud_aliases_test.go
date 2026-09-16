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
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// feedsTokenUser13 is a feeds-scoped API token for user 13 (see the feeds
// fixtures); it authenticates the v2 notifications Atom feed via HTTP Basic.
const feedsTokenUser13 = "tk_feeds_access_token_user_0013_feed0013"

// TestHumaNonCRUDAliases covers the three non-REST endpoints mounted under
// /api/v2. Health and the Atom feed are Huma operations (so they appear in the
// OpenAPI spec); the WebSocket upgrade stays a raw echo route (OpenAPI can't
// model WebSockets). Each authenticates itself, so the group's JWT middleware
// must let them through.
func TestHumaNonCRUDAliases(t *testing.T) {
	t.Run("health is public and returns OK", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		rec := humaRequest(t, e, http.MethodGet, "/api/v2/health", "", "", "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), "OK")
	})

	t.Run("ws is reachable without a JWT", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		// A plain GET without the upgrade headers makes websocket.Accept reject
		// the request (typically 400). The point is that it reaches the handler
		// at all — not a 401 from the JWT middleware nor a 404 for an unmounted
		// route.
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/ws", "", "", "")
		assert.NotEqual(t, http.StatusUnauthorized, rec.Code, "ws must not be blocked by v2 JWT auth")
		assert.NotEqual(t, http.StatusNotFound, rec.Code, "ws must be mounted under /api/v2")
	})

	t.Run("atom feed is basic-auth-gated, not JWT-gated", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		t.Run("without credentials returns a basic-auth challenge", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v2/notifications.atom", nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			// The JWT middleware skips this path, so the handler's own HTTP Basic
			// auth gates it instead: a 401 carrying a Basic challenge, not the JWT
			// middleware's JSON error.
			require.Equal(t, http.StatusUnauthorized, rec.Code)
			assert.Contains(t, strings.ToLower(rec.Header().Get(echo.HeaderWWWAuthenticate)), "basic",
				"expected a Basic auth challenge, got %q", rec.Header().Get(echo.HeaderWWWAuthenticate))
		})

		t.Run("with a feeds API token returns an atom feed", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v2/notifications.atom", nil)
			req.SetBasicAuth("user13", feedsTokenUser13)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
			assert.True(t, strings.HasPrefix(rec.Header().Get(echo.HeaderContentType), "application/atom+xml"),
				"expected atom content type, got %q", rec.Header().Get(echo.HeaderContentType))
			assert.Contains(t, rec.Body.String(), "<feed", "expected an atom feed body")
		})
	})
}

// TestHumaNonCRUDAliasesInSpec is the load-bearing assertion: health and the
// Atom feed must show up as operations in the generated v2 OpenAPI document,
// while the raw WebSocket route must not (WebSockets can't be modeled).
func TestHumaNonCRUDAliasesInSpec(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/v2/openapi.json", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var spec map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &spec))

	paths, ok := spec["paths"].(map[string]any)
	require.True(t, ok, "spec must have a paths object")

	t.Run("health is a documented GET", func(t *testing.T) {
		op, ok := paths["/health"].(map[string]any)
		require.True(t, ok, "/health must be in the spec paths")
		_, ok = op["get"].(map[string]any)
		assert.True(t, ok, "/health must document a GET operation")
	})

	t.Run("notifications.atom is a documented GET with basic-auth security", func(t *testing.T) {
		op, ok := paths["/notifications.atom"].(map[string]any)
		require.True(t, ok, "/notifications.atom must be in the spec paths")
		get, ok := op["get"].(map[string]any)
		require.True(t, ok, "/notifications.atom must document a GET operation")

		// It returns application/atom+xml, not JSON.
		responses, _ := get["responses"].(map[string]any)
		ok200, _ := responses["200"].(map[string]any)
		content, _ := ok200["content"].(map[string]any)
		_, hasAtom := content["application/atom+xml"]
		assert.True(t, hasAtom, "200 response must declare application/atom+xml content")

		// The op documents its HTTP Basic auth honestly.
		security, _ := get["security"].([]any)
		require.NotEmpty(t, security, "op must declare a security requirement")
		first, _ := security[0].(map[string]any)
		_, hasBasic := first["BasicAuth"]
		assert.True(t, hasBasic, "op must require the BasicAuth scheme")

		comps, _ := spec["components"].(map[string]any)
		schemes, _ := comps["securitySchemes"].(map[string]any)
		basic, ok := schemes["BasicAuth"].(map[string]any)
		require.True(t, ok, "BasicAuth security scheme must be declared")
		assert.Equal(t, "http", basic["type"])
		assert.Equal(t, "basic", basic["scheme"])
	})

	t.Run("ws is absent from the spec", func(t *testing.T) {
		_, ok := paths["/ws"]
		assert.False(t, ok, "WebSockets can't be modeled in OpenAPI; /ws must stay a raw route")
	})
}
