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

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth/openid"
	"code.vikunja.io/api/pkg/user"
	"github.com/labstack/echo/v5"
)

type IsAdminPatch struct {
	// Pointer to distinguish "omitted" from false; an empty body would silently demote otherwise.
	IsAdmin *bool `json:"is_admin"`
}

// PatchAdmin toggles a user's instance-admin flag.
// @Summary Promote or demote a user (admin)
// @Description Toggle the instance-admin flag on a user. Demoting the last remaining admin is refused with 400.
// @tags admin
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "User ID"
// @Param body body admin.IsAdminPatch true "New admin value"
// @Success 200 {object} admin.User
// @Failure 400 {object} web.HTTPError
// @Failure 404 {object} web.HTTPError
// @Router /admin/users/{id}/admin [patch]
func PatchAdmin(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id < 1 {
		return user.ErrUserDoesNotExist{UserID: id}
	}

	body := &IsAdminPatch{}
	if err := c.Bind(body); err != nil {
		return models.ErrInvalidData{Message: "invalid body"}
	}
	if body.IsAdmin == nil {
		return models.ErrInvalidData{Message: "is_admin is required"}
	}

	s := db.NewSession()
	defer s.Close()

	target := &user.User{ID: id}
	has, err := s.Get(target)
	if err != nil {
		return err
	}
	if !has {
		return user.ErrUserDoesNotExist{UserID: id}
	}

	if !*body.IsAdmin {
		if err := user.GuardLastAdmin(s, target); err != nil {
			_ = s.Rollback()
			return err
		}
	}

	target.IsAdmin = *body.IsAdmin
	if _, err := s.ID(target.ID).Cols("is_admin").Update(target); err != nil {
		_ = s.Rollback()
		return err
	}
	if err := s.Commit(); err != nil {
		return err
	}

	providers, err := openid.GetAllProviders()
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, newAdminUser(target, providers))
}
