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
	"github.com/labstack/echo/v5"
)

type OwnerPatch struct {
	OwnerID int64 `json:"owner_id"`
}

// PatchProjectOwner reassigns a project's owner.
// @Summary Reassign project owner (admin)
// @Description Reassign a project's owner. The existing update endpoint doesn't allow owner changes — this is the admin-only escape hatch.
// @tags admin
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "Project ID"
// @Param body body admin.OwnerPatch true "New owner"
// @Success 200 {object} models.Project
// @Failure 400 {object} web.HTTPError
// @Failure 404 {object} web.HTTPError
// @Router /admin/projects/{id}/owner [patch]
func PatchProjectOwner(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		return models.ErrProjectDoesNotExist{ID: id}
	}

	body := &OwnerPatch{}
	if err := c.Bind(body); err != nil || body.OwnerID < 1 {
		return models.ErrInvalidData{Message: "invalid body"}
	}

	s := db.NewSession()
	defer s.Close()

	p, err := models.ReassignProjectOwner(s, id, body.OwnerID)
	if err != nil {
		return err
	}
	if err := s.Commit(); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, p)
}
