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
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/license"

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
	// PUT is the authoritative update verb for API tokens — AutoPatch's
	// PATCH twin must not clobber it.
	assert.Equal(t, "PUT", labels["update"].Method)
	assert.Equal(t, "DELETE", labels["delete"].Method)
}

// TestCollectRoutesV2_Patch pins PUT as the stored update verb regardless of
// the order echo lists the AutoPatch twin, while a native PATCH with no PUT
// twin (as the admin routes have) is collected as-is.
func TestCollectRoutesV2_Patch(t *testing.T) {
	t.Run("PATCH listed before PUT still stores PUT", func(t *testing.T) {
		apiTokenRoutes = make(map[string]APITokenRoute)
		apiTokenRoutesV2 = make(map[string]APITokenRoute)

		CollectRoutesForAPITokenUsage(echo.RouteInfo{Method: "PATCH", Path: "/api/v2/labels/:id"}, true)
		CollectRoutesForAPITokenUsage(echo.RouteInfo{Method: "PUT", Path: "/api/v2/labels/:id"}, true)

		assert.Equal(t, "PUT", apiTokenRoutesV2["labels"]["update"].Method)
		assert.Len(t, apiTokenRoutesV2["labels"], 1)
	})

	t.Run("non-CRUD PUT twin is not suffixed", func(t *testing.T) {
		apiTokenRoutes = make(map[string]APITokenRoute)
		apiTokenRoutesV2 = make(map[string]APITokenRoute)

		CollectRoutesForAPITokenUsage(echo.RouteInfo{Method: "PUT", Path: "/api/v2/tasks/:task/position"}, true)
		CollectRoutesForAPITokenUsage(echo.RouteInfo{Method: "PATCH", Path: "/api/v2/tasks/:task/position"}, true)

		require.Contains(t, apiTokenRoutesV2["tasks"], "position")
		assert.Equal(t, "PUT", apiTokenRoutesV2["tasks"]["position"].Method)
		assert.NotContains(t, apiTokenRoutesV2["tasks"], "position_patch")
	})

	t.Run("native PATCH without a PUT twin is collected", func(t *testing.T) {
		apiTokenRoutes = make(map[string]APITokenRoute)
		apiTokenRoutesV2 = make(map[string]APITokenRoute)

		CollectRoutesForAPITokenUsage(echo.RouteInfo{Method: "PATCH", Path: "/api/v2/teams/:id/archive"}, true)

		require.Contains(t, apiTokenRoutesV2, "teams")
		require.Contains(t, apiTokenRoutesV2["teams"], "archive")
		assert.Equal(t, "PATCH", apiTokenRoutesV2["teams"]["archive"].Method)

		token := &APIToken{APIPermissions: APIPermissions{"teams": []string{"archive"}}}
		req := httptest.NewRequest("PATCH", "/api/v2/teams/:id/archive", nil)
		c := echo.New().NewContext(req, httptest.NewRecorder())
		assert.True(t, CanDoAPIRoute(c, token))
	})
}

