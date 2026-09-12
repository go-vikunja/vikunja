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
	"slices"
	"sort"

	"code.vikunja.io/api/pkg/models"
)

type ConnectionSettings struct {
	Endpoint string                          `json:"endpoint" doc:"Absolute URL of the streamable HTTP endpoint."`
	Routes   map[string]models.APITokenRoute `json:"routes" doc:"Token permissions usable by exposed MCP tools, including mcp.access."`
	Presets  TokenPresets                    `json:"presets"`
}

type TokenPresets struct {
	ReadOnly map[string][]string   `json:"read_only" doc:"Wildcard permissions expanded against routes."`
	Typed    models.APIPermissions `json:"typed" doc:"Exact permissions used by the first-class tools."`
	Full     map[string]string     `json:"full" doc:"Wildcard granting all advertised permissions."`
}

func ConnectionInfo(endpoint string) ConnectionSettings {
	routes := map[string]models.APITokenRoute{
		"mcp": {"access": {Path: RoutePrefix, Method: "ANY"}},
	}
	typed := models.APIPermissions{}
	exposed := snapshotTools()
	for group, permissions := range models.GetAPITokenRoutes() {
		for permission := range permissions {
			token := &models.APIToken{APIPermissions: models.APIPermissions{group: {permission}}}
			for _, tool := range exposed {
				if !token.CanUseRoute(tool.echoPath, tool.op.Method) {
					continue
				}
				if routes[group] == nil {
					routes[group] = models.APITokenRoute{}
				}
				routes[group][permission] = &models.RouteDetail{Path: tool.echoPath, Method: tool.op.Method}
				if tool.tier == TierTyped {
					if !slices.Contains(typed[group], permission) {
						typed[group] = append(typed[group], permission)
					}
				}
			}
		}
	}
	for _, permissions := range typed {
		sort.Strings(permissions)
	}
	return ConnectionSettings{Endpoint: endpoint, Routes: routes, Presets: TokenPresets{
		ReadOnly: map[string][]string{"*": {"read_one", "read_all"}},
		Typed:    typed,
		Full:     map[string]string{"*": "*"},
	}}
}
