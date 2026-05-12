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

package models

import (
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCanDoAPIRoute_BulkLabelTask(t *testing.T) {
	// Reset apiTokenRoutes to isolate this test
	apiTokenRoutes = make(map[string]APITokenRoute)

	// Register the standard CRUD routes for tasks_labels first
	CollectRoutesForAPITokenUsage(echo.RouteInfo{
		Method: "PUT",
		Path:   "/api/v1/tasks/:projecttask/labels",
	}, true)
	CollectRoutesForAPITokenUsage(echo.RouteInfo{
		Method: "DELETE",
		Path:   "/api/v1/tasks/:projecttask/labels/:label",
	}, true)

	// Now register the bulk route
	CollectRoutesForAPITokenUsage(echo.RouteInfo{
		Method: "POST",
		Path:   "/api/v1/tasks/:projecttask/labels/bulk",
	}, true)

	// Verify that the tasks_labels route group exists
	routes, has := apiTokenRoutes["tasks_labels"]
	require.True(t, has, "tasks_labels route group should exist")

	// The bulk route should be registered as "update_bulk" under tasks_labels
	bulkRoute, has := routes["update_bulk"]
	require.True(t, has, "update_bulk should exist in tasks_labels routes")
	assert.Equal(t, "/api/v1/tasks/:projecttask/labels/bulk", bulkRoute.Path)
	assert.Equal(t, "POST", bulkRoute.Method)
}

func TestIsV2Path(t *testing.T) {
	cases := map[string]bool{
		"/api/v2":         true,
		"/api/v2/":        true,
		"/api/v2/labels":  true,
		"/api/v1/labels":  false,
		"/api/v1/api/v2":  false, // prefix is authoritative
		"":                false,
		"/api/v20/labels": false, // only exact /api/v2 prefix counts
		"/api/v2labels":   false,
	}
	for path, want := range cases {
		t.Run(path, func(t *testing.T) {
			assert.Equal(t, want, isV2Path(path))
		})
	}
}

func TestStripAPIVersion(t *testing.T) {
	cases := map[string]string{
		"/api/v1/labels":     "labels",
		"/api/v2/labels":     "labels",
		"/api/v2/labels/42":  "labels/42",
		"/api/v1/tasks/bulk": "tasks/bulk",
		"/api/v3/labels":     "/api/v3/labels", // unknown versions pass through
		"/labels":            "/labels",
		"":                   "",
	}
	for path, want := range cases {
		t.Run(path, func(t *testing.T) {
			assert.Equal(t, want, stripAPIVersion(path))
		})
	}
}

// TestCollectRoutesV2 verifies that /api/v2 routes are stored in the v2
// shadow table under the same (group, permission) keys their v1 counterparts
// would use. This is what lets a token scoped on `labels.read_one` authorise
// both /api/v1/labels/{id} and /api/v2/labels/{id}.
func TestCollectRoutesV2(t *testing.T) {
	apiTokenRoutes = make(map[string]APITokenRoute)
	apiTokenRoutesV2 = make(map[string]APITokenRoute)

	CollectRoutesForAPITokenUsage(echo.RouteInfo{Method: "GET", Path: "/api/v2/labels"}, true)
	CollectRoutesForAPITokenUsage(echo.RouteInfo{Method: "GET", Path: "/api/v2/labels/:id"}, true)
	CollectRoutesForAPITokenUsage(echo.RouteInfo{Method: "POST", Path: "/api/v2/labels"}, true)
	CollectRoutesForAPITokenUsage(echo.RouteInfo{Method: "PUT", Path: "/api/v2/labels/:id"}, true)
	CollectRoutesForAPITokenUsage(echo.RouteInfo{Method: "DELETE", Path: "/api/v2/labels/:id"}, true)
	CollectRoutesForAPITokenUsage(echo.RouteInfo{Method: "PATCH", Path: "/api/v2/labels/:id"}, true)

	// v1 map stays untouched.
	assert.Empty(t, apiTokenRoutes, "v2 routes must not land in the v1 table")

	labels, has := apiTokenRoutesV2["labels"]
	require.True(t, has, "labels group should exist in v2 table")
	assert.Equal(t, "GET", labels["read_all"].Method)
	assert.Equal(t, "/api/v2/labels", labels["read_all"].Path)
	assert.Equal(t, "GET", labels["read_one"].Method)
	assert.Equal(t, "POST", labels["create"].Method)
	// PUT is the authoritative update verb for API tokens — PATCH is
	// skipped during collection so it doesn't clobber PUT.
	assert.Equal(t, "PUT", labels["update"].Method)
	assert.Equal(t, "DELETE", labels["delete"].Method)
}

// TestGetRouteDetail_V2Verbs verifies the v2 verb mapping: POST→create,
// PUT/PATCH→update. v1 inverts POST and PUT so we need a separate mapping
// path.
func TestGetRouteDetail_V2Verbs(t *testing.T) {
	cases := []struct {
		method, path, wantPerm string
	}{
		{"GET", "/api/v2/labels", "read_all"},
		{"GET", "/api/v2/labels/:id", "read_one"},
		{"POST", "/api/v2/labels", "create"},
		{"PUT", "/api/v2/labels/:id", "update"},
		{"PATCH", "/api/v2/labels/:id", "update"},
		{"DELETE", "/api/v2/labels/:id", "delete"},
	}
	for _, c := range cases {
		t.Run(c.method+" "+c.path, func(t *testing.T) {
			perm, _ := getRouteDetail(echo.RouteInfo{Method: c.method, Path: c.path})
			assert.Equal(t, c.wantPerm, perm)
		})
	}
}

// End-to-end CanDoAPIRoute coverage for /api/v2 is provided by the Label
// integration test in pkg/webtests/huma_label_test.go (see the token-auth
// scenarios in that file) which exercises the full auth pipeline.
