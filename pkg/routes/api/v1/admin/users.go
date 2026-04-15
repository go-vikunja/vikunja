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
	"code.vikunja.io/api/pkg/user"
	"github.com/labstack/echo/v5"
)

// AdminUser re-exposes fields that the default User JSON view hides.
type AdminUser struct {
	*user.User
	IsAdmin bool        `json:"is_admin"`
	Status  user.Status `json:"status"`
}

// ListUsers returns paginated users for the admin panel with optional search.
// @Summary List users (admin)
// @Description Paginated list of all users on the instance. Supports search by username/email. Exposes fields hidden from the normal user API (is_admin, status).
// @tags admin
// @Produce json
// @Security JWTKeyAuth
// @Param s query string false "Search string matched against username and email."
// @Param page query int false "Page number, defaults to 1."
// @Param per_page query int false "Items per page, defaults to the service setting."
// @Success 200 {array} admin.AdminUser
// @Failure 404 {object} web.HTTPError
// @Router /admin/users [get]
func ListUsers(c *echo.Context) error {
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

	query := c.QueryParam("s")

	var users []*user.User
	sess := s.Limit(perPage, (page-1)*perPage).OrderBy("id ASC")
	if query != "" {
		q := "%" + query + "%"
		sess = sess.Where("username LIKE ? OR email LIKE ?", q, q)
	}
	if err := sess.Find(&users); err != nil {
		return err
	}

	out := make([]*AdminUser, 0, len(users))
	for _, u := range users {
		out = append(out, &AdminUser{User: u, IsAdmin: u.IsAdmin, Status: u.Status})
	}

	return c.JSON(http.StatusOK, out)
}
