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

type StatusPatch struct {
	Status user.Status `json:"status"`
}

// PatchStatus sets a user's status.
// @Summary Set a user's status (admin)
// @Description Change a user's status without requiring them to log in.
// @tags admin
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "User ID"
// @Param body body admin.StatusPatch true "Status"
// @Success 200 {object} admin.User
// @Failure 400 {object} web.HTTPError
// @Failure 404 {object} web.HTTPError
// @Router /admin/users/{id}/status [patch]
func PatchStatus(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		return user.ErrUserDoesNotExist{UserID: id}
	}
	body := &StatusPatch{}
	if err := c.Bind(body); err != nil {
		return models.ErrInvalidData{Message: "invalid body"}
	}
	if body.Status < user.StatusActive || body.Status > user.StatusAccountLocked {
		return models.ErrInvalidData{Message: "invalid status"}
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

	// Disabling/locking an admin is equivalent to demoting them.
	if target.IsAdmin && (body.Status == user.StatusDisabled || body.Status == user.StatusAccountLocked) {
		if err := user.GuardLastAdmin(s, target); err != nil {
			_ = s.Rollback()
			return err
		}
	}

	if err := user.SetUserStatus(s, target, body.Status); err != nil {
		_ = s.Rollback()
		return err
	}
	if err := s.Commit(); err != nil {
		return err
	}

	// Refresh locally; GetUserByID refuses disabled accounts.
	target.Status = body.Status
	providers, err := openid.GetAllProviders()
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, newAdminUser(target, providers))
}

// DeleteUser removes a user either immediately or through the self-deletion flow.
// @Summary Delete a user (admin)
// @Description Delete a user. With mode=now the user is removed immediately. With mode=scheduled (the default, matching the CLI) the user receives a confirmation email and is scheduled for deletion just like a self-initiated account deletion.
// @tags admin
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "User ID"
// @Param mode query string false "Deletion mode: 'now' for immediate deletion, 'scheduled' (default) to trigger the email-confirmation self-deletion flow."
// @Success 204
// @Failure 400 {object} web.HTTPError
// @Failure 404 {object} web.HTTPError
// @Router /admin/users/{id} [delete]
func DeleteUser(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		return user.ErrUserDoesNotExist{UserID: id}
	}

	mode := c.QueryParam("mode")
	if mode == "" {
		mode = "scheduled"
	}
	if mode != "now" && mode != "scheduled" {
		return models.ErrInvalidData{Message: "invalid mode, expected 'now' or 'scheduled'"}
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

	if mode == "now" {
		if err := user.GuardLastAdmin(s, target); err != nil {
			_ = s.Rollback()
			return err
		}
		if err := models.DeleteUser(s, target); err != nil {
			_ = s.Rollback()
			return err
		}
	} else {
		if err := user.RequestDeletion(s, target); err != nil {
			_ = s.Rollback()
			return err
		}
	}

	if err := s.Commit(); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}
