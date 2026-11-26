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
	"strings"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth/oauth2server"
	apiv2 "code.vikunja.io/api/pkg/routes/api/v2"
)

// ConnectionInfo reports which token permissions the tools of this module can actually use.
func (m *Module) ConnectionInfo() apiv2.ConnectionSettings {
	routes := map[string]models.APITokenRoute{
		"mcp": {"access": {Path: RoutePrefix, Method: "ANY"}},
	}
	typed := models.APIPermissions{}
	for group, permissions := range models.GetAPITokenRoutes() {
		for permission := range permissions {
			token := &models.APIToken{APIPermissions: models.APIPermissions{group: {permission}}}
			for _, t := range m.order {
				if !token.CanUseRoute(t.echoPath, t.op.Method) {
					continue
				}
				if routes[group] == nil {
					routes[group] = models.APITokenRoute{}
				}
				routes[group][permission] = &models.RouteDetail{
					Path:   t.echoPath,
					Method: t.op.Method,
				}
				if t.typed && !slices.Contains(typed[group], permission) {
					typed[group] = append(typed[group], permission)
				}
			}
		}
	}
	for _, permissions := range typed {
		sort.Strings(permissions)
	}
	return apiv2.ConnectionSettings{
		Endpoint: m.ResourceURL(),
		Routes:   routes,
		Presets: apiv2.TokenPresets{
			ReadOnly: map[string][]string{"*": {"read_one", "read_all"}},
			Typed:    typed,
			Full:     map[string]string{"*": "*"},
		},
	}
}

// ResourceURL is the endpoint's absolute URL, the resource indicator OAuth clients request tokens for.
func (m *Module) ResourceURL() string {
	return strings.TrimRight(config.ServicePublicURL.GetString(), "/") + RoutePrefix
}

// ScopePermissions expands an OAuth scope the way the frontend expands the read-only and full presets.
func (m *Module) ScopePermissions(scope oauth2server.Scope) models.APIPermissions {
	readOnly := scope == oauth2server.ScopeMCPRead
	if !readOnly && scope != oauth2server.ScopeMCPWrite {
		return nil
	}
	permissions := models.APIPermissions{}
	for group, routes := range m.ConnectionInfo().Routes {
		for permission := range routes {
			if readOnly && group != "mcp" && permission != "read_one" && permission != "read_all" {
				continue
			}
			permissions[group] = append(permissions[group], permission)
		}
	}
	for _, granted := range permissions {
		sort.Strings(granted)
	}
	return permissions
}
