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

package models

import (
	"code.vikunja.io/api/pkg/notifications"
	"code.vikunja.io/api/pkg/web"
	"xorm.io/xorm"
)

// DatabaseNotifications is a wrapper around the crud operations that come with a database notification.
type DatabaseNotifications struct {
	notifications.DatabaseNotification

	// Whether or not to mark this notification as read or unread.
	// True is read, false is unread.
	Read bool `xorm:"-" json:"read"`

	web.CRUDable    `xorm:"-" json:"-"`
	web.Permissions `xorm:"-" json:"-"`
}

// ReadAll returns all database notifications for a user
// @Summary Get all notifications for the current user
// @Description Returns an array with all notifications for the current user.
// @tags subscriptions
// @Accept json
// @Produce json
// @Param page query int false "The page number. Used for pagination. If not provided, the first page of results is returned."
// @Param per_page query int false "The maximum number of items per page. Note this parameter is limited by the configured maximum of items per page."
// @Security JWTKeyAuth
// @Success 200 {array} notifications.DatabaseNotification "The notifications"
// @Failure 403 {object} web.HTTPError "Link shares cannot have notifications."
// @Failure 500 {object} models.Message "Internal error"
// @Router /notifications [get]
func (d *DatabaseNotifications) ReadAll(s *xorm.Session, a web.Auth, _ string, page int, perPage int) (ls interface{}, resultCount int, numberOfEntries int64, err error) {
	if _, is := a.(*LinkSharing); is {
		return nil, 0, 0, ErrGenericForbidden{}
	}

	limit, start := getLimitFromPageIndex(page, perPage)
	return notifications.GetNotificationsForUser(s, a.GetID(), limit, start)
}

// CanUpdate checks if a user can mark a notification as read.
func (d *DatabaseNotifications) CanUpdate(s *xorm.Session, a web.Auth) (bool, error) {
	if _, is := a.(*LinkSharing); is {
		return false, nil
	}

	return notifications.CanMarkNotificationAsRead(s, &d.DatabaseNotification, a.GetID())
}

// Update marks a notification as read.
// @Summary Mark a notification as (un-)read
// @Description Marks a notification as either read or unread. A user can only mark their own notifications as read.
// @tags subscriptions
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "Notification ID"
// @Success 200 {object} models.DatabaseNotifications "The notification to mark as read."
// @Failure 403 {object} web.HTTPError "The user does not have access to that notification."
// @Failure 403 {object} web.HTTPError "Link shares cannot have notifications."
// @Failure 404 {object} web.HTTPError "The notification does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /notifications/{id} [post]
func (d *DatabaseNotifications) Update(s *xorm.Session, _ web.Auth) (err error) {
	return notifications.MarkNotificationAsRead(s, &d.DatabaseNotification, d.Read)
}