// TestCollectRoutes_TimeEntriesV2 pins the v2-only time-entries resource to a
// snake_case "time_entries" group (not the "other" catch-all, not a hyphenated
// key the frontend's snake_case transform would mangle on save).
func TestCollectRoutes_TimeEntriesV2(t *testing.T) {
	apiTokenRoutes = make(map[string]APITokenRoute)
	apiTokenRoutesV2 = make(map[string]APITokenRoute)

	CollectRoutesForAPITokenUsage(echo.RouteInfo{Method: "GET", Path: "/api/v2/time-entries"}, true)
	CollectRoutesForAPITokenUsage(echo.RouteInfo{Method: "GET", Path: "/api/v2/time-entries/:id"}, true)
	CollectRoutesForAPITokenUsage(echo.RouteInfo{Method: "POST", Path: "/api/v2/time-entries"}, true)
	CollectRoutesForAPITokenUsage(echo.RouteInfo{Method: "PUT", Path: "/api/v2/time-entries/:id"}, true)
	CollectRoutesForAPITokenUsage(echo.RouteInfo{Method: "DELETE", Path: "/api/v2/time-entries/:id"}, true)

	_, isOther := apiTokenRoutesV2["other"]
	assert.False(t, isOther, "time-entries CRUD must not fall into the 'other' bucket")

	_, hyphenated := apiTokenRoutesV2["time-entries"]
	assert.False(t, hyphenated, "group key must be canonicalised to snake_case")

	te, has := apiTokenRoutesV2["time_entries"]
	require.True(t, has, "time_entries group should exist in the v2 table")
	assert.Equal(t, "GET", te["read_all"].Method)
	assert.Equal(t, "/api/v2/time-entries", te["read_all"].Path)
	assert.Equal(t, "GET", te["read_one"].Method)
	assert.Equal(t, "POST", te["create"].Method)
	assert.Equal(t, "PUT", te["update"].Method)
	assert.Equal(t, "DELETE", te["delete"].Method)
}

// TestGetAPITokenRoutes_ExposesV2Only verifies the /routes payload merges
// v2-only groups (time_entries has no v1 counterpart) so token clients can
// discover and grant them, without mutating the v1 table itself.
func TestGetAPITokenRoutes_ExposesV2Only(t *testing.T) {
	apiTokenRoutes = make(map[string]APITokenRoute)
	apiTokenRoutesV2 = make(map[string]APITokenRoute)
	license.SetForTests([]license.Feature{license.FeatureTimeTracking})
	defer license.ResetForTests()

	CollectRoutesForAPITokenUsage(echo.RouteInfo{Method: "GET", Path: "/api/v1/labels"}, true)
	CollectRoutesForAPITokenUsage(echo.RouteInfo{Method: "GET", Path: "/api/v2/time-entries"}, true)

	routes := GetAPITokenRoutes()

	_, hasLabels := routes["labels"]
	assert.True(t, hasLabels, "v1 groups stay exposed")

	te, hasTE := routes["time_entries"]
	require.True(t, hasTE, "v2-only time_entries must be exposed via /routes")
	assert.Equal(t, "GET", te["read_all"].Method)

	_, v1HasTE := apiTokenRoutes["time_entries"]
	assert.False(t, v1HasTE, "the merge must not mutate the v1 table")
}

// TestGetAPITokenRoutes_LicenseFilter verifies the /routes payload omits
// routes whose license feature is off — including licensed permissions nested
// inside always-available groups (tasks.time_entries) — while validation and
// authorisation of existing tokens stay unfiltered.
func TestGetAPITokenRoutes_LicenseFilter(t *testing.T) {
	resetAPITokenRoutes()

	CollectRoutesForAPITokenUsage(echo.RouteInfo{Method: "GET", Path: "/api/v1/labels"}, true)
	CollectRoutesForAPITokenUsage(echo.RouteInfo{Method: "GET", Path: "/api/v2/tasks"}, true)
	CollectRoutesForAPITokenUsage(echo.RouteInfo{Method: "GET", Path: "/api/v2/time-entries"}, true)
	CollectRoutesForAPITokenUsage(echo.RouteInfo{Method: "GET", Path: "/api/v2/tasks/:task_id/time-entries"}, true)

	t.Run("unlicensed", func(t *testing.T) {
		license.ResetForTests()
		routes := GetAPITokenRoutes()

		assert.Contains(t, routes, "labels")
		assert.NotContains(t, routes, "admin")
		assert.NotContains(t, routes, "time_entries")
		require.Contains(t, routes, "tasks")
		assert.Contains(t, routes["tasks"], "read_all")
		assert.NotContains(t, routes["tasks"], "time_entries")

		// Existing tokens with gated scopes must keep validating — the
		// request-time gates already make them inert.
		perms := APIPermissions{"time_entries": []string{"read_all"}}
		assert.NoError(t, PermissionsAreValid(perms))
	})

	t.Run("licensed", func(t *testing.T) {
		license.SetForTests([]license.Feature{license.FeatureTimeTracking, license.FeatureAdminPanel})
		defer license.ResetForTests()
		routes := GetAPITokenRoutes()

		assert.Contains(t, routes, "admin")
		require.Contains(t, routes, "time_entries")
		assert.Contains(t, routes["time_entries"], "read_all")
		require.Contains(t, routes, "tasks")
		assert.Contains(t, routes["tasks"], "time_entries")
	})
}

