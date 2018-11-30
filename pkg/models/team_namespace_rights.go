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

// CanCreate checks if one can create a new team <-> namespace relation
func (tn *TeamNamespace) CanCreate(a web.Auth) bool {
	user := getUserForRights(a)

	n, err := GetNamespaceByID(tn.NamespaceID)
	if err != nil {
		log.Log.Error("Error occurred during CanCreate for TeamNamespace: %s", err)
		return false
	}
	return n.IsAdmin(user)
}

// CanDelete checks if a user can remove a team from a namespace. Only namespace admins can do that.
func (tn *TeamNamespace) CanDelete(a web.Auth) bool {
	user := getUserForRights(a)

	n, err := GetNamespaceByID(tn.NamespaceID)
	if err != nil {
		log.Log.Error("Error occurred during CanDelete for TeamNamespace: %s", err)
		return false
	}
	return n.IsAdmin(user)
}

// CanUpdate checks if a user can update a team from a  Only namespace admins can do that.
func (tn *TeamNamespace) CanUpdate(a web.Auth) bool {
	user := getUserForRights(a)

	n, err := GetNamespaceByID(tn.NamespaceID)
	if err != nil {
		log.Log.Error("Error occurred during CanUpdate for TeamNamespace: %s", err)
		return false
	}
	return n.IsAdmin(user)
}
