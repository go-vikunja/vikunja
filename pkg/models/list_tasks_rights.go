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
)

// CanDelete checks if the user can delete an task
func (i *ListTask) CanDelete(doer *User) bool {
	// Get the task
	lI, err := GetListTaskByID(i.ID)
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
func (i *ListTask) CanUpdate(doer *User) bool {
	// Get the task
	lI, err := GetListTaskByID(i.ID)
	if err != nil {
		log.Log.Error("Error occurred during CanDelete for ListTask: %s", err)
		return false
	}

	// A user can update an task if he has write acces to its list
	l := &List{ID: lI.ListID}
	l.ReadOne()
	return l.CanWrite(doer)
}

// CanCreate determines if a user has the right to create a list task
func (i *ListTask) CanCreate(doer *User) bool {
	// A user can create an task if he has write acces to its list
	l := &List{ID: i.ListID}
	l.ReadOne()
	return l.CanWrite(doer)
}
