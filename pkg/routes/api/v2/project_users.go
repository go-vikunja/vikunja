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

package v2

import (
	"math"
	"net/http"
	"strconv"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/services"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/labstack/echo/v4"
)

// ProjectUsers is a struct to handle project user routes
type ProjectUsers struct{}

// Get returns all users for a project
func (pu *ProjectUsers) Get(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	auth, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID").SetInternal(err)
	}

	pageStr := c.QueryParam("page")
	if pageStr == "" {
		pageStr = "1"
	}
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid page number").SetInternal(err)
	}

	perPageStr := c.QueryParam("per_page")
	if perPageStr == "" {
		perPageStr = "20"
	}
	perPage, err := strconv.Atoi(perPageStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid per_page number").SetInternal(err)
	}
	search := c.QueryParam("s")

	pus := services.NewProjectUsersService()
	users, resultCount, totalItems, err := pus.Get(s, projectID, auth, search, page, perPage)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	c.Response().Header().Set("x-pagination-total-pages", strconv.FormatFloat(math.Ceil(float64(totalItems)/float64(perPage)), 'f', 0, 64))
	c.Response().Header().Set("x-pagination-result-count", strconv.Itoa(resultCount))
	c.Response().Header().Set("Access-Control-Expose-Headers", "x-pagination-total-pages, x-pagination-result-count")

	if users, ok := users.([]*models.UserWithPermission); ok {
		for _, u := range users {
			u.Links = models.Links{"self": {HREF: "/api/v2/users/" + strconv.FormatInt(u.User.ID, 10), Method: "GET"}}
		}
	}

	return c.JSON(http.StatusOK, users)
}

// Post adds a user to a project
func (pu *ProjectUsers) Post(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	auth, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID").SetInternal(err)
	}

	puModel := new(models.ProjectUser)
	if err := c.Bind(puModel); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project user object provided.").SetInternal(err)
	}
	puModel.ProjectID = projectID

	if err := c.Validate(puModel); err != nil {
		return err
	}

	pus := services.NewProjectUsersService()
	puModel, err = pus.Create(s, puModel, auth)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	if err := s.Commit(); err != nil {
		return handler.HandleHTTPError(err)
	}

	return c.JSON(http.StatusCreated, puModel)
}

// Put updates a user's permissions on a project
func (pu *ProjectUsers) Put(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	auth, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID").SetInternal(err)
	}

	userID, err := strconv.ParseInt(c.Param("userid"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID").SetInternal(err)
	}

	puModel := new(models.ProjectUser)
	if err := c.Bind(puModel); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project user object provided.").SetInternal(err)
	}
	puModel.ProjectID = projectID
	puModel.UserID = userID

	if err := c.Validate(puModel); err != nil {
		return err
	}

	pus := services.NewProjectUsersService()
	puModel, err = pus.Update(s, puModel, auth)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	if err := s.Commit(); err != nil {
		return handler.HandleHTTPError(err)
	}

	return c.JSON(http.StatusOK, puModel)
}

// Delete removes a user from a project
func (pu *ProjectUsers) Delete(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	auth, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID").SetInternal(err)
	}

	userID, err := strconv.ParseInt(c.Param("userid"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID").SetInternal(err)
	}

	pus := services.NewProjectUsersService()
	if err := pus.Delete(s, projectID, userID, auth); err != nil {
		return handler.HandleHTTPError(err)
	}

	if err := s.Commit(); err != nil {
		return handler.HandleHTTPError(err)
	}

	return c.NoContent(http.StatusNoContent)
}
