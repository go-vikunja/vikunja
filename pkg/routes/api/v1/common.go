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

package v1

import (
	"code.vikunja.io/api/pkg/models"
	"github.com/labstack/echo/v4"
)

// APIRoute defines the metadata for a single API endpoint.
// This centralizes route definition to avoid duplication across route files.
type APIRoute struct {
	Method          string
	Path            string
	Handler         echo.HandlerFunc
	PermissionScope string
	// AdminOnly indicates that this route requires an admin-level API token.
	// Standard tokens will be denied access even if they have the correct permission scope.
	AdminOnly bool
}

// registerRoutes takes a slice of APIRoute structs and registers them with both the Echo router
// and our token permission system. This replaces the old "magic" permission detection.
// This is the v1-specific wrapper that calls RegisterRoutes with "v1" version.
func registerRoutes(a *echo.Group, routes []APIRoute) {
	RegisterRoutes(a, routes, "v1")
}

// RegisterRoutes is the exported version that can be used by both v1 and v2 routes.
// It takes a slice of APIRoute structs and registers them with both the Echo router
// and our token permission system with the specified API version.
func RegisterRoutes(a *echo.Group, routes []APIRoute, version string) {
	// Build the API prefix based on the version (e.g., "/api/v1" or "/api/v2")
	apiPrefix := "/api/" + version

	for _, route := range routes {
		// Build the full path by combining API prefix with route path
		fullPath := apiPrefix + route.Path

		// 1. FIRST: Explicitly register the route and its permission scope
		//    with our API token system using the FULL path (with /api/{version} prefix).
		//    This must happen BEFORE a.Add() so that the OnAddRouteHandler
		//    check can see it already exists.
		models.CollectRoute(route.Method, fullPath, route.PermissionScope, route.AdminOnly)

		// 2. THEN: Register the route with the Echo web server
		//    This triggers OnAddRouteHandler, but our explicit registration
		//    will be detected and the legacy system will skip it.
		a.Add(route.Method, route.Path, route.Handler)
	}
}
