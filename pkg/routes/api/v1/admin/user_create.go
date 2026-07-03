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
	"errors"
	"net/http"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth/openid"
	"code.vikunja.io/api/pkg/routes/api/shared"

	"github.com/labstack/echo/v5"
)

// CreateUser provisions a new account on behalf of an instance admin.
// @Summary Create a user (admin)
// @Description Create a new local user account. Respects the admin-only fields `is_admin` and `skip_email_confirm`. The public registration toggle is bypassed.
// @tags admin
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param body body models.CreateUserBody true "The user to create"
// @Success 200 {object} shared.AdminUser
// @Failure 400 {object} web.HTTPError
// @Router /admin/users [post]
func CreateUser(c *echo.Context) error {
	body := &models.CreateUserBody{}
	if err := c.Bind(body); err != nil {
		return c.JSON(http.StatusBadRequest, models.Message{Message: "No or invalid user model provided."})
	}
	if err := c.Validate(body); err != nil {
		e := models.ValidationHTTPError{}
		if is := errors.As(err, &e); is {
			return c.JSON(e.HTTPCode, e)
		}
		return err
	}

	s := db.NewSession()
	defer s.Close()

	newUser, err := models.CreateUserAsAdmin(s, body)
	if err != nil {
		_ = s.Rollback()
		events.CleanupPending(s)
		return err
	}
	// CreateUserAsAdmin commits internally; the events RegisterUser queued on s
	// (user.created) still need to be dispatched here.
	events.DispatchPending(c.Request().Context(), s)

	providers, err := openid.GetAllProviders()
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, shared.NewAdminUser(newUser, providers))
}
