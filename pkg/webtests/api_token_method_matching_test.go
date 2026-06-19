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
	"net/http/httptest"
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAPITokenMethodMatching is the standing guard for GHSA-v479-vf79-mg83:
// for every advertised permission it builds a single-permission token and
// asserts CanDoAPIRoute matches exactly the permission's stored (method,
// path) across every registered route. Any future contributor who adds a
// non-CRUD route on a shared path, or otherwise reintroduces method
// confusion, fails here. The tasks.read_all quirk is the only exception.
func TestAPITokenMethodMatching(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)

	type apiRoute struct{ Method, Path string }
	var allRoutes []apiRoute
	for _, r := range e.Router().Routes() {
		if !strings.HasPrefix(r.Path, "/api/v1") {
			continue
		}
		if r.Method == "echo_route_not_found" {
			continue
		}
		allRoutes = append(allRoutes, apiRoute{Method: r.Method, Path: r.Path})
	}
	require.NotEmpty(t, allRoutes, "echo router should have registered routes")

	advertised := models.GetAPITokenRoutes()
	require.NotEmpty(t, advertised, "GetAPITokenRoutes should be populated by RegisterRoutes")

	// Spec the matcher must conform to.
	expectedAuthorized := func(group, perm string, rd *models.RouteDetail, method, path string) bool {
		if rd.Method == method && rd.Path == path {
			return true
		}
		if group == "tasks" && perm == "read_all" && method == "GET" &&
			(path == "/api/v1/tasks" || path == "/api/v1/projects/:project/tasks") {
			return true
		}
		return false
	}

	for group, perms := range advertised {
		for perm, rd := range perms {
			token := &models.APIToken{
				APIPermissions: models.APIPermissions{group: []string{perm}},
			}

			req := httptest.NewRequest(rd.Method, rd.Path, nil)
			c := e.NewContext(req, httptest.NewRecorder())
			assert.Truef(t, models.CanDoAPIRoute(c, token),
				"%s.%s must authorize its own stored route %s %s",
				group, perm, rd.Method, rd.Path,
			)

			for _, r := range allRoutes {
				want := expectedAuthorized(group, perm, rd, r.Method, r.Path)
				req := httptest.NewRequest(r.Method, r.Path, nil)
				c := e.NewContext(req, httptest.NewRecorder())
				got := models.CanDoAPIRoute(c, token)
				assert.Equalf(t, want, got,
					"token %s.%s (stored for %s %s) on request %s %s: got=%v want=%v",
					group, perm, rd.Method, rd.Path,
					r.Method, r.Path, got, want,
				)
			}
		}
	}
}