func TestValidateInstanceBotPermissions(t *testing.T) {
	require.NoError(t, validateInstanceBotPermissions(APIPermissions{"admin": {"users_list", "users_create"}}))
	require.NoError(t, validateInstanceBotPermissions(APIPermissions{}))

	for _, perms := range []APIPermissions{
		{"tasks": {"read_all"}},
		{"admin": {"users_list"}, "tasks": {"read_all"}},
		{"caldav": {"access"}},
	} {
		err := validateInstanceBotPermissions(perms)
		require.Error(t, err, "%v", perms)
		assert.True(t, IsErrInstanceBotScopeNotAllowed(err))
	}
}

// TestCanDoAPIRoute_TimeEntriesHyphenLegacy proves a token stored under the old
// hyphenated "time-entries" key still validates and authorises — no migration.
func TestCanDoAPIRoute_TimeEntriesHyphenLegacy(t *testing.T) {
	apiTokenRoutes = make(map[string]APITokenRoute)
	apiTokenRoutesV2 = make(map[string]APITokenRoute)

	CollectRoutesForAPITokenUsage(echo.RouteInfo{Method: "GET", Path: "/api/v2/time-entries"}, true)

	for _, key := range []string{"time_entries", "time-entries"} {
		t.Run(key, func(t *testing.T) {
			perms := APIPermissions{key: []string{"read_all"}}
			require.NoError(t, PermissionsAreValid(perms), "%s must validate", key)

			token := &APIToken{APIPermissions: perms}
			req := httptest.NewRequest("GET", "/api/v2/time-entries", nil)
			c := echo.New().NewContext(req, httptest.NewRecorder())
			assert.True(t, CanDoAPIRoute(c, token), "%s must authorise", key)
		})
	}
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

// TestCanDoAPIRoute_V2PatchAliasesPut verifies that a token granted the
// "update" permission on a v2 resource can issue PATCH requests against
// the same path as the stored PUT route. Huma's AutoPatch synthesises
// PATCH for every PUT — the matcher accepts it as an alias so token
// holders aren't forced to use PUT exclusively.
func TestCanDoAPIRoute_V2PatchAliasesPut(t *testing.T) {
	apiTokenRoutes = make(map[string]APITokenRoute)
	apiTokenRoutesV2 = make(map[string]APITokenRoute)
	apiTokenRoutes["caldav"] = APITokenRoute{
		"access": &RouteDetail{Path: "/dav/*", Method: "ANY"},
	}

	CollectRoutesForAPITokenUsage(echo.RouteInfo{Method: "PUT", Path: "/api/v2/labels/:id"}, true)
	CollectRoutesForAPITokenUsage(echo.RouteInfo{Method: "PATCH", Path: "/api/v2/labels/:id"}, true)

	token := &APIToken{
		APIPermissions: APIPermissions{"labels": []string{"update"}},
	}

	e := echo.New()

	t.Run("PUT is allowed (stored verb)", func(t *testing.T) {
		req := httptest.NewRequest("PUT", "/api/v2/labels/:id", nil)
		c := e.NewContext(req, httptest.NewRecorder())
		assert.True(t, CanDoAPIRoute(c, token))
	})

	t.Run("PATCH is allowed via alias", func(t *testing.T) {
		req := httptest.NewRequest("PATCH", "/api/v2/labels/:id", nil)
		c := e.NewContext(req, httptest.NewRecorder())
		assert.True(t, CanDoAPIRoute(c, token))
	})

	t.Run("PATCH on a different path is rejected", func(t *testing.T) {
		req := httptest.NewRequest("PATCH", "/api/v2/projects/:id", nil)
		c := e.NewContext(req, httptest.NewRecorder())
		assert.False(t, CanDoAPIRoute(c, token))
	})

	t.Run("v1 PATCH stays rejected", func(t *testing.T) {
		// The alias must not bleed onto v1 — v1 has no AutoPatch and
		// never registers PATCH on update routes.
		apiTokenRoutes["labels"] = APITokenRoute{
			"update": &RouteDetail{Path: "/api/v1/labels/:id", Method: "POST"},
		}
		v1Token := &APIToken{
			APIPermissions: APIPermissions{"labels": []string{"update"}},
		}
		req := httptest.NewRequest("PATCH", "/api/v1/labels/:id", nil)
		c := e.NewContext(req, httptest.NewRecorder())
		assert.False(t, CanDoAPIRoute(c, v1Token))
	})
}

// TestCanDoAPIRoute_V2TasksReadAll verifies that tasks.read_all authorises
// both the global /api/v2/tasks and project-scoped /api/v2/projects/:project/tasks
// endpoints. Both normalise to tasks.read_all via getRouteGroupName, but only
// one RouteDetail survives in the map — the special case in CanDoAPIRoute must
// accept either path.
func TestCanDoAPIRoute_V2TasksReadAll(t *testing.T) {
	apiTokenRoutes = make(map[string]APITokenRoute)
	apiTokenRoutesV2 = make(map[string]APITokenRoute)
	apiTokenRoutes["caldav"] = APITokenRoute{
		"access": &RouteDetail{Path: "/dav/*", Method: "ANY"},
	}

	CollectRoutesForAPITokenUsage(echo.RouteInfo{Method: "GET", Path: "/api/v2/tasks"}, true)
	CollectRoutesForAPITokenUsage(echo.RouteInfo{Method: "GET", Path: "/api/v2/projects/:project/tasks"}, true)

	token := &APIToken{
		APIPermissions: APIPermissions{"tasks": []string{"read_all"}},
	}

	e := echo.New()

	t.Run("global /api/v2/tasks", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v2/tasks", nil)
		c := e.NewContext(req, httptest.NewRecorder())
		assert.True(t, CanDoAPIRoute(c, token))
	})

	t.Run("project-scoped /api/v2/projects/:project/tasks", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v2/projects/:project/tasks", nil)
		c := e.NewContext(req, httptest.NewRecorder())
		assert.True(t, CanDoAPIRoute(c, token))
	})
}

