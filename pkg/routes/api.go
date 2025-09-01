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

package routes

import (
	"code.vikunja.io/api/pkg/models"
	"github.com/labstack/echo/v4"
)

// APIRoute defines all the necessary metadata for a single API endpoint.
// This struct serves as the new "Single Source of Truth" for all information
// about a single API endpoint, replacing the fragile implicit permission mapping.
type APIRoute struct {
	Method          string            // HTTP method (GET, POST, PUT, DELETE, etc.)
	Path            string            // URL path pattern
	Handler         echo.HandlerFunc  // The handler function
	PermissionScope string            // Explicit permission scope (read_all, create, update, delete, etc.)
}

// Register takes a slice of APIRoute structs and registers them with both the Echo router
// and our token permission system. This replaces the old "magic" permission detection
// with explicit, declarative route definitions.
func Register(a *echo.Group, routes []APIRoute) {
	for _, route := range routes {
		// 1. Register the route with the Echo web server
		a.Add(route.Method, route.Path, route.Handler)

		// 2. Explicitly register the route and its permission scope
		//    with our API token system. This replaces the old "magic".
		models.CollectRoute(route.Method, route.Path, route.PermissionScope)
	}
}
