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
	"xorm.io/xorm"
)

// CanRead implements the read permission check for a link share
func (share *LinkSharing) CanRead(s *xorm.Session, u *user.User) (bool, int, error) {
	// Don't allow creating link shares if the user itself authenticated with a link share
	if u == nil {
		return false, 0, nil
	}

	l, err := GetProjectByShareHash(s, share.Hash)
	if err != nil {
		return false, 0, err
	}
	return l.CanRead(s, u)
}

// CanDelete implements the delete permission check for a link share
func (share *LinkSharing) CanDelete(s *xorm.Session, u *user.User) (bool, error) {
	return share.canDoLinkShare(s, u)
}

// CanUpdate implements the update permission check for a link share
func (share *LinkSharing) CanUpdate(s *xorm.Session, u *user.User) (bool, error) {
	return share.canDoLinkShare(s, u)
}

// CanCreate implements the create permission check for a link share
func (share *LinkSharing) CanCreate(s *xorm.Session, u *user.User) (bool, error) {
	return share.canDoLinkShare(s, u)
}

func (share *LinkSharing) canDoLinkShare(s *xorm.Session, u *user.User) (bool, error) {
	// Don't allow creating link shares if the user itself authenticated with a link share
	if u == nil {
		return false, nil
	}

	l, err := GetProjectSimpleByID(s, share.ProjectID)
	if err != nil {
		return false, err
	}

	// Check if the user is admin when the link permission is admin
	if share.Permission == PermissionAdmin {
		return l.IsAdmin(s, u)
	}

	return l.CanWrite(s, u)
}