// TestCollectRoutes_V2TasksBulkCreate pins the bulk create route to tasks.create_bulk instead of projects.tasks_bulk.
func TestCollectRoutes_V2TasksBulkCreate(t *testing.T) {
	apiTokenRoutes = make(map[string]APITokenRoute)
	apiTokenRoutesV2 = make(map[string]APITokenRoute)

	CollectRoutesForAPITokenUsage(echo.RouteInfo{
		Method: "POST",
		Path:   "/api/v2/projects/:project/tasks/bulk",
	}, true)

	tasks, has := apiTokenRoutesV2["tasks"]
	require.True(t, has, "tasks route group should exist")

	bulkRoute, has := tasks["create_bulk"]
	require.True(t, has, "create_bulk should exist in tasks routes")
	assert.Equal(t, "/api/v2/projects/:project/tasks/bulk", bulkRoute.Path)
	assert.Equal(t, "POST", bulkRoute.Method)

	_, underProjects := apiTokenRoutesV2["projects"]
	assert.False(t, underProjects, "bulk task create must not file under projects")
}

// TestCanDoAPIRoute_V2TasksBulkCreate verifies that tasks.create_bulk, not tasks.create, authorises the bulk create route.
func TestCanDoAPIRoute_V2TasksBulkCreate(t *testing.T) {
	apiTokenRoutes = make(map[string]APITokenRoute)
	apiTokenRoutesV2 = make(map[string]APITokenRoute)

	CollectRoutesForAPITokenUsage(echo.RouteInfo{
		Method: "POST",
		Path:   "/api/v2/projects/:project/tasks/bulk",
	}, true)

	e := echo.New()

	t.Run("create_bulk permission is allowed", func(t *testing.T) {
		token := &APIToken{
			APIPermissions: APIPermissions{"tasks": []string{"create_bulk"}},
		}
		req := httptest.NewRequest("POST", "/api/v2/projects/:project/tasks/bulk", nil)
		c := e.NewContext(req, httptest.NewRecorder())
		assert.True(t, CanDoAPIRoute(c, token))
	})

	t.Run("create permission alone is rejected", func(t *testing.T) {
		token := &APIToken{
			APIPermissions: APIPermissions{"tasks": []string{"create"}},
		}
		req := httptest.NewRequest("POST", "/api/v2/projects/:project/tasks/bulk", nil)
		c := e.NewContext(req, httptest.NewRecorder())
		assert.False(t, CanDoAPIRoute(c, token))
	})
}

