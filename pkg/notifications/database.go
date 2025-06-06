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

package notifications

import (
	"time"

	"xorm.io/xorm"
)

// DatabaseNotification represents a notification that was saved to the database
type DatabaseNotification struct {
	// The unique, numeric id of this notification.
	ID int64 `xorm:"bigint autoincr not null unique pk" json:"id" param:"notificationid"`

	// The ID of the notifiable this notification is associated with.
	NotifiableID int64 `xorm:"bigint not null" json:"-"`
	// The actual content of the notification.
	Notification interface{} `xorm:"json not null" json:"notification"`
	// The name of the notification
	Name string `xorm:"varchar(250) index not null" json:"name"`
	// The thing the notification is about. Used to check if a notification for this thing already happened or not.
	SubjectID int64 `xorm:"bigint null" json:"-"`

	// When this notification is marked as read, this will be updated with the current timestamp.
	ReadAt time.Time `xorm:"datetime null" json:"read_at"`

	// A timestamp when this notification was created. You cannot change this value.
	Created time.Time `xorm:"created not null" json:"created"`
}

// TableName resolves to a better table name for notifications
func (d *DatabaseNotification) TableName() string {
	return "notifications"
}

// GetNotificationsForUser returns all notifications for a user. It is possible to limit the amount of notifications
// to return with the limit and start parameters.
// We're not passing a user object in directly because every other package imports this one so we'd get import cycles.
func GetNotificationsForUser(s *xorm.Session, notifiableID int64, limit, start int) (notifications []*DatabaseNotification, resultCount int, total int64, err error) {
	err = s.
		Where("notifiable_id = ?", notifiableID).
		Limit(limit, start).
		OrderBy("id DESC").
		Find(&notifications)
	if err != nil {
		return nil, 0, 0, err
	}

	total, err = s.
		Where("notifiable_id = ?", notifiableID).
		Count(&DatabaseNotification{})
	return notifications, len(notifications), total, err
}

func GetNotificationsForNameAndUser(s *xorm.Session, notifiableID int64, event string, subjectID int64) (notifications []*DatabaseNotification, err error) {
	notifications = []*DatabaseNotification{}
	err = s.Where("notifiable_id = ? AND name = ? AND subject_id = ?", notifiableID, event, subjectID).
		Find(&notifications)
	return
}

// CanMarkNotificationAsRead checks if a user can mark a notification as read.
func CanMarkNotificationAsRead(s *xorm.Session, notification *DatabaseNotification, notifiableID int64) (can bool, err error) {
	can, err = s.
		Where("notifiable_id = ? AND id = ?", notifiableID, notification.ID).
		NoAutoCondition().
		Get(notification)
	return
}

// MarkNotificationAsRead marks a notification as read. It should be called only after CanMarkNotificationAsRead has
// been called.
func MarkNotificationAsRead(s *xorm.Session, notification *DatabaseNotification, read bool) (err error) {
	notification.ReadAt = time.Time{}
	if read {
		notification.ReadAt = time.Now()
	}

	_, err = s.
		Where("id = ?", notification.ID).
		Cols("read_at").
		Update(notification)
	return
}

func MarkAllNotificationsAsRead(s *xorm.Session, userID int64) (err error) {
	_, err = s.
		Where("notifiable_id = ?", userID).
		Cols("read_at").
		Update(&DatabaseNotification{ReadAt: time.Now()})
	return
}
