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

package yaegi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/log"
	"github.com/labstack/echo/v5"
)

func TestPluginRoutesServeHTTP(t *testing.T) {
	log.InitLogger()
	loaded, err := LoadPluginFull(examplePluginDir)
	if err != nil {
		t.Fatalf("LoadPluginFull failed: %v", err)
	}

	if loaded.UnauthRouter == nil {
		t.Fatal("UnauthRouter is nil — cannot test route registration")
	}

	// Create a real Echo instance and register the plugin's unauthenticated routes
	e := echo.New()
	g := e.Group("/plugins")
	loaded.UnauthRouter.RegisterUnauthenticatedRoutes(g)

	// Make an HTTP request to the plugin's /status endpoint
	req := httptest.NewRequest(http.MethodGet, "/plugins/status", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body: %s", rec.Code, rec.Body.String())
	}

	body := rec.Body.String()
	if !strings.Contains(body, "example") {
		t.Errorf("response body should contain plugin name 'example', got: %s", body)
	}
	if !strings.Contains(body, "ok") {
		t.Errorf("response body should contain status 'ok', got: %s", body)
	}

	t.Logf("HTTP response: %s", body)
}
