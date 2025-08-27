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

// GetUserOrLinkShareUser returns either a user or a link share disguised as a user.
func GetUserOrLinkShareUser(s *xorm.Session, a web.Auth) (uu *user.User, err error) {
	if u, is := a.(*user.User); is {
		uu, err = user.GetUserByID(s, u.ID)
		return
	}

	if ls, is := a.(*LinkSharing); is {
		l, err := GetLinkShareByID(s, ls.ID)
		if err != nil {
			return nil, err
		}
		return l.ToUser(), nil
	}

	return
}

// GetUsersOrLinkSharesFromIDsFunc is a function that returns all users or pseudo link shares from a slice of ids.
// It is used to break the dependency cycle between the models and services packages.
var GetUsersOrLinkSharesFromIDsFunc func(s *xorm.Session, ids []int64) (users map[int64]*user.User, err error)

// Returns all users or pseudo link shares from a slice of ids. ids < 0 are considered to be a link share in that case.
func GetUsersOrLinkSharesFromIDs(s *xorm.Session, ids []int64) (users map[int64]*user.User, err error) {
	if GetUsersOrLinkSharesFromIDsFunc == nil {
		panic("GetUsersOrLinkSharesFromIDsFunc is not set")
	}
	return GetUsersOrLinkSharesFromIDsFunc(s, ids)
}
