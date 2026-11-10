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

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/routes/api/shared"

	"github.com/danielgtaylor/huma/v2"
)

// VikunjaInfos keeps the shared schema name; fields added here stay off the frozen /api/v1.
type VikunjaInfos struct {
	shared.VikunjaInfos
	GlobalUserSearchEnabled bool `json:"global_user_search_enabled" doc:"Whether the user search matches every user partially by name, username and email and returns emails. When true, the per-user discoverable_by_name and discoverable_by_email settings have no effect."`
}

type infoBody struct {
	Body VikunjaInfos
}

// RegisterInfoRoutes wires the public instance-info endpoint onto the Huma API.
func RegisterInfoRoutes(api huma.API) {
	Register(api, huma.Operation{
		OperationID: "info",
		Summary:     "Instance info",
		Description: "Returns version, frontend URL, motd and the enabled features of this Vikunja instance. Public — no authentication required.",
		Method:      http.MethodGet,
		Path:        "/info",
		Tags:        []string{"service"},
		// Public: opt out of the globally-applied auth. The path is also listed
		// in unauthenticatedAPIPaths so the token middleware lets it through.
		Security: []map[string][]string{},
	}, info)
}

func init() { AddRouteRegistrar(RegisterInfoRoutes) }

func info(_ context.Context, _ *struct{}) (*infoBody, error) {
	return &infoBody{Body: VikunjaInfos{
		VikunjaInfos:            shared.BuildInfo(), //nolint:contextcheck // OIDC provider init deliberately uses a background context — provider lifetime exceeds the request
		GlobalUserSearchEnabled: config.ServiceEnableGlobalUserSearch.GetBool(),
	}}, nil
}
