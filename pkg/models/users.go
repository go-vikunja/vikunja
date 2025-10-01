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
	// Case 1: The auth object is already a *user.User (could also be a link share proxy with negative ID)
	if u, is := a.(*user.User); is {
		// Negative user IDs represent link share principals in the service layer.
		if u.ID < 0 {
			shareID := u.ID * -1
			l, err := GetLinkShareByID(s, shareID)
			if err != nil {
				return nil, err
			}
			if l == nil {
				return nil, ErrProjectShareDoesNotExist{ID: shareID}
			}
			if NewUserProxyFromLinkShareFunc == nil {
				panic("NewUserProxyFromLinkShareFunc is not set")
			}
			return NewUserProxyFromLinkShareFunc(l), nil
		}
		uu, err = user.GetUserByID(s, u.ID)
		return
	}

	// Case 2: The auth object is the legacy *LinkSharing instance
	if ls, is := a.(*LinkSharing); is {
		l, err := GetLinkShareByID(s, ls.ID)
		if err != nil {
			return nil, err
		}
		if NewUserProxyFromLinkShareFunc == nil {
			panic("NewUserProxyFromLinkShareFunc is not set")
		}
		return NewUserProxyFromLinkShareFunc(l), nil
	}

	// Unknown auth type → return nil, nil (caller should treat as unauthenticated)
	return nil, nil
}

// NewUserProxyFromLinkShareFunc is a function that returns a user proxy from a link share.
// It is used to break the dependency cycle between the models and services packages.
var NewUserProxyFromLinkShareFunc func(share *LinkSharing) *user.User

// GetUsersOrLinkSharesFromIDsFunc is a function that returns all users or pseudo link shares from a slice of ids.
// It is used to break the dependency cycle between the models and services packages.
var GetUsersOrLinkSharesFromIDsFunc func(s *xorm.Session, ids []int64) (users map[int64]*user.User, err error)

// Returns all users or pseudo link shares from a slice of ids. ids < 0 are considered to be a link share in that case.
//
// @Deprecated This function is deprecated and will be removed in a future version. Use services.UserService.GetUsersAndProxiesFromIDs instead.
func GetUsersOrLinkSharesFromIDs(s *xorm.Session, ids []int64) (users map[int64]*user.User, err error) {
	if GetUsersOrLinkSharesFromIDsFunc == nil {
		panic("GetUsersOrLinkSharesFromIDsFunc is not set")
	}
	return GetUsersOrLinkSharesFromIDsFunc(s, ids)
}
