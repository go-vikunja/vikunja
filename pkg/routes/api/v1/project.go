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

// RegisterProjects registers all project routes
func RegisterProjects(a *echo.Group) {
	a.GET("/projects", GetAllProjects)
	a.GET("/projects/:project", GetProject)
	a.POST("/projects/:project", UpdateProject)
	a.DELETE("/projects/:project", DeleteProject)
	a.PUT("/projects", CreateProject)
}

func GetAllProjects(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	auth, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return handler.HandleHTTPError(err)
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
	p := new(models.Project)
	_ = c.Bind(p)

	ps := services.NewProjectService()
	projects, resultCount, total, err := ps.GetAll(s, auth, p, search, page, perPage)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	if projects, ok := projects.([]*models.Project); ok {
		for _, p := range projects {
			services.AddProjectLinks(c, p)
		}
	}

	var numberOfPages = math.Ceil(float64(total) / float64(perPage))
	if page < 0 {
		numberOfPages = 1
	}
	if resultCount == 0 {
		numberOfPages = 0
	}

	c.Response().Header().Set("x-pagination-total-pages", strconv.FormatFloat(numberOfPages, 'f', 0, 64))
	c.Response().Header().Set("x-pagination-result-count", strconv.Itoa(resultCount))
	c.Response().Header().Set("Access-Control-Expose-Headers", "x-pagination-total-pages, x-pagination-result-count")

	return c.JSON(http.StatusOK, projects)
}

func GetProject(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	auth, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	projectID, err := strconv.ParseInt(c.Param("project"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID").SetInternal(err)
	}

	ps := services.NewProjectService()
	p, err := ps.Get(s, projectID, auth)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	perm, err := ps.GetMaxPermission(s, p.ID, auth)
	if err != nil {
		return handler.HandleHTTPError(err)
	}
	c.Response().Header().Set("x-max-permission", strconv.Itoa(int(perm)))

	return c.JSON(http.StatusOK, p)
}

func UpdateProject(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	auth, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	projectID, err := strconv.ParseInt(c.Param("project"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID").SetInternal(err)
	}

	updatePayload := new(models.Project)
	if err := c.Bind(updatePayload); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project object provided.").SetInternal(err)
	}

	// Set the ID from the URL param, not the payload
	updatePayload.ID = projectID

	if err := c.Validate(updatePayload); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error()).SetInternal(err)
	}

	ps := services.NewProjectService()
	p, err := ps.Update(s, updatePayload, auth)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	if err := s.Commit(); err != nil {
		return handler.HandleHTTPError(err)
	}

	return c.JSON(http.StatusOK, p)
}

func DeleteProject(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	auth, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	projectID, err := strconv.ParseInt(c.Param("project"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID").SetInternal(err)
	}

	ps := services.NewProjectService()
	if err := ps.Delete(s, projectID, auth); err != nil {
		return handler.HandleHTTPError(err)
	}

	if err := s.Commit(); err != nil {
		return handler.HandleHTTPError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

func CreateProject(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	auth, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	p := new(models.Project)
	if err := c.Bind(p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project object provided.").SetInternal(err)
	}

	if err := c.Validate(p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error()).SetInternal(err)
	}

	ps := services.NewProjectService()
	p, err = ps.Create(s, p, auth)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	if err := s.Commit(); err != nil {
		return handler.HandleHTTPError(err)
	}

	return c.JSON(http.StatusCreated, p)
}
