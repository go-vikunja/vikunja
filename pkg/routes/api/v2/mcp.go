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

package apiv2

import (
	"context"
	"net/http"
	"strings"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/modules/mcp"
	"code.vikunja.io/api/pkg/user"
	"github.com/danielgtaylor/huma/v2"
)

func init() { AddRouteRegistrar(RegisterMCPRoutes) }

func RegisterMCPRoutes(api huma.API) {
	Register(api, huma.Operation{
		OperationID: "mcp-info",
		Summary:     "Get MCP connection settings",
		Description: "Returns the endpoint and usable token permissions with presets derived from exposed tools. Requires a user session; API tokens and link shares are rejected.",
		Method:      http.MethodGet, Path: "/mcp/info", Tags: []string{"mcp"},
	}, mcpInfo)
}

func mcpInfo(ctx context.Context, _ *struct{}) (*singleBody[mcp.ConnectionSettings], error) {
	if ec := echoContextFromCtx(ctx); ec != nil && ec.Get("api_token") != nil {
		return nil, huma.Error403Forbidden("API tokens cannot access MCP connection settings")
	}
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	if _, err = user.GetFromAuth(a); err != nil {
		return nil, translateDomainError(err)
	}
	endpoint := strings.TrimRight(config.ServicePublicURL.GetString(), "/") + mcp.RoutePrefix
	info := mcp.ConnectionInfo(endpoint)
	return &singleBody[mcp.ConnectionSettings]{Body: &info}, nil
}
