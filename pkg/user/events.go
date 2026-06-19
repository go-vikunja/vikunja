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

package user

// CreatedEvent represents a CreatedEvent event
type CreatedEvent struct {
	User *User
}

// Name defines the name for CreatedEvent
func (t *CreatedEvent) Name() string {
	return "user.created"
}

// LoginSucceededEvent is fired after a user successfully authenticated,
// regardless of the auth provider (local, LDAP, OpenID).
type LoginSucceededEvent struct {
	User *User `json:"user"`
}

// Name defines the name for LoginSucceededEvent
func (t *LoginSucceededEvent) Name() string {
	return "user.login.succeeded"
}

// LoginFailedEvent is fired for every failed password check of a known user.
type LoginFailedEvent struct {
	User *User `json:"user"`
}

// Name defines the name for LoginFailedEvent
func (t *LoginFailedEvent) Name() string {
	return "user.login.failed"
}

// LogoutEvent is fired when a user destroys their session.
type LogoutEvent struct {
	UserID int64 `json:"user_id"`
}

// Name defines the name for LogoutEvent
func (t *LogoutEvent) Name() string {
	return "user.logout"
}
