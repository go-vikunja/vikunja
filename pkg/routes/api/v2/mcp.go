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

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/humabridge"
	"code.vikunja.io/api/pkg/user"

	"github.com/danielgtaylor/huma/v2"
)

// ConnectionSettings lives here and not in the mcp module because that module imports this package.
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

// RegisterMCPInfo is called explicitly instead of through AddRouteRegistrar: settings are derived
// from the MCP module, which can only be built once every other operation is registered.
func RegisterMCPInfo(api huma.API, settings func() ConnectionSettings) {
	Register(api, huma.Operation{
		OperationID: "mcp-info",
		Summary:     "Get MCP connection settings",
		Description: "Returns the endpoint and usable token permissions with presets derived from exposed tools. Requires a user session; API tokens and link shares are rejected.",
		Method:      http.MethodGet,
		Path:        "/mcp/info",
		Tags:        []string{"mcp"},
	}, func(ctx context.Context, _ *struct{}) (*singleBody[ConnectionSettings], error) {
		return mcpInfo(ctx, settings)
	})
}

func mcpInfo(ctx context.Context, settings func() ConnectionSettings) (*singleBody[ConnectionSettings], error) {
	if ec := humabridge.EchoContextFrom(ctx); ec != nil && (*ec).Get("api_token") != nil {
		return nil, huma.Error403Forbidden("API tokens cannot access MCP connection settings")
	}
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	if _, err = user.GetFromAuth(a); err != nil {
		return nil, translateDomainError(err)
	}
	info := settings()
	return &singleBody[ConnectionSettings]{Body: &info}, nil
}
