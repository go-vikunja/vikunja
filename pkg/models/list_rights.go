//  Vikunja is a todo-list application to facilitate your life.
//  Copyright 2018 Vikunja and contributors. All rights reserved.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU General Public License as published by
//  the Free Software Foundation, either version 3 of the License, or
//  (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU General Public License for more details.
//
//  You should have received a copy of the GNU General Public License
//  along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/web"
)

// IsAdmin returns whether the user has admin rights on the list or not
func (l *List) IsAdmin(a web.Auth) bool {
	u := getUserForRights(a)

	// Owners are always admins
	if l.OwnerID == u.ID {
		return true
	}

	// Check individual rights
	if l.checkListUserRight(u, UserRightAdmin) {
		return true
	}

	return l.checkListTeamRight(u, TeamRightAdmin)
}

// CanWrite return whether the user can write on that list or not
func (l *List) CanWrite(a web.Auth) bool {
	user := getUserForRights(a)

	// Admins always have write access
	if l.IsAdmin(user) {
		return true
	}

	// Check individual rights
	if l.checkListUserRight(user, UserRightWrite) {
		return true
	}

	return l.checkListTeamRight(user, TeamRightWrite)
}

// CanRead checks if a user has read access to a list
func (l *List) CanRead(a web.Auth) bool {
	user := getUserForRights(a)

	// Admins always have read access
	if l.IsAdmin(user) {
		return true
	}

	// Check individual rights
	if l.checkListUserRight(user, UserRightRead) {
		return true
	}

	if l.checkListTeamRight(user, TeamRightRead) {
		return true
	}

	// Users who are able to write should also be able to read
	return l.CanWrite(a)
}

// CanDelete checks if the user can delete a list
func (l *List) CanDelete(a web.Auth) bool {
	doer := getUserForRights(a)
	return l.IsAdmin(doer)
}

// CanUpdate checks if the user can update a list
func (l *List) CanUpdate(a web.Auth) bool {
	doer := getUserForRights(a)
	return l.CanWrite(doer)
}

// CanCreate checks if the user can update a list
func (l *List) CanCreate(a web.Auth) bool {
	// A user can create a list if he has write access to the namespace
	n, _ := GetNamespaceByID(l.NamespaceID)
	return n.CanWrite(a)
}

func (l *List) checkListTeamRight(user *User, r TeamRight) bool {
	exists, err := x.Select("l.*").
		Table("list").
		Alias("l").
		Join("LEFT", []string{"team_namespaces", "tn"}, " l.namespace_id = tn.namespace_id").
		Join("LEFT", []string{"team_members", "tm"}, "tm.team_id = tn.team_id").
		Join("LEFT", []string{"team_list", "tl"}, "l.id = tl.list_id").
		Join("LEFT", []string{"team_members", "tm2"}, "tm2.team_id = tl.team_id").
		Where("((tm.user_id = ? AND tn.right = ?) OR (tm2.user_id = ? AND tl.right = ?)) AND l.id = ?",
			user.ID, r, user.ID, r, l.ID).
		Exist(&List{})
	if err != nil {
		log.Log.Error("Error occurred during checkListTeamRight for List: %s", err)
		return false
	}

	return exists
}

func (l *List) checkListUserRight(user *User, r UserRight) bool {
	exists, err := x.Select("l.*").
		Table("list").
		Alias("l").
		Join("LEFT", []string{"users_namespace", "un"}, "un.namespace_id = l.namespace_id").
		Join("LEFT", []string{"users_list", "ul"}, "ul.list_id = l.id").
		Join("LEFT", []string{"namespaces", "n"}, "n.id = l.namespace_id").
		Where("((ul.user_id = ? AND ul.right = ?) "+
			"OR (un.user_id = ? AND un.right = ?) "+
			"OR n.owner_id = ?)"+
			"AND l.id = ?",
			user.ID, r, user.ID, r, user.ID, l.ID).
		Exist(&List{})
	if err != nil {
		log.Log.Error("Error occurred during checkListUserRight for List: %s", err)
		return false
	}

	return exists
}
