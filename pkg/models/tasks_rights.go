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
	"code.vikunja.io/web"
	"xorm.io/xorm"
)

// CanDelete checks if the user can delete an task
func (t *Task) CanDelete(s *xorm.Session, a web.Auth) (bool, error) {
	return t.canDoTask(s, a)
}

// CanUpdate determines if a user has the right to update a list task
func (t *Task) CanUpdate(s *xorm.Session, a web.Auth) (bool, error) {
	return t.canDoTask(s, a)
}

// CanCreate determines if a user has the right to create a list task
func (t *Task) CanCreate(s *xorm.Session, a web.Auth) (bool, error) {
	// A user can do a task if he has write acces to its list
	l := &List{ID: t.ListID}
	return l.CanWrite(s, a)
}

// CanRead determines if a user can read a task
func (t *Task) CanRead(s *xorm.Session, a web.Auth) (canRead bool, maxRight int, err error) {
	// Get the task, error out if it doesn't exist
	*t, err = GetTaskByIDSimple(s, t.ID)
	if err != nil {
		return
	}

	// A user can read a task if it has access to the list
	l := &List{ID: t.ListID}
	return l.CanRead(s, a)
}

// CanWrite checks if a user has write access to a task
func (t *Task) CanWrite(s *xorm.Session, a web.Auth) (canWrite bool, err error) {
	return t.canDoTask(s, a)
}

// Helper function to check if a user can do stuff on a list task
func (t *Task) canDoTask(s *xorm.Session, a web.Auth) (bool, error) {
	// Get the task
	ot, err := GetTaskByIDSimple(s, t.ID)
	if err != nil {
		return false, err
	}

	// Check if we're moving the task into a different list to check if the user has sufficient rights for that on the new list
	if t.ListID != 0 && t.ListID != ot.ListID {
		newList := &List{ID: t.ListID}
		can, err := newList.CanWrite(s, a)
		if err != nil {
			return false, err
		}
		if !can {
			return false, ErrGenericForbidden{}
		}
	}

	// A user can do a task if it has write acces to its list
	l := &List{ID: ot.ListID}
	return l.CanWrite(s, a)
}
