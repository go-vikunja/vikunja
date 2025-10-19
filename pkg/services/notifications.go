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

package services

import (
	"encoding/json"
	"time"

	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/notifications"
	"xorm.io/xorm"
)

// NotificationsService handles notification operations
type NotificationsService struct {
	Session *xorm.Session
}

// NewNotificationsService creates a new notifications service
func NewNotificationsService(s *xorm.Session) *NotificationsService {
	return &NotificationsService{
		Session: s,
	}
}

// GetNotificationsForUser returns all notifications for a user with pagination
func (s *NotificationsService) GetNotificationsForUser(notifiableID int64, limit, start int) (notifs []*notifications.DatabaseNotification, resultCount int, total int64, err error) {
	err = s.Session.
		Where("notifiable_id = ?", notifiableID).
		Limit(limit, start).
		OrderBy("id DESC").
		Find(&notifs)
	if err != nil {
		return nil, 0, 0, err
	}

	total, err = s.Session.
		Where("notifiable_id = ?", notifiableID).
		Count(&notifications.DatabaseNotification{})
	return notifs, len(notifs), total, err
}

// GetNotificationsForNameAndUser returns notifications for a specific event and user
func (s *NotificationsService) GetNotificationsForNameAndUser(notifiableID int64, event string, subjectID int64) (notifs []*notifications.DatabaseNotification, err error) {
	notifs = []*notifications.DatabaseNotification{}
	err = s.Session.
		Where("notifiable_id = ? AND name = ? AND subject_id = ?", notifiableID, event, subjectID).
		Find(&notifs)
	return
}

// CanMarkNotificationAsRead checks if a user can mark a notification as read
func (s *NotificationsService) CanMarkNotificationAsRead(notification *notifications.DatabaseNotification, notifiableID int64) (can bool, err error) {
	can, err = s.Session.
		Where("notifiable_id = ? AND id = ?", notifiableID, notification.ID).
		NoAutoCondition().
		Get(notification)
	return
}

// MarkNotificationAsRead marks a notification as read or unread
func (s *NotificationsService) MarkNotificationAsRead(notification *notifications.DatabaseNotification, read bool) (err error) {
	notification.ReadAt = time.Time{}
	if read {
		notification.ReadAt = time.Now()
	}

	_, err = s.Session.
		Where("id = ?", notification.ID).
		Cols("read_at").
		Update(notification)
	return
}

// MarkAllNotificationsAsRead marks all notifications for a user as read
func (s *NotificationsService) MarkAllNotificationsAsRead(userID int64) (err error) {
	_, err = s.Session.
		Where("notifiable_id = ?", userID).
		Cols("read_at").
		Update(&notifications.DatabaseNotification{ReadAt: time.Now()})
	return
}

// DeleteNotification deletes a single notification if it belongs to the user
func (s *NotificationsService) DeleteNotification(notificationID, userID int64) (err error) {
	_, err = s.Session.
		Where("id = ? AND notifiable_id = ?", notificationID, userID).
		Delete(&notifications.DatabaseNotification{})
	return
}

// DeleteAllReadNotifications deletes all read notifications for a user
func (s *NotificationsService) DeleteAllReadNotifications(userID int64) (err error) {
	_, err = s.Session.
		Where("notifiable_id = ? AND read_at IS NOT NULL AND read_at != ?", userID, time.Time{}).
		Delete(&notifications.DatabaseNotification{})
	return
}

// Notify sends a notification to a notifiable entity (usually a user)
// This handles both email and database notifications
func (s *NotificationsService) Notify(notifiable notifications.Notifiable, notification notifications.Notification) (err error) {
	should, err := notifiable.ShouldNotify()
	if err != nil || !should {
		log.Debugf("Not notifying user %d because they are disabled", notifiable.RouteForDB())
		return err
	}

	err = s.notifyMail(notifiable, notification)
	if err != nil {
		return
	}

	return s.notifyDB(notifiable, notification)
}

// notifyMail sends a mail notification
func (s *NotificationsService) notifyMail(notifiable notifications.Notifiable, notification notifications.Notification) error {
	mail := notification.ToMail(notifiable.Lang())
	if mail == nil {
		return nil
	}

	to, err := notifiable.RouteForMail()
	if err != nil {
		return err
	}
	mail.To(to)

	return notifications.SendMail(mail, notifiable.Lang())
}

// notifyDB saves a notification to the database
func (s *NotificationsService) notifyDB(notifiable notifications.Notifiable, notification notifications.Notification) (err error) {
	dbContent := notification.ToDB()
	if dbContent == nil {
		return nil
	}

	content, err := json.Marshal(dbContent)
	if err != nil {
		return err
	}

	dbNotification := &notifications.DatabaseNotification{
		NotifiableID: notifiable.RouteForDB(),
		Notification: content,
		Name:         notification.Name(),
	}

	if subject, is := notification.(notifications.SubjectID); is {
		dbNotification.SubjectID = subject.SubjectID()
	}

	_, err = s.Session.Insert(dbNotification)
	if err != nil {
		return err
	}

	return nil
}
