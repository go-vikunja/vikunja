//  Vikunja is a todo-list application to facilitate your life.
//  Copyright 2019 Vikunja and contributors. All rights reserved.
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

// CanCreate checks if a user can add a new assignee
func (la *ListTaskAssginee) CanCreate(a web.Auth) bool {
	return canDoListTaskAssingee(la.TaskID, a)
}

// CanCreate checks if a user can add a new assignee
func (ba *BulkAssignees) CanCreate(a web.Auth) bool {
	return canDoListTaskAssingee(ba.TaskID, a)
}

// CanDelete checks if a user can delete an assignee
func (la *ListTaskAssginee) CanDelete(a web.Auth) bool {
	return canDoListTaskAssingee(la.TaskID, a)
}

func canDoListTaskAssingee(taskID int64, a web.Auth) bool {
	// Check if the current user can edit the list
	list, err := GetListSimplByTaskID(taskID)
	if err != nil {
		log.Log.Errorf("Error during canDoListTaskAssingee for ListTaskAssginee: %v", err)
		return false
	}
	return list.CanCreate(a)
}
