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
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/labstack/echo/v4"
)

// ProjectTeams is a struct to handle project team routes
type ProjectTeams struct{}

// Get handles getting all teams in a project
func (pt *ProjectTeams) Get(c echo.Context) error {
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

	tp := &models.TeamProject{ProjectID: projectID}
	teams, resultCount, total, err := tp.ReadAll(s, auth, search, page, perPage)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	c.Response().Header().Set("x-pagination-total-pages", strconv.FormatFloat(math.Ceil(float64(total)/float64(perPage)), 'f', 0, 64))
	c.Response().Header().Set("x-pagination-result-count", strconv.Itoa(resultCount))
	c.Response().Header().Set("Access-Control-Expose-Headers", "x-pagination-total-pages, x-pagination-result-count")

	for _, t := range teams.([]*models.TeamWithPermission) {
		t.AddLinks(c)
	}

	return c.JSON(http.StatusOK, teams)
}

// Post adds a team to a project
func (pt *ProjectTeams) Post(c echo.Context) error {
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

	p := &models.Project{ID: projectID}
	_, perm, err := p.CanRead(s, auth)
	if err != nil {
		return handler.HandleHTTPError(err)
	}
	if perm < int(models.PermissionAdmin) {
		return echo.ErrForbidden
	}

	tp := new(models.TeamProject)
	if err := c.Bind(tp); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid team project object provided.").SetInternal(err)
	}
	tp.ProjectID = projectID

	if err := c.Validate(tp); err != nil {
		return err
	}

	if err := tp.Create(s, auth); err != nil {
		return handler.HandleHTTPError(err)
	}

	if err := s.Commit(); err != nil {
		return handler.HandleHTTPError(err)
	}

	return c.JSON(http.StatusCreated, tp)
}

// Put updates a team on a project
func (pt *ProjectTeams) Put(c echo.Context) error {
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

	teamID, err := strconv.ParseInt(c.Param("teamid"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid team ID").SetInternal(err)
	}

	p := &models.Project{ID: projectID}
	_, perm, err := p.CanRead(s, auth)
	if err != nil {
		return handler.HandleHTTPError(err)
	}
	if perm < int(models.PermissionAdmin) {
		return echo.ErrForbidden
	}

	tp := new(models.TeamProject)
	if err := c.Bind(tp); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid team project object provided.").SetInternal(err)
	}
	tp.ProjectID = projectID
	tp.TeamID = teamID

	if err := c.Validate(tp); err != nil {
		return err
	}

	if err := tp.Update(s, auth); err != nil {
		return handler.HandleHTTPError(err)
	}

	if err := s.Commit(); err != nil {
		return handler.HandleHTTPError(err)
	}

	return c.JSON(http.StatusOK, tp)
}

// Delete removes a team from a project
func (pt *ProjectTeams) Delete(c echo.Context) error {
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

	teamID, err := strconv.ParseInt(c.Param("teamid"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid team ID").SetInternal(err)
	}

	p := &models.Project{ID: projectID}
	_, perm, err := p.CanRead(s, auth)
	if err != nil {
		return handler.HandleHTTPError(err)
	}
	if perm < int(models.PermissionAdmin) {
		return echo.ErrForbidden
	}

	tp := &models.TeamProject{
		ProjectID: projectID,
		TeamID:    teamID,
	}

	if err := tp.Delete(s, auth); err != nil {
		return handler.HandleHTTPError(err)
	}

	if err := s.Commit(); err != nil {
		return handler.HandleHTTPError(err)
	}

	return c.NoContent(http.StatusNoContent)
}
