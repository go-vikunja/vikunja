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

// doerFromAuth resolves the authenticated principal into a full user for event payloads. The JWT
// only carries id + username, so without a re-fetch notifications and emails render the
// auto-generated username instead of the display name (#2720). Status errors (disabled/locked) are
// swallowed because their user is still populated and some flows act on behalf of such accounts
// (e.g. user deletion deletes that user's tasks); the partial principal is used as a last resort.
func doerFromAuth(s *xorm.Session, a web.Auth) *user.User {
	if a == nil {
		return nil
	}

	doer, err := GetUserOrLinkShareUser(s, a)
	if err != nil && !user.IsErrUserStatusError(err) {
		doer = nil
	}
	if doer != nil && doer.ID != 0 {
		return doer
	}

	if u, is := a.(*user.User); is {
		return u
	}
	if share, is := a.(*LinkSharing); is {
		return share.toUser()
	}
	return &user.User{ID: a.GetID()}
}

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
		return l.toUser(), nil
	}

	return
}

// Returns all users or pseudo link shares from a slice of ids. ids < 0 are considered to be a link share in that case.
func getUsersOrLinkSharesFromIDs(s *xorm.Session, ids []int64) (users map[int64]*user.User, err error) {
	users = make(map[int64]*user.User)
	var userIDs []int64
	var linkShareIDs []int64
	for _, id := range ids {
		if id < 0 {
			linkShareIDs = append(linkShareIDs, id*-1)
			continue
		}

		userIDs = append(userIDs, id)
	}

	if len(userIDs) > 0 {
		users, err = user.GetUsersByIDs(s, userIDs)
		if err != nil {
			return
		}
	}

	if len(linkShareIDs) == 0 {
		return
	}

	shares, err := GetLinkSharesByIDs(s, linkShareIDs)
	if err != nil {
		return nil, err
	}

	for _, share := range shares {
		users[share.ID*-1] = share.toUser()
	}

	return
}
