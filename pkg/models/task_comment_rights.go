// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2021 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"code.vikunja.io/web"
	"xorm.io/xorm"
)

// CanRead checks if a user can read a comment
func (tc *TaskComment) CanRead(s *xorm.Session, a web.Auth) (bool, int, error) {
	t := Task{ID: tc.TaskID}
	return t.CanRead(s, a)
}

// CanDelete checks if a user can delete a comment
func (tc *TaskComment) CanDelete(s *xorm.Session, a web.Auth) (bool, error) {
	t := Task{ID: tc.TaskID}
	return t.CanWrite(s, a)
}

// CanUpdate checks if a user can update a comment
func (tc *TaskComment) CanUpdate(s *xorm.Session, a web.Auth) (bool, error) {
	t := Task{ID: tc.TaskID}
	return t.CanWrite(s, a)
}

// CanCreate checks if a user can create a new comment
func (tc *TaskComment) CanCreate(s *xorm.Session, a web.Auth) (bool, error) {
	t := Task{ID: tc.TaskID}
	return t.CanWrite(s, a)
}
