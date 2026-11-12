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
	"context"
	"encoding/json"
	"slices"

	"code.vikunja.io/api/pkg/mail"

	"code.vikunja.io/api/pkg/db"

	"xorm.io/xorm"
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

type ThreadID interface {
	ThreadID() string
}

// Titler is an optional capability for notifications that can render a
// one-line, translated title. Used as the mail subject when ToMail does not
// set one explicitly, and as the item title in the notifications feed.
type Titler interface {
	ToTitle(lang string) string
}

// ProjectID reports the project a notification is about, 0 if it is
// account-scoped, ProjectIDUnresolved if it is about a project it cannot name.
type ProjectID interface {
	ProjectID() int64
}

// PersistedNotification is stored and read back, so it must declare its project.
type PersistedNotification interface {
	Notification
	ProjectID
}

// ProjectIDUnresolved marks a project-scoped notification whose project could
// not be determined. 0 would make the row account-scoped and hand its payload
// back unchecked.
const ProjectIDUnresolved int64 = -1

// ProjectIDOf returns the project a notification is about. Unregistered types
// need not implement ProjectID and count as account-scoped.
func ProjectIDOf(n Notification) int64 {
	if p, is := n.(ProjectID); is {
		return p.ProjectID()
	}
	return 0
}

var registry = map[string]func() PersistedNotification{}

// Register makes a notification type discoverable by name. It should be
// called from init() in the package that defines the type. Only notifications
// that persist to the database need to register, since only persisted
// notifications are re-hydrated from JSON (e.g. by the feed handler).
// The name is derived from the notification's own Name() method, so it stays
// in one place.
func Register(factory func() PersistedNotification) {
	registry[factory().Name()] = factory
}

// Lookup returns a fresh, empty instance of the notification type registered
// under the given name. The second return value is false if no type is
// registered with that name.
func Lookup(name string) (Notification, bool) {
	f, ok := registry[name]
	if !ok {
		return nil, false
	}
	return f(), true
}

// RegisteredNames returns every registered notification name, sorted.
func RegisteredNames() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

// Notifiable is an entity which can be notified. Usually a user.
type Notifiable interface {
	// RouteForMail should return the email address this notifiable has.
	RouteForMail() (string, error)
	// RouteForDB should return the id of the notifiable entity to save it in the database.
	RouteForDB() int64
	// ShouldNotify provides a last-minute way to cancel a notification. It will be called immediately before
	// sending a notification. An optional session can be passed to reuse an existing transaction.
	ShouldNotify(sessions ...*xorm.Session) (should bool, err error)
	// Lang provides the language which should be used for translations in the mail.
	Lang() string
}

// Notify commits notification content and its mail job together.
func Notify(notifiable Notifiable, notification Notification, sessions ...*xorm.Session) error {
	return NotifyQueued(notifiable, notification, "mail", "", sessions...)
}

func NotifyQueued(notifiable Notifiable, notification Notification, queue, key string, sessions ...*xorm.Session) error {
	if isUnderTest {
		sentTestNotifications = append(sentTestNotifications, notification)
		return nil
	}
	var s *xorm.Session
	owned := len(sessions) == 0 || sessions[0] == nil
	if owned {
		s = db.NewSession()
		defer s.Close()
	} else {
		s = sessions[0]
	}
	should, err := notifiable.ShouldNotify(s)
	if err != nil || !should {
		return err
	}
	opts, err := prepareMail(notifiable, notification)
	if err != nil {
		return err
	}
	if opts != nil {
		if err := mail.Enqueue(s, opts, queue, key); err != nil {
			return err
		}
	}
	if err := notifyDB(notifiable, notification, s); err != nil {
		return err
	}
	if owned {
		return s.Commit()
	}
	return nil
}

func NotifyDB(notifiable Notifiable, notification Notification, s *xorm.Session) error {
	should, err := notifiable.ShouldNotify(s)
	if err != nil || !should {
		return err
	}
	return notifyDB(notifiable, notification, s)
}

func SendMailNow(ctx context.Context, notifiable Notifiable, notification Notification) error {
	opts, err := prepareMail(notifiable, notification)
	if err != nil || opts == nil {
		return err
	}
	return mail.Deliver(ctx, opts)
}

func prepareMail(notifiable Notifiable, notification Notification) (*mail.Opts, error) {
	m := notification.ToMail(notifiable.Lang())
	if m == nil {
		return nil, nil
	}
	if m.subject == "" {
		if t, ok := notification.(Titler); ok {
			m.subject = t.ToTitle(notifiable.Lang())
		}
	}
	to, err := notifiable.RouteForMail()
	if err != nil {
		return nil, err
	}
	m.To(to)
	if thread, ok := notification.(ThreadID); ok {
		m.ThreadID(thread.ThreadID())
	}
	return RenderMail(m, notifiable.Lang())
}

func notifyMail(notifiable Notifiable, notification Notification) error {
	opts, err := prepareMail(notifiable, notification)
	if err != nil || opts == nil {
		return err
	}
	return mail.SendMail(opts)
}

func notifyDB(notifiable Notifiable, notification Notification, s *xorm.Session) error {
	dbContent := notification.ToDB()
	if dbContent == nil {
		return nil
	}

	content, err := json.Marshal(dbContent)
	if err != nil {
		return err
	}

	dbNotification := &DatabaseNotification{
		NotifiableID: notifiable.RouteForDB(),
		Notification: json.RawMessage(content),
		Name:         notification.Name(),
	}

	if subject, is := notification.(SubjectID); is {
		dbNotification.SubjectID = subject.SubjectID()
	}
	dbNotification.ProjectID = ProjectIDOf(notification)

	_, err = s.Insert(dbNotification)
	return err
}
