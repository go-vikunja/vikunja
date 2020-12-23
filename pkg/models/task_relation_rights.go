// Copyright 2018-2020 Vikunja and contriubtors. All rights reserved.
//
// This file is part of Vikunja.
//
// Vikunja is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Vikunja is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with Vikunja.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"code.vikunja.io/web"
	"xorm.io/xorm"
)

// CanDelete checks if a user can delete a task relation
func (rel *TaskRelation) CanDelete(s *xorm.Session, a web.Auth) (bool, error) {
	// A user can delete a relation if it can update the base task
	baseTask := &Task{ID: rel.TaskID}
	return baseTask.CanUpdate(s, a)
}

// CanCreate checks if a user can create a new relation between two relations
func (rel *TaskRelation) CanCreate(s *xorm.Session, a web.Auth) (bool, error) {
	// Check if the relation kind is valid
	if !rel.RelationKind.isValid() {
		return false, ErrInvalidRelationKind{Kind: rel.RelationKind}
	}

	// Needs have write access to the base task and at least read access to the other task
	baseTask := &Task{ID: rel.TaskID}
	has, err := baseTask.CanUpdate(s, a)
	if err != nil || !has {
		return false, err
	}

	// We explicitly don't check if the two tasks are on the same list.
	otherTask := &Task{ID: rel.OtherTaskID}
	has, _, err = otherTask.CanRead(s, a)
	if err != nil {
		return false, err
	}
	return has, nil
}
