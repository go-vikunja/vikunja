//   Vikunja is a todo-list application to facilitate your life.
//   Copyright 2019 Vikunja and contributors. All rights reserved.
//
//   This program is free software: you can redistribute it and/or modify
//   it under the terms of the GNU General Public License as published by
//   the Free Software Foundation, either version 3 of the License, or
//   (at your option) any later version.
//
//   This program is distributed in the hope that it will be useful,
//   but WITHOUT ANY WARRANTY; without even the implied warranty of
//   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//   GNU General Public License for more details.
//
//   You should have received a copy of the GNU General Public License
//   along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import "code.vikunja.io/web"

// CanRead implements the read right check for a link share
func (share *LinkSharing) CanRead(a web.Auth) (bool, error) {
	// Don't allow creating link shares if the user itself authenticated with a link share
	if _, is := a.(*LinkSharing); is {
		return false, nil
	}

	l, err := GetListByShareHash(share.Hash)
	if err != nil {
		return false, err
	}
	return l.CanRead(a)
}

// CanDelete implements the delete right check for a link share
func (share *LinkSharing) CanDelete(a web.Auth) (bool, error) {
	return share.canDoLinkShare(a)
}

// CanUpdate implements the update right check for a link share
func (share *LinkSharing) CanUpdate(a web.Auth) (bool, error) {
	return share.canDoLinkShare(a)
}

// CanCreate implements the create right check for a link share
func (share *LinkSharing) CanCreate(a web.Auth) (bool, error) {
	return share.canDoLinkShare(a)
}

func (share *LinkSharing) canDoLinkShare(a web.Auth) (bool, error) {
	// Don't allow creating link shares if the user itself authenticated with a link share
	if _, is := a.(*LinkSharing); is {
		return false, nil
	}

	l, err := GetListSimplByTaskID(share.ListID)
	if err != nil {
		return false, err
	}
	return l.CanWrite(a)
}
