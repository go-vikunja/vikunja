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

	"code.vikunja.io/api/pkg/modules/humaecho5"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/autopatch"
	"github.com/labstack/echo/v5"
)

// GroupPrefix is the URL prefix the Echo group for /api/v2 is mounted at.
// Exported for tests and to keep humaecho5's internal dispatch aligned
// with the real router paths.
const GroupPrefix = "/api/v2"

// NewAPI mounts Huma on the /api/v2 group and returns the Huma API for
// route registration. Configuration lives in this function; per-resource
// Register* calls happen in sibling files (labels.go, projects.go, ...).
func NewAPI(e *echo.Echo, g *echo.Group) huma.API {
	cfg := huma.DefaultConfig("Vikunja API", "2.0.0")
	// Serve the spec under the group so it lands at /api/v2/openapi.{json,yaml}.
	cfg.OpenAPIPath = "/openapi"
	// Huma's built-in docs would load from unpkg.com — unacceptable for
	// self-hosted Vikunja. Disable and serve Scalar ourselves (see Task D1).
	cfg.DocsPath = ""
	// Error shape is RFC 9457 problem+json by default — no override needed.
	// Partial-update permissive: non-pointer fields do not become required
	// at the schema layer (legacy /v1 handlers are permissive by convention;
	// govalidator enforces real field-level rules later).
	cfg.FieldsOptionalByDefault = true

	api := humaecho5.NewWithGroup(e, g, GroupPrefix, cfg)
	oapi := api.OpenAPI()
	if oapi.Components.SecuritySchemes == nil {
		oapi.Components.SecuritySchemes = map[string]*huma.SecurityScheme{}
	}
	// Two security schemes share the Authorization: Bearer header: JWT
	// (issued via /api/v1/login, refresh tokens rotated via cookie) and
	// Vikunja API tokens (tk_... prefix, scoped permissions). v1 declared
	// only JWTKeyAuth and conflated both under it; v2 declares them
	// separately so generated SDKs and /api/v2/docs distinguish them.
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
	// Applied globally to every registered operation; the handful of public
	// endpoints (spec, docs) explicitly opt out with Security: []map[...]{}.
	oapi.Security = []map[string][]string{
		{"JWTKeyAuth": {}},
		{"APITokenAuth": {}},
	}
	return api
}

// Register wraps huma.Register with verb-based DefaultStatus defaults so
// per-operation registrations don't have to spell out the obvious cases:
// POST → 201 Created, DELETE → 204 No Content. Anything else (including an
// explicit DefaultStatus on the operation) is left untouched.
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

// EnableAutoPatch registers a PATCH operation for every resource that has
// both a GET and a PUT already registered. Must be called AFTER all
// per-resource Register* calls — AutoPatch walks the already-registered
// operations to synthesize its PATCH counterparts.
func EnableAutoPatch(api huma.API) {
	autopatch.AutoPatch(api)
}
