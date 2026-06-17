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
	"reflect"

	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/notifications"
	"code.vikunja.io/api/pkg/user"

	"xorm.io/xorm"
)

// maxNotificationUserRefreshDepth bounds the reflection walk so an unexpectedly
// deep payload cannot recurse without end.
const maxNotificationUserRefreshDepth = 8

// refreshNotificationsUsers reloads every embedded user of each notification
// from the database. Notifications serialized before the acting user was
// resolved with its full profile (#2720) stored only id+username, so without
// this they keep rendering the auto-generated username instead of the display
// name. It runs at read time and is not persisted; one cache is shared across
// the batch so a user recurring across notifications is fetched only once.
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

	refreshUsersInValue(s, reflect.ValueOf(typed), cache, 0)
	dbn.Notification = typed
}

func refreshUsersInValue(s *xorm.Session, v reflect.Value, cache map[int64]*user.User, depth int) {
	if depth > maxNotificationUserRefreshDepth || !v.IsValid() {
		return
	}

	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			return
		}
		if u, is := v.Interface().(*user.User); is {
			refreshUser(s, u, cache)
			return
		}
		refreshUsersInValue(s, v.Elem(), cache, depth+1)
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if !v.Type().Field(i).IsExported() {
				continue
			}
			refreshUsersInValue(s, v.Field(i), cache, depth+1)
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			refreshUsersInValue(s, v.Index(i), cache, depth+1)
		}
	case reflect.Map:
		for _, key := range v.MapKeys() {
			refreshUsersInValue(s, v.MapIndex(key), cache, depth+1)
		}
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
