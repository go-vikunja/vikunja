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

// RegisterSavedFilters registers the saved filter routes.
func RegisterSavedFilters(a *echo.Group) {
	a.GET("/filters", handler.WithDBAndUser(GetAllSavedFilters, true))
	a.GET("/filters/:filter", handler.WithDBAndUser(GetSavedFilter, true))
	a.PUT("/filters", handler.WithDBAndUser(CreateSavedFilter, true))
	a.POST("/filters/:filter", handler.WithDBAndUser(UpdateSavedFilter, true))
	a.DELETE("/filters/:filter", handler.WithDBAndUser(DeleteSavedFilter, true))
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