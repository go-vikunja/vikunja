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

// CanDelete checks if the user can delete an task
func (t *ListTask) CanDelete(a web.Auth) bool {
	doer := getUserForRights(a)

	// Get the task
	lI, err := GetListTaskByID(t.ID)
	if err != nil {
		log.Log.Error("Error occurred during CanDelete for ListTask: %s", err)
		return false
	}

	// A user can delete an task if he has write acces to its list
	l := &List{ID: lI.ListID}
	l.ReadOne()
	return l.CanWrite(doer)
}

// CanUpdate determines if a user has the right to update a list task
func (t *ListTask) CanUpdate(a web.Auth) bool {
	doer := getUserForRights(a)

	// Get the task
	lI, err := getTaskByIDSimple(t.ID)
	if err != nil {
		log.Log.Error("Error occurred during CanUpdate (getTaskByIDSimple) for ListTask: %s", err)
		return false
	}

	// A user can update an task if he has write acces to its list
	l := &List{ID: lI.ListID}
	err = l.GetSimpleByID()
	if err != nil {
		log.Log.Error("Error occurred during CanUpdate (ReadOne) for ListTask: %s", err)
		return false
	}
	return l.CanWrite(doer)
}

// CanCreate determines if a user has the right to create a list task
func (t *ListTask) CanCreate(a web.Auth) bool {
	doer := getUserForRights(a)

	// A user can create an task if he has write acces to its list
	l := &List{ID: t.ListID}
	l.ReadOne()
	return l.CanWrite(doer)
}

// CanRead determines if a user can read a task
func (t *ListTask) CanRead(a web.Auth) bool {
	// A user can read a task if it has access to the list
	list := &List{ID: t.ListID}
	return list.CanRead(a)
}
