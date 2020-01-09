// Vikunja is a todo-list application to facilitate your life.
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
	"code.vikunja.io/web"
	"github.com/go-xorm/builder"
)

// CanWrite checks if a user has write access to a namespace
func (n *Namespace) CanWrite(a web.Auth) (bool, error) {
	return n.checkRight(a, RightWrite, RightAdmin)
}

// IsAdmin returns true or false if the user is admin on that namespace or not
func (n *Namespace) IsAdmin(a web.Auth) (bool, error) {
	return n.checkRight(a, RightAdmin)
}

// CanRead checks if a user has read access to that namespace
func (n *Namespace) CanRead(a web.Auth) (bool, error) {
	return n.checkRight(a, RightRead, RightWrite, RightAdmin)
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
	if _, is := a.(*LinkSharing); is {
		return false, nil
	}

	// This is currently a dummy function, later on we could imagine global limits etc.
	return true, nil
}

func (n *Namespace) checkRight(a web.Auth, rights ...Right) (bool, error) {

	// If the auth is a link share, don't do anything
	if _, is := a.(*LinkSharing); is {
		return false, nil
	}

	// Get the namespace and check the right
	err := n.GetSimpleByID()
	if err != nil {
		return false, err
	}

	if a.GetID() == n.OwnerID {
		return true, nil
	}

	/*
		The following loop creates an sql condition like this one:

		namespaces.owner_id = 1 OR
		(users_namespace.user_id = 1 AND users_namespace.right = 1) OR
		(team_members.user_id = 1 AND team_namespaces.right = 1) OR


		for each passed right. That way, we can check with a single sql query (instead if 8)
		if the user has the right to see the list or not.
	*/

	var conds []builder.Cond
	conds = append(conds, builder.Eq{"namespaces.owner_id": a.GetID()})
	for _, r := range rights {
		// User conditions
		// If the namespace was shared directly with the user and the user has the right
		conds = append(conds, builder.And(
			builder.Eq{"users_namespace.user_id": a.GetID()},
			builder.Eq{"users_namespace.right": r},
		))

		// Team rights
		// If the namespace was shared directly with the team and the team has the right
		conds = append(conds, builder.And(
			builder.Eq{"team_members.user_id": a.GetID()},
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
