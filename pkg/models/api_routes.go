// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

var apiTokenRoutes = map[string]*APITokenRoute{}

func init() {
	apiTokenRoutes = make(map[string]*APITokenRoute)
}

type APITokenRoute struct {
	Create  *RouteDetail `json:"create,omitempty"`
	ReadOne *RouteDetail `json:"read_one,omitempty"`
	ReadAll *RouteDetail `json:"read_all,omitempty"`
	Update  *RouteDetail `json:"update,omitempty"`
	Delete  *RouteDetail `json:"delete,omitempty"`
}

type RouteDetail struct {
	Path   string `json:"path"`
	Method string `json:"method"`
}

func getRouteGroupName(path string) string {
	parts := strings.Split(strings.TrimPrefix(path, "/api/v1/"), "/")
	filteredParts := []string{}
	for _, part := range parts {
		if strings.HasPrefix(part, ":") {
			continue
		}

		filteredParts = append(filteredParts, part)
	}

	finalName := strings.Join(filteredParts, "_")
	switch finalName {
	case "projects_tasks":
		fallthrough
	case "tasks_all":
		return "tasks"
	default:
		return finalName
	}
}

// CollectRoutesForAPITokenUsage gets called for every added APITokenRoute and builds a list of all routes we can use for the api tokens.
func CollectRoutesForAPITokenUsage(route echo.Route) {

	if !strings.Contains(route.Name, "(*WebHandler)") {
		return
	}

	routeGroupName := getRouteGroupName(route.Path)

	if routeGroupName == "subscriptions" ||
		routeGroupName == "tokens" ||
		strings.HasSuffix(routeGroupName, "_bulk") {
		return
	}

	_, has := apiTokenRoutes[routeGroupName]
	if !has {
		apiTokenRoutes[routeGroupName] = &APITokenRoute{}
	}

	if strings.Contains(route.Name, "CreateWeb") {
		apiTokenRoutes[routeGroupName].Create = &RouteDetail{
			Path:   route.Path,
			Method: route.Method,
		}
	}
	if strings.Contains(route.Name, "ReadOneWeb") {
		apiTokenRoutes[routeGroupName].ReadOne = &RouteDetail{
			Path:   route.Path,
			Method: route.Method,
		}
	}
	if strings.Contains(route.Name, "ReadAllWeb") {
		apiTokenRoutes[routeGroupName].ReadAll = &RouteDetail{
			Path:   route.Path,
			Method: route.Method,
		}
	}
	if strings.Contains(route.Name, "UpdateWeb") {
		apiTokenRoutes[routeGroupName].Update = &RouteDetail{
			Path:   route.Path,
			Method: route.Method,
		}
	}
	if strings.Contains(route.Name, "DeleteWeb") {
		apiTokenRoutes[routeGroupName].Delete = &RouteDetail{
			Path:   route.Path,
			Method: route.Method,
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
	return c.JSON(http.StatusOK, apiTokenRoutes)
}

// CanDoAPIRoute checks if a token is allowed to use the current api route
func CanDoAPIRoute(c echo.Context, token *APIToken) (can bool) {
	path := c.Path()
	if path == "" {
		// c.Path() is empty during testing, but returns the path which the route used during registration
		// which is what we need.
		path = c.Request().URL.Path
	}

	routeGroupName := getRouteGroupName(path)

	group, hasGroup := token.Permissions[routeGroupName]
	if !hasGroup {
		return false
	}

	var route string
	routes, has := apiTokenRoutes[routeGroupName]
	if !has {
		return false
	}

	if routes.Create != nil && routes.Create.Path == path && routes.Create.Method == c.Request().Method {
		route = "create"
	}
	if routes.ReadOne != nil && routes.ReadOne.Path == path && routes.ReadOne.Method == c.Request().Method {
		route = "read_one"
	}
	if routes.ReadAll != nil && routes.ReadAll.Path == path && routes.ReadAll.Method == c.Request().Method {
		route = "read_all"
	}
	if routes.Update != nil && routes.Update.Path == path && routes.Update.Method == c.Request().Method {
		route = "update"
	}
	if routes.Delete != nil && routes.Delete.Path == path && routes.Delete.Method == c.Request().Method {
		route = "delete"
	}

	for _, p := range group {
		if p == route {
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
			if method == "create" && routes.Create == nil {
				return &ErrInvalidAPITokenPermission{
					Group:      key,
					Permission: method,
				}
			}
			if method == "read_one" && routes.ReadOne == nil {
				return &ErrInvalidAPITokenPermission{
					Group:      key,
					Permission: method,
				}
			}
			if method == "read_all" && routes.ReadAll == nil {
				return &ErrInvalidAPITokenPermission{
					Group:      key,
					Permission: method,
				}
			}
			if method == "update" && routes.Update == nil {
				return &ErrInvalidAPITokenPermission{
					Group:      key,
					Permission: method,
				}
			}
			if method == "delete" && routes.Delete == nil {
				return &ErrInvalidAPITokenPermission{
					Group:      key,
					Permission: method,
				}
			}
		}
	}

	return nil
}
