// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2020 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/web"
	"xorm.io/builder"
)

// CanWrite return whether the user can write on that list or not
func (l *List) CanWrite(a web.Auth) (bool, error) {

	// Get the list and check the right
	originalList := &List{ID: l.ID}
	err := originalList.GetSimpleByID()
	if err != nil {
		return false, err
	}

	// We put the result of the is archived check in a separate variable to be able to return it later without
	// needing to recheck it again
	errIsArchived := originalList.CheckIsArchived()

	var canWrite bool

	// Check if we're dealing with a share auth
	shareAuth, ok := a.(*LinkSharing)
	if ok {
		return originalList.ID == shareAuth.ListID &&
			(shareAuth.Right == RightWrite || shareAuth.Right == RightAdmin), errIsArchived
	}

	// Check if the user is either owner or can write to the list
	if originalList.isOwner(&user.User{ID: a.GetID()}) {
		canWrite = true
	}

	if canWrite {
		return canWrite, errIsArchived
	}

	canWrite, err = originalList.checkRight(a, RightWrite, RightAdmin)
	if err != nil {
		return false, err
	}
	return canWrite, errIsArchived
}

// CanRead checks if a user has read access to a list
func (l *List) CanRead(a web.Auth) (bool, error) {
	// Check if the user is either owner or can read
	if err := l.GetSimpleByID(); err != nil {
		return false, err
	}

	// Check if we're dealing with a share auth
	shareAuth, ok := a.(*LinkSharing)
	if ok {
		return l.ID == shareAuth.ListID &&
			(shareAuth.Right == RightRead || shareAuth.Right == RightWrite || shareAuth.Right == RightAdmin), nil
	}

	if l.isOwner(&user.User{ID: a.GetID()}) {
		return true, nil
	}
	return l.checkRight(a, RightRead, RightWrite, RightAdmin)
}

// CanUpdate checks if the user can update a list
func (l *List) CanUpdate(a web.Auth) (canUpdate bool, err error) {
	canUpdate, err = l.CanWrite(a)
	// If the list is archived and the user tries to un-archive it, let the request through
	if IsErrListIsArchived(err) && !l.IsArchived {
		err = nil
	}
	return canUpdate, err
}

// CanDelete checks if the user can delete a list
func (l *List) CanDelete(a web.Auth) (bool, error) {
	return l.IsAdmin(a)
}

// CanCreate checks if the user can create a list
func (l *List) CanCreate(a web.Auth) (bool, error) {
	// A user can create a list if he has write access to the namespace
	n := &Namespace{ID: l.NamespaceID}
	return n.CanWrite(a)
}

// IsAdmin returns whether the user has admin rights on the list or not
func (l *List) IsAdmin(a web.Auth) (bool, error) {
	originalList := &List{ID: l.ID}
	err := originalList.GetSimpleByID()
	if err != nil {
		return false, err
	}

	// Check if we're dealing with a share auth
	shareAuth, ok := a.(*LinkSharing)
	if ok {
		return originalList.ID == shareAuth.ListID && shareAuth.Right == RightAdmin, nil
	}

	// Check all the things
	// Check if the user is either owner or can write to the list
	// Owners are always admins
	if originalList.isOwner(&user.User{ID: a.GetID()}) {
		return true, nil
	}
	return originalList.checkRight(a, RightAdmin)
}

// Little helper function to check if a user is list owner
func (l *List) isOwner(u *user.User) bool {
	return l.OwnerID == u.ID
}

// Checks n different rights for any given user
func (l *List) checkRight(a web.Auth, rights ...Right) (bool, error) {

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
			builder.Eq{"ul.user_id": a.GetID()},
			builder.Eq{"ul.right": r},
		))
		// If the namespace this list belongs to was shared directly with the user and the user has the right
		conds = append(conds, builder.And(
			builder.Eq{"un.user_id": a.GetID()},
			builder.Eq{"un.right": r},
		))

		// Team rights
		// If the list was shared directly with the team and the team has the right
		conds = append(conds, builder.And(
			builder.Eq{"tm2.user_id": a.GetID()},
			builder.Eq{"tl.right": r},
		))
		// If the namespace this list belongs to was shared directly with the team and the team has the right
		conds = append(conds, builder.And(
			builder.Eq{"tm.user_id": a.GetID()},
			builder.Eq{"tn.right": r},
		))
	}

	// If the user is the owner of a namespace, it has any right, all the time
	conds = append(conds, builder.Eq{"n.owner_id": a.GetID()})

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
	return exists, err
}
