// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2021 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package v1

import (
	"net/http"
	"time"

	"code.vikunja.io/api/pkg/user"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"

	"code.vikunja.io/api/pkg/db"

	"code.vikunja.io/web/handler"
	"github.com/labstack/echo/v4"
)

type userWithSettings struct {
	user.User
	Settings            *UserSettings `json:"settings"`
	DeletionScheduledAt time.Time     `json:"deletion_scheduled_at"`
	IsLocalUser         bool          `json:"is_local_user"`
}

// UserShow gets all informations about the current user
// @Summary Get user information
// @Description Returns the current user object.
// @tags user
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Success 200 {object} user.User
// @Failure 404 {object} web.HTTPError "User does not exist."
// @Failure 500 {object} models.Message "Internal server error."
// @Router /user [get]
func UserShow(c echo.Context) error {
	a, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Error getting current user.")
	}

	s := db.NewSession()
	defer s.Close()

	u, err := models.GetUserOrLinkShareUser(s, a)
	if err != nil {
		return handler.HandleHTTPError(err, c)
	}

	us := &userWithSettings{
		User: *u,
		Settings: &UserSettings{
			Name:                         u.Name,
			EmailRemindersEnabled:        u.EmailRemindersEnabled,
			DiscoverableByName:           u.DiscoverableByName,
			DiscoverableByEmail:          u.DiscoverableByEmail,
			OverdueTasksRemindersEnabled: u.OverdueTasksRemindersEnabled,
			DefaultListID:                u.DefaultListID,
			WeekStart:                    u.WeekStart,
			Language:                     u.Language,
		},
		DeletionScheduledAt: u.DeletionScheduledAt,
		IsLocalUser:         u.Issuer == user.IssuerLocal,
	}

	return c.JSON(http.StatusOK, us)
}
