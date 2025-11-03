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
	"net/http"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/notifications"
	"github.com/labstack/echo/v4"
)

// MarkAllNotificationsAsRead marks all notifications of a user as read
// @Summary Mark all notifications of a user as read
// @tags sharing
// @Accept json
// @Produce json
// @Success 200 {object} models.Message "All notifications marked as read."
// @Failure 500 {object} models.Message "Internal error"
// @Router /notifications [post]
func MarkAllNotificationsAsRead(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	a, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return err
	}

	if _, is := a.(*models.LinkSharing); is {
		return echo.ErrForbidden
	}

	err = notifications.MarkAllNotificationsAsRead(s, a.GetID())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, models.Message{Message: "success"})
}
