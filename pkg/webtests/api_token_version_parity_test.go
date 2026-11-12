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
	"regexp"
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/license"
	"code.vikunja.io/api/pkg/models"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type tokenRoute struct{ Method, Path string }

var pathParam = regexp.MustCompile(`:[^/]+`)

func pathShape(path string) string {
	path = strings.TrimPrefix(strings.TrimPrefix(path, "/api/v1"), "/api/v2")
	return pathParam.ReplaceAllString(path, ":")
}

// GHSA-j5v7-qcwp-5g32
func TestAPITokenPermissionVersionParity(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)

	license.SetForTests([]license.Feature{license.FeatureAdminPanel, license.FeatureTimeTracking, license.FeatureUserInvites})
	defer license.ResetForTests()

	var routes []tokenRoute
	for _, r := range e.Router().Routes() {
		// PATCH is only an alias of the stored v2 PUT.
		if r.Method == echo.RouteNotFound || r.Method == http.MethodPatch && strings.HasPrefix(r.Path, "/api/v2") {
			continue
		}
		if strings.HasPrefix(r.Path, "/api/v1/") || strings.HasPrefix(r.Path, "/api/v2/") {
			routes = append(routes, tokenRoute{Method: r.Method, Path: r.Path})
		}
	}
	require.NotEmpty(t, routes)

	authorized := func(group, perm string) []tokenRoute {
		token := &models.APIToken{APIPermissions: models.APIPermissions{group: {perm}}}
		var got []tokenRoute
		for _, r := range routes {
			c := e.NewContext(httptest.NewRequest(r.Method, r.Path, nil), httptest.NewRecorder())
			if models.CanDoAPIRoute(c, token) {
				got = append(got, r)
			}
		}
		return got
	}

	cases := []struct {
		group, perm string
		want        []tokenRoute
	}{
		{"projects", "background", []tokenRoute{
			{http.MethodGet, "/api/v1/projects/:project/background"},
			{http.MethodGet, "/api/v2/projects/:project/background"},
		}},
		{"projects", "background_delete", []tokenRoute{
			{http.MethodDelete, "/api/v1/projects/:project/background"},
			{http.MethodDelete, "/api/v2/projects/:project/background"},
		}},
		{"projects", "backgrounds_upload", []tokenRoute{
			{http.MethodPut, "/api/v1/projects/:project/backgrounds/upload"},
			{http.MethodPut, "/api/v2/projects/:project/backgrounds/upload"},
		}},
		{"projects", "views_buckets", []tokenRoute{
			{http.MethodGet, "/api/v1/projects/:project/views/:view/buckets"},
			{http.MethodGet, "/api/v2/projects/:project/views/:view/buckets"},
		}},
		{"projects", "views_buckets_create", []tokenRoute{
			{http.MethodPut, "/api/v1/projects/:project/views/:view/buckets"},
			{http.MethodPost, "/api/v2/projects/:project/views/:view/buckets"},
		}},
		{"projects", "views_buckets_update", []tokenRoute{
			{http.MethodPost, "/api/v1/projects/:project/views/:view/buckets/:bucket"},
			{http.MethodPut, "/api/v2/projects/:project/views/:view/buckets/:bucket"},
		}},
		{"projects", "views_buckets_delete", []tokenRoute{
			{http.MethodDelete, "/api/v1/projects/:project/views/:view/buckets/:bucket"},
			{http.MethodDelete, "/api/v2/projects/:project/views/:view/buckets/:bucket"},
		}},
		{"projects", "views_buckets_tasks_update", []tokenRoute{
			{http.MethodPost, "/api/v1/projects/:project/views/:view/buckets/:bucket/tasks"},
			{http.MethodPut, "/api/v2/projects/:project/views/:view/buckets/:bucket/tasks"},
		}},
		{"projects", "views_buckets_tasks", []tokenRoute{
			{http.MethodGet, "/api/v2/projects/:project/views/:view/buckets/tasks"},
		}},
		{"reactions", "delete", []tokenRoute{
			{http.MethodPost, "/api/v1/:entitykind/:entityid/reactions/delete"},
			{http.MethodPost, "/api/v2/:entitykind/:entityid/reactions/delete"},
		}},
		{"notifications", "update", []tokenRoute{
			{http.MethodPost, "/api/v1/notifications/:notificationid"},
			{http.MethodPut, "/api/v2/notifications/:notificationid"},
		}},
		{"notifications", "mark_all_as_read", []tokenRoute{
			{http.MethodPost, "/api/v1/notifications"},
			{http.MethodPost, "/api/v2/notifications"},
		}},
		{"teams", "members_admin", []tokenRoute{
			{http.MethodPost, "/api/v1/teams/:team/members/:user/admin"},
			{http.MethodPost, "/api/v2/teams/:team/members/:user/admin"},
		}},
		{"admin", "users_create", []tokenRoute{
			{http.MethodPost, "/api/v1/admin/users"},
			{http.MethodPost, "/api/v2/admin/users"},
		}},
	}
	for _, tc := range cases {
		t.Run(tc.group+"."+tc.perm, func(t *testing.T) {
			assert.ElementsMatch(t, tc.want, authorized(tc.group, tc.perm))
		})
	}

	t.Run("every permission covers one resource and at most one route per version", func(t *testing.T) {
		for group, perms := range models.GetAPITokenRoutes() {
			for perm := range perms {
				// Two list endpoints share tasks.read_all by design.
				if group == "tasks" && perm == "read_all" {
					continue
				}
				got := authorized(group, perm)
				perVersion := map[bool]int{}
				for _, r := range got {
					perVersion[strings.HasPrefix(r.Path, "/api/v2/")]++
					assert.Equalf(t, pathShape(got[0].Path), pathShape(r.Path),
						"%s.%s authorizes different resources: %v", group, perm, got)
				}
				assert.LessOrEqualf(t, perVersion[false], 1, "%s.%s authorizes several v1 routes: %v", group, perm, got)
				assert.LessOrEqualf(t, perVersion[true], 1, "%s.%s authorizes several v2 routes: %v", group, perm, got)
			}
		}
	})
}
