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
	"github.com/go-xorm/builder"
)

// CanWrite return whether the user can write on that list or not
func (l *List) CanWrite(a web.Auth) bool {
	user := getUserForRights(a)

	// Check all the things
	// Check if the user is either owner or can write to the list
	return l.isOwner(user) || l.checkRight(user, RightWrite, RightAdmin)
}

// CanRead checks if a user has read access to a list
func (l *List) CanRead(a web.Auth) bool {
	user := getUserForRights(a)

	// Check all the things
	// Check if the user is either owner or can read
	return l.isOwner(user) || l.checkRight(user, RightRead, RightWrite, RightAdmin)
}

// CanUpdate checks if the user can update a list
func (l *List) CanUpdate(a web.Auth) bool {
	return l.CanWrite(a)
}

// CanDelete checks if the user can delete a list
func (l *List) CanDelete(a web.Auth) bool {
	return l.IsAdmin(a)
}

// CanCreate checks if the user can update a list
func (l *List) CanCreate(a web.Auth) bool {
	// A user can create a list if he has write access to the namespace
	n, _ := GetNamespaceByID(l.NamespaceID)
	return n.CanWrite(a)
}

// IsAdmin returns whether the user has admin rights on the list or not
func (l *List) IsAdmin(a web.Auth) bool {
	user := getUserForRights(a)

	// Check all the things
	// Check if the user is either owner or can write to the list
	// Owners are always admins
	return l.isOwner(user) || l.checkRight(user, RightAdmin)
}

// Little helper function to check if a user is list owner
func (l *List) isOwner(u *User) bool {
	return l.OwnerID == u.ID
}

// Checks n different rights for any given user
func (l *List) checkRight(user *User, rights ...Right) bool {

	/*
			The following loop creates an sql condition like this one:

		    (ul.user_id = 1 AND ul.right = 1) OR (un.user_id = 1 AND un.right = 1) OR
			(tm.user_id = 1 AND tn.right = 1) OR (tm2.user_id = 1 AND tl.right = 1) OR

			for each passed right. That way, we can check with a single sql query (instead if 8)
			if the user has the right to see the list or not.
	*/

	var conds []builder.Cond
	for _, r := range rights {
		// User conditions
		// If the list was shared directly with the user and the user has the right
		conds = append(conds, builder.And(
			builder.Eq{"ul.user_id": user.ID},
			builder.Eq{"ul.right": r},
		))
		// If the namespace this list belongs to was shared directly with the user and the user has the right
		conds = append(conds, builder.And(
			builder.Eq{"un.user_id": user.ID},
			builder.Eq{"un.right": r},
		))

		// Team rights
		// If the list was shared directly with the team and the team has the right
		conds = append(conds, builder.And(
			builder.Eq{"tm2.user_id": user.ID},
			builder.Eq{"tl.right": r},
		))
		// If the namespace this list belongs to was shared directly with the team and the team has the right
		conds = append(conds, builder.And(
			builder.Eq{"tm.user_id": user.ID},
			builder.Eq{"tn.right": r},
		))
	}

	exists, err := x.Select("l.*").
		Table("list").
		Alias("l").
		// User stuff
		Join("LEFT", []string{"users_namespace", "un"}, "un.namespace_id = l.namespace_id").
		Join("LEFT", []string{"users_list", "ul"}, "ul.list_id = l.id").
		Join("LEFT", []string{"namespaces", "n"}, "n.id = l.namespace_id").
		// Team stuff
		Join("LEFT", []string{"team_namespaces", "tn"}, " l.namespace_id = tn.namespace_id").
		Join("LEFT", []string{"team_members", "tm"}, "tm.team_id = tn.team_id").
		Join("LEFT", []string{"team_list", "tl"}, "l.id = tl.list_id").
		Join("LEFT", []string{"team_members", "tm2"}, "tm2.team_id = tl.team_id").
		// The actual condition
		Where(builder.And(
			builder.Or(
				conds...,
			),
			builder.Eq{"l.id": l.ID},
		)).
		Exist(&List{})
	if err != nil {
		log.Log.Error("Error occurred during checkRight for list: %s", err)
		return false
	}

	return exists
}
