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
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"
	"xorm.io/xorm"
)

func (t *APIToken) CanDelete(s *xorm.Session, a web.Auth) (bool, error) {
	token, err := GetAPITokenByID(s, t.ID)
	if err != nil {
		return false, err
	}

	if token.OwnerID == a.GetID() {
		*t = *token
		return true, nil
	}

	// Allow deletion if the token belongs to a bot owned by the caller.
	botUser, err := user.GetUserByID(s, token.OwnerID)
	if err != nil {
		if user.IsErrUserDoesNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if botUser.IsBot() && botUser.BotOwnerID == a.GetID() {
		*t = *token
		return true, nil
	}

	return false, nil
}

func (t *APIToken) CanCreate(_ *xorm.Session, _ web.Auth) (bool, error) {
	return true, nil
}
