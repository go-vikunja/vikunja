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
	"strconv"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/notifications"
	"code.vikunja.io/api/pkg/services"
	"github.com/labstack/echo/v4"
)

// RegisterNotifications registers the notification routes
func RegisterNotifications(a *echo.Group) {
	a.GET("/notifications", getNotifications)
	a.POST("/notifications/:notificationid", markNotificationAsRead)
	a.POST("/notifications", markAllNotificationsAsRead)
}

// getNotifications returns all notifications for the authenticated user
// @Summary Get all notifications for the current user
// @Description Returns an array with all notifications for the current user.
// @tags notifications
// @Accept json
// @Produce json
// @Param page query int false "The page number. Used for pagination. If not provided, the first page of results is returned."
// @Param per_page query int false "The maximum number of items per page. Note this parameter is limited by the configured maximum of items per page."
// @Security JWTKeyAuth
// @Success 200 {array} notifications.DatabaseNotification "The notifications"
// @Failure 403 {object} models.Message "Link shares cannot have notifications."
// @Failure 500 {object} models.Message "Internal error"
// @Router /notifications [get]
func getNotifications(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	a, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return err
	}

	if _, is := a.(*models.LinkSharing); is {
		return echo.ErrForbidden
	}

	// Parse pagination parameters
	page := c.QueryParam("page")
	pageNum := 1
	if page != "" {
		pageNum, err = strconv.Atoi(page)
		if err != nil || pageNum < 1 {
			pageNum = 1
		}
	}

	perPage := c.QueryParam("per_page")
	perPageNum := 50 // default
	if perPage != "" {
		perPageNum, err = strconv.Atoi(perPage)
		if err != nil || perPageNum < 1 {
			perPageNum = 50
		}
		if perPageNum > 100 {
			perPageNum = 100 // max limit
		}
	}

	// Calculate limit and offset
	limit := perPageNum
	offset := (pageNum - 1) * perPageNum

	service := services.NewNotificationsService(s)
	notifs, resultCount, total, err := service.GetNotificationsForUser(a.GetID(), limit, offset)
	if err != nil {
		return err
	}

	// Set pagination headers
	c.Response().Header().Set("x-pagination-total-pages", strconv.FormatInt((total+int64(perPageNum)-1)/int64(perPageNum), 10))
	c.Response().Header().Set("x-pagination-result-count", strconv.Itoa(resultCount))

	return c.JSON(http.StatusOK, notifs)
}

// markNotificationAsRead marks a notification as read or unread
// @Summary Mark a notification as (un-)read
// @Description Marks a notification as either read or unread. A user can only mark their own notifications as read.
// @tags notifications
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "Notification ID"
// @Param notification body notifications.DatabaseNotification true "The notification with the read status"
// @Success 200 {object} notifications.DatabaseNotification "The notification"
// @Failure 403 {object} models.Message "The user does not have access to that notification."
// @Failure 403 {object} models.Message "Link shares cannot have notifications."
// @Failure 404 {object} models.Message "The notification does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /notifications/{id} [post]
func markNotificationAsRead(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	a, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return err
	}

	if _, is := a.(*models.LinkSharing); is {
		return echo.ErrForbidden
	}

	// Parse notification ID
	notificationID, err := strconv.ParseInt(c.Param("notificationid"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid notification ID")
	}

	// Parse request body to get read status
	type markReadRequest struct {
		Read bool `json:"read"`
	}
	var req markReadRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Get the notification and check permissions
	notif := &notifications.DatabaseNotification{ID: notificationID}
	service := services.NewNotificationsService(s)

	canMark, err := service.CanMarkNotificationAsRead(notif, a.GetID())
	if err != nil {
		return err
	}
	if !canMark {
		return echo.ErrForbidden
	}

	// Mark as read/unread
	err = service.MarkNotificationAsRead(notif, req.Read)
	if err != nil {
		return err
	}

	if err := s.Commit(); err != nil {
		return err
	}

	// Reload the notification to return updated data
	s = db.NewSession()
	defer s.Close()
	updated := &notifications.DatabaseNotification{ID: notificationID}
	exists, err := s.Get(updated)
	if err != nil {
		return err
	}
	if !exists {
		return echo.NewHTTPError(http.StatusNotFound, "Notification not found")
	}

	return c.JSON(http.StatusOK, updated)
}

// markAllNotificationsAsRead marks all notifications of a user as read
// @Summary Mark all notifications of a user as read
// @tags notifications
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Success 200 {object} models.Message "All notifications marked as read."
// @Failure 403 {object} models.Message "Link shares cannot have notifications."
// @Failure 500 {object} models.Message "Internal error"
// @Router /notifications [post]
func markAllNotificationsAsRead(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	a, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return err
	}

	if _, is := a.(*models.LinkSharing); is {
		return echo.ErrForbidden
	}

	service := services.NewNotificationsService(s)
	err = service.MarkAllNotificationsAsRead(a.GetID())
	if err != nil {
		return err
	}

	if err := s.Commit(); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, models.Message{Message: "success"})
}
