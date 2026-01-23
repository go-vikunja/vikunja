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

package models

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"
)

var apiTokenRoutes = map[string]APITokenRoute{}

func init() {
	apiTokenRoutes = make(map[string]APITokenRoute)
}

type APITokenRoute map[string]*RouteDetail

type RouteDetail struct {
	Path   string `json:"path"`
	Method string `json:"method"`
}

func getRouteGroupName(path string) (finalName string, filteredParts []string) {
	parts := strings.Split(strings.TrimPrefix(path, "/api/v1/"), "/")
	filteredParts = []string{}
	for _, part := range parts {
		if strings.HasPrefix(part, ":") {
			continue
		}

		filteredParts = append(filteredParts, part)
	}

	finalName = strings.Join(filteredParts, "_")
	switch finalName {
	case "projects_tasks":
		fallthrough
	case "tasks_all":
		return "tasks", []string{"tasks"}
	default:
		return finalName, filteredParts
	}
}

// getRouteDetail determines the API permission type from the route's HTTP method and path.
// In Echo v5, route.Name is auto-generated as METHOD:PATH, so we derive permissions from
// the HTTP method and path structure instead of the handler function name.
func getRouteDetail(route echo.RouteInfo) (method string, detail *RouteDetail) {
	detail = &RouteDetail{
		Path:   route.Path,
		Method: route.Method,
	}

	// Check if path ends with a parameter (e.g., /:id, /:task, /:project)
	pathParts := strings.Split(route.Path, "/")
	lastPart := ""
	if len(pathParts) > 0 {
		lastPart = pathParts[len(pathParts)-1]
	}
	endsWithParam := strings.HasPrefix(lastPart, ":")

	switch route.Method {
	case http.MethodGet:
		if endsWithParam {
			return "read_one", detail
		}
		return "read_all", detail
	case http.MethodPut:
		// PUT is used for creating resources in this codebase
		return "create", detail
	case http.MethodPost:
		// POST is used for updating resources
		return "update", detail
	case http.MethodDelete:
		return "delete", detail
	}

	return "", detail
}

func ensureAPITokenRoutesGroup(group string) {
	if _, has := apiTokenRoutes[group]; !has {
		apiTokenRoutes[group] = make(APITokenRoute)
	}
}

// isStandardCRUDRoute checks if a route follows the standard CRUD pattern.
// In Echo v5, route.Name is auto-generated as METHOD:PATH, so we can no longer
// check for "(*WebHandler)" in the name. Instead, we identify CRUD routes by:
// 1. Path structure: simple /resource or /resource/:param patterns
// 2. HTTP method: GET, PUT, POST, DELETE matching CRUD semantics
//
// Standard CRUD routes have paths like:
// - /projects, /tasks, /teams, /labels, /notifications, /webhooks, /filters, etc.
// - /projects/:project, /tasks/:task, /teams/:team, etc.
//
// Non-CRUD routes have paths with additional segments or special paths like:
// - /user/settings/email, /projects/:project/background, /backgrounds/unsplash/search
func isStandardCRUDRoute(routeGroupName string, routeParts []string, _ string) bool {
	// Standard CRUD resource groups that follow the WebHandler pattern
	crudResources := map[string]bool{
		"projects":             true,
		"tasks":                true,
		"teams":                true,
		"labels":               true,
		"filters":              true,
		"notifications":        true,
		"webhooks":             true,
		"reactions":            true,
		"shares":               true,
		"buckets":              true,
		"views":                true,
		"assignees":            true,
		"comments":             true,
		"relations":            true,
		"attachments":          true,
		"projects_views":       true,
		"projects_teams":       true,
		"projects_users":       true,
		"projects_shares":      true,
		"projects_webhooks":    true,
		"projects_buckets":     true,
		"tasks_attachments":    true,
		"tasks_assignees":      true,
		"tasks_labels":         true,
		"tasks_comments":       true,
		"tasks_relations":      true,
		"teams_members":        true,
		"projects_views_tasks": true,
	}

	// Check if this is a standard CRUD resource
	if crudResources[routeGroupName] {
		return true
	}

	// Also check the base resource for nested paths
	if len(routeParts) > 0 && crudResources[routeParts[0]] {
		// For single-segment paths, it's CRUD if it's a known resource
		if len(routeParts) == 1 {
			return true
		}
	}

	return false
}

