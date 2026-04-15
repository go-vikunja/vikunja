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

package admin

import (
	"net/http"
	"strconv"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"github.com/labstack/echo/v5"
)

// ListProjects returns all projects on the instance, paginated.
// @Summary List projects (admin)
// @Description Paginated list of every project on the instance, regardless of ownership.
// @tags admin
// @Produce json
// @Security JWTKeyAuth
// @Param page query int false "Page number, defaults to 1."
// @Param per_page query int false "Items per page, defaults to the service setting."
// @Success 200 {array} models.Project
// @Failure 404 {object} web.HTTPError
// @Router /admin/projects [get]
func ListProjects(c *echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
	if perPage < 1 {
		perPage = config.ServiceMaxItemsPerPage.GetInt()
	}

	var projects []*models.Project
	if err := s.Limit(perPage, (page-1)*perPage).OrderBy("id DESC").Find(&projects); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, projects)
}
