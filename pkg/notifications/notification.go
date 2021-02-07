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

package notifications

import (
	"encoding/json"
	"time"

	"code.vikunja.io/api/pkg/db"
)

// Notification is a notification which can be sent via mail or db.
type Notification interface {
	ToMail() *Mail
	ToDB() interface{}
}

// Notifiable is an entity which can be notified. Usually a user.
type Notifiable interface {
	// Should return the email address this notifiable has.
	RouteForMail() string
	// Should return the id of the notifiable entity
	RouteForDB() int64
}

// DatabaseNotification represents a notification that was saved to the database
type DatabaseNotification struct {
	// The unique, numeric id of this notification.
	ID int64 `xorm:"bigint autoincr not null unique pk" json:"id"`

	// The ID of the notifiable this notification is associated with.
	NotifiableID int64 `xorm:"bigint not null" json:"-"`
	// The actual content of the notification.
	Notification interface{} `xorm:"json not null" json:"notification"`

	// A timestamp when this notification was created. You cannot change this value.
	Created time.Time `xorm:"created not null" json:"created"`
}

// TableName resolves to a better table name for notifications
func (d *DatabaseNotification) TableName() string {
	return "notifications"
}

// Notify notifies a notifiable of a notification
func Notify(notifiable Notifiable, notification Notification) (err error) {

	err = notifyMail(notifiable, notification)
	if err != nil {
		return
	}

	return notifyDB(notifiable, notification)
}

func notifyMail(notifiable Notifiable, notification Notification) error {
	mail := notification.ToMail()
	if mail == nil {
		return nil
	}

	mail.To(notifiable.RouteForMail())

	return SendMail(mail)
}

func notifyDB(notifiable Notifiable, notification Notification) (err error) {

	dbContent := notification.ToDB()
	if dbContent == nil {
		return nil
	}

	content, err := json.Marshal(dbContent)
	if err != nil {
		return err
	}

	s := db.NewSession()
	dbNotification := &DatabaseNotification{
		NotifiableID: notifiable.RouteForDB(),
		Notification: content,
	}

	_, err = s.Insert(dbNotification)
	if err != nil {
		_ = s.Rollback()
		return err
	}

	return s.Commit()
}
