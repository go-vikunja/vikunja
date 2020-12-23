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

// CanCreate checks if a user can add a new assignee
func (la *TaskAssginee) CanCreate(s *xorm.Session, a web.Auth) (bool, error) {
	return canDoTaskAssingee(s, la.TaskID, a)
}

// CanCreate checks if a user can add a new assignee
func (ba *BulkAssignees) CanCreate(s *xorm.Session, a web.Auth) (bool, error) {
	return canDoTaskAssingee(s, ba.TaskID, a)
}

// CanDelete checks if a user can delete an assignee
func (la *TaskAssginee) CanDelete(s *xorm.Session, a web.Auth) (bool, error) {
	return canDoTaskAssingee(s, la.TaskID, a)
}

func canDoTaskAssingee(s *xorm.Session, taskID int64, a web.Auth) (bool, error) {
	// Check if the current user can edit the list
	list, err := GetListSimplByTaskID(s, taskID)
	if err != nil {
		return false, err
	}
	return list.CanUpdate(s, a)
}
