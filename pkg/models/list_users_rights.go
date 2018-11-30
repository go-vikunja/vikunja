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

// CanCreate checks if the user can create a new user <-> list relation
func (lu *ListUser) CanCreate(a web.Auth) bool {
	doer := getUserForRights(a)

	// Get the list and check if the user has write access on it
	l := List{ID: lu.ListID}
	if err := l.GetSimpleByID(); err != nil {
		log.Log.Error("Error occurred during CanCreate for ListUser: %s", err)
		return false
	}
	return l.CanWrite(doer)
}

// CanDelete checks if the user can delete a user <-> list relation
func (lu *ListUser) CanDelete(a web.Auth) bool {
	doer := getUserForRights(a)

	// Get the list and check if the user has write access on it
	l := List{ID: lu.ListID}
	if err := l.GetSimpleByID(); err != nil {
		log.Log.Error("Error occurred during CanDelete for ListUser: %s", err)
		return false
	}
	return l.CanWrite(doer)
}

// CanUpdate checks if the user can update a user <-> list relation
func (lu *ListUser) CanUpdate(a web.Auth) bool {
	doer := getUserForRights(a)

	// Get the list and check if the user has write access on it
	l := List{ID: lu.ListID}
	if err := l.GetSimpleByID(); err != nil {
		log.Log.Error("Error occurred during CanUpdate for ListUser: %s", err)
		return false
	}
	return l.CanWrite(doer)
}
