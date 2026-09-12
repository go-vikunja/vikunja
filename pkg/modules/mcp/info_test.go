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

package mcp

import (
	"net/http"
	"testing"

	"code.vikunja.io/api/pkg/models"
	"github.com/danielgtaylor/huma/v2"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

func TestMCPConnectionInfo(t *testing.T) {
	for _, route := range []echo.RouteInfo{
		{Path: "/api/v2/tasks/:id", Method: http.MethodPut},
		{Path: "/api/v2/tasks/:id", Method: http.MethodGet},
		{Path: "/api/v2/projects/:project/labels", Method: http.MethodGet},
		{Path: "/api/v2/projects/:project/webhooks", Method: http.MethodGet},
		{Path: "/api/v1/labels/:id", Method: http.MethodGet},
	} {
		models.CollectRoutesForAPITokenUsage(route, true)
	}
	toolsMu.Lock()
	previous := toolOrder
	toolOrder = []*tool{
		{echoPath: "/api/v2/tasks/:id", op: &huma.Operation{Method: http.MethodPatch}, tier: TierTyped},
		{echoPath: "/api/v2/tasks/:id", op: &huma.Operation{Method: http.MethodGet}, tier: TierTyped},
		{echoPath: "/api/v2/projects/:project/labels", op: &huma.Operation{Method: http.MethodGet}, tier: TierCatalog},
		{echoPath: "/api/v2/labels/:id", op: &huma.Operation{Method: http.MethodGet}, tier: TierTyped},
	}
	toolsMu.Unlock()
	t.Cleanup(func() { toolsMu.Lock(); toolOrder = previous; toolsMu.Unlock() })
	info := ConnectionInfo("https://example.com/api/v2/mcp")
	for _, tc := range []struct {
		group, permission string
		included          bool
	}{
		{"mcp", "access", true},
		{"tasks", "update", true},
		{"tasks", "read_one", true},
		{"projects", "labels", true},
		{"projects_webhooks", "read_all", false},
		{"labels", "read_one", false},
	} {
		_, has := info.Routes[tc.group][tc.permission]
		assert.Equal(t, tc.included, has, "%s:%s", tc.group, tc.permission)
	}
	assert.Equal(t, models.APIPermissions{"tasks": {"read_one", "update"}}, info.Presets.Typed)
	assert.Equal(t, http.MethodPatch, info.Routes["tasks"]["update"].Method)
	assert.Equal(t, map[string][]string{"*": {"read_one", "read_all"}}, info.Presets.ReadOnly)
	assert.Equal(t, map[string]string{"*": "*"}, info.Presets.Full)
}
