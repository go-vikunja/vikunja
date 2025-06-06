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
	"encoding/json"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
)

// Notification is a notification which can be sent via mail or db.
type Notification interface {
	ToMail(lang string) *Mail
	ToDB() interface{}
	Name() string
}

type SubjectID interface {
	SubjectID() int64
}

type NotificationWithSubject interface {
	Notification
	SubjectID
}

// Notifiable is an entity which can be notified. Usually a user.
type Notifiable interface {
	// RouteForMail should return the email address this notifiable has.
	RouteForMail() (string, error)
	// RouteForDB should return the id of the notifiable entity to save it in the database.
	RouteForDB() int64
	// ShouldNotify provides a last-minute way to cancel a notification. It will be called immediately before
	// sending a notification.
	ShouldNotify() (should bool, err error)
	// Lang provides the language which should be used for translations in the mail.
	Lang() string
}

// Notify notifies a notifiable of a notification
func Notify(notifiable Notifiable, notification Notification) (err error) {
	if isUnderTest {
		sentTestNotifications = append(sentTestNotifications, notification)
		return nil
	}

	should, err := notifiable.ShouldNotify()
	if err != nil || !should {
		log.Debugf("Not notifying user %d because they are disabled", notifiable.RouteForDB())
		return err
	}

	err = notifyMail(notifiable, notification)
	if err != nil {
		return
	}

	return notifyDB(notifiable, notification)
}

func notifyMail(notifiable Notifiable, notification Notification) error {
	mail := notification.ToMail(notifiable.Lang())
	if mail == nil {
		return nil
	}

	to, err := notifiable.RouteForMail()
	if err != nil {
		return err
	}
	mail.To(to)

	return SendMail(mail, notifiable.Lang())
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
		Name:         notification.Name(),
	}

	if subject, is := notification.(SubjectID); is {
		dbNotification.SubjectID = subject.SubjectID()
	}

	_, err = s.Insert(dbNotification)
	if err != nil {
		_ = s.Rollback()
		return err
	}

	return s.Commit()
}
