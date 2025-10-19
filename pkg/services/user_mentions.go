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
	"regexp"
	"strings"

	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/notifications"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/xorm"
)

// UserMentionsService handles user mention functionality across tasks and comments
type UserMentionsService struct{}

// NewUserMentionsService creates a new UserMentionsService instance
func NewUserMentionsService() *UserMentionsService {
	return &UserMentionsService{}
}

// FindMentionedUsersInText extracts @username mentions from text and returns the corresponding users
func (ums *UserMentionsService) FindMentionedUsersInText(s *xorm.Session, text string) (users map[int64]*user.User, err error) {
	// Match @username pattern (letters, numbers, underscore)
	reg := regexp.MustCompile(`@\w+`)
	matches := reg.FindAllString(text, -1)
	if matches == nil {
		return
	}

	// Extract unique usernames
	usernames := []string{}
	for _, match := range matches {
		usernames = append(usernames, strings.TrimPrefix(match, "@"))
	}

	// Look up users by username
	return user.GetUsersByUsername(s, usernames, true)
}

// NotificationSubject represents an entity that can receive notifications
type NotificationSubject interface {
	CanRead(s *xorm.Session, a web.Auth) (bool, int, error)
}

// NotifyMentionedUsers finds mentioned users in text and sends notifications to those with access
func (ums *UserMentionsService) NotifyMentionedUsers(
	sess *xorm.Session,
	subject NotificationSubject,
	text string,
	notification notifications.NotificationWithSubject,
) (users map[int64]*user.User, err error) {
	users, err = ums.FindMentionedUsersInText(sess, text)
	if err != nil {
		return
	}

	if len(users) == 0 {
		return
	}

	log.Debugf("Processing %d mentioned users for subject %d", len(users), notification.SubjectID())

	var notified int
	for _, u := range users {
		// Check if user has read access to the subject
		can, _, err := subject.CanRead(sess, u)
		if err != nil {
			return users, err
		}

		if !can {
			continue
		}

		// Don't notify a user if they were already notified for this subject
		dbn, err := notifications.GetNotificationsForNameAndUser(sess, u.ID, notification.Name(), notification.SubjectID())
		if err != nil {
			return users, err
		}

		if len(dbn) > 0 {
			continue
		}

		// Send notification
		err = notifications.Notify(u, notification)
		if err != nil {
			return users, err
		}
		notified++
	}

	log.Debugf("Notified %d mentioned users for subject %d", notified, notification.SubjectID())

	return
}
