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

	"code.vikunja.io/api/pkg/log"

	"github.com/labstack/echo/v5"
)

var apiTokenRoutes = map[string]APITokenRoute{}

// apiTokenRoutesV2 is a shadow routing table for /api/v2 routes keyed under the
// same (group, permission) names as their /api/v1 equivalents. The frontend
// token UI keeps listing only v1 routes via apiTokenRoutes; CanDoAPIRoute
// consults this second table so a token granted e.g. `labels.read_one` also
// authorizes the v2 endpoint.
var apiTokenRoutesV2 = map[string]APITokenRoute{}

func init() {
	apiTokenRoutes = make(map[string]APITokenRoute)
	apiTokenRoutesV2 = make(map[string]APITokenRoute)
	apiTokenRoutes["caldav"] = APITokenRoute{
		"access": &RouteDetail{
			Path:   "/dav/*",
			Method: "ANY",
		},
	}
}

type APITokenRoute map[string]*RouteDetail

type RouteDetail struct {
	Path   string `json:"path"`
	Method string `json:"method"`
}

// isV2Path reports whether the given route path lives under /api/v2.
func isV2Path(path string) bool {
	return strings.HasPrefix(path, "/api/v2/") || path == "/api/v2"
}

// stripAPIVersion removes the /api/v1/ or /api/v2/ prefix, so both versions
// are normalized to the same group name for token permission purposes.
func stripAPIVersion(path string) string {
	if stripped := strings.TrimPrefix(path, "/api/v1/"); stripped != path {
		return stripped
	}
	if stripped := strings.TrimPrefix(path, "/api/v2/"); stripped != path {
		return stripped
	}
	return path
}

func getRouteGroupName(path string) (finalName string, filteredParts []string) {
	parts := strings.Split(stripAPIVersion(path), "/")
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
//
// v1 and v2 have inverted create/update verbs: v1 uses PUT for create and POST
// for update, v2 follows REST conventions (POST create, PUT/PATCH update).
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
	v2 := isV2Path(route.Path)

	switch route.Method {
	case http.MethodGet:
		if endsWithParam {
			return "read_one", detail
		}
		return "read_all", detail
	case http.MethodPut:
		if v2 {
			// v2: PUT replaces an existing resource → update.
			return "update", detail
		}
		// v1: PUT is used for creating resources.
		return "create", detail
	case http.MethodPost:
		if v2 {
			// v2: POST creates a new resource on the collection.
			return "create", detail
		}
		// v1: POST is used for updating resources.
		return "update", detail
	case http.MethodPatch:
		// Both v1 and v2 use PATCH for partial updates; v2 introduces
		// PATCH via Huma's AutoPatch synthesizer.
		return "update", detail
	case http.MethodDelete:
		return "delete", detail
	}

	return "", detail
}

