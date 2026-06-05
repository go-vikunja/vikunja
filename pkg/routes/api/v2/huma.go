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

// Package apiv2 wires Huma onto the /api/v2 Echo group.
package apiv2

import (
	"context"
	"net/http"
	"strings"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/modules/humaecho5"
	"code.vikunja.io/api/pkg/version"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/autopatch"
	"github.com/labstack/echo/v5"
)

// GroupPrefix is the URL prefix the Echo group for /api/v2 is mounted at.
const GroupPrefix = "/api/v2"

// NewAPI mounts Huma on the /api/v2 group. Per-resource Register* calls
// live in sibling files.
func NewAPI(e *echo.Echo, g *echo.Group) huma.API {
	cfg := huma.DefaultConfig("Vikunja API", version.Version)
	cfg.OpenAPIPath = "/openapi"
	// Huma's built-in docs would load from unpkg.com — we serve Scalar locally instead.
	cfg.DocsPath = ""
	// Match v1's permissive partial-update convention; govalidator enforces real rules.
	cfg.FieldsOptionalByDefault = true

	api := humaecho5.NewWithGroup(e, g, GroupPrefix, cfg)
	oapi := api.OpenAPI()
	if oapi.Components.SecuritySchemes == nil {
		oapi.Components.SecuritySchemes = map[string]*huma.SecurityScheme{}
	}
	// v1 conflated JWTs and tk_-prefixed API tokens under JWTKeyAuth; v2
	// declares them separately so SDK generators and /api/v2/docs distinguish them.
	oapi.Components.SecuritySchemes["JWTKeyAuth"] = &huma.SecurityScheme{
		Type:         "http",
		Scheme:       "bearer",
		BearerFormat: "JWT",
		Description:  "User session JWT issued via /api/v1/login.",
	}
	oapi.Components.SecuritySchemes["APITokenAuth"] = &huma.SecurityScheme{
		Type:        "http",
		Scheme:      "bearer",
		Description: "Vikunja API token (tk_ prefix) with scoped permissions. Created via /api/v1/tokens.",
	}
	// Applied globally; public endpoints (spec, docs) opt out with an empty Security list.
	oapi.Security = []map[string][]string{
		{"JWTKeyAuth": {}},
		{"APITokenAuth": {}},
	}
	// The relative entry MUST stay at index 0. Huma's SchemaLinkTransformer
	// reads Servers[0] in three places (getAPIPrefix, addSchemaField, the
	// runtime Transform fallback). With a path-bearing absolute URL at
	// index 0, the runtime fallback concatenates that URL onto a ref that
	// already includes /api/v2, producing a double-prefixed $schema link
	// like https://host/api/v2/api/v2/schemas/Label.json. A relative URL
	// at index 0 keeps the prefix in the transformer's bookkeeping while
	// the absolute URL at index 1 advertises the deployment URL to SDK
	// generators and docs UIs.
	servers := []*huma.Server{{URL: GroupPrefix}}
	if publicURL := strings.TrimRight(config.ServicePublicURL.GetString(), "/"); publicURL != "" {
		servers = append(servers, &huma.Server{URL: publicURL + GroupPrefix})
	}
	oapi.Servers = servers
	return api
}

// Register wraps huma.Register with verb-based DefaultStatus: POST → 201,
// DELETE → 204. Anything else (including an explicit op.DefaultStatus) is untouched.
func Register[I, O any](api huma.API, op huma.Operation, handler func(context.Context, *I) (*O, error)) {
	if op.DefaultStatus == 0 {
		switch op.Method {
		case http.MethodPost:
			op.DefaultStatus = http.StatusCreated
		case http.MethodDelete:
			op.DefaultStatus = http.StatusNoContent
		}
	}
	huma.Register(api, op, handler)
}

// EnableAutoPatch synthesises a PATCH for every resource that already
// registered GET + PUT. Must be called AFTER all Register* calls.
func EnableAutoPatch(api huma.API) {
	autopatch.AutoPatch(api)

	// AutoPatch names each synthesised PATCH after the GET operation
	// ("Patch labels-read"), which reads poorly in the docs nav. Rewrite
	// the summary from the sibling PUT so it reads like "Update a label
	// (partial)". Only touch summaries AutoPatch generated (the "Patch "
	// prefix) so a hand-registered PATCH is left alone.
	for _, item := range api.OpenAPI().Paths {
		if item == nil || item.Patch == nil || item.Put == nil {
			continue
		}
		if item.Put.Summary != "" && strings.HasPrefix(item.Patch.Summary, "Patch ") {
			item.Patch.Summary = item.Put.Summary + " (partial)"
		}
	}
}
