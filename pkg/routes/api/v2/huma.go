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
	"code.vikunja.io/api/pkg/version"

	"github.com/danielgtaylor/huma/v2"
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
	// Partial-update permissive: non-pointer fields do not become required
	// at the schema layer (legacy /v1 handlers are permissive by convention;
	// govalidator enforces real field-level rules later).
	cfg.FieldsOptionalByDefault = true

	return humaecho5.NewWithGroup(e, g, GroupPrefix, cfg)
}