func ensureAPITokenRoutesGroup(target map[string]APITokenRoute, group string) {
	if _, has := target[group]; !has {
		target[group] = make(APITokenRoute)
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

	// Check if this is a bulk variant of a known CRUD resource
	if strings.HasSuffix(routeGroupName, "_bulk") {
		parent := strings.TrimSuffix(routeGroupName, "_bulk")
		if crudResources[parent] {
			return true
		}
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
//
// v1 and v2 routes are keyed identically — both write into a map addressed by
// the same (group, permission) name derived from the path without the
// /api/vN prefix. v2 routes land in apiTokenRoutesV2 so the frontend token
// UI (which reads apiTokenRoutes) keeps showing the stable v1-named groups
// while CanDoAPIRoute consults both tables when authorising a request.
func CollectRoutesForAPITokenUsage(route echo.RouteInfo, requiresJWT bool) {

	if route.Method == "echo_route_not_found" {
		return
	}

	if !requiresJWT {
		return
	}

	routeGroupName, routeParts := getRouteGroupName(route.Path)

	if routeGroupName == "token_test" ||
		routeGroupName == "subscriptions" ||
		routeGroupName == "tokens" ||
		routeGroupName == "*" ||
		strings.HasPrefix(routeGroupName, "user_") {
		return
	}

	target := apiTokenRoutes
	if isV2Path(route.Path) {
		target = apiTokenRoutesV2
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
				ensureAPITokenRoutesGroup(target, "notifications")

				target["notifications"]["mark_all_as_read"] = routeDetail
				return
			}

			ensureAPITokenRoutesGroup(target, "other")

			_, exists := target["other"][routeGroupName]
			if exists {
				routeGroupName += "_" + strings.ToLower(route.Method)
			}
			target["other"][routeGroupName] = routeDetail
			return
		}

		subkey := strings.Join(routeParts[1:], "_")

		if _, has := target[routeParts[0]]; !has {
			target[routeParts[0]] = make(APITokenRoute)
		}

		if _, has := target[routeParts[0]][subkey]; has {
			subkey += "_" + strings.ToLower(route.Method)
		}

		target[routeParts[0]][subkey] = routeDetail

		return
	}

	if strings.HasSuffix(routeGroupName, "_bulk") {
		parent := strings.TrimSuffix(routeGroupName, "_bulk")
		ensureAPITokenRoutesGroup(target, parent)

		method, routeDetail := getRouteDetail(route)
		target[parent][method+"_bulk"] = routeDetail
		return
	}

	_, has := target[routeGroupName]
	if !has {
		target[routeGroupName] = make(APITokenRoute)
	}

	method, routeDetail := getRouteDetail(route)
	if method != "" {
		target[routeGroupName][method] = routeDetail
	}

	// Handle task attachments specially - they use custom handlers not WebHandler
	if routeGroupName == "tasks_attachments" {
		// PUT is upload (create), GET with :attachment param is download (read_one)
		if route.Method == http.MethodPut {
			target[routeGroupName]["create"] = &RouteDetail{
				Path:   route.Path,
				Method: route.Method,
			}
		}
		if route.Method == http.MethodGet && strings.HasSuffix(route.Path, ":attachment") {
			target[routeGroupName]["read_one"] = &RouteDetail{
				Path:   route.Path,
				Method: route.Method,
			}
		}
	}

}

// GetAPITokenRoutes exposes the registered scoped-token routes so tests
// and the /api/v1/routes handler share a single source of truth.
func GetAPITokenRoutes() map[string]APITokenRoute {
	return apiTokenRoutes
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
	return c.JSON(http.StatusOK, GetAPITokenRoutes())
}

// CanDoAPIRoute checks if a token is allowed to use the current api route.
//
// Each permission is authoritative: the request is allowed only if the
// stored (Path, Method) for that permission matches exactly. This closes
// GHSA-v479-vf79-mg83 and the wider method/sub-resource confusion it
// enabled. The one exception is the tasks.read_all quirk handled below.
//
// Tokens are granted by (group, permission) name (e.g. labels.read_one),
// so a single permission can legitimately match both the v1 and v2 routes
// for the same resource. We consult apiTokenRoutes for v1 and the
// apiTokenRoutesV2 shadow map for v2.
func CanDoAPIRoute(c *echo.Context, token *APIToken) (can bool) {
	path := c.Path()
	if path == "" {
		// c.Path() is empty during testing, but returns the path which
		// the route used during registration which is what we need.
		path = c.Request().URL.Path
	}
	method := c.Request().Method

	for group, perms := range token.APIPermissions {
		tables := []APITokenRoute{apiTokenRoutes[group], apiTokenRoutesV2[group]}
		for _, routes := range tables {
			if routes == nil {
				continue
			}
			for _, p := range perms {
				rd := routes[p]
				if rd == nil {
					continue
				}
				if rd.Method == method && rd.Path == path {
					return true
				}
				// Two list endpoints share tasks.read_all but only one
				// survives collection, so allow either explicitly.
				if group == "tasks" && p == "read_all" && method == http.MethodGet &&
					(path == "/api/v1/tasks" || path == "/api/v1/projects/:project/tasks") {
					return true
				}
			}
		}
	}

	log.Debugf("[auth] Token %d tried to use route %s %s which is not covered by its permissions %v",
		token.ID, method, path, token.APIPermissions)

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
