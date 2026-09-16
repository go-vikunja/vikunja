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
	"context"
	"net/http"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/models"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	config.ServicePublicURL.Set("https://example.com/")
	m := &Module{order: []*tool{
		{echoPath: "/api/v2/tasks/:id", op: &huma.Operation{Method: http.MethodPatch}, typed: true},
		{echoPath: "/api/v2/tasks/:id", op: &huma.Operation{Method: http.MethodGet}, typed: true},
		{echoPath: "/api/v2/projects/:project/labels", op: &huma.Operation{Method: http.MethodGet}},
		{echoPath: "/api/v2/labels/:id", op: &huma.Operation{Method: http.MethodGet}, typed: true},
	}}
	info := m.ConnectionInfo()
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
	assert.Equal(t, "https://example.com/api/v2/mcp", info.Endpoint)
	assert.Equal(t, models.APIPermissions{"tasks": {"read_one", "update"}}, info.Presets.Typed)
	assert.Equal(t, http.MethodPatch, info.Routes["tasks"]["update"].Method)
	assert.Equal(t, map[string][]string{"*": {"read_one", "read_all"}}, info.Presets.ReadOnly)
	assert.Equal(t, map[string]string{"*": "*"}, info.Presets.Full)
}

func TestMCPConnectionInfoAllowlist(t *testing.T) {
	_, api := humatest.New(t)
	for _, op := range []huma.Operation{
		{OperationID: "tasks-list", Method: http.MethodGet, Path: "/tasks"},
		{OperationID: "teams-list", Method: http.MethodGet, Path: "/teams"},
		{OperationID: "filters-preview", Method: http.MethodGet, Path: "/filters"},
	} {
		huma.Register(api, op, func(context.Context, *struct{}) (*struct{}, error) {
			return nil, nil
		})
		models.CollectRoutesForAPITokenUsage(echo.RouteInfo{
			Path:   "/api/v2" + op.Path,
			Method: op.Method,
		}, true)
	}
	require.Contains(t, models.GetAPITokenRoutes()["filters"], "read_all")
	m, err := New(api, nil)
	require.NoError(t, err)

	info := m.ConnectionInfo()
	assert.Contains(t, info.Routes["mcp"], "access")
	assert.Contains(t, info.Routes["tasks"], "read_all")
	assert.Contains(t, info.Routes["teams"], "read_all")
	assert.NotContains(t, info.Routes, "filters")
	assert.Equal(t, models.APIPermissions{"tasks": {"read_all"}}, info.Presets.Typed)
}
