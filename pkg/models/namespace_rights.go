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
	"code.vikunja.io/web"
	"github.com/go-xorm/builder"
)

// CanWrite checks if a user has write access to a namespace
func (n *Namespace) CanWrite(a web.Auth) (bool, error) {

	// Get the namespace and check the right
	originalNamespace := &Namespace{ID: n.ID}
	err := originalNamespace.GetSimpleByID()
	if err != nil {
		return false, err
	}

	u := getUserForRights(a)
	if originalNamespace.isOwner(u) {
		return true, nil
	}
	return originalNamespace.checkRight(u, RightWrite, RightAdmin)
}

// IsAdmin returns true or false if the user is admin on that namespace or not
func (n *Namespace) IsAdmin(a web.Auth) (bool, error) {
	originalNamespace := &Namespace{ID: n.ID}
	err := originalNamespace.GetSimpleByID()
	if err != nil {
		return false, err
	}

	u := getUserForRights(a)
	if originalNamespace.isOwner(u) {
		return true, nil
	}
	return originalNamespace.checkRight(u, RightAdmin)
}

// CanRead checks if a user has read access to that namespace
func (n *Namespace) CanRead(a web.Auth) (bool, error) {
	originalNamespace := &Namespace{ID: n.ID}
	err := originalNamespace.GetSimpleByID()
	if err != nil {
		return false, err
	}

	u := getUserForRights(a)
	if originalNamespace.isOwner(u) {
		return true, nil
	}
	return n.checkRight(u, RightRead, RightWrite, RightAdmin)
}

// CanUpdate checks if the user can update the namespace
func (n *Namespace) CanUpdate(a web.Auth) (bool, error) {
	return n.IsAdmin(a)
}

// CanDelete checks if the user can delete a namespace
func (n *Namespace) CanDelete(a web.Auth) (bool, error) {
	return n.IsAdmin(a)
}

// CanCreate checks if the user can create a new namespace
func (n *Namespace) CanCreate(a web.Auth) (bool, error) {
	// This is currently a dummy function, later on we could imagine global limits etc.
	return true, nil
}

// Small helper function to check if a user owns the namespace
func (n *Namespace) isOwner(user *User) bool {
	return n.OwnerID == user.ID
}

func (n *Namespace) checkRight(user *User, rights ...Right) (bool, error) {

	/*
		The following loop creates an sql condition like this one:

		namespaces.owner_id = 1 OR
		(users_namespace.user_id = 1 AND users_namespace.right = 1) OR
		(team_members.user_id = 1 AND team_namespaces.right = 1) OR


		for each passed right. That way, we can check with a single sql query (instead if 8)
		if the user has the right to see the list or not.
	*/

	var conds []builder.Cond
	conds = append(conds, builder.Eq{"namespaces.owner_id": user.ID})
	for _, r := range rights {
		// User conditions
		// If the namespace was shared directly with the user and the user has the right
		conds = append(conds, builder.And(
			builder.Eq{"users_namespace.user_id": user.ID},
			builder.Eq{"users_namespace.right": r},
		))

		// Team rights
		// If the namespace was shared directly with the team and the team has the right
		conds = append(conds, builder.And(
			builder.Eq{"team_members.user_id": user.ID},
			builder.Eq{"team_namespaces.right": r},
		))
	}

	exists, err := x.Select("namespaces.*").
		Table("namespaces").
		// User stuff
		Join("LEFT", "users_namespace", "users_namespace.namespace_id = namespaces.id").
		// Teams stuff
		Join("LEFT", "team_namespaces", "namespaces.id = team_namespaces.namespace_id").
		Join("LEFT", "team_members", "team_members.team_id = team_namespaces.team_id").
		// The actual condition
		Where(builder.And(
			builder.Or(
				conds...,
			),
			builder.Eq{"namespaces.id": n.ID},
		)).
		Exist(&List{})
	return exists, err
}
