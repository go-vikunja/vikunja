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
	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/xorm"
)

// CanCreate checks if a user can create a bot user.
func (b *BotUser) CanCreate(_ *xorm.Session, a web.Auth) (bool, error) {
	if !config.ServiceEnableBotUsers.GetBool() {
		return false, &user.ErrBotUsersDisabled{}
	}
	u, ok := a.(*user.User)
	if !ok || u.IsBot() {
		return false, nil
	}
	return true, nil
}

// CanRead checks if a user can read a bot user.
func (b *BotUser) CanRead(s *xorm.Session, a web.Auth) (bool, int, error) {
	ok, err := b.isOwner(s, a)
	return ok, 0, err
}

// CanUpdate checks if a user can update a bot user.
func (b *BotUser) CanUpdate(s *xorm.Session, a web.Auth) (bool, error) { return b.isOwner(s, a) }

// CanDelete checks if a user can delete a bot user.
func (b *BotUser) CanDelete(s *xorm.Session, a web.Auth) (bool, error) { return b.isOwner(s, a) }

func (b *BotUser) isOwner(s *xorm.Session, a web.Auth) (bool, error) {
	if !config.ServiceEnableBotUsers.GetBool() {
		return false, &user.ErrBotUsersDisabled{}
	}
	u, err := user.GetUserByID(s, b.ID)
	if err != nil {
		if user.IsErrUserDoesNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return u.BotOwnerID == a.GetID(), nil
}
