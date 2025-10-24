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
	"net/http"
	"strconv"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/services"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web/handler"
	"github.com/labstack/echo/v4"
	"xorm.io/xorm"
)

// SavedFilterRoutes defines all saved filter API routes with their explicit permission scopes.
// This enables API tokens to be scoped for saved filter management operations.
var SavedFilterRoutes = []APIRoute{
	{Method: "GET", Path: "/filters", Handler: handler.WithDBAndUser(GetAllSavedFilters, true), PermissionScope: "read_all"},
	{Method: "GET", Path: "/filters/:filter", Handler: handler.WithDBAndUser(GetSavedFilter, true), PermissionScope: "read_one"},
	{Method: "PUT", Path: "/filters", Handler: handler.WithDBAndUser(CreateSavedFilter, true), PermissionScope: "create"},
	{Method: "POST", Path: "/filters/:filter", Handler: handler.WithDBAndUser(UpdateSavedFilter, true), PermissionScope: "update"},
	{Method: "DELETE", Path: "/filters/:filter", Handler: handler.WithDBAndUser(DeleteSavedFilter, true), PermissionScope: "delete"},
}

// RegisterSavedFilters registers the saved filter routes.
func RegisterSavedFilters(a *echo.Group) {
	registerRoutes(a, SavedFilterRoutes)
}

func GetAllSavedFilters(s *xorm.Session, u *user.User, c echo.Context) error {
	search := c.QueryParam("search")

	sfService := services.NewSavedFilterService(s.Engine())
	filters, err := sfService.GetAllForUser(s, u, search)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, filters)
}

func GetSavedFilter(s *xorm.Session, u *user.User, c echo.Context) error {
	filterID, err := strconv.ParseInt(c.Param("filter"), 10, 64)
	if err != nil {
		return err
	}

	sfService := services.NewSavedFilterService(s.Engine())
	filter, err := sfService.Get(s, filterID, u)
	if err != nil {
		return err
	}

	// Saved filters can only be accessed by their owner, so permission is always Admin
	c.Response().Header().Set("x-max-permission", "2") // PermissionAdmin = 2
	c.Response().Header().Set("Access-Control-Expose-Headers", "x-max-permission")

	return c.JSON(http.StatusOK, filter)
}

func CreateSavedFilter(s *xorm.Session, u *user.User, c echo.Context) error {
	sf := &models.SavedFilter{}
	if err := c.Bind(sf); err != nil {
		return err
	}

	sfService := services.NewSavedFilterService(s.Engine())
	if err := sfService.Create(s, sf, u); err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, sf)
}

func UpdateSavedFilter(s *xorm.Session, u *user.User, c echo.Context) error {
	filterID, err := strconv.ParseInt(c.Param("filter"), 10, 64)
	if err != nil {
		return err
	}

	sf := &models.SavedFilter{}
	if err := c.Bind(sf); err != nil {
		return err
	}
	sf.ID = filterID

	sfService := services.NewSavedFilterService(s.Engine())
	if err := sfService.Update(s, sf, u); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, sf)
}

func DeleteSavedFilter(s *xorm.Session, u *user.User, c echo.Context) error {
	filterID, err := strconv.ParseInt(c.Param("filter"), 10, 64)
	if err != nil {
		return err
	}

	sfService := services.NewSavedFilterService(s.Engine())
	if err := sfService.Delete(s, filterID, u); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}
