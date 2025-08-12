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

func (r *Reaction) setEntityKindFromString() (err error) {
	switch r.EntityKindString {
	case "tasks":
		r.EntityKind = ReactionKindTask
		return
	case "comments":
		r.EntityKind = ReactionKindComment
		return
	}

	return ErrInvalidReactionEntityKind{
		Kind: r.EntityKindString,
	}
}

func (r *Reaction) CanRead(s *xorm.Session, a web.Auth) (bool, int, error) {
	t, err := r.getTask(s)
	if err != nil {
		return false, 0, err
	}
	return t.CanRead(s, a)
}

func (r *Reaction) CanDelete(s *xorm.Session, a web.Auth) (bool, error) {
	t, err := r.getTask(s)
	if err != nil {
		return false, err
	}
	return t.CanUpdate(s, a)
}

func (r *Reaction) CanCreate(s *xorm.Session, a web.Auth) (bool, error) {
	t, err := r.getTask(s)
	if err != nil {
		return false, err
	}
	return t.CanUpdate(s, a)
}

func (r *Reaction) getTask(s *xorm.Session) (t *Task, err error) {
	err = r.setEntityKindFromString()
	if err != nil {
		return
	}

	t = &Task{ID: r.EntityID}

	if r.EntityKind == ReactionKindComment {
		tc := &TaskComment{ID: r.EntityID}
		err = getTaskCommentSimple(s, tc)
		if err != nil {
			return
		}
		t.ID = tc.TaskID
	}

	return
}