// CollectRoutesForAPITokenUsage gets called for every added APITokenRoute and builds a list of all routes we can use for the api tokens.
// The requiresJWT parameter indicates if this route is protected by JWT authentication.
func CollectRoutesForAPITokenUsage(route echo.RouteInfo, requiresJWT bool) {

	if route.Method == "echo_route_not_found" {
		return
	}

	if !requiresJWT {
		return
	}

	routeGroupName, routeParts := getRouteGroupName(route.Path)

	if routeGroupName == "tokenTest" ||
		routeGroupName == "subscriptions" ||
		routeGroupName == "tokens" ||
		routeGroupName == "*" ||
		strings.HasPrefix(routeGroupName, "user_") {
		return
	}

	// Check if this is a standard CRUD route using path-based heuristics
	// In Echo v5, we can no longer rely on route.Name containing "(*WebHandler)"
	isCRUD := isStandardCRUDRoute(routeGroupName, routeParts, route.Method)

	// Special case for task attachments which use custom handlers
	isAttachmentRoute := routeGroupName == "tasks_attachments"

	if !isCRUD && !isAttachmentRoute {
		routeDetail := &RouteDetail{
			Path:   route.Path,
			Method: route.Method,
		}
		// We're trying to add routes to the routes of a matching "parent" - for
		// example, projects_background should show up under "projects".
		// To do this, we check if the route is a sub route of some other route
		// and if that's the case, add it to its parent instead.
		// Otherwise, we add it to the "other" key.
		if len(routeParts) == 1 {
			if routeGroupName == "notifications" && route.Method == http.MethodPost {
				ensureAPITokenRoutesGroup("notifications")

				apiTokenRoutes["notifications"]["mark_all_as_read"] = routeDetail
				return
			}

			ensureAPITokenRoutesGroup("other")

			_, exists := apiTokenRoutes["other"][routeGroupName]
			if exists {
				routeGroupName += "_" + strings.ToLower(route.Method)
			}
			apiTokenRoutes["other"][routeGroupName] = routeDetail
			return
		}

		subkey := strings.Join(routeParts[1:], "_")

		if _, has := apiTokenRoutes[routeParts[0]]; !has {
			apiTokenRoutes[routeParts[0]] = make(APITokenRoute)
		}

		if _, has := apiTokenRoutes[routeParts[0]][subkey]; has {
			subkey += "_" + strings.ToLower(route.Method)
		}

		apiTokenRoutes[routeParts[0]][subkey] = routeDetail

		return
	}

	if strings.HasSuffix(routeGroupName, "_bulk") {
		parent := strings.TrimSuffix(routeGroupName, "_bulk")
		ensureAPITokenRoutesGroup(parent)

		method, routeDetail := getRouteDetail(route)
		apiTokenRoutes[parent][method+"_bulk"] = routeDetail
		return
	}

	_, has := apiTokenRoutes[routeGroupName]
	if !has {
		apiTokenRoutes[routeGroupName] = make(APITokenRoute)
	}

	method, routeDetail := getRouteDetail(route)
	if method != "" {
		apiTokenRoutes[routeGroupName][method] = routeDetail
	}

	// Handle task attachments specially - they use custom handlers not WebHandler
	if routeGroupName == "tasks_attachments" {
		// PUT is upload (create), GET with :attachment param is download (read_one)
		if route.Method == http.MethodPut {
			apiTokenRoutes[routeGroupName]["create"] = &RouteDetail{
				Path:   route.Path,
				Method: route.Method,
			}
		}
		if route.Method == http.MethodGet && strings.HasSuffix(route.Path, ":attachment") {
			apiTokenRoutes[routeGroupName]["read_one"] = &RouteDetail{
				Path:   route.Path,
				Method: route.Method,
			}
		}
	}

}

// GetAvailableAPIRoutesForToken returns a list of all API routes which are available for token usage.
// @Summary Get a list of all token api routes
// @Description Returns a list of all API routes which are available to use with an api token, not a user login.
// @tags api
// @Produce json
// @Security JWTKeyAuth
// @Success 200 {array} models.APITokenRoute "The list of all routes."
// @Router /routes [get]
func GetAvailableAPIRoutesForToken(c *echo.Context) error {
	return c.JSON(http.StatusOK, apiTokenRoutes)
}

// CanDoAPIRoute checks if a token is allowed to use the current api route
func CanDoAPIRoute(c *echo.Context, token *APIToken) (can bool) {
	path := c.Path()
	if path == "" {
		// c.Path() is empty during testing, but returns the path which
		// the route used during registration which is what we need.
		path = c.Request().URL.Path
	}

	routeGroupName, routeParts := getRouteGroupName(path)

	routeGroupName = strings.TrimSuffix(routeGroupName, "_bulk")

	if routeGroupName == "user" ||
		routeGroupName == "users" ||
		routeGroupName == "routes" {
		routeGroupName = "other"
	}

	group, hasGroup := token.APIPermissions[routeGroupName]
	if !hasGroup {
		group, hasGroup = token.APIPermissions[routeParts[0]]
		if !hasGroup {
			return false
		}
	}

	var route string
	routes, has := apiTokenRoutes[routeGroupName]
	if !has {
		routes, has = apiTokenRoutes[routeParts[0]]
		if !has {
			return false
		}
		route = strings.Join(routeParts[1:], "_")
	}

	// The tasks read_all route is available as /:project/tasks and /tasks/all - therefore we need this workaround here.
	if routeGroupName == "tasks" && path == "/api/v1/projects/:project/tasks" && c.Request().Method == http.MethodGet {
		route = "read_all"
	}

	for _, p := range group {
		if route == "" && routes[p] != nil && routes[p].Path == path && routes[p].Method == c.Request().Method {
			return true
		}
		if route != "" && p == route {
			return true
		}
	}

	return false
}

func PermissionsAreValid(permissions APIPermissions) (err error) {

	for key, methods := range permissions {
		routes, has := apiTokenRoutes[key]
		if !has {
			return &ErrInvalidAPITokenPermission{
				Group: key,
			}
		}

		for _, method := range methods {
			if routes[method] == nil {
				return &ErrInvalidAPITokenPermission{
					Group:      key,
					Permission: method,
				}
			}
		}
	}

	return nil
}
