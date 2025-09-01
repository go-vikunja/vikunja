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
}

// registerRoutes takes a slice of APIRoute structs and registers them with both the Echo router
// and our token permission system. This replaces the old "magic" permission detection.
func registerRoutes(a *echo.Group, routes []APIRoute) {
	for _, route := range routes {
		// 1. Register the route with the Echo web server
		a.Add(route.Method, route.Path, route.Handler)

		// 2. Explicitly register the route and its permission scope
		//    with our API token system. This replaces the old "magic".
		models.CollectRoute(route.Method, route.Path, route.PermissionScope)
	}
}
