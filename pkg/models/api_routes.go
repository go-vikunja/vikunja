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
	"reflect"
	"regexp"
	"runtime"
	"strings"

	"code.vikunja.io/api/pkg/log"

	"github.com/labstack/echo/v4"
)

var (
	apiTokenRoutes  = map[string]map[string]APITokenRoute{}
	apiPrefixRegex  = regexp.MustCompile(`^/api/v[0-9]+/`)
	apiVersionRegex = regexp.MustCompile(`^/api/(v[0-9]+)/`)
)

func init() {
	apiTokenRoutes = make(map[string]map[string]APITokenRoute)
}

type APITokenRoute map[string]*RouteDetail

type RouteDetail struct {
	Path   string `json:"path"`
	Method string `json:"method"`
}

func getRouteAPIVersion(path string) string {
	matches := apiVersionRegex.FindStringSubmatch(path)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func getRouteGroupName(path string) (finalName string, filteredParts []string) {
	path = apiPrefixRegex.ReplaceAllString(path, "")
	parts := strings.Split(path, "/")
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

func getRouteDetail(route echo.Route) (method string, detail *RouteDetail) {
	if strings.Contains(route.Name, "CreateWeb") {
		return "create", &RouteDetail{
			Path:   route.Path,
			Method: route.Method,
		}
	}
	if strings.Contains(route.Name, "ReadOneWeb") {
		return "read_one", &RouteDetail{
			Path:   route.Path,
			Method: route.Method,
		}
	}
	if strings.Contains(route.Name, "ReadAllWeb") {
		return "read_all", &RouteDetail{
			Path:   route.Path,
			Method: route.Method,
		}
	}
	if strings.Contains(route.Name, "UpdateWeb") {
		return "update", &RouteDetail{
			Path:   route.Path,
			Method: route.Method,
		}
	}
	if strings.Contains(route.Name, "DeleteWeb") {
		return "delete", &RouteDetail{
			Path:   route.Path,
			Method: route.Method,
		}
	}

	return "", &RouteDetail{
		Path:   route.Path,
		Method: route.Method,
	}
}

func ensureAPITokenRoutesGroup(version, group string) {
	if _, has := apiTokenRoutes[version]; !has {
		apiTokenRoutes[version] = make(map[string]APITokenRoute)
	}
	if _, has := apiTokenRoutes[version][group]; !has {
		apiTokenRoutes[version][group] = make(APITokenRoute)
	}
}

// CollectRoutesForAPITokenUsage gets called for every added APITokenRoute and builds a list of all routes we can use for the api tokens.
func CollectRoutesForAPITokenUsage(route echo.Route, middlewares []echo.MiddlewareFunc) {

	if route.Method == "echo_route_not_found" {
		return
	}

	seenJWT := false
	for _, middleware := range middlewares {
		if strings.Contains(runtime.FuncForPC(reflect.ValueOf(middleware).Pointer()).Name(), "github.com/labstack/echo-jwt/") {
			seenJWT = true
		}
	}

	if !seenJWT {
		return
	}

	routeGroupName, routeParts := getRouteGroupName(route.Path)
	apiVersion := getRouteAPIVersion(route.Path)
	if apiVersion == "" {
		// No api version, no tokens
		return
	}

	if routeGroupName == "tokenTest" ||
		routeGroupName == "subscriptions" ||
		routeGroupName == "tokens" ||
		routeGroupName == "*" ||
		strings.HasPrefix(routeGroupName, "user_") {
		return
	}

	if apiVersion == "v2" {
		method := ""
		switch route.Method {
		case http.MethodPost:
			method = "create"
		case http.MethodGet:
			method = "read_all"
			if strings.Contains(route.Path, "/:") {
				method = "read_one"
			}
		case http.MethodPut:
			method = "update"
		case http.MethodDelete:
			method = "delete"
		}
		if method == "" {
			return
		}

		ensureAPITokenRoutesGroup(apiVersion, routeGroupName)
		routeDetail := &RouteDetail{
			Path:   route.Path,
			Method: route.Method,
		}
		apiTokenRoutes[apiVersion][routeGroupName][method] = routeDetail
		return
	}

	if !strings.Contains(route.Name, "(*WebHandler)") && !strings.Contains(route.Name, "Attachment") {
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
				ensureAPITokenRoutesGroup(apiVersion, "notifications")

				apiTokenRoutes[apiVersion]["notifications"]["mark_all_as_read"] = routeDetail
				return
			}

			ensureAPITokenRoutesGroup(apiVersion, "other")

			_, exists := apiTokenRoutes[apiVersion]["other"][routeGroupName]
			if exists {
				routeGroupName += "_" + strings.ToLower(route.Method)
			}
			apiTokenRoutes[apiVersion]["other"][routeGroupName] = routeDetail
			return
		}

		subkey := strings.Join(routeParts[1:], "_")

		ensureAPITokenRoutesGroup(apiVersion, routeParts[0])

		if _, has := apiTokenRoutes[apiVersion][routeParts[0]][subkey]; has {
			subkey += "_" + strings.ToLower(route.Method)
		}

		apiTokenRoutes[apiVersion][routeParts[0]][subkey] = routeDetail

		return
	}

	if strings.HasSuffix(routeGroupName, "_bulk") {
		parent := strings.TrimSuffix(routeGroupName, "_bulk")
		ensureAPITokenRoutesGroup(apiVersion, parent)

		method, routeDetail := getRouteDetail(route)
		apiTokenRoutes[apiVersion][parent][method+"_bulk"] = routeDetail
		return
	}

	ensureAPITokenRoutesGroup(apiVersion, routeGroupName)

	method, routeDetail := getRouteDetail(route)
	if method != "" {
		apiTokenRoutes[apiVersion][routeGroupName][method] = routeDetail
	}

	if routeGroupName == "tasks_attachments" {
		if strings.Contains(route.Name, "UploadTaskAttachment") {
			apiTokenRoutes[apiVersion][routeGroupName]["create"] = &RouteDetail{
				Path:   route.Path,
				Method: route.Method,
			}
		}
		if strings.Contains(route.Name, "GetTaskAttachment") {
			apiTokenRoutes[apiVersion][routeGroupName]["read_one"] = &RouteDetail{
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
func GetAvailableAPIRoutesForToken(c echo.Context) error {
	// We merge all versions into one map to make it easier for the frontend
	// to display the routes.
	mergedRoutes := make(map[string]APITokenRoute)
	for version, groups := range apiTokenRoutes {
		for group, routes := range groups {
			mergedRoutes[version+"_"+group] = routes
		}
	}
	return c.JSON(http.StatusOK, mergedRoutes)
}

// CanDoAPIRoute checks if a token is allowed to use the current api route
func CanDoAPIRoute(c echo.Context, token *APIToken) (can bool) {
	path := c.Path()
	if path == "" {
		// c.Path() is empty during testing, but returns the path which
		// the route used during registration which is what we need.
		path = c.Request().URL.Path
	}

	routeGroupName, routeParts := getRouteGroupName(path)
	apiVersion := getRouteAPIVersion(path)

	routeGroupName = strings.TrimSuffix(routeGroupName, "_bulk")

	if routeGroupName == "user" ||
		routeGroupName == "users" ||
		routeGroupName == "routes" {
		routeGroupName = "other"
	}

	// The frontend sends permissions with the version prefixed, like "v1_projects"
	group, hasGroup := token.APIPermissions[apiVersion+"_"+routeGroupName]
	if !hasGroup {
		group, hasGroup = token.APIPermissions[apiVersion+"_"+routeParts[0]]
		if !hasGroup {
			// For backwards compatibility, we also check without the version prefix
			group, hasGroup = token.APIPermissions[routeGroupName]
			if !hasGroup {
				group, hasGroup = token.APIPermissions[routeParts[0]]
				if !hasGroup {
					return false
				}
			}
		}
	}

	var route string
	routes, has := apiTokenRoutes[apiVersion][routeGroupName]
	if !has {
		routes, has = apiTokenRoutes[apiVersion][routeParts[0]]
		if !has {
			return false
		}
		route = strings.Join(routeParts[1:], "_")
	}

	// The tasks read_all route is available as /:project/tasks and /tasks/all - therefore we need this workaround here.
	if routeGroupName == "tasks" && path == "/api/v1/projects/:project/tasks" && c.Request().Method == http.MethodGet {
		route = "read_all"
	}

	// We need to remove the /api/v... prefix from the path to compare it with the stored path.
	pathWithoutPrefix := apiPrefixRegex.ReplaceAllString(path, "")
	for _, p := range group {
		if route == "" && routes[p] != nil {
			// We only check the path without the version prefix, because the version is already checked.
			routePathWithoutPrefix := apiPrefixRegex.ReplaceAllString(routes[p].Path, "")
			if routePathWithoutPrefix == pathWithoutPrefix && routes[p].Method == c.Request().Method {
				return true
			}
		}
		if route != "" && p == route {
			return true
		}
	}

	log.Debugf("[auth] Token %d tried to use route %s which requires permission %s but has only %v", token.ID, path, route, token.APIPermissions)

	return false
}

func PermissionsAreValid(permissions APIPermissions) (err error) {
	for key, methods := range permissions {
		parts := strings.SplitN(key, "_", 2)
		if len(parts) != 2 {
			return &ErrInvalidAPITokenPermission{
				Group: key,
			}
		}
		version, group := parts[0], parts[1]

		versionedRoutes, has := apiTokenRoutes[version]
		if !has {
			return &ErrInvalidAPITokenPermission{
				Group: key,
			}
		}

		routes, has := versionedRoutes[group]
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