// End-to-end CanDoAPIRoute coverage for /api/v2 is provided by the Label
// integration test in pkg/webtests/huma_label_test.go (see the token-auth
// scenarios in that file) which exercises the full auth pipeline.

// Guards cross-group expansions (GHSA-9rg3-v78m-26q8).
func TestCanDoAPIRoute_ExpandScopes(t *testing.T) {
	apiTokenRoutes = make(map[string]APITokenRoute)
	apiTokenRoutesV2 = make(map[string]APITokenRoute)

	for _, r := range []echo.RouteInfo{
		{Method: "GET", Path: "/api/v1/tasks"},
		{Method: "GET", Path: "/api/v1/tasks/:projecttask"},
		{Method: "GET", Path: "/api/v1/projects/:project/tasks"},
		{Method: "GET", Path: "/api/v1/projects/:project/tasks/by-index/:index"},
		{Method: "GET", Path: "/api/v1/projects/:project/views/:view/tasks"},
		{Method: "GET", Path: "/api/v1/projects/:project/views/:view/buckets"},
		{Method: "GET", Path: "/api/v2/tasks"},
		{Method: "GET", Path: "/api/v2/tasks/:projecttask"},
		{Method: "GET", Path: "/api/v2/projects/:project/tasks"},
		{Method: "GET", Path: "/api/v2/projects/:project/tasks/by-index/:index"},
		{Method: "GET", Path: "/api/v2/projects/:project/views/:view/tasks"},
		{Method: "GET", Path: "/api/v2/projects/:project/views/:view/buckets/tasks"},
	} {
		CollectRoutesForAPITokenUsage(r, true)
	}
	CollectRoutesForAPITokenUsage(echo.RouteInfo{Method: "GET", Path: "/api/v1/tasks/:task/comments"}, true)
	CollectRoutesForAPITokenUsage(echo.RouteInfo{Method: "GET", Path: "/api/v1/:entitykind/:entityid/reactions"}, true)
	CollectRoutesForAPITokenUsage(echo.RouteInfo{Method: "GET", Path: "/api/v1/time-entries"}, true)

	e := echo.New()
	do := func(_ *testing.T, url string, token *APIToken) bool {
		req := httptest.NewRequest("GET", url, nil)
		c := e.NewContext(req, httptest.NewRecorder())
		return CanDoAPIRoute(c, token)
	}

	basePerms := APIPermissions{
		"tasks":                []string{"read_all", "read_one"},
		"projects":             []string{"tasks_by_index", "views_buckets", "views_buckets_tasks"},
		"projects_views_tasks": []string{"read_all"},
	}
	tasksOnly := &APIToken{APIPermissions: basePerms}
	withScopesPerms := APIPermissions{
		"tasks_comments": []string{"read_all"},
		"reactions":      []string{"read_all"},
		"time_entries":   []string{"read_all"},
	}
	for k, v := range basePerms {
		withScopesPerms[k] = v
	}
	withScopes := &APIToken{APIPermissions: withScopesPerms}

	t.Run("each expand value requires its own scope", func(t *testing.T) {
		for _, expand := range []string{"comments", "comment_count", "reactions", "time_entries_count"} {
			assert.False(t, do(t, "/api/v1/tasks?expand="+expand, tasksOnly),
				"a tasks-only token must not expand %s", expand)
			assert.False(t, do(t, "/api/v2/tasks?expand="+expand, tasksOnly),
				"a tasks-only token must not expand %s", expand)
			assert.True(t, do(t, "/api/v1/tasks?expand="+expand, withScopes),
				"a token with the expansion scope may expand %s", expand)
			assert.True(t, do(t, "/api/v2/tasks?expand="+expand, withScopes))
		}
	})

	t.Run("repeated and comma-mixed expand params", func(t *testing.T) {
		assert.False(t, do(t, "/api/v1/tasks?expand=subtasks&expand=comments", tasksOnly))
		assert.True(t, do(t, "/api/v1/tasks?expand=subtasks&expand=comments", withScopes))
		assert.False(t, do(t, "/api/v2/tasks?expand=buckets,reactions", tasksOnly))
		assert.True(t, do(t, "/api/v2/tasks?expand=buckets,reactions", withScopes))
	})

	t.Run("unprotected expansions stay available to a tasks-only token", func(t *testing.T) {
		for _, expand := range []string{"subtasks", "buckets", "is_unread"} {
			assert.True(t, do(t, "/api/v1/tasks?expand="+expand, tasksOnly))
		}
		assert.True(t, do(t, "/api/v1/tasks?expand=subtasks&expand=buckets&expand=is_unread", tasksOnly))
		assert.True(t, do(t, "/api/v1/tasks", tasksOnly), "no expand at all must stay allowed")
	})

	t.Run("protected route shapes", func(t *testing.T) {
		for _, path := range []string{
			"/api/v1/tasks/:projecttask",
			"/api/v1/projects/:project/tasks",
			"/api/v1/projects/:project/tasks/by-index/:index",
			"/api/v1/projects/:project/views/:view/tasks",
			"/api/v1/projects/:project/views/:view/buckets",
			"/api/v2/tasks/:projecttask",
			"/api/v2/projects/:project/tasks",
			"/api/v2/projects/:project/tasks/by-index/:index",
			"/api/v2/projects/:project/views/:view/tasks",
			"/api/v2/projects/:project/views/:view/buckets/tasks",
		} {
			assert.False(t, do(t, path+"?expand=comments", tasksOnly), "%s must require the comment scope", path)
			assert.True(t, do(t, path+"?expand=comments", withScopes), "%s must allow the expansion with the scope", path)
		}
	})

	t.Run("expand is ignored on unrelated routes", func(t *testing.T) {
		CollectRoutesForAPITokenUsage(echo.RouteInfo{Method: "GET", Path: "/api/v1/projects"}, true)
		projectsToken := &APIToken{APIPermissions: APIPermissions{"projects": []string{"read_all"}}}
		assert.True(t, do(t, "/api/v1/projects?expand=comments", projectsToken),
			"expand on a route which does not consume it must not require any scope")
	})
}

