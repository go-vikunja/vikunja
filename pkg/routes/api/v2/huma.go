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
	"code.vikunja.io/api/pkg/modules/humaecho5"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/autopatch"
	"github.com/labstack/echo/v5"
)

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

	api := humaecho5.NewWithGroup(e, g, cfg)
	if api.OpenAPI().Components.SecuritySchemes == nil {
		api.OpenAPI().Components.SecuritySchemes = map[string]*huma.SecurityScheme{}
	}
	api.OpenAPI().Components.SecuritySchemes["JWTKeyAuth"] = &huma.SecurityScheme{
		Type:         "http",
		Scheme:       "bearer",
		BearerFormat: "JWT",
	}
	autopatch.AutoPatch(api)
	return api
}
