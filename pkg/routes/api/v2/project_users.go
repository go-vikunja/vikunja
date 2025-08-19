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
	"fmt"
	"math"
	"net/http"
	"strconv"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"
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

	puModel := &models.ProjectUser{ProjectID: projectID}

	users, resultCount, totalItems, err := puModel.ReadAll(s, auth, search, page, perPage)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	c.Response().Header().Set("x-pagination-total-pages", strconv.FormatFloat(math.Ceil(float64(totalItems)/float64(perPage)), 'f', 0, 64))
	c.Response().Header().Set("x-pagination-result-count", strconv.Itoa(resultCount))
	c.Response().Header().Set("Access-Control-Expose-Headers", "x-pagination-total-pages, x-pagination-result-count")

	usersResponse := make([]*ProjectUserResponse, len(users.([]*models.UserWithPermission)))
	for i, u := range users.([]*models.UserWithPermission) {
		usersResponse[i] = &ProjectUserResponse{
			UserWithPermission: u,
			Links: &ProjectUserLinks{
				Self:    fmt.Sprintf("/api/v2/users/%d", u.User.ID),
				Project: fmt.Sprintf("/api/v2/projects/%d", projectID),
			},
		}
	}

	return c.JSON(http.StatusOK, usersResponse)
}

type ProjectUserLinks struct {
	Self    string `json:"self"`
	Project string `json:"project"`
}

type ProjectUserResponse struct {
	*models.UserWithPermission
	Links *ProjectUserLinks `json:"_links"`
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

	p := &models.Project{ID: projectID}
	_, perm, err := p.CanRead(s, auth)
	if err != nil {
		return handler.HandleHTTPError(err)
	}
	if perm < int(models.PermissionAdmin) {
		return echo.ErrForbidden
	}

	puModel := new(models.ProjectUser)
	if err := c.Bind(puModel); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project user object provided.").SetInternal(err)
	}
	puModel.ProjectID = projectID

	if err := c.Validate(puModel); err != nil {
		return err
	}

	if err := puModel.Create(s, auth); err != nil {
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

	p := &models.Project{ID: projectID}
	_, perm, err := p.CanRead(s, auth)
	if err != nil {
		return handler.HandleHTTPError(err)
	}
	if perm < int(models.PermissionAdmin) {
		return echo.ErrForbidden
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

	if err := puModel.Update(s, auth); err != nil {
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

	p := &models.Project{ID: projectID}
	_, perm, err := p.CanRead(s, auth)
	if err != nil {
		return handler.HandleHTTPError(err)
	}
	if perm < int(models.PermissionAdmin) {
		return echo.ErrForbidden
	}

	puModel := &models.ProjectUser{
		ProjectID: projectID,
		UserID:    userID,
	}

	if err := puModel.Delete(s, auth); err != nil {
		return handler.HandleHTTPError(err)
	}

	if err := s.Commit(); err != nil {
		return handler.HandleHTTPError(err)
	}

	return c.NoContent(http.StatusNoContent)
}