// TestAdminTokenScopes covers the hand-named admin scopes: /routes lists only
// the new names, each name authorises its v1 and v2 route, the
// collision-derived keys on pre-existing tokens still authorise, and an
// admin-only token stays out of everything else.
func TestAdminTokenScopes(t *testing.T) {
	resetAPITokenRoutes()
	license.SetForTests([]license.Feature{license.FeatureAdminPanel})
	defer license.ResetForTests()

	// Derived collection must not reintroduce the old keys.
	CollectRoutesForAPITokenUsage(echo.RouteInfo{Method: "GET", Path: "/api/v1/admin/users"}, true)
	CollectRoutesForAPITokenUsage(echo.RouteInfo{Method: "POST", Path: "/api/v2/admin/users"}, true)
	CollectRoutesForAPITokenUsage(echo.RouteInfo{Method: "PATCH", Path: "/api/v2/admin/users/:id/admin"}, true)
	CollectRoutesForAPITokenUsage(echo.RouteInfo{Method: "GET", Path: "/api/v2/tasks"}, true)

	e := echo.New()
	can := func(token *APIToken, method, path string) bool {
		c := e.NewContext(httptest.NewRequest(method, path, nil), httptest.NewRecorder())
		return CanDoAPIRoute(c, token)
	}

	t.Run("routes lists only the named scopes", func(t *testing.T) {
		names := make([]string, 0, len(adminTokenRoutes))
		for _, r := range adminTokenRoutes {
			names = append(names, r.name)
		}
		listed := make([]string, 0, len(GetAPITokenRoutes()["admin"]))
		for name := range GetAPITokenRoutes()["admin"] {
			listed = append(listed, name)
		}
		assert.ElementsMatch(t, names, listed)
	})

	t.Run("each scope authorises its own route on both versions", func(t *testing.T) {
		for _, r := range adminTokenRoutes {
			token := &APIToken{APIPermissions: APIPermissions{"admin": []string{r.name}}}
			require.NoError(t, PermissionsAreValid(token.APIPermissions))
			assert.True(t, can(token, r.method, "/api/v2/"+r.path), "%s must authorise v2", r.name)
			assert.Equal(t, !r.v2Only, can(token, r.method, "/api/v1/"+r.path), "%s on v1", r.name)
			// No method confusion: GET must not unlock a PATCH on the same path and vice versa.
			other := http.MethodGet
			if r.method == http.MethodGet {
				other = http.MethodPatch
			}
			assert.False(t, can(token, other, "/api/v2/"+r.path), "%s must not authorise %s", r.name, other)
		}
	})

	t.Run("legacy keys still authorise", func(t *testing.T) {
		byName := map[string]string{}
		for _, r := range adminTokenRoutes {
			byName[r.name] = r.method + " " + r.path
		}
		for old, current := range legacyAdminScopes {
			parts := strings.SplitN(byName[current], " ", 2)
			token := &APIToken{APIPermissions: APIPermissions{"admin": []string{old}}}
			assert.True(t, can(token, parts[0], "/api/v2/"+parts[1]), "legacy %s must authorise %s", old, current)
		}
		// Unchanged names need no alias.
		assert.True(t, can(&APIToken{APIPermissions: APIPermissions{"admin": []string{"users_delete"}}}, http.MethodDelete, "/api/v1/admin/users/:id"))
		assert.True(t, can(&APIToken{APIPermissions: APIPermissions{"admin": []string{"overview"}}}, http.MethodGet, "/api/v2/admin/overview"))
	})

	t.Run("v1 PATCH admin routes", func(t *testing.T) {
		for _, key := range []string{"users_set_status", "users_status"} {
			token := &APIToken{APIPermissions: APIPermissions{"admin": []string{key}}}
			assert.True(t, can(token, http.MethodPatch, "/api/v1/admin/users/:id/status"), key)
			assert.True(t, can(token, http.MethodPatch, "/api/v2/admin/users/:id/status"), key)
		}
	})

	t.Run("admin-only token is denied elsewhere", func(t *testing.T) {
		all := make([]string, 0, len(adminTokenRoutes))
		for _, r := range adminTokenRoutes {
			all = append(all, r.name)
		}
		token := &APIToken{APIPermissions: APIPermissions{"admin": all}}
		assert.False(t, can(token, http.MethodGet, "/api/v2/tasks"))
		assert.False(t, can(token, http.MethodGet, "/api/v1/admin/users/:id"))
	})

	t.Run("legacy keys are not aliased outside the admin group", func(t *testing.T) {
		token := &APIToken{APIPermissions: APIPermissions{"tasks": []string{"users"}}}
		assert.False(t, can(token, http.MethodGet, "/api/v2/admin/users"))
	})
}
