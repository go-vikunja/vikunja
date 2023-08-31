// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2023 Vikunja and contributors. All rights reserved.
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

package routes

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
)

var apiTokenRoutes = map[string]*APITokenRoute{}

func init() {
	apiTokenRoutes = make(map[string]*APITokenRoute)
}

type APITokenRoute struct {
	Create  string `json:"create,omitempty"`
	ReadOne string `json:"read_one,omitempty"`
	ReadAll string `json:"read_all,omitempty"`
	Update  string `json:"update,omitempty"`
	Delete  string `json:"delete,omitempty"`
}

func getRouteGroupName(route echo.Route) string {
	parts := strings.Split(strings.TrimPrefix(route.Path, "/api/v1/"), "/")
	filteredParts := []string{}
	for _, part := range parts {
		if strings.HasPrefix(part, ":") {
			continue
		}

		filteredParts = append(filteredParts, part)
	}

	finalName := strings.Join(filteredParts, "_")
	switch finalName {
	case "tasks_all":
		return "tasks"
	default:
		return finalName
	}
}

// gets called for every added APITokenRoute and builds a list of all routes we can use for the api tokens.
func collectRoutesForAPITokenUsage(route echo.Route) {

	if !strings.Contains(route.Name, "(*WebHandler)") {
		return
	}

	routeGroupName := getRouteGroupName(route)

	if routeGroupName == "subscriptions" || routeGroupName == "notifications" || strings.HasSuffix(routeGroupName, "_bulk") {
		return
	}

	_, has := apiTokenRoutes[routeGroupName]
	if !has {
		apiTokenRoutes[routeGroupName] = &APITokenRoute{}
	}

	if strings.Contains(route.Name, "CreateWeb") {
		apiTokenRoutes[routeGroupName].Create = route.Path
	}
	if strings.Contains(route.Name, "ReadWeb") {
		apiTokenRoutes[routeGroupName].ReadOne = route.Path
	}
	if strings.Contains(route.Name, "ReadAllWeb") {
		apiTokenRoutes[routeGroupName].ReadAll = route.Path
	}
	if strings.Contains(route.Name, "UpdateWeb") {
		apiTokenRoutes[routeGroupName].Update = route.Path
	}
	if strings.Contains(route.Name, "DeleteWeb") {
		apiTokenRoutes[routeGroupName].Delete = route.Path
	}
}

// GetAvailableAPIRoutesForToken returns a list of all API routes which are available for token usage.
// @Summary Get a list of all token api routes
// @Description Returns a list of all API routes which are available to use with an api token, not a user login.
// @tags opi
// @Produce json
// @Security JWTKeyAuth
// @Success 200 {array} routes.APITokenRoute "The list of all routes."
// @Router /routes [get]
func GetAvailableAPIRoutesForToken(c echo.Context) error {
	return c.JSON(http.StatusOK, apiTokenRoutes)
}
