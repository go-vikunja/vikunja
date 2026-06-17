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
	"encoding/json"

	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/notifications"
	"code.vikunja.io/api/pkg/user"

	"xorm.io/xorm"
)

// refreshNotificationsUsers reloads each notification's embedded users from the
// database. Notifications serialized before the acting user was resolved with
// its full profile (#2720) stored only id+username, so without this they keep
// rendering the auto-generated username instead of the display name. It runs at
// read time and is not persisted; one cache is shared across the batch so a
// user recurring across notifications is fetched only once.
func refreshNotificationsUsers(s *xorm.Session, dbNotifications []*notifications.DatabaseNotification) {
	cache := make(map[int64]*user.User)
	for _, dbn := range dbNotifications {
		refreshNotificationUsers(s, dbn, cache)
	}
}

func refreshNotificationUsers(s *xorm.Session, dbn *notifications.DatabaseNotification, cache map[int64]*user.User) {
	typed, ok := notifications.Lookup(dbn.Name)
	if !ok {
		return
	}

	raw, err := json.Marshal(dbn.Notification)
	if err != nil {
		log.Errorf("Could not marshal notification %d to refresh its users: %v", dbn.ID, err)
		return
	}
	if err := json.Unmarshal(raw, typed); err != nil {
		log.Errorf("Could not unmarshal notification %d to refresh its users: %v", dbn.ID, err)
		return
	}

	for _, u := range notificationUsers(typed) {
		refreshUser(s, u, cache)
	}
	dbn.Notification = typed
}

// notificationUsers returns the user fields a stored notification renders, so
// they can be reloaded. New notification types carrying a user belong here.
func notificationUsers(n notifications.Notification) []*user.User {
	switch n := n.(type) {
	case *TaskCommentNotification:
		return []*user.User{n.Doer}
	case *TaskAssignedNotification:
		return []*user.User{n.Doer, n.Assignee}
	case *TaskDeletedNotification:
		return []*user.User{n.Doer}
	case *ProjectCreatedNotification:
		return []*user.User{n.Doer}
	case *TeamMemberAddedNotification:
		return []*user.User{n.Doer, n.Member}
	case *UserMentionedInTaskNotification:
		return []*user.User{n.Doer}
	default:
		return nil
	}
}

// refreshUser overwrites the user in place with its current database row. A
// disabled or locked account is still returned fully populated, so only a
// missing user or a real database error leaves the stored value untouched.
func refreshUser(s *xorm.Session, u *user.User, cache map[int64]*user.User) {
	if u == nil || u.ID == 0 {
		return
	}

	fresh, cached := cache[u.ID]
	if !cached {
		loaded, err := user.GetUserByID(s, u.ID)
		if err != nil && !user.IsErrUserStatusError(err) {
			cache[u.ID] = nil
			return
		}
		fresh = loaded
		cache[u.ID] = fresh
	}

	if fresh != nil {
		*u = *fresh
	}
}
