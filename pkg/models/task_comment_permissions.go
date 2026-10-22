// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"code.vikunja.io/api/pkg/web"
	"xorm.io/xorm"
)

// CanRead checks if a user can read a comment
func (tc *TaskComment) CanRead(s *xorm.Session, a web.Auth) (bool, int, error) {
	t := Task{ID: tc.TaskID}
	return t.CanRead(s, a)
}

func (tc *TaskComment) canUserModifyTaskComment(s *xorm.Session, a web.Auth) (bool, error) {
	t := Task{ID: tc.TaskID}
	canWriteTask, err := t.CanWrite(s, a)
	if err != nil {
		return false, err
	}
	if !canWriteTask {
		return false, nil
	}

	savedComment := &TaskComment{
		ID:     tc.ID,
		TaskID: tc.TaskID,
	}
	err = getTaskCommentSimple(s, savedComment)
	if err != nil {
		return false, err
	}

	return a.GetID() == savedComment.AuthorID, nil
}

// CanDelete checks if a user can delete a comment
func (tc *TaskComment) CanDelete(s *xorm.Session, a web.Auth) (bool, error) {
	return tc.canUserModifyTaskComment(s, a)
}

// CanUpdate checks if a user can update a comment
func (tc *TaskComment) CanUpdate(s *xorm.Session, a web.Auth) (bool, error) {
	return tc.canUserModifyTaskComment(s, a)
}

// CanCreate checks if a user can create a new comment
func (tc *TaskComment) CanCreate(s *xorm.Session, a web.Auth) (bool, error) {
	t := Task{ID: tc.TaskID}
	return t.CanWrite(s, a)
}
